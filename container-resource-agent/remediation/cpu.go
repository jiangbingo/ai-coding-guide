// Package remediation CPU 密集型任务调度优化
package remediation

import (
	"fmt"
	"time"
)

// CPURemediation CPU 问题修复器
type CPURemediation struct {
	Detector *Detector
}

// CPUIssue CPU 问题类型
type CPUIssue string

const (
	CPUIssueThrottling      CPUIssue = "throttling"       // CPU 限流
	CPUIssueContention      CPUIssue = "contention"       // CPU 争抢
	CPUIssueImbalance       CPUIssue = "imbalance"        // 负载不均衡
	CPUIssueBurstViolation  CPUIssue = "burst_violation"  // 突发超限
	CPUIssueSchedulingDelay CPUIssue = "scheduling_delay" // 调度延迟
)

// CPUDiagnosis CPU 诊断结果
type CPUDiagnosis struct {
	Issue       CPUIssue    `json:"issue"`
	Severity    Severity    `json:"severity"`
	Container   string      `json:"container"`
	Pod         string      `json:"pod"`
	Namespace   string      `json:"namespace"`
	Details     string      `json:"details"`
	Metrics     CPUMetrics  `json:"metrics"`
	Suggestions []string    `json:"suggestions"`
}

// CPUMetrics CPU 指标
type CPUMetrics struct {
	UsageCoreSeconds    float64   `json:"usage_core_seconds"`
	ThrottledTime       float64   `json:"throttled_time"`
	ThrottledPeriods    int64     `json:"throttled_periods"`
	TotalPeriods        int64     `json:"total_periods"`
	ThrottleRatio       float64   `json:"throttle_ratio"`
	SchedulingDelay     float64   `json:"scheduling_delay"`
	CPUUtilization      float64   `json:"cpu_utilization"`
	ThrottlingRate      float64   `json:"throttling_rate"`
	Timestamp           time.Time `json:"timestamp"`
}

// NewCPURemediation 创建 CPU 修复器
func NewCPURemediation(detector *Detector) *CPURemediation {
	return &CPURemediation{
		Detector: detector,
	}
}

// Diagnose 诊断 CPU 问题
func (r *CPURemediation) Diagnose(container, pod, namespace string) (*CPUDiagnosis, error) {
	metrics, err := r.getMetrics(container, pod, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU metrics: %w", err)
	}

	// 检查 CPU 限流
	if metrics.ThrottleRatio > 0.1 {
		return &CPUDiagnosis{
			Issue:     CPUIssueThrottling,
			Severity:  SeverityWarning,
			Container: container,
			Pod:       pod,
			Namespace: namespace,
			Details:   fmt.Sprintf("CPU throttling ratio %.2f%% is too high", metrics.ThrottleRatio*100),
			Metrics:   metrics,
			Suggestions: []string{
				"增加 CPU limit",
				"优化应用减少 CPU 使用",
				"检查是否有无限循环",
				"考虑使用 CPU Manager 静态策略",
			},
		}, nil
	}

	// 检查调度延迟
	if metrics.SchedulingDelay > 100*time.Millisecond {
		return &CPUDiagnosis{
			Issue:     CPUIssueSchedulingDelay,
			Severity:  SeverityWarning,
			Container: container,
			Pod:       pod,
			Namespace: namespace,
			Details:   fmt.Sprintf("Scheduling delay %.2fms is high", float64(metrics.SchedulingDelay)/1e6),
			Metrics:   metrics,
			Suggestions: []string{
				"检查节点负载",
				"使用 CPU pinning",
				"减少容器数量",
				"优化调度策略",
			},
		}, nil
	}

	// 检查 CPU 利用率
	if metrics.CPUUtilization > 0.9 {
		return &CPUDiagnosis{
			Issue:     CPUIssueContention,
			Severity:  SeverityCritical,
			Container: container,
			Pod:       pod,
			Namespace: namespace,
			Details:   fmt.Sprintf("CPU utilization %.2f%% is very high", metrics.CPUUtilization*100),
			Metrics:   metrics,
			Suggestions: []string{
				"扩展副本数",
				"优化算法减少计算量",
				"使用异步处理",
				"考虑 GPU 加速",
			},
		}, nil
	}

	return &CPUDiagnosis{
		Issue:     "",
		Severity:  SeverityNone,
		Container: container,
		Pod:       pod,
		Namespace: namespace,
		Details:   "No CPU issues detected",
		Metrics:   metrics,
	}, nil
}

