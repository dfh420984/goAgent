package mcp

import (
	"context"

	"goagent/pkg/mcp"
)

// ToolWrapper MCP 工具包装器
type ToolWrapper struct {
	mcpTool *mcp.Tool
}

// NewToolWrapper 创建 MCP 工具包装器
func NewToolWrapper(mcpTool *mcp.Tool) *ToolWrapper {
	return &ToolWrapper{
		mcpTool: mcpTool,
	}
}

// Name 返回工具名称
func (w *ToolWrapper) Name() string {
	return w.mcpTool.Name()
}

// Description 返回工具描述
func (w *ToolWrapper) Description() string {
	return w.mcpTool.Description()
}

// Parameters 返回参数定义
func (w *ToolWrapper) Parameters() map[string]interface{} {
	return w.mcpTool.InputSchema()
}

// Execute 执行工具
func (w *ToolWrapper) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	return w.mcpTool.Execute(ctx, args)
}
