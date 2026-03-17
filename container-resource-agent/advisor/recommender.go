package advisor

import (
	"context"
	"fmt"
	"math"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ResourceRecommender 资源推荐器
type ResourceRecommender struct {
	client        kubernetes.Interface
	metricsClient metricsv1.Interface
}

// NewResourceRecommender 创建资源推荐器
func NewResourceRecommender(client kubernetes.Interface, metricsClient metricsv1.Interface) *ResourceRecommender {
	return &ResourceRecommender{
		client:        client,
		metricsClient: metricsClient,
	}
}

// ResourceRecommendation 资源推荐
type ResourceRecommendation struct {
	Namespace     string
	PodName       string
	ContainerName string

	// 当前配置
	CurrentCPURequest    string
	CurrentCPULimit      string
	CurrentMemoryRequest string
	CurrentMemoryLimit   string

	// 推荐配置
	RecommendedCPURequest    string
	RecommendedCPULimit      string
	RecommendedMemoryRequest string
	RecommendedMemoryLimit   string

	// 实际使用（过去 7 天）
	AvgCPUUsage    string
	PeakCPUUsage   string
	AvgMemoryUsage string
	PeakMemoryUsage string

	// 节省百分比
	CPUSavings    float64
	MemorySavings float64

	// QoS 类别
	CurrentQoS     v1.PodQOSClass
	RecommendedQoS v1.PodQOSClass

	// 原因
	Reasons []string
}

// RecommendPodResources 推荐 Pod 资源配置
func (rr *ResourceRecommender) RecommendPodResources(ctx context.Context, namespace, podName string) (*ResourceRecommendation, error) {
	// 获取 Pod
	pod, err := rr.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 失败: %w", err)
	}

	// 获取 Pod 指标（如果可用）
	var podMetrics *metricsv1.PodMetrics
	if rr.metricsClient != nil {
		podMetrics, _ = rr.metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	}

	recommendation := &ResourceRecommendation{
		Namespace:   namespace,
		PodName:     podName,
		CurrentQoS:  pod.Status.QOSClass,
		Reasons:     make([]string, 0),
	}

	// 分析每个容器
	for i, container := range pod.Spec.Containers {
		rec := rr.recommendContainerResources(container, podMetrics, i)
		if i == 0 {
			// 使用第一个容器的推荐
			recommendation.ContainerName = container.Name
			recommendation.CurrentCPURequest = resourceToString(container.Resources.Requests.Cpu())
			recommendation.CurrentCPULimit = resourceToString(container.Resources.Limits.Cpu())
			recommendation.CurrentMemoryRequest = resourceToString(container.Resources.Requests.Memory())
			recommendation.CurrentMemoryLimit = resourceToString(container.Resources.Limits.Memory())

			recommendation.RecommendedCPURequest = rec.CPURequest
			recommendation.RecommendedCPULimit = rec.CPULimit
			recommendation.RecommendedMemoryRequest = rec.MemoryRequest
			recommendation.RecommendedMemoryLimit = rec.MemoryLimit

			recommendation.AvgCPUUsage = rec.AvgCPUUsage
			recommendation.PeakCPUUsage = rec.PeakCPUUsage
			recommendation.AvgMemoryUsage = rec.AvgMemoryUsage
			recommendation.PeakMemoryUsage = rec.PeakMemoryUsage

			recommendation.Reasons = append(recommendation.Reasons, rec.Reasons...)
		}
	}

	// 计算 QoS 变化
	recommendation.RecommendedQoS = rr.calculateRecommendedQoS(recommendation)

	return recommendation, nil
}

// containerRecommendation 容器推荐
type containerRecommendation struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string

	AvgCPUUsage    string
	PeakCPUUsage   string
	AvgMemoryUsage string
	PeakMemoryUsage string

	Reasons []string
}

