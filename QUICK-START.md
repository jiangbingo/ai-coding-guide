# 🚀 快速开始指南

## ✅ 已完成的工作

### 1. 测试框架配置

**已创建文件**：
- `package.json` - 项目配置和依赖
- `tests/unit/typewriter.test.js` - 单元测试示例
- `tests/system/navigation.test.js` - 系统测试示例

**测试类型**：
- ✅ UT（单元测试）- 测试 JavaScript 函数
- ✅ SCT（系统测试）- 测试完整用户流程
- 🚧 MCT（模块测试）- 测试 Skills 集成（待实现）

### 2. MCP 集成

**已创建文件**：
- `mcp/src/server.ts` - MCP 服务器实现

**已实现功能**：
- ✅ 5 个工具（KVM 优化、性能分析、内核调试、NUMA 分析、VirtIO 配置）
- ✅ 3 个资源（内核文档、KVM 最佳实践、性能工具参考）
- ✅ 2 个提示词（KVM 调优、内核调试工作流）

---

## 📦 安装依赖

```bash
cd /Users/jiangbin/.openclaw/workspace/ai-coding-guide

# 安装所有依赖
npm install

# 安装 Playwright 浏览器
npx playwright install
```

---

## 🧪 运行测试

### 单元测试（UT）

```bash
# 运行所有单元测试
npm run test:unit

# 运行单个测试文件
npx jest tests/unit/typewriter.test.js

# 生成覆盖率报告
npm run test:coverage
```

**预期输出**：
```
PASS  tests/unit/typewriter.test.js
  Typewriter Effect
    ✓ should type text character by character (52ms)
    ✓ should handle empty string (5ms)
    ✓ should handle single character (3ms)
    ✓ should respect speed parameter (205ms)
    ✓ should handle null element gracefully (1ms)

Test Suites: 1 passed, 1 total
Tests:       5 passed, 5 total
```

### 系统测试（SCT）

```bash
# 运行所有系统测试
npm run test:system

# 运行单个测试文件
npx playwright test tests/system/navigation.test.js

# 以 UI 模式运行
npx playwright test --ui
```

**预期输出**：
```
Running 6 tests using 1 worker

✓ navigation.test.js:3:1 › Navigation Flow › should load homepage successfully (1s)
✓ navigation.test.js:10:3 › Navigation Flow › should navigate to section on keyboard press (524ms)
✓ navigation.test.js:21:3 › Navigation Flow › should toggle accordion on click (645ms)

6 passed (3s)
```

### 所有测试

```bash
# 运行所有测试
npm test

# 监视模式
npm run test:watch
```

---

## 🔧 启动 MCP Server

### 构建

```bash
# 编译 TypeScript
npm run build:mcp
```

### 测试连接

```bash
# 方法 1: 直接运行
npm run mcp:start

# 方法 2: 使用 Claude Code
# 在 Claude Code 中配置 MCP 服务器
```

### 在 Claude Code 中配置

编辑 `~/.claude/config.json`：

```json
{
  "mcpServers": {
    "huawei-kernel-tools": {
      "command": "node",
      "args": ["/Users/jiangbin/.openclaw/workspace/ai-coding-guide/mcp/dist/server.js"]
    }
  }
}
```

### 验证 MCP 集成

```bash
# 运行验证脚本
npm run mcp:validate
```

---

## 🎯 下一步工作

### Week 1: 完善测试框架

**Day 1-2**：
- [ ] 添加更多单元测试（手风琴、键盘快捷键、移动端菜单）
- [ ] 提高测试覆盖率到 80%

**Day 3-4**：
- [ ] 添加更多系统测试（响应式布局、性能测试）
- [ ] 配置 CI/CD 流水线

**Day 5**：
- [ ] 实现 MCT（模块测试）框架
- [ ] 编写 Skills 集成测试

### Week 2: 完善 MCP 集成

**Day 1-2**：
- [ ] 实现真实的资源加载（从文件或 API）
- [ ] 添加更多工具（DPDK、eBPF 等）

**Day 3-4**：
- [ ] 编写 MCP 测试
- [ ] 优化错误处理

**Day 5**：
- [ ] 集成到 CI/CD
- [ ] 编写文档

### Week 3: 文档和优化

**Day 1-2**：
- [ ] 编写测试框架文档
- [ ] 编写 MCP 集成文档

**Day 3-4**：
- [ ] 性能优化
- [ ] 提高测试覆盖率

**Day 5**：
- [ ] 版本发布
- [ ] 社区推广

---

## 📊 当前状态

### 测试覆盖率

| 类型 | 当前 | 目标 |
|------|------|------|
| 单元测试 | 5 个测试 | 20+ 个测试 |
| 系统测试 | 6 个测试 | 15+ 个测试 |
| 模块测试 | 0 | 10+ 个测试 |
| MCP 测试 | 0 | 5+ 个测试 |

### MCP 工具

| 工具 | 状态 | 功能 |
|------|------|------|
| kvm_optimize | ✅ 完成 | KVM 性能优化 |
| perf_analyze | ✅ 完成 | 性能瓶颈分析 |
| kernel_debug | ✅ 完成 | 内核调试助手 |
| numa_analyze | ✅ 完成 | NUMA 亲和性分析 |
| virtio_configure | ✅ 完成 | VirtIO 设备配置 |

---

## 🐛 常见问题

### Q1: 测试失败怎么办？

```bash
# 查看详细错误信息
npx jest --verbose

# 更新快照
npx jest --updateSnapshot
```

### Q2: Playwright 浏览器未安装？

```bash
# 安装浏览器
npx playwright install

# 安装系统依赖
npx playwright install-deps
```

### Q3: MCP Server 无法启动？

```bash
# 检查 TypeScript 编译
npm run build:mcp

# 检查依赖
npm install

# 查看错误日志
node mcp/dist/server.js
```

---

## 📚 参考资源

### 测试框架
- [Jest Documentation](https://jestjs.io/)
- [Playwright Documentation](https://playwright.dev/)
- [Testing Library](https://testing-library.com/)

### MCP
- [Model Context Protocol Spec](https://modelcontextprotocol.io/)
- [MCP TypeScript SDK](https://github.com/modelcontextprotocol/typescript-sdk)

### 项目文档
- [开发计划](./DEVELOPMENT-PLAN.md)
- [华为云测试工作流](./docs/HUAWEI-CLOUD-TEST-WORKFLOW.md)
- [多 Agent 架构](./docs/MULTI-AGENT-ARCHITECTURE.md)

---

**创建时间**：2026-03-18
**状态**：✅ 测试框架和 MCP 集成已完成基础配置