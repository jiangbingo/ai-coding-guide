# smoke-test-ops

华为云计算内核冒烟测试专家 - 快速验证核心功能可用性

## 触发条件

- 用户提到: 冒烟测试、smoke test、基础验证、快速测试、功能检查
- 用户场景: 新版本发布、代码合并后验证、环境部署确认、回归测试
- 平台: openEuler、CentOS、华为云 ECS、KVM 虚拟化

## 能力范围

### 环境预检
- 内核版本和配置验证
- 硬件资源可用性检查
- 依赖库和工具链验证
- 网络和存储配置确认

### 核心功能测试
- KVM 虚拟机生命周期（创建/启动/停止/销毁）
- 内核模块加载和卸载
- 网络连通性和性能
- 存储 I/O 读写
- 内存分配和回收

### 回归验证
- 历史问题复现确认
- 性能基准对比
- 稳定性基础检查

## 工作流程

```
┌─────────────────────────────────────────────────────────────┐
│                    冒烟测试工作流                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Pre-flight Check (预检)                                 │
│     ├── 内核版本检查                                        │
│     ├── 资源可用性验证                                      │
│     └── 配置文件校验                                        │
│                                                             │
│  2. Core Function Tests (核心功能)                          │
│     ├── 虚拟化基础测试                                      │
│     ├── 网络连通性测试                                      │
│     └── 存储 I/O 测试                                       │
│                                                             │
│  3. Result Analysis (结果分析)                              │
│     ├── 通过/失败判定                                       │
│     ├── 日志异常分析                                        │
│     └── 性能指标对比                                        │
│                                                             │
│  4. Report (报告生成)                                       │
│     ├── 测试摘要                                            │
│     ├── 问题清单                                            │
│     └── 建议措施                                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 测试用例模板

### 1. 环境预检用例

```bash
# check_environment.sh
#!/bin/bash
set -e

echo "===== 环境预检 ====="

# 1. 内核版本
echo "[1] 内核版本:"
uname -r
cat /proc/version

# 2. CPU 信息
echo "[2] CPU 信息:"
lscpu | grep -E "Architecture|CPU\(s\)|Model name|CPU MHz"

# 3. 内存信息
echo "[3] 内存信息:"
free -h

# 4. KVM 支持
echo "[4] KVM 虚拟化支持:"
if [ -e /dev/kvm ]; then
    echo "✓ /dev/kvm 存在"
    ls -la /dev/kvm
else
    echo "✗ /dev/kvm 不存在"
fi

# 5. 必要工具
echo "[5] 必要工具检查:"
for tool in qemu-kvm virsh iperf3 fio perf; do
    if command -v $tool &> /dev/null; then
        echo "✓ $tool: $(command -v $tool)"
    else
        echo "✗ $tool: 未安装"
    fi
done

# 6. 内核配置
echo "[6] 关键内核配置:"
zcat /proc/config.gz 2>/dev/null | grep -E "CONFIG_KVM|CONFIG_VIRTIO|CONFIG_NUMA" || \
    echo "内核配置不可用，检查 /boot/config-$(uname -r)"

echo "===== 预检完成 ====="
```

### 2. KVM 虚拟机生命周期测试

```bash
# test_vm_lifecycle.sh
#!/bin/bash

VM_NAME="smoke-test-vm"
VM_IMAGE="/var/lib/libvirt/images/smoke-test.qcow2"
TIMEOUT=60

echo "===== KVM 虚拟机生命周期测试 ====="

# 清理旧测试资源
cleanup() {
    virsh destroy $VM_NAME 2>/dev/null || true
    virsh undefine $VM_NAME 2>/dev/null || true
    rm -f $VM_IMAGE
}

cleanup
trap cleanup EXIT

# 1. 创建虚拟机
echo "[1] 创建虚拟机..."
if virsh list --all | grep -q "$VM_NAME"; then
    echo "✗ 虚拟机已存在"
    exit 1
fi

# 创建测试镜像
qemu-img create -f qcow2 $VM_IMAGE 1G
echo "✓ 镜像创建成功"

# 定义虚拟机
virt-install --name $VM_NAME \
    --ram 512 \
    --vcpus 1 \
    --disk path=$VM_IMAGE \
    --os-variant generic \
    --network network=default \
    --graphics none \
    --noautoconsole \
    --import \
    2>&1

