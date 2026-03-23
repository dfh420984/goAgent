package tools

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
)

// ShellExecutorTool Shell 命令执行工具
type ShellExecutorTool struct{}

func NewShellExecutorTool() *ShellExecutorTool {
	return &ShellExecutorTool{}
}

func (t *ShellExecutorTool) Name() string {
	return "shell_executor"
}

func (t *ShellExecutorTool) Description() string {
	return "执行系统 shell 命令。注意：此工具可能危险，应该限制在安全的环境中使用"
}

func (t *ShellExecutorTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "要执行的 shell 命令",
			},
			"args": map[string]interface{}{
				"type":        "array",
				"description": "命令参数列表",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"command"},
	}
}

func (t *ShellExecutorTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	command, ok := args["command"].(string)
	if !ok {
		return "", fmt.Errorf("command is required")
	}

	// 安全检查：禁止某些危险命令
	blockedCommands := []string{"rm -rf", "del /f", "format", "mkfs"}
	for _, blocked := range blockedCommands {
		if len(command) >= len(blocked) && command[:len(blocked)] == blocked {
			return "", fmt.Errorf("blocked dangerous command: %s", blocked)
		}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Command output: %s\nError: %v", string(output), err), nil
	}

	return fmt.Sprintf("Command executed successfully:\n%s", string(output)), nil
}
