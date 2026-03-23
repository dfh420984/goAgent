package config

import (
	"encoding/json"
	"os"
)

// Config 配置结构
type Config struct {
	LLM      LLMConfig      `json:"llm"`
	Tools    ToolsConfig    `json:"tools"`
	Workflow WorkflowConfig `json:"workflow"`
	Server   ServerConfig   `json:"server"`
}

// LLMConfig LLM 配置
type LLMConfig struct {
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	MaxTokens   int    `json:"max_tokens"`
	Temperature float32 `json:"temperature"`
}

// ToolsConfig 工具配置
type ToolsConfig struct {
	Enabled []string `json:"enabled"`
	Timeout int      `json:"timeout"`
}

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	MaxSteps    int  `json:"max_steps"`
	EnableAudit bool `json:"enable_audit"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port        int      `json:"port"`
	Host        string   `json:"host"`
	CORSOrigins []string `json:"cors_origins"`
}

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		LLM: LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o-mini",
			MaxTokens:   4096,
			Temperature: 0.7,
		},
		Tools: ToolsConfig{
			Enabled: []string{"http_client", "file_processor", "shell_executor"},
			Timeout: 30,
		},
		Workflow: WorkflowConfig{
			MaxSteps:    50,
			EnableAudit: true,
		},
		Server: ServerConfig{
			Port:        8080,
			Host:        "localhost",
			CORSOrigins: []string{"http://localhost:5173"},
		},
	}
}
