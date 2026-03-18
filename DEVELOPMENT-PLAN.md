# ai-coding-guide 项目开发计划

## 📋 项目概况

**项目名称**：AI 代码入门指南
**项目类型**：静态网站 + OpenClaw Skills 集合
**当前状态**：基础框架完成，需要添加测试框架和 MCP 集成
**目标用户**：资深 C/云计算开发者（10+ 年经验，AI 零基础）

---

## 🎯 开发目标

### Phase 1: 测试框架集成（UT/SCT/MCT）

#### 1.1 单元测试（UT - Unit Testing）

**目标**：为现有的 JavaScript 功能添加单元测试

**测试范围**：
```javascript
// 需要测试的模块
js/main.js
├── 打字机效果函数
├── 滚动动画触发器
├── 手风琴折叠逻辑
├── 键盘快捷键处理
└── 移动端菜单切换
```

**技术选型**：
- **Jest** - 测试框架
- **@testing-library/dom** - DOM 测试工具
- **jsdom** - 浏览器环境模拟

**示例测试**：
```javascript
// tests/unit/typewriter.test.js
describe('Typewriter Effect', () => {
  test('should type text character by character', async () => {
    const element = document.createElement('span');
    await typewriterEffect(element, 'Hello', 50);
    expect(element.textContent).toBe('Hello');
  });

  test('should handle empty string', async () => {
    const element = document.createElement('span');
    await typewriterEffect(element, '', 50);
    expect(element.textContent).toBe('');
  });
});
```

#### 1.2 系统测试（SCT - System Component Testing）

**目标**：测试完整的用户流程和组件交互

**测试范围**：
```yaml
系统级测试:
  - 导航流程测试
  - 章节跳转测试
  - 响应式布局测试
  - 性能测试（加载时间）
  - 跨浏览器兼容性测试
```

**技术选型**：
- **Playwright** - 端到端测试
- **Lighthouse** - 性能测试
- **BrowserStack** - 跨浏览器测试

**示例测试**：
```javascript
// tests/system/navigation.test.js
describe('Navigation Flow', () => {
  test('should navigate to section on keyboard press', async () => {
    await page.goto('http://localhost:8000');
    await page.keyboard.press('1');
    await page.waitForSelector('#concepts');
    const isVisible = await page.isVisible('#concepts');
    expect(isVisible).toBe(true);
  });

  test('should toggle accordion on click', async () => {
    await page.goto('http://localhost:8000');
    await page.click('.scenario-header:first-child');
    const content = await page.$eval('.scenario-content', el => 
      window.getComputedStyle(el).display
    );
    expect(content).toBe('block');
  });
});
```

#### 1.3 模块测试（MCT - Module/Integration Testing）

**目标**：测试 Skills 模块的集成和协作

**测试范围**：
```yaml
Skills 集成测试:
  - Skill 加载测试
  - Skill 触发词匹配测试
  - Skill 输出格式测试
  - 多 Skill 协作测试
  - Skills 依赖关系测试
```

**技术选型**：
- **自定义测试框架** - 基于 Node.js
- **Mock LLM** - 模拟 AI 响应
- **Snapshot Testing** - 输出一致性验证

**示例测试**：
```javascript
// tests/module/skills-integration.test.js
describe('Skills Integration', () => {
  test('should load all skills correctly', async () => {
    const skills = await loadSkills('./skills');
    expect(skills.length).toBeGreaterThan(0);
    skills.forEach(skill => {
      expect(skill).toHaveProperty('name');
      expect(skill).toHaveProperty('description');
      expect(skill).toHaveProperty('triggers');
    });
  });

  test('should trigger KVM skill on keyword', async () => {
    const input = '帮我优化 KVM 虚拟机性能';
    const skill = matchSkill(input);
    expect(skill.name).toBe('kvm-virt-optimization');
  });

  test('skills should chain correctly', async () => {
    const workflow = await executeSkillChain([
      'kernel-debug-assistant',
      'kunpeng-perf-tuning'
    ]);
    expect(workflow.success).toBe(true);
  });
});
```

---

### Phase 2: MCP（Model Context Protocol）集成

#### 2.1 MCP 架构设计

**目标**：将 Skills 转换为 MCP 服务器，支持标准化工具调用

