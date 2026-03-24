package tools

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"
)

// TemplateTool 模板转换工具
type TemplateTool struct{}

// NewTemplateTool 创建模板转换工具
func NewTemplateTool() *TemplateTool {
	return &TemplateTool{}
}

// Name 工具名称
func (t *TemplateTool) Name() string {
	return "template"
}

// Description 工具描述
func (t *TemplateTool) Description() string {
	return "数据格式转换和模板渲染。支持 JSON、CSV、XML 等格式转换，以及自定义模板渲染。"
}

// Parameters 工具参数定义
func (t *TemplateTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"input": map[string]interface{}{
				"type":        "string",
				"description": "输入数据（JSON 格式）",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "目标格式：csv, xml, markdown, text",
				"enum":        []string{"csv", "xml", "markdown", "text"},
			},
			"template": map[string]interface{}{
				"type":        "string",
				"description": "自定义模板（Go template 语法）",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"description": "要输出的字段列表",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"input", "format"},
	}
}

// Execute 执行模板转换
func (t *TemplateTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	inputArg, exists := args["input"]
	if !exists {
		return "", fmt.Errorf("input 参数是必需的")
	}

	// 处理 nil 值
	if inputArg == nil {
		return "", fmt.Errorf("input 参数不能为 nil")
	}

	var inputStr string
	var ok bool
	if inputStr, ok = inputArg.(string); !ok {
		// 如果不是字符串，尝试转换为 JSON
		jsonData, err := json.Marshal(inputArg)
		if err != nil {
			return "", fmt.Errorf("input 参数类型错误且无法转换为 JSON: %w", err)
		}
		inputStr = string(jsonData)
	}

	format, ok := args["format"].(string)
	if !ok {
		return "", fmt.Errorf("format 参数是必需的")
	}

	// 解析输入数据
	var data interface{}
	if err := json.Unmarshal([]byte(inputStr), &data); err != nil {
		return "", fmt.Errorf("解析输入 JSON 失败：%w", err)
	}

	// 根据格式进行转换
	switch format {
	case "csv":
		return t.convertToCSV(data, args)

	case "xml":
		return t.convertToXML(data, args)

	case "markdown":
		return t.convertToMarkdown(data, args)

	case "text":
		return t.convertToText(data, args)

	default:
		return "", fmt.Errorf("不支持的格式：%s", format)
	}
}

// convertToCSV 转换为 CSV 格式
func (t *TemplateTool) convertToCSV(data interface{}, args map[string]interface{}) (string, error) {
	// 期望数据是数组
	var dataArray []interface{}

	// 尝试不同的数据类型
	switch v := data.(type) {
	case []interface{}:
		dataArray = v
	case string:
		// 如果是字符串，尝试解析为 JSON
		if err := json.Unmarshal([]byte(v), &dataArray); err != nil {
			return "", fmt.Errorf("解析 JSON 字符串失败：%w", err)
		}
	default:
		return "", fmt.Errorf("CSV 转换需要数组类型的数据，当前类型：%T", data)
	}

	if len(dataArray) == 0 {
		return "", fmt.Errorf("CSV 转换需要非空数组")
	}

	// 获取字段列表
	var fields []string
	if fieldsArg, ok := args["fields"].([]interface{}); ok {
		fields = make([]string, len(fieldsArg))
		for i, f := range fieldsArg {
			if fs, ok := f.(string); ok {
				fields[i] = fs
			}
		}
	}

	// 如果没有指定字段，从第一个对象获取所有键
	if len(fields) == 0 {
		if firstObj, ok := dataArray[0].(map[string]interface{}); ok {
			fields = make([]string, 0, len(firstObj))
			for k := range firstObj {
				fields = append(fields, k)
			}
		}
	}

	// 创建 CSV writer
	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	// 写入表头
	if err := writer.Write(fields); err != nil {
		return "", fmt.Errorf("写入 CSV 表头失败：%w", err)
	}

	// 写入数据行
	for _, item := range dataArray {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		record := make([]string, len(fields))
		for i, field := range fields {
			if val, exists := obj[field]; exists && val != nil {
				record[i] = fmt.Sprintf("%v", val)
			}
		}

		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("写入 CSV 行失败：%w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV 刷新失败：%w", err)
	}

	return sb.String(), nil
}

