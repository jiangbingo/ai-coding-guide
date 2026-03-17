# cgroups v1 vs v2 知识库

## 1. 架构差异

### cgroups v1
```
/sys/fs/cgroup/
├── cpu/                  # CPU 子系统
│   └── cpu.shares
│   └── cpu.cfs_quota_us
│   └── cpu.cfs_period_us
├── memory/               # 内存子系统
│   └── memory.limit_in_bytes
│   └── memory.soft_limit_in_bytes
│   └── memory.oom_control
├── blkio/                # I/O 子系统
│   └── blkio.weight
│   └── blkio.throttle.*
└── cpuset/               # CPU 亲和性
    └── cpuset.cpus
    └── cpuset.mems
```

### cgroups v2
```
/sys/fs/cgroup/
├── cgroup.controllers    # 可用控制器列表
├── cgroup.subtree_control # 子树启用的控制器
├── cpu.weight            # CPU 权重 (1-10000)
├── cpu.max               # CPU 配额 (quota period)
├── memory.max            # 内存硬限制
├── memory.low            # 内存软限制
├── memory.oom.group      # OOM 组控制
├── io.weight             # I/O 权重 (1-10000)
└── io.max                # I/O 限流
```

## 2. CPU 资源控制

### cgroups v1

| 参数 | 说明 | 默认值 | 范围 |
|------|------|--------|------|
| `cpu.shares` | CPU 时间分配权重 | 1024 | 2-262144 |
| `cpu.cfs_quota_us` | CFS 配额（微秒） | -1（无限制） | >= 1000 |
| `cpu.cfs_period_us` | CFS 周期（微秒） | 100000 | 1000-1000000 |

**计算公式**:
```
CPU 核数 = cpu.cfs_quota_us / cpu.cfs_period_us
```

**示例**:
```bash
# 限制为 1.5 核
echo 150000 > cpu.cfs_quota_us
echo 100000 > cpu.cfs_period_us
```

### cgroups v2

| 参数 | 说明 | 默认值 | 范围 |
|------|------|--------|------|
| `cpu.weight` | CPU 权重 | 100 | 1-10000 |
| `cpu.max` | 最大 CPU 时间 | max | quota period |

**权重转换公式**:
```
v2_weight = (v1_shares * 10000) / 1024
```

**示例**:
```bash
# 限制为 1.5 核
echo "150000 100000" > cpu.max

# 设置权重为 512（相当于 v1 的 512 shares）
echo 5000 > cpu.weight
```

### Kubernetes 映射

| K8s 参数 | cgroups v1 | cgroups v2 |
|----------|-----------|-----------|
| `spec.containers[].resources.requests.cpu` | `cpu.shares` (乘以 1024) | `cpu.weight` (归一化) |
| `spec.containers[].resources.limits.cpu` | `cpu.cfs_quota_us` / `cpu.cfs_period_us` | `cpu.max` |

**转换示例**:
```yaml
# Pod 配置
resources:
  requests:
    cpu: "500m"    # 0.5 核
  limits:
    cpu: "1000m"   # 1 核
```

```bash
# cgroups v1
cpu.shares = 512           # 500m * 1024
cpu.cfs_quota_us = 100000  # 1000m * 100000

# cgroups v2
cpu.weight = 50            # 归一化后
cpu.max = "100000 100000"  # 1 核
```

## 3. 内存资源控制

### cgroups v1

| 参数 | 说明 | 行为 |
|------|------|------|
| `memory.limit_in_bytes` | 硬限制 | 超过触发 OOM |
| `memory.soft_limit_in_bytes` | 软限制 | 内存紧张时回收 |
| `memory.memsw.limit_in_bytes` | 内存+Swap 限制 | 需要启用 swap |
| `memory.oom_control` | OOM 控制 | 禁用/启用 OOM Killer |

**关键行为**:
- 超过 `limit_in_bytes` → 立即 OOM Kill
- 超过 `soft_limit` → 内存回收，不 kill
- `oom_control=1` → 暂停进程而非 kill

### cgroups v2

| 参数 | 说明 | 行为 |
|------|------|------|
| `memory.max` | 硬限制 | 超过触发 OOM |
| `memory.min` | 最小保证 | 不会被回收 |
| `memory.low` | 软限制 | 尽量保护 |
| `memory.high` | 节流阈值 | 超过后限速 |
| `memory.swap.max` | Swap 限制 | 0 表示禁用 |
| `memory.oom.group` | OOM 组 | 组内进程一起 kill |

**内存层级保护**:
```
memory.min < memory.low < memory.high < memory.max
   ↓           ↓           ↓            ↓
  保护       尽量保护     节流         OOM
```

