# kunpeng-perf-tuning

> 鲲鹏 920 性能调优专家 - 针对华为云 ECS / openEuler

## 触发条件

- 鲲鹏 920、ARM64、openEuler
- 跨 NUMA 访问、内存带宽、缓存优化
- SPEC CPU、数据库性能调优
- 华为云 ECS 性能问题

## 鲲鹏 920 硬件特性

### 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                   鲲鹏 920 双路服务器                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Socket 0                    │  Socket 1                    │
│  ┌─────────────────────┐     │  ┌─────────────────────┐     │
│  │ Node 0    Node 1    │     │  │ Node 2    Node 3    │     │
│  │ CPU 0-31  CPU 32-63 │     │  │ CPU 64-95 CPU 96-127│     │
│  │ L3: 32MB  L3: 32MB  │     │  │ L3: 32MB  L3: 32MB  │     │
│  │ DDR4     DDR4       │     │  │ DDR4     DDR4       │     │
│  └─────────────────────┘     │  └─────────────────────┘     │
│                                                             │
│  跨 Socket 延迟: ~310ns     │  本地内存延迟: ~130ns        │
│  跨 Node 延迟:  ~180ns      │  带宽: ~120 GB/s             │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 内存访问延迟矩阵

```
        Node0   Node1   Node2   Node3
Node0    130     180     310     310
Node1    180     130     310     310
Node2    310     310     130     180
Node3    310     310     180     130

单位: ns (纳秒)
- 本地访问: 130ns
- 同 Socket 跨 Node: 180ns (+38%)
- 跨 Socket: 310ns (+138%)
```

### 缓存层次

```
L1 I-Cache: 64KB / core
L1 D-Cache: 64KB / core
L2 Cache:   512KB / core pair (共享)
L3 Cache:   32MB / DIE (共享)

预取:
- L1 预取距离: 2-4 cache lines
- L2 预取距离: 8-16 cache lines
```

## 诊断脚本

### 1. NUMA 访问模式分析

```bash
#!/bin/bash
# numa-access-analysis.sh - NUMA 访问模式分析

echo "=== NUMA 访问模式分析 ==="

# NUMA 拓扑
echo "[1] NUMA 拓扑"
numactl -H

# 当前进程 NUMA 命中率
echo -e "\n[2] NUMA 命中率 (需要 numactl --stat)"
if command -v numastat &> /dev/null; then
    numastat -p $$ 2>/dev/null || echo "运行 numastat -p <pid> 查看进程统计"
fi

# 使用 eBPF 追踪跨 NUMA 访问
echo -e "\n[3] 跨 NUMA 访问追踪 (eBPF)"
cat << 'EOF'
bpftrace -e '
kprobe:handle_mm_fault {
    @numa_access[pid] = count();
}
interval:s:5 {
    print(@numa_access);
    clear(@numa_access);
}
'
EOF
```

### 2. 内存带宽测试

```bash
#!/bin/bash
# memory-bandwidth-test.sh - 内存带宽测试

echo "=== 内存带宽测试 ==="

# 使用 stream benchmark
if command -v stream &> /dev/null; then
    echo "[使用 STREAM benchmark]"
    stream
else
    echo "[使用 sysbench]"
    sysbench memory --memory-block-size=1G \
        --memory-total-size=100G --memory-oper=write run
fi

# NUMA 感知带宽测试
echo -e "\n[2] 各 NUMA Node 带宽"
for node in 0 1 2 3; do
    echo "Node $node:"
    numactl --cpunodebind=$node --membind=$node \
        dd if=/dev/zero of=/dev/null bs=1M count=10000 2>&1 | grep -E "copied|bytes"
done
```

### 3. 缓存效率分析

```bash
#!/bin/bash
# cache-efficiency.sh - 缓存效率分析

echo "=== 缓存效率分析 ==="

# 使用 perf 分析缓存
echo "[1] 缓存 Miss 率"
perf stat -e cache-references,cache-misses,L1-dcache-load-misses,\
llc-load-misses,llc-store-misses \
    -p $(pgrep -n your_app) -- sleep 10

# 使用 BCC llcstat
echo -e "\n[2] LLC 统计 (需要 root)"
if command -v llcstat.py &> /dev/null; then
    llcstat.py 5
else
    echo "安装 BCC: yum install bcc-tools"
fi

# False Sharing 检测
echo -e "\n[3] False Sharing 检测"
perf c2c record -a -- sleep 10 2>/dev/null && \
    perf c2c report --stdio | head -50 || \
    echo "需要 CONFIG_SAMPLES=y 和 CONFIG_PERF_EVENTS=y"
```

### 4. CPU 热点分析

```bash
#!/bin/bash
# cpu-hotspot.sh - CPU 热点分析

echo "=== CPU 热点分析 ==="

# perf 采样
echo "[1] CPU 采样 (30秒)"
perf record -g -a -- sleep 30

# 生成报告
echo -e "\n[2] 热点报告"
perf report --stdio | head -50

# 生成火焰图
echo -e "\n[3] 火焰图生成"
if [ -d "/opt/FlameGraph" ]; then
    perf script | /opt/FlameGraph/stackcollapse-perf.pl | \
        /opt/FlameGraph/flamegraph.pl > /tmp/cpu_flame.svg
    echo "火焰图已生成: /tmp/cpu_flame.svg"
else
    echo "克隆 FlameGraph: git clone https://github.com/brendangregg/FlameGraph"
fi
```

