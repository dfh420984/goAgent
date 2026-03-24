# 工作流工具实现总结

## ✅ 已完成的工具

### 1. 数据库查询工具 (`db_query`)
**文件**: `internal/tools/db_query.go`
**功能**: 
- 执行 SQL 查询（支持 SELECT, INSERT, UPDATE, DELETE）
- 支持 SQLite3 数据库
- 输出格式：JSON, CSV, Table

**依赖**:
```bash
go get github.com/mattn/go-sqlite3
```

**使用示例**:
```yaml
- id: query_data
  type: tool
  tool: db_query
  params:
    query: "SELECT * FROM users WHERE created_at > '2024-01-01'"
    output_format: csv
```

### 2. 模板转换工具 (`template`)
**文件**: `internal/tools/template.go`
**功能**:
- 数据格式转换（JSON → CSV/XML/Markdown/Text）
- 自定义模板渲染（Go template 语法）
- 内置函数：now, date, timestamp, upper, lower, title

**使用示例**:
```yaml
- id: convert_data
  type: tool
  tool: template
  params:
    input: '{"data": [{"name": "Alice", "age": 30}]}'
    format: csv
    fields: ["name", "age"]
```

### 3. 天气查询工具 (`weather_query`)
**文件**: `internal/tools/weather.go`
**功能**:
- 全球天气查询
- 实时天气 + 7 天预报
- 使用 OpenMeteo 免费 API（无需 API Key）
- 包含温度、湿度、风速、降水等信息

**使用示例**:
```yaml
- id: check_weather
  type: tool
  tool: weather_query
  params:
    city: "Beijing"
    country: "CN"
    days: 3
```

### 4. 已有工具（之前实现）
- **file_processor**: 文件读写、编辑、删除
- **http_client**: HTTP 请求（GET/POST/PUT/DELETE）
- **shell_executor**: 执行 Shell 命令（带安全检查）

## ✅ 工作流执行器增强

### 文件: `internal/workflow/executor.go`

**新增功能**:
1. ✅ Tool 节点执行（调用工具注册表）
2. ✅ LLM 节点执行（调用 LLM 客户端）
3. ✅ Parallel 节点执行（并行分支）
4. ✅ Join 节点执行（合并结果）
5. ✅ Condition 节点执行（条件判断，基础版）
6. ✅ 变量替换系统（支持 `{{variable}}` 格式）
7. ✅ 内置变量：`{{date}}`, `{{timestamp}}`

**执行器构造函数**:
```go
executor := workflow.NewExecutor(
    ctx,           // context
    workflow,      // 工作流定义
    llmClient,     // LLM 客户端
    toolRegistry,  // 工具注册表
)
```

## ✅ 工作流加载器

### 文件: `internal/workflow/loader.go`

**功能**:
- 从 YAML 文件加载工作流定义
- 自动验证工作流结构
- 检查节点 ID 唯一性
- 验证节点引用关系

**使用示例**:
```go
wf, err := workflow.LoadWorkflow("configs/workflows/report_gen.yaml")
if err != nil {
    log.Fatal(err)
}
```

## ✅ 命令行工具

### 文件: `cmd/workflow/main.go`

**功能**:
- 运行指定的工作流
- 显示执行进度和结果
- 支持命令行参数指定工作流文件

**使用方式**:
```bash
# 运行默认工作流
go run cmd/workflow/main.go

# 运行指定工作流
go run cmd/workflow/main.go configs/workflows/data_export.yaml
```

## 📋 工作流配置文件

### 示例 1: 报表生成 (`report_gen.yaml`)
```yaml
name: 报表生成工作流
description: 自动生成日报/周报
version: "1.0"

nodes:
  - id: start
    type: trigger
    next: collect_data

  - id: collect_data
    type: parallel
    branches:
      - nodes:
          - id: get_metrics
            type: tool
            tool: http_client
            params:
              url: "http://api.internal/metrics"
        next: merge_data
      - nodes:
          - id: get_logs
            type: tool
            tool: http_client
            params:
              url: "http://api.internal/logs"
        next: merge_data

  - id: merge_data
    type: join
    next: generate_report

  - id: generate_report
    type: llm
    prompt: "请根据以下数据生成日报：{{merged_data}}"
    next: save_report

  - id: save_report
    type: tool
    tool: file_processor
    params:
      action: write
      path: "./reports/daily_{{date}}.md"
    next: end

  - id: end
    type: terminator
```