**示例配置**:
```bash
# 设置内存限制
echo "2G" > memory.max        # 硬限制 2GB
echo "1G" > memory.high       # 节流阈值 1GB
echo "512M" > memory.low      # 软限制 512MB
echo "256M" > memory.min      # 最小保证 256MB

# 禁用 swap
echo "0" > memory.swap.max
```

### Kubernetes 映射

| K8s 参数 | cgroups v1 | cgroups v2 |
|----------|-----------|-----------|
| `requests.memory` | `memory.soft_limit_in_bytes` | `memory.min` |
| `limits.memory` | `memory.limit_in_bytes` | `memory.max` |

## 4. I/O 资源控制

### cgroups v1 (blkio)

| 参数 | 说明 |
|------|------|
| `blkio.weight` | I/O 权重 (10-1000) |
| `blkio.throttle.read_bps_device` | 读带宽限制 |
| `blkio.throttle.write_bps_device` | 写带宽限制 |
| `blkio.throttle.read_iops_device` | 读 IOPS 限制 |
| `blkio.throttle.write_iops_device` | 写 IOPS 限制 |

**示例**:
```bash
# 设置权重
echo 500 > blkio.weight

# 限制读带宽为 100MB/s
echo "8:0 104857600" > blkio.throttle.read_bps_device
```

### cgroups v2

| 参数 | 说明 | 范围 |
|------|------|------|
| `io.weight` | I/O 权重 | 1-10000 |
| `io.max` | I/O 限流 | 格式: "MAJOR:MINOR op limit" |

**示例**:
```bash
# 设置权重
echo 500 > io.weight

# 限制读带宽
echo "8:0 rbps=104857600 wiops=100" > io.max
```

### Kubernetes 支持

Kubernetes 1.25+ 通过临时约束支持 I/O 限制:
```yaml
# 注意：这是 Alpha 特性
spec:
  containers:
  - name: app
    resources:
      limits:
        hugepages-2Mi: 100Mi
```

## 5. 迁移指南

### 检测当前版本

```bash
# 检查 cgroups 版本
mount | grep cgroup

# v1: 多个挂载点
# v2: 单一挂载点 /sys/fs/cgroup

# 或检查文件
if [ -f /sys/fs/cgroup/cgroup.controllers ]; then
  echo "cgroups v2"
else
  echo "cgroups v1"
fi
```

### 运行时配置

**containerd (config.toml)**:
```toml
[plugins."io.containerd.grpc.v1.cri"]
  # 强制使用 cgroups v2
  disable_cgroup = false
  systemd_cgroup = true  # v2 推荐

[plugins."io.containerd.grpc.v1.cri".containerd]
  snapshotter = "overlayfs"
```

**Docker (daemon.json)**:
```json
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "cgroup-parent": "kubepods"
}
```

### Kubernetes 配置

**kubelet 参数**:
```bash
# cgroups v2
--cgroup-driver=systemd \
--cgroup-root=/ \
--feature-gates=KubeletCgroupDriverFromCRI=true
```

**内核参数 (GRUB)**:
```
GRUB_CMDLINE_LINUX="cgroup_no_v1=all systemd.unified_cgroup_hierarchy=1"
```

### 常见问题

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| `mount: /sys/fs/cgroup: cgroup2 already mounted` | 已启用 v2 | 无需操作 |
| `failed to write to memory.max: invalid argument` | 权限/格式错误 | 检查权限和单位 |
| `cgroup v2 not supported` | 内核版本低 | 升级到 4.5+ |
| `systemd cgroup driver mismatch` | 配置不一致 | 统一使用 systemd |

## 6. 调试命令

### 查看资源使用

```bash
# cgroups v1
cat /sys/fs/cgroup/memory/kubepods/memory.usage_in_bytes
cat /sys/fs/cgroup/cpu/kubepods/cpuacct.usage

# cgroups v2
cat /sys/fs/cgroup/kubepods/memory.current
cat /sys/fs/cgroup/kubepods/cpu.stat
```

### 实时监控

```bash
# 安装 systemd-cgtop
systemd-cgtop

# 输出示例
Control Group           Tasks   %CPU   Memory  Input/s Output/s
/kubepods               150     45.2   4.2G    10M     5M
/kubepods/besteffort    50      10.1   1.5G    2M      1M
/kubepods/burstable     80      30.5   2.0G    6M      3M
/kubepods/guaranteed    20      4.6    700M    2M      1M
```

### 压力测试

```bash
# CPU 压力
stress-ng --cpu 4 --timeout 60s

# 内存压力
stress-ng --vm 2 --vm-bytes 80% --timeout 60s

# I/O 压力
stress-ng --iomix 4 --timeout 60s
```
