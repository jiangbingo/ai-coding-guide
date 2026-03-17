# kvm-virt-optimization

> KVM 虚拟化优化专家 - 针对华为云 ECS / 鲲鹏平台

## 触发条件

当用户提到以下关键词时自动激活：

- KVM、QEMU、libvirt、虚拟化
- 虚拟机性能、VM 优化、virtio
- vCPU、内存气球、大页、NUMA
- SR-IOV、DPDK、vhost、VFIO
- 华为云 ECS、鲲鹏虚拟化

## 能力范围

### 1. 架构优化能力
- vCPU 和内存配置优化
- NUMA 拓扑感知调度
- I/O 路径优化（virtio/vhost/SR-IOV）
- 网络和存储性能调优

### 2. 配置生成能力
- libvirt XML 优化配置
- QEMU 命令行参数
- sysctl 内核参数
- cgroups 资源限制

### 3. 故障排查能力
- 虚拟机性能问题诊断
- I/O 延迟分析
- 网络吞吐问题
- 内存 ballooning 问题

## 工作流程

```
┌─────────────────────────────────────────────────────────────┐
│                   KVM 优化工作流                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Step 1: 环境评估                                           │
│  ├── 宿主机硬件配置 (CPU/内存/存储/网络)                    │
│  ├── NUMA 拓扑分析                                          │
│  ├── 当前虚拟机配置                                         │
│  └── 工作负载特征                                           │
│                                                             │
│  Step 2: 瓶颈识别                                           │
│  ├── CPU 性能分析                                           │
│  ├── 内存带宽和延迟                                         │
│  ├── 网络 I/O 吞吐/延迟                                     │
│  └── 存储 I/O 性能                                          │
│                                                             │
│  Step 3: 优化方案                                           │
│  ├── vCPU 绑定和调度                                        │
│  ├── 内存大页和 NUMA 绑定                                   │
│  ├── I/O 路径选择 (virtio/vhost/SR-IOV)                    │
│  └── 中断和 vhost 线程优化                                  │
│                                                             │
│  Step 4: 配置实施                                           │
│  ├── 生成优化后的 XML 配置                                  │
│  ├── 系统参数调整                                           │
│  ├── 应用并验证                                             │
│  └── 性能基准测试                                           │
│                                                             │
│  Step 5: 持续监控                                           │
│  ├── 性能指标监控                                           │
│  ├── 异常检测                                               │
│  └── 动态调整建议                                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 知识库

### KVM 性能优化层级

```
┌─────────────────────────────────────────────────────────────┐
│                   优化层级金字塔                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Level 5: 应用层优化                                        │
│  ├── 客户机内核参数                                         │
│  └── 应用程序配置                                           │
│                                                             │
│  Level 4: I/O 路径优化                                      │
│  ├── virtio 多队列                                          │
│  ├── vhost 线程绑定                                         │
│  └── SR-IOV / DPDK                                          │
│                                                             │
│  Level 3: 内存优化                                          │
│  ├── 大页 (HugePages)                                       │
│  ├── NUMA 绑定                                              │
│  └── KSM / 内存合并                                         │
│                                                             │
│  Level 2: vCPU 优化                                         │
│  ├── CPU 绑定 (pinning)                                     │
│  ├── CPU 亲和性                                             │
│  └── 实时调度                                               │
│                                                             │
│  Level 1: 硬件层优化                                        │
│  ├── NUMA 拓扑感知                                          │
│  ├── 中断亲和性                                             │
│  └── 电源管理                                               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 鲲鹏 920 虚拟化特性

```
硬件虚拟化支持:
- ARMv8.2-A Virtualization Extensions
- Stage-2 页表
- 虚拟定时器
- GICv4 (虚拟中断)

性能特性:
- 64 vCPU per VM
- 512 GB 内存 per VM
- SMMUv3 (IOMMU)
- CCIX 支持

推荐配置:
- vCPU 绑定到同一 NUMA node
- 使用 64KB 大页
- virtio-net 多队列 (2-8 队列)
- vhost 线程独立 CPU
```

### virtio vs vhost vs SR-IOV

| 方案 | 延迟 | 吞吐 | CPU 开销 | 灵活性 | 适用场景 |
|------|------|------|----------|--------|----------|
| **virtio** | 高 | 低 | 高 | 最高 | 开发/测试 |
| **vhost** | 中 | 中 | 中 | 高 | 通用生产 |
| **vhost-user** | 低 | 高 | 低 | 中 | 高性能网络 |
| **SR-IOV** | 最低 | 最高 | 最低 | 低 | 极致性能 |
| **DPDK** | 最低 | 最高 | 可配置 | 低 | NFV/数据面 |

