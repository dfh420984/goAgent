package mcp

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"goagent/internal/tools"
	"goagent/pkg/mcp"
)

// Loader MCP 工具加载器
type Loader struct {
	configFile string
}

// NewLoader 创建 MCP 加载器
func NewLoader(configFile string) *Loader {
	return &Loader{
		configFile: configFile,
	}
}

// LoadAndRegister 加载 MCP 工具并注册到工具注册表
func (l *Loader) LoadAndRegister(registry *tools.ToolRegistry) ([]*mcp.Client, error) {
	ctx := context.Background()

	// 读取 MCP 配置文件
	configContent, err := os.ReadFile(l.configFile)
	if err != nil {
		log.Printf("⚠️  读取 MCP 配置失败：%v", err)
		return nil, err
	}

	var serverConfigs map[string]mcp.ClientConfig
	if err := json.Unmarshal(configContent, &serverConfigs); err != nil {
		log.Printf("⚠️  解析 MCP 配置失败：%v", err)
		return nil, err
	}

	// 创建并初始化 MCP 客户端
	mcpClients := make([]*mcp.Client, 0)
	for name, config := range serverConfigs {
		log.Printf("   连接 MCP 服务器：%s", name)
		client := mcp.NewClient(&config)

		// 刷新工具列表
		if err := client.RefreshTools(ctx); err != nil {
			log.Printf("   ⚠️  %s: 连接失败 - %v", name, err)
			continue
		}

		// 注册所有 MCP 工具
		for _, mcpTool := range client.GetTools() {
			wrapper := NewToolWrapper(mcpTool)
			registry.Register(wrapper)
			log.Printf("   ✅ 注册 MCP 工具：%s", wrapper.Name())
		}

		mcpClients = append(mcpClients, client)
	}

	return mcpClients, nil
}
