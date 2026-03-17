# Skill 创建完整指南

> 本指南帮助华为云杭州团队成员创建自己的 Claude Code Skills

## 什么是 Skill？

Skill 是 Claude Code 的**能力扩展模块**，让 AI 掌握特定领域的专业知识。

```
没有 Skill 时：
你: 帮我分析 KVM 性能问题
AI: 请提供更多信息...（不知道从何入手）

有 Skill 后：
你: 帮我分析 KVM 性能问题
AI: [自动激活 huawei-kvm-debug]
    我来帮你诊断 KVM 性能问题。
    首先检查 VM Exit 分布...
    然后分析 virtio 路径...
    （按照专业流程执行）
```

## Skill 文件结构

```
~/.claude/skills/
└── my-skill/
    └── SKILL.md              # 必须是这个文件名
```

### SKILL.md 核心模板

```markdown
# skill-name

> 一句话描述这个 Skill 的作用

## 触发条件

当用户提到以下关键词时自动激活：
- 关键词 1
- 关键词 2
- 关键词 3

## 能力范围

### 1. 诊断能力
- 能做什么

### 2. 优化能力
- 能做什么

### 3. 验证能力
- 能做什么

## 工作流程

```
Step 1: xxx
Step 2: xxx
Step 3: xxx
```

## 知识库

### 参数表格
| 参数 | 默认值 | 说明 |

### 命令速查
```bash
# 常用命令
```

## Prompt 模板

### 诊断 Prompt
```markdown
# 诊断模板
```

## 使用示例

### 示例 1：xxx
```
用户: xxx
AI: [激活 skill-name]
    xxx
```
```

## 完整示例：创建一个 KVM 内存诊断 Skill

### 步骤 1：创建目录

```bash
mkdir -p ~/.claude/skills/kvm-memory-diag
```

### 步骤 2：编写 SKILL.md

```markdown
# kvm-memory-diag

> KVM 虚拟机内存问题诊断专家

## 触发条件

- 内存泄漏、OOM、balloon
- EPT/NPT 页表问题
- 内存气球、大页配置
- 虚拟机内存性能差

## 能力范围

### 1. 内存使用分析
- 虚拟机内存分配统计
- 大页使用情况
- 内存气球状态

### 2. 页表效率分析
- EPT/NPT Miss 率
- Stage-2 页表大小

### 3. NUMA 亲和性
- 跨 NUMA 内存访问
- 内存绑定建议

## 工作流程

```
┌─────────────────────────────────────────────────────────────┐
│                    内存诊断工作流                            │
├─────────────────────────────────────────────────────────────┤
│  Step 1: 信息收集                                           │
│  ├── 虚拟机内存配置                                         │
│  ├── 当前使用情况                                           │
│  └── NUMA 分配                                              │
│                                                             │
│  Step 2: 问题识别                                           │
│  ├── OOM 检查                                               │
│  ├── 页表压力                                               │
│  └── 跨 NUMA 访问                                           │
│                                                             │
│  Step 3: 优化建议                                           │
│  ├── 大页配置                                               │
│  ├── NUMA 绑定                                              │
│  └── 内存限制调整                                           │
└─────────────────────────────────────────────────────────────┘
```

## 诊断脚本

### 内存使用统计

```bash
#!/bin/bash
# kvm-memory-stats.sh

VM_NAME=$1
if [ -z "$VM_NAME" ]; then
    echo "用法: $0 <vm-name>"
    exit 1
fi

echo "=== 虚拟机内存统计 ==="

# 获取虚拟机 PID
PID=$(virsh qemu-monitor-command $VM_NAME --hmp info status 2>/dev/null | \
      grep -oP 'pid=\K[0-9]+' || \
      pgrep -f "qemu.*$VM_NAME")

if [ -z "$PID" ]; then
    echo "错误: 找不到虚拟机进程"
    exit 1
fi

echo "[1] 虚拟机进程信息"
ps -p $PID -o pid,vsz,rss,comm

echo -e "\n[2] 内存统计 (cgroup)"
cat /sys/fs/cgroup/machine.slice/machine-qemu*.scope/$VM_NAME/memory.stat 2>/dev/null || \
    echo "cgroup v2 路径可能不同"

echo -e "\n[3] 大页使用"
grep -E "HugePages|hugepages" /proc/meminfo

echo -e "\n[4] NUMA 内存分布"
numastat -p $PID 2>/dev/null || echo "需要 numactl 包"

echo -e "\n[5] 内存气球状态"
virsh qemu-monitor-command $VM_NAME --hmp info balloon 2>/dev/null || \
    echo "气球驱动未启用"
```

### EPT/NPT 分析

```bash
#!/bin/bash
# ept-analysis.sh