### libvirt XML 关键配置

```xml
<!-- vCPU 绑定 -->
<vcpu placement='static' cpuset='2-7,10-15'>12</vcpu>
<cputune>
  <vcpupin vcpu='0' cpuset='2'/>
  <vcpupin vcpu='1' cpuset='3'/>
  <emulatorpin cpuset='0-1'/>
</cputune>

<!-- NUMA 绑定 -->
<numatune>
  <memory mode='strict' nodeset='0'/>
  <memnode cellid='0' mode='strict' nodeset='0'/>
</numatune>

<!-- 大页 -->
<memoryBacking>
  <hugepages>
    <page size='64' unit='KiB' nodeset='0'/>
  </hugepages>
</memoryBacking>

<!-- virtio-net 多队列 -->
<interface type='bridge'>
  <model type='virtio'/>
  <driver name='vhost' queues='4' rx_queue_size='1024' tx_queue_size='1024'/>
</interface>

<!-- virtio-blk 多队列 -->
<disk type='block' device='disk'>
  <driver name='qemu' type='raw' cache='none' io='native' queues='4'/>
</disk>
```

### 性能监控命令

```bash
# 宿主机 CPU 使用
mpstat -P ALL 1

# 虚拟机 vCPU 使用
virsh cpu-stats <vm-name> 1 5

# 内存统计
virsh memtune <vm-name>
cat /proc/meminfo | grep Huge

# 网络 I/O
virsh domifstat <vm-name> vnet0

# 磁盘 I/O
virsh domblkstat <vm-name> vda

# NUMA 统计
numastat -p $(pgrep qemu-kvm)
```

## Prompt 模板

### 性能优化 Prompt

```markdown
# KVM 虚拟机性能优化

## 环境信息
- 宿主机: {{host_info}}  # 鲲鹏 920 / x86_64
- 内存: {{memory}}
- 存储: {{storage_type}}  # NVMe / SSD / HDD
- 网络: {{network_type}}  # 10G / 25G / 100G

## 虚拟机配置
- vCPU: {{vcpu_count}}
- 内存: {{vm_memory}}
- 磁盘: {{disk_config}}
- 网络: {{network_config}}

## 工作负载
- 类型: {{workload_type}}  # Web服务 / 数据库 / NFV
- 指标: {{target_metrics}}  # 延迟 < 100us / 吞吐 > 10Gbps

请生成：
1. 完整的优化配置
2. libvirt XML 关键部分
3. 宿主机内核参数
4. 性能验证方法
```

### 故障诊断 Prompt

```markdown
# KVM 性能问题诊断

## 问题描述
- 症状: {{symptom}}  # 延迟高 / 吞吐低 / CPU 高
- 触发条件: {{trigger}}
- 持续时间: {{duration}}

## 环境数据
{{diagnostic_data}}

请分析：
1. 问题根因
2. 瓶颈位置
3. 优化建议
4. 预期效果
```

## 使用示例

### 示例 1：高性能 NFV 虚拟机配置

```
用户: 我需要在鲲鹏 920 上部署一个 NFV 虚拟机，要求网络吞吐 20Gbps+，延迟 < 50us。

AI: [激活 kvm-virt-optimization]

针对 NFV 高性能场景，我来生成完整的优化配置。

## 1. 宿主机准备

### 内核参数 (/etc/sysctl.d/99-nfv.conf)
```bash
# CPU 隔离（假设 64 核，隔离 4-63 给 VM）
isolcpus=4-63 nohz_full=4-63 rcu_nocbs=4-63

# 网络优化
net.core.somaxconn = 65535
net.core.netdev_max_backlog = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 15

# 大页配置
vm.nr_hugepages = 32768  # 64KB * 32768 = 2GB
```

### 中断亲和性（移到系统 CPU）
```bash
# 将网卡中断绑定到 CPU 0-3
for irq in $(grep eth0 /proc/interrupts | cut -d: -f1); do
    echo 0f > /proc/irq/$irq/smp_affinity
