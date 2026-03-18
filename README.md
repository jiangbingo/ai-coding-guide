# AI 代码入门指南

> 给资深 C/云计算开发者的 AI 编程实用手册

## 📖 项目简介

这是一个面向资深 C/云计算开发者（10+ 年经验，AI 零基础）的单页科普+实战网站。旨在帮助开发者快速上手 AI 编程工具，提升开发效率。

## ✨ 特色

- **零门槛入门**：无需 AI 背景，直接上手实战
- **场景适配**：专为 C/云计算开发场景定制
- **实用导向**：提供可直接复用的提示词模板
- **Neo-Terminal 美学**：深色终端风格，开发者友好

## 🛠️ 技术栈

- **HTML5** - 语义化标签
- **CSS3** - Flexbox/Grid 布局，CSS 变量，动画
- **JavaScript (ES6+)** - 原生 JS，无框架依赖
- **Prism.js** - 代码语法高亮
- **Google Fonts** - JetBrains Mono + IBM Plex Sans

## 📁 项目结构

```
ai-coding-guide/
├── index.html          # 主页面
├── css/
│   └── style.css       # 样式文件
├── js/
│   └── main.js         # 交互逻辑
├── assets/
│   └── icons/          # 工具图标（预留）
└── README.md           # 项目说明
```

## 🚀 快速开始

### 本地预览

```bash
# 方法 1: Python HTTP Server
cd ai-coding-guide
python3 -m http.server 8000
# 访问 http://localhost:8000

# 方法 2: Node.js serve
npx serve .
# 访问 http://localhost:3000

# 方法 3: 直接用浏览器打开
open index.html
```

### 部署到 GitHub Pages

1. 推送到 GitHub 仓库
2. 进入仓库 Settings → Pages
3. 选择 Branch: main, Folder: / (root)
4. 保存后等待部署完成

## 📱 响应式设计

- 桌面端：完整布局，双栏 Hero 区
- 平板端：自适应网格布局
- 移动端：单栏布局，折叠导航

## 🎨 设计系统

### 色彩

| 名称 | 色值 | 用途 |
|------|------|------|
| 背景主色 | `#0a0a0f` | 页面背景 |
| 背景辅色 | `#141419` | 卡片背景 |
| 文字主色 | `#e4e4e7` | 正文文字 |
| 琥珀色 | `#fbbf24` | 主要强调色 |
| 绿色 | `#10b981` | 成功/终端 |
| 蓝色 | `#3b82f6` | 链接/步骤 |
| 紫色 | `#8b5cf6` | AI 主题色 |

### 字体

- **等宽字体**: JetBrains Mono - 标题、代码
- **无衬线字体**: IBM Plex Sans - 正文

## ⌨️ 键盘快捷键

| 按键 | 功能 |
|------|------|
| `1-5` | 跳转到对应章节 |
| `ESC` | 关闭移动端菜单 |

## 📦 内容模块

1. **Hero 区** - 打字机动画效果
2. **基础概念** - 4 类 AI 工具介绍
3. **工具选型** - Claude Code 重点推荐
4. **实战流程** - 5 步开发流程
5. **场景示例** - 5 个 C/云计算场景
6. **资源推荐** - 学习资源链接

## 🚀 在线访问

- **主页面**: [https://jiangbingo.github.io/ai-coding-guide/](https://jiangbingo.github.io/ai-coding-guide/)
- **华为云 Agent 系统展示**: [https://jiangbingo.github.io/ai-coding-guide/agents-showcase.html](https://jiangbingo.github.io/ai-coding-guide/agents-showcase.html)

### 华为云内核开发 Agent 系统

基于 MCP 协议的多 Agent 协作系统，包含：

- 🧪 **UT Agent** - 单元测试自动化
- 🔧 **SCT Agent** - 系统组件测试
- 🔗 **MCT Agent** - 模块集成测试
- 🐛 **Debug Agent** - 内核调试
- ⚡ **Perf Agent** - 性能分析

详细文档：
- [Agent 系统设计](./docs/HUAWEI-CLOUD-AGENT-SYSTEM.md)
- [Agent 配置指南](./agents/README.md)
- [开发计划](./DEVELOPMENT-PLAN.md)

## 🔧 自定义

### 修改配色

编辑 `css/style.css` 中的 CSS 变量：

```css
:root {
  --bg-primary: #0a0a0f;
  --accent-amber: #fbbf24;
  /* ... */
}
```

### 添加新场景

在 `index.html` 的 `.scenario-accordion` 中添加：

```html
<div class="scenario-item">
  <button class="scenario-header">
    <!-- 场景标题 -->
  </button>
  <div class="scenario-content">
    <!-- 场景内容 -->
  </div>
</div>
```

## 📄 License

MIT License - 自由使用和修改

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

Made with ❤️ for developers
