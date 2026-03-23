package workflow

import (
	"context"
)

// NodeType 节点类型
type NodeType string

const (
	NodeTrigger   NodeType = "trigger"
	NodeTool      NodeType = "tool"
	NodeLLM       NodeType = "llm"
	NodeParallel  NodeType = "parallel"
	NodeJoin      NodeType = "join"
	NodeCondition NodeType = "condition"
	NodeLoop      NodeType = "loop"
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
	workflow *Workflow
	ctx      context.Context
	variables map[string]interface{}
}

// NewExecutor 创建工作流执行器
func NewExecutor(ctx context.Context, workflow *Workflow) *Executor {
	return &Executor{
		workflow:  workflow,
		ctx:       ctx,
		variables: make(map[string]interface{}),
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
	// TODO: 需要调用工具注册表执行工具
	return nil, nil
}

func (e *Executor) executeLLMNode(node *Node) (interface{}, error) {
	// TODO: 需要调用 LLM 客户端
	return nil, nil
}

func (e *Executor) executeParallelNode(node *Node) (interface{}, error) {
	// TODO: 并行执行多个分支
	return nil, nil
}

func (e *Executor) executeJoinNode(node *Node) (interface{}, error) {
	// TODO: 合并并行分支的结果
	return nil, nil
}

func (e *Executor) executeConditionNode(node *Node) (interface{}, error) {
	// TODO: 条件判断
	return nil, nil
}
