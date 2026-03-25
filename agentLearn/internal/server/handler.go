package server

import (
	"github.com/gin-gonic/gin"
	"net/http"

	"goagent/internal/agent"
	"goagent/pkg/llm"
)

// Handler HTTP 请求处理器
type Handler struct {
	engine *agent.Engine
}

// NewHandler 创建 HTTP 处理器
func NewHandler(engine *agent.Engine) *Handler {
	return &Handler{
		engine: engine,
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

// HandleChat 处理聊天请求（非流式）
func (h *Handler) HandleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.engine.Chat(c.Request.Context(), req.Message, nil)
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

// HandleChatStream 处理聊天请求（流式）
func (h *Handler) HandleChatStream(c *gin.Context) {
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
		_, _ = c.Writer.WriteString("data: " + token + "\n\n")
		flusher.Flush()
	}

	result, err := h.engine.ChatStream(c.Request.Context(), req.Message, streamCallback)
	if err != nil {
		_, _ = c.Writer.WriteString("data: [ERROR: " + err.Error() + "]\n\n")
		flusher.Flush()
		return
	}

	// 发送结束标记
	_, _ = c.Writer.WriteString("data: [DONE]\n\n")
	flusher.Flush()

	_ = result
}

// HandleGetHistory 获取历史记录
func (h *Handler) HandleGetHistory(c *gin.Context) {
	history := h.engine.GetHistory()
	c.JSON(http.StatusOK, gin.H{"history": history})
}

// HandleClearHistory 清空历史记录
func (h *Handler) HandleClearHistory(c *gin.Context) {
	h.engine.ClearHistory()
	c.JSON(http.StatusOK, gin.H{"success": true})
}