// recommendContainerResources 推荐容器资源
func (rr *ResourceRecommender) recommendContainerResources(container v1.Container, podMetrics *metricsv1.PodMetrics, index int) *containerRecommendation {
	rec := &containerRecommendation{
		Reasons: make([]string, 0),
	}

	// 获取当前配置
	currentCPURequest := container.Resources.Requests.Cpu()
	currentCPULimit := container.Resources.Limits.Cpu()
	currentMemRequest := container.Resources.Requests.Memory()
	currentMemLimit := container.Resources.Limits.Memory()

	// 获取实际使用（如果有指标）
	var cpuUsage, memUsage int64
	if podMetrics != nil && index < len(podMetrics.Containers) {
		containerMetric := podMetrics.Containers[index]
		cpuUsage = containerMetric.Usage.Cpu().MilliValue()
		memUsage = containerMetric.Usage.Memory().Value()
	}

	// 推荐 CPU
	rec.CPURequest, rec.CPULimit = rr.recommendCPU(currentCPURequest, currentCPULimit, cpuUsage)

	// 推荐内存
	rec.MemoryRequest, rec.MemoryLimit = rr.recommendMemory(currentMemRequest, currentMemLimit, memUsage)

	// 设置使用统计
	if cpuUsage > 0 {
		rec.AvgCPUUsage = fmt.Sprintf("%dm", cpuUsage)
		rec.PeakCPUUsage = fmt.Sprintf("%dm", cpuUsage) // 简化，实际需要历史数据
	}
	if memUsage > 0 {
		rec.AvgMemoryUsage = fmt.Sprintf("%dMi", memUsage/1024/1024)
		rec.PeakMemoryUsage = fmt.Sprintf("%dMi", memUsage/1024/1024)
	}

	return rec
}

// recommendCPU 推荐 CPU 配置
func (rr *ResourceRecommender) recommendCPU(currentRequest, currentLimit *resource.Quantity, usage int64) (string, string) {
	// 如果没有使用数据，保持原样
	if usage == 0 {
		if currentRequest != nil {
			return currentRequest.String(), resourceToString(currentLimit)
		}
		return "100m", "500m" // 默认值
	}

	// 推荐策略：
	// Request = max(usage * 1.2, 100m)  // 使用量 + 20% 缓冲
	// Limit = max(usage * 2, 500m)      // 使用量 * 2，允许突发

	requestMilli := int64(float64(usage) * 1.2)
	if requestMilli < 100 {
		requestMilli = 100
	}

	limitMilli := int64(float64(usage) * 2)
	if limitMilli < 500 {
		limitMilli = 500
	}

	// 如果当前限制更合理，保持
	if currentLimit != nil && currentLimit.MilliValue() > limitMilli {
		limitMilli = currentLimit.MilliValue()
	}

	return fmt.Sprintf("%dm", requestMilli), fmt.Sprintf("%dm", limitMilli)
}

// recommendMemory 推荐内存配置
func (rr *ResourceRecommender) recommendMemory(currentRequest, currentLimit *resource.Quantity, usage int64) (string, string) {
	// 如果没有使用数据，保持原样
	if usage == 0 {
		if currentRequest != nil {
			return currentRequest.String(), resourceToString(currentLimit)
		}
		return "128Mi", "256Mi" // 默认值
	}

	// 推荐策略：
	// Request = usage * 1.3  // 使用量 + 30% 缓冲
	// Limit = usage * 1.5    // 使用量 + 50% 缓冲

	// 转换为 MiB
	usageMiB := usage / 1024 / 1024

	requestMiB := int64(float64(usageMiB) * 1.3)
	if requestMiB < 64 {
		requestMiB = 64
	}

	limitMiB := int64(float64(usageMiB) * 1.5)
	if limitMiB < 128 {
		limitMiB = 128
	}

	// 向上取整到 64MiB 边界
	requestMiB = roundTo(requestMiB, 64)
	limitMiB = roundTo(limitMiB, 64)

	return fmt.Sprintf("%dMi", requestMiB), fmt.Sprintf("%dMi", limitMiB)
}

