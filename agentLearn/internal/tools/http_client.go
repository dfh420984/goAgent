package tools

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// HTTPClientTool HTTP 请求工具
type HTTPClientTool struct {
	client *resty.Client
}

func NewHTTPClientTool() *HTTPClientTool {
	return &HTTPClientTool{
		client: resty.New(),
	}
}

func (t *HTTPClientTool) Name() string {
	return "http_client"
}

func (t *HTTPClientTool) Description() string {
	return "发送 HTTP 请求（GET/POST/PUT/DELETE），调用外部 API 接口"
}

func (t *HTTPClientTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"method": map[string]interface{}{
				"type":        "string",
				"description": "HTTP 方法",
				"enum":        []string{"GET", "POST", "PUT", "DELETE"},
			},
			"url": map[string]interface{}{
				"type":        "string",
				"description": "请求 URL",
			},
			"headers": map[string]interface{}{
				"type":        "object",
				"description": "请求头",
			},
			"body": map[string]interface{}{
				"type":        "string",
				"description": "请求体（JSON 格式）",
			},
		},
		"required": []string{"method", "url"},
	}
}

func (t *HTTPClientTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	method, ok := args["method"].(string)
	if !ok {
		return "", fmt.Errorf("method is required")
	}

	url, ok := args["url"].(string)
	if !ok {
		return "", fmt.Errorf("url is required")
	}

	req := t.client.R().SetContext(ctx)

	// 设置 headers
	if headers, ok := args["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if vs, ok := v.(string); ok {
				req.SetHeader(k, vs)
			}
		}
	}

	// 设置 body
	if body, ok := args["body"].(string); ok && body != "" {
		req.SetBody(body)
	}

	// 执行请求
	var resp *resty.Response
	var err error

	switch method {
	case "GET":
		resp, err = req.Get(url)
	case "POST":
		resp, err = req.Post(url)
	case "PUT":
		resp, err = req.Put(url)
	case "DELETE":
		resp, err = req.Delete(url)
	default:
		return "", fmt.Errorf("unsupported method: %s", method)
	}

	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}

	return fmt.Sprintf("Status: %d\nBody: %s", resp.StatusCode(), resp.String()), nil
}
