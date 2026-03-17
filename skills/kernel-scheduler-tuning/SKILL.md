# kernel-scheduler-tuning

> 内核调度参数优化专家 - 针对华为云 ECS / openEuler / 鲲鹏平台

## 触发条件

当用户提到以下关键词时自动激活：

- 调度延迟、CPU 隔离、isolcpus、nohz_full
- NUMA 亲和性、实时性、低延迟
- 内核调度、CFS、调度器参数
- 鲲鹏 920、openEuler、华为云 ECS

## 能力范围

### 1. 诊断能力
- 收集当前调度状态信息
- 分析调度延迟抖动原因
- 识别 CPU 隔离泄漏问题
- 检测 NUMA 亲和性问题

### 2. 优化能力
- 生成内核启动参数配置（GRUB）
- 配置 sysctl 调度器参数
- 设计 cgroups v2 CPU 隔离方案
- IRQ 亲和性优化

### 3. 验证能力
- 隔离效果验证脚本
- 延迟基准测试（cyclictest）
- 负载下稳定性验证

## 工作流程

```
┌─────────────────────────────────────────────────────────────┐
│                    调度优化工作流                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Step 1: 信息收集                                           │
│  ├── CPU 拓扑 (lscpu -e)                                   │
│  ├── NUMA 信息 (numactl -H)                                │
│  ├── 当前隔离状态 (isolcpus, nohz_full)                     │
│  └── 调度器参数 (sysctl kernel.sched_*)                     │
│                                                             │
│  Step 2: 问题分析                                           │
│  ├── 延迟测试 (cyclictest)                                  │
│  ├── 中断分布检查 (/proc/interrupts)                        │
│  └── 进程分布分析 (/proc/*/status)                          │
│                                                             │
│  Step 3: 方案生成                                           │
│  ├── GRUB 配置                                              │
│  ├── sysctl 配置                                            │
│  ├── cgroups 配置                                           │
│  └── IRQ 绑定脚本                                           │
│                                                             │
│  Step 4: 验证效果                                           │
│  ├── 重新运行延迟测试                                       │
│  ├── 对比优化前后指标                                       │
│  └── 生成验证报告                                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 知识库

### CFS 调度器核心参数

| 参数 | 默认值 | 说明 | 调优建议 |
|------|--------|------|---------|
| `sched_min_granularity_ns` | 10ms | 最小调度粒度 | 高并发场景可减小 |
| `sched_wakeup_granularity_ns` | 15ms | 唤醒抢占粒度 | 低延迟场景减小 |
| `sched_latency_ns` | 20ms | 目标调度延迟 | 一般保持默认 |
| `sched_migration_cost_ns` | 5ms | 迁移成本阈值 | NUMA 场景可增大 |

### CPU 隔离层级

```
Level 1: isolcpus      → 基础隔离，调度器不均衡
Level 2: nohz_full     → 关闭时钟 tick
Level 3: rcu_nocbs     → RCU 回调卸载
Level 4: irqaffinity   → 中断亲和性绑定
Level 5: cpusets       → 进程绑定
```

### 鲲鹏 920 特定优化

```bash
# NUMA 拓扑（双路 128 核）
Socket 0: Node 0-1 (每 Node 32 Core)
Socket 1: Node 2-3 (每 Node 32 Core)

# 内存访问延迟 (ns)
        Node0   Node1   Node2   Node3
Node0    130     180     310     310
Node1    180     130     310     310
Node2    310     310     130     180
Node3    310     310     180     130
```

## Prompt 模板

### 诊断 Prompt

```markdown
# 调度性能诊断

## 系统信息
- 平台: {{platform}}  # 如: 鲲鹏 920 / openEuler 22.03
- 内核: {{kernel_version}}
- CPU: {{cpu_info}}
- 问题: {{problem_description}}

请执行以下诊断：
1. 收集 CPU 拓扑和 NUMA 信息
2. 检查当前隔离配置
3. 运行延迟测试
4. 分析瓶颈原因
5. 给出优化建议
```

### 优化 Prompt

```markdown
# 生成调度优化方案

