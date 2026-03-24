package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"goagent/internal/tools"
	"goagent/pkg/llm"
)

// NodeType 节点类型
type NodeType string

const (
	NodeTrigger    NodeType = "trigger"
	NodeTool       NodeType = "tool"
	NodeLLM        NodeType = "llm"
	NodeParallel   NodeType = "parallel"
	NodeJoin       NodeType = "join"
	NodeCondition  NodeType = "condition"
	NodeLoop       NodeType = "loop"
	NodeTerminator NodeType = "terminator"
)

// Node 工作流节点
type Node struct {
	ID       string                 `json:"id"`
	Type     NodeType               `json:"type"`
	Name     string                 `json:"name"`
	Next     string                 `json:"next,omitempty"`
	Tool     string                 `json:"tool,omitempty"`
	Prompt   string                 `json:"prompt,omitempty"`
	Params   map[string]interface{} `json:"params,omitempty"`
	Branches []Branch               `json:"branches,omitempty"`
}

// Branch 并行分支
type Branch struct {
	Nodes []Node `json:"nodes"`
}

// Workflow 工作流定义
type Workflow struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Nodes       []Node `json:"nodes"`
}

// Executor 工作流执行器
type Executor struct {
	workflow     *Workflow
	ctx          context.Context
	variables    map[string]interface{}
	llmClient    llm.Client
	toolRegistry *tools.ToolRegistry
}

// NewExecutor 创建工作流执行器
func NewExecutor(ctx context.Context, workflow *Workflow, llmClient llm.Client, toolRegistry *tools.ToolRegistry) *Executor {
	return &Executor{
		workflow:     workflow,
		ctx:          ctx,
		variables:    make(map[string]interface{}),
		llmClient:    llmClient,
		toolRegistry: toolRegistry,
	}
}

// Execute 执行工作流
func (e *Executor) Execute(startNodeID string) (map[string]interface{}, error) {
	// 找到起始节点
	startNode := e.findNode(startNodeID)
	if startNode == nil {
		return nil, nil
	}

	// 执行节点链
	currentNode := startNode
	for currentNode != nil {
		select {
		case <-e.ctx.Done():
			return nil, e.ctx.Err()
		default:
		}

		result, err := e.executeNode(currentNode)
		if err != nil {
			return nil, err
		}

		// 保存结果到变量
		if result != nil {
			e.variables[currentNode.ID] = result
		}

		// 移动到下一个节点
		if currentNode.Next != "" {
			currentNode = e.findNode(currentNode.Next)
		} else {
			currentNode = nil
		}
	}

	return e.variables, nil
}

func (e *Executor) findNode(id string) *Node {
	for i := range e.workflow.Nodes {
		if e.workflow.Nodes[i].ID == id {
			return &e.workflow.Nodes[i]
		}
	}
	return nil
}

func (e *Executor) executeNode(node *Node) (interface{}, error) {
	switch node.Type {
	case NodeTrigger:
		return e.executeTriggerNode(node)
	case NodeTool:
		return e.executeToolNode(node)
	case NodeLLM:
		return e.executeLLMNode(node)
	case NodeParallel:
		return e.executeParallelNode(node)
	case NodeJoin:
		return e.executeJoinNode(node)
	case NodeCondition:
		return e.executeConditionNode(node)
	case NodeTerminator:
		return nil, nil
	default:
		return nil, nil
	}
}

func (e *Executor) executeTriggerNode(node *Node) (interface{}, error) {
	return nil, nil
}

func (e *Executor) executeToolNode(node *Node) (interface{}, error) {
	if node.Tool == "" {
		return nil, fmt.Errorf("tool node %s 未指定工具名称", node.Name)
	}

	// 获取工具
	tool, ok := e.toolRegistry.Get(node.Tool)
	if !ok {
		return nil, fmt.Errorf("工具 %s 不存在", node.Tool)
	}

	// 准备参数（支持变量替换）
	params := e.replaceVariables(node.Params)

	// 执行工具
	result, err := tool.Execute(e.ctx, params)
	if err != nil {
		return nil, fmt.Errorf("执行工具 %s 失败：%w", node.Tool, err)
	}

	return result, nil
}

