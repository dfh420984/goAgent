# 🚀 GoAgent TaskRunner 快速启动指南

## ⚡ 5 分钟快速开始

### 第一步：配置 API Key

1. 打开 `.env` 文件
2. 设置你的 API Key（选择以下任一）：

```bash
# 方案 A: 使用 DeepSeek（推荐，国内访问快）
DEEPSEEK_API_KEY=sk_your_deepseek_key_here
DEEPSEEK_BASE_URL=https://api.deepseek.com/v1

# 方案 B: 使用 OpenAI
OPENAI_API_KEY=sk-your_openai_key_here
OPENAI_BASE_URL=https://api.openai.com/v1

# 方案 C: 使用其他兼容平台（如 Moonshot、Zhipu）
```

### 第二步：启动项目

#### Windows 用户（推荐）
直接双击运行 `start.bat`

或使用 PowerShell：
```powershell
.\start.ps1
```

#### 手动启动（分两个终端）

**终端 1 - 后端服务：**
```bash
cd agentLearn
go run cmd/server/main.go
```

**终端 2 - 前端开发服务器：**
```bash
cd agentLearn/frontend
npm install
npm run dev
```

### 第三步：访问应用

浏览器打开：**http://localhost:5173**

---

## 💡 测试示例

### 1. 基础对话
```
你好，请介绍一下你自己
```

### 2. 文件操作
```
请在 ./test 目录下创建一个 hello.txt 文件，内容为 "Hello, Agent!"
```

```
读取刚刚创建的 hello.txt 文件内容
```

### 3. HTTP 请求
```
访问 https://api.github.com/users/octocat 并告诉我这个用户的信息
```

### 4. Shell 命令
```
执行命令：echo Hello World
```

```
查看当前目录下的文件列表
```

---

## 🔧 常见问题

### Q1: 提示 "API key is required"
**解决**: 确保 `.env` 文件中设置了正确的 API Key，并且没有注释掉

### Q2: 前端无法连接后端
**解决**: 
1. 检查后端是否在 8080 端口启动
2. 查看前端代理配置（`frontend/vite.config.ts`）
3. 确保 CORS 配置正确

### Q3: npm install 失败
**解决**:
```bash
# 使用淘宝镜像
npm config set registry https://registry.npmmirror.com
npm install
```

### Q4: go mod tidy 下载依赖慢
**解决**:
```bash
# 使用 GOPROXY
export GOPROXY=https://goproxy.cn,direct
go mod tidy
```

---

## 📊 项目架构预览

```
┌─────────────┐         ┌──────────────┐
│   前端      │  HTTP   │   后端 API   │
│ React + TS  │ ◄────► │  Gin Server  │
└─────────────┘         └──────┬───────┘
                               │
                        ┌──────▼───────┐
                        │  Agent 引擎  │
                        │  ReAct Loop  │
                        └──────┬───────┘
                               │
                    ┌──────────┼──────────┐
                    │          │          │
             ┌──────▼──┐ ┌────▼────┐ ┌──▼──────┐
             │ 文件工具│ │HTTP 工具 │ │Shell 工具│
             └─────────┘ └─────────┘ └─────────┘
```

---

## 🎯 下一步学习建议

1. **阅读核心代码**
   - `internal/agent/engine.go` - Agent 引擎实现
   - `cmd/server/main.go` - API 服务入口
   - `frontend/src/App.tsx` - 前端组件

2. **理解关键概念**
   - Function Calling 原理
   - ReAct Loop 执行流程
   - SSE 流式传输机制

3. **扩展功能**
   - 添加新的工具（如数据库查询）
   - 自定义工作流定义
   - 实现记忆系统

4. **准备面试**
   - 阅读 README.md 中的面试要点
   - 整理技术难点和解决方案
   - 准备项目演示

---

## 📚 相关资源

- [完整文档](README.md)
- [BabyAgent 教程](../baby-agent/README.md)
- [OpenAI Function Calling 文档](https://platform.openai.com/docs/guides/function-calling)

---

**祝你学习愉快！🎉**
