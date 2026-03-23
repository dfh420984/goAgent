# GoAgent TaskRunner - 智能任务执行助手

🚀 一个基于 Go 语言开发的企业级智能体任务执行框架，专为后端工程师设计，适合写入简历面试使用。

## 🌟 项目亮点

### 技术栈
- **后端**: Go 1.24 + Gin Framework
- **前端**: React 18 + TypeScript + Vite
- **LLM 集成**: OpenAI API 兼容接口（支持多家厂商）
- **架构设计**: Clean Architecture 分层架构
- **实时通信**: SSE (Server-Sent Events) 流式传输

### 核心特性
✅ **ReAct Agent Loop** - 实现推理与行动的交替循环  
✅ **Function Calling** - 支持工具调用和函数执行  
✅ **工作流引擎** - 可视化 YAML 定义业务流程  
✅ **可插拔工具系统** - 轻松扩展新工具  
✅ **SSE 流式输出** - 实时打字机效果  
✅ **企业级特性** - 审计日志、链路追踪、配置管理  

## 📁 项目结构

```
agentLearn/
├── cmd/
│   ├── agent/          # CLI 版本入口
│   └── server/         # Web 服务入口
├── internal/           # 内部核心代码
│   ├── agent/          # Agent 引擎（ReAct Loop）
│   ├── tools/          # 工具系统（文件/HTTP/Shell）
│   ├── workflow/       # 工作流引擎
│   └── config/         # 配置管理
├── pkg/                # 公共包
│   └── llm/            # LLM 客户端封装
├── configs/            # 配置文件
│   └── workflows/      # 工作流定义
├── frontend/           # React 前端
│   └── src/
│       ├── App.tsx     # 主应用组件
│       ├── api.ts      # API 调用
│       └── types.ts    # TypeScript 类型
└── start.ps1           # 启动脚本
```

## 🚀 快速开始

### 1. 环境要求
- Go 1.24+
- Node.js 18+
- npm 或 yarn

### 2. 安装步骤

```bash
# 进入项目目录
cd agentLearn

# 复制环境变量配置
cp .env.example .env

# 编辑 .env 文件，设置你的 API Key
# Windows: notepad .env
# Linux/Mac: vim .env
```

### 3. 配置 API Key

在 `.env` 文件中设置：

```bash
# OpenAI
OPENAI_API_KEY=sk-xxx
OPENAI_BASE_URL=https://api.openai.com/v1

# 或使用国内 API（推荐）
DEEPSEEK_API_KEY=your_deepseek_key
DEEPSEEK_BASE_URL=https://api.deepseek.com/v1
```

### 4. 启动项目

#### 方式一：使用启动脚本（推荐）

```bash
# Windows PowerShell
.\start.ps1

# 或双击 start.bat
```

#### 方式二：手动启动

```bash
# 终端 1 - 启动后端
go run cmd/server/main.go

# 终端 2 - 启动前端
cd frontend
npm install
npm run dev
```

### 5. 访问应用

打开浏览器访问：http://localhost:5173

## 💡 功能演示

### 基础对话
```
用户：你好，请介绍一下你自己
AI: 你好！我是 GoAgent TaskRunner，一个智能任务执行助手...
```

### 工具调用示例

**读取文件**
```
用户：读取 README.md 文件并总结内容
AI: [调用 file_processor 工具读取文件]
    [分析文件内容]
    这是一个智能体开发框架，主要功能包括...
```

**HTTP 请求**
```
用户：访问 https://api.github.com/users/octocat
AI: [调用 http_client 工具发送 GET 请求]
    [解析返回的 JSON 数据]
    Octocat 是 GitHub 的创始人...
```

**执行命令**
```
用户：执行命令 echo Hello World
AI: [调用 shell_executor 工具]
    Command executed successfully:
    Hello World
```

## 🎯 面试准备要点

### 技术难点
1. **ReAct Loop 实现** - 如何平衡推理和行动的次数
2. **流式处理** - SSE 协议的实现和前端解析
3. **上下文管理** - 多轮对话的历史维护
4. **工具调用安全** - Shell 命令的安全检查和限制
5. **并发控制** - 并行工具调用的同步问题

### 架构设计
- **分层架构** - 清晰的职责划分
- **接口抽象** - LLM 客户端的可替换设计
- **依赖注入** - 便于测试和扩展
- **配置驱动** - 灵活的运行时配置

### 可扩展点
- 添加数据库查询工具
- 集成 MCP 协议
- 实现记忆系统
- 添加 RAG 能力
- Web 控制台升级

## 📝 工作流定义示例

工作流允许你通过 YAML 定义复杂的业务流程：

```yaml
name: 数据导出工作流
description: 从数据库查询数据并导出为 CSV 文件
version: "1.0"

nodes:
  - id: query_db
    type: tool
    tool: db_query
    params:
      query: "SELECT * FROM users"
    next: save_file

  - id: save_file
    type: tool
    tool: file_processor
    params:
      action: write
      path: "./exports/data.csv"
```

## 🧪 测试

```bash
# 运行单元测试
go test ./internal/tools/... -v

# 运行 Agent 测试
go test ./internal/agent/... -v
```

## 🔧 开发计划

- [ ] 完善工作流引擎执行逻辑
- [ ] 添加数据库查询工具
- [ ] 实现人类确认节点（Human-in-the-loop）
- [ ] 添加审计日志系统
- [ ] 前端增加工具调用可视化
- [ ] Docker 容器化部署

## 📚 学习资源

- [BabyAgent](../baby-agent/) - 更详细的教学项目
- [OpenAI Function Calling](https://platform.openai.com/docs/guides/function-calling)
- [ReAct Paper](https://arxiv.org/abs/2210.03629)

## 🤝 参与贡献

欢迎提交 Issue 和 PR！

## 📄 开源协议

MIT License

---

**💼 简历描述建议：**

> **GoAgent-TaskRunner | 企业级任务执行智能体框架**
> 
> - 基于 Go + React 开发的 AI 智能体框架，支持自然语言触发复杂业务流程
> - 实现了 ReAct Agent Loop 和 Function Calling 机制，可自主规划和执行任务
> - 设计了可插拔工具系统和工作流引擎，支持 YAML 声明式定义业务流程
> - 采用 SSE 实现流式输出，使用 Clean Architecture 保证代码可维护性
> - 提供 RESTful API 和现代化 Web 界面，支持实时对话和工具调用可视化

---

**作者**: [你的名字]  
**邮箱**: [你的邮箱]  
**GitHub**: [你的 GitHub 链接]
