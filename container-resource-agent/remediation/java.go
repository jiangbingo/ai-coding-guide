package remediation

import (
	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// JavaMemoryFixer Java 应用内存配置修复器
type JavaMemoryFixer struct {
	client kubernetes.Interface
}

// NewJavaMemoryFixer 创建 Java 内存修复器
func NewJavaMemoryFixer(client kubernetes.Interface) *JavaMemoryFixer {
	return &JavaMemoryFixer{
		client: client,
	}
}

// JavaMemoryIssue Java 内存问题
type JavaMemoryIssue struct {
	Namespace      string
	PodName        string
	ContainerName  string
	IssueType      string // Overconfigured, Underconfigured, MissingLimits
	Description    string
	CurrentConfig  string
	RecommendedConfig string
	Severity       string // Critical, High, Medium, Low
}

// DiagnoseJavaMemory 诊断 Java 内存配置
func (jmf *JavaMemoryFixer) DiagnoseJavaMemory(ctx context.Context, namespace, podName string) ([]*JavaMemoryIssue, error) {
	// 获取 Pod
	pod, err := jmf.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 失败: %w", err)
	}

	var issues []*JavaMemoryIssue

	for _, container := range pod.Spec.Containers {
		// 检查是否是 Java 应用
		if !isJavaContainer(container) {
			continue
		}

		// 检查 JVM 内存配置
		containerIssues := jmf.diagnoseContainer(container, pod)
		issues = append(issues, containerIssues...)
	}

	return issues, nil
}

// diagnoseContainer 诊断容器 JVM 配置
func (jmf *JavaMemoryFixer) diagnoseContainer(container v1.Container, pod *v1.Pod) []*JavaMemoryIssue {
	var issues []*JavaMemoryIssue

	// 获取内存限制
	memLimit := container.Resources.Limits.Memory()
	memRequest := container.Resources.Requests.Memory()

	// 获取 JVM 参数
	var jvmOpts string
	for _, env := range container.Env {
		if env.Name == "JAVA_OPTS" || env.Name == "JAVA_TOOL_OPTIONS" || env.Name == "JDK_JAVA_OPTIONS" {
			jvmOpts = env.Value
			break
		}
	}

	// 检查问题

	// 1. 没有设置内存限制
	if memLimit == nil || memLimit.IsZero() {
		issues = append(issues, &JavaMemoryIssue{
			Namespace:      pod.Namespace,
			PodName:        pod.Name,
			ContainerName:  container.Name,
			IssueType:      "MissingLimits",
			Description:    "Java 容器没有设置内存限制",
			Severity:       "Critical",
			RecommendedConfig: "设置 limits.memory 并配置 -XX:MaxRAMPercentage",
		})
		return issues
	}

	// 2. 检查 JVM 堆内存配置
	if jvmOpts == "" {
		// 没有显式 JVM 配置，检查是否使用了容器感知
		issues = append(issues, jmf.checkContainerAware(container, pod))
	} else {
		// 有 JVM 配置，检查是否合理
		issues = append(issues, jmf.checkJVMConfig(container, pod, jvmOpts, memLimit)...)
	}

	// 3. 检查元空间配置
	issues = append(issues, jmf.checkMetaspace(container, pod, jvmOpts)...)

	// 4. 检查直接内存配置
	issues = append(issues, jmf.checkDirectMemory(container, pod, jvmOpts)...)

	// 5. 检查 GC 配置
	issues = append(issues, jmf.checkGCConfig(container, pod, jvmOpts)...)

	return issues
}

// checkContainerAware 检查容器感知配置
func (jmf *JavaMemoryFixer) checkContainerAware(container v1.Container, pod *v1.Pod) *JavaMemoryIssue {
	// Java 10+ 默认启用容器感知
	// 但建议显式配置

	memLimit := container.Resources.Limits.Memory()
	limitBytes := memLimit.Value()

	// 推荐使用百分比配置
	return &JavaMemoryIssue{
		Namespace:     pod.Namespace,
		PodName:       pod.Name,
		ContainerName: container.Name,
		IssueType:     "NoExplicitConfig",
		Description:   "建议显式配置 JVM 内存参数",
		Severity:      "Medium",
		CurrentConfig: "无",
		RecommendedConfig: fmt.Sprintf(`环境变量配置:
JAVA_OPTS: "-XX:MaxRAMPercentage=75.0 -XX:InitialRAMPercentage=50.0"

或固定值配置（容器内存限制 %d MiB）:
JAVA_OPTS: "-Xmx%dM -Xms%dM"`, limitBytes/1024/1024, limitBytes*75/100/1024/1024, limitBytes*50/100/1024/1024),
	}
}

