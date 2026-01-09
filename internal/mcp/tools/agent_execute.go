package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/domain"
	"github.com/victorzhuk/go-ent/internal/skill"
)

type AgentExecuteInput struct {
	Path       string                 `json:"path"`
	Task       string                 `json:"task"`
	TaskType   string                 `json:"task_type,omitempty"`
	Files      []string               `json:"files,omitempty"`
	MaxBudget  int                    `json:"max_budget,omitempty"`
	ForceRole  string                 `json:"force_role,omitempty"`
	ForceModel string                 `json:"force_model,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

type AgentExecuteResponse struct {
	Role       string   `json:"role"`
	Model      string   `json:"model"`
	Skills     []string `json:"skills"`
	Reason     string   `json:"reason"`
	Complexity string   `json:"complexity"`
	Message    string   `json:"message"`
}

func registerAgentExecute(s *mcp.Server, registry *skill.Registry) {
	tool := &mcp.Tool{
		Name:        "agent_execute",
		Description: "Execute a task with automatic agent selection based on complexity",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Path to project directory",
				},
				"task": map[string]any{
					"type":        "string",
					"description": "Task description to execute",
				},
				"task_type": map[string]any{
					"type":        "string",
					"description": "Type of task: feature, bugfix, refactor, test, documentation, architecture",
				},
				"files": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "List of files involved in the task",
				},
				"max_budget": map[string]any{
					"type":        "integer",
					"description": "Maximum token budget for execution (0 = unlimited)",
				},
				"force_role": map[string]any{
					"type":        "string",
					"description": "Override automatic role selection (architect, senior, developer, ops, reviewer)",
				},
				"force_model": map[string]any{
					"type":        "string",
					"description": "Override automatic model selection (opus, sonnet, haiku)",
				},
				"context": map[string]any{
					"type":        "object",
					"description": "Additional context for task execution",
				},
			},
			"required": []string{"path", "task"},
		},
	}

	handler := makeAgentExecuteHandler(registry)
	mcp.AddTool(s, tool, handler)
}

func makeAgentExecuteHandler(registry *skill.Registry) func(context.Context, *mcp.CallToolRequest, AgentExecuteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AgentExecuteInput) (*mcp.CallToolResult, any, error) {
		if input.Path == "" {
			return nil, nil, fmt.Errorf("path is required")
		}
		if input.Task == "" {
			return nil, nil, fmt.Errorf("task is required")
		}

		taskType := agent.TaskTypeFeature
		if input.TaskType != "" {
			taskType = agent.TaskType(input.TaskType)
		}

		task := agent.Task{
			Description: input.Task,
			Type:        taskType,
			Files:       input.Files,
			Metadata:    input.Context,
		}

		maxBudget := 0
		if input.MaxBudget > 0 {
			maxBudget = input.MaxBudget
		}

		selector := agent.NewSelector(agent.Config{
			MaxBudget:  maxBudget,
			StrictMode: false,
		}, registry)

		result, err := selector.Select(ctx, task)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error selecting agent: %v", err),
				}},
			}, nil, nil
		}

		if input.ForceRole != "" {
			result.Role = domain.AgentRole(input.ForceRole)
			result.Reason = fmt.Sprintf("manually overridden to %s", input.ForceRole)
		}

		if input.ForceModel != "" {
			result.Model = input.ForceModel
		}

		complexity := "unknown"
		analyzer := agent.NewComplexity()
		complexityResult := analyzer.Analyze(task)
		complexity = complexityResult.Level.String()

		response := AgentExecuteResponse{
			Role:       result.Role.String(),
			Model:      result.Model,
			Skills:     result.Skills,
			Reason:     result.Reason,
			Complexity: complexity,
			Message:    fmt.Sprintf("Selected %s agent with %s model (complexity: %s)", result.Role, result.Model, complexity),
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
			}, nil, nil
		}

		msg := fmt.Sprintf("✅ Agent selected for execution\n\n```json\n%s\n```\n\n", string(data))
		msg += fmt.Sprintf("**Next Steps:**\n")
		msg += fmt.Sprintf("- The %s agent (%s) will handle this task\n", result.Role, result.Model)
		msg += fmt.Sprintf("- Complexity level: %s\n", complexity)
		msg += fmt.Sprintf("- Reason: %s\n", result.Reason)

		if len(result.Skills) > 0 {
			msg += fmt.Sprintf("\n**Available Skills:** %v\n", result.Skills)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}
}