// Remediate 修复 CPU 问题
func (r *CPURemediation) Remediate(diagnosis *CPUDiagnosis) (*RemediationResult, error) {
	switch diagnosis.Issue {
	case CPUIssueThrottling:
		return r.remediateThrottling(diagnosis)
	case CPUIssueContention:
		return r.remediateContention(diagnosis)
	case CPUIssueSchedulingDelay:
		return r.remediateSchedulingDelay(diagnosis)
	default:
		return &RemediationResult{
			Success: true,
			Message: "No remediation needed",
		}, nil
	}
}

// remediateThrottling 修复 CPU 限流
func (r *CPURemediation) remediateThrottling(diagnosis *CPUDiagnosis) (*RemediationResult, error) {
	// 分析当前配置
	currentLimit := r.getCurrentCPULimit(diagnosis)
	currentRequest := r.getCurrentCPURequest(diagnosis)

	// 推荐新配置
	recommendedLimit := currentLimit * 1.5
	recommendedRequest := currentRequest

	// 如果 request 太低，也需要调整
	if currentRequest < currentLimit*0.5 {
		recommendedRequest = currentLimit * 0.7
	}

	return &RemediationResult{
		Success: true,
		Message: "CPU throttling remediation plan generated",
		Actions: []RemediationAction{
			{
				Type:        "update_resource_limit",
				Description: fmt.Sprintf("Increase CPU limit from %.2f to %.2f cores", currentLimit, recommendedLimit),
				Parameters: map[string]interface{}{
					"container":        diagnosis.Container,
					"current_limit":    currentLimit,
					"recommended_limit": recommendedLimit,
				},
			},
			{
				Type:        "update_resource_request",
				Description: fmt.Sprintf("Adjust CPU request from %.2f to %.2f cores", currentRequest, recommendedRequest),
				Parameters: map[string]interface{}{
					"container":          diagnosis.Container,
					"current_request":    currentRequest,
					"recommended_request": recommendedRequest,
				},
			},
		},
		Manifest: r.generateCPUResourcePatch(diagnosis, recommendedRequest, recommendedLimit),
	}, nil
}

// remediateContention 修复 CPU 争抢
func (r *CPURemediation) remediateContention(diagnosis *CPUDiagnosis) (*RemediationResult, error) {
	return &RemediationResult{
		Success: true,
		Message: "CPU contention remediation plan generated",
		Actions: []RemediationAction{
			{
				Type:        "enable_cpu_manager",
				Description: "Enable CPU Manager static policy for dedicated CPUs",
				Parameters: map[string]interface{}{
					"policy": "static",
				},
			},
			{
				Type:        "set_cpu_affinity",
				Description: "Configure CPU affinity for the workload",
				Parameters: map[string]interface{}{
					"container": diagnosis.Container,
				},
			},
			{
				Type:        "scale_horizontally",
				Description: "Consider horizontal scaling to distribute load",
				Parameters: map[string]interface{}{
					"current_replicas": r.getCurrentReplicas(diagnosis),
				},
			},
		},
		Manifest: r.generateCPUManagerConfig(),
	}, nil
}

// remediateSchedulingDelay 修复调度延迟
func (r *CPURemediation) remediateSchedulingDelay(diagnosis *CPUDiagnosis) (*RemediationResult, error) {
	return &RemediationResult{
		Success: true,
		Message: "Scheduling delay remediation plan generated",
		Actions: []RemediationAction{
			{
				Type:        "enable_cpu_pinning",
				Description: "Enable CPU pinning to reduce scheduling overhead",
				Parameters: map[string]interface{}{
					"container": diagnosis.Container,
				},
			},
			{
				Type:        "adjust_cpu_quota",
				Description: "Optimize CPU quota settings",
				Parameters: map[string]interface{}{
					"container": diagnosis.Container,
				},
			},
			{
				Type:        "check_node_overcommit",
				Description: "Check for node CPU overcommit",
				Parameters: map[string]interface{}{
					"node": r.getNodeForPod(diagnosis),
				},
			},
		},
		Manifest: r.generateCPUPinningConfig(diagnosis),
	}, nil
}

// Helper methods

