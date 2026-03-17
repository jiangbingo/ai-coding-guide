package remediation

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// MultiTenantIsolator 多租户资源隔离器
type MultiTenantIsolator struct {
	client kubernetes.Interface
}

// NewMultiTenantIsolator 创建多租户隔离器
func NewMultiTenantIsolator(client kubernetes.Interface) *MultiTenantIsolator {
	return &MultiTenantIsolator{
		client: client,
	}
}

// TenantIsolationIssue 租户隔离问题
type TenantIsolationIssue struct {
	TenantID     string
	Namespace    string
	IssueType    string // NoQuota, OverQuota, NoLimitRange, SharedNode, NoNetworkPolicy
	Severity     string // Critical, High, Medium, Low
	Description  string
	Remediation  string
}

// DiagnoseTenantIsolation 诊断租户隔离
func (mti *MultiTenantIsolator) DiagnoseTenantIsolation(ctx context.Context, namespace string) ([]*TenantIsolationIssue, error) {
	var issues []*TenantIsolationIssue

	// 1. 检查 ResourceQuota
	quotaIssues := mti.checkResourceQuota(ctx, namespace)
	issues = append(issues, quotaIssues...)

	// 2. 检查 LimitRange
	limitIssues := mti.checkLimitRange(ctx, namespace)
	issues = append(issues, limitIssues...)

	// 3. 检查 NetworkPolicy
	networkIssues := mti.checkNetworkPolicy(ctx, namespace)
	issues = append(issues, networkIssues...)

	// 4. 检查节点隔离
	nodeIssues := mti.checkNodeIsolation(ctx, namespace)
	issues = append(issues, nodeIssues...)

	// 5. 检查 RBAC
	rbacIssues := mti.checkRBAC(ctx, namespace)
	issues = append(issues, rbacIssues...)

	return issues, nil
}

// checkResourceQuota 检查资源配额
func (mti *MultiTenantIsolator) checkResourceQuota(ctx context.Context, namespace string) []*TenantIsolationIssue {
	var issues []*TenantIsolationIssue

	quotas, err := mti.client.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues
	}

	if len(quotas.Items) == 0 {
		issues = append(issues, &TenantIsolationIssue{
			TenantID:    mti.extractTenantID(namespace),
			Namespace:   namespace,
			IssueType:   "NoQuota",
			Severity:    "Critical",
			Description: "命名空间没有配置 ResourceQuota，租户可能消耗过多资源",
			Remediation: `apiVersion: v1
kind: ResourceQuota
metadata:
  name: tenant-quota
  namespace: ` + namespace + `
spec:
  hard:
    pods: "50"
    requests.cpu: "10"
    requests.memory: "20Gi"
    limits.cpu: "20"
    limits.memory: "40Gi"`,
		})
	}

	return issues
}

// checkLimitRange 检查限制范围
func (mti *MultiTenantIsolator) checkLimitRange(ctx context.Context, namespace string) []*TenantIsolationIssue {
	var issues []*TenantIsolationIssue

	limits, err := mti.client.CoreV1().LimitRanges(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues
	}

	if len(limits.Items) == 0 {
		issues = append(issues, &TenantIsolationIssue{
			TenantID:    mti.extractTenantID(namespace),
			Namespace:   namespace,
			IssueType:   "NoLimitRange",
			Severity:    "High",
			Description: "命名空间没有配置 LimitRange，Pod 可能没有资源限制",
			Remediation: `apiVersion: v1
kind: LimitRange
metadata:
  name: tenant-limits
  namespace: ` + namespace + `
spec:
  limits:
  - type: Container
    default:
      cpu: "500m"
      memory: "512Mi"
    defaultRequest:
      cpu: "100m"
      memory: "128Mi"
    max:
      cpu: "4"
      memory: "8Gi"`,
		})
	}

	return issues
}

// checkNetworkPolicy 检查网络策略
func (mti *MultiTenantIsolator) checkNetworkPolicy(ctx context.Context, namespace string) []*TenantIsolationIssue {
	var issues []*TenantIsolationIssue

	policies, err := mti.client.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues
	}

	if len(policies.Items) == 0 {
		issues = append(issues, &TenantIsolationIssue{
			TenantID:    mti.extractTenantID(namespace),
			Namespace:   namespace,
			IssueType:   "NoNetworkPolicy",
			Severity:    "High",
			Description: "命名空间没有配置 NetworkPolicy，租户间网络未隔离",
			Remediation: `# 默认拒绝所有入站流量
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all-ingress
  namespace: ` + namespace + `
spec:
  podSelector: {}
  policyTypes:
  - Ingress

---
# 允许命名空间内部通信
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-within-namespace
  namespace: ` + namespace + `
spec:
  podSelector: {}
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: ` + namespace,
		})
	}

	return issues
}

