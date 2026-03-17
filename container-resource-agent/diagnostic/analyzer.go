package diagnostic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ResourceAnalyzer 容器资源分析器
type ResourceAnalyzer struct {
	client    kubernetes.Interface
	cgroupRoot string
}

// NewResourceAnalyzer 创建资源分析器
func NewResourceAnalyzer(client kubernetes.Interface) *ResourceAnalyzer {
	cgroupRoot := "/sys/fs/cgroup"
	if _, err := os.Stat(cgroupRoot); os.IsNotExist(err) {
		cgroupRoot = "/sys/fs/cgroup/unified"
	}
	return &ResourceAnalyzer{
		client:    client,
		cgroupRoot: cgroupRoot,
	}
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	Namespace     string
	PodName       string
	ContainerName string
	QoSClass      v1.PodQOSClass

	// CPU 使用
	CPUUsageNanoCores uint64
	CPUQuota          int64
	CPUPeriod         int64
	CPUShares         uint64
	CPUUtilization    float64 // 百分比

	// 内存使用
	MemoryUsageBytes     uint64
	MemoryWorkingSet     uint64
	MemoryLimit          uint64
	MemoryRequest        uint64
	MemoryUtilization    float64 // 百分比
	MemoryCache          uint64
	MemoryRSS            uint64

	// OOM 相关
	OOMCount      uint64
	OOMKillCount  uint64
	LastOOMTime   *time.Time

	// I/O 使用
	IoReadBytes  uint64
	IoWriteBytes uint64
	IoReadOps    uint64
	IoWriteOps   uint64

	// 诊断信息
	Warnings []string
	Errors   []string
}

// AnalyzePod 分析 Pod 资源使用
func (ra *ResourceAnalyzer) AnalyzePod(ctx context.Context, namespace, podName string) (*ResourceUsage, error) {
	pod, err := ra.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 失败: %w", err)
	}

	usages := make([]*ResourceUsage, 0)
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.ContainerID == "" {
			continue
		}

		usage, err := ra.analyzeContainer(pod, containerStatus)
		if err != nil {
			return nil, fmt.Errorf("分析容器 %s 失败: %w", containerStatus.Name, err)
		}
		usages = append(usages, usage)
	}

	// 聚合所有容器的使用情况
	return ra.aggregateUsage(usages), nil
}

// analyzeContainer 分析单个容器
func (ra *ResourceAnalyzer) analyzeContainer(pod *v1.Pod, status v1.ContainerStatus) (*ResourceUsage, error) {
	containerID := extractContainerID(status.ContainerID)
	cgroupPath := ra.findCgroupPath(pod, containerID)
	if cgroupPath == "" {
		return nil, fmt.Errorf("找不到 cgroup 路径")
	}

	usage := &ResourceUsage{
		Namespace:     pod.Namespace,
		PodName:       pod.Name,
		ContainerName: status.Name,
		QoSClass:      pod.Status.QOSClass,
	}

	// 读取 CPU 使用
	if err := ra.readCPUStats(cgroupPath, usage); err != nil {
		usage.Errors = append(usage.Errors, fmt.Sprintf("读取 CPU 统计失败: %v", err))
	}

	// 读取内存使用
	if err := ra.readMemoryStats(cgroupPath, usage); err != nil {
		usage.Errors = append(usage.Errors, fmt.Sprintf("读取内存统计失败: %v", err))
	}

	// 读取 I/O 使用
	if err := ra.readIOStats(cgroupPath, usage); err != nil {
		usage.Warnings = append(usage.Warnings, fmt.Sprintf("读取 I/O 统计失败: %v", err))
	}

	// 读取 OOM 信息
	if err := ra.readOOMStats(cgroupPath, usage); err != nil {
		usage.Warnings = append(usage.Warnings, fmt.Sprintf("读取 OOM 统计失败: %v", err))
	}

	// 计算使用率
	ra.calculateUtilization(usage, pod, status.Name)

	// 诊断问题
	ra.diagnose(usage, pod, status.Name)

	return usage, nil
}

