package tools

import (
	"context"
	"encoding/json"

	"goagent/pkg/llm"
)

// ToolRegistry 工具注册表
type ToolRegistry struct {
	tools map[string]Tool
}

// Tool 工具接口
type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]interface{}
	Execute(ctx context.Context, args map[string]interface{}) (string, error)
}

// NewToolRegistry 创建工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
func (r *ToolRegistry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get 获取工具
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// List 列出所有工具
func (r *ToolRegistry) List() []Tool {
	result := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		result = append(result, tool)
	}
	return result
}

// ToLLMTools 转换为 LLM 工具定义格式
func (r *ToolRegistry) ToLLMTools() []llm.ToolDefinition {
	result := make([]llm.ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		result = append(result, llm.ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  tool.Parameters(),
		})
	}
	return result
}

// ExecuteTool 执行工具
func (r *ToolRegistry) ExecuteTool(ctx context.Context, name string, argsJSON string) (string, error) {
	tool, ok := r.Get(name)
	if !ok {
		return "", nil
	}

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}

	return tool.Execute(ctx, args)
}
