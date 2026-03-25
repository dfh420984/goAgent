package bootstrap

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"goagent/internal/agent"
	"goagent/internal/config"
	mcp "goagent/internal/mcp"
	"goagent/internal/server"
	"goagent/internal/tools"
	"goagent/pkg/llm"
	pkgmcp "goagent/pkg/mcp"
)

// App 应用实例
type App struct {
	engine       *agent.Engine
	toolRegistry *tools.ToolRegistry
	mcpClients   []*pkgmcp.Client
}

// Bootstrap 初始化应用
func Bootstrap() (*App, error) {
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

	// 创建工具注册表并注册本地工具
	toolRegistry := tools.NewToolRegistry()
	toolRegistry.Register(tools.NewFileProcessorTool())
	toolRegistry.Register(tools.NewHTTPClientTool())
	toolRegistry.Register(tools.NewShellExecutorTool())
	toolRegistry.Register(tools.NewWeatherTool())

	// 加载 MCP 配置并注册 MCP 工具
	log.Println("📦 正在加载 MCP 服务器配置...")
	mcpLoader := mcp.NewLoader("mcp-servers.json")
	mcpClients, err := mcpLoader.LoadAndRegister(toolRegistry)
	if err != nil {
		log.Printf("⚠️  MCP 加载失败：%v", err)
	} else if len(mcpClients) > 0 {
		log.Printf("✅ 成功加载 %d 个 MCP 服务器", len(mcpClients))
	} else {
		log.Println("⚠️  未找到可用的 MCP 服务器")
	}

	// 创建 Agent 引擎
	systemPrompt := "你是一个智能助手，可以帮助用户完成各种任务。你可以使用以下工具：文件处理、HTTP 请求、Shell 命令、天气查询。此外，你还可以使用文件系统工具来读取、写入和管理文件，使用内存工具来存储和检索知识。"
	engine := agent.NewEngine(llmClient, toolRegistry, systemPrompt)

	return &App{
		engine:       engine,
		toolRegistry: toolRegistry,
		mcpClients:   mcpClients,
	}, nil
}

// CreateServer 创建 HTTP 服务器
func (app *App) CreateServer(cfg *config.ServerConfig) *server.Server {
	handler := server.NewHandler(app.engine)
	serverConfig := &server.Config{
		Host: cfg.Host,
		Port: cfg.Port,
	}

	// 支持环境变量覆盖端口
	if envPort := os.Getenv("SERVER_PORT"); envPort != "" {
		var port int
		if _, err := fmt.Sscanf(envPort, "%d", &port); err == nil {
			serverConfig.Port = port
		}
	}

	return server.NewServer(serverConfig, handler)
}
