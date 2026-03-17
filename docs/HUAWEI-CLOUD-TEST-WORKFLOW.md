# 华为云计算内核冒烟测试工作流

> 针对华为云杭州团队的完整测试工作流设计

## 目录

1. [工作流概述](#工作流概述)
2. [测试环境准备](#测试环境准备)
3. [测试用例设计](#测试用例设计)
4. [自动化执行](#自动化执行)
5. [结果分析与报告](#结果分析与报告)
6. [CI/CD 集成](#cicd-集成)
7. [AI 辅助测试](#ai-辅助测试)

---

## 工作流概述

### 整体架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          冒烟测试自动化工作流                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐ │
│   │ Trigger │───▶│ Prepare │───▶│ Execute │───▶│ Analyze │───▶│ Report  │ │
│   │ 触发器  │    │ 环境准备 │    │ 测试执行 │    │ 结果分析 │    │ 报告生成 │ │
│   └─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘ │
│        │              │              │              │              │        │
│        ▼              ▼              ▼              ▼              ▼        │
│   ┌─────────────────────────────────────────────────────────────────────┐ │
│   │                        AI Agent 辅助层                                │ │
│   │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────────────┐ │ │
│   │  │诊断 Agent │  │分析 Agent │  │报告 Agent │  │CodeArts 集成 Agent│ │ │
│   │  └───────────┘  └───────────┘  └───────────┘  └───────────────────┘ │ │
│   └─────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 触发条件

| 触发类型 | 描述 | 测试范围 |
|---------|------|---------|
| **代码提交** | MR/Push 到 main 分支 | 快速测试 |
| **定时触发** | 每日夜间构建 | 完整测试 |
| **手动触发** | 发布前验证 | 完整测试 + 性能测试 |
| **环境变更** | 基础设施更新 | 完整测试 |

---

## 测试环境准备

### 1. 环境检查清单

```bash
#!/bin/bash
# scripts/env_check.sh

set -e

echo "=========================================="
echo "  华为云计算内核测试环境检查"
echo "=========================================="

PASS=0
FAIL=0

check_item() {
    local name=$1
    local condition=$2
    if eval "$condition"; then
        echo "✓ $name"
        ((PASS++))
    else
        echo "✗ $name"
        ((FAIL++))
    fi
}

# 1. 操作系统检查
echo ""
echo "[1] 操作系统检查"
check_item "openEuler/CentOS" "grep -qE 'openEuler|CentOS' /etc/os-release"
check_item "内核版本 >= 5.10" "uname -r | grep -qE '^5\.1[0-9]|^[6-9]\.'"
check_item "架构支持" "uname -m | grep -qE 'x86_64|aarch64'"

# 2. 虚拟化支持
echo ""
echo "[2] 虚拟化支持"
check_item "KVM 模块" "lsmod | grep -q kvm"
check_item "/dev/kvm 存在" "test -e /dev/kvm"
check_item "libvirtd 运行" "systemctl is-active libvirtd &>/dev/null"

# 3. 网络配置
echo ""
echo "[3] 网络配置"
check_item "网桥存在" "brctl show 2>/dev/null | grep -q virbr || ip link show type bridge | grep -q ."
check_item "IP 转发" "sysctl -n net.ipv4.ip_forward | grep -q 1"

# 4. 存储配置
echo ""
echo "[4] 存储配置"
check_item "libvirt 目录存在" "test -d /var/lib/libvirt"
check_item "磁盘空间充足 (>10G)" "df /var | awk 'NR==2 {exit (\$4 > 10485760)}'"

# 5. 工具链
echo ""
echo "[5] 工具链检查"
for tool in qemu-kvm virsh fio iperf3 perf bpftool; do
    check_item "$tool" "command -v $tool &>/dev/null"
done

# 6. 权限检查
echo ""
echo "[6] 权限检查"
check_item "root 或 sudo 权限" "id -u | grep -q 0 || sudo -n true 2>/dev/null"

# 汇总
echo ""
echo "=========================================="
echo "  检查结果: $PASS 通过, $FAIL 失败"
echo "=========================================="

if [ $FAIL -gt 0 ]; then
    exit 1
fi
```

### 2. 测试资源配置

```yaml
# config/test_resources.yaml

# 虚拟机模板
vm_templates:
  minimal:
    vcpus: 1
    memory: 512M
    disk: 1G
    network: default
    os: generic

  standard:
    vcpus: 2
    memory: 2G
    disk: 10G
    network: default
    os: openeuler-22.03

# 测试镜像
test_images:
  - name: smoke-test-base
    url: obs://huawei-cloud-mirror/smoke-test-base.qcow2
    checksum: sha256:abc123...

  - name: openeuler-minimal
    url: obs://huawei-cloud-mirror/openeuler-minimal.qcow2
    checksum: sha256:def456...

# 网络配置
networks:
  default:
    type: nat
    subnet: 192.168.122.0/24

  isolated:
    type: isolated
    subnet: 10.0.0.0/24

# 性能基准
baselines:
  vm_boot_time: 30s
  network_latency: 1ms
  disk_iops_4k: 10000
  disk_throughput: 100MB/s
```

---

## 测试用例设计

### 1. 测试用例分类

```
测试用例
├── P0 - 阻塞级（必须通过）
│   ├── KVM 虚拟机启动
│   ├── 网络基础连通
│   └── 存储 I/O 可用
│
├── P1 - 严重级（核心功能）
│   ├── 虚拟机迁移
│   ├── 快照恢复
│   ├── 热插拔设备
│   └── 内核模块加载
│
├── P2 - 一般级（扩展功能）
│   ├── NUMA 绑定
│   ├── CPU 亲和性
│   ├── 内存大页
│   └── SR-IOV
│
└── P3 - 建议级（性能验证）
    ├── 网络吞吐量
    ├── 存储 IOPS
    └── 并发压力
```

### 2. 测试用例模板

```yaml
# testcases/kvm_lifecycle.yaml

id: TC-KVM-001
name: KVM 虚拟机生命周期测试
priority: P0
tags: [kvm, virtualization, critical]

preconditions:
  - libvirtd 服务运行
  - 测试镜像存在
  - 网络配置正确

test_steps:
  - step: 1
    action: 创建虚拟机
    command: virsh define vm.xml
    expected: 退出码 0

  - step: 2
    action: 启动虚拟机
    command: virsh start test-vm
    expected: 退出码 0, 状态 running
    timeout: 30s

  - step: 3
    action: 验证运行状态
    command: virsh domstate test-vm
    expected: 输出 "running"

  - step: 4
    action: 暂停虚拟机
    command: virsh suspend test-vm
    expected: 状态 paused

  - step: 5
    action: 恢复虚拟机
    command: virsh resume test-vm
    expected: 状态 running

  - step: 6
    action: 关闭虚拟机
    command: virsh shutdown test-vm
    expected: 状态 "shut off"
    timeout: 60s

  - step: 7
    action: 销毁虚拟机
    command: virsh undefine test-vm
    expected: 退出码 0

cleanup:
  - virsh destroy test-vm 2>/dev/null || true
  - virsh undefine test-vm 2>/dev/null || true

metrics:
  boot_time:
    description: 启动时间
    threshold: 30s
    collection: virsh dominfo test-vm | grep "CPU time"
```

### 3. 测试矩阵

| 用例 ID | 用例名称 | openEuler 22.03 | CentOS 7.9 | 鲲鹏 920 | x86_64 |
|---------|---------|-----------------|------------|----------|--------|
| TC-KVM-001 | 虚拟机生命周期 | ✓ | ✓ | ✓ | ✓ |
| TC-NET-001 | 网络连通性 | ✓ | ✓ | ✓ | ✓ |
| TC-IO-001 | 存储 I/O | ✓ | ✓ | ✓ | ✓ |
| TC-PERF-001 | eBPF 追踪 | ✓ | ○ | ✓ | ✓ |
| TC-NUMA-001 | NUMA 绑定 | ✓ | ○ | ✓ | ○ |

---

## 自动化执行

### 1. 测试执行脚本

```bash
#!/bin/bash
# scripts/run_smoke_tests.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORT_DIR="${PROJECT_ROOT}/reports/$(date +%Y%m%d_%H%M%S)"

# 创建报告目录
mkdir -p "$REPORT_DIR"

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$REPORT_DIR/test.log"
}

# 测试结果统计
TOTAL=0
PASSED=0
FAILED=0
SKIPPED=0

run_test() {
    local test_name=$1
    local test_script=$2
    local priority=${3:-P1}

    ((TOTAL++))

    log "========== 开始测试: $test_name [$priority] =========="

    local start_time=$(date +%s)

    if bash "$test_script" > "$REPORT_DIR/${test_name}.log" 2>&1; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        log "✓ $test_name 通过 (耗时: ${duration}s)"
        ((PASSED++))
        echo "$test_name,PASS,$duration,$priority" >> "$REPORT_DIR/results.csv"
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        log "✗ $test_name 失败 (耗时: ${duration}s)"
        ((FAILED++))
        echo "$test_name,FAIL,$duration,$priority" >> "$REPORT_DIR/results.csv"

        # P0 测试失败立即退出
        if [ "$priority" == "P0" ]; then
            log "P0 测试失败，终止测试"
            exit 1
        fi
    fi
}

# 环境检查
log "========== 环境检查 =========="
if ! bash "$SCRIPT_DIR/env_check.sh"; then
    log "环境检查失败，请先解决环境问题"
    exit 1
fi

# 执行测试
log "========== 开始冒烟测试 =========="

# P0 测试（阻塞级）
run_test "kvm_lifecycle" "$SCRIPT_DIR/tests/test_kvm_lifecycle.sh" "P0"
run_test "network_basic" "$SCRIPT_DIR/tests/test_network_basic.sh" "P0"
run_test "storage_io" "$SCRIPT_DIR/tests/test_storage_io.sh" "P0"

# P1 测试（严重级）
run_test "kernel_modules" "$SCRIPT_DIR/tests/test_kernel_modules.sh" "P1"
run_test "vm_snapshot" "$SCRIPT_DIR/tests/test_vm_snapshot.sh" "P1"
run_test "vm_migration" "$SCRIPT_DIR/tests/test_vm_migration.sh" "P1"

# P2 测试（一般级）
run_test "numa_binding" "$SCRIPT_DIR/tests/test_numa_binding.sh" "P2"
run_test "hugepages" "$SCRIPT_DIR/tests/test_hugepages.sh" "P2"

# 生成报告
log "========== 生成测试报告 =========="
python3 "$SCRIPT_DIR/generate_report.py" \
    --results "$REPORT_DIR/results.csv" \
    --logs "$REPORT_DIR" \
    --output "$REPORT_DIR/report.html"

# 汇总
log "========== 测试汇总 =========="
log "总计: $TOTAL, 通过: $PASSED, 失败: $FAILED, 跳过: $SKIPPED"
log "报告目录: $REPORT_DIR"

if [ $FAILED -gt 0 ]; then
    exit 1
fi
```

### 2. Python 测试框架

```python
#!/usr/bin/env python3
# tests/smoke_test_framework.py

import subprocess
import time
import json
from dataclasses import dataclass
from typing import Optional, List, Callable
from enum import Enum
import xml.etree.ElementTree as ET

class Priority(Enum):
    P0 = "P0"  # 阻塞级
    P1 = "P1"  # 严重级
    P2 = "P2"  # 一般级
    P3 = "P3"  # 建议级

@dataclass
class TestResult:
    name: str
    passed: bool
    duration: float
    priority: Priority
    message: str = ""
    details: Optional[dict] = None

class SmokeTest:
    def __init__(self, name: str, priority: Priority = Priority.P1):
        self.name = name
        self.priority = priority
        self.results: List[TestResult] = []

    def run_command(self, cmd: str, timeout: int = 60) -> tuple:
        """执行命令并返回结果"""
        try:
            result = subprocess.run(
                cmd,
                shell=True,
                capture_output=True,
                text=True,
                timeout=timeout
            )
            return result.returncode, result.stdout, result.stderr
        except subprocess.TimeoutExpired:
            return -1, "", "Command timed out"

    def assert_command(self, cmd: str, expected_code: int = 0, timeout: int = 60) -> bool:
        """断言命令返回指定退出码"""
        code, stdout, stderr = self.run_command(cmd, timeout)
        return code == expected_code

    def assert_output_contains(self, cmd: str, pattern: str, timeout: int = 60) -> bool:
        """断言命令输出包含指定内容"""
        code, stdout, stderr = self.run_command(cmd, timeout)
        return pattern in stdout or pattern in stderr

class KVMSmokeTest(SmokeTest):
    """KVM 虚拟化冒烟测试"""

    def __init__(self):
        super().__init__("KVM Smoke Test", Priority.P0)
        self.vm_name = "smoke-test-vm"

    def setup(self):
        """测试前准备"""
        # 清理旧资源
        self.run_command(f"virsh destroy {self.vm_name} 2>/dev/null")
        self.run_command(f"virsh undefine {self.vm_name} 2>/dev/null")

    def teardown(self):
        """测试后清理"""
        self.run_command(f"virsh destroy {self.vm_name} 2>/dev/null")
        self.run_command(f"virsh undefine {self.vm_name} 2>/dev/null")

    def test_vm_define(self) -> TestResult:
        """测试虚拟机定义"""
        start = time.time()

        # 创建虚拟机 XML
        xml = f"""
        <domain type='kvm'>
          <name>{self.vm_name}</name>
          <memory unit='MiB'>512</memory>
          <vcpu>1</vcpu>
          <os><type arch='x86_64'>hvm</type></os>
          <devices>
            <disk type='file'>
              <driver name='qemu' type='qcow2'/>
              <source file='/tmp/{self.vm_name}.qcow2'/>
              <target dev='vda'/>
            </disk>
          </devices>
        </domain>
        """

        # 创建磁盘
        self.run_command(f"qemu-img create -f qcow2 /tmp/{self.vm_name}.qcow2 1G")

        # 定义虚拟机
        code, _, stderr = self.run_command(f"echo '{xml}' | virsh define /dev/stdin")

        return TestResult(
            name="vm_define",
            passed=(code == 0),
            duration=time.time() - start,
            priority=Priority.P0,
            message=stderr if code != 0 else "OK"
        )

    def test_vm_lifecycle(self) -> TestResult:
        """测试虚拟机生命周期"""
        start = time.time()
        errors = []

        # 启动
        code, _, stderr = self.run_command(f"virsh start {self.vm_name}", timeout=30)
        if code != 0:
            errors.append(f"Start failed: {stderr}")

        time.sleep(5)

        # 检查状态
        code, stdout, _ = self.run_command(f"virsh domstate {self.vm_name}")
        if "running" not in stdout:
            errors.append(f"Not running: {stdout}")

        # 停止
        code, _, stderr = self.run_command(f"virsh shutdown {self.vm_name}", timeout=60)
        if code != 0:
            self.run_command(f"virsh destroy {self.vm_name}")

        return TestResult(
            name="vm_lifecycle",
            passed=len(errors) == 0,
            duration=time.time() - start,
            priority=Priority.P0,
            message="; ".join(errors) if errors else "OK"
        )

    def run_all(self) -> List[TestResult]:
        """运行所有测试"""
        self.setup()
        try:
            self.results.append(self.test_vm_define())
            self.results.append(self.test_vm_lifecycle())
        finally:
            self.teardown()
        return self.results

class TestRunner:
    """测试运行器"""

    def __init__(self):
        self.results: List[TestResult] = []

    def run_test(self, test: SmokeTest):
        """运行单个测试"""
        results = test.run_all()
        self.results.extend(results)

    def generate_report(self, output_path: str):
        """生成 HTML 报告"""
        html = """
        <!DOCTYPE html>
        <html>
        <head>
            <title>冒烟测试报告</title>
            <style>
                body { font-family: sans-serif; margin: 20px; }
                .pass { color: green; }
                .fail { color: red; }
                table { border-collapse: collapse; width: 100%; }
                th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
            </style>
        </head>
        <body>
            <h1>冒烟测试报告</h1>
            <p>执行时间: {timestamp}</p>
            <table>
                <tr><th>测试名称</th><th>优先级</th><th>结果</th><th>耗时</th><th>信息</th></tr>
                {rows}
            </table>
            <h2>汇总</h2>
            <p>总计: {total}, 通过: {passed}, 失败: {failed}</p>
        </body>
        </html>
        """

        rows = ""
        for r in self.results:
            status = "✓ 通过" if r.passed else "✗ 失败"
            cls = "pass" if r.passed else "fail"
            rows += f"""
                <tr>
                    <td>{r.name}</td>
                    <td>{r.priority.value}</td>
                    <td class="{cls}">{status}</td>
                    <td>{r.duration:.2f}s</td>
                    <td>{r.message}</td>
                </tr>
            """

        passed = sum(1 for r in self.results if r.passed)
        failed = len(self.results) - passed

        html = html.format(
            timestamp=time.strftime("%Y-%m-%d %H:%M:%S"),
            rows=rows,
            total=len(self.results),
            passed=passed,
            failed=failed
        )

        with open(output_path, 'w') as f:
            f.write(html)

# 使用示例
if __name__ == "__main__":
    runner = TestRunner()
    runner.run_test(KVMSmokeTest())
    runner.generate_report("/tmp/smoke_report.html")
```

---

## 结果分析与报告

### 1. 结果分析逻辑

```python
#!/usr/bin/env python3
# scripts/analyze_results.py

import json
import re
from dataclasses import dataclass
from typing import List, Dict
from collections import defaultdict

@dataclass
class FailureAnalysis:
    test_name: str
    error_type: str
    root_cause: str
    suggestion: str
    related_issues: List[str]

class ResultAnalyzer:
    """测试结果分析器"""

    # 已知错误模式
    ERROR_PATTERNS = {
        "permission_denied": {
            "pattern": r"Permission denied|Access denied|操作不允许",
            "root_cause": "权限不足",
            "suggestion": "检查用户权限或使用 sudo 执行"
        },
        "resource_busy": {
            "pattern": r"Device or resource busy|资源忙",
            "root_cause": "资源被占用",
            "suggestion": "检查是否有其他进程占用资源"
        },
        "timeout": {
            "pattern": r"timeout|超时|timed out",
            "root_cause": "操作超时",
            "suggestion": "检查系统负载或增加超时时间"
        },
        "memory_error": {
            "pattern": r"Out of memory|内存不足|Cannot allocate memory",
            "root_cause": "内存不足",
            "suggestion": "检查系统内存使用或释放部分内存"
        },
        "network_error": {
            "pattern": r"Network is unreachable|网络不可达|Connection refused",
            "root_cause": "网络问题",
            "suggestion": "检查网络配置和防火墙规则"
        },
        "kvm_error": {
            "pattern": r"KVM|virtualization|虚拟化",
            "root_cause": "虚拟化问题",
            "suggestion": "检查 KVM 模块和 BIOS 虚拟化设置"
        }
    }

    def __init__(self, results_file: str):
        with open(results_file) as f:
            self.results = json.load(f)

    def analyze_failures(self) -> List[FailureAnalysis]:
        """分析失败用例"""
        analyses = []

        for result in self.results:
            if result['status'] == 'FAIL':
                error_log = self._load_error_log(result['name'])
                analysis = self._analyze_error(result['name'], error_log)
                analyses.append(analysis)

        return analyses

    def _load_error_log(self, test_name: str) -> str:
        """加载错误日志"""
        try:
            with open(f"logs/{test_name}.log") as f:
                return f.read()
        except:
            return ""

    def _analyze_error(self, test_name: str, error_log: str) -> FailureAnalysis:
        """分析单个错误"""
        for error_type, config in self.ERROR_PATTERNS.items():
            if re.search(config['pattern'], error_log, re.IGNORECASE):
                return FailureAnalysis(
                    test_name=test_name,
                    error_type=error_type,
                    root_cause=config['root_cause'],
                    suggestion=config['suggestion'],
                    related_issues=self._find_related_issues(error_type)
                )

        # 未知错误类型
        return FailureAnalysis(
            test_name=test_name,
            error_type="unknown",
            root_cause="需要进一步分析",
            suggestion="查看详细日志进行人工分析",
            related_issues=[]
        )

    def _find_related_issues(self, error_type: str) -> List[str]:
        """查找相关问题"""
        # 可以关联到内部 issue tracker
        return []

    def generate_summary(self) -> Dict:
        """生成测试摘要"""
        total = len(self.results)
        passed = sum(1 for r in self.results if r['status'] == 'PASS')
        failed = total - passed

        by_priority = defaultdict(lambda: {'pass': 0, 'fail': 0})
        for r in self.results:
            if r['status'] == 'PASS':
                by_priority[r['priority']]['pass'] += 1
            else:
                by_priority[r['priority']]['fail'] += 1

        return {
            'total': total,
            'passed': passed,
            'failed': failed,
            'pass_rate': f"{passed/total*100:.1f}%",
            'by_priority': dict(by_priority)
        }
```

### 2. 报告模板

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>冒烟测试报告 - {{ date }}</title>
    <style>
        :root {
            --pass-color: #10b981;
            --fail-color: #ef4444;
            --warn-color: #f59e0b;
            --info-color: #3b82f6;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #f5f5f5;
        }
        .card {
            background: white;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .summary-grid {
            display: grid;
            grid-template-columns: repeat(4, 1fr);
            gap: 16px;
        }
        .stat-card {
            text-align: center;
            padding: 20px;
            background: linear-gradient(135deg, var(--info-color), #6366f1);
            color: white;
            border-radius: 8px;
        }
        .stat-card.pass { background: linear-gradient(135deg, var(--pass-color), #059669); }
        .stat-card.fail { background: linear-gradient(135deg, var(--fail-color), #dc2626); }
        .stat-value { font-size: 36px; font-weight: bold; }
        .stat-label { font-size: 14px; opacity: 0.9; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #eee; }
        th { background: #f9f9f9; font-weight: 600; }
        .badge { padding: 4px 8px; border-radius: 4px; font-size: 12px; }
        .badge-P0 { background: #fef2f2; color: #991b1b; }
        .badge-P1 { background: #fff7ed; color: #9a3412; }
        .badge-P2 { background: #fefce8; color: #854d0e; }
        .status-pass { color: var(--pass-color); }
        .status-fail { color: var(--fail-color); }
        .progress-bar {
            height: 8px;
            background: #e5e7eb;
            border-radius: 4px;
            overflow: hidden;
        }
        .progress-fill {
            height: 100%;
            background: var(--pass-color);
        }
    </style>
</head>
<body>
    <h1>🧪 冒烟测试报告</h1>
    <p>执行时间: {{ date }} | 环境: {{ environment }}</p>

    <!-- 摘要卡片 -->
    <div class="summary-grid">
        <div class="stat-card">
            <div class="stat-value">{{ total }}</div>
            <div class="stat-label">总用例</div>
        </div>
        <div class="stat-card pass">
            <div class="stat-value">{{ passed }}</div>
            <div class="stat-label">通过</div>
        </div>
        <div class="stat-card fail">
            <div class="stat-value">{{ failed }}</div>
            <div class="stat-label">失败</div>
        </div>
        <div class="stat-card">
            <div class="stat-value">{{ duration }}</div>
            <div class="stat-label">总耗时</div>
        </div>
    </div>

    <!-- 通过率 -->
    <div class="card">
        <h3>通过率</h3>
        <div class="progress-bar">
            <div class="progress-fill" style="width: {{ pass_rate }}%"></div>
        </div>
        <p style="text-align: right;">{{ pass_rate }}%</p>
    </div>

    <!-- 详细结果 -->
    <div class="card">
        <h3>测试结果详情</h3>
        <table>
            <thead>
                <tr>
                    <th>用例名称</th>
                    <th>优先级</th>
                    <th>状态</th>
                    <th>耗时</th>
                    <th>信息</th>
                </tr>
            </thead>
            <tbody>
                {% for result in results %}
                <tr>
                    <td>{{ result.name }}</td>
                    <td><span class="badge badge-{{ result.priority }}">{{ result.priority }}</span></td>
                    <td class="status-{{ result.status|lower }}">{{ result.status }}</td>
                    <td>{{ result.duration }}s</td>
                    <td>{{ result.message }}</td>
                </tr>
                {% endfor %}
            </tbody>
        </table>
    </div>

    <!-- 失败分析 -->
    {% if failures %}
    <div class="card">
        <h3>❌ 失败分析</h3>
        {% for failure in failures %}
        <div style="margin-bottom: 16px; padding: 16px; background: #fef2f2; border-radius: 8px;">
            <h4>{{ failure.test_name }}</h4>
            <p><strong>错误类型:</strong> {{ failure.error_type }}</p>
            <p><strong>根因:</strong> {{ failure.root_cause }}</p>
            <p><strong>建议:</strong> {{ failure.suggestion }}</p>
        </div>
        {% endfor %}
    </div>
    {% endif %}
</body>
</html>
```

---

## CI/CD 集成

### 1. Jenkins Pipeline

```groovy
// Jenkinsfile
pipeline {
    agent {
        label 'huawei-cloud-runner'  // 华为云构建节点
    }

    environment {
        TEST_ENV = 'smoke-test'
        REPORT_DIR = "reports/${BUILD_NUMBER}"
    }

    stages {
        stage('环境准备') {
            steps {
                sh '''
                    echo "========== 检查测试环境 =========="
                    bash scripts/env_check.sh
                '''
            }
        }

        stage('P0 测试') {
            steps {
                sh '''
                    echo "========== 执行 P0 测试 =========="
                    bash scripts/run_p0_tests.sh
                '''
            }
            post {
                failure {
                    echo "P0 测试失败，终止构建"
                    script {
                        currentBuild.result = 'FAILURE'
                    }
                }
            }
        }

        stage('P1 测试') {
            steps {
                sh '''
                    echo "========== 执行 P1 测试 =========="
                    bash scripts/run_p1_tests.sh || true
                '''
            }
        }

        stage('生成报告') {
            steps {
                sh '''
                    python3 scripts/generate_report.py \
                        --results "${REPORT_DIR}/results.json" \
                        --output "${REPORT_DIR}/report.html"
                '''
                archiveArtifacts artifacts: "${REPORT_DIR}/**", fingerprint: true
                publishHTML(target: [
                    allowMissing: false,
                    alwaysLinkToLastBuild: true,
                    keepAll: true,
                    reportDir: "${REPORT_DIR}",
                    reportFiles: 'report.html',
                    reportName: 'Smoke Test Report'
                ])
            }
        }

        stage('结果通知') {
            steps {
                script {
                    def result = currentBuild.result ?: 'SUCCESS'
                    def color = result == 'SUCCESS' ? 'green' : 'red'

                    // 企业微信通知
                    sh """
                        curl -X POST "${WEBHOOK_URL}" \
                            -H 'Content-Type: application/json' \
                            -d '{
                                "msgtype": "markdown",
                                "markdown": {
                                    "content": "## 冒烟测试完成\\n> 状态: ${result}\\n> 构建: #${BUILD_NUMBER}\\n> [查看报告](${BUILD_URL}Smoke_20Test_20Report/)"
                                }
                            }'
                    """
                }
            }
        }
    }

    post {
        always {
            cleanWs()
        }
    }
}
```

### 2. GitLab CI

```yaml
# .gitlab-ci.yml

stages:
  - prepare
  - test
  - report

variables:
  TEST_ENV: "smoke-test"
  REPORT_DIR: "reports"

# 环境准备
env_check:
  stage: prepare
  tags:
    - huawei-cloud
  script:
    - bash scripts/env_check.sh
  artifacts:
    paths:
      - env_check.log
    expire_in: 1 day

# P0 测试
p0_tests:
  stage: test
  tags:
    - huawei-cloud
  needs:
    - env_check
  script:
    - bash scripts/run_p0_tests.sh
  artifacts:
    paths:
      - reports/
    expire_in: 7 days
    when: always

# P1 测试
p1_tests:
  stage: test
  tags:
    - huawei-cloud
  needs:
    - env_check
  script:
    - bash scripts/run_p1_tests.sh
  allow_failure: true
  artifacts:
    paths:
      - reports/
    expire_in: 7 days
    when: always

# 生成报告
generate_report:
  stage: report
  tags:
    - huawei-cloud
  needs:
    - p0_tests
    - p1_tests
  script:
    - python3 scripts/generate_report.py
  artifacts:
    paths:
      - reports/report.html
    expire_in: 30 days

# 通知
notify:
  stage: report
  tags:
    - huawei-cloud
  needs:
    - generate_report
  script:
    - |
      curl -X POST "${WEBHOOK_URL}" \
        -H 'Content-Type: application/json' \
        -d "{\"msgtype\": \"text\", \"text\": {\"content\": \"冒烟测试完成: ${CI_PIPELINE_STATUS}\"}}"
  when: always
```

---

## AI 辅助测试

### 1. AI 辅助测试生成

```python
#!/usr/bin/env python3
# ai_test_generator.py

from anthropic import Anthropic
import os

client = Anthropic(api_key=os.environ.get("ANTHROPIC_API_KEY"))

def generate_test_case(feature_description: str, test_type: str = "smoke") -> str:
    """使用 AI 生成测试用例"""

    prompt = f"""
    请根据以下功能描述生成冒烟测试用例：

    功能描述：{feature_description}

    要求：
    1. 测试脚本使用 Bash 编写
    2. 包含前置条件检查
    3. 包含清理逻辑
    4. 有明确的通过/失败判断
    5. 输出格式符合测试框架规范

    请输出完整的测试脚本。
    """

    response = client.messages.create(
        model="claude-sonnet-4-20250514",
        max_tokens=4096,
        messages=[{"role": "user", "content": prompt}]
    )

    return response.content[0].text

def analyze_test_failure(log_content: str) -> dict:
    """使用 AI 分析测试失败原因"""

    prompt = f"""
    请分析以下测试失败日志，找出根本原因：

    日志内容：
    {log_content}

    请输出：
    1. 错误类型
    2. 根本原因
    3. 修复建议
    4. 相关代码位置（如果能确定）
    """

    response = client.messages.create(
        model="claude-sonnet-4-20250514",
        max_tokens=2048,
        messages=[{"role": "user", "content": prompt}]
    )

    return {
        "analysis": response.content[0].text,
        "model": "claude-sonnet-4-20250514"
    }

# 使用示例
if __name__ == "__main__":
    # 生成测试用例
    test_script = generate_test_case(
        "KVM 虚拟机热迁移功能，支持在同一宿主机内迁移虚拟机"
    )
    print(test_script)

    # 分析失败
    # analysis = analyze_test_failure(open("test.log").read())
    # print(analysis["analysis"])
```

### 2. MCP 工具集成

```python
#!/usr/bin/env python3
# mcp_smoke_test_server.py

from mcp.server.fastmcp import FastMCP
import subprocess
import json

mcp = FastMCP("SmokeTest")

@mcp.tool()
def run_smoke_test(module: str, quick: bool = True) -> dict:
    """运行冒烟测试

    Args:
        module: 测试模块 (kvm/network/storage/kernel)
        quick: 是否快速测试
    """
    cmd = f"bash scripts/run_smoke_tests.sh --module {module}"
    if quick:
        cmd += " --quick"

    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)

    return {
        "success": result.returncode == 0,
        "output": result.stdout,
        "errors": result.stderr
    }

@mcp.tool()
def get_test_report(build_number: int) -> dict:
    """获取测试报告

    Args:
        build_number: 构建号
    """
    report_path = f"reports/{build_number}/results.json"
    try:
        with open(report_path) as f:
            return json.load(f)
    except FileNotFoundError:
        return {"error": f"Report not found for build {build_number}"}

@mcp.tool()
def analyze_failure(test_name: str, log_path: str) -> dict:
    """分析测试失败

    Args:
        test_name: 测试名称
        log_path: 日志文件路径
    """
    try:
        with open(log_path) as f:
            log_content = f.read()

        # 使用 AI 分析
        from ai_test_generator import analyze_test_failure
        return analyze_test_failure(log_content)
    except Exception as e:
        return {"error": str(e)}

@mcp.resource("test://history/{limit}")
def get_test_history(limit: int = 10) -> str:
    """获取测试历史记录"""
    import glob
    reports = sorted(glob.glob("reports/*/results.json"), reverse=True)[:limit]

    history = []
    for report in reports:
        with open(report) as f:
            history.append(json.load(f))

    return json.dumps(history, indent=2)

if __name__ == "__main__":
    mcp.run()
```

---

## 附录

### A. 常见问题排查

| 问题 | 可能原因 | 排查步骤 |
|------|---------|---------|
| VM 启动失败 | KVM 模块未加载 | `lsmod \| grep kvm` |
| 网络不通 | 网桥配置错误 | `brctl show` |
| I/O 超时 | 磁盘性能问题 | `iostat -x 1` |
| 内存不足 | 资源竞争 | `free -h` |

### B. 测试环境规格

| 组件 | 最低要求 | 推荐配置 |
|------|---------|---------|
| CPU | 4 核 | 8+ 核 |
| 内存 | 8 GB | 32+ GB |
| 存储 | 50 GB | 500+ GB SSD |
| 网络 | 1 Gbps | 10+ Gbps |

### C. 联系方式

- 测试框架问题：测试团队
- 环境问题：运维团队
- 业务逻辑问题：开发团队