done
```

## 2. libvirt XML 配置

```xml
<domain type='kvm'>
  <name>nfv-vm</name>
  <memory unit='GiB'>8</memory>
  <currentMemory unit='GiB'>8</currentMemory>
  <memoryBacking>
    <hugepages>
      <page size='64' unit='KiB'/>
    </hugepages>
    <locked/>
  </memoryBacking>

  <vcpu placement='static'>16</vcpu>
  <cputune>
    <!-- vCPU 绑定到隔离 CPU -->
    <vcpupin vcpu='0' cpuset='4'/>
    <vcpupin vcpu='1' cpuset='5'/>
    <vcpupin vcpu='2' cpuset='6'/>
    <vcpupin vcpu='3' cpuset='7'/>
    <vcpupin vcpu='4' cpuset='8'/>
    <vcpupin vcpu='5' cpuset='9'/>
    <vcpupin vcpu='6' cpuset='10'/>
    <vcpupin vcpu='7' cpuset='11'/>
    <vcpupin vcpu='8' cpuset='12'/>
    <vcpupin vcpu='9' cpuset='13'/>
    <vcpupin vcpu='10' cpuset='14'/>
    <vcpupin vcpu='11' cpuset='15'/>
    <vcpupin vcpu='12' cpuset='16'/>
    <vcpupin vcpu='13' cpuset='17'/>
    <vcpupin vcpu='14' cpuset='18'/>
    <vcpupin vcpu='15' cpuset='19'/>
    <!-- vhost 线程绑定 -->
    <emulatorpin cpuset='2-3'/>
  </cputune>

  <numatune>
    <memory mode='strict' nodeset='0'/>
  </numatune>

  <cpu mode='host-passthrough' check='none'>
    <topology sockets='1' cores='16' threads='1'/>
    <numa>
      <cell id='0' cpus='0-15' memory='8' unit='GiB' memAccess='shared'/>
    </numa>
  </cpu>

  <devices>
    <!-- 高性能网卡：多队列 vhost -->
    <interface type='bridge'>
      <source bridge='br0'/>
      <model type='virtio'/>
      <driver name='vhost' queues='8'
              rx_queue_size='4096' tx_queue_size='4096'>
        <host csum='off' gso='off' tso4='off' tso6='off' ecn='off' ufo='off'/>
        <guest csum='off' gso='off' tso4='off' tso6='off' ecn='off' ufo='off'/>
      </driver>
    </interface>

    <!-- 高性能存储 -->
    <disk type='block' device='disk'>
      <driver name='qemu' type='raw' cache='none' io='native'
              queues='8' discard='unmap'/>
      <source dev='/dev/nvme0n1'/>
      <target dev='vda' bus='virtio'/>
    </disk>
  </devices>
</domain>
```

## 3. 客户机内核参数

```bash
# /etc/sysctl.conf in guest
net.core.rps_sock_flow_entries = 0  # 禁用 RPS，让 vhost 做分发
net.ipv4.tcp_low_latency = 1
```

## 4. 性能验证

```bash
# 网络吞吐测试
iperf3 -c <server> -t 60 -P 8

# 网络延迟测试
./l2fwd -c 0xff -n 4 -- -p 0x3  # DPDK L2 转发

# CPU 效率
perf stat -e cycles,instructions -p $(pgrep qemu-kvm) sleep 10
```

预期性能：
- 吞吐: 25+ Gbps (单流 10G, 多流叠加)
- 延迟: 30-50 us (L2 转发)
- CPU 效率: > 80% 在数据面
```

### 示例 2：存储 I/O 优化

```
用户: 我的 KVM 虚拟机磁盘 I/O 很慢，如何优化？

AI: [激活 kvm-virt-optimization]

## 存储 I/O 优化诊断

### 第一步：评估当前配置

```bash
# 1. 查看当前磁盘配置
virsh dumpxml <vm> | grep -A 10 "<disk"

# 2. 宿主机存储性能基线
fio --name=host-baseline --filename=/dev/nvme0n1 \
    --ioengine=libaio --direct=1 --bs=4k \
    --rw=randread --iodepth=64 --numjobs=4 --runtime=30

# 3. 虚拟机内存储性能
fio --name=vm-test --filename=/dev/vda \
    --ioengine=libaio --direct=1 --bs=4k \
    --rw=randread --iodepth=64 --numjobs=4 --runtime=30
```

### 第二步：优化方案

#### 1. virtio-blk 多队列配置
```xml
<disk type='block' device='disk'>
  <driver name='qemu' type='raw'
          cache='none'      <!-- 无缓存，O_DIRECT -->
          io='native'       <!-- Linux AIO -->
          queues='8'        <!-- 多队列 -->
          discard='unmap'/>
  <source dev='/dev/nvme0n1'/>
  <target dev='vda' bus='virtio'/>
</disk>
```

#### 2. vhost-user-blk（极致性能）
```xml
<disk type='vhostuser' device='disk'>
  <driver type='raw'/>
  <source type='unix' path='/tmp/vhost.0'/>
  <target dev='vda' bus='virtio'/>
