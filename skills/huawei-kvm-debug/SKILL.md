# huawei-kvm-debug

> 华为云 KVM 内核调试专家 - 针对鲲鹏平台 + openEuler

## 触发条件

- VM Exit 频繁、虚拟化性能差
- virtio 前后端通信问题
- 鲲鹏 920 特有的虚拟化问题
- ARM64 KVM 调试

## 核心诊断脚本

### 1. VM Exit 分析（最关键的性能指标）

```bash
#!/bin/bash
# vmexit-analyzer.sh - 分析 VM Exit 原因

echo "=== VM Exit 分析 ==="

# 检查 perf 是否支持 kvm 访客事件
perf list | grep -q "kvm:" && HAS_KVM_EVENTS=1

if [ "$HAS_KVM_EVENTS" = "1" ]; then
    echo "[使用 perf kvm 追踪]"
    perf kvm --host stat live -p $(pgrep qemu-kvm) --timeout 30
else
    echo "[使用 debugfs 追踪]"
    # ARM64 KVM 没有 vmx/svm，使用 tracefs
    echo 1 > /sys/kernel/debug/tracing/events/kvm/enable
    cat /sys/kernel/debug/tracing/trace_pipe | head -100
    echo 0 > /sys/kernel/debug/tracing/events/kvm/enable
fi

# 分析 VM Exit 分布
echo -e "\n=== VM Exit 热点 ==="
cat /sys/kernel/debug/kvm/vmexit_stats 2>/dev/null || \
    echo "需要内核 CONFIG_KVM_EXIT_TIMING=y"
```

### 2. virtio 性能诊断

```bash
#!/bin/bash
# virtio-perf-diag.sh - virtio 性能诊断

echo "=== virtio 性能诊断 ==="

# 检查 virtio 设备
echo "[1] virtio 设备列表"
ls -la /sys/bus/virtio/devices/

# virtio-net 队列配置
echo -e "\n[2] virtio-net 队列"
for iface in /sys/class/net/vnet*; do
    if [ -d "$iface" ]; then
        name=$(basename $iface)
        echo "  $name:"
        cat /sys/class/net/$name/queues/*/rx_max_pending 2>/dev/null | head -1
    fi
done

# vhost 线程 CPU 使用
echo -e "\n[3] vhost 线程"
ps -eLo pid,tid,comm,%cpu | grep vhost

# virtio 中断分布
echo -e "\n[4] virtio 中断"
grep virtio /proc/interrupts
```

### 3. 鲲鹏 920 虚拟化特性检查

```bash
#!/bin/bash
# kunpeng-virt-check.sh - 鲲鹏虚拟化特性检查

echo "=== 鲲鹏 920 虚拟化特性检查 ==="

# CPU 特性
echo "[1] ARM64 虚拟化扩展"
cat /proc/cpuinfo | grep -E "Features|model name" | head -4

# SMMU (IOMMU) 状态
echo -e "\n[2] SMMU 状态"
ls -la /sys/class/iommu/ 2>/dev/null || echo "SMMU 未启用"

# GIC 版本
echo -e "\n[3] GIC 版本"
cat /sys/firmware/devicetree/base/interrupt-controller*/compatible 2>/dev/null || \
    dmesg | grep -i "GIC" | head -3

# 内存大页
echo -e "\n[4] 大页配置"
cat /proc/meminfo | grep -i huge

# NUMA 拓扑
echo -e "\n[5] NUMA 拓扑"
numactl -H 2>/dev/null || lscpu | grep -i numa
```

### 4. KVM ARM64 调试

```bash
#!/bin/bash
# kvm-arm64-debug.sh - ARM64 KVM 特有调试

echo "=== KVM ARM64 调试 ==="

# Stage-2 页表统计
echo "[1] Stage-2 页表"
cat /sys/kernel/debug/kvm/*/mmu_stats 2>/dev/null || \
    echo "需要挂载 debugfs: mount -t debugfs none /sys/kernel/debug"

# VGIC (虚拟中断控制器) 状态
echo -e "\n[2] VGIC 统计"
cat /sys/kernel/debug/kvm/*/vgic-state 2>/dev/null || \
    echo "VGIC 调试信息不可用"

# 定时器虚拟化
echo -e "\n[3] 虚拟定时器"
dmesg | grep -i "kvm.*timer" | tail -5

# PMU 虚拟化
echo -e "\n[4] PMU 直通状态"
cat /sys/bus/event_source/devices/armv8_pmuv3*/cpumask 2>/dev/null || \
    echo "PMU 直通未启用"
```

