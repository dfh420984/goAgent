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
				// 打印工具调用信息
				fmt.Printf("\n[TOOL CALL] 准备调用工具：%s\n", tc.Name)
				fmt.Printf("[TOOL CALL] 参数：%s\n", tc.Arguments)

				// 执行工具
				toolResult, err := e.toolRegistry.ExecuteTool(ctx, tc.Name, tc.Arguments)
				if err != nil {
					fmt.Printf("[TOOL ERROR] 工具执行失败：%v\n", err)
					return nil, fmt.Errorf("tool execution failed: %w", err)
				}

				// 打印工具执行结果
				fmt.Printf("[TOOL RESULT] 工具 '%s' 执行成功\n", tc.Name)
				if len(toolResult) > 200 {
					// 如果结果太长，只显示前 200 个字符
					fmt.Printf("[TOOL RESULT] 结果预览：%s...\n", toolResult[:200])
				} else {
					fmt.Printf("[TOOL RESULT] 结果：%s\n", toolResult)
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
			fmt.Println("[INFO] 继续下一轮 LLM 调用...")
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

	// 第一次流式调用
	resp, err := e.llmClient.ChatStream(ctx, e.messages, e.toolRegistry.ToLLMTools(), streamCallback)
	if err != nil {
		return nil, err
	}

	if len(resp.ToolCalls) > 0 {
		result.ToolCallCount += len(resp.ToolCalls)

		// 执行每个工具调用
		for _, tc := range resp.ToolCalls {
			// 发送工具调用通知
			streamCallback(fmt.Sprintf("\n\n🔧 [正在调用工具：%s]\n", tc.Name))
			streamCallback(fmt.Sprintf("参数：%s\n", tc.Arguments))

			// 打印工具调用信息
			fmt.Printf("\n[TOOL CALL] 准备调用工具：%s\n", tc.Name)
			fmt.Printf("[TOOL CALL] 参数：%s\n", tc.Arguments)

			// 执行工具
			toolResult, err := e.toolRegistry.ExecuteTool(ctx, tc.Name, tc.Arguments)
			if err != nil {
				fmt.Printf("[TOOL ERROR] 工具执行失败：%v\n", err)
				streamCallback(fmt.Sprintf("❌ [工具执行失败：%v]\n", err))
				return nil, fmt.Errorf("tool execution failed: %w", err)
			}

			// 打印工具执行结果
			fmt.Printf("[TOOL RESULT] 工具 '%s' 执行成功\n", tc.Name)
			if len(toolResult) > 200 {
				fmt.Printf("[TOOL RESULT] 结果预览：%s...\n", toolResult[:200])
			} else {
				fmt.Printf("[TOOL RESULT] 结果：%s\n", toolResult)
			}

			// 发送工具执行结果（截断显示）
			displayResult := toolResult
			if len(toolResult) > 500 {
				displayResult = toolResult[:500] + "..."
			}
			streamCallback(fmt.Sprintf("✅ [工具执行成功]\n结果：%s\n", displayResult))

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

		// 使用更新后的历史再次调用 LLM 获取最终响应
		streamCallback("\n🤔 [工具调用完成，正在生成最终回复...]\n\n")
		fmt.Println("[INFO] 工具调用完成，继续获取最终响应...")

		// 使用流式方式获取最终响应
		finalResp, err := e.llmClient.ChatStream(ctx, e.messages, e.toolRegistry.ToLLMTools(), streamCallback)
		if err != nil {
			return nil, err
		}

		resp.Content = finalResp.Content
		resp.Usage.TotalTokens += finalResp.Usage.TotalTokens
	}

	fmt.Printf("\n[FIRST RESPONSE] 最终响应长度：%d 字符\n", len(resp.Content))

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