echo "=== EPT/NPT 页表分析 ==="

# KVM MMU 统计 (需要 debugfs)
if mountpoint -q /sys/kernel/debug; then
    echo "[1] KVM MMU 统计"
    for vm in /sys/kernel/debug/kvm/*/; do
        echo "VM: $(basename $vm)"
        cat $vm/mmu_stats 2>/dev/null || echo "MMU 统计不可用"
    done
else
    echo "请先挂载 debugfs: mount -t debugfs none /sys/kernel/debug"
fi

# Stage-2 页表统计
echo -e "\n[2] 页表效率"
perf stat -e dTLB-load-misses,dTLB-store-misses \
    -a -- sleep 5 2>&1 | grep -E "dTLB|seconds"
```

### 内存带宽测试

```bash
#!/bin/bash
# memory-bw-test.sh

echo "=== 内存带宽测试 ==="

# 使用 mbw (需要安装)
if command -v mbw &> /dev/null; then
    echo "[使用 mbw]"
    mbw -n 10 256
else
    echo "[使用 dd 粗略估计]"
    dd if=/dev/zero of=/dev/null bs=1M count=10000 2>&1 | tail -1
fi

# NUMA 感知测试
echo -e "\n[各 NUMA Node 带宽]"
numactl -H | grep -E "^node [0-9]" | while read line; do
    node=$(echo $line | cut -d: -f1 | grep -oP '\d+')
    echo "Node $node:"
    numactl --cpunodebind=$node --membind=$node \
        dd if=/dev/zero of=/dev/null bs=1M count=5000 2>&1 | tail -1
done
```

## 优化建议模板

### 1. 大页配置

```xml
<!-- libvirt XML 配置 -->
<memoryBacking>
  <hugepages>
    <page size='64' unit='KiB' nodeset='0'/>
  </hugepages>
  <locked/>
</memoryBacking>
```

### 2. NUMA 绑定

```xml
<numatune>
  <memory mode='strict' nodeset='0'/>
  <memnode cellid='0' mode='strict' nodeset='0'/>
</numatune>
```

### 3. 内存限制

```xml
<memory unit='GiB'>8</memory>
<currentMemory unit='GiB'>8</currentMemory>
```

## 使用示例

### 示例 1：诊断虚拟机内存不足

```
用户: 我的虚拟机经常 OOM，帮我看看

AI: [激活 kvm-memory-diag]

我来帮你诊断虚拟机内存问题。

## 第一步：收集信息

请运行以下命令：

```bash
# 1. 查看虚拟机内存配置
virsh dominfo <vm-name> | grep -E "Memory|CPU"

# 2. 查看当前内存使用
virsh qemu-monitor-command <vm-name> --hmp info balloon

# 3. 查看 cgroup 内存限制
cat /sys/fs/cgroup/machine.slice/machine-*/memory.max
```

请把输出发给我，我来分析。
```

### 示例 2：优化内存性能

```
用户: 虚拟机内存性能差，怎么优化？

AI: [激活 kvm-memory-diag]

## 内存性能优化检查清单

### 1. 大页配置
- [ ] 是否启用 64KB 大页？
- [ ] 大页数量是否足够？

### 2. NUMA 亲和性
- [ ] 虚拟机内存是否绑定到同一 NUMA Node？
- [ ] vCPU 和内存是否在同一 Node？

### 3. 内存气球
- [ ] 是否启用了内存气球？
- [ ] 气球是否占用了过多内存？

### 4. 页表效率
- [ ] EPT/NPT Miss 率是否过高？

请运行诊断脚本：
```bash
./kvm-memory-stats.sh <vm-name>
```

我来分析结果并给出优化建议。
```

## 注意事项

1. 生产环境谨慎使用诊断脚本
2. 需要 root 权限访问部分调试信息
3. 优化后务必验证效果
```

### 步骤 3：验证 Skill

```bash
# 1. 复制到 Claude Code Skills 目录
cp -r kvm-memory-diag ~/.claude/skills/

# 2. 在 Claude Code 中测试
# 你: 帮我诊断 KVM 虚拟机内存问题
# AI 应该会自动激活 kvm-memory-diag
```

## Skill 最佳实践

### 1. 触发条件要精准

```markdown
## 触发条件

# ✅ 好的触发条件 - 具体且独特
- KVM、QEMU、virtio
- VM Exit、EPT、Stage-2
- vCPU 绑定、NUMA 亲和性

# ❌ 差的触发条件 - 太泛化
- 虚拟化（可能匹配 Docker、K8s）
- 性能（太宽泛）
- Linux（几乎所有场景）
```

### 2. 诊断脚本要实用

```bash
# ✅ 好的脚本 - 有输出、有解释
echo "=== 内存分析 ==="
echo "[1] 当前内存使用"
free -h