func (r *CPURemediation) getMetrics(container, pod, namespace string) (CPUMetrics, error) {
	// 从 cgroup 获取 CPU 指标
	// 这里是示例实现
	return CPUMetrics{
		UsageCoreSeconds:    3600,
		ThrottledTime:       180,
		ThrottledPeriods:    1000,
		TotalPeriods:        10000,
		ThrottleRatio:       0.1,
		SchedulingDelay:     50 * time.Millisecond,
		CPUUtilization:      0.85,
		ThrottlingRate:      0.05,
		Timestamp:           time.Now(),
	}, nil
}

func (r *CPURemediation) getCurrentCPULimit(diagnosis *CPUDiagnosis) float64 {
	// 获取当前 CPU limit
	return 2.0 // 示例值
}

func (r *CPURemediation) getCurrentCPURequest(diagnosis *CPUDiagnosis) float64 {
	// 获取当前 CPU request
	return 0.5 // 示例值
}

func (r *CPURemediation) getCurrentReplicas(diagnosis *CPUDiagnosis) int {
	// 获取当前副本数
	return 3
}

func (r *CPURemediation) getNodeForPod(diagnosis *CPUDiagnosis) string {
	// 获取 Pod 所在节点
	return "node-1"
}

func (r *CPURemediation) generateCPUResourcePatch(diagnosis *CPUDiagnosis, request, limit float64) string {
	return fmt.Sprintf(`
# CPU Resource Patch
apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: %s
        resources:
          requests:
            cpu: "%.2f"
          limits:
            cpu: "%.2f"
`, diagnosis.Pod, diagnosis.Namespace, diagnosis.Container, request, limit)
}

func (r *CPURemediation) generateCPUManagerConfig() string {
	return `
# Kubelet CPU Manager Configuration
cpuManagerPolicy: static
cpuManagerReconcilePeriod: 10s
topologyManagerPolicy: best-effort
`
}

func (r *CPURemediation) generateCPUPinningConfig(diagnosis *CPUDiagnosis) string {
	return fmt.Sprintf(`
# CPU Pinning Configuration
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
cpuManagerPolicy: static
reservedSystemCPUs: "0-1"
---
apiVersion: v1
kind: Pod
metadata:
  name: %s
  namespace: %s
spec:
  containers:
  - name: %s
    resources:
      limits:
        cpu: "2"
      requests:
        cpu: "2"
`, diagnosis.Pod, diagnosis.Namespace, diagnosis.Container)
}

// CPUBurstOptimizer CPU 突发优化器
type CPUBurstOptimizer struct {
	BurstWindow    time.Duration
	BurstThreshold float64
}

// NewCPUBurstOptimizer 创建突发优化器
func NewCPUBurstOptimizer(window time.Duration, threshold float64) *CPUBurstOptimizer {
	return &CPUBurstOptimizer{
		BurstWindow:    window,
		BurstThreshold: threshold,
	}
}

// Optimize 优化 CPU 突发配置
func (o *CPUBurstOptimizer) Optimize(diagnosis *CPUDiagnosis) string {
	return fmt.Sprintf(`
# CPU Burst Optimization
# Enable CPU burst for short CPU-intensive tasks
cpu.cfs_burst_us: %d
cpu.cfs_burst_percent: 100

# Original limit
cpu.max: 200000 100000  # 2 cores

# After burst
# Can burst up to 4 cores for %v
`, o.BurstWindow.Microseconds(), o.BurstWindow)
}

// CPUQuotaCalculator CPU 配额计算器
type CPUQuotaCalculator struct{}

// Calculate 计算最优 CPU 配额
func (c *CPUQuotaCalculator) Calculate(requestCores, limitCores float64, isGuaranteed bool) CPURatio {
	quota := limitCores * 100000
	period := int64(100000)

	if isGuaranteed {
		// Guaranteed: quota = limit
		quota = limitCores * 100000
	} else {
		// Burstable: 可能需要更宽松的配额
		quota = limitCores * 100000
	}

	return CPURatio{
		Quota:  int64(quota),
		Period: period,
		Shares: int64(requestCores * 1024),
	}
}

// CPURatio CPU 配额比例
type CPURatio struct {
	Quota  int64 `json:"quota"`
	Period int64 `json:"period"`
	Shares int64 `json:"shares"`
}

// String 返回 cgroup v2 格式
func (r CPURatio) String() string {
	return fmt.Sprintf("%d %d", r.Quota, r.Period)
}
