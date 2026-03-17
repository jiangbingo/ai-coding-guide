package diagnostic

import (
	"context"
	"fmt"
	"sort"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// FragmentAnalyzer 节点资源碎片分析器
type FragmentAnalyzer struct {
	client kubernetes.Interface
}

// NewFragmentAnalyzer 创建碎片分析器
func NewFragmentAnalyzer(client kubernetes.Interface) *FragmentAnalyzer {
	return &FragmentAnalyzer{
		client: client,
	}
}

// NodeFragment 节点碎片情况
type NodeFragment struct {
	NodeName string

	// CPU 碎片
	CPUCapacity       int64 // MilliValue
	CPUAllocatable    int64
	CPURequested      int64
	CPUAvailable      int64
	CPUFragmentation  float64 // 碎片率
	CPULargestBlock   int64   // 最大可分配块

	// 内存碎片
	MemoryCapacity    int64 // Bytes
	MemoryAllocatable int64
	MemoryRequested   int64
	MemoryAvailable   int64
	MemoryFragmentation float64
	MemoryLargestBlock int64

	// Pod 数量
	PodCapacity       int
	PodAllocated      int
	PodAvailable      int

	// 碎片原因
	FragmentationReasons []string

	// 建议
	Recommendations []string
}

// AnalyzeNodeFragment 分析单个节点碎片
func (fa *FragmentAnalyzer) AnalyzeNodeFragment(ctx context.Context, nodeName string) (*NodeFragment, error) {
	// 获取节点信息
	node, err := fa.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点失败: %w", err)
	}

	// 获取节点上的 Pod
	pods, err := fa.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	fragment := &NodeFragment{
		NodeName: node.Name,
	}

	// 分析 CPU
	fa.analyzeCPUFragment(node, pods, fragment)

	// 分析内存
	fa.analyzeMemoryFragment(node, pods, fragment)

	// 分析 Pod 数量
	fa.analyzePodFragment(node, pods, fragment)

	// 生成建议
	fa.generateRecommendations(fragment)

	return fragment, nil
}

// analyzeCPUFragment 分析 CPU 碎片
func (fa *FragmentAnalyzer) analyzeCPUFragment(node *v1.Node, pods *v1.PodList, fragment *NodeFragment) {
	// 获取 CPU 容量
	fragment.CPUCapacity = node.Status.Capacity.Cpu().MilliValue()
	fragment.CPUAllocatable = node.Status.Allocatable.Cpu().MilliValue()

	// 计算已请求 CPU
	var totalRequested int64
	for _, pod := range pods.Items {
		if pod.Status.Phase != v1.PodRunning && pod.Status.Phase != v1.PodPending {
			continue
		}

		for _, container := range pod.Spec.Containers {
			if request := container.Resources.Requests.Cpu(); request != nil {
				totalRequested += request.MilliValue()
			}
		}
	}

	fragment.CPURequested = totalRequested
	fragment.CPUAvailable = fragment.CPUAllocatable - totalRequested

	// 计算碎片率
	// 碎片率 = 1 - (可用 / 可分配) * (最大块 / 可用)
	// 如果可用为 0，碎片率为 100%
	if fragment.CPUAvailable == 0 {
		fragment.CPUFragmentation = 100
	} else {
		// 简化计算：假设最大块等于可用（实际需要更复杂的 bin packing 算法）
		fragment.CPULargestBlock = fragment.CPUAvailable
		fragment.CPUFragmentation = 0
	}

	// 检查小碎片
	if fragment.CPUAvailable > 0 && fragment.CPUAvailable < 500 {
		fragment.FragmentationReasons = append(fragment.FragmentationReasons,
			fmt.Sprintf("CPU 可用空间较小 (%dm)，可能导致无法调度大 Pod", fragment.CPUAvailable))
	}
}

