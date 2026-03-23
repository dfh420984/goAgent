package tools_test

import (
	"context"
	"testing"

	"goagent/internal/tools"
)

func TestFileProcessorTool(t *testing.T) {
	tool := tools.NewFileProcessorTool()
	ctx := context.Background()

	// 测试写入文件
	result, err := tool.Execute(ctx, map[string]interface{}{
		"action":  "write",
		"path":    "./test_data/test.txt",
		"content": "Hello, World!",
	})
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	t.Logf("Write result: %s", result)

	// 测试读取文件
	result, err = tool.Execute(ctx, map[string]interface{}{
		"action": "read",
		"path":   "./test_data/test.txt",
	})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if result != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", result)
	}
	t.Logf("Read result: %s", result)
}

func TestHTTPClientTool(t *testing.T) {
	tool := tools.NewHTTPClientTool()
	ctx := context.Background()

	// 测试 GET 请求
	result, err := tool.Execute(ctx, map[string]interface{}{
		"method": "GET",
		"url":    "https://httpbin.org/get",
	})
	if err != nil {
		t.Fatalf("HTTP GET failed: %v", err)
	}
	t.Logf("HTTP GET result: %s", result)
}

func TestToolRegistry(t *testing.T) {
	registry := tools.NewToolRegistry()
	
	// 注册工具
	registry.Register(tools.NewFileProcessorTool())
	registry.Register(tools.NewHTTPClientTool())

	// 测试获取工具
	tool, ok := registry.Get("file_processor")
	if !ok {
		t.Fatal("Failed to get file_processor tool")
	}
	if tool.Name() != "file_processor" {
		t.Errorf("Expected file_processor, got %s", tool.Name())
	}

	// 测试列出所有工具
	allTools := registry.List()
	if len(allTools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(allTools))
	}
}
