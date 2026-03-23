package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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

// StreamToolCall 流式工具调用片段
type StreamToolCall struct {
	Index    int    `json:"index"`
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
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

	// 设置超时
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	fmt.Printf("[LLM] Calling API: %s/model/%s\n", c.config.BaseURL, c.config.Model)

	resp, err := c.client.R().
		SetContext(ctxWithTimeout).
		SetHeader("Authorization", "Bearer "+c.config.APIKey).
		SetBody(req).
		SetDoNotParseResponse(true).
		Post("/chat/completions")

	fmt.Printf("[LLM DEBUG] Request - Model: %s, Messages: %+v\n", c.config.Model, messages) // 调试日志
	if err != nil {
		fmt.Printf("[LLM DEBUG] HTTP Error: %v\n", err) // 调试日志
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}
	defer resp.RawBody().Close()

	fmt.Printf("[LLM DEBUG] Response Status: %d\n", resp.StatusCode()) // 调试日志

	// 使用 bufio.Scanner 逐行读取流式响应
	scanner := bufio.NewScanner(resp.RawBody())

	// 用于合并工具调用的片段
	type ToolCallBuilder struct {
		ID        string
		Name      string
		Type      string
		Arguments strings.Builder
	}
	toolCallBuilders := make(map[int]*ToolCallBuilder)

	for scanner.Scan() {
		line := scanner.Text()
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
					Content   string           `json:"content"`
					ToolCalls []StreamToolCall `json:"tool_calls"`
					Role      string           `json:"role"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
			fmt.Printf("[LLM DEBUG] Unmarshal error: %v\n", err) // 调试日志
			continue
		}

		fmt.Printf("[LLM DEBUG] Decoded data: %+v\n", data) // 调试日志

		if len(data.Choices) > 0 {
			delta := data.Choices[0].Delta
			if delta.Content != "" {
				fullContent.WriteString(delta.Content)
				callback(delta.Content)
			}

			// 处理工具调用（需要合并片段）
			if len(delta.ToolCalls) > 0 {
				for _, tc := range delta.ToolCalls {
					// 获取或创建 builder
					builder, exists := toolCallBuilders[tc.Index]
					if !exists {
						builder = &ToolCallBuilder{}
						toolCallBuilders[tc.Index] = builder
					}

					// 累积信息
					if tc.ID != "" {
						builder.ID = tc.ID
					}
					if tc.Type != "" {
						builder.Type = tc.Type
					}
					if tc.Function.Name != "" {
						builder.Name = tc.Function.Name
					}
					// 累积参数字符串
					if tc.Function.Arguments != "" {
						builder.Arguments.WriteString(tc.Function.Arguments)
					}

					fmt.Printf("[TOOL CHUNK] Index: %d, Name: %s, Args: %s\n",
						tc.Index, tc.Function.Name, tc.Function.Arguments)
				}
			}
		}
	}

	// 将合并后的工具调用转换为最终格式
	for _, builder := range toolCallBuilders {
		toolCalls = append(toolCalls, ToolCall{
			ID:        builder.ID,
			Name:      builder.Name,
			Arguments: builder.Arguments.String(),
		})
		fmt.Printf("[TOOL MERGED] Final tool call: Name=%s, Args=%s\n",
			builder.Name, builder.Arguments.String())
	}

	result := &Response{
		Content:   fullContent.String(),
		ToolCalls: toolCalls,
	}

	return result, nil
}
