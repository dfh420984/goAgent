# MCP 集成快速指南

## 📚 什么是 MCP？

**MCP (Model Context Protocol)** 是一个开放协议，用于让 AI Agent 以统一方式发现和调用外部工具。

### 核心概念

- **Client**: 嵌入在 Agent 中的客户端（我们的 `pkg/mcp/client.go`）
- **Server**: 提供工具的服务端（如文件系统服务器、内存服务器）
- **Tool**: 可被调用的功能接口

## 🎯 本 Demo 包含什么

```
agentLearn/
├── pkg/mcp/
│   └── client.go          # MCP 客户端封装
├── cmd/mcp_demo/
│   └── main.go            # 演示程序
├── mcp-servers.json       # MCP 服务器配置
└── run_mcp_demo.bat       # Windows 启动脚本
```

## 🚀 快速开始

### 方法 1: 使用批处理（推荐）

```bash
# 双击运行或在命令行执行
run_mcp_demo.bat
```

### 方法 2: 手动运行

```bash
# 1. 确保已安装 Node.js
node --version

# 2. 安装 MCP 服务器
npm install -g @modelcontextprotocol/server-filesystem

# 3. 运行演示
go run cmd/mcp_demo/main.go
```

## 📋 配置文件说明

`mcp-servers.json` 支持两种模式：

### stdio 模式（本地进程）

```json
{
  "filesystem": {
    "name": "filesystem",
    "command": "npx",
    "args": [
      "-y",
      "@modelcontextprotocol/server-filesystem",
      "."
    ]
  }
}
```

### HTTP/SSE 模式（远程服务）

```json
{
  "remote_server": {
    "name": "remote",
    "type": "http",
    "url": "http://localhost:8080/mcp"
  }
}
```

## 🛠️ 常用 MCP 服务器

### 官方服务器

1. **文件系统服务器**
   ```bash
   npm install -g @modelcontextprotocol/server-filesystem
   ```
   - `list_directory`: 列出目录内容
   - `read_file`: 读取文件
   - `write_file`: 写入文件

2. **内存服务器**
   ```bash
   npm install -g @modelcontextprotocol/server-memory
   ```
   - `create_memory`: 创建记忆
   - `get_memories`: 获取记忆
   - `delete_memory`: 删除记忆

3. **数据库服务器**
   ```bash
   npm install -g @modelcontextprotocol/server-sqlite
   ```
   - `query`: 执行 SQL 查询

## 💡 代码示例

### 在 Agent 中集成 MCP

```go
import (
    "agentLearn/pkg/mcp"
)

// 1. 加载配置
config := &mcp.ClientConfig{
    Name:    "filesystem",
    Command: "npx",
    Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "."},
}

// 2. 创建客户端
client := mcp.NewClient(config)

// 3. 刷新工具列表
ctx := context.Background()
if err := client.RefreshTools(ctx); err != nil {
    log.Fatal(err)
}

// 4. 获取工具
tools := client.GetTools()
for _, tool := range tools {
    fmt.Printf("工具：%s\n", tool.Name())
    fmt.Printf("描述：%s\n", tool.Description())
}

// 5. 调用工具
result, err := tools[0].Execute(ctx, map[string]interface{}{
    "path": "./data.txt",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println("结果:", result)
```

### 与 LLM 配合使用

```go
// 将 MCP 工具注册到 OpenAI API
func buildTools(mcpClients []*mcp.Client) []map[string]interface{} {
    tools := make([]map[string]interface{}, 0)
    
    for _, client := range mcpClients {
        for _, tool := range client.GetTools() {
            tools = append(tools, tool.Info())
        }
    }
    
    return tools
}

// 在 LLM 请求中使用
messages := []openai.ChatCompletionMessageParamUnion{
    openai.UserMessage("请列出当前目录下的所有文件"),
}

resp, err := client.ChatCompletion(ctx, openai.ChatCompletionNewParams{
    Messages: messages,
    Tools:    buildTools(mcpClients),
})

// 如果 LLM 决定调用工具
if len(resp.Choices) > 0 && resp.Choices[0].Message.ToolCalls != nil {
    toolCall := resp.Choices[0].Message.ToolCalls[0]
    
    // 查找对应的 MCP 工具
    var targetTool *mcp.Tool
    for _, client := range mcpClients {
        for _, tool := range client.GetTools() {
            if tool.Name() == toolCall.Function.Name {
                targetTool = tool
                break
            }
        }
    }
    
    // 执行工具
    var args map[string]interface{}
    json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
    result, _ := targetTool.Execute(ctx, args)
}
```

## 🔍 调试技巧

### 启用详细日志

在 `pkg/mcp/client.go` 中添加：

```go
log.SetLevel(log.DebugLevel)
```

### 查看原始通信

修改 `callTool` 方法：

```go
log.Printf("[MCP Request] → %s %+v\n", toolName, arguments)
result, err := c.session.CallTool(...)
log.Printf("[MCP Response] ← %s\n", result)
```

## ❓ 常见问题

### Q: 提示 "npx: command not found"
A: 需要先安装 Node.js: https://nodejs.org/

### Q: 工具列表为空
A: 检查：
1. MCP 服务器是否正确安装
2. 网络连接是否正常
3. 查看错误日志

### Q: 调用超时
A: 可能是服务器执行时间过长，可以：
1. 增加 timeout 配置
2. 优化服务器性能
3. 检查服务器是否卡住

## 📖 深入学习

- **baby-agent/ch04**: 完整的 Agent 集成示例
- **baby-agent/ch04/mcp.go**: 更详细的 MCP 实现
- **MCP 官方文档**: https://modelcontextprotocol.io/

## 🎓 下一步

1. ✅ 运行 demo，了解 MCP 基本用法
2. 📝 尝试添加新的 MCP 服务器配置
3. 🔧 将 MCP 工具集成到您的工作流中
4. 🤖 让 LLM 自动调用 MCP 工具

祝您使用愉快！🎉