// analyzeMemoryFragment 分析内存碎片
func (fa *FragmentAnalyzer) analyzeMemoryFragment(node *v1.Node, pods *v1.PodList, fragment *NodeFragment) {
	// 获取内存容量
	fragment.MemoryCapacity = node.Status.Capacity.Memory().Value()
	fragment.MemoryAllocatable = node.Status.Allocatable.Memory().Value()

	// 计算已请求内存
	var totalRequested int64
	for _, pod := range pods.Items {
		if pod.Status.Phase != v1.PodRunning && pod.Status.Phase != v1.PodPending {
			continue
		}

		for _, container := range pod.Spec.Containers {
			if request := container.Resources.Requests.Memory(); request != nil {
				totalRequested += request.Value()
			}
		}
	}

	fragment.MemoryRequested = totalRequested
	fragment.MemoryAvailable = fragment.MemoryAllocatable - totalRequested

	// 计算碎片率
	if fragment.MemoryAvailable == 0 {
		fragment.MemoryFragmentation = 100
	} else {
		fragment.MemoryLargestBlock = fragment.MemoryAvailable
		fragment.MemoryFragmentation = 0
	}

	// 检查小碎片
	if fragment.MemoryAvailable > 0 && fragment.MemoryAvailable < 512*1024*1024 {
		fragment.FragmentationReasons = append(fragment.FragmentationReasons,
			fmt.Sprintf("内存可用空间较小 (%d MiB)，可能导致无法调度大 Pod", fragment.MemoryAvailable/1024/1024))
	}
}

// analyzePodFragment 分析 Pod 数量碎片
func (fa *FragmentAnalyzer) analyzePodFragment(node *v1.Node, pods *v1.PodList, fragment *NodeFragment) {
	// 获取 Pod 容量
	podCapacity := node.Status.Capacity.Pods()
	if podCapacity == nil {
		fragment.PodCapacity = 110 // 默认值
	} else {
		fragment.PodCapacity = int(podCapacity.Value())
	}

	// 计算已分配 Pod
	var allocatedPods int
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodPending {
			allocatedPods++
		}
	}

	fragment.PodAllocated = allocatedPods
	fragment.PodAvailable = fragment.PodCapacity - allocatedPods

	// 检查 Pod 数量限制
	if fragment.PodAvailable < 10 {
		fragment.FragmentationReasons = append(fragment.FragmentationReasons,
			fmt.Sprintf("Pod 可用数量较少 (%d)，可能限制新调度", fragment.PodAvailable))
	}
}

// generateRecommendations 生成建议
func (fa *FragmentAnalyzer) generateRecommendations(fragment *NodeFragment) {
	// CPU 碎片建议
	if fragment.CPUAvailable > 0 && fragment.CPUAvailable < 1000 {
		fragment.Recommendations = append(fragment.Recommendations,
			"CPU 资源碎片化，考虑调度小 Pod 或迁移大 Pod")
	}

	// 内存碎片建议
	if fragment.MemoryAvailable > 0 && fragment.MemoryAvailable < 1024*1024*1024 {
		fragment.Recommendations = append(fragment.Recommendations,
			"内存资源碎片化，考虑调度小 Pod 或迁移大 Pod")
	}

	// 完全空闲
	if fragment.CPUAvailable == fragment.CPUAllocatable && fragment.MemoryAvailable == fragment.MemoryAllocatable {
		fragment.Recommendations = append(fragment.Recommendations,
			"节点完全空闲，考虑迁移或删除")
	}
}

// ClusterFragmentSummary 集群碎片摘要
type ClusterFragmentSummary struct {
	TotalNodes          int
	HighFragmentNodes   []string // 高碎片节点
	EmptyNodes          []string // 空闲节点
	CriticalNodes       []string // 关键节点（资源紧张）

	TotalCPUAvailable   int64
	TotalMemoryAvailable int64

	// 碎片分布
	CPUFragmentDistribution map[string]int // 碎片率范围 -> 节点数量
	MemFragmentDistribution map[string]int

	// 建议
	Recommendations []string
}