**架构**：
```
┌─────────────────────────────────────────────────────────────┐
│                     MCP Architecture                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   ┌─────────────┐       ┌─────────────┐                   │
│   │ Claude Code │◄─────►│ MCP Client  │                   │
│   │   (用户)     │       │  (协议层)    │                   │
│   └─────────────┘       └─────────────┘                   │
│                               │                             │
│                               ▼                             │
│   ┌─────────────────────────────────────────────────────┐ │
│   │                  MCP Server                          │ │
│   │  ┌────────────────────────────────────────────────┐ │ │
│   │  │            Tool Registry                        │ │ │
│   │  │  ┌─────────┐ ┌─────────┐ ┌─────────┐         │ │ │
│   │  │  │ KVM Opt │ │ Perf    │ │ Kernel  │         │ │ │
│   │  │  │ Tool    │ │ Analysis│ │ Debug   │         │ │ │
│   │  │  └─────────┘ └─────────┘ └─────────┘         │ │ │
│   │  └────────────────────────────────────────────────┘ │ │
│   │                                                       │ │
│   │  ┌────────────────────────────────────────────────┐ │ │
│   │  │            Resource Provider                    │ │ │
│   │  │  ┌─────────┐ ┌─────────┐ ┌─────────┐         │ │ │
│   │  │  │ Kernel  │ │ Perf    │ │ Config  │         │ │ │
│   │  │  │ Docs    │ │ Data    │ │ Files   │         │ │ │
│   │  │  └─────────┘ └─────────┘ └─────────┘         │ │ │
│   │  └────────────────────────────────────────────────┘ │ │
│   └─────────────────────────────────────────────────────┘ │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

#### 2.2 MCP Server 实现

**文件结构**：
```
mcp/
├── package.json
├── tsconfig.json
├── src/
│   ├── server.ts           # MCP 服务器入口
│   ├── tools/
│   │   ├── kvm-optimizer.ts
│   │   ├── perf-analyzer.ts
│   │   └── kernel-debugger.ts
│   ├── resources/
│   │   ├── kernel-docs.ts
│   │   └── perf-data.ts
│   └── prompts/
│       ├── kvm-tuning.ts
│       └── numa-analysis.ts
└── tests/
    ├── tools/
    └── resources/
```

**示例代码**：
```typescript
// src/server.ts
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';

const server = new Server({
  name: 'huawei-cloud-kernel-tools',
  version: '1.0.0',
}, {
  capabilities: {
    tools: {},
    resources: {},
    prompts: {},
  },
});

// 注册工具
server.setRequestHandler(ListToolsRequestSchema, async () => {
  return {
    tools: [
      {
        name: 'kvm_optimize',
        description: '优化 KVM 虚拟机性能',
        inputSchema: {
          type: 'object',
          properties: {
            vm_type: { type: 'string', enum: ['kvm', 'qemu', 'xen'] },
            workload: { type: 'string', enum: ['cpu-intensive', 'memory-intensive', 'io-intensive'] },
          },
          required: ['vm_type'],
        },
      },
    ],
  };
});

// 工具调用处理
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;
  
  if (name === 'kvm_optimize') {
    const result = await optimizeKVM(args);
    return {
      content: [{
        type: 'text',
        text: JSON.stringify(result, null, 2),
      }],
    };
  }
  
  throw new Error(`Unknown tool: ${name}`);
});

// 启动服务器
const transport = new StdioServerTransport();
await server.connect(transport);
```

#### 2.3 MCP 工具示例

```typescript
// src/tools/kvm-optimizer.ts
export async function optimizeKVM(args: {
  vm_type: string;
  workload?: string;
}): Promise<any> {
  const recommendations = [];
  
  // CPU 优化
  if (args.workload === 'cpu-intensive') {
    recommendations.push({
      category: 'CPU',
      priority: 'HIGH',
      items: [
        '启用 vCPU 绑定（CPU pinning）',
        '禁用 CPU 超额订阅',
        '使用 host-passthrough 模式',
      ],
    });
  }
  
  // 内存优化
  recommendations.push({
    category: 'Memory',
    priority: 'MEDIUM',
    items: [
      '启用大页内存（HugePages）',
      '配置 NUMA 亲和性',
      '禁用 KSM（Kernel Samepage Merging）',
    ],
  });
  
  // 网络 I/O 优化
  if (args.workload === 'io-intensive') {
    recommendations.push({
      category: 'Network',
      priority: 'HIGH',
      items: [
        '使用 virtio-net + vhost-net',
        '启用多队列（multi-queue）',
        '配置 RSS（Receive Side Scaling）',
      ],
    });
  }
  
  return {
    summary: `KVM ${args.vm_type} 优化建议`,
    recommendations,
    generatedAt: new Date().toISOString(),
  };
}
```

#### 2.4 MCP 资源示例

```typescript
// src/resources/kernel-docs.ts
export const kernelDocsResource = {
  uri: 'kernel://docs',
  name: 'Linux Kernel Documentation',
  description: 'Linux 内核文档资源',
  mimeType: 'text/markdown',
  
  async read() {
    const docs = await fetchKernelDocs();
    return {
      contents: [{
        uri: 'kernel://docs',
        mimeType: 'text/markdown',
        text: docs,
      }],
    };
  },
};