// calculateRecommendedQoS 计算推荐的 QoS
func (rr *ResourceRecommender) calculateRecommendedQoS(rec *ResourceRecommendation) v1.PodQOSClass {
	// 如果 Request == Limit，则是 Guaranteed
	if rec.RecommendedCPURequest == rec.RecommendedCPULimit &&
		rec.RecommendedMemoryRequest == rec.RecommendedMemoryLimit {
		return v1.PodQOSGuaranteed
	}

	// 如果有 Request 或 Limit，则是 Burstable
	if rec.RecommendedCPURequest != "" || rec.RecommendedMemoryRequest != "" {
		return v1.PodQOSBurstable
	}

	// 否则是 BestEffort
	return v1.PodQOSBestEffort
}

// NodeReservation 节点资源预留
type NodeReservation struct {
	NodeName string

	// 预留资源
	KubeReservedCPU    string
	KubeReservedMemory string
	SystemReservedCPU  string
	SystemReservedMemory string
	EvictionThresholdMemory string

	// 可分配资源
	AllocatableCPU    string
	AllocatableMemory string

	// 建议
	Recommendations []string
}

// RecommendNodeReservation 推荐节点预留
func (rr *ResourceRecommender) RecommendNodeReservation(ctx context.Context, nodeName string) (*NodeReservation, error) {
	// 获取节点信息
	node, err := rr.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点失败: %w", err)
	}

	reservation := &NodeReservation{
		NodeName:        node.Name,
		Recommendations: make([]string, 0),
	}

	// 获取节点容量
	nodeCPU := node.Status.Capacity.Cpu()
	nodeMem := node.Status.Capacity.Memory()

	// 计算预留资源
	// Kube Reserved: 为 Kubernetes 组件预留
	// System Reserved: 为系统进程预留
	// Eviction Threshold: 驱逐阈值

	cpuMilli := nodeCPU.MilliValue()
	memBytes := nodeMem.Value()

	// 根据节点大小调整预留
	if cpuMilli >= 8000 { // 8 核以上
		reservation.KubeReservedCPU = "500m"
		reservation.KubeReservedMemory = "1Gi"
		reservation.SystemReservedCPU = "500m"
		reservation.SystemReservedMemory = "1Gi"
		reservation.EvictionThresholdMemory = "500Mi"
	} else if cpuMilli >= 4000 { // 4-8 核
		reservation.KubeReservedCPU = "300m"
		reservation.KubeReservedMemory = "512Mi"
		reservation.SystemReservedCPU = "300m"
		reservation.SystemReservedMemory = "512Mi"
		reservation.EvictionThresholdMemory = "250Mi"
	} else { // 4 核以下
		reservation.KubeReservedCPU = "100m"
		reservation.KubeReservedMemory = "256Mi"
		reservation.SystemReservedCPU = "100m"
		reservation.SystemReservedMemory = "256Mi"
		reservation.EvictionThresholdMemory = "100Mi"
	}

	// 计算可分配资源
	kubeCPU, _ := resource.ParseQuantity(reservation.KubeReservedCPU)
	kubeMem, _ := resource.ParseQuantity(reservation.KubeReservedMemory)
	sysCPU, _ := resource.ParseQuantity(reservation.SystemReservedCPU)
	sysMem, _ := resource.ParseQuantity(reservation.SystemReservedMemory)
	evictMem, _ := resource.ParseQuantity(reservation.EvictionThresholdMemory)

	allocatableCPU := cpuMilli - kubeCPU.MilliValue() - sysCPU.MilliValue()
	allocatableMem := memBytes - kubeMem.Value() - sysMem.Value() - evictMem.Value()

	reservation.AllocatableCPU = fmt.Sprintf("%dm", allocatableCPU)
	reservation.AllocatableMemory = fmt.Sprintf("%dMi", allocatableMem/1024/1024)

	// 生成建议
	reservation.Recommendations = append(reservation.Recommendations,
		"在 kubelet 配置中设置以下参数：",
		fmt.Sprintf("  --kube-reserved=cpu=%s,memory=%s", reservation.KubeReservedCPU, reservation.KubeReservedMemory),
		fmt.Sprintf("  --system-reserved=cpu=%s,memory=%s", reservation.SystemReservedCPU, reservation.SystemReservedMemory),
		fmt.Sprintf("  --eviction-hard=memory.available<%s,nodefs.available<10%%", reservation.EvictionThresholdMemory),
	)

	return reservation, nil
}

