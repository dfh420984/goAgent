# 工作流工具使用指南

## 📦 已实现的工具

### 1. **文件处理工具** (`file_processor`)
- **功能**: 读取、写入、编辑、删除文件
- **操作类型**: read, write, append, delete
- **参数**:
  - `action`: 操作类型
  - `path`: 文件路径
  - `content`: 文件内容（写入时需要）

**示例**:
```yaml
- id: read_file
  type: tool
  tool: file_processor
  params:
    action: read
    path: "./data.txt"
```

### 2. **HTTP 客户端工具** (`http_client`)
- **功能**: 发送 HTTP 请求（GET/POST/PUT/DELETE）
- **参数**:
  - `method`: HTTP 方法
  - `url`: 请求 URL
  - `headers`: 请求头（可选）
  - `body`: 请求体（可选）

**示例**:
```yaml
- id: call_api
  type: tool
  tool: http_client
  params:
    method: GET
    url: "https://api.example.com/data"
```

### 3. **Shell 命令执行工具** (`shell_executor`)
- **功能**: 执行系统 shell 命令（带安全检查）
- **参数**:
  - `command`: 要执行的命令
  - `args`: 命令参数（可选）

**示例**:
```yaml
- id: run_command
  type: tool
  tool: shell_executor
  params:
    command: "ls -la"
```

### 4. **天气查询工具** (`weather_query`) ⭐ NEW
- **功能**: 查询全球任意地点的实时天气和天气预报
- **API**: OpenMeteo（免费，无需 API Key）
- **参数**:
  - `city`: 城市名称（必需）
  - `country`: 国家代码（可选）
  - `days`: 预报天数 1-7 天（可选，默认 1 天）

**示例**:
```yaml
- id: check_weather
  type: tool
  tool: weather_query
  params:
    city: "Beijing"
    country: "CN"
    days: 3
```

**输出**:
```
📍 地点：Beijing, CN (纬度：39.90, 经度：116.40)

🌡️ 当前天气：
   温度：25.5°C
   湿度：60%
   风速：12.5 km/h (东南)
   天气：☀️ 晴朗

📅 天气预报：
   01 月 15 日：☀️ 晴朗  18°C ~ 28°C
   01 月 16 日：⛅ 部分多云  16°C ~ 25°C
   01 月 17 日：🌧️ 中雨  15°C ~ 22°C  🌧️ 降水：5.2mm

💡 数据来源：Open-Meteo (免费天气 API)
```

### 5. **数据库查询工具** (`db_query`) ⭐ NEW
- **功能**: 执行 SQL 查询，从数据库获取数据
- **后端**: SQLite3（可扩展支持其他数据库）
- **参数**:
  - `query`: SQL 查询语句（必需）
  - `params`: 查询参数（可选，防止 SQL 注入）
  - `output_format`: 输出格式 json/csv/table（可选，默认 json）

**示例**:
```yaml
- id: query_users
  type: tool
  tool: db_query
  params:
    query: "SELECT * FROM users WHERE created_at > '2024-01-01'"
    output_format: csv
```

**注意**: 需要在 `main.go` 中初始化数据库连接：
```go
dbTool, err := tools.NewDBQueryTool("./data.db")
if err != nil {
    log.Fatal(err)
}
defer dbTool.Close()
toolRegistry.Register(dbTool)
```

### 6. **模板转换工具** (`template`) ⭐ NEW
- **功能**: 数据格式转换和模板渲染
- **支持的格式**: CSV, XML, Markdown, Text
- **参数**:
  - `input`: 输入数据（JSON 格式，必需）
  - `format`: 目标格式 csv/xml/markdown/text（必需）
  - `template`: 自定义模板（可选，Go template 语法）
  - `fields`: 要输出的字段列表（可选）

**示例 1 - 转换为 CSV**:
```yaml
- id: convert_to_csv
  type: tool
  tool: template
  params:
    input: '{"data": [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}]}'
    format: csv
    fields: ["name", "age"]
```

**示例 2 - 使用自定义模板**:
```yaml
- id: render_report
  type: tool
  tool: template
  params:
    input: '{"title": "日报", "content": "今天完成了..."}'
    format: text
    template: |
      报告标题：{{.title}}
      生成时间：{{now}}
      报告内容：{{.content}}
```

**内置模板函数**:
- `{{now}}`: 当前时间（2006-01-02 15:04:05）
- `{{date}}`: 当前日期（2006-01-02）
- `{{timestamp}}`: 时间戳（20060102150405）
- `{{upper "text"}}`: 转大写
- `{{lower "TEXT"}}`: 转小写
- `{{title "hello"}}`: 首字母大写