// 注册资源
server.setRequestHandler(ListResourcesRequestSchema, async () => {
  return {
    resources: [
      {
        uri: 'kernel://docs',
        name: 'Linux Kernel Documentation',
        mimeType: 'text/markdown',
      },
      {
        uri: 'perf://data',
        name: 'Performance Metrics',
        mimeType: 'application/json',
      },
    ],
  };
});
```

#### 2.5 MCP Prompts 示例

```typescript
// src/prompts/kvm-tuning.ts
export const kvmTuningPrompt = {
  name: 'kvm-tuning',
  description: 'KVM 虚拟机性能调优提示',
  arguments: [
    {
      name: 'workload_type',
      description: '工作负载类型',
      required: true,
    },
  ],
  
  async generate(args: { workload_type: string }) {
    return {
      description: `KVM ${args.workload_type} 工作负载调优`,
      messages: [
        {
          role: 'user',
          content: {
            type: 'text',
            text: `请针对 ${args.workload_type} 工作负载，提供 KVM 虚拟机性能调优建议。

重点关注：
1. CPU 调度和绑定策略
2. 内存分配和 NUMA 亲和性
3. 网络 I/O 优化（virtio 配置）
4. 存储 I/O 优化（磁盘调度器）

请提供具体的配置示例和验证方法。`,
          },
        },
      ],
    };
  },
};
```

---

### Phase 3: CI/CD 集成

#### 3.1 测试流水线

```yaml
# .github/workflows/test.yml
name: Test Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: npm ci
      - run: npm run test:unit
      - uses: codecov/codecov-action@v3

  system-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      - run: npm ci
      - run: npm run test:system
      - uses: actions/upload-artifact@v3
        with:
          name: playwright-report
          path: playwright-report/

  mcp-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      - run: npm ci
      - run: npm run test:mcp
      - run: npm run mcp:validate

  deploy:
    needs: [unit-tests, system-tests, mcp-tests]
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm run build
      - uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./dist
```

#### 3.2 测试覆盖率目标

```yaml
Coverage Targets:
  Unit Tests:
    - Statements: 80%
    - Branches: 75%
    - Functions: 80%
    - Lines: 80%
  
  System Tests:
    - Critical Paths: 100%
    - User Flows: 90%
    - Edge Cases: 70%
  
  MCP Integration:
    - Tool Coverage: 100%
    - Resource Coverage: 80%
    - Prompt Coverage: 60%
```

---

## 📅 实施计划

### Week 1: 测试框架搭建

**Day 1-2: UT 框架**
- 安装 Jest 和相关依赖
- 编写第一个单元测试
- 配置测试覆盖率报告

**Day 3-4: SCT 框架**
- 安装 Playwright
- 编写第一个系统测试
- 配置 CI/CD 流水线

**Day 5: MCT 框架**
- 设计 Skills 测试框架
- 编写 Skills 加载测试
- 配置 Mock LLM

### Week 2: MCP 集成

**Day 1-2: MCP Server 基础**
- 初始化 MCP 项目
- 实现基础工具注册
- 测试 Claude Code 集成

**Day 3-4: 工具和资源**
- 实现 KVM 优化工具
- 实现性能分析工具
- 实现文档资源

**Day 5: Prompts 和测试**
- 实现 Prompts
- 编写 MCP 测试
- 集成到 CI/CD

### Week 3: 文档和优化

**Day 1-2: 文档编写**
- 测试框架文档
- MCP 集成文档
- API 参考文档

**Day 3-4: 性能优化**
- 测试执行优化
- 覆盖率提升
- 文档补充

**Day 5: 发布**
- 版本发布
- 更新主网站
- 社区推广

---

## 📊 成功指标

### 测试指标

| 指标 | 当前 | 目标 |
|------|------|------|
| 单元测试覆盖率 | 0% | 80% |
| 系统测试通过率 | N/A | 95% |
| MCP 工具数量 | 0 | 8 |
| 测试执行时间 | N/A | < 5 分钟 |

### 质量指标

| 指标 | 目标 |
|------|------|
| Bug 发现率 | > 1 bug/测试 |
| 回归测试 | 100% 自动化 |
| 文档完整性 | 100% 覆盖 |
| 用户满意度 | > 4.5/5 |

---

## 🚀 快速开始

### 安装依赖

```bash
cd ai-coding-guide

# 安装测试依赖
npm install --save-dev \
  jest @types/jest ts-jest \
  @testing-library/dom jsdom \
  playwright @playwright/test

# 安装 MCP 依赖
npm install --save-dev \
  @modelcontextprotocol/sdk \
  typescript ts-node
```

### 运行测试

```bash
# 单元测试
npm run test:unit

# 系统测试
npm run test:system

# MCP 测试
npm run test:mcp

# 所有测试
npm test
```

### 启动 MCP Server

```bash
# 构建
npm run build:mcp

# 启动
npm run mcp:start

# 验证
npm run mcp:validate
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
- [Claude Code MCP Integration](https://docs.anthropic.com/claude-code/mcp)

### 华为云

- [华为云测试工作流](./docs/HUAWEI-CLOUD-TEST-WORKFLOW.md)
- [多 Agent 架构](./docs/MULTI-AGENT-ARCHITECTURE.md)

---

**创建时间**：2026-03-18
**预计完成**：2026-04-08（3 周）
**负责人**：OpenClaw AI Agent