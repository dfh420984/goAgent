package server

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// Config 服务器配置
type Config struct {
	Host string
	Port int
}

// Server HTTP 服务器
type Server struct {
	engine *gin.Engine
	config *Config
}

// NewServer 创建 HTTP 服务器
func NewServer(config *Config, handler *Handler) *Server {
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
		api.POST("/chat", handler.HandleChat)
		api.POST("/chat/stream", handler.HandleChatStream)
		api.GET("/history", handler.HandleGetHistory)
		api.DELETE("/history", handler.HandleClearHistory)
	}

	return &Server{
		engine: r,
		config: config,
	}
}

// Run 启动服务器
func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	log.Printf("Starting server on %s", addr)
	return s.engine.Run(addr)
}
