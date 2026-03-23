package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"goagent/internal/agent"
	"goagent/internal/config"
	"goagent/internal/tools"
	"goagent/pkg/llm"
)

var engine *agent.Engine

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
	engine = agent.NewEngine(llmClient, toolRegistry, systemPrompt)

	// 设置 Gin 模式
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	// CORS 配置
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API 路由
	api := r.Group("/api")
	{
		api.POST("/chat", handleChat)
		api.POST("/chat/stream", handleChatStream)
		api.GET("/history", handleGetHistory)
		api.DELETE("/history", handleClearHistory)
	}

	// 静态文件服务（生产环境部署前端）
	// r.StaticFS("/frontend", http.Dir("../frontend/dist"))

	port := cfg.Server.Port
	if envPort := os.Getenv("SERVER_PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", &port)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Message string `json:"message" binding:"required"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Response string    `json:"response"`
	Usage    llm.Usage `json:"usage"`
	Duration string    `json:"duration"`
	Error    string    `json:"error,omitempty"`
}

func handleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := engine.Chat(c.Request.Context(), req.Message, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{
		Response: result.Response,
		Usage:    result.TotalUsage,
		Duration: result.Duration.String(),
	})
}

func handleChatStream(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置 SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	// 流式回调
	streamCallback := func(token string) {
		fmt.Fprintf(c.Writer, "data: %s\n\n", token)
		flusher.Flush()
	}

	result, err := engine.ChatStream(c.Request.Context(), req.Message, streamCallback)
	if err != nil {
		fmt.Fprintf(c.Writer, "data: [ERROR: %v]\n\n", err)
		flusher.Flush()
		return
	}

	// 发送结束标记
	fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
	flusher.Flush()

	_ = result
}

func handleGetHistory(c *gin.Context) {
	history := engine.GetHistory()
	c.JSON(http.StatusOK, gin.H{"history": history})
}

func handleClearHistory(c *gin.Context) {
	engine.ClearHistory()
	c.JSON(http.StatusOK, gin.H{"success": true})
}
