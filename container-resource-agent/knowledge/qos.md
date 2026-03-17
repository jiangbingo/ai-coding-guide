# Kubernetes QoS 类别详解

## 1. QoS 类别定义

Kubernetes 根据 Pod 的资源配置自动分配 QoS 类别：

| QoS 类别 | 优先级 | 条件 | 特点 |
|----------|--------|------|------|
| **Guaranteed** | 最高 | 所有容器都设置了 requests=limits | 资源保证，OOM 最后被杀 |
| **Burstable** | 中等 | 至少一个容器设置了 requests 或 limits | 弹性资源，可能被驱逐 |
| **BestEffort** | 最低 | 没有任何资源配置 | 使用剩余资源，首先被杀 |

## 2. Guaranteed 类别

### 配置要求

**必须满足**:
1. Pod 中所有容器都必须设置 CPU 和内存的 `requests` 和 `limits`
2. 所有容器的 `requests.cpu == limits.cpu`
3. 所有容器的 `requests.memory == limits.memory`

### 配置示例

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: guaranteed-pod
spec:
  containers:
  - name: app
    image: nginx
    resources:
      requests:
        cpu: "1"        # 必须等于 limits
        memory: "1Gi"   # 必须等于 limits
      limits:
        cpu: "1"        # 必须等于 requests
        memory: "1Gi"   # 必须等于 requests
  - name: sidecar
    image: busybox
    resources:
      requests:
        cpu: "500m"
        memory: "512Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
```

### cgroups 映射

```bash
# cgroups v2 路径
/sys/fs/cgroup/kubepods/guaranteed/pod<uid>/

# CPU 配置
cpu.max = "100000 100000"    # 1 核
cpu.weight = 10000           # 最高权重

# 内存配置
memory.min = "1G"            # 保证最小
memory.max = "1G"            # 硬限制
```

### 适用场景

- 数据库（MySQL、PostgreSQL）
- 消息队列（Kafka、RabbitMQ）
- 缓存服务（Redis、Memcached）
- 关键业务应用
- 对延迟敏感的服务

### 最佳实践

```yaml
# 推荐：关键业务 Guaranteed 配置
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql
spec:
  template:
    spec:
      containers:
      - name: mysql
        image: mysql:8.0
        resources:
          requests:
            cpu: "2"
            memory: "4Gi"
          limits:
            cpu: "2"
            memory: "4Gi"
        env:
        - name: MYSQL_INNODB_BUFFER_POOL_SIZE
          value: "3G"  # 为 OS 和其他预留空间
```

## 3. Burstable 类别

### 配置要求

**满足以下任一条件**:
1. Pod 中至少有一个容器设置了 `requests` 或 `limits`
2. 但不满足 Guaranteed 的所有条件

### 配置示例

```yaml
# 示例 1：只设置 requests
apiVersion: v1
kind: Pod
metadata:
  name: burstable-pod-1
spec:
  containers:
  - name: app
    image: nginx
    resources:
      requests:
        cpu: "500m"
        memory: "512Mi"
      # 没有 limits

# 示例 2：requests < limits
apiVersion: v1
kind: Pod
metadata:
  name: burstable-pod-2
spec:
  containers:
  - name: app
    image: nginx
    resources:
      requests:
        cpu: "500m"
        memory: "512Mi"
      limits:
        cpu: "1"        # > requests
        memory: "1Gi"   # > requests

# 示例 3：只设置 limits（requests 默认等于 limits）
apiVersion: v1
kind: Pod
metadata:
  name: burstable-pod-3
spec:
  containers:
  - name: app
    image: nginx
    resources:
      limits:
        cpu: "1"
        memory: "1Gi"
      # K8s 自动设置 requests = limits，但这仍然是 Burstable
      # 因为其他容器可能没有设置
```

### cgroups 映射

```bash
# cgroups v2 路径
/sys/fs/cgroup/kubepods/burstable/pod<uid>/