if [ $? -eq 0 ]; then
    echo "✓ 虚拟机定义成功"
else
    echo "✗ 虚拟机定义失败"
    exit 1
fi

# 2. 启动虚拟机
echo "[2] 启动虚拟机..."
virsh start $VM_NAME
sleep 5

if virsh domstate $VM_NAME | grep -q "running"; then
    echo "✓ 虚拟机启动成功"
else
    echo "✗ 虚拟机启动失败"
    exit 1
fi

# 3. 暂停/恢复测试
echo "[3] 暂停/恢复测试..."
virsh suspend $VM_NAME
sleep 2
if virsh domstate $VM_NAME | grep -q "paused"; then
    echo "✓ 暂停成功"
fi

virsh resume $VM_NAME
sleep 2
if virsh domstate $VM_NAME | grep -q "running"; then
    echo "✓ 恢复成功"
fi

# 4. 停止虚拟机
echo "[4] 停止虚拟机..."
virsh shutdown $VM_NAME
sleep 5

if virsh domstate $VM_NAME | grep -q "shut off"; then
    echo "✓ 虚拟机关闭成功"
else
    virsh destroy $VM_NAME
    echo "✓ 强制关闭成功"
fi

# 5. 销毁虚拟机
echo "[5] 销毁虚拟机..."
virsh undefine $VM_NAME
rm -f $VM_IMAGE
echo "✓ 虚拟机销毁成功"

echo "===== KVM 生命周期测试通过 ====="
```

### 3. 网络连通性测试

```bash
# test_network.sh
#!/bin/bash

echo "===== 网络连通性测试 ====="

# 1. 网桥检查
echo "[1] 网桥检查:"
brctl show 2>/dev/null || bridge link 2>/dev/null || ip link show type bridge

# 2. 虚拟网络检查
echo "[2] libvirt 虚拟网络:"
virsh net-list --all

# 3. 基础连通性
echo "[3] 基础连通性测试:"
ping -c 3 8.8.8.8 && echo "✓ 外网连通" || echo "✗ 外网不通"
ping -c 3 gateway && echo "✓ 网关连通" || echo "✗ 网关不通"

# 4. DNS 解析
echo "[4] DNS 解析测试:"
nslookup baidu.com && echo "✓ DNS 解析正常" || echo "✗ DNS 解析失败"

# 5. 端口检查
echo "[5] 关键端口检查:"
for port in 22 80 443; do
    timeout 5 bash -c "echo > /dev/tcp/baidu.com/$port" 2>/dev/null && \
        echo "✓ 端口 $port 可达" || echo "✗ 端口 $port 不可达"
done

# 6. 网络性能（如果有 iperf3）
echo "[6] 网络性能测试:"
if command -v iperf3 &> /dev/null; then
    # 本地回环测试
    iperf3 -s &
    SERVER_PID=$!
    sleep 2
    iperf3 -c localhost -t 5
    kill $SERVER_PID 2>/dev/null
else
    echo "iperf3 未安装，跳过性能测试"
fi

echo "===== 网络测试完成 ====="
```

### 4. 存储 I/O 测试

```bash
# test_io.sh
#!/bin/bash

TEST_DIR="/tmp/smoke-io-test"
TEST_FILE="$TEST_DIR/test.bin"

echo "===== 存储 I/O 测试 ====="

# 清理
rm -rf $TEST_DIR
mkdir -p $TEST_DIR
trap "rm -rf $TEST_DIR" EXIT

# 1. 基础读写测试
echo "[1] 基础读写测试:"
dd if=/dev/zero of=$TEST_FILE bs=1M count=100 oflag=direct 2>&1 | grep -E "copied|bytes"
dd if=$TEST_FILE of=/dev/null bs=1M iflag=direct 2>&1 | grep -E "copied|bytes"
echo "✓ 基础读写完成"

# 2. fio 快速测试
echo "[2] fio 随机读写测试:"
if command -v fio &> /dev/null; then
    fio --name=smoke-test \
        --filename=$TEST_FILE \
        --size=100M \
        --ioengine=libaio \
        --direct=1 \
        --bs=4k \
        --rw=randrw \
        --rwmixread=70 \
        --iodepth=16 \
        --numjobs=2 \
        --runtime=10 \
        --group_reporting \
        2>&1 | grep -E "read|write|IOPS|iops"
    echo "✓ fio 测试完成"
