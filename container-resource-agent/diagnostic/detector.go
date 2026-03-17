package diagnostic

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ResourceContention 资源争抢检测结果
type ResourceContention struct {
	NodeName        string
	Severity        string // Critical, High, Medium, Low
	Type            string // CPU, Memory, IO
	Description     string
	AffectedPods    []string
	Recommendations []string
	Timestamp       time.Time
}

// ContentationDetector 资源争抢检测器
type ContentationDetector struct {
	client kubernetes.Interface
	analyzer *ResourceAnalyzer
}

// NewContentationDetector 创建争抢检测器
func NewContentationDetector(client kubernetes.Interface) *ContentationDetector {
	return &ContentationDetector{
		client:   client,
		analyzer: NewResourceAnalyzer(client),
	}
}

// DetectNodeContention 检测节点资源争抢
func (cd *ContentationDetector) DetectNodeContention(ctx context.Context, nodeName string) ([]*ResourceContention, error) {
	// 获取节点信息
	node, err := cd.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点失败: %w", err)
	}

	// 获取节点上的 Pod
	pods, err := cd.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	var contentions []*ResourceContention

	// 检测 CPU 争抢
	cpuContention := cd.detectCPUContention(node, pods)
	contentions = append(contentions, cpuContention...)

	// 检测内存争抢
	memContention := cd.detectMemoryContention(node, pods)
	contentions = append(contentions, memContention...)

	// 检测 I/O 争抢
	ioContention := cd.detectIOContention(node, pods)
	contentions = append(contentions, ioContention...)

	// 检测 QoS 不平衡
	qosContention := cd.detectQoSImbalance(node, pods)
	contentions = append(contentions, qosContention...)

	return contentions, nil
}

