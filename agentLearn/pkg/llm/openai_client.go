package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/go-resty/resty/v2"
)

// OpenAICompatibleClient OpenAI 兼容的客户端
type OpenAICompatibleClient struct {
	client *resty.Client
	config Config
}

// ChatRequest 聊天请求体
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// Tool 工具定义（OpenAI 格式）
type Tool struct {
	Type     string         `json:"type"`
	Function ToolDefinition `json:"function"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
		Delta        struct {
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls"`
		} `json:"delta"`
	} `json:"choices"`
	Usage Usage `json:"usage"`
}

func NewOpenAICompatibleClient(cfg Config) *OpenAICompatibleClient {
	return &OpenAICompatibleClient{
		client: resty.New().SetBaseURL(cfg.BaseURL),
		config: cfg,
	}
}

func (c *OpenAICompatibleClient) Chat(ctx context.Context, messages []Message, tools []ToolDefinition) (*Response, error) {
	req := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
	}

	if len(tools) > 0 {
		for _, tool := range tools {
			req.Tools = append(req.Tools, Tool{
				Type:     "function",
				Function: tool,
			})
		}
	}

	var resp ChatResponse
	_, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+c.config.APIKey).
		SetBody(req).
		SetResult(&resp).
		Post("/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := resp.Choices[0]
	result := &Response{
		Content:      choice.Message.Content,
		Usage:        resp.Usage,
		FinishReason: choice.FinishReason,
	}

	return result, nil
}

func (c *OpenAICompatibleClient) ChatStream(ctx context.Context, messages []Message, tools []ToolDefinition, callback func(string)) (*Response, error) {
	req := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		Stream:      true,
	}

	if len(tools) > 0 {
		for _, tool := range tools {
			req.Tools = append(req.Tools, Tool{
				Type:     "function",
				Function: tool,
			})
		}
	}

	var fullContent strings.Builder
	var toolCalls []ToolCall

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+c.config.APIKey).
		SetBody(req).
		SetDoNotParseResponse(true).
		Post("/chat/completions")

	if err != nil {
		fmt.Printf("[LLM DEBUG] HTTP Error: %v\n", err) // 调试日志
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}
	defer resp.RawBody().Close()

	// 读取原始响应内容用于调试
	rawBody, err := io.ReadAll(resp.RawBody())
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 按行解析 SSE 格式
	lines := strings.Split(string(rawBody), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}

		// 去掉 "data: " 前缀
		jsonData := strings.TrimPrefix(line, "data:")
		jsonData = strings.TrimSpace(jsonData)

		if jsonData == "[DONE]" {
			break
		}

		var data struct {
			Choices []struct {
				Delta struct {
					Content   string     `json:"content"`
					ToolCalls []ToolCall `json:"tool_calls"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
			fmt.Printf("[LLM DEBUG] Unmarshal error: %v\n", err) // 调试日志
			continue
		}

		if len(data.Choices) > 0 {
			delta := data.Choices[0].Delta
			if delta.Content != "" {
				fullContent.WriteString(delta.Content)
				callback(delta.Content)
			}
			if len(delta.ToolCalls) > 0 {
				toolCalls = append(toolCalls, delta.ToolCalls...)
			}
		}
	}

	result := &Response{
		Content:   fullContent.String(),
		ToolCalls: toolCalls,
	}

	return result, nil
}