# CPU 配置（requests = 500m, limits = 1）
cpu.max = "100000 100000"    # 1 核上限
cpu.weight = 5000            # 基于请求的权重

# 内存配置
memory.min = "512M"          # 请求量
memory.max = "1G"            # 限制量
```

### 适用场景

- Web 前端应用
- API 服务
- 批处理任务
- 开发/测试环境
- 负载波动大的应用

### 最佳实践

```yaml
# 推荐：Web 应用 Burstable 配置
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-api
spec:
  template:
    spec:
      containers:
      - name: api
        image: my-api:v1
        resources:
          requests:
            cpu: "250m"     # 正常负载
            memory: "256Mi"
          limits:
            cpu: "1"        # 峰值可达到 1 核
            memory: "512Mi" # 峰值可达到 512Mi
```

### 超配比建议

| 资源类型 | 开发/测试 | 生产环境 |
|----------|-----------|----------|
| CPU | 3:1 - 5:1 | 2:1 - 3:1 |
| 内存 | 1.5:1 - 2:1 | 1.2:1 - 1.5:1 |

```yaml
# 节点超配比配置（kubelet）
# /var/lib/kubelet/config.yaml
evictionHard:
  memory.available: "500Mi"
  nodefs.available: "10%"
  imagefs.available: "15%"
```

## 4. BestEffort 类别

### 配置要求

Pod 中所有容器都没有设置 `requests` 和 `limits`。

### 配置示例

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: besteffort-pod
spec:
  containers:
  - name: app
    image: nginx
    # 没有任何资源配置
```

### cgroups 映射

```bash
# cgroups v2 路径
/sys/fs/cgroup/kubepods/besteffort/pod<uid>/

# CPU 配置
cpu.max = "max"              # 无限制
cpu.weight = 2               # 最低权重

# 内存配置
memory.min = "0"             # 无保证
memory.max = "max"           # 无限制
```

### 风险

1. **资源饥饿**: 可能获得很少的 CPU 时间
2. **OOM 优先**: 内存不足时首先被 kill
3. **驱逐优先**: 节点压力大时首先被驱逐
4. **性能不稳定**: 受其他 Pod 影响大

### 适用场景

- 批处理任务（非关键）
- 内部工具
- CI/CD 任务
- 临时调试 Pod
- **不建议用于生产环境**

## 5. QoS 与 OOM 关系

### OOM 分数计算

```
OOM 分数 = (进程内存使用 / 总内存) * 1000 + OOM 分数调整值
```

### OOM 分数调整值

| QoS 类别 | OOM 分数调整值 | 行为 |
|----------|---------------|------|
| Guaranteed | -998 | 最后被 kill |
| Burstable | 0 - 999 | 基于内存使用 |
| BestEffort | 1000 | 首先被 kill |

### OOM Kill 顺序

```
1. BestEffort Pod（最高优先级被 kill）
   ↓
2. Burstable Pod（按内存使用排序）
   ↓
3. Guaranteed Pod（最后被 kill）
```

### 配置 OOM 分数

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: custom-oom-pod
spec:
  containers:
  - name: app
    image: nginx
    resources:
      requests:
        memory: "512Mi"
      limits:
        memory: "1Gi"
  # 自定义 OOM 分数调整（需要特性门控）
  oomScoreAdj: -500  # 降低被 kill 优先级
```

## 6. QoS 与资源驱逐

### 驱逐信号

| 信号 | 描述 | 阈值 |
|------|------|------|
| `memory.available` | 可用内存 | 默认 100Mi |
| `nodefs.available` | 节点文件系统可用 | 默认 10% |
| `nodefs.inodesFree` | 节点 inode 可用 | 默认 5% |
| `imagefs.available` | 镜像文件系统可用 | 默认 15% |

### 驱逐顺序

1. **BestEffort Pod** - 首先被驱逐
2. **Burstable Pod** - 按资源使用率排序
3. **Guaranteed Pod** - 最后被驱逐（几乎不会）

### 驱逐配置

```yaml
# kubelet 配置
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
evictionHard:
  memory.available: "500Mi"
  nodefs.available: "10%"
  nodefs.inodesFree: "5%"
  imagefs.available: "15%"
