# SQLite 数据库快速入门指南

## 📦 已创建的数据库结构

### 数据表 (4 张)

1. **users** - 用户信息表
   - id: 主键
   - username: 用户名
   - email: 邮箱
   - age: 年龄
   - created_at: 创建时间
   - updated_at: 更新时间

2. **logs** - 系统日志表
   - id: 主键
   - level: 日志级别 (INFO/WARNING/ERROR/DEBUG)
   - message: 日志内容
   - source: 来源模块
   - created_at: 创建时间

3. **tasks** - 任务管理表
   - id: 主键
   - title: 任务标题
   - description: 任务描述
   - status: 状态 (pending/in_progress/completed)
   - priority: 优先级 (1=高，2=中，3=低)
   - due_date: 截止日期
   - created_at: 创建时间
   - completed_at: 完成时间

4. **metrics** - 系统指标表
   - id: 主键
   - name: 指标名称
   - value: 指标值
   - unit: 单位
   - timestamp: 时间戳

## 🚀 快速开始

### 方法 1: 使用批处理文件（推荐）

```bash
# 双击运行或在命令行执行
.\init_db.bat
```

### 方法 2: 使用 PowerShell 脚本

```bash
# 初始化数据库
.\db_manager.ps1

# 或带参数运行
.\db_manager.ps1 -action init
```

### 方法 3: 直接运行 Go 程序

```bash
go run cmd/init_db/main.go
```

##  使用 PowerShell 管理工具

运行交互式管理工具：

```bash
.\db_manager.ps1
```

### 菜单选项

```
1. 初始化数据库（创建表和示例数据）
2. 查看数据库表结构
3. 查询用户数据
4. 查询日志数据
5. 查询任务数据
6. 查询指标数据
7. 清空所有数据
8. 删除数据库文件
0. 退出
```

## 🔍 常用 SQL 查询示例

### 使用 sqlite3 命令行工具

```bash
# 进入 SQLite 交互模式
sqlite3 data.db

# 查看所有表
.tables

# 查看表结构
.schema users

# 查询用户
SELECT * FROM users;

# 查询高优先级任务
SELECT * FROM tasks WHERE priority = 1;

# 查询最近的日志
SELECT * FROM logs ORDER BY created_at DESC LIMIT 10;

# 统计日志级别
SELECT level, COUNT(*) as count FROM logs GROUP BY level;

# 退出
.exit
```

### 在 Go 代码中使用

```go
package main

import (
    "database/sql"
    "log"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    db, err := sql.Open("sqlite3", "./data.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // 查询用户
    rows, err := db.Query("SELECT id, username, email FROM users")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    for rows.Next() {
        var id int
        var username, email string
        if err := rows.Scan(&id, &username, &email); err != nil {
            log.Fatal(err)
        }
        log.Printf("用户：%s (%s)", username, email)
    }
}
```

## 📊 示例数据说明

### users 表（5 条记录）
- 张三 (28 岁)
- 李四 (32 岁)
- 王五 (25 岁)
- 赵六 (30 岁)
- 钱七 (27 岁)

### logs 表（6 条记录）
- INFO: 系统启动成功
- INFO: 用户登录成功
- WARNING: 内存使用率超过 80%
- ERROR: 数据库连接超时
- INFO: 数据备份完成
- DEBUG: API 请求处理完成

### tasks 表（5 条记录）
- 完成项目报告（高优先级，进行中）
- 代码审查（中优先级，待处理）
- 数据库优化（高优先级，待处理）
- 单元测试（低优先级，已完成）
- 文档更新（中优先级，待处理）

### metrics 表（5 条记录）
- CPU 使用率：45.5%
- 内存使用率：68.2%
- 磁盘使用率：52.0%
- 网络吞吐量：125.8 MB/s
- 请求响应时间：89.5 ms

## 🔧 在工作流中使用数据库

### 1. 在 main.go 中注册数据库工具

```go
// 创建数据库工具
dbTool, err := tools.NewDBQueryTool("./data.db")
if err != nil {
    log.Fatal(err)
}
defer dbTool.Close()

// 注册到工具注册表
toolRegistry.Register(dbTool)
```

### 2. 在工作流 YAML 中使用

```yaml
- id: query_users
  type: tool
  tool: db_query
  params:
    query: "SELECT * FROM users WHERE age > 25"
    output_format: json

- id: query_tasks
  type: tool
  tool: db_query
  params:
    query: "SELECT title, status FROM tasks WHERE priority = 1"
    output_format: csv
```

### 3. 结合模板工具转换数据

```yaml
- id: export_users
  type: tool
  tool: db_query
  params:
    query: "SELECT username, email FROM users"
    output_format: json

- id: convert_to_csv
  type: tool
  tool: template
  params:
    input: "{{query_users}}"
    format: csv
    fields: ["username", "email"]
```

## 🛠️ 数据库管理命令

### 备份数据库
```bash
# 导出为 SQL 文件
sqlite3 data.db ".dump" > backup.sql

# 恢复
sqlite3 data.db < backup.sql
```

### 导出为 CSV
```bash
# 导出用户表
sqlite3 -header -csv data.db "SELECT * FROM users;" > users.csv

# 导出任务表
sqlite3 -header -csv data.db "SELECT * FROM tasks;" > tasks.csv
```

### 数据库优化
```bash
# 清理未使用的空间
sqlite3 data.db "VACUUM;"

# 检查数据库完整性
sqlite3 data.db "PRAGMA integrity_check;"
```

## ⚠️ 注意事项

1. **数据库文件位置**: 相对路径 `./data.db`，相对于程序运行目录
2. **并发访问**: SQLite 支持读写锁，高并发场景建议使用其他数据库
3. **数据持久化**: 数据库文件会一直存在，除非手动删除
4. **字符编码**: 使用 UTF-8 编码，支持中文
5. **大小限制**: SQLite 数据库最大 140TB（通常够用）

## 🎯 下一步

数据库创建完成后，您可以：

1. ✅ 运行工作流测试数据库查询功能
2. ✅ 使用 PowerShell 工具查看和管理数据
3. ✅ 在工作流 YAML 中配置数据库查询节点
4. ✅ 结合模板工具导出数据为 CSV/Markdown 等格式

## 📚 相关资源

- [SQLite 官方文档](https://www.sqlite.org/docs.html)
- [SQLite 教程](https://www.sqlitetutorial.net/)
- [Go SQLite3 驱动](https://github.com/mattn/go-sqlite3)
