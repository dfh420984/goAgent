package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ClientConfig MCP 客户端配置
type ClientConfig struct {
	Name    string            `json:"name"`
	Command string            `json:"command"` // stdio 模式下的命令
	Args    []string          `json:"args"`    // stdio 模式下的参数
	Env     map[string]string `json:"env"`     // 环境变量
	Url     string            `json:"url"`     // HTTP/SSE 模式下的 URL
}

// IsStdio 判断是否为 stdio 模式
func (c *ClientConfig) IsStdio() bool {
	return c.Command != ""
}

// IsHttp 判断是否为 HTTP 模式
func (c *ClientConfig) IsHttp() bool {
	return c.Url != ""
}

// Client MCP 客户端
type Client struct {
	name    string
	config  *ClientConfig
	client  *mcp.Client
	session *mcp.ClientSession
	tools   []*Tool
}

// NewClient 创建新的 MCP 客户端
func NewClient(config *ClientConfig) *Client {
	return &Client{
		name:   config.Name,
		config: config,
		client: mcp.NewClient(&mcp.Implementation{
			Name:    "agentlearn-mcp-client",
			Version: "v1.0.0",
		}, nil),
		tools: make([]*Tool, 0),
	}
}

// Name 返回客户端名称
func (c *Client) Name() string {
	return c.name
}

// connect 连接到 MCP 服务器
func (c *Client) connect(ctx context.Context) error {
	// 如果已经连接，跳过
	if c.session != nil && c.session.Ping(ctx, &mcp.PingParams{}) == nil {
		return nil
	}

	var err error
	if c.config.IsStdio() {
		// stdio 模式：通过子进程连接
		cmd := exec.Command(c.config.Command, c.config.Args...)
		for k, v := range c.config.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
		transport := &mcp.CommandTransport{Command: cmd}
		c.session, err = c.client.Connect(ctx, transport, nil)
	} else if c.config.IsHttp() {
		// HTTP/SSE 模式：通过网络连接
		transport := &mcp.StreamableClientTransport{Endpoint: c.config.Url}
		c.session, err = c.client.Connect(ctx, transport, nil)
	} else {
		err = fmt.Errorf("无效的 MCP 服务器配置")
	}

	return err
}

// RefreshTools 刷新工具列表
func (c *Client) RefreshTools(ctx context.Context) error {
	if err := c.connect(ctx); err != nil {
		return err
	}

	// 获取工具列表
	mcpToolResult, err := c.session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return fmt.Errorf("获取工具列表失败：%w", err)
	}

	// 转换为内部 Tool 对象
	c.tools = make([]*Tool, 0)
	for _, mcpTool := range mcpToolResult.Tools {
		tool := &Tool{
			client:  c,
			name:    mcpTool.Name,
			session: c.session,
			mcpTool: mcpTool,
		}
		c.tools = append(c.tools, tool)
	}

	log.Printf("[MCP] 从服务器 %s 加载了 %d 个工具\n", c.name, len(c.tools))
	return nil
}

// GetTools 获取所有工具
func (c *Client) GetTools() []*Tool {
	return c.tools
}

// callTool 调用 MCP 工具
func (c *Client) callTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	if err := c.connect(ctx); err != nil {
		return "", err
	}

	// 序列化参数
	argsJSON, err := json.Marshal(arguments)
	if err != nil {
		return "", fmt.Errorf("参数序列化失败：%w", err)
	}

	// 调用工具
	result, err := c.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: json.RawMessage(argsJSON),
	})
	if err != nil {
		return "", fmt.Errorf("调用工具失败：%w", err)
	}

	// 解析结果
	var builder strings.Builder
	for _, content := range result.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			builder.WriteString(textContent.Text)
		}
	}

	return builder.String(), nil
}

// Tool MCP 工具封装
type Tool struct {
	name    string
	client  *Client
	session *mcp.ClientSession
	mcpTool *mcp.Tool
}

// Name 返回工具名称（带命名空间）
func (t *Tool) Name() string {
	return fmt.Sprintf("mcp_%s__%s", t.client.Name(), t.name)
}

// Description 返回工具描述
func (t *Tool) Description() string {
	if t.mcpTool.Description != "" {
		return t.mcpTool.Description
	}
	return "MCP 工具：" + t.name
}

// InputSchema 返回输入参数 Schema
func (t *Tool) InputSchema() map[string]interface{} {
	// 在 v1.4.0 中，InputSchema 是 *jsonschema.Schema 类型
	if t.mcpTool.InputSchema == nil {
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}
	// 转换为 map[string]interface{}
	schemaMap, err := json.Marshal(t.mcpTool.InputSchema)
	if err != nil {
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}
	var result map[string]interface{}
	if err := json.Unmarshal(schemaMap, &result); err != nil {
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}
	return result
}

// Execute 执行工具
func (t *Tool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	log.Printf("[MCP] 调用工具：%s, 参数：%+v\n", t.name, args)

	result, err := t.client.callTool(ctx, t.name, args)
	if err != nil {
		return "", err
	}

	log.Printf("[MCP] 工具 %s 执行成功，结果长度：%d\n", t.name, len(result))
	return result, nil
}

// Info 返回工具信息（用于 OpenAI API）
func (t *Tool) Info() map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        t.Name(),
			"description": t.Description(),
			"parameters":  t.InputSchema(),
		},
	}
}
