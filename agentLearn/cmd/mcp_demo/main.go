package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"goagent/pkg/mcp"
)

func main() {
	fmt.Println("🚀 MCP 集成演示")
	fmt.Println("================\n")

	ctx := context.Background()

	// 1. 加载 MCP 服务器配置
	fmt.Println("📋 步骤 1: 加载 MCP 服务器配置")
	configContent, err := os.ReadFile("mcp-servers.json")
	if err != nil {
		log.Fatalf("读取配置文件失败：%v", err)
	}

	var serverConfigs map[string]mcp.ClientConfig
	if err := json.Unmarshal(configContent, &serverConfigs); err != nil {
		log.Fatalf("解析配置文件失败：%v", err)
	}
	fmt.Printf("✅ 找到 %d 个 MCP 服务器配置\n\n", len(serverConfigs))

	// 2. 创建并初始化 MCP 客户端
	fmt.Println("🔧 步骤 2: 创建 MCP 客户端")
	mcpClients := make([]*mcp.Client, 0)

	for name, config := range serverConfigs {
		fmt.Printf("   创建客户端：%s\n", name)
		client := mcp.NewClient(&config)

		// 刷新工具列表
		if err := client.RefreshTools(ctx); err != nil {
			fmt.Printf("   ⚠️  警告：无法连接 MCP 服务器 %s: %v\n", name, err)
			fmt.Println("   提示：确保已安装 Node.js 和 npx，并且网络正常")
			continue
		}

		mcpClients = append(mcpClients, client)
		fmt.Printf("   ✅ %s: 加载了 %d 个工具\n", name, len(client.GetTools()))
	}
	fmt.Println()

	// 3. 列出所有可用工具
	fmt.Println("🛠️  步骤 3: 显示所有可用工具")
	allTools := make([]*mcp.Tool, 0)
	for _, client := range mcpClients {
		tools := client.GetTools()
		allTools = append(allTools, tools...)

		for _, tool := range tools {
			fmt.Printf("   • %s\n", tool.Name())
			fmt.Printf("     描述：%s\n", tool.Description())

			// 显示参数信息
			schema := tool.InputSchema()
			if props, ok := schema["properties"].(map[string]interface{}); ok {
				if len(props) > 0 {
					fmt.Printf("     参数：%v\n", props)
				}
			}
			fmt.Println()
		}
	}

	if len(allTools) == 0 {
		fmt.Println("⚠️  没有找到可用的工具")
		fmt.Println("\n💡 提示:")
		fmt.Println("   1. 确保已安装 Node.js (https://nodejs.org/)")
		fmt.Println("   2. 运行以下命令安装 MCP 服务器:")
		fmt.Println("      npm install -g @modelcontextprotocol/server-filesystem")
		fmt.Println("      npm install -g @modelcontextprotocol/server-memory")
		return
	}

	// 4. 演示工具调用
	fmt.Println("🎯 步骤 4: 演示工具调用")
	fmt.Println("-------------------")

	// 查找文件系统工具
	var listFilesTool *mcp.Tool
	for _, tool := range allTools {
		if tool.Name() == "mcp_filesystem__list_directory" ||
			tool.Name() == "mcp_filesystem__read_file" {
			listFilesTool = tool
			break
		}
	}

	if listFilesTool != nil {
		fmt.Printf("\n📁 演示调用：%s\n", listFilesTool.Name())

		// 准备参数
		args := map[string]interface{}{
			"path": ".",
		}
		fmt.Printf("   参数：%+v\n", args)

		// 执行工具
		result, err := listFilesTool.Execute(ctx, args)
		if err != nil {
			fmt.Printf("   ❌ 调用失败：%v\n", err)
		} else {
			fmt.Printf("   ✅ 调用成功!\n")
			fmt.Printf("   结果:\n%s\n", result)
		}
	} else {
		fmt.Println("   ⚠️  未找到合适的演示工具")
	}

	// 5. 总结
	fmt.Println("\n📊 总结")
	fmt.Println("------")
	fmt.Printf("✅ 成功连接 %d 个 MCP 服务器\n", len(mcpClients))
	fmt.Printf("✅ 总共发现 %d 个工具\n", len(allTools))
	fmt.Println("\n💡 下一步:")
	fmt.Println("   - 将这些工具集成到您的 Agent 中")
	fmt.Println("   - 让 LLM 自动决定调用哪个工具")
	fmt.Println("   - 参考 baby-agent/ch04 查看完整的 Agent 集成示例")
}
