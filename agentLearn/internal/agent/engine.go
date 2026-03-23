package agent

import (
	"context"
	"fmt"
	"time"

	"goagent/internal/tools"
	"goagent/pkg/llm"
)

// Engine Agent 引擎
type Engine struct {
	llmClient    llm.Client
	toolRegistry *tools.ToolRegistry
	messages     []llm.Message
	systemPrompt string
}

// NewEngine 创建 Agent 引擎
func NewEngine(llmClient llm.Client, toolRegistry *tools.ToolRegistry, systemPrompt string) *Engine {
	return &Engine{
		llmClient:    llmClient,
		toolRegistry: toolRegistry,
		messages:     make([]llm.Message, 0),
		systemPrompt: systemPrompt,
	}
}

// Chat 对话方法
func (e *Engine) Chat(ctx context.Context, userMessage string, streamCallback func(string)) (*ChatResult, error) {
	// 添加用户消息
	e.messages = append(e.messages, llm.Message{
		Role:    llm.RoleUser,
		Content: userMessage,
	})

	result := &ChatResult{
		StartTime: time.Now(),
	}

	// 执行 ReAct Loop
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// 调用 LLM
		resp, err := e.llmClient.Chat(ctx, e.messages, e.toolRegistry.ToLLMTools())
		if err != nil {
			return nil, err
		}

		result.TotalUsage.PromptTokens += resp.Usage.PromptTokens
		result.TotalUsage.CompletionTokens += resp.Usage.CompletionTokens
		result.TotalUsage.TotalTokens += resp.Usage.TotalTokens

		// 如果有工具调用
		if len(resp.ToolCalls) > 0 {
			result.ToolCallCount += len(resp.ToolCalls)

			// 执行每个工具调用
			for _, tc := range resp.ToolCalls {
				toolResult, err := e.toolRegistry.ExecuteTool(ctx, tc.Name, tc.Arguments)
				if err != nil {
					return nil, fmt.Errorf("tool execution failed: %w", err)
				}

				// 添加工具调用和结果到消息历史
				e.messages = append(e.messages, llm.Message{
					Role:    llm.RoleAssistant,
					Content: fmt.Sprintf("Calling tool: %s with args: %s", tc.Name, tc.Arguments),
				})
				e.messages = append(e.messages, llm.Message{
					Role:    llm.RoleUser,
					Content: fmt.Sprintf("Tool result: %s", toolResult),
				})
			}
			continue
		}

		// 没有工具调用，返回最终响应
		e.messages = append(e.messages, llm.Message{
			Role:    llm.RoleAssistant,
			Content: resp.Content,
		})

		result.Response = resp.Content
		result.Iterations = i + 1
		result.Duration = time.Since(result.StartTime)

		return result, nil
	}

	return nil, fmt.Errorf("max iterations reached")
}

// ChatStream 流式对话
func (e *Engine) ChatStream(ctx context.Context, userMessage string, streamCallback func(string)) (*ChatResult, error) {
	// 添加用户消息
	e.messages = append(e.messages, llm.Message{
		Role:    llm.RoleUser,
		Content: userMessage,
	})

	result := &ChatResult{
		StartTime: time.Now(),
	}

	// 暂时简化处理，后续可以支持流式工具调用
	resp, err := e.llmClient.ChatStream(ctx, e.messages, e.toolRegistry.ToLLMTools(), streamCallback)
	if err != nil {
		return nil, err
	}

	e.messages = append(e.messages, llm.Message{
		Role:    llm.RoleAssistant,
		Content: resp.Content,
	})

	result.Response = resp.Content
	result.TotalUsage = resp.Usage
	result.Duration = time.Since(result.StartTime)

	return result, nil
}

// ClearHistory 清空历史消息
func (e *Engine) ClearHistory() {
	e.messages = make([]llm.Message, 0)
}

// GetHistory 获取历史消息
func (e *Engine) GetHistory() []llm.Message {
	return e.messages
}

// ChatResult 对话结果
type ChatResult struct {
	Response      string
	ToolCallCount int
	Iterations    int
	TotalUsage    llm.Usage
	Duration      time.Duration
	StartTime     time.Time
}
