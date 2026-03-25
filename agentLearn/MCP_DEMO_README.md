# MCP 集成 Demo 说明

## 📦 项目结构

```
agentLearn/
├── pkg/mcp/
│   └── client.go              # MCP 客户端核心封装
├── cmd/
│   ├── mcp_demo/              # 基础演示程序
│   │   └── main.go
│   └── workflow_mcp_demo/     # 工作流集成演示
│       └── main.go
├── mcp-servers.json           # MCP 服务器配置
├── run_mcp_demo.bat           # 快速启动脚本
├── run_workflow_mcp_demo.bat  # 工作流集成启动脚本
└── MCP_QUICKSTART.md          # 快速指南
```

## 🎯 实现的功能

### 1. MCP 客户端封装 (`pkg/mcp/client.go`)

**核心组件：**

- `ClientConfig`: MCP 服务器配置
- `Client`: MCP 客户端，管理连接和工具
- `Tool`: MCP 工具封装，提供统一接口

**主要方法：**

```go
// 创建客户端
client := mcp.NewClient(&mcp.ClientConfig{
    Name:    "filesystem",
    Command: "npx",
    Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "."},
})

// 刷新工具列表
err := client.RefreshTools(ctx)

// 获取所有工具
tools := client.GetTools()

// 调用工具
result, err := tool.Execute(ctx, args)
```

### 2. 支持的传输模式

#### stdio 模式（本地进程间通信）

```json
{
  "filesystem": {
    "name": "filesystem",
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
  }
}
```

**特点：**
- ✅ 低延迟，适合本地工具
- ✅ 简单可靠，无需网络
- ❌ 只能本地使用

#### HTTP/SSE 模式（远程服务）

```json
{
  "remote": {
    "name": "remote",
    "type": "http",
    "url": "http://localhost:8080/mcp"
  }
}
```

**特点：**
- ✅ 支持远程调用
- ✅ 可跨机器部署
- ❌ 需要网络连接

### 3. 工具包装器 (`workflow_mcp_demo/main.go`)

将 MCP 工具包装为工作流工具接口：

```go
type MCPToolWrapper struct {
    mcpTool *mcp.Tool
}

func (w *MCPToolWrapper) Name() string {
    return w.mcpTool.Name()
}

func (w *MCPToolWrapper) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    return w.mcpTool.Execute(ctx, args)
}
```

## 🚀 运行演示

### Demo 1: 基础 MCP 演示

```bash
# Windows
run_mcp_demo.bat

# 或手动运行
go run cmd/mcp_demo/main.go
```

**演示内容：**
1. 加载 MCP 服务器配置
2. 创建并初始化 MCP 客户端
3. 显示所有可用工具
4. 实际调用一个 MCP 工具

### Demo 2: 工作流集成演示

```bash
# Windows
run_workflow_mcp_demo.bat

# 或手动运行
go run cmd/workflow_mcp_demo/main.go
```

**演示内容：**
1. 加载 MCP 配置
2. 初始化 MCP 客户端
3. 将 MCP 工具注册到工作流 registry
4. 演示在工作流中调用 MCP 工具

## 💡 代码示例详解

### 示例 1: 简单的文件读取

```go
package main

import (
    "context"
    "agentLearn/pkg/mcp"
)

func main() {
    ctx := context.Background()
    
    // 1. 创建配置
    config := &mcp.ClientConfig{
        Name:    "filesystem",
        Command: "npx",
        Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "."},
    }
    
    // 2. 创建客户端
    client := mcp.NewClient(config)
    
    // 3. 刷新工具
    if err := client.RefreshTools(ctx); err != nil {
        panic(err)
    }
    
    // 4. 查找 read_file 工具
    var readTool *mcp.Tool
    for _, tool := range client.GetTools() {
        if tool.Name() == "mcp_filesystem__read_file" {
            readTool = tool
            break
        }
    }
    
    // 5. 调用工具
    result, err := readTool.Execute(ctx, map[string]interface{}{
        "path": "./README.md",
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Println("文件内容:", result)
}
```

### 示例 2: 在工作流 YAML 中使用

```yaml
name: 文件处理工作流
description: 使用 MCP 工具处理文件
version: "1.0"

nodes:
  - id: start
    type: trigger
    next: list_files

  - id: list_files
    type: tool
    tool: mcp_filesystem__list_directory
    params:
      path: "./data"
    next: read_first_file

  - id: read_first_file
    type: tool
    tool: mcp_filesystem__read_file
    params:
      path: "{{list_files}}/file1.txt"
    next: process_data

  - id: process_data
    type: llm
    prompt: "请处理以下文件内容：{{read_first_file}}"
    next: end

  - id: end
    type: terminator
```