</disk>
```

启动 vhost-user-blk 后端：
```bash
spdk_tgt &
rpc.py bdev_nvme_attach_controller -b Nvme0 -t PCIe -a 0000:01:00.0
rpc.py vhost_create_blk_controller vhost.0 Nvme0n1
```

#### 3. 客户机内核参数
```bash
# 增加virtio队列深度
echo 1024 > /sys/block/vda/queue/nr_requests

# 使用 mq 调度器
echo mq-deadline > /sys/block/vda/queue/scheduler
```

### 优化效果对比

| 配置 | 4K 随机读 IOPS | 4K 随机写 IOPS | 延迟 |
|------|----------------|----------------|------|
| 默认 virtio | 150K | 120K | 40us |
| 多队列 virtio | 400K | 350K | 15us |
| vhost-user-blk | 800K+ | 700K+ | 5us |

### 第三方工具推荐
- **SPDK**: 用户态存储栈，最高性能
- **NVMe over Fabrics**: 分布式存储
- **virtiofs**: 高性能文件系统共享
```

### 示例 3：NUMA 亲和性优化

```
用户: 我的鲲鹏 920 双路服务器上跑 KVM，虚拟机性能不稳定。

AI: [激活 kvm-virt-optimization]

鲲鹏 920 双路有 4 个 NUMA node，NUMA 亲和性对性能影响巨大。

## NUMA 拓扑分析

```bash
# 查看 NUMA 拓扑
numactl -H

# 预期输出（鲲鹏 920 双路）:
# available: 4 nodes (0-3)
# node 0 size: 65536 MB
# node 1 size: 65536 MB
# node 2 size: 65536 MB
# node 3 size: 65536 MB
#
# node distances:
# node   0   1   2   3
#   0:  10  20  40  40
#   1:  20  10  40  40
#   2:  40  40  10  20
#   3:  40  40  20  10
```

## 优化方案

### 1. 单 NUMA 虚拟机（小规格）
```xml
<!-- 所有资源绑定到 Node 0 -->
<numatune>
  <memory mode='strict' nodeset='0'/>
</numatune>
<cputune>
  <vcpupin vcpu='0' cpuset='0'/>
  <vcpupin vcpu='1' cpuset='1'/>
  <vcpupin vcpu='2' cpuset='2'/>
  <vcpupin vcpu='3' cpuset='3'/>
</cputune>
```

### 2. 跨 NUMA 虚拟机（大规格）
```xml
<!-- 使用 2 个 NUMA node -->
<numatune>
  <memory mode='interleave' nodeset='0,1'/>
</numatune>
<cpu>
  <numa>
    <cell id='0' cpus='0-7' memory='4' unit='GiB' nodeset='0'/>
    <cell id='1' cpus='8-15' memory='4' unit='GiB' nodeset='1'/>
  </numa>
</cpu>
```

### 3. QEMU 进程 NUMA 绑定
```bash
# 启动后绑定 QEMU 进程
pid=$(pgrep -f "qemu.*vm-name")
numactl --cpunodebind=0 --membind=0 --pid=$pid
```

### 验证 NUMA 效果

```bash
# 检查 QEMU 进程 NUMA 统计
numastat -p $(pgrep qemu-kvm)

# 期望: 绝大部分内存在同一 node
# Node 0  Node 1  Node 2  Node 3
# 8192    0       0       0       <-- 好
# 2048    2048    2048    2048    <-- 差，跨 NUMA
```

### 性能影响

| 场景 | 本地 NUMA | 跨 NUMA | 性能损失 |
|------|-----------|---------|----------|
| 内存访问 | ~130ns | ~310ns | 138% |
| 内存带宽 | 100% | 60% | 40% |
| 应用延迟 | 1x | 1.5-2x | 50-100% |
```

## 输出格式

优化报告包含：

1. **当前状态评估** - 配置和性能基线
2. **瓶颈识别** - 主要性能限制因素
3. **优化方案** - 分层优化建议
4. **配置文件** - 可直接使用的 XML/sysctl
5. **验证方法** - 如何验证优化效果
6. **监控建议** - 持续监控指标

## 注意事项

1. **生产变更** - 在维护窗口进行，提前备份配置
2. **逐步优化** - 一次改一个参数，验证效果
3. **性能基线** - 优化前建立基线，便于对比
4. **资源预留** - 给宿主机预留足够资源
5. **监控告警** - 设置性能退化告警

## 相关 Skills

- `kernel-scheduler-tuning` - 调度器参数优化
- `cloud-performance-analysis` - 性能分析
- `kernel-debug-assistant` - 内核调试
- `smoke-test-ops` - 虚拟化冒烟测试