// AnalyzeClusterFragment 分析集群碎片
func (fa *FragmentAnalyzer) AnalyzeClusterFragment(ctx context.Context) (*ClusterFragmentSummary, error) {
	// 获取所有节点
	nodes, err := fa.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点列表失败: %w", err)
	}

	summary := &ClusterFragmentSummary{
		TotalNodes:              len(nodes.Items),
		HighFragmentNodes:       make([]string, 0),
		EmptyNodes:              make([]string, 0),
		CriticalNodes:           make([]string, 0),
		CPUFragmentDistribution: make(map[string]int),
		MemFragmentDistribution: make(map[string]int),
		Recommendations:         make([]string, 0),
	}

	for _, node := range nodes.Items {
		fragment, err := fa.AnalyzeNodeFragment(ctx, node.Name)
		if err != nil {
			continue
		}

		// 累计可用资源
		summary.TotalCPUAvailable += fragment.CPUAvailable
		summary.TotalMemoryAvailable += fragment.MemoryAvailable

		// 分类节点
		if fragment.CPUAvailable == fragment.CPUAllocatable &&
			fragment.MemoryAvailable == fragment.MemoryAllocatable {
			summary.EmptyNodes = append(summary.EmptyNodes, node.Name)
		} else if fragment.CPUAvailable < 500 || fragment.MemoryAvailable < 512*1024*1024 {
			summary.CriticalNodes = append(summary.CriticalNodes, node.Name)
		}

		// 碎片分布
		cpuFragRange := getFragmentRange(fragment.CPUFragmentation)
		summary.CPUFragmentDistribution[cpuFragRange]++

		memFragRange := getFragmentRange(fragment.MemoryFragmentation)
		summary.MemFragmentDistribution[memFragRange]++
	}

	// 生成集群级建议
	if len(summary.EmptyNodes) > 2 {
		summary.Recommendations = append(summary.Recommendations,
			fmt.Sprintf("有 %d 个空闲节点，考虑释放资源", len(summary.EmptyNodes)))
	}

	if len(summary.CriticalNodes) > 0 {
		summary.Recommendations = append(summary.Recommendations,
			fmt.Sprintf("有 %d 个资源紧张的节点，考虑扩容或迁移", len(summary.CriticalNodes)))
	}

	return summary, nil
}

// getFragmentRange 获取碎片率范围
func getFragmentRange(fragmentation float64) string {
	switch {
	case fragmentation == 0:
		return "0%"
	case fragmentation < 25:
		return "0-25%"
	case fragmentation < 50:
		return "25-50%"
	case fragmentation < 75:
		return "50-75%"
	default:
		return "75-100%"
	}
}

// ResourceBinPacking 资源装箱分析
type ResourceBinPacking struct {
	client kubernetes.Interface
}

// NewResourceBinPacking 创建装箱分析器
func NewResourceBinPacking(client kubernetes.Interface) *ResourceBinPacking {
	return &ResourceBinPacking{
		client: client,
	}
}

// PackingResult 装箱结果
type PackingResult struct {
	CurrentUtilization float64 // 当前利用率
	OptimalUtilization float64 // 理论最优利用率
	Efficiency         float64 // 装箱效率
	Savings            int     // 可节省节点数
	MigrationPlan      []MigrationItem
}

// MigrationItem 迁移项
type MigrationItem struct {
	PodName      string
	SourceNode   string
	TargetNode   string
	ResourceType string // CPU, Memory
	Reason       string
}