### 示例 3: 与 LLM 配合使用（Agent 模式）

```go
// 构建工具列表给 LLM
func buildToolsForLLM(mcpClients []*mcp.Client) []map[string]interface{} {
    tools := make([]map[string]interface{}, 0)
    
    for _, client := range mcpClients {
        for _, tool := range client.GetTools() {
            tools = append(tools, tool.Info())
        }
    }
    
    return tools
}

// 处理 LLM 的工具调用
func handleLLMToolCall(ctx context.Context, 
                      mcpClients []*mcp.Client, 
                      toolName string, 
                      args string) (string, error) {
    
    // 查找对应的 MCP 工具
    var targetTool *mcp.Tool
    for _, client := range mcpClients {
        for _, tool := range client.GetTools() {
            if tool.Name() == toolName {
                targetTool = tool
                break
            }
        }
    }
    
    if targetTool == nil {
        return "", fmt.Errorf("未找到工具：%s", toolName)
    }
    
    // 解析参数并执行
    var argsMap map[string]interface{}
    json.Unmarshal([]byte(args), &argsMap)
    
    return targetTool.Execute(ctx, argsMap)
}
```

## 🔍 调试技巧

### 1. 启用详细日志

在 `pkg/mcp/client.go` 的 `callTool` 方法中添加：

```go
log.Printf("[MCP] 调用工具：%s, 参数：%+v\n", toolName, arguments)
log.Printf("[MCP] 请求 JSON: %s\n", argsJSON)
```

### 2. 查看工具列表

```go
for _, tool := range client.GetTools() {
    fmt.Printf("工具：%s\n", tool.Name())
    fmt.Printf("  描述：%s\n", tool.Description())
    fmt.Printf("  参数：%+v\n", tool.InputSchema())
}
```

### 3. 测试连接

```go
if err := client.RefreshTools(ctx); err != nil {
    log.Printf("连接失败：%v", err)
    log.Printf("请检查:")
    log.Printf("  1. Node.js 是否安装")
    log.Printf("  2. npx 是否可用")
    log.Printf("  3. MCP 服务器是否正确安装")
}
```

## 📊 架构图

```
┌─────────────────┐
│   Agent/LLM     │
│                 │
│  决定调用工具    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Tool Registry   │
│                 │
│  管理所有工具    │
└────────┬────────┘
         │
         ├──────────────┬──────────────┐
         │              │              │
         ▼              ▼              ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│ WeatherTool │ │ DBQueryTool │ │ MCP Wrapper │
│ (原生工具)   │ │ (原生工具)   │ │ (适配器)    │
└─────────────┘ └─────────────┘ └──────┬──────┘
                                       │
                                       ▼
                                ┌─────────────┐
                                │ MCP Client  │
                                │             │
                                │ 协议转换层   │
                                └──────┬──────┘
                                       │
                         ┌─────────────┴─────────────┐
                         │                           │
                         ▼                           ▼
                  ┌─────────────┐            ┌─────────────┐
                  │ stdio 传输   │            │ HTTP 传输    │
                  │ (本地进程)   │            │ (远程服务)   │
                  └──────┬──────┘            └──────┬──────┘
                         │                           │
                         ▼                           ▼
                  ┌─────────────┐            ┌─────────────┐
                  │ MCP Server  │            │ MCP Server  │
                  │ filesystem  │            │ memory      │
                  └─────────────┘            └─────────────┘
```

## 🎓 深入学习

### 参考资源

1. **baby-agent/ch04**: 完整的 Agent + MCP 集成示例
2. **baby-agent/ch04/mcp.go**: 更详细的 MCP 实现
3. **MCP 官方文档**: https://modelcontextprotocol.io/
4. **MCP SDK**: https://github.com/modelcontextprotocol/go-sdk

### 推荐的 MCP 服务器

```bash
# 文件系统
npm install -g @modelcontextprotocol/server-filesystem

# 内存/知识库
npm install -g @modelcontextprotocol/server-memory

# SQLite 数据库
npm install -g @modelcontextprotocol/server-sqlite

# Git 操作
npm install -g @modelcontextprotocol/server-git

# Puppeteer (浏览器自动化)
npm install -g @modelcontextprotocol/server-puppeteer
```

## 🎉 总结

通过这个 Demo，您可以：

✅ 理解 MCP 的基本原理和架构
✅ 学会配置和连接 MCP 服务器
✅ 将 MCP 工具集成到工作流系统
✅ 在实际项目中应用 MCP 协议

祝您使用愉快！如有问题，请参考 `MCP_QUICKSTART.md` 或查看 baby-agent 的完整实现。
