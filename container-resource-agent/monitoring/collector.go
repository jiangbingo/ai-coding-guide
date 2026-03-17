package monitoring

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

// MetricsCollector 指标采集器
type MetricsCollector struct {
	client        kubernetes.Interface
	metricsClient metricsv1.Interface
}

// NewMetricsCollector 创建指标采集器
func NewMetricsCollector(client kubernetes.Interface, metricsClient metricsv1.Interface) *MetricsCollector {
	return &MetricsCollector{
		client:        client,
		metricsClient: metricsClient,
	}
}

// CollectNodeMetrics 采集节点指标
func (mc *MetricsCollector) CollectNodeMetrics(ctx context.Context, nodeName string) (*NodeMetrics, error) {
	// 获取节点信息
	node, err := mc.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点失败: %w", err)
	}

	metrics := &NodeMetrics{
		NodeName:  nodeName,
		Timestamp: time.Now(),
	}

	// 采集容量指标
	metrics.CPUCapacity = node.Status.Capacity.Cpu().MilliValue()
	metrics.MemoryCapacity = node.Status.Capacity.Memory().Value()
	metrics.PodCapacity = node.Status.Capacity.Pods().Value()

	// 采集可分配指标
	metrics.CPUAllocatable = node.Status.Allocatable.Cpu().MilliValue()
	metrics.MemoryAllocatable = node.Status.Allocatable.Memory().Value()

	// 采集使用指标（如果有 Metrics Server）
	if mc.metricsClient != nil {
		nodeMetrics, err := mc.metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, nodeName, metav1.GetOptions{})
		if err == nil {
			metrics.CPUUsage = nodeMetrics.Usage.Cpu().MilliValue()
			metrics.MemoryUsage = nodeMetrics.Usage.Memory().Value()
		}
	}

	// 计算使用率
	if metrics.CPUAllocatable > 0 {
		metrics.CPUUtilization = float64(metrics.CPUUsage) / float64(metrics.CPUAllocatable) * 100
	}
	if metrics.MemoryAllocatable > 0 {
		metrics.MemoryUtilization = float64(metrics.MemoryUsage) / float64(metrics.MemoryAllocatable) * 100
	}

	// 采集条件
	for _, condition := range node.Status.Conditions {
		switch condition.Type {
		case v1.NodeReady:
			metrics.Ready = condition.Status == v1.ConditionTrue
		case v1.NodeMemoryPressure:
			metrics.MemoryPressure = condition.Status == v1.ConditionTrue
		case v1.NodeDiskPressure:
			metrics.DiskPressure = condition.Status == v1.ConditionTrue
		case v1.NodePIDPressure:
			metrics.PIDPressure = condition.Status == v1.ConditionTrue
		}
	}

	return metrics, nil
}

// CollectPodMetrics 采集 Pod 指标
func (mc *MetricsCollector) CollectPodMetrics(ctx context.Context, namespace, podName string) (*PodMetrics, error) {
	// 获取 Pod 信息
	pod, err := mc.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 失败: %w", err)
	}

	metrics := &PodMetrics{
		Namespace:  namespace,
		PodName:    podName,
		NodeName:   pod.Spec.NodeName,
		QoSClass:   string(pod.Status.QOSClass),
		Phase:      string(pod.Status.Phase),
		Timestamp:  time.Now(),
	}

	// 采集资源请求和限制
	for _, container := range pod.Spec.Containers {
		if request := container.Resources.Requests.Cpu(); request != nil {
			metrics.CPURequest += request.MilliValue()
		}
		if limit := container.Resources.Limits.Cpu(); limit != nil {
			metrics.CPULimit += limit.MilliValue()
		}
		if request := container.Resources.Requests.Memory(); request != nil {
			metrics.MemoryRequest += request.Value()
		}
		if limit := container.Resources.Limits.Memory(); limit != nil {
			metrics.MemoryLimit += limit.Value()
		}
	}

	// 采集使用指标（如果有 Metrics Server）
	if mc.metricsClient != nil {
		podMetrics, err := mc.metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err == nil {
			for _, container := range podMetrics.Containers {
				metrics.CPUUsage += container.Usage.Cpu().MilliValue()
				metrics.MemoryUsage += container.Usage.Memory().Value()
			}
		}
	}

	// 计算使用率
	if metrics.CPURequest > 0 {
		metrics.CPURequestUtilization = float64(metrics.CPUUsage) / float64(metrics.CPURequest) * 100
	}
	if metrics.CPULimit > 0 {
		metrics.CPULimitUtilization = float64(metrics.CPUUsage) / float64(metrics.CPULimit) * 100
	}
	if metrics.MemoryRequest > 0 {
		metrics.MemoryRequestUtilization = float64(metrics.MemoryUsage) / float64(metrics.MemoryRequest) * 100
	}
	if metrics.MemoryLimit > 0 {
		metrics.MemoryLimitUtilization = float64(metrics.MemoryUsage) / float64(metrics.MemoryLimit) * 100
	}

	return metrics, nil
}