## 🛠️ 工作流节点类型

### 1. **Trigger** (`trigger`)
- 工作流的起始节点
- 不需要特殊配置

### 2. **Tool** (`tool`)
- 调用已注册的工具
- 需要指定 `tool` 和 `params`

### 3. **LLM** (`llm`)
- 调用大语言模型
- 需要指定 `prompt`（支持变量替换）

### 4. **Parallel** (`parallel`)
- 并行执行多个分支
- 每个分支包含一系列节点

### 5. **Join** (`join`)
- 合并并行分支的结果
- 自动存储到 `merged_data` 变量

### 6. **Condition** (`condition`)
- 条件判断（暂未完全实现）
- 根据条件执行不同分支

### 7. **Terminator** (`terminator`)
- 工作流结束节点

## 📝 变量替换

工作流支持在 `params` 和 `prompt` 中使用变量：

### 变量格式
```yaml
params:
  path: "./reports/daily_{{date}}.md"
  query: "SELECT * FROM logs WHERE date = '{{timestamp}}'"
```

### 可用变量
- `{{date}}`: 当前日期
- `{{timestamp}}`: 当前时间戳
- `{{node_id}}`: 前一个节点的输出结果

## 🚀 使用示例

### 示例 1: 运行工作流
```bash
# 运行默认工作流（report_gen.yaml）
go run cmd/workflow/main.go

# 运行指定工作流
go run cmd/workflow/main.go configs/workflows/data_export.yaml
```

### 示例 2: 在代码中使用
```go
// 加载工作流
wf, err := workflow.LoadWorkflow("configs/workflows/report_gen.yaml")
if err != nil {
    log.Fatal(err)
}

// 创建执行器
executor := workflow.NewExecutor(ctx, wf, llmClient, toolRegistry)

// 执行工作流
results, err := executor.Execute("start")
if err != nil {
    log.Fatal(err)
}

// 处理结果
for key, value := range results {
    fmt.Printf("%s: %v\n", key, value)
}
```

## 🔧 扩展新工具

### 步骤 1: 实现 Tool 接口
```go
type MyCustomTool struct {
    // 工具状态
}

func NewMyCustomTool() *MyCustomTool {
    return &MyCustomTool{}
}

func (t *MyCustomTool) Name() string {
    return "my_custom_tool"
}

func (t *MyCustomTool) Description() string {
    return "我的自定义工具描述"
}

func (t *MyCustomTool) Parameters() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type": "string",
                "description": "参数 1 描述",
            },
        },
        "required": []string{"param1"},
    }
}

func (t *MyCustomTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    // 实现工具逻辑
    param1 := args["param1"].(string)
    result := doSomething(param1)
    return result, nil
}
```

### 步骤 2: 注册工具
```go
toolRegistry := tools.NewToolRegistry()
toolRegistry.Register(tools.NewMyCustomTool())
```

### 步骤 3: 在工作流中使用
```yaml
- id: use_custom
  type: tool
  tool: my_custom_tool
  params:
    param1: "value"
```

## 📋 工作流配置文件示例

### 报表生成工作流 (`report_gen.yaml`)
```yaml
name: 报表生成工作流
description: 自动生成日报/周报
version: "1.0"

nodes:
  - id: start
    type: trigger
    name: 开始
    next: collect_data

  - id: collect_data
    type: parallel
    name: 并行收集数据
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
    name: 合并数据
    next: generate_report

  - id: generate_report
    type: llm
    name: 生成报表
    prompt: "请根据以下数据生成日报：{{merged_data}}"
    next: save_report

  - id: save_report
    type: tool
    name: 保存报告
    tool: file_processor
    params:
      action: write
      path: "./reports/daily_{{date}}.md"
    next: end

  - id: end
    type: terminator
    name: 结束
```

## 🎯 最佳实践

1. **错误处理**: 工具执行失败时会停止整个工作流
2. **超时控制**: 使用 context 设置超时
3. **日志记录**: 在关键节点添加日志
4. **变量命名**: 使用有意义的节点 ID 便于调试
5. **并行优化**: 将独立的任务放在 parallel 分支中

## 📚 相关资源

- [Go 语言模板语法](https://pkg.go.dev/text/template)
- [OpenMeteo API 文档](https://open-meteo.com/en/docs)
- [SQLite3 文档](https://www.sqlite.org/docs.html)
