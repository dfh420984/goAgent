package tools

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// DBQueryTool 数据库查询工具
type DBQueryTool struct {
	db *sql.DB
}

// NewDBQueryTool 创建数据库查询工具
func NewDBQueryTool(dbPath string) (*DBQueryTool, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("无法连接数据库：%w", err)
	}

	return &DBQueryTool{
		db: db,
	}, nil
}

// Name 工具名称
func (t *DBQueryTool) Name() string {
	return "db_query"
}

// Description 工具描述
func (t *DBQueryTool) Description() string {
	return "执行 SQL 查询，从数据库获取数据。支持 SELECT、INSERT、UPDATE、DELETE 等操作。"
}

// Parameters 工具参数定义
func (t *DBQueryTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "SQL 查询语句",
			},
			"params": map[string]interface{}{
				"type":        "array",
				"description": "查询参数（用于防止 SQL 注入）",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"output_format": map[string]interface{}{
				"type":        "string",
				"description": "输出格式：json, csv, table",
				"enum":        []string{"json", "csv", "table"},
			},
		},
		"required": []string{"query"},
	}
}

// Execute 执行数据库查询
func (t *DBQueryTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("query 参数是必需的")
	}

	outputFormat := "json"
	if format, ok := args["output_format"].(string); ok {
		outputFormat = format
	}

	// 执行查询
	rows, err := t.db.QueryContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("查询执行失败：%w", err)
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("获取列名失败：%w", err)
	}

	// 读取所有数据
	var results []map[string]interface{}
	for rows.Next() {
		// 创建扫描器
		values := make([]interface{}, len(columns))
		scanArgs := make([]interface{}, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		// 扫描数据
		if err := rows.Scan(scanArgs...); err != nil {
			return "", fmt.Errorf("扫描行失败：%w", err)
		}

		// 转换为 map
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// 处理 nil 值
			if val == nil {
				rowMap[col] = nil
			} else {
				// 尝试转换为具体类型
				switch v := val.(type) {
				case []byte:
					rowMap[col] = string(v)
				default:
					rowMap[col] = v
				}
			}
		}
		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("遍历结果失败：%w", err)
	}

	// 根据格式输出
	switch outputFormat {
	case "json":
		jsonData, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return "", fmt.Errorf("JSON 格式化失败：%w", err)
		}
		return string(jsonData), nil

	case "csv":
		var sb strings.Builder
		writer := csv.NewWriter(&sb)

		// 写入表头
		if err := writer.Write(columns); err != nil {
			return "", fmt.Errorf("写入 CSV 表头失败：%w", err)
		}

		// 写入数据行
		for _, row := range results {
			record := make([]string, len(columns))
			for i, col := range columns {
				if val, ok := row[col]; ok && val != nil {
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

	case "table":
		var sb strings.Builder
		// 写入表头
		sb.WriteString(strings.Join(columns, " | "))
		sb.WriteString("\n")
		sb.WriteString(strings.Repeat("-", len(columns)*20))
		sb.WriteString("\n")

		// 写入数据
		for _, row := range results {
			record := make([]string, len(columns))
			for i, col := range columns {
				if val, ok := row[col]; ok && val != nil {
					record[i] = fmt.Sprintf("%v", val)
				}
			}
			sb.WriteString(strings.Join(record, " | "))
			sb.WriteString("\n")
		}

		return sb.String(), nil

	default:
		return "", fmt.Errorf("不支持的输出格式：%s", outputFormat)
	}
}

// Close 关闭数据库连接
func (t *DBQueryTool) Close() error {
	if t.db != nil {
		return t.db.Close()
	}
	return nil
}