// CollectNamespaceMetrics 采集命名空间指标
func (mc *MetricsCollector) CollectNamespaceMetrics(ctx context.Context, namespace string) (*NamespaceMetrics, error) {
	metrics := &NamespaceMetrics{
		Namespace: namespace,
		Timestamp: time.Now(),
	}

	// 获取命名空间中的 Pod
	pods, err := mc.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	// 统计 Pod 数量
	for _, pod := range pods.Items {
		switch pod.Status.Phase {
		case v1.PodRunning:
			metrics.RunningPods++
		case v1.PodPending:
			metrics.PendingPods++
		case v1.PodFailed:
			metrics.FailedPods++
		case v1.PodSucceeded:
			metrics.SucceededPods++
		}
		metrics.TotalPods++

		// 累计资源
		for _, container := range pod.Spec.Containers {
			if request := container.Resources.Requests.Cpu(); request != nil {
				metrics.TotalCPURequest += request.MilliValue()
			}
			if limit := container.Resources.Limits.Cpu(); limit != nil {
				metrics.TotalCPULimit += limit.MilliValue()
			}
			if request := container.Resources.Requests.Memory(); request != nil {
				metrics.TotalMemoryRequest += request.Value()
			}
			if limit := container.Resources.Limits.Memory(); limit != nil {
				metrics.TotalMemoryLimit += limit.Value()
			}
		}
	}

	// 获取 ResourceQuota
	quotas, err := mc.client.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err == nil && len(quotas.Items) > 0 {
		for _, quota := range quotas.Items {
			if cpu := quota.Status.Hard.Cpu(); cpu != nil {
				metrics.QuotaCPU = cpu.MilliValue()
			}
			if mem := quota.Status.Hard.Memory(); mem != nil {
				metrics.QuotaMemory = mem.Value()
			}
			if cpu := quota.Status.Used.Cpu(); cpu != nil {
				metrics.QuotaCPUUsed = cpu.MilliValue()
			}
			if mem := quota.Status.Used.Memory(); mem != nil {
				metrics.QuotaMemoryUsed = mem.Value()
			}
		}
	}

	return metrics, nil
}