// detectCPUContention 检测 CPU 争抢
func (cd *ContentationDetector) detectCPUContention(node *v1.Node, pods *v1.PodList) []*ResourceContention {
	var contentions []*ResourceContention

	// 计算节点 CPU 资源
	nodeCPU := node.Status.Capacity.Cpu()
	allocatableCPU := node.Status.Allocatable.Cpu()

	// 计算总请求和限制
	var totalCPURequest, totalCPULimit resource.Quantity
	var overLimitPods []string

	for _, pod := range pods.Items {
		if pod.Status.Phase != v1.PodRunning {
			continue
		}

		for _, container := range pod.Spec.Containers {
			if request := container.Resources.Requests.Cpu(); request != nil {
				totalCPURequest.Add(*request)
			}
			if limit := container.Resources.Limits.Cpu(); limit != nil {
				totalCPULimit.Add(*limit)
			}
		}

		// 检查是否超配
		if isOvercommitted(&pod) {
			overLimitPods = append(overLimitPods, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
		}
	}

	// 计算超配率
	allocatableMilli := allocatableCPU.MilliValue()
	requestMilli := totalCPURequest.MilliValue()
	limitMilli := totalCPULimit.MilliValue()

	requestOvercommit := float64(requestMilli) / float64(allocatableMilli)
	limitOvercommit := float64(limitMilli) / float64(allocatableMilli)

	// 判断争抢程度
	if limitOvercommit > 3.0 {
		contentions = append(contentions, &ResourceContention{
			NodeName:    node.Name,
			Severity:    "Critical",
			Type:        "CPU",
			Description: fmt.Sprintf("CPU 限制超配率 %.2f，严重争抢风险", limitOvercommit),
			AffectedPods: overLimitPods,
			Recommendations: []string{
				"减少 CPU 限制超配",
				"将关键业务迁移到 Guaranteed QoS",
				"考虑水平扩展或增加节点",
			},
			Timestamp: time.Now(),
		})
	} else if limitOvercommit > 2.0 {
		contentions = append(contentions, &ResourceContention{
			NodeName:    node.Name,
			Severity:    "High",
			Type:        "CPU",
			Description: fmt.Sprintf("CPU 限制超配率 %.2f，存在争抢", limitOvercommit),
			AffectedPods: overLimitPods,
			Recommendations: []string{
				"监控 CPU 节流情况",
				"考虑调整资源配置",
			},
			Timestamp: time.Now(),
		})
	} else if requestOvercommit > 1.5 {
		contentions = append(contentions, &ResourceContention{
			NodeName:    node.Name,
			Severity:    "Medium",
			Type:        "CPU",
			Description: fmt.Sprintf("CPU 请求超配率 %.2f，轻微争抢", requestOvercommit),
			Recommendations: []string{
				"监控 CPU 使用情况",
				"优化资源配置",
			},
			Timestamp: time.Now(),
		})
	}

	// 检查 BestEffort Pod 数量
	var bestEffortCount int
	for _, pod := range pods.Items {
		if pod.Status.QOSClass == v1.PodQOSBestEffort {
			bestEffortCount++
		}
	}

	if bestEffortCount > 5 {
		contentions = append(contentions, &ResourceContention{
			NodeName:    node.Name,
			Severity:    "Medium",
			Type:        "CPU",
			Description: fmt.Sprintf("节点上有 %d 个 BestEffort Pod，可能影响性能", bestEffortCount),
			Recommendations: []string{
				"为 Pod 设置资源限制",
				"使用 LimitRange 强制资源配置",
			},
			Timestamp: time.Now(),
		})
	}

	return contentions
}

// detectMemoryContention 检测内存争抢
func (cd *ContentationDetector) detectMemoryContention(node *v1.Node, pods *v1.PodList) []*ResourceContention {
	var contentions []*ResourceContention

	// 计算节点内存资源
	allocatableMem := node.Status.Allocatable.Memory()

	// 计算总请求和限制
	var totalMemRequest, totalMemLimit resource.Quantity
	var noLimitPods []string
	var highUsagePods []string

	for _, pod := range pods.Items {
		if pod.Status.Phase != v1.PodRunning {
			continue
		}

		podHasLimit := false
		for _, container := range pod.Spec.Containers {
			if request := container.Resources.Requests.Memory(); request != nil {
				totalMemRequest.Add(*request)
			}
			if limit := container.Resources.Limits.Memory(); limit != nil {
				totalMemLimit.Add(*limit)
				podHasLimit = true
			}
		}

		if !podHasLimit {
			noLimitPods = append(noLimitPods, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
		}

		// 检查内存使用率高的 Pod
		for _, containerStatus := range pod.Status.ContainerStatuses {
			// 这里简化处理，实际应该读取 cgroup 数据
			_ = containerStatus
		}
	}

	// 计算超配率
	allocatableBytes := allocatableMem.Value()
	requestBytes := totalMemRequest.Value()
	limitBytes := totalMemLimit.Value()

	requestOvercommit := float64(requestBytes) / float64(allocatableBytes)
	limitOvercommit := float64(limitBytes) / float64(allocatableBytes)

	// 判断争抢程度（内存超配更危险）
	if limitOvercommit > 1.5 {
		contentions = append(contentions, &ResourceContention{
			NodeName:    node.Name,
			Severity:    "Critical",
			Type:        "Memory",
			Description: fmt.Sprintf("内存限制超配率 %.2f，存在 OOM 风险", limitOvercommit),
			AffectedPods: noLimitPods,
			Recommendations: []string{
				"立即减少内存超配",
				"为所有 Pod 设置内存限制",
				"考虑增加节点内存或水平扩展",
			},
			Timestamp: time.Now(),
		})
	} else if limitOvercommit > 1.2 {
		contentions = append(contentions, &ResourceContention{
			NodeName:    node.Name,
			Severity:    "High",
			Type:        "Memory",
			Description: fmt.Sprintf("内存限制超配率 %.2f，存在争抢", limitOvercommit),
			Recommendations: []string{
				"监控内存使用",
				"减少内存超配",
			},
			Timestamp: time.Now(),
		})
	} else if requestOvercommit > 1.0 {
		contentions = append(contentions, &ResourceContention{
			NodeName:    node.Name,
			Severity:    "Medium",
			Type:        "Memory",
			Description: fmt.Sprintf("内存请求超配率 %.2f", requestOvercommit),
			Recommendations: []string{
				"监控内存压力",
			},
			Timestamp: time.Now(),
		})
	}

	// 检查无内存限制的 Pod
	if len(noLimitPods) > 0 && len(noLimitPods) <= 10 {
		contentions = append(contentions, &ResourceContention{
			NodeName:    node.Name,
			Severity:    "High",
			Type:        "Memory",
			Description: fmt.Sprintf("%d 个 Pod 没有设置内存限制", len(noLimitPods)),
			AffectedPods: noLimitPods,
			Recommendations: []string{
				"为所有 Pod 设置内存限制",
				"使用 LimitRange 强制限制",
			},
			Timestamp: time.Now(),
		})
	}

	// 检查高内存使用 Pod
	if len(highUsagePods) > 0 {
		contentions = append(contentions, &ResourceContention{
			NodeName:    node.Name,
			Severity:    "Medium",
			Type:        "Memory",
			Description: "存在高内存使用的 Pod",
			AffectedPods: highUsagePods,
			Recommendations: []string{
				"检查内存泄漏",
				"增加内存限制",
			},
			Timestamp: time.Now(),
		})
	}

	return contentions
}

// detectIOContention 检测 I/O 争抢
func (cd *ContentationDetector) detectIOContention(node *v1.Node, pods *v1.PodList) []*ResourceContention {
	var contentions []*ResourceContention

	// 统计高 I/O 需求的 Pod
	var heavyIOPods []string
	var noIOPSLimit []string

	for _, pod := range pods.Items {
		if pod.Status.Phase != v1.PodRunning {
			continue
		}

		// 检查是否有 volume 挂载
		hasVolume := len(pod.Spec.Volumes) > 0

		// 检查是否是数据库类应用（通过标签或名称判断）
		isDatabase := isDatabasePod(&pod)

		if hasVolume && isDatabase {
			heavyIOPods = append(heavyIOPods, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
		}

		// Kubernetes 默认不支持 IOPS 限制，记录提示
		if isDatabase {
			noIOPSLimit = append(noIOPSLimit, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
		}
	}

	if len(heavyIOPods) > 3 {
		contentions = append(contentions, &ResourceContention{
			NodeName:    node.Name,
			Severity:    "Medium",
			Type:        "IO",
			Description: fmt.Sprintf("节点上有 %d 个高 I/O 需求的 Pod", len(heavyIOPods)),
			AffectedPods: heavyIOPods,
			Recommendations: []string{
				"考虑使用专用存储节点",
				"使用 Local PV 或高性能存储",
				"分离高 I/O 负载",
			},
			Timestamp: time.Now(),
		})
	}

	return contentions
}

// detectQoSImbalance 检测 QoS 不平衡
func (cd *ContentationDetector) detectQoSImbalance(node *v1.Node, pods *v1.PodList) []*ResourceContention {
	var contentions []*ResourceContention

	// 统计各 QoS 类别数量
	qosCounts := make(map[v1.PodQOSClass]int)
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			qosCounts[pod.Status.QOSClass]++
		}
	}

	total := qosCounts[v1.PodQOSGuaranteed] + qosCounts[v1.PodQOSBurstable] + qosCounts[v1.PodQOSBestEffort]

	// 检查 BestEffort 比例
	bestEffortRatio := float64(qosCounts[v1.PodQOSBestEffort]) / float64(total)
	if bestEffortRatio > 0.3 {
		contentions = append(contentions, &ResourceContention{
			NodeName:    node.Name,
			Severity:    "Medium",
			Type:        "QoS",
			Description: fmt.Sprintf("BestEffort Pod 占比 %.1f%%，资源隔离不足", bestEffortRatio*100),
			Recommendations: []string{
				"为 Pod 设置资源限制",
				"使用 ResourceQuota 限制 BestEffort Pod",
			},
			Timestamp: time.Now(),
		})
	}

	// 检查是否有足够的 Guaranteed Pod
	guaranteedRatio := float64(qosCounts[v1.PodQOSGuaranteed]) / float64(total)
	if total > 10 && guaranteedRatio < 0.1 {
		contentions = append(contentions, &ResourceContention{
			NodeName:    node.Name,
			Severity:    "Low",
			Type:        "QoS",
			Description: fmt.Sprintf("Guaranteed Pod 占比 %.1f%%，关键业务可能受影响", guaranteedRatio*100),
			Recommendations: []string{
				"为关键业务设置 Guaranteed QoS",
			},
			Timestamp: time.Now(),
		})
	}

	return contentions
}

// DetectClusterContention 检测集群级别的资源争抢
func (cd *ContentationDetector) DetectClusterContention(ctx context.Context) ([]*ResourceContention, error) {
	// 获取所有节点
	nodes, err := cd.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点列表失败: %w", err)
	}

	var allContentions []*ResourceContention

	for _, node := range nodes.Items {
		contentions, err := cd.DetectNodeContention(ctx, node.Name)
		if err != nil {
			// 记录错误但继续处理其他节点
			allContentions = append(allContentions, &ResourceContention{
				NodeName:    node.Name,
				Severity:    "Low",
				Type:        "Unknown",
				Description: fmt.Sprintf("检测失败: %v", err),
				Timestamp:   time.Now(),
			})
			continue
		}

		allContentions = append(allContentions, contentions...)
	}

	return allContentions, nil
}

// isOvercommitted 检查 Pod 是否超配
func isOvercommitted(pod *v1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		requestCPU := container.Resources.Requests.Cpu()
		limitCPU := container.Resources.Limits.Cpu()

		if requestCPU != nil && limitCPU != nil {
			if requestCPU.MilliValue() < limitCPU.MilliValue() {
				return true
			}
		}

		requestMem := container.Resources.Requests.Memory()
		limitMem := container.Resources.Limits.Memory()

		if requestMem != nil && limitMem != nil {
			if requestMem.Value() < limitMem.Value() {
				return true
			}
		}
	}

	return false
}

// isDatabasePod 检查是否是数据库 Pod
func isDatabasePod(pod *v1.Pod) bool {
	dbKeywords := []string{"mysql", "postgres", "mongodb", "redis", "elasticsearch", "kafka", "zookeeper"}

	name := strings.ToLower(pod.Name)
	for _, keyword := range dbKeywords {
		if strings.Contains(name, keyword) {
			return true
		}
	}

	// 检查标签
	if app := pod.Labels["app"]; app != "" {
		app = strings.ToLower(app)
		for _, keyword := range dbKeywords {
			if strings.Contains(app, keyword) {
				return true
			}
		}
	}

	return false
}
