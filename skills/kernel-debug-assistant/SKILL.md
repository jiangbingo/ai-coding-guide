# kernel-debug-assistant

> 内核调试助手 - 崩溃分析、内存泄漏、竞态条件诊断

## 触发条件

当用户提到以下关键词时自动激活：

- 崩溃、crash、panic、kernel oops
- coredump、vmcore、kdump
- 内存泄漏、OOM、out of memory
- 竞态条件、死锁、race condition
- 内核调试、kgdb、ftrace
- Call Trace、堆栈分析

## 能力范围

### 1. 崩溃分析能力
- 解析 kernel oops/call trace
- 分析 vmcore 转储文件
- 定位崩溃根因
- 生成修复建议

### 2. 内存问题诊断
- 内存泄漏检测和定位
- OOM 问题分析
- 内存碎片化评估
- Slab 分配器问题

### 3. 并发问题诊断
- 死锁检测和分析
- 竞态条件定位
- 锁依赖分析
- 内存序问题

## 工作流程

```
┌─────────────────────────────────────────────────────────────┐
│                    内核调试工作流                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Step 1: 信息收集                                           │
│  ├── 获取崩溃日志 (dmesg, /var/log/messages)               │
│  ├── 收集 vmcore (如果配置了 kdump)                         │
│  ├── 了解触发条件                                           │
│  └── 获取内核版本和配置                                     │
│                                                             │
│  Step 2: 日志分析                                           │
│  ├── 解析 Call Trace                                        │
│  ├── 识别崩溃类型 (panic/oops/BUG)                         │
│  ├── 定位崩溃函数                                           │
│  └── 分析寄存器状态                                         │
│                                                             │
│  Step 3: 根因分析                                           │
│  ├── 分析源代码                                             │
│  ├── 检查并发访问                                           │
│  ├── 验证内存操作                                           │
│  └── 识别潜在 Bug                                           │
│                                                             │
│  Step 4: 修复建议                                           │
│  ├── 生成补丁代码                                           │
│  ├── 添加防御性检查                                         │
│  ├── 修改并发控制                                           │
│  └── 验证方案有效性                                         │
│                                                             │
│  Step 5: 预防措施                                           │
│  ├── 添加调试代码                                           │
│  ├── 增强日志输出                                           │
│  ├── 编写回归测试                                           │
│  └── 代码审查要点                                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 知识库

### 崩溃类型识别

| 类型 | 特征 | 常见原因 |
|------|------|----------|
| **Kernel Panic** | 系统完全停止 | 致命错误、硬件故障 |
| **Kernel Oops** | 继续运行但不可靠 | 空指针、非法地址 |
| **BUG()** | 主动触发 | 断言失败、非法状态 |
| **BUG_ON()** | 条件触发 | 前置条件不满足 |
| **divide error** | 除零错误 | 未检查除数 |
| **invalid opcode** | 非法指令 | 函数指针损坏 |

### Call Trace 解读

```
[  123.456789] BUG: unable to handle kernel NULL pointer dereference at 0000000000000018
[  123.456790] IP: [<ffffffffa0123456>] my_function+0x56/0xa0 [mymodule]
[  123.456791] PGD 0
[  123.456792] Oops: 0002 [#1] SMP
[  123.456793] CPU: 2 PID: 1234 Comm: myprocess Tainted: P           O 4.19.0
[  123.456794] Hardware name: Huawei TaiShan 200, BIOS 1.82
[  123.456795] task: ffff800012345678 task.stack: ffff800056789abc
[  123.456796] PC is at my_function+0x56/0xa0
[  123.456797] LR is at caller_function+0x28/0x40
[  123.456798] pc : [<ffffffffa0123456>] lr : [<ffffffffa0123498>] pstate: 60000005
[  123.456799] sp : ffff800056789c00
[  123.456800] Call trace:
[  123.456801]  my_function+0x56/0xa0
[  123.456802]  caller_function+0x28/0x40
[  123.456803]  process_one_work+0x198/0x340
[  123.456804]  worker_thread+0x48/0x4a0
```

### 调试工具

| 工具 | 用途 | 命令示例 |
|------|------|----------|
| **crash** | vmcore 分析 | `crash vmlinux vmcore` |
| **gdb** | 用户态 coredump | `gdb vmlinux core` |
| **ftrace** | 函数追踪 | `echo function > current_tracer` |
| **kprobes** | 动态探测 | `perf probe -a 'my_func'` |
| **eBPF** | 动态追踪 | `bpftrace -e 'kprobe:my_func {}'` |
| **kgdb** | 内核调试器 | `kgdboc=ttyS0,115200` |

### 内存问题模式

```
┌─────────────────────────────────────────────────────────────┐
│                    内存问题诊断树                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  内存问题                                                   │
│  ├── 泄漏 (Leak)                                           │
│  │   ├── kmalloc 无 kfree                                  │
│  │   ├── 引用计数未释放                                    │
│  │   └── 循环引用                                          │
│  │                                                         │
│  ├── 越界 (OOB)                                            │
│  │   ├── 数组越界                                          │
│  │   ├── 缓冲区溢出                                        │
│  │   └── Use-After-Free                                    │
│  │                                                         │
│  ├── 野指针 (Dangling)                                     │
│  │   ├── 释放后使用                                        │
│  │   ├── 未初始化指针                                      │
│  │   └── 竞态导致释放                                      │
│  │                                                         │
│  └── OOM (Out of Memory)                                   │
│      ├── 真实内存不足                                      │
│      ├── 内存碎片化                                        │
│      └── 内存限制过低                                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 并发问题模式

| 模式 | 症状 | 诊断方法 |
|------|------|----------|
| **ABBA 死锁** | 进程挂起 | `cat /proc/lockdep_chains` |
| **竞态条件** | 间歇性崩溃 | KASAN + 压力测试 |
| **原子上下文睡眠** | BUG: scheduling while atomic | `might_sleep()` 检查 |
| **中断安全违规** | 数据损坏 | `lockdep` 警告 |
| **内存序问题** | 状态不一致 | `WRITE_ONCE/READ_ONCE` |

## Prompt 模板

### 崩溃分析 Prompt

```markdown
# 内核崩溃分析

## 系统信息
- 内核版本: {{kernel_version}}
- 平台: {{platform}}  # 鲲鹏 920 / x86_64
- 触发场景: {{trigger}}

## 崩溃日志
```
{{call_trace}}
```

请分析：
1. 崩溃类型和原因
2. 问题函数和代码位置
3. 根因分析
4. 修复建议
5. 预防措施
```

### 内存泄漏分析 Prompt

```markdown
# 内存泄漏诊断

## 系统信息
- 内核版本: {{kernel_version}}
- 问题模块: {{module_name}}
- 症状: {{symptom}}  # 内存持续增长 / OOM

## 已有数据
{{meminfo}}  # /proc/meminfo, slabinfo

请分析：
1. 泄漏类型判断
2. 可能的泄漏点
3. 追踪和定位方法
4. 修复建议
```

## 使用示例

### 示例 1：空指针崩溃分析

```
用户: 我的内核模块崩溃了，日志如下：

[  123.456789] BUG: unable to handle kernel NULL pointer dereference at 0000000000000018
[  123.456790] IP: [<ffffffffa0123456>] my_probe+0x56/0xa0 [mymodule]
[  123.456801] Call trace:
[  123.456802]  my_probe+0x56/0xa0
[  123.456803]  pci_device_probe+0x88/0x120

AI: [激活 kernel-debug-assistant]

## 崩溃分析

### 问题识别
- **崩溃类型**: NULL 指针解引用
- **访问地址**: 0x18（偏移 24 字节）
- **崩溃位置**: `my_probe+0x56` in `mymodule`

### 根因分析

访问地址 `0x18` 表示通过空指针访问结构体成员，偏移 24 字节。
在 64 位系统上，这对应结构体中第 4 个指针成员（0, 8, 16, 24）。

可能的代码模式：
```c
// 崩溃代码推测
struct device *dev = NULL;  // 某个路径未初始化
dev->driver_data;           // 偏移 0x18 的成员访问
```

### 修复建议

1. **添加 NULL 检查**
```c
static int my_probe(struct pci_dev *pdev, const struct pci_device_id *id)
{
    struct my_device *dev;

    // 检查 pci_dev 是否有效
    if (!pdev) {
        pr_err("my_probe: pdev is NULL\n");
        return -EINVAL;
    }

    // 检查 id 是否有效
    if (!id) {
        pr_err("my_probe: id is NULL\n");
        return -EINVAL;
    }

    // ... 原有逻辑
}
```

2. **添加防御性编程**
```c
#define SAFE_ACCESS(ptr, member, default_val) \
    ((ptr) ? (ptr)->member : (default_val))
```

### 验证方法
```bash
# 使用 kprobes 追踪函数参数
perf probe -a 'my_probe pdev arg1=%x0 id arg2=%x1'

# 触发后查看
perf script
```
```

### 示例 2：内存泄漏诊断

```
用户: 我的内核模块运行一段时间后内存持续增长，怀疑有泄漏。

AI: [激活 kernel-debug-assistant]

## 内存泄漏诊断

### 第一步：确认泄漏

```bash
# 1. 查看整体内存使用
cat /proc/meminfo | grep -E "Mem|Slab|KernelStack|PageTables"

# 2. 查看 slab 统计
cat /proc/slabinfo | head -20
slabtop -o | head -20

# 3. 监控内存变化（5分钟）
for i in {1..10}; do
    echo "=== $(date) ==="
    grep -E "MemAvailable|Slab" /proc/meminfo
    sleep 30
done
```

### 第二步：定位泄漏源

```bash
# 方法1：使用 kmemleak（需要启用 CONFIG_DEBUG_KMEMLEAK）
echo scan > /sys/kernel/debug/kmemleak
cat /sys/kernel/debug/kmemleak

# 方法2：使用 ftrace 追踪 kmalloc/kfree
echo 1 > /sys/kernel/debug/tracing/events/kmem/kmalloc/enable
echo 1 > /sys/kernel/debug/tracing/events/kmem/kfree/enable
cat /sys/kernel/debug/tracing/trace_pipe > kmalloc_trace.log &
# 运行一段时间后停止
echo 0 > /sys/kernel/debug/tracing/events/kmem/kmalloc/enable
echo 0 > /sys/kernel/debug/tracing/events/kmem/kfree/enable

# 分析未匹配的分配
# 需要写脚本解析 kmalloc_trace.log
```

### 第三步：使用 eBPF 定位

```bash
# 使用 bpftrace 追踪内存分配
bpftrace -e '
kprobe:kmalloc {
    @alloc[ustack] = count();
}
kprobe:kfree {
    @free[ustack] = count();
}
interval:s:30 {
    print(@alloc);
    print(@free);
    clear(@alloc);
    clear(@free);
}
'
```

### 常见泄漏模式

| 模式 | 代码特征 | 修复方法 |
|------|----------|----------|
| 错误路径遗漏 | `if (error) return;` 无释放 | 使用 goto 统一退出 |
| 引用计数泄漏 | `get_device()` 无 `put_device()` | 检查所有返回路径 |
| 定时器泄漏 | `mod_timer()` 后未删除 | 模块卸载时清理 |
| 工作队列泄漏 | `queue_work()` 未取消 | `cancel_work_sync()` |

请运行以上诊断命令，把输出发给我分析。
```

### 示例 3：死锁分析

```
用户: 进程卡住了，dmesg 显示：
INFO: task myprocess:1234 blocked for more than 120 seconds.
      Tainted: P           O 4.19.0
"echo 0 > /proc/sys/kernel/hung_task_timeout_secs" disables this message.

AI: [激活 kernel-debug-assistant]

## 死锁诊断

### 第一步：获取锁状态

```bash
# 1. 查看被阻塞的任务栈
cat /proc/1234/stack

# 2. 查看锁依赖链（如果启用 lockdep）
cat /proc/lockdep_chains
cat /proc/lockdep_stats

# 3. 查看当前持有锁
cat /proc/lock_stats

# 4. 查看所有任务状态
ps auxw | grep -E "D|state"
```

### 第二步：分析调用栈

```bash
# 使用 SysRq 获取所有任务堆栈
echo t > /proc/sysrq-trigger
dmesg | grep -A 30 "myprocess"
```

### 第三步：使用 eBPF 追踪

```bash
# 追踪锁等待
bpftrace -e '
kprobe:__lock_acquire {
    @lock_caller[ustack] = count();
}
kprobe:mutex_lock_slowpath {
    printf("PID %d waiting for lock at %s\n", pid, kstack);
}
'
```

### 常见死锁模式

| 模式 | 描述 | 诊断特征 |
|------|------|----------|
| **ABBA** | 两个锁相反顺序获取 | lockdep 报告 circular dependency |
| **自死锁** | 同一线程重复获取 | 递归锁或信号量误用 |
| **中断死锁** | 中断上下文获取睡眠锁 | `scheduling while atomic` |
| **饥饿** | 长期无法获取锁 | 单个任务长时间 blocked |

### 解决方案

1. **统一锁顺序**
```c
// 定义锁获取顺序
// A -> B -> C (必须按此顺序)
mutex_lock(&lock_a);
mutex_lock(&lock_b);
mutex_lock(&lock_c);
// 释放可以任意顺序
```

2. **使用 lockdep 验证**
```bash
# 启用 lockdep
echo 1 > /proc/sys/kernel/lockdep

# 查看报告
dmesg | grep -i lockdep
```

3. **避免中断上下文睡眠**
```c
// 错误
irq_handler() {
    mutex_lock(&lock);  // 可能睡眠！
}

// 正确
irq_handler() {
    spin_lock(&lock);   // 不会睡眠
}
```

请提供 `/proc/1234/stack` 的输出，我来分析具体阻塞点。
```

## 输出格式

调试报告包含：

1. **问题摘要** - 崩溃类型和严重程度
2. **调用栈分析** - 关键函数和代码位置
3. **根因诊断** - 深层原因分析
4. **修复方案** - 具体代码修改建议
5. **验证步骤** - 如何验证修复有效
6. **预防措施** - 防止再次发生的方法

## 注意事项

1. **生产环境** - 调试工具可能影响性能，谨慎使用
2. **转储分析** - vmcore 包含敏感信息，注意安全
3. **符号表** - 确保有匹配的 vmlinux 和调试符号
4. **版本一致性** - 分析时使用相同内核版本的源码
5. **回归测试** - 修复后必须进行压力测试验证

## 相关 Skills

- `kernel-scheduler-tuning` - 调度器参数优化
- `cloud-performance-analysis` - 性能分析
- `kvm-virt-optimization` - KVM 虚拟化调试