// AnalyzeBinPacking 分析装箱效率
func (rbp *ResourceBinPacking) AnalyzeBinPacking(ctx context.Context) (*PackingResult, error) {
	// 获取所有节点和 Pod
	nodes, err := rbp.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点列表失败: %w", err)
	}

	pods, err := rbp.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	result := &PackingResult{
		MigrationPlan: make([]MigrationItem, 0),
	}

	// 计算当前利用率
	var totalCPU, usedCPU int64
	var totalMem, usedMem int64

	nodeResources := make(map[string]struct {
		cpuAllocatable int64
		cpuUsed        int64
		memAllocatable int64
		memUsed        int64
	})

	for _, node := range nodes.Items {
		cpuAlloc := node.Status.Allocatable.Cpu().MilliValue()
		memAlloc := node.Status.Allocatable.Memory().Value()

		totalCPU += cpuAlloc
		totalMem += memAlloc

		nodeResources[node.Name] = struct {
			cpuAllocatable int64
			cpuUsed        int64
			memAllocatable int64
			memUsed        int64
		}{
			cpuAllocatable: cpuAlloc,
			memAllocatable: memAlloc,
		}
	}

	for _, pod := range pods.Items {
		if pod.Spec.NodeName == "" {
			continue
		}

		var podCPU, podMem int64
		for _, container := range pod.Spec.Containers {
			if request := container.Resources.Requests.Cpu(); request != nil {
				podCPU += request.MilliValue()
			}
			if request := container.Resources.Requests.Memory(); request != nil {
				podMem += request.Value()
			}
		}

		usedCPU += podCPU
		usedMem += podMem

		if res, ok := nodeResources[pod.Spec.NodeName]; ok {
			res.cpuUsed += podCPU
			res.memUsed += podMem
			nodeResources[pod.Spec.NodeName] = res
		}
	}

	// 计算利用率
	if totalCPU > 0 && totalMem > 0 {
		cpuUtil := float64(usedCPU) / float64(totalCPU)
		memUtil := float64(usedMem) / float64(totalMem)
		result.CurrentUtilization = (cpuUtil + memUtil) / 2 * 100
	}

	// 计算理论最优（假设 80% 目标利用率）
	result.OptimalUtilization = 80

	// 计算效率
	if result.OptimalUtilization > 0 {
		result.Efficiency = result.CurrentUtilization / result.OptimalUtilization * 100
	}

	// 计算可节省节点数
	// 假设每个节点平均利用率为 currentUtilization
	// 如果能优化到 optimalUtilization，可以节省的节点数
	if result.CurrentUtilization > 0 {
		// 简化计算
		avgUtil := result.CurrentUtilization
		if avgUtil < 50 {
			// 低利用率，可以节省节点
			potentialSavings := float64(len(nodes.Items)) * (50 - avgUtil) / 100
			result.Savings = int(potentialSavings)
		}
	}

	return result, nil
}

// FindBestFitNode 找到最适合的节点
func (rbp *ResourceBinPacking) FindBestFitNode(ctx context.Context, cpuRequired, memRequired int64) (string, error) {
	// 获取所有节点
	nodes, err := rbp.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("获取节点列表失败: %w", err)
	}

	// 获取所有 Pod
	pods, err := rbp.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	// 计算每个节点的可用资源
	nodeAvailable := make(map[string]struct {
		cpu int64
		mem int64
	})

	for _, node := range nodes.Items {
		cpuAlloc := node.Status.Allocatable.Cpu().MilliValue()
		memAlloc := node.Status.Allocatable.Memory().Value()

		nodeAvailable[node.Name] = struct {
			cpu int64
			mem int64
		}{
			cpu: cpuAlloc,
			mem: memAlloc,
		}
	}

	// 减去已使用的资源
	for _, pod := range pods.Items {
		if pod.Spec.NodeName == "" {
			continue
		}

		var podCPU, podMem int64
		for _, container := range pod.Spec.Containers {
			if request := container.Resources.Requests.Cpu(); request != nil {
				podCPU += request.MilliValue()
			}
			if request := container.Resources.Requests.Memory(); request != nil {
				podMem += request.Value()
			}
		}

		if res, ok := nodeAvailable[pod.Spec.NodeName]; ok {
			res.cpu -= podCPU
			res.mem -= podMem
			nodeAvailable[pod.Spec.NodeName] = res
		}
	}

	// 找到最适合的节点（Best Fit）
	type nodeScore struct {
		name  string
		score float64
	}

	var candidates []nodeScore

	for nodeName, res := range nodeAvailable {
		// 检查是否有足够资源
		if res.cpu >= cpuRequired && res.mem >= memRequired {
			// 计算得分（越接近越好）
			cpuScore := float64(cpuRequired) / float64(res.cpu)
			memScore := float64(memRequired) / float64(res.mem)
			score := (cpuScore + memScore) / 2

			candidates = append(candidates, nodeScore{
				name:  nodeName,
				score: score,
			})
		}
	}

	// 按得分排序（越高越好）
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	if len(candidates) > 0 {
		return candidates[0].name, nil
	}

	return "", fmt.Errorf("没有找到合适的节点")
}