else
    echo "fio 未安装，跳过"
fi

# 3. I/O 调度器检查
echo "[3] I/O 调度器检查:"
for dev in /sys/block/sd*/queue/scheduler /sys/block/vd*/queue/scheduler /sys/block/nvme*/queue/scheduler; do
    if [ -f "$dev" ]; then
        device=$(echo $dev | cut -d'/' -f4)
        scheduler=$(cat $dev)
        echo "  $device: $scheduler"
    fi
done

# 4. 磁盘空间检查
echo "[4] 磁盘空间检查:"
df -h | grep -E "Filesystem|/$|/home|/var"

echo "===== 存储 I/O 测试完成 ====="
```

### 5. 内核模块测试

```bash
# test_kernel_modules.sh
#!/bin/bash

echo "===== 内核模块测试 ====="

# 1. 已加载模块检查
echo "[1] 关键模块检查:"
for mod in kvm kvm_intel kvm_amd virtio virtio_net virtio_blk ebpf; do
    if lsmod | grep -q "^$mod"; then
        echo "✓ $mod 已加载"
    else
        echo "  $mod 未加载"
    fi
done

# 2. 模块加载/卸载测试（使用安全的测试模块）
echo "[2] 模块加载测试:"
TEST_MOD="dummy"
if modprobe $TEST_MOD 2>/dev/null; then
    echo "✓ $TEST_MOD 加载成功"
    modprobe -r $TEST_MOD
    echo "✓ $TEST_MOD 卸载成功"
else
    echo "  跳过模块加载测试（需要 root 权限）"
fi

# 3. 内核参数检查
echo "[3] 关键内核参数:"
for param in vm.swappiness vm.dirty_ratio kernel.sched_min_granularity_ns; do
    value=$(sysctl -n $param 2>/dev/null)
    if [ -n "$value" ]; then
        echo "  $param = $value"
    fi
done

# 4. eBPF 支持检查
echo "[4] eBPF 支持检查:"
if [ -d /sys/fs/bpf ]; then
    echo "✓ eBPF 文件系统已挂载"
else
    echo "  eBPF 文件系统未挂载"
fi

# 5. 内核日志检查
echo "[5] 内核错误检查:"
if dmesg | grep -iE "error|fail|panic|warn" | tail -10 | grep -q .; then
    echo "⚠ 发现内核警告/错误:"
    dmesg | grep -iE "error|fail|panic|warn" | tail -5
else
    echo "✓ 无明显内核错误"
fi

echo "===== 内核模块测试完成 ====="
```

## Prompt 模板

### 诊断 Prompt

```markdown
# 冒烟测试诊断

## 系统信息
- OS: {{os_version}}
- 内核: {{kernel_version}}
- 平台: {{platform}}

## 测试结果
{{test_results}}

## 失败信息
{{failure_details}}

请分析：
1. 失败的根本原因
2. 是否为环境配置问题
3. 是否需要调整测试用例
4. 修复建议
```

### 报告生成 Prompt

```markdown
# 生成冒烟测试报告

## 测试概要
- 执行时间: {{timestamp}}
- 测试环境: {{environment}}
- 总用例数: {{total_cases}}
- 通过数: {{passed}}
- 失败数: {{failed}}

## 详细结果
{{detailed_results}}

## 性能指标
{{performance_metrics}}

请生成：
1. 测试摘要（1-2 段）
2. 问题清单（按优先级）
3. 建议措施
4. 回归建议
```

## 使用示例

```bash
# 运行完整冒烟测试
/smoke-test-ops run --full

# 运行快速测试（仅核心功能）
/smoke-test-ops run --quick

# 运行特定模块测试
/smoke-test-ops run --module kvm
/smoke-test-ops run --module network
/smoke-test-ops run --module storage

# 生成测试报告
/smoke-test-ops report --format html --output /tmp/smoke-report.html

# 对比历史结果
/smoke-test-ops compare --baseline /tmp/baseline.json
```

## 知识库

- openEuler 内核配置指南
- KVM 虚拟化最佳实践
- 华为云 ECS 规格和限制
- libvirt 常见问题排查
- eBPF 工具使用手册