// checkNodeIsolation 检查节点隔离
func (mti *MultiTenantIsolator) checkNodeIsolation(ctx context.Context, namespace string) []*TenantIsolationIssue {
	var issues []*TenantIsolationIssue

	// 获取命名空间中的 Pod
	pods, err := mti.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues
	}

	// 统计节点使用情况
	nodeTenants := make(map[string]map[string]int) // node -> tenant -> count

	for _, pod := range pods.Items {
		if pod.Spec.NodeName == "" {
			continue
		}

		tenantID := mti.extractTenantID(namespace)
		if nodeTenants[pod.Spec.NodeName] == nil {
			nodeTenants[pod.Spec.NodeName] = make(map[string]int)
		}
		nodeTenants[pod.Spec.NodeName][tenantID]++
	}

	// 检查是否有节点被多个租户共享
	for nodeName, tenants := range nodeTenants {
		if len(tenants) > 1 {
			issues = append(issues, &TenantIsolationIssue{
				TenantID:    mti.extractTenantID(namespace),
				Namespace:   namespace,
				IssueType:   "SharedNode",
				Severity:    "Medium",
				Description: fmt.Sprintf("节点 %s 被多个租户共享", nodeName),
				Remediation: `# 使用节点污点和容忍度实现节点隔离
kubectl taint nodes ` + nodeName + ` tenant=exclusive:NoSchedule

# 或使用 NodeSelector/PodAntiAffinity
spec:
  affinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchLabels:
            tenant: other-tenant
        topologyKey: kubernetes.io/hostname`,
			})
		}
	}

	return issues
}

// checkRBAC 检查 RBAC
func (mti *MultiTenantIsolator) checkRBAC(ctx context.Context, namespace string) []*TenantIsolationIssue {
	var issues []*TenantIsolationIssue

	// 获取命名空间的角色绑定
	roleBindings, err := mti.client.RbacV1().RoleBindings(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues
	}

	// 检查是否有跨命名空间的权限
	clusterRoleBindings, err := mti.client.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues
	}

	// 检查是否有过于宽松的权限
	for _, crb := range clusterRoleBindings.Items {
		for _, subject := range crb.Subjects {
			if subject.Namespace == namespace {
				issues = append(issues, &TenantIsolationIssue{
					TenantID:    mti.extractTenantID(namespace),
					Namespace:   namespace,
					IssueType:   "ClusterRoleBinding",
					Severity:    "Medium",
					Description: fmt.Sprintf("租户用户 %s 绑定了 ClusterRole %s", subject.Name, crb.RoleRef.Name),
					Remediation: "建议使用 Role 和 RoleBinding 限制权限范围",
				})
			}
		}
	}

	// 检查是否缺少 RoleBinding
	if len(roleBindings.Items) == 0 {
		issues = append(issues, &TenantIsolationIssue{
			TenantID:    mti.extractTenantID(namespace),
			Namespace:   namespace,
			IssueType:   "NoRoleBinding",
			Severity:    "Low",
			Description: "命名空间没有配置 RoleBinding，租户用户可能没有访问权限",
			Remediation: `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: tenant-role
  namespace: ` + namespace + `
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: tenant-binding
  namespace: ` + namespace + `
subjects:
- kind: User
  name: tenant-user
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: tenant-role
  apiGroup: rbac.authorization.k8s.io`,
		})
	}

	return issues
}

// ApplyTenantIsolation 应用租户隔离配置
func (mti *MultiTenantIsolator) ApplyTenantIsolation(ctx context.Context, namespace string, config TenantIsolationConfig) error {
	// 1. 创建 ResourceQuota
	if config.EnableQuota {
		quota := &v1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "tenant-quota",
				Namespace: namespace,
			},
			Spec: v1.ResourceQuotaSpec{
				Hard: config.QuotaLimits,
			},
		}
		_, err := mti.client.CoreV1().ResourceQuotas(namespace).Create(ctx, quota, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("创建 ResourceQuota 失败: %w", err)
		}
	}

	// 2. 创建 LimitRange
	if config.EnableLimitRange {
		limitRange := &v1.LimitRange{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "tenant-limits",
				Namespace: namespace,
			},
			Spec: v1.LimitRangeSpec{
				Limits: config.LimitRanges,
			},
		}
		_, err := mti.client.CoreV1().LimitRanges(namespace).Create(ctx, limitRange, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("创建 LimitRange 失败: %w", err)
		}
	}

	// 3. 创建 NetworkPolicy
	if config.EnableNetworkIsolation {
		// 创建默认拒绝策略
		// 创建命名空间内部允许策略
		// ... (需要 networking.k8s.io/v1 包)
	}

	return nil
}

// TenantIsolationConfig 租户隔离配置
type TenantIsolationConfig struct {
	EnableQuota             bool
	EnableLimitRange        bool
	EnableNetworkIsolation  bool
	EnableNodeIsolation     bool

	QuotaLimits             v1.ResourceList
	LimitRanges             []v1.LimitRangeItem
	NetworkPolicies         []string // NetworkPolicy YAML
	NodeSelector            map[string]string
}

// extractTenantID 从命名空间提取租户 ID
func (mti *MultiTenantIsolator) extractTenantID(namespace string) string {
	// 简单实现：直接使用命名空间名
	// 实际可能需要根据标签或命名规则提取
	return namespace
}

// GenerateTenantNamespace 生成租户命名空间配置
func (mti *MultiTenantIsolator) GenerateTenantNamespace(tenantID string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("tenant-%s", tenantID),
			Labels: map[string]string{
				"tenant":                   tenantID,
				"pod-security.kubernetes.io/enforce": "restricted",
				"pod-security.kubernetes.io/audit":   "restricted",
				"pod-security.kubernetes.io/warn":    "restricted",
			},
		},
	}
}
