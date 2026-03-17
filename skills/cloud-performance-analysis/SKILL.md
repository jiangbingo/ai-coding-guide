# cloud-performance-analysis

> 云性能分析专家 - 基于 perf/eBPF/BCC 的性能诊断与优化

## 触发条件

当用户提到以下关键词时自动激活：

- 性能分析、性能瓶颈、火焰图、hotspot
- perf、eBPF、BCC、bpftrace
- CPU 占用高、延迟高、吞吐量低
- 系统调优、性能优化、压测分析
- 鲲鹏、openEuler、华为云 ECS

## 能力范围

### 1. 性能采集能力
- perf 事件采集和分析
- eBPF 动态追踪
- 火焰图生成和解读
- 系统调用追踪

### 2. 分析诊断能力
- 识别 CPU 热点函数
- 分析内存访问模式
- 诊断 I/O 瓶颈
- 锁竞争分析

### 3. 优化建议能力
- 生成优化方案
- 代码层面建议
- 系统配置调优
- 性能回归预防

## 工作流程

```
┌─────────────────────────────────────────────────────────────┐
│                    性能分析工作流                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Step 1: 问题定义                                           │
│  ├── 明确性能指标（延迟/吞吐/资源利用率）                    │
│  ├── 确定基线和目标                                         │
│  └── 了解系统架构和约束                                     │
│                                                             │
│  Step 2: 数据采集                                           │
│  ├── CPU 分析 (perf record/report)                         │
│  ├── 内存分析 (perf mem, page-faults)                      │
│  ├── I/O 分析 (iostat, blktrace)                           │
│  └── 网络分析 (tcpdump, eBPF)                              │
│                                                             │
│  Step 3: 火焰图分析                                         │
│  ├── CPU 火焰图生成                                         │
│  ├── 识别热点调用栈                                         │
│  ├── 分析函数占比                                           │
│  └── 定位优化目标                                           │
│                                                             │
│  Step 4: 深度诊断                                           │
│  ├── eBPF 追踪关键路径                                      │
│  ├── 分析锁等待时间                                         │
│  ├── 检查缓存命中率                                         │
│  └── NUMA 访问模式分析                                      │
│                                                             │
│  Step 5: 优化建议                                           │
│  ├── 代码优化建议                                           │
│  ├── 编译器优化选项                                         │
│  ├── 系统配置调整                                           │
│  └── 架构优化方案                                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 知识库

### Perf 常用命令

| 命令 | 用途 | 示例 |
|------|------|------|
| `perf record` | 采样记录 | `perf record -g -p <pid> -- sleep 30` |
| `perf report` | 分析报告 | `perf report --stdio` |
| `perf top` | 实时热点 | `perf top -p <pid>` |
| `perf stat` | 统计事件 | `perf stat -e cycles,instructions ./app` |
| `perf mem` | 内存分析 | `perf mem record -p <pid>` |
| `perf lock` | 锁分析 | `perf lock record -p <pid>` |

### eBPF/BCC 工具

| 工具 | 用途 | 场景 |
|------|------|------|
| `execsnoop` | 追踪新进程 | 查看短生命周期进程 |
| `opensnoop` | 追踪文件打开 | I/O 问题诊断 |
| `biolatency` | 块 I/O 延迟 | 存储性能分析 |
| `tcpconnect` | TCP 连接追踪 | 网络问题诊断 |
| `profile` | CPU 采样 | 生成火焰图数据 |
| `offcputime` | Off-CPU 分析 | 等待时间分析 |
| `llcstat` | 缓存统计 | NUMA/缓存优化 |

### 鲲鹏 920 特性

```
架构: ARMv8.2-A
核心: 64 核 / 双路 128 核
频率: 2.6 GHz
L1 Cache: 64KB I-Cache + 64KB D-Cache
L2 Cache: 512KB (per core pair)
L3 Cache: 共享，每 DIE 32MB

NUMA 拓扑:
- Socket 0: Node 0, Node 1
- Socket 1: Node 2, Node 3
- 跨 Socket 延迟: ~310ns
- 本地内存延迟: ~130ns
```

### 性能反模式识别

| 反模式 | 症状 | 诊断方法 |
|--------|------|----------|
| **热锁竞争** | 高 sys CPU，低用户 CPU | `perf lock`, `offcputime` |
| **缓存抖动** | 多核性能随核心数下降 | `perf c2c`, `llcstat` |
| **内存带宽饱和** | 吞吐量不随核心增加 | `perf stat -e bus-cycles` |
| **False Sharing** | 多线程性能差 | 代码审查 + `perf c2c` |
| **系统调用风暴** | 高内核态 CPU | `perf record -e syscalls:sys_enter_*` |
| **I/O 等待** | 高 iowait | `iostat`, `biolatency` |

## Prompt 模板

### 性能诊断 Prompt

```markdown
# 性能问题诊断

## 系统信息
- 平台: {{platform}}  # 鲲鹏 920 / x86_64
- OS: {{os_version}}  # openEuler 22.03
- 内核: {{kernel_version}}
- 应用: {{application}}