// convertToXML 转换为 XML 格式
func (t *TemplateTool) convertToXML(data interface{}, args map[string]interface{}) (string, error) {
	var sb strings.Builder

	sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")

	// 处理不同的数据类型
	switch v := data.(type) {
	case []interface{}:
		// 直接处理数组
		sb.WriteString("<root>\n")
		for _, item := range v {
			sb.WriteString("  <item>\n")
			if obj, ok := item.(map[string]interface{}); ok {
				for k, val := range obj {
					sb.WriteString(fmt.Sprintf("    <%s>%v</%s>\n", k, val, k))
				}
			}
			sb.WriteString("  </item>\n")
		}
		sb.WriteString("</root>")
	case string:
		// 如果是字符串，尝试解析为 JSON
		var parsedData interface{}
		if err := json.Unmarshal([]byte(v), &parsedData); err != nil {
			return "", fmt.Errorf("解析 JSON 字符串失败：%w", err)
		}
		// 递归调用
		return t.convertToXML(parsedData, args)
	case map[string]interface{}:
		// 单个对象
		sb.WriteString("<root>\n")
		for k, val := range v {
			sb.WriteString(fmt.Sprintf("  <%s>%v</%s>\n", k, val, k))
		}
		sb.WriteString("</root>")
	default:
		return "", fmt.Errorf("XML 转换不支持的数据类型：%T", data)
	}

	return sb.String(), nil
}

// convertToMarkdown 转换为 Markdown 表格
func (t *TemplateTool) convertToMarkdown(data interface{}, args map[string]interface{}) (string, error) {
	// 期望数据是数组
	var dataArray []interface{}

	// 尝试不同的数据类型
	switch v := data.(type) {
	case []interface{}:
		dataArray = v
	case string:
		// 如果是字符串，尝试解析为 JSON
		if err := json.Unmarshal([]byte(v), &dataArray); err != nil {
			return "", fmt.Errorf("解析 JSON 字符串失败：%w", err)
		}
	default:
		return "", fmt.Errorf("Markdown 转换需要数组类型的数据，当前类型：%T", data)
	}

	if len(dataArray) == 0 {
		return "", fmt.Errorf("Markdown 转换需要非空数组")
	}

	// 获取字段列表
	var fields []string
	if fieldsArg, ok := args["fields"].([]interface{}); ok {
		fields = make([]string, len(fieldsArg))
		for i, f := range fieldsArg {
			if fs, ok := f.(string); ok {
				fields[i] = fs
			}
		}
	}

	// 如果没有指定字段，从第一个对象获取所有键
	if len(fields) == 0 {
		if firstObj, ok := dataArray[0].(map[string]interface{}); ok {
			fields = make([]string, 0, len(firstObj))
			for k := range firstObj {
				fields = append(fields, k)
			}
		}
	}

	var sb strings.Builder

	// 表头
	sb.WriteString("| ")
	for _, field := range fields {
		sb.WriteString(field)
		sb.WriteString(" | ")
	}
	sb.WriteString("\n")

	// 分隔线
	sb.WriteString("|")
	for range fields {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")

	// 数据行
	for _, item := range dataArray {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		sb.WriteString("| ")
		for _, field := range fields {
			if val, exists := obj[field]; exists && val != nil {
				sb.WriteString(fmt.Sprintf("%v | ", val))
			} else {
				sb.WriteString(" | ")
			}
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// convertToText 转换为纯文本
func (t *TemplateTool) convertToText(data interface{}, args map[string]interface{}) (string, error) {
	// 检查是否有自定义模板
	if tmplStr, ok := args["template"].(string); ok && tmplStr != "" {
		return t.renderTemplate(data, tmplStr)
	}

	// 默认 JSON 格式化输出
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 格式化失败：%w", err)
	}

	return string(jsonData), nil
}

// renderTemplate 渲染 Go template
func (t *TemplateTool) renderTemplate(data interface{}, tmplStr string) (string, error) {
	// 添加自定义函数
	funcMap := template.FuncMap{
		"now": func() string {
			return time.Now().Format("2006-01-02 15:04:05")
		},
		"date": func() string {
			return time.Now().Format("2006-01-02")
		},
		"timestamp": func() string {
			return time.Now().Format("20060102150405")
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": func(s string) string {
			return strings.Title(strings.ToLower(s))
		},
	}

	tmpl, err := template.New("output").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("解析模板失败：%w", err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("执行模板失败：%w", err)
	}

	return sb.String(), nil
}