evictionSoft:
  memory.available: "750Mi"
  nodefs.available: "15%"
evictionSoftGracePeriod:
  memory.available: "1m30s"
  nodefs.available: "2m"
evictionMinimumReclaim:
  memory.available: "0Mi"
  nodefs.available: "500Mi"
```

## 7. QoS 选择决策树

```
开始
  │
  ├─ 是否关键业务？── 是 ── Guaranteed
  │     │
  │     └─ 否
  │           │
  │           ├─ 负载是否可预测？── 是 ── Guaranteed
  │           │     │
  │           │     └─ 否
  │           │           │
  │           │           └─ Burstable（设置合理 requests/limits）
  │           │
  │           └─ 是否开发/测试？── 是 ── BestEffort 或 Burstable
  │                 │
  │                 └─ 否 ── Burstable
  │
  └─ 是否批处理任务？── 是 ── Burstable 或 BestEffort
        │
        └─ 否 ── 评估业务重要性
```

## 8. 监控指标

### Prometheus 指标

```promql
# 按 QoS 统计 Pod 数量
count(kube_pod_info) by (qos_class)

# Guaranteed Pod CPU 使用
sum(rate(container_cpu_usage_seconds_total{container!=""}[5m]))
  by (pod) * on(pod) kube_pod_labels{label_qos_class="Guaranteed"}

# Burstable Pod 内存使用率
sum(container_memory_working_set_bytes{container!=""})
  by (pod) / sum(kube_pod_container_resource_limits{resource="memory"})
  by (pod) * on(pod) kube_pod_labels{label_qos_class="Burstable"}

# BestEffort Pod 数量趋势
count(kube_pod_info{qos_class="BestEffort"})
```

### Grafana 查询

```json
{
  "title": "QoS 资源分布",
  "targets": [
    {
      "expr": "sum(kube_pod_container_resource_requests{resource='cpu'}) by (qos_class)",
      "legendFormat": "{{qos_class}} CPU Requests"
    },
    {
      "expr": "sum(kube_pod_container_resource_requests{resource='memory'}) by (qos_class)",
      "legendFormat": "{{qos_class}} Memory Requests"
    }
  ]
}
```

## 9. 常见问题

### Q1: 为什么设置了 limits 还是 Burstable？

**原因**: 只有 limits 没有 requests 时，K8s 会自动设置 requests=limits，但如果 Pod 中有其他容器没有设置，整体仍是 Burstable。

**解决**:
```yaml
# 错误：导致 Burstable
spec:
  containers:
  - name: app
    resources:
      limits: {cpu: "1", memory: "1Gi"}
  - name: sidecar
    # 没有设置资源

# 正确：所有容器都设置
spec:
  containers:
  - name: app
    resources:
      requests: {cpu: "1", memory: "1Gi"}
      limits: {cpu: "1", memory: "1Gi"}
  - name: sidecar
    resources:
      requests: {cpu: "100m", memory: "128Mi"}
      limits: {cpu: "100m", memory: "128Mi"}
```

### Q2: Guaranteed Pod 也会 OOM 吗？

**会**，Guaranteed 只保证:
1. 资源不会低于 requests
2. OOM 优先级最低
3. 驱逐优先级最低

但如果应用内存泄漏或超出 limits，仍会 OOM。

### Q3: 如何强制所有 Pod 都有资源限制？

使用 LimitRange:
```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: default-limits
spec:
  limits:
  - type: Container
    default:        # 默认 limits
      cpu: "500m"
      memory: "512Mi"
    defaultRequest: # 默认 requests
      cpu: "100m"
      memory: "128Mi"
```