// checkJVMConfig 检查 JVM 配置
func (jmf *JavaMemoryFixer) checkJVMConfig(container v1.Container, pod *v1.Pod, jvmOpts string, memLimit *resource.Quantity) []*JavaMemoryIssue {
	var issues []*JavaMemoryIssue

	limitBytes := memLimit.Value()

	// 解析 -Xmx 参数
	xmx := parseXmx(jvmOpts)
	if xmx > 0 {
		// 检查是否超过容器限制的 85%
		if xmx > limitBytes*85/100 {
			issues = append(issues, &JavaMemoryIssue{
				Namespace:     pod.Namespace,
				PodName:       pod.Name,
				ContainerName: container.Name,
				IssueType:     "Overconfigured",
				Description:   fmt.Sprintf("JVM 堆内存 %d MiB 超过容器限制的 85%%", xmx/1024/1024),
				Severity:      "High",
				CurrentConfig: fmt.Sprintf("-Xmx%dM", xmx/1024/1024),
				RecommendedConfig: fmt.Sprintf("-Xmx%dM (容器限制的 75%%)", limitBytes*75/100/1024/1024),
			})
		}

		// 检查是否低于容器限制的 50%
		if xmx < limitBytes*50/100 {
			issues = append(issues, &JavaMemoryIssue{
				Namespace:     pod.Namespace,
				PodName:       pod.Name,
				ContainerName: container.Name,
				IssueType:     "Underconfigured",
				Description:   fmt.Sprintf("JVM 堆内存 %d MiB 低于容器限制的 50%%，可能浪费资源", xmx/1024/1024),
				Severity:      "Medium",
				CurrentConfig: fmt.Sprintf("-Xmx%dM", xmx/1024/1024),
				RecommendedConfig: fmt.Sprintf("-Xmx%dM (容器限制的 75%%)", limitBytes*75/100/1024/1024),
			})
		}
	}

	// 检查 -Xms 和 -Xmx 是否一致
	xms := parseXms(jvmOpts)
	if xms > 0 && xmx > 0 && xms != xmx {
		issues = append(issues, &JavaMemoryIssue{
			Namespace:     pod.Namespace,
			PodName:       pod.Name,
			ContainerName: container.Name,
			IssueType:     "InconsistentHeap",
			Description:   "JVM 初始堆和最大堆不一致，可能导致性能问题",
			Severity:      "Medium",
			CurrentConfig: fmt.Sprintf("-Xms%dM -Xmx%dM", xms/1024/1024, xmx/1024/1024),
			RecommendedConfig: fmt.Sprintf("-Xms%dM -Xmx%dM", xmx/1024/1024, xmx/1024/1024),
		})
	}

	return issues
}

// checkMetaspace 检查元空间配置
func (jmf *JavaMemoryFixer) checkMetaspace(container v1.Container, pod *v1.Pod, jvmOpts string) []*JavaMemoryIssue {
	var issues []*JavaMemoryIssue

	// 检查是否有元空间限制
	if !strings.Contains(jvmOpts, "-XX:MaxMetaspaceSize") {
		issues = append(issues, &JavaMemoryIssue{
			Namespace:     pod.Namespace,
			PodName:       pod.Name,
			ContainerName: container.Name,
			IssueType:     "UnlimitedMetaspace",
			Description:   "元空间没有限制，可能导致内存泄漏",
			Severity:      "High",
			CurrentConfig: "无限制",
			RecommendedConfig: "-XX:MaxMetaspaceSize=256M",
		})
	}

	return issues
}

// checkDirectMemory 检查直接内存配置
func (jmf *JavaMemoryFixer) checkDirectMemory(container v1.Container, pod *v1.Pod, jvmOpts string) []*JavaMemoryIssue {
	var issues []*JavaMemoryIssue

	// 如果使用了 Netty 等框架，需要配置直接内存
	// 检查是否有限制
	if !strings.Contains(jvmOpts, "-XX:MaxDirectMemorySize") {
		// 检查是否是网络应用
		if isNetworkApplication(container) {
			issues = append(issues, &JavaMemoryIssue{
				Namespace:     pod.Namespace,
				PodName:       pod.Name,
				ContainerName: container.Name,
				IssueType:     "MissingDirectMemory",
				Description:   "网络应用建议配置直接内存限制",
				Severity:      "Medium",
				CurrentConfig: "无",
				RecommendedConfig: "-XX:MaxDirectMemorySize=512M",
			})
		}
	}

	return issues
}

// checkGCConfig 检查 GC 配置
func (jmf *JavaMemoryFixer) checkGCConfig(container v1.Container, pod *v1.Pod, jvmOpts string) []*JavaMemoryIssue {
	var issues []*JavaMemoryIssue

	// 检查 GC 配置
	memLimit := container.Resources.Limits.Memory()
	limitGB := memLimit.Value() / 1024 / 1024 / 1024

	// 根据内存大小推荐 GC
	if limitGB >= 4 {
		// 大内存建议使用 G1GC
		if !strings.Contains(jvmOpts, "-XX:+UseG1GC") && !strings.Contains(jvmOpts, "-XX:+UseZGC") {
			issues = append(issues, &JavaMemoryIssue{
				Namespace:     pod.Namespace,
				PodName:       pod.Name,
				ContainerName: container.Name,
				IssueType:     "SuboptimalGC",
				Description:   fmt.Sprintf("内存 %d GB 建议使用 G1GC 或 ZGC", limitGB),
				Severity:      "Low",
				CurrentConfig: "默认",
				RecommendedConfig: "-XX:+UseG1GC",
			})
		}
	} else if limitGB <= 2 {
		// 小内存建议使用 SerialGC
		if !strings.Contains(jvmOpts, "-XX:+UseSerialGC") {
			issues = append(issues, &JavaMemoryIssue{
				Namespace:     pod.Namespace,
				PodName:       pod.Name,
				ContainerName: container.Name,
				IssueType:     "SuboptimalGC",
				Description:   fmt.Sprintf("内存 %d GB 建议使用 SerialGC", limitGB),
				Severity:      "Low",
				CurrentConfig: "默认",
				RecommendedConfig: "-XX:+UseSerialGC",
			})
		}
	}

	return issues
}