func (e *Executor) executeLLMNode(node *Node) (interface{}, error) {
	if node.Prompt == "" {
		return nil, fmt.Errorf("llm node %s 未指定 prompt", node.Name)
	}

	// 替换 prompt 中的变量
	prompt := e.replaceVariableString(node.Prompt)

	// 调用 LLM
	resp, err := e.llmClient.Chat(e.ctx, []llm.Message{
		{
			Role:    llm.RoleUser,
			Content: prompt,
		},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("调用 LLM 失败：%w", err)
	}

	return resp.Content, nil
}

func (e *Executor) executeParallelNode(node *Node) (interface{}, error) {
	if len(node.Branches) == 0 {
		return nil, fmt.Errorf("parallel node %s 没有分支", node.Name)
	}

	// 使用 WaitGroup 等待所有分支完成
	var wg sync.WaitGroup
	results := make(map[string]interface{}, len(node.Branches))
	errors := make([]error, 0)
	mu := sync.Mutex{}

	// 并行执行每个分支
	for i, branch := range node.Branches {
		wg.Add(1)
		go func(index int, branch Branch) {
			defer wg.Done()

			// 执行分支中的节点链
			branchResult := make(map[string]interface{})
			for _, n := range branch.Nodes {
				result, err := e.executeNode(&n)
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("分支 %d 节点 %s 执行失败：%w", index, n.Name, err))
					mu.Unlock()
					return
				}
				if result != nil {
					branchResult[n.ID] = result
				}
			}

			// 保存分支结果
			mu.Lock()
			results[fmt.Sprintf("branch_%d", index)] = branchResult
			mu.Unlock()
		}(i, branch)
	}

	// 等待所有分支完成
	wg.Wait()

	if len(errors) > 0 {
		// 返回第一个错误
		return nil, errors[0]
	}

	return results, nil
}

func (e *Executor) executeJoinNode(node *Node) (interface{}, error) {
	// 合并所有并行分支的结果
	merged := make(map[string]interface{})

	// 从变量中获取所有分支的结果
	for _, value := range e.variables {
		if branchResult, ok := value.(map[string]interface{}); ok {
			for k, v := range branchResult {
				merged[k] = v
			}
		}
	}

	// 存储合并后的结果
	e.variables["merged_data"] = merged

	return merged, nil
}

func (e *Executor) executeConditionNode(node *Node) (interface{}, error) {
	// 根据条件判断走哪个分支
	if len(node.Branches) == 0 {
		return nil, fmt.Errorf("condition node %s 没有分支", node.Name)
	}

	// TODO: 实现条件表达式求值
	// 暂时默认执行第一个分支
	if len(node.Branches) > 0 {
		branch := node.Branches[0]
		for _, n := range branch.Nodes {
			result, err := e.executeNode(&n)
			if err != nil {
				return nil, err
			}
			if result != nil {
				e.variables[n.ID] = result
			}
		}
	}

	return nil, nil
}

// replaceVariables 递归替换 map 中的变量
func (e *Executor) replaceVariables(params map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range params {
		result[k] = e.replaceValue(v)
	}
	return result
}

// replaceValue 替换值中的变量
func (e *Executor) replaceValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return e.replaceVariableString(v)
	case map[string]interface{}:
		return e.replaceVariables(v)
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = e.replaceValue(item)
		}
		return result
	default:
		return value
	}
}

// replaceVariableString 替换字符串中的变量
func (e *Executor) replaceVariableString(s string) string {
	// 支持 {{variable_name}} 格式的变量替换
	for key, value := range e.variables {
		placeholder := "{{" + key + "}}"
		s = strings.ReplaceAll(s, placeholder, fmt.Sprintf("%v", value))

		// 也支持 {{date}} {{timestamp}} 等特殊变量
		placeholder = "{{date}}"
		s = strings.ReplaceAll(s, placeholder, time.Now().Format("2006-01-02"))

		placeholder = "{{timestamp}}"
		s = strings.ReplaceAll(s, placeholder, time.Now().Format("20060102150405"))
	}
	return s
}