## 问题描述
- 症状: {{symptom}}  # 如：延迟高、吞吐低
- 基线: {{baseline}}
- 目标: {{target}}

## 已有数据
{{perf_data}}  # perf report / 火焰图

请分析：
1. 识别主要瓶颈
2. 定位热点代码
3. 给出优化建议
4. 预估优化效果
```

### 火焰图分析 Prompt

```markdown
# 火焰图分析

## 火焰图数据
{{flamegraph_svg_or_text}}

## 分析重点
- 关注区域: {{focus_area}}  # 如：网络 I/O、内存分配
- 排除区域: {{exclude_area}}

请：
1. 识别 Top 5 热点调用栈
2. 分析每个热点的优化潜力
3. 检测是否存在异常平坦区域
4. 给出具体优化建议
```

## 使用示例

### 示例 1：CPU 热点分析

```
用户: 我的 KVM 虚拟机宿主机 CPU 使用率很高，帮我分析下热点。

AI: [激活 cloud-performance-analysis]

## 第一步：采集性能数据

请运行以下命令采集 CPU 热点：

```bash
# 1. 全局 CPU 采样 (30秒)
perf record -g -a -- sleep 30

# 2. 生成报告
perf report --stdio > perf_report.txt

# 3. 生成火焰图（需要 FlameGraph 工具）
perf script | stackcollapse-perf.pl | flamegraph.pl > cpu.svg
```

或者使用 BCC 工具：
```bash
# 实时查看热点
profile -F 99 -a 30 > profile.out
```

请把 perf_report.txt 的前 100 行发给我，我来分析热点。
```

### 示例 2：内存访问分析

```
用户: 我的程序在鲲鹏上性能不如 x86，怀疑是内存访问模式问题。

AI: [激活 cloud-performance-analysis]

鲲鹏 920 的 NUMA 架构与 x86 差异较大，需要分析内存访问模式。

## 诊断步骤

### 1. NUMA 访问分析
```bash
# 查看当前 NUMA 命中/未命中统计
numastat -p $(pidof your_app)

# 使用 eBPF 追踪跨 NUMA 访问
bpftrace -e '
kprobe:handle_mm_fault {
    @numa_access[pid, cpu] = count();
}
'
```

### 2. 缓存命中率分析
```bash
# LLC 缓存统计
perf stat -e cache-references,cache-misses,llc_prefetch.miss \
    -p $(pidof your_app) -- sleep 10

# 使用 BCC llcstat
llcstat.py 10 1
```

### 3. 内存带宽分析
```bash
# ARM 平台内存带宽监控
perf stat -e armv8_pmuv3_0/bus_cycles/,armv8_pmuv3_0/stall_slot/ \
    -p $(pidof your_app) -- sleep 10
```

### 鲲鹏优化建议
1. **NUMA 绑定**: 使用 `numactl --cpunodebind=0 --membind=0` 绑定到本地 NUMA
2. **页大小优化**: 考虑使用 64KB 大页（鲲鹏默认）
3. **预取优化**: 鲲鹏对软件预取敏感，适当增加 `__builtin_prefetch`
```

### 示例 3：锁竞争诊断

```
用户: 多线程程序性能随核心数增加反而下降，怀疑是锁问题。

AI: [激活 cloud-performance-analysis]

这是典型的锁竞争症状。让我帮你诊断。

## 锁竞争诊断

### 1. 快速检查
```bash
# 查看 futex 等待
perf top -e futex_wait -a

# 锁统计
perf lock record -a -- sleep 10
perf lock report
```

### 2. Off-CPU 分析
```bash
# 使用 offcputime 查看等待时间
offcputime.py -p $(pidof your_app) 10
```

### 3. eBPF 锁追踪
```bash
# 追踪 mutex 锁竞争
bpftrace -e '
kprobe:mutex_lock {
    @lock_caller[ustack] = count();
}
'
```

### 常见锁优化方案

| 问题 | 解决方案 |
|------|----------|
| 热点全局锁 | 分片锁、RCU、无锁数据结构 |
| 锁粒度太粗 | 细粒度锁、读写锁 |
| 锁持有时间长 | 减少临界区、延迟处理 |
| False Sharing | 缓存行对齐、填充 |

请运行以上命令，把输出发给我分析。
```

## 输出格式

性能分析报告包含：

1. **执行摘要** - 1-2 段问题描述和主要发现
2. **瓶颈分析** - Top 5 热点和调用栈
3. **根因分析** - 深层原因剖析
4. **优化建议** - 具体可执行的改进方案
5. **预期效果** - 优化后的性能预估
6. **验证方法** - 如何验证优化效果

## 注意事项

1. **生产环境采样** - 使用 `-F 99` 等低频率采样，避免影响业务
2. **多次采集** - 至少采集 3 次取平均值，避免偶然性
3. **上下文理解** - 了解业务逻辑，避免误判
4. **对比验证** - 优化前后用相同方法对比
5. **长期监控** - 建立性能基线，持续监控回归

## 相关 Skills

- `kernel-scheduler-tuning` - 调度器参数优化
- `kernel-debug-assistant` - 内核问题调试
- `kvm-virt-optimization` - KVM 性能优化