## 常见问题诊断流程

### 问题 1：虚拟机网络吞吐量低

```
诊断流程：
1. 检查 virtio-net 队列数
   → 单队列是常见瓶颈

2. 检查 vhost 线程 CPU 亲和性
   → vhost 和 vCPU 在同一 CPU 会竞争

3. 检查网卡多队列
   → 物理网卡队列数要匹配

4. 检查 DPDK/vhost-user
   → 极致性能需要用户态网络
```

### 问题 2：虚拟机延迟抖动

```
诊断流程：
1. 分析 VM Exit 原因
   → 频繁 Exit 导致延迟

2. 检查宿主机调度
   → vCPU 被抢占

3. 检查中断干扰
   → 网卡中断和 vCPU 同 CPU

4. 启用 CPU 隔离
   → isolcpus + cpuset
```

### 问题 3：内存性能差

```
诊断流程：
1. 检查大页使用
   → 4KB 页导致 TLB miss

2. 检查 NUMA 亲和性
   → 跨 NUMA 访问延迟 2-3x

3. 检查 EPT/NPT 压力
   → Stage-2 页表 Miss

4. 检查内存气球
   → 气球占用过多内存
```

## 鲲鹏 920 特有优化

### 1. NUMA 感知 vCPU 绑定

```bash
# 鲲鹏 920 双路 128 核 NUMA 拓扑
# Socket 0: Node 0 (CPU 0-31), Node 1 (CPU 32-63)
# Socket 1: Node 2 (CPU 64-95), Node 3 (CPU 96-127)

# 最佳实践：vCPU 绑定到同一 Node
virsh vcpupin vm-name 0 2-7    # vCPU 0 → Node 0
virsh vcpupin vm-name 1 34-39  # vCPU 1 → Node 1
```

### 2. 64KB 大页（鲲鹏默认）

```bash
# 鲲鹏 920 默认页大小 64KB
# 相比 x86 的 4KB，TLB 覆盖范围更大

# 配置 64KB 大页
echo 1024 > /sys/kernel/mm/hugepages/hugepages-65536kB/nr_hugepages

# libvirt 配置
<memoryBacking>
  <hugepages>
    <page size='64' unit='KiB'/>
  </hugepages>
</memoryBacking>
```

### 3. GICv4 优化

```bash
# 鲲鹏 920 支持 GICv4，支持虚拟中断直接注入
# 减少中断相关的 VM Exit

# 检查 GIC 版本
dmesg | grep "GIC"

# 启用 LPI (Locality-specific Peripheral Interrupts)
echo 1 > /sys/kernel/debug/irq/gicv4_enabled 2>/dev/null
```

## 与华为云工具集成

### 使用华为云 MCP

```bash
# 华为云 MCP 工具（如果有）
# 用于云原生环境的问题诊断

# 示例：查询 ECS 实例信息
huaweicloud ecs show-instance --instance-id i-xxx

# 查询监控指标
huaweicloud cloud-eye get-metric-data \
    --namespace SYS.ECS \
    --metric-name cpu_utilisation
```

## 输出格式

诊断报告包含：

1. **系统概览** - 硬件和虚拟化配置
2. **性能瓶颈** - VM Exit 热点、资源竞争
3. **配置建议** - 具体优化参数
4. **验证脚本** - 如何验证优化效果

## 注意事项

1. **生产环境** - 诊断脚本可能影响性能，谨慎使用
2. **权限要求** - 需要 root 权限访问 debugfs
3. **内核版本** - 某些特性需要 5.10+ 内核
4. **华为 BIOS** - 部分优化需要在 BIOS 中开启

## 相关 Skills

- `kernel-scheduler-tuning` - 宿主机调度优化
- `kvm-virt-optimization` - 通用 KVM 优化
- `cloud-performance-analysis` - 性能分析工具