// FixJavaMemory 修复 Java 内存配置
func (jmf *JavaMemoryFixer) FixJavaMemory(ctx context.Context, namespace, podName string, dryRun bool) error {
	// 获取 Pod
	pod, err := jmf.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 Pod 失败: %w", err)
	}

	// 诊断问题
	issues, err := jmf.DiagnoseJavaMemory(ctx, namespace, podName)
	if err != nil {
		return err
	}

	if len(issues) == 0 {
		fmt.Println("没有发现 Java 内存配置问题")
		return nil
	}

	// 生成修复建议
	fmt.Println("发现的 Java 内存问题:")
	for _, issue := range issues {
		fmt.Printf("- [%s] %s: %s\n", issue.Severity, issue.ContainerName, issue.Description)
		fmt.Printf("  当前配置: %s\n", issue.CurrentConfig)
		fmt.Printf("  推荐配置: %s\n", issue.RecommendedConfig)
	}

	if dryRun {
		fmt.Println("\nDry run 模式，不执行修复")
		return nil
	}

	// 实际修复需要修改 Deployment/StatefulSet 等
	fmt.Println("\n注意: 实际修复需要修改 Deployment/StatefulSet 配置")
	fmt.Println("请手动应用以下配置:")

	for _, issue := range issues {
		if issue.RecommendedConfig != "" {
			fmt.Printf("\n# %s (%s)\n%s\n", issue.ContainerName, issue.IssueType, issue.RecommendedConfig)
		}
	}

	return nil
}

// 辅助函数

func isJavaContainer(container v1.Container) bool {
	// 检查镜像名
	image := strings.ToLower(container.Image)
	if strings.Contains(image, "java") ||
		strings.Contains(image, "jdk") ||
		strings.Contains(image, "openjdk") ||
		strings.Contains(image, "tomcat") ||
		strings.Contains(image, "spring") ||
		strings.Contains(image, "wildfly") ||
		strings.Contains(image, "jboss") {
		return true
	}

	// 检查命令
	for _, cmd := range container.Command {
		if strings.Contains(cmd, "java") {
			return true
		}
	}

	for _, cmd := range container.Args {
		if strings.Contains(cmd, "java") || strings.Contains(cmd, "-jar") {
			return true
		}
	}

	// 检查环境变量
	for _, env := range container.Env {
		if strings.HasPrefix(env.Name, "JAVA_") {
			return true
		}
	}

	return false
}

func parseXmx(jvmOpts string) int64 {
	// 解析 -Xmx 参数
	// 格式: -Xmx512m, -Xmx2g, -Xmx512M, -Xmx2G

	parts := strings.Fields(jvmOpts)
	for _, part := range parts {
		if strings.HasPrefix(part, "-Xmx") {
			value := part[4:]
			return parseMemoryValue(value)
		}
	}

	return 0
}

func parseXms(jvmOpts string) int64 {
	// 解析 -Xms 参数
	parts := strings.Fields(jvmOpts)
	for _, part := range parts {
		if strings.HasPrefix(part, "-Xms") {
			value := part[4:]
			return parseMemoryValue(value)
		}
	}

	return 0
}

func parseMemoryValue(value string) int64 {
	value = strings.TrimSpace(value)
	if len(value) == 0 {
		return 0
	}

	var multiplier int64 = 1
	lastChar := strings.ToLower(string(value[len(value)-1]))

	switch lastChar {
	case 'g':
		multiplier = 1024 * 1024 * 1024
		value = value[:len(value)-1]
	case 'm':
		multiplier = 1024 * 1024
		value = value[:len(value)-1]
	case 'k':
		multiplier = 1024
		value = value[:len(value)-1]
	}

	var result int64
	fmt.Sscanf(value, "%d", &result)
	return result * multiplier
}

func isNetworkApplication(container v1.Container) bool {
	// 检查是否是网络应用（使用 Netty 等）
	for _, env := range container.Env {
		if strings.Contains(env.Name, "NETTY") ||
			strings.Contains(env.Name, "VERTX") ||
			strings.Contains(env.Name, "REACTOR") {
			return true
		}
	}

	// 检查镜像
	image := strings.ToLower(container.Image)
	if strings.Contains(image, "netty") ||
		strings.Contains(image, "vertx") ||
		strings.Contains(image, "spring-webflux") {
		return true
	}

	return false
}