### 示例 2: 数据导出 (`data_export.yaml`)
```yaml
name: 数据导出工作流
description: 从数据库查询数据并导出为 CSV 文件
version: "1.0"

nodes:
  - id: start
    type: trigger
    next: query_db

  - id: query_db
    type: tool
    tool: db_query
    params:
      query: "SELECT * FROM users WHERE created_at > '{{start_date}}'"
    next: transform_data

  - id: transform_data
    type: tool
    tool: template
    params:
      format: csv
    next: save_file

  - id: save_file
    type: tool
    tool: file_processor
    params:
      path: "./exports/data_{{timestamp}}.csv"
    next: end

  - id: end
    type: terminator
```

## 🔧 如何扩展新工具

### 步骤 1: 创建工具文件
```go
// internal/tools/my_tool.go
package tools

import (
    "context"
    "fmt"
)

type MyTool struct{}

func NewMyTool() *MyTool {
    return &MyTool{}
}

func (t *MyTool) Name() string {
    return "my_tool"
}

func (t *MyTool) Description() string {
    return "我的工具描述"
}

func (t *MyTool) Parameters() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type": "string",
                "description": "参数 1",
            },
        },
        "required": []string{"param1"},
    }
}

func (t *MyTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    param1 := args["param1"].(string)
    // 实现逻辑
    result := doSomething(param1)
    return fmt.Sprintf("结果：%s", result), nil
}
```

### 步骤 2: 注册工具
在 `cmd/workflow/main.go` 中添加：
```go
toolRegistry.Register(tools.NewMyTool())
```

### 步骤 3: 在工作流中使用
```yaml
- id: use_my_tool
  type: tool
  tool: my_tool
  params:
    param1: "value"
```

## 🎯 变量替换系统

### 支持的变量格式

1. **节点结果变量**: `{{node_id}}`
   - 自动替换为前一个节点的输出

2. **内置时间变量**:
   - `{{date}}` → "2024-01-15"
   - `{{timestamp}}` → "20240115153045"

3. **在 prompt 中使用**:
```yaml
- id: generate_report
  type: llm
  prompt: |
    请根据以下数据生成报告：
    - 指标数据：{{get_metrics}}
    - 日志数据：{{get_logs}}
    - 日期：{{date}}
```

4. **在 params 中使用**:
```yaml
- id: save_file
  type: tool
  tool: file_processor
  params:
    action: write
    path: "./reports/daily_{{date}}.md"
    content: "{{merged_data}}"
```

## 📦 依赖安装

```bash
# 进入项目目录
cd agentLearn

# 安装 SQLite3 驱动
go get github.com/mattn/go-sqlite3

# 安装 YAML 解析库
go get gopkg.in/yaml.v3

# 更新依赖
go mod tidy
```

## 🚀 运行测试

```bash
# 测试天气工具
go run cmd/workflow/main.go

# 查看工具列表
go run cmd/workflow/main.go --list-tools

# 测试特定工作流
go run cmd/workflow/main.go configs/workflows/report_gen.yaml
```

## 📝 注意事项

1. **数据库工具**: 需要预先创建 SQLite 数据库文件
2. **文件路径**: 使用相对路径时相对于程序运行目录
3. **超时控制**: 使用 context 设置合理的超时时间
4. **错误处理**: 工具执行失败会停止整个工作流
5. **并行执行**: parallel 节点的分支之间无法共享变量

## 🎓 学习资源

- [Go 模板语法](https://pkg.go.dev/text/template)
- [OpenMeteo API](https://open-meteo.com/en/docs)
- [SQLite3 教程](https://www.sqlite.org/docs.html)
- [YAML 格式规范](https://yaml.org/spec/)