// QoSRecommendation QoS 推荐
type QoSRecommendation struct {
	Namespace string
	PodName   string

	CurrentQoS v1.PodQOSClass

	// 推荐 QoS
	RecommendedQoS v1.PodQOSClass
	Reason         string

	// 如果需要升级到 Guaranteed
	RequiredCPURequest    string
	RequiredCPULimit      string
	RequiredMemoryRequest string
	RequiredMemoryLimit   string

	// 配置建议
	ConfigSnippet string
}

// RecommendQoS 推荐 QoS 类别
func (rr *ResourceRecommender) RecommendQoS(ctx context.Context, namespace, podName string) (*QoSRecommendation, error) {
	// 获取 Pod
	pod, err := rr.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 失败: %w", err)
	}

	rec := &QoSRecommendation{
		Namespace:  namespace,
		PodName:    podName,
		CurrentQoS: pod.Status.QOSClass,
	}

	// 根据 Pod 标签和用途判断推荐 QoS
	rec.RecommendedQoS, rec.Reason = rr.determineOptimalQoS(pod)

	// 如果需要升级到 Guaranteed，生成配置
	if rec.RecommendedQoS == v1.PodQOSGuaranteed && rec.CurrentQoS != v1.PodQOSGuaranteed {
		rec.RequiredCPURequest = "500m"
		rec.RequiredCPULimit = "500m"
		rec.RequiredMemoryRequest = "512Mi"
		rec.RequiredMemoryLimit = "512Mi"

		rec.ConfigSnippet = fmt.Sprintf(`
resources:
  requests:
    cpu: "%s"
    memory: "%s"
  limits:
    cpu: "%s"
    memory: "%s"
`, rec.RequiredCPURequest, rec.RequiredMemoryRequest, rec.RequiredCPULimit, rec.RequiredMemoryLimit)
	}

	return rec, nil
}

// determineOptimalQoS 确定最优 QoS
func (rr *ResourceRecommender) determineOptimalQoS(pod *v1.Pod) (v1.PodQOSClass, string) {
	// 检查标签
	labels := pod.Labels

	// 关键业务标签
	if labels["critical"] == "true" || labels["priority"] == "high" {
		return v1.PodQOSGuaranteed, "关键业务应使用 Guaranteed QoS"
	}

	// 数据库应用
	if isDatabaseApp(pod) {
		return v1.PodQOSGuaranteed, "数据库应用应使用 Guaranteed QoS"
	}

	// Web 服务
	if labels["app-type"] == "web" || labels["app-type"] == "api" {
		return v1.PodQOSBurstable, "Web/API 服务适合使用 Burstable QoS"
	}

	// 批处理
	if labels["app-type"] == "batch" || labels["app-type"] == "job" {
		if labels["priority"] == "low" {
			return v1.PodQOSBestEffort, "低优先级批处理任务可以使用 BestEffort QoS"
		}
		return v1.PodQOSBurstable, "批处理任务适合使用 Burstable QoS"
	}

	// 默认推荐 Burstable
	return v1.PodQOSBurstable, "默认推荐 Burstable QoS，平衡性能和资源利用"
}

// isDatabaseApp 检查是否是数据库应用
func isDatabaseApp(pod *v1.Pod) bool {
	dbKeywords := []string{"mysql", "postgres", "mongodb", "redis", "elasticsearch", "kafka"}

	name := pod.Name
	for _, keyword := range dbKeywords {
		if containsIgnoreCase(name, keyword) {
			return true
		}
	}

	if app := pod.Labels["app"]; app != "" {
		for _, keyword := range dbKeywords {
			if containsIgnoreCase(app, keyword) {
				return true
			}
		}
	}

	return false
}

// 辅助函数
func resourceToString(q *resource.Quantity) string {
	if q == nil {
		return ""
	}
	return q.String()
}

func roundTo(value, step int64) int64 {
	return ((value + step - 1) / step) * step
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
		 (len(s) > len(substr) &&
		  containsLower(lower(s), lower(substr))))
}

func lower(s string) string {
	return strings.ToLower(s)
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
