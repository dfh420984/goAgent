package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"goagent/internal/agent"
	"goagent/internal/config"
	"goagent/internal/tools"
	"goagent/pkg/llm"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// 加载配置
	cfg, err := config.LoadConfig("configs/config.json")
	if err != nil {
		log.Printf("Warning: failed to load config, using default: %v", err)
		cfg = config.DefaultConfig()
	}

	// 从环境变量获取 API Key
	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	if apiKey == "" {
		log.Fatal("Error: OPENAI_API_KEY environment variable is required")
	}

	// 创建 LLM 客户端
	llmClient, err := llm.NewClient(llm.Config{
		APIKey:      apiKey,
		BaseURL:     baseURL,
		Model:       cfg.LLM.Model,
		MaxTokens:   cfg.LLM.MaxTokens,
		Temperature: cfg.LLM.Temperature,
	})
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}

	// 创建工具注册表并注册工具
	toolRegistry := tools.NewToolRegistry()
	toolRegistry.Register(tools.NewFileProcessorTool())
	toolRegistry.Register(tools.NewHTTPClientTool())
	toolRegistry.Register(tools.NewShellExecutorTool())

	// 创建 Agent 引擎
	systemPrompt := "你是一个智能助手，可以帮助用户完成各种任务。你可以使用以下工具：文件处理、HTTP 请求、Shell 命令。"
	engine := agent.NewEngine(llmClient, toolRegistry, systemPrompt)

	// 测试对话
	ctx := context.Background()
	fmt.Println("=== GoAgent TaskRunner ===")
	fmt.Println("输入 'quit' 退出程序")
	fmt.Println()

	for {
		fmt.Print("You: ")
		var input string
		fmt.Scanln(&input)

		if input == "quit" || input == "exit" {
			break
		}

		result, err := engine.Chat(ctx, input, nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Assistant: %s\n", result.Response)
		fmt.Printf("(Tokens: %d, Duration: %v, Tool calls: %d)\n", 
			result.TotalUsage.TotalTokens, result.Duration, result.ToolCallCount)
		fmt.Println()
	}

	fmt.Println("Goodbye!")
}