## 需求
- 隔离 CPU: {{isolated_cpus}}  # 如: 2-7,10-15
- 系统 CPU: {{system_cpus}}    # 如: 0-1,8-9
- 目标延迟: {{target_latency}} # 如: < 100us
- 工作负载: {{workload_type}}  # 如: 实时数据处理

请生成：
1. 完整的 GRUB 配置
2. sysctl 调度器参数配置
3. cgroups v2 隔离配置
4. IRQ 亲和性绑定脚本
5. 验证测试脚本
```

## 使用示例

### 示例 1：基础调度优化

```
用户: 我需要优化华为云 ECS 的内核调度参数，环境是鲲鹏 920 + openEuler 22.03，
     当前实时任务有延迟抖动，P99 约 500us，希望降到 100us 以下。

AI: [激活 kernel-scheduler-tuning Skill]

我来帮你诊断和优化调度延迟问题。

## 第一步：信息收集

请运行以下命令收集系统信息：

```bash
# 1. CPU 拓扑
lscpu -e

# 2. NUMA 信息
numactl -H

# 3. 当前隔离状态
cat /sys/devices/system/cpu/isolated
cat /sys/devices/system/cpu/nohz_full
cat /proc/cmdline

# 4. 调度器参数
sysctl -a | grep kernel.sched

# 5. 延迟测试 (需要安装 rt-tests)
cyclictest -m -Sp90 -i200 -h400 -D30
```

请把输出发给我，我会根据结果生成针对性的优化方案。
```

### 示例 2：CPU 隔离配置

```
用户: 帮我配置 CPU 隔离，保留 0-3 给系统，隔离 4-63 给业务。

AI: [激活 kernel-scheduler-tuning Skill]

根据你的需求（鲲鹏 920 双路 64 核），我来生成完整的隔离方案：

## 1. GRUB 配置

编辑 `/etc/default/grub`：

```bash
GRUB_CMDLINE_LINUX_DEFAULT="
  # CPU 隔离
  isolcpus=4-63

  # 完全静态 tick
  nohz_full=4-63
  nohz=on

  # RCU 回调卸载
  rcu_nocbs=4-63
  rcu_nocb_poll=1

  # 关闭 NUMA 自动平衡
  numa_balancing=disable

  # 性能模式
  cpufreq.default_governor=performance

  # 大页配置
  default_hugepagesz=1G
  hugepagesz=1G
  hugepages=64
"
```

应用配置：
```bash
grub2-mkconfig -o /boot/grub2/grub.cfg
reboot
```

## 2. sysctl 配置

创建 `/etc/sysctl.d/99-scheduler-tuning.conf`：

```ini
# CFS 调度器参数
kernel.sched_min_granularity_ns = 10000000
kernel.sched_wakeup_granularity_ns = 15000000
kernel.sched_latency_ns = 20000000
kernel.sched_migration_cost_ns = 5000000

# NUMA 平衡关闭
kernel.numa_balancing = 0

# 实时调度
kernel.sched_rt_runtime_us = -1
```

应用：
```bash
sysctl -p /etc/sysctl.d/99-scheduler-tuning.conf
```

## 3. 验证脚本

```bash
#!/bin/bash
# verify_isolation.sh

echo "隔离状态:"
cat /sys/devices/system/cpu/isolated

echo -e "\n延迟测试 (30秒):"
cyclictest -m -Sp90 -i200 -h400 -D30
```
```

## 输出格式

优化方案包含以下部分：

1. **诊断报告** - 当前状态和问题分析
2. **优化配置** - GRUB、sysctl、cgroups 配置
3. **实施脚本** - 可直接执行的 Shell 脚本
4. **验证方法** - 效果验证和对比
5. **注意事项** - 风险提示和回滚方案

## 注意事项

1. **内核参数修改需要重启** - 提前规划维护窗口
2. **隔离 CPU 数量** - 至少保留 2 个系统 CPU
3. **NUMA 亲和性** - 确保业务进程和内存在同一节点
4. **验证充分** - 生产环境前在测试环境验证
