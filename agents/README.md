# 华为云内核开发 Agent 系统 - 宱总配置

## 📋 系统概述

本文档定义了针对华为云杭州团队内核开发的完整 Agent 系统，配置和使用方法。

---

## 🤖 Agent 列表

### 1. UT Agent（单元测试 Agent）

**功能**：
- 分析内核模块代码结构
- 自动生成单元测试用例
- 执行测试并收集覆盖率
- 识别未测试的代码路径
- 自动修复失败的测试

**位置**：`agents/ut-agent/src/server.ts`

**MCP 工具**：
- `analyze_kernel_module` - 分析模块结构
- `generate_unit_tests` - 生成测试用例
- `execute_tests` - 执行测试
- `identify_untested_paths` - 识别未覆盖路径
- `auto_fix_test` - 自动修复测试

**使用示例**：
```
用户: 为 virtio_net.c 生成单元测试

UT Agent: 分析中...
发现 15 个可测试函数：
- virtio_net_open
- virtio_net_close
- virtio_net_send
- ...

已生成 45 个测试用例
覆盖率：85%
```

---

### 2. SCT Agent（系统组件测试 Agent）

**功能**：
- 识别组件依赖关系
- 生成集成测试场景
- 自动配置测试环境
- 执行冒烟测试
- 检测回归问题

**位置**：`agents/sct-agent/src/server.ts`

**MCP 工具**：
- `analyze_component_dependencies` - 分析依赖
- `generate_integration_tests` - 生成集成测试
- `prepare_test_environment` - 准备环境
- `execute_smoke_tests` - 执行冒烟测试
- `detect_regression` - 检测回归

**使用示例**：
```
用户: 对 virtio-net 组件进行系统测试

SCT Agent: 分析依赖...
virtio-net 依赖：
- virtio-core
- vhost-net
- qemu-kvm

生成测试场景：
1. 基础连通性测试
2. 多队列性能测试
3. 热迁移兼容性测试

执行测试：
✓ 连通性测试 (1/1 通过)
⚠️ 性能测试 (发现问题)
```

---

### 3. MCT Agent（模块/集成测试 Agent）

**功能**：
- 分析模块间接口
- 生成端到端测试
- 模拟真实工作负载
- 性能基准测试
- 稳定性测试

**位置**：`agents/mct-agent/src/server.ts`（待实现）

**MCP 工具**（计划）：
- `design_test_topology` - 设计测试拓扑
- `generate_e2e_tests` - 生成端到端测试
- `execute_workload_tests` - 执行工作负载测试
- `run_benchmark_tests` - 运行基准测试
- `stability_test` - 稳定性测试

---

### 4. Debug Agent（调试 Agent）

**功能**：
- 分析内核崩溃日志
- 识别死锁和竞态条件
- 内存泄漏检测
- 性能瓶颈定位
- 提供修复建议

**位置**：`agents/debug-agent/src/server.ts`（待实现）

**MCP 工具**（计划）：
- `analyze_kernel_crash` - 分析崩溃
- `detect_deadlock` - 检测死锁
- `detect_memory_leak` - 检测内存泄漏
- `identify_bottleneck` - 识别瓶颈
- `suggest_fix` - 建议修复

---

### 5. Perf Agent（性能分析 Agent）

**功能**：
- 性能数据采集（perf, eBPF）
- 瓶颈识别
- 优化建议生成
- 性能回归检测
- 基准对比

**位置**：`agents/perf-agent/src/server.ts`（待实现）

**MCP 工具**（计划）：
- `collect_perf_data` - 采集性能数据
- `analyze_performance` - 分析性能
- `identify_bottlenecks` - 识别瓶颈
- `generate_optimizations` - 生成优化建议
- `compare_baselines` - 基准对比

---

## 🔧 MCP 配置

### Claude Code 配置

编辑 `~/.claude/config.json`：

```json
{
  "mcpServers": {
    "ut-agent": {
      "command": "node",
      "args": ["/Users/jiangbin/.openclaw/workspace/ai-coding-guide/agents/ut-agent/dist/server.js"],
      "env": {
        "KERNEL_SRC": "/path/to/kernel/source",
        "TEST_FRAMEWORK": "kunit"
      }
    },
    "sct-agent": {
      "command": "node",
      "args": ["/Users/jiangbin/.openclaw/workspace/ai-coding-guide/agents/sct-agent/dist/server.js"],
      "env": {
        "LIBVIRT_URI": "qemu:///system",
        "TEST_ENV": "vm"
      }
    },
    "mct-agent": {
      "command": "node",
      "args": ["/Users/jiangbin/.openclaw/workspace/ai-coding-guide/agents/mct-agent/dist/server.js"],
      "env": {
        "TEST_DURATION": "24h",
        "WORKLOAD_TYPE": "mixed"
      }
    },
    "debug-agent": {
      "command": "node",
      "args": ["/Users/jiangbin/.openclaw/workspace/ai-coding-guide/agents/debug-agent/dist/server.js"],
      "env": {
        "KERNEL_VERSION": "5.10.0",
        "CRASH_DIR": "/var/crash"
      }
    },
    "perf-agent": {
      "command": "node",
      "args": ["/Users/jiangbin/.openclaw/workspace/ai-coding-guide/agents/perf-agent/dist/server.js"],
      "env": {
        "PERF_DATA_DIR": "/var/perf",
        "BASELINE_VERSION": "v1.0.0"
      }
    }
  }
}
```

---

## 🚀 快速开始

### 1. 构建 Agent

```bash
cd /Users/jiangbin/.openclaw/workspace/ai-coding-guide

# 构建 UT Agent
cd agents/ut-agent
npm install
npm run build

# 构建 SCT Agent
cd ../sct-agent
npm install
npm run build

# 构建 MCT Agent（待实现）
# cd ../mct-agent
# npm install
# npm run build

# 构建 Debug Agent（待实现）
# cd ../debug-agent
# npm install
# npm run build

# 构建 Perf Agent（待实现）
# cd ../perf-agent
# npm install
# npm run build
```

### 2. 配置 Claude Code

```bash
# 编辑配置文件
vim ~/.claude/config.json

# 重启 Claude Code
claude-code restart
```

### 3. 测试 Agent

```bash
# 测试 UT Agent
echo '测试 UT Agent' | node agents/ut-agent/dist/server.js

# 测试 SCT Agent
echo '测试 SCT Agent' | node agents/sct-agent/dist/server.js
```

---

## 📊 使用场景

### 场景 1：新功能开发

```
1. 使用 UT Agent 生成单元测试
2. 使用 SCT Agent 测试组件集成
3. 使用 MCT Agent 进行端到端测试
4. 使用 Perf Agent 验证性能
```

### 场景 2：Bug 修复

```
1. 使用 Debug Agent 分析问题
2. 使用 UT Agent 添加回归测试
3. 使用 SCT Agent 验证修复
4. 使用 Perf Agent 检查性能影响
```

### 场景 3：性能优化

```
1. 使用 Perf Agent 分析性能
2. 使用 Debug Agent 识别瓶颈
3. 使用 MCT Agent 验证优化效果
4. 使用 SCT Agent 检查回归
```

---

## 📚 相关文档

- [华为云 Agent 系统设计](./HUAWEI-CLOUD-AGENT-SYSTEM.md)
- [华为云测试工作流](./HUAWEI-CLOUD-TEST-WORKFLOW.md)
- [多 Agent 协作架构](./MULTI-AGENT-ARCHITECTURE.md)

---

**创建时间**：2026-03-18
**状态**：UT Agent ✅ 完成，SCT Agent ✅ 完成，其他 Agent 🚧 开发中