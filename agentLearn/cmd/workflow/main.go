package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"goagent/internal/config"
	"goagent/internal/tools"
	"goagent/internal/workflow"
	"goagent/pkg/llm"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
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

	// 加载配置
	cfg, err := config.LoadConfig("configs/config.json")
	if err != nil {
		log.Printf("Warning: failed to load config, using default: %v", err)
		cfg = config.DefaultConfig()
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
	toolRegistry.Register(tools.NewWeatherTool())

	// 注册数据库查询工具（如果需要使用）
	dbTool, err := tools.NewDBQueryTool("./data.db")
	if err == nil {
		toolRegistry.Register(dbTool)
		defer dbTool.Close()
	}

	// 注册模板转换工具
	toolRegistry.Register(tools.NewTemplateTool())

	// 加载工作流
	workflowFile := "configs/workflows/data_export.yaml"
	if len(os.Args) > 1 {
		workflowFile = os.Args[1]
	}

	fmt.Printf("🚀 加载工作流：%s\n", workflowFile)
	wf, err := workflow.LoadWorkflow(workflowFile)
	if err != nil {
		log.Fatalf("加载工作流失败：%v", err)
	}

	fmt.Printf("✅ 工作流加载成功：%s\n", wf.Name)
	fmt.Printf("📝 描述：%s\n", wf.Description)
	fmt.Printf("📦 节点数量：%d\n\n", len(wf.Nodes))

	// 创建工作流执行器
	ctx := context.Background()
	executor := workflow.NewExecutor(ctx, wf, llmClient, toolRegistry)

	// 执行工作流
	fmt.Println("⚙️  开始执行工作流...")
	results, err := executor.Execute("start")
	if err != nil {
		log.Fatalf("执行工作流失败：%v", err)
	}

	// 输出结果
	fmt.Println("\n✅ 工作流执行完成！")
	fmt.Println("📊 执行结果:")
	for key, value := range results {
		fmt.Printf("  %s: %v\n", key, value)
	}
}