// findCgroupPath 查找 cgroup 路径
func (ra *ResourceAnalyzer) findCgroupPath(pod *v1.Pod, containerID string) string {
	qosClass := strings.ToLower(string(pod.Status.QOSClass))
	podUID := string(pod.UID)

	// cgroups v2 路径模式
	paths := []string{
		filepath.Join(ra.cgroupRoot, "kubepods", qosClass, fmt.Sprintf("pod%s", podUID), containerID),
		filepath.Join(ra.cgroupRoot, "kubepods", qosClass, fmt.Sprintf("pod%s", podUID), fmt.Sprintf("cri-containerd-%s", containerID)),
		filepath.Join(ra.cgroupRoot, "kubepods", qosClass, fmt.Sprintf("pod%s", podUID)),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// readCPUStats 读取 CPU 统计
func (ra *ResourceAnalyzer) readCPUStats(cgroupPath string, usage *ResourceUsage) error {
	// 读取 CPU 使用
	statPath := filepath.Join(cgroupPath, "cpu.stat")
	if data, err := os.ReadFile(statPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				switch parts[0] {
				case "usage_usec":
					value, _ := strconv.ParseUint(parts[1], 10, 64)
					usage.CPUUsageNanoCores = value * 1000
				}
			}
		}
	}

	// 读取 CPU 配额
	quotaPath := filepath.Join(cgroupPath, "cpu.max")
	if data, err := os.ReadFile(quotaPath); err == nil {
		parts := strings.Fields(string(data))
		if len(parts) >= 2 {
			if parts[0] != "max" {
				usage.CPUQuota, _ = strconv.ParseInt(parts[0], 10, 64)
			}
			usage.CPUPeriod, _ = strconv.ParseInt(parts[1], 10, 64)
		}
	}

	// 读取 CPU 权重
	weightPath := filepath.Join(cgroupPath, "cpu.weight")
	if data, err := os.ReadFile(weightPath); err == nil {
		weight := strings.TrimSpace(string(data))
		usage.CPUShares, _ = strconv.ParseUint(weight, 10, 64)
	}

	return nil
}

// readMemoryStats 读取内存统计
func (ra *ResourceAnalyzer) readMemoryStats(cgroupPath string, usage *ResourceUsage) error {
	// 读取当前内存使用
	currentPath := filepath.Join(cgroupPath, "memory.current")
	if data, err := os.ReadFile(currentPath); err == nil {
		usage.MemoryUsageBytes, _ = strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	}

	// 读取内存峰值
	peakPath := filepath.Join(cgroupPath, "memory.peak")
	if data, err := os.ReadFile(peakPath); err == nil {
		peak, _ := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
		_ = peak // 可用于分析
	}

	// 读取内存限制
	maxPath := filepath.Join(cgroupPath, "memory.max")
	if data, err := os.ReadFile(maxPath); err == nil {
		max := strings.TrimSpace(string(data))
		if max != "max" {
			usage.MemoryLimit, _ = strconv.ParseUint(max, 10, 64)
		}
	}

	// 读取详细统计
	statPath := filepath.Join(cgroupPath, "memory.stat")
	if data, err := os.ReadFile(statPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				switch parts[0] {
				case "file":
					usage.MemoryCache, _ = strconv.ParseUint(parts[1], 10, 64)
				case "anon":
					usage.MemoryRSS, _ = strconv.ParseUint(parts[1], 10, 64)
				}
			}
		}
	}

	// 计算 Working Set = Usage - Cache
	usage.MemoryWorkingSet = usage.MemoryUsageBytes - usage.MemoryCache

	return nil
}

// readIOStats 读取 I/O 统计
func (ra *ResourceAnalyzer) readIOStats(cgroupPath string, usage *ResourceUsage) error {
	statPath := filepath.Join(cgroupPath, "io.stat")
	if data, err := os.ReadFile(statPath); err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// 格式: 253:0 rbytes=123 wbytes=456 rios=7 wios=8
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		for _, part := range parts[1:] {
			kv := strings.Split(part, "=")
			if len(kv) != 2 {
				continue
			}

			value, _ := strconv.ParseUint(kv[1], 10, 64)
			switch kv[0] {
			case "rbytes":
				usage.IoReadBytes += value
			case "wbytes":
				usage.IoWriteBytes += value
			case "rios":
				usage.IoReadOps += value
			case "wios":
				usage.IoWriteOps += value
			}
		}
	}

	return nil
}

// readOOMStats 读取 OOM 统计
func (ra *ResourceAnalyzer) readOOMStats(cgroupPath string, usage *ResourceUsage) error {
	// 读取 OOM 事件
	eventsPath := filepath.Join(cgroupPath, "memory.events")
	if data, err := os.ReadFile(eventsPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				switch parts[0] {
				case "oom":
					usage.OOMCount, _ = strconv.ParseUint(parts[1], 10, 64)
				case "oom_kill":
					usage.OOMKillCount, _ = strconv.ParseUint(parts[1], 10, 64)
				}
			}
		}
	}

	return nil
}