## 优化建议

### 1. NUMA 绑定策略

```bash
# 应用绑定到单个 NUMA Node
numactl --cpunodebind=0 --membind=0 your_app

# 多实例，每个实例一个 Node
for i in 0 1 2 3; do
    numactl --cpunodebind=$i --membind=$i \
        your_app --instance=$i &
done

# 线程池绑定
taskset -c 0-31 your_app  # 绑定到 Node 0
```

### 2. 内存大页配置

```bash
# 64KB 大页（鲲鹏默认）
echo 4096 > /sys/kernel/mm/hugepages/hugepages-65536kB/nr_hugepages

# 1GB 大页（需要内核支持）
echo 64 > /sys/kernel/mm/hugepages/hugepages-1048576kB/nr_hugepages

# 编译时启用大页
gcc -DUSE_HUGEPAGE your_app.c -o your_app
```

### 3. 编译器优化选项

```makefile
# 鲲鹏 920 专用优化
CFLAGS += -march=armv8.2-a+crypto+fp16+dotprod
CFLAGS += -mtune=tsv110
CFLAGS += -O3 -ffast-math

# 自动向量化
CFLAGS += -ftree-vectorize -fopt-info-vec

# LTO (链接时优化)
CFLAGS += -flto
LDFLAGS += -flto

# NUMA 感知内存分配
LDFLAGS += -lnuma
```

### 4. 应用层优化

```c
// 预取优化
#define PREFETCH(addr) __builtin_prefetch(addr, 0, 3)

// 缓存行对齐
#define CACHE_ALIGNED __attribute__((aligned(64)))

// 避免 False Sharing
struct counter {
    volatile long value;
    char padding[56];  // 填充到 64 字节
} CACHE_ALIGNED;

// NUMA 感知分配
void *numa_alloc(size_t size, int node) {
    void *ptr = numa_alloc_onnode(size, node);
    return ptr;
}
```

## 数据库优化（MySQL/PostgreSQL）

### MySQL on 鲲鹏

```ini
# my.cnf
[mysqld]
# 缓冲池绑定 NUMA
innodb_numa_interleave = 1

# 大页
large-pages = 1

# I/O 优化
innodb_flush_method = O_DIRECT
innodb_io_capacity = 4000

# 并发
innodb_thread_concurrency = 0
innodb_read_io_threads = 16
innodb_write_io_threads = 16

# ARM64 优化
innodb_spin_wait_delay = 20
innodb_max_spin_wait_attempts = 60
```

### PostgreSQL on 鲲鹏

```ini
# postgresql.conf
# 内存
shared_buffers = 32GB
huge_pages = on
work_mem = 64MB

# 并发
max_worker_processes = 64
max_parallel_workers_per_gather = 8

# WAL
wal_buffers = 64MB
checkpoint_completion_target = 0.9
```

## SPEC CPU 优化

```bash
# SPEC CPU 2017 鲲鹏优化配置
# config.cfg

# 编译器
CC = gcc
CXX = g++
FC = gfortran

# 优化选项
OPTIMIZE = -O3 -march=armv8.2-a -mtune=tsv110 \
           -ffast-math -funroll-loops \
           -fno-plt -fno-ipa-sra

# OpenMP
EXTRA_CFLAGS = -fopenmp -DSPEC_OPENMP
EXTRA_LDFLAGS = -fopenmp

# 大页
HUGETLB_MORECORE = yes
HUGETLB_VERBOSE = 0

# NUMA
numactl --interleave=all runspec ...
```

## 验证脚本

```bash
#!/bin/bash
# verify-tuning.sh - 验证调优效果

echo "=== 调优效果验证 ==="

# 测试前基线
echo "[1] 收集基线数据"
perf stat -e cycles,instructions,cache-misses \
    -a -- sleep 10 > /tmp/baseline.txt

# 应用优化
echo "[2] 应用优化..."
source /tmp/tuning.sh

# 测试后数据
echo "[3] 收集优化后数据"
perf stat -e cycles,instructions,cache-misses \
    -a -- sleep 10 > /tmp/optimized.txt

# 对比
echo "[4] 对比结果"
diff /tmp/baseline.txt /tmp/optimized.txt
```

## 输出格式

性能分析报告包含：

1. **硬件概览** - NUMA 拓扑、缓存配置
2. **瓶颈识别** - 内存带宽、缓存 Miss、跨 NUMA 访问
3. **优化建议** - 具体配置和代码修改
4. **验证方法** - 如何验证优化效果
5. **预期收益** - 性能提升预估

## 注意事项

1. **NUMA 亲和性** - 确保应用和内存在同一 Node
2. **大页配置** - 减少页表开销
3. **编译优化** - 使用 ARM64 专用选项
4. **监控验证** - 优化后持续监控

## 相关 Skills

- `kernel-scheduler-tuning` - 调度器优化
- `huawei-kvm-debug` - KVM 虚拟化调试
- `cloud-performance-analysis` - 通用性能分析
