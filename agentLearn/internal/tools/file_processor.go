package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FileProcessorTool 文件处理工具
type FileProcessorTool struct{}

func NewFileProcessorTool() *FileProcessorTool {
	return &FileProcessorTool{}
}

func (t *FileProcessorTool) Name() string {
	return "file_processor"
}

func (t *FileProcessorTool) Description() string {
	return "读取、写入、编辑文件。支持读取文件内容、写入新文件、追加内容等操作"
}

func (t *FileProcessorTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "操作类型：read, write, append, delete",
				"enum":        []string{"read", "write", "append", "delete"},
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "文件路径",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "文件内容（写入时需要）",
			},
		},
		"required": []string{"action", "path"},
	}
}

func (t *FileProcessorTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, ok := args["action"].(string)
	if !ok {
		return "", fmt.Errorf("action is required")
	}

	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path is required")
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	switch action {
	case "read":
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return string(content), nil

	case "write":
		content, ok := args["content"].(string)
		if !ok {
			return "", fmt.Errorf("content is required for write action")
		}
		if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("failed to write file: %w", err)
		}
		return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil

	case "append":
		content, ok := args["content"].(string)
		if !ok {
			return "", fmt.Errorf("content is required for append action")
		}
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()
		if _, err := f.WriteString(content); err != nil {
			return "", fmt.Errorf("failed to append to file: %w", err)
		}
		return fmt.Sprintf("Successfully appended content to %s", path), nil

	case "delete":
		if err := os.Remove(path); err != nil {
			return "", fmt.Errorf("failed to delete file: %w", err)
		}
		return fmt.Sprintf("Successfully deleted %s", path), nil

	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}