// calculateUtilization 计算使用率
func (ra *ResourceAnalyzer) calculateUtilization(usage *ResourceUsage, pod *v1.Pod, containerName string) {
	// 查找容器资源配置
	for _, container := range pod.Spec.Containers {
		if container.Name != containerName {
			continue
		}

		// CPU 使用率
		if usage.CPUQuota > 0 && usage.CPUPeriod > 0 {
			cpuLimit := float64(usage.CPUQuota) / float64(usage.CPUPeriod)
			// 使用率 = 实际使用 / 限制 * 100
			// 注意：这里需要累计值和时间窗口来计算
			_ = cpuLimit
		}

		// 内存使用率
		if usage.MemoryLimit > 0 {
			usage.MemoryUtilization = float64(usage.MemoryUsageBytes) / float64(usage.MemoryLimit) * 100
		}

		// 记录请求值
		if request := container.Resources.Requests.Memory(); request != nil {
			usage.MemoryRequest = request.Value()
		}

		break
	}
}

// diagnose 诊断问题
func (ra *ResourceAnalyzer) diagnose(usage *ResourceUsage, pod *v1.Pod, containerName string) {
	// 检查内存使用率
	if usage.MemoryUtilization > 90 {
		usage.Warnings = append(usage.Warnings, "内存使用率超过 90%，存在 OOM 风险")
	}
	if usage.MemoryUtilization > 95 {
		usage.Errors = append(usage.Errors, "内存使用率超过 95%，立即 OOM 风险")
	}

	// 检查 OOM 历史
	if usage.OOMCount > 0 {
		usage.Warnings = append(usage.Warnings, fmt.Sprintf("发生过 %d 次 OOM 事件", usage.OOMCount))
	}
	if usage.OOMKillCount > 0 {
		usage.Errors = append(usage.Errors, fmt.Sprintf("发生过 %d 次 OOM Kill", usage.OOMKillCount))
	}

	// 检查 QoS 类别
	if usage.QoSClass == v1.PodQOSBestEffort {
		usage.Warnings = append(usage.Warnings, "Pod QoS 为 BestEffort，没有资源保证")
	}

	// 检查内存限制
	if usage.MemoryLimit == 0 {
		usage.Warnings = append(usage.Warnings, "没有设置内存限制，可能导致资源争抢")
	}

	// 检查 CPU 配额
	if usage.CPUQuota == 0 {
		usage.Warnings = append(usage.Warnings, "没有设置 CPU 限制，可能导致 CPU 争抢")
	}
}

// aggregateUsage 聚合使用情况
func (ra *ResourceAnalyzer) aggregateUsage(usages []*ResourceUsage) *ResourceUsage {
	if len(usages) == 0 {
		return nil
	}

	// 使用第一个作为基础
	aggregate := usages[0]

	for i := 1; i < len(usages); i++ {
		u := usages[i]

		aggregate.CPUUsageNanoCores += u.CPUUsageNanoCores
		aggregate.MemoryUsageBytes += u.MemoryUsageBytes
		aggregate.MemoryWorkingSet += u.MemoryWorkingSet
		aggregate.MemoryCache += u.MemoryCache
		aggregate.MemoryRSS += u.MemoryRSS
		aggregate.IoReadBytes += u.IoReadBytes
		aggregate.IoWriteBytes += u.IoWriteBytes
		aggregate.IoReadOps += u.IoReadOps
		aggregate.IoWriteOps += u.IoWriteOps

		if u.OOMCount > aggregate.OOMCount {
			aggregate.OOMCount = u.OOMCount
		}
		if u.OOMKillCount > aggregate.OOMKillCount {
			aggregate.OOMKillCount = u.OOMKillCount
		}

		aggregate.Warnings = append(aggregate.Warnings, u.Warnings...)
		aggregate.Errors = append(aggregate.Errors, u.Errors...)
	}

	// 重新计算使用率
	if aggregate.MemoryLimit > 0 {
		aggregate.MemoryUtilization = float64(aggregate.MemoryUsageBytes) / float64(aggregate.MemoryLimit) * 100
	}

	return aggregate
}

// extractContainerID 从容器 ID 字符串中提取 ID
func extractContainerID(containerID string) string {
	// 格式: containerd://<id> 或 docker://<id>
	parts := strings.Split(containerID, "://")
	if len(parts) == 2 {
		return parts[1]
	}
	return containerID
}
