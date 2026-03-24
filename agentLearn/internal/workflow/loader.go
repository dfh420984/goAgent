package workflow

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadWorkflow 从 YAML 文件加载工作流定义
func LoadWorkflow(filename string) (*Workflow, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取工作流文件失败：%w", err)
	}

	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("解析工作流 YAML 失败：%w", err)
	}

	// 验证工作流
	if err := validateWorkflow(&workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// validateWorkflow 验证工作流定义
func validateWorkflow(workflow *Workflow) error {
	if workflow.Name == "" {
		return fmt.Errorf("工作流名称不能为空")
	}

	if len(workflow.Nodes) == 0 {
		return fmt.Errorf("工作流必须至少有一个节点")
	}

	// 检查节点 ID 是否唯一
	nodeIDs := make(map[string]bool)
	for _, node := range workflow.Nodes {
		if node.ID == "" {
			return fmt.Errorf("节点 ID 不能为空")
		}
		if nodeIDs[node.ID] {
			return fmt.Errorf("节点 ID %s 重复", node.ID)
		}
		nodeIDs[node.ID] = true
	}

	// 检查 next 引用是否有效
	for _, node := range workflow.Nodes {
		if node.Next != "" {
			if !nodeIDs[node.Next] {
				return fmt.Errorf("节点 %s 的 next 引用了不存在的节点 %s", node.ID, node.Next)
			}
		}
	}

	return nil
}
