package main

import (
	"goagent/internal/config"
	"log"

	"goagent/internal/bootstrap"
)

func main() {
	// 初始化应用
	app, err := bootstrap.Bootstrap()
	if err != nil {
		log.Fatalf("Failed to bootstrap application: %v", err)
	}

	// 加载配置创建服务器
	cfg, _ := config.LoadConfig("configs/config.json")
	server := app.CreateServer(&cfg.Server)

	// 启动服务器
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
