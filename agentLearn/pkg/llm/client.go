package llm

import (
	"context"
	"errors"
)

// Message 表示对话消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// ToolDefinition 工具定义
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
	Index     int    `json:"index"` // 用于流式合并
}

// Response LLM 响应
type Response struct {
	Content      string
	ToolCalls    []ToolCall
	Usage        Usage
	FinishReason string
}

// Usage Token 使用情况
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Client LLM 客户端接口
type Client interface {
	Chat(ctx context.Context, messages []Message, tools []ToolDefinition) (*Response, error)
	ChatStream(ctx context.Context, messages []Message, tools []ToolDefinition, callback func(string)) (*Response, error)
}

// Config 客户端配置
type Config struct {
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float32
}

// NewClient 根据配置创建客户端
func NewClient(cfg Config) (Client, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("API key is required")
	}

	// 默认使用 OpenAI 兼容客户端
	return NewOpenAICompatibleClient(cfg), nil
}
