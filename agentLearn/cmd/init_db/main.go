package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 如果数据库文件不存在，会自动创建
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal("打开数据库失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	fmt.Println("✅ 数据库创建成功：./data.db")

	// 创建示例表
	createTables(db)

	// 插入示例数据
	insertSampleData(db)

	fmt.Println("✅ 示例数据插入成功")
	fmt.Println("\n📊 数据库结构:")
	showTables(db)
}

// 创建表结构
func createTables(db *sql.DB) {
	// 用户表
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL,
		age INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createUsersTable)
	if err != nil {
		log.Fatal("创建 users 表失败:", err)
	}
	fmt.Println("   - 创建 users 表成功")

	// 日志表
	createLogsTable := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		level TEXT NOT NULL,
		message TEXT NOT NULL,
		source TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createLogsTable)
	if err != nil {
		log.Fatal("创建 logs 表失败:", err)
	}
	fmt.Println("   - 创建 logs 表成功")

	// 任务表
	createTasksTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT DEFAULT 'pending',
		priority INTEGER DEFAULT 0,
		due_date DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME
	);`

	_, err = db.Exec(createTasksTable)
	if err != nil {
		log.Fatal("创建 tasks 表失败:", err)
	}
	fmt.Println("   - 创建 tasks 表成功")

	// 指标表
	createMetricsTable := `
	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		value REAL NOT NULL,
		unit TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createMetricsTable)
	if err != nil {
		log.Fatal("创建 metrics 表失败:", err)
	}
	fmt.Println("   - 创建 metrics 表成功")
}

// 插入示例数据
func insertSampleData(db *sql.DB) {
	// 插入用户数据
	users := []struct {
		username string
		email    string
		age      int
	}{
		{"张三", "zhangsan@example.com", 28},
		{"李四", "lisi@example.com", 32},
		{"王五", "wangwu@example.com", 25},
		{"赵六", "zhaoliu@example.com", 30},
		{"钱七", "qianqi@example.com", 27},
	}

	for _, user := range users {
		_, err := db.Exec(
			"INSERT INTO users (username, email, age) VALUES (?, ?, ?)",
			user.username, user.email, user.age,
		)
		if err != nil {
			log.Printf("插入用户 %s 失败：%v", user.username, err)
		}
	}
	fmt.Println("   - 插入 5 条用户数据")

	// 插入日志数据
	logs := []struct {
		level   string
		message string
		source  string
	}{
		{"INFO", "系统启动成功", "system"},
		{"INFO", "用户登录成功", "auth"},
		{"WARNING", "内存使用率超过 80%", "monitor"},
		{"ERROR", "数据库连接超时", "database"},
		{"INFO", "数据备份完成", "backup"},
		{"DEBUG", "API 请求处理完成", "api"},
	}

	for _, logEntry := range logs {
		_, err := db.Exec(
			"INSERT INTO logs (level, message, source) VALUES (?, ?, ?)",
			logEntry.level, logEntry.message, logEntry.source,
		)
		if err != nil {
			log.Printf("插入日志失败：%v", err)
		}
	}
	fmt.Println("   - 插入 6 条日志数据")

	// 插入任务数据
	tasks := []struct {
		title       string
		description string
		status      string
		priority    int
	}{
		{"完成项目报告", "编写 Q4 季度项目总结报告", "in_progress", 1},
		{"代码审查", "审查新提交的代码", "pending", 2},
		{"数据库优化", "优化慢查询语句", "pending", 1},
		{"单元测试", "为核心模块编写单元测试", "completed", 3},
		{"文档更新", "更新 API 文档", "pending", 2},
	}

	for _, task := range tasks {
		_, err := db.Exec(
			"INSERT INTO tasks (title, description, status, priority) VALUES (?, ?, ?, ?)",
			task.title, task.description, task.status, task.priority,
		)
		if err != nil {
			log.Printf("插入任务失败：%v", err)
		}
	}
	fmt.Println("   - 插入 5 条任务数据")

	// 插入指标数据
	metrics := []struct {
		name  string
		value float64
		unit  string
	}{
		{"CPU 使用率", 45.5, "%"},
		{"内存使用率", 68.2, "%"},
		{"磁盘使用率", 52.0, "%"},
		{"网络吞吐量", 125.8, "MB/s"},
		{"请求响应时间", 89.5, "ms"},
	}

	for _, metric := range metrics {
		_, err := db.Exec(
			"INSERT INTO metrics (name, value, unit) VALUES (?, ?, ?)",
			metric.name, metric.value, metric.unit,
		)
		if err != nil {
			log.Printf("插入指标失败：%v", err)
		}
	}
	fmt.Println("   - 插入 5 条指标数据")
}

// 显示表结构
func showTables(db *sql.DB) {
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		log.Fatal("查询表失败:", err)
	}
	defer rows.Close()

	fmt.Println("\n📋 数据表列表:")
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Printf("扫描表名失败：%v", err)
			continue
		}
		if tableName != "sqlite_sequence" { // 跳过系统表
			fmt.Printf("  - %s\n", tableName)
		}
	}

	// 显示每个表的记录数
	tables := []string{"users", "logs", "tasks", "metrics"}
	for _, table := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err == nil {
			fmt.Printf("    %s: %d 条记录\n", table, count)
		}
	}
}