// CollectClusterMetrics 采集集群指标
func (mc *MetricsCollector) CollectClusterMetrics(ctx context.Context) (*ClusterMetrics, error) {
	metrics := &ClusterMetrics{
		Timestamp: time.Now(),
	}

	// 获取所有节点
	nodes, err := mc.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点列表失败: %w", err)
	}

	metrics.TotalNodes = len(nodes.Items)

	for _, node := range nodes.Items {
		// 累计容量
		metrics.TotalCPUCapacity += node.Status.Capacity.Cpu().MilliValue()
		metrics.TotalMemoryCapacity += node.Status.Capacity.Memory().Value()

		// 累计可分配
		metrics.TotalCPUAllocatable += node.Status.Allocatable.Cpu().MilliValue()
		metrics.TotalMemoryAllocatable += node.Status.Allocatable.Memory().Value()

		// 统计节点状态
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
				metrics.ReadyNodes++
			}
		}
	}

	// 获取所有 Pod
	pods, err := mc.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	metrics.TotalPods = len(pods.Items)

	// 统计 QoS 分布
	for _, pod := range pods.Items {
		switch pod.Status.QOSClass {
		case v1.PodQOSGuaranteed:
			metrics.GuaranteedPods++
		case v1.PodQOSBurstable:
			metrics.BurstablePods++
		case v1.PodQOSBestEffort:
			metrics.BestEffortPods++
		}

		// 累计资源请求
		for _, container := range pod.Spec.Containers {
			if request := container.Resources.Requests.Cpu(); request != nil {
				metrics.TotalCPURequest += request.MilliValue()
			}
			if request := container.Resources.Requests.Memory(); request != nil {
				metrics.TotalMemoryRequest += request.Value()
			}
		}
	}

	// 计算利用率
	if metrics.TotalCPUAllocatable > 0 {
		metrics.CPUUtilization = float64(metrics.TotalCPURequest) / float64(metrics.TotalCPUAllocatable) * 100
	}
	if metrics.TotalMemoryAllocatable > 0 {
		metrics.MemoryUtilization = float64(metrics.TotalMemoryRequest) / float64(metrics.TotalMemoryAllocatable) * 100
	}

	return metrics, nil
}

// NodeMetrics 节点指标
type NodeMetrics struct {
	NodeName    string
	Timestamp   time.Time

	// 容量
	CPUCapacity    int64 // MilliValue
	MemoryCapacity int64 // Bytes
	PodCapacity    int64

	// 可分配
	CPUAllocatable    int64
	MemoryAllocatable int64

	// 使用
	CPUUsage    int64
	MemoryUsage int64

	// 使用率
	CPUUtilization    float64
	MemoryUtilization float64

	// 条件
	Ready          bool
	MemoryPressure bool
	DiskPressure   bool
	PIDPressure    bool
}

// PodMetrics Pod 指标
type PodMetrics struct {
	Namespace string
	PodName   string
	NodeName  string
	QoSClass  string
	Phase     string
	Timestamp time.Time

	// 资源配置
	CPURequest    int64
	CPULimit      int64
	MemoryRequest int64
	MemoryLimit   int64

	// 使用
	CPUUsage    int64
	MemoryUsage int64

	// 使用率
	CPURequestUtilization    float64
	CPULimitUtilization      float64
	MemoryRequestUtilization float64
	MemoryLimitUtilization   float64
}

// NamespaceMetrics 命名空间指标
type NamespaceMetrics struct {
	Namespace string
	Timestamp time.Time

	// Pod 数量
	TotalPods      int
	RunningPods    int
	PendingPods    int
	FailedPods     int
	SucceededPods  int

	// 资源请求
	TotalCPURequest    int64
	TotalCPULimit      int64
	TotalMemoryRequest int64
	TotalMemoryLimit   int64

	// 配额
	QuotaCPU       int64
	QuotaMemory    int64
	QuotaCPUUsed   int64
	QuotaMemoryUsed int64
}

// ClusterMetrics 集群指标
type ClusterMetrics struct {
	Timestamp time.Time

	// 节点
	TotalNodes int
	ReadyNodes int

	// Pod
	TotalPods       int
	GuaranteedPods  int
	BurstablePods   int
	BestEffortPods  int

	// 容量
	TotalCPUCapacity    int64
	TotalMemoryCapacity int64

	// 可分配
	TotalCPUAllocatable    int64
	TotalMemoryAllocatable int64

	// 请求
	TotalCPURequest    int64
	TotalMemoryRequest int64

	// 利用率
	CPUUtilization    float64
	MemoryUtilization float64
}