echo -e "\n[2] 大页配置"
grep Huge /proc/meminfo

# ❌ 差的脚本 - 只有命令没有说明
free -h
grep Huge /proc/meminfo
```

### 3. 优化建议要具体

```markdown
# ✅ 好的建议 - 可直接执行
## 优化配置

编辑 `/etc/libvirt/qemu/vm.xml`：

```xml
<memoryBacking>
  <hugepages>
    <page size='64' unit='KiB'/>
  </hugepages>
</memoryBacking>
```

应用：
```bash
virsh define /etc/libvirt/qemu/vm.xml
virsh reboot vm
```

# ❌ 差的建议 - 太抽象
## 优化配置

启用大页可以提高内存性能。
```

### 4. 工作流程要清晰

```
# ✅ 好的工作流 - 步骤明确

Step 1: 信息收集
├── CPU 拓扑
├── NUMA 信息
└── 当前配置

Step 2: 问题诊断
├── 运行测试
├── 分析结果
└── 识别瓶颈

Step 3: 优化方案
├── 生成配置
├── 应用变更
└── 验证效果

# ❌ 差的工作流 - 缺少细节

Step 1: 分析
Step 2: 优化
Step 3: 验证
```

## 高级技巧

### 1. 条件激活

```markdown
## 触发条件

当满足以下条件时激活：
1. 用户提到 KVM/QEMU
2. 且 环境是 openEuler 或华为云
3. 且 问题涉及性能或稳定性

优先级：
- 高优先级：VM Exit、EPT、virtio
- 中优先级：虚拟机、虚拟化
- 低优先级：性能优化（通用）
```

### 2. 与其他 Skill 协作

```markdown
## 相关 Skills

- `kernel-scheduler-tuning`: 当发现调度器问题时转发
- `cloud-performance-analysis`: 当需要通用性能分析时调用
- `huawei-kvm-debug`: 当需要更深入调试时调用

协作示例：
1. 本 Skill 发现内存问题 → 转发 kernel-scheduler-tuning 处理 NUMA
2. 本 Skill 发现性能瓶颈 → 调用 cloud-performance-analysis 生成火焰图
```

### 3. 知识库分层

```markdown
## 知识库

### 基础知识（所有用户）
- 内存层次结构
- 虚拟化基本概念

### 进阶知识（运维人员）
- cgroups 配置
- libvirt XML 参数

### 专家知识（内核开发者）
- EPT/NPT 实现原理
- Stage-2 页表结构
```

## 团队共享 Skill

### 方案 1：Git 仓库共享

```bash
# 创建团队 Skills 仓库
git init huawei-cloud-skills
cd huawei-cloud-skills

# 添加 Skills
mkdir -p skills/kvm-memory-diag
echo "# kvm-memory-diag" > skills/kvm-memory-diag/SKILL.md

# 推送到内网 GitLab
git remote add origin git@internal.gitlab.com:team/skills.git
git push -u origin main

# 团队成员同步
cd ~/.claude/skills
git clone git@internal.gitlab.com:team/skills.git huawei-cloud
```

### 方案 2：直接复制

```bash
# 打包 Skills
tar czf skills.tar.gz -C ~/.claude/skills .

# 分享给团队成员
# 解压到各自的 ~/.claude/skills/ 目录
```

## 调试 Skill

```bash
# 查看 Claude Code 是否识别到 Skill
ls -la ~/.claude/skills/*/

# 检查 SKILL.md 语法
cat ~/.claude/skills/my-skill/SKILL.md

# 测试触发
# 在 Claude Code 中说：帮我 <触发关键词>
```

## 常见问题

### Q: Skill 没有被激活？

A: 检查以下几点：
1. 文件名必须是 `SKILL.md`（大写）
2. 目录结构正确：`~/.claude/skills/<name>/SKILL.md`
3. 触发条件中的关键词足够具体

### Q: 如何更新 Skill？

A: 直接编辑 `SKILL.md` 文件，Claude Code 会自动读取最新内容。

### Q: 如何删除 Skill？

A: 删除对应的目录：
```bash
rm -rf ~/.claude/skills/my-skill
```

### Q: 多个 Skill 可能同时触发？

A: Claude Code 会根据上下文选择最相关的 Skill。可以在 Skill 中定义优先级。

## 总结

创建一个好的 Skill 需要：

1. **精准的触发条件** - 让 AI 知道什么时候用
2. **清晰的工作流程** - 让 AI 知道怎么做
3. **实用的脚本** - 用户可以直接运行
4. **具体的建议** - 可直接执行的配置
5. **完整的示例** - 展示使用场景

按照本指南的模板和最佳实践，你可以为团队创建高效的自定义 Skills！
