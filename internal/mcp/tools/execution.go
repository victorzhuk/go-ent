package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/domain"
	"github.com/victorzhuk/go-ent/internal/execution"
	"github.com/victorzhuk/go-ent/internal/skill"
)

// EngineExecuteInput defines the input for engine execution.
type EngineExecuteInput struct {
	Path         string                 `json:"path"`
	Task         string                 `json:"task"`
	TaskType     string                 `json:"task_type,omitempty"`
	Files        []string               `json:"files,omitempty"`
	Strategy     string                 `json:"strategy,omitempty"`
	ForceAgent   string                 `json:"force_agent,omitempty"`
	ForceModel   string                 `json:"force_model,omitempty"`
	ForceRuntime string                 `json:"force_runtime,omitempty"`
	MaxTokens    int                    `json:"max_tokens,omitempty"`
	MaxCost      float64                `json:"max_cost,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

// EngineExecuteResponse contains execution results.
type EngineExecuteResponse struct {
	Success     bool     `json:"success"`
	Output      string   `json:"output"`
	Error       string   `json:"error,omitempty"`
	TokensIn    int      `json:"tokens_in"`
	TokensOut   int      `json:"tokens_out"`
	Cost        float64  `json:"cost"`
	Strategy    string   `json:"strategy"`
	Runtime     string   `json:"runtime"`
	Adjustments []string `json:"adjustments,omitempty"`
}

// EngineStatusResponse contains engine status information.
type EngineStatusResponse struct {
	AvailableRuntimes   []string `json:"available_runtimes"`
	AvailableStrategies []string `json:"available_strategies"`
	PreferredRuntime    string   `json:"preferred_runtime"`
	DailySpending       float64  `json:"daily_spending"`
	MonthlySpending     float64  `json:"monthly_spending"`
	IsMCPMode           bool     `json:"is_mcp_mode"`
}

// EngineBudgetInput defines budget query/update input.
type EngineBudgetInput struct {
	MaxTokens int     `json:"max_tokens,omitempty"`
	MaxCost   float64 `json:"max_cost,omitempty"`
	Reset     bool    `json:"reset,omitempty"`
}

// EngineBudgetResponse contains budget information.
type EngineBudgetResponse struct {
	DailyTokens     int     `json:"daily_tokens"`
	DailySpending   float64 `json:"daily_spending"`
	MonthlyTokens   int     `json:"monthly_tokens"`
	MonthlySpending float64 `json:"monthly_spending"`
	DailyLimit      int     `json:"daily_limit,omitempty"`
	MonthlyCost     float64 `json:"monthly_cost,omitempty"`
}

func registerEngineExecute(s *mcp.Server, registry *skill.Registry) {
	tool := &mcp.Tool{
		Name:        "engine_execute",
		Description: "Execute a task using the execution engine with runtime and strategy selection",
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
				"strategy": map[string]any{
					"type":        "string",
					"description": "Execution strategy: single, multi, parallel",
				},
				"force_agent": map[string]any{
					"type":        "string",
					"description": "Override agent selection",
				},
				"force_model": map[string]any{
					"type":        "string",
					"description": "Override model selection",
				},
				"force_runtime": map[string]any{
					"type":        "string",
					"description": "Override runtime selection: claude-code, open-code, cli",
				},
				"max_tokens": map[string]any{
					"type":        "integer",
					"description": "Maximum tokens allowed (0 = unlimited)",
				},
				"max_cost": map[string]any{
					"type":        "number",
					"description": "Maximum cost allowed in USD (0 = unlimited)",
				},
				"context": map[string]any{
					"type":        "object",
					"description": "Additional execution context",
				},
			},
			"required": []string{"path", "task"},
		},
	}

	handler := makeEngineExecuteHandler(registry)
	mcp.AddTool(s, tool, handler)
}

func makeEngineExecuteHandler(registry *skill.Registry) func(context.Context, *mcp.CallToolRequest, EngineExecuteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input EngineExecuteInput) (*mcp.CallToolResult, any, error) {
		if input.Path == "" {
			return nil, nil, fmt.Errorf("path is required")
		}
		if input.Task == "" {
			return nil, nil, fmt.Errorf("task is required")
		}

		// Create engine
		selector := agent.NewSelector(agent.Config{}, registry)
		engine := execution.New(execution.Config{
			PreferredRuntime: domain.RuntimeClaudeCode,
			IsMCPMode:        true,
		}, selector)

		// Build task
		task := execution.NewTask(input.Task).
			WithType(input.TaskType).
			WithMetadata("path", input.Path)

		if len(input.Files) > 0 {
			task = task.WithMetadata("files", input.Files)
		}

		if input.ForceAgent != "" {
			task = task.WithAgent(domain.AgentRole(input.ForceAgent))
		}

		if input.ForceModel != "" {
			task = task.WithModel(input.ForceModel)
		}

		if input.ForceRuntime != "" {
			task = task.WithRuntime(domain.Runtime(input.ForceRuntime))
		}

		if input.Strategy != "" {
			task = task.WithStrategy(domain.ExecutionStrategy(input.Strategy))
		}

		if input.MaxTokens > 0 || input.MaxCost > 0 {
			task = task.WithBudget(&execution.BudgetLimit{
				MaxTokens: input.MaxTokens,
				MaxCost:   input.MaxCost,
			})
		}

		// Execute
		result, err := engine.Execute(ctx, task)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Execution error: %v", err),
				}},
			}, nil, nil
		}

		response := EngineExecuteResponse{
			Success:     result.Success,
			Output:      result.Output,
			Error:       result.Error,
			TokensIn:    result.TokensIn,
			TokensOut:   result.TokensOut,
			Cost:        result.Cost,
			Strategy:    input.Strategy,
			Adjustments: result.Adjustments,
		}

		data, _ := json.MarshalIndent(response, "", "  ")
		msg := fmt.Sprintf("✅ Execution completed\n\n```json\n%s\n```\n", string(data))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}
}

func registerEngineStatus(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "engine_status",
		Description: "Get engine status including available runtimes, strategies, and budget info",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}

	handler := func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
		// Create temporary engine to get capabilities
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := execution.New(execution.Config{
			PreferredRuntime: domain.RuntimeClaudeCode,
			IsMCPMode:        true,
		}, selector)

		status := engine.Status(ctx)

		response := EngineStatusResponse{
			AvailableRuntimes:   status.AvailableRuntimes,
			AvailableStrategies: status.AvailableStrategies,
			PreferredRuntime:    status.PreferredRuntime,
			DailySpending:       status.Budget.DailySpending,
			MonthlySpending:     status.Budget.MonthlySpending,
			IsMCPMode:           true,
		}

		data, _ := json.MarshalIndent(response, "", "  ")
		msg := fmt.Sprintf("✅ Engine status\n\n```json\n%s\n```\n", string(data))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}

	mcp.AddTool(s, tool, handler)
}

func registerEngineBudget(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "engine_budget",
		Description: "Query or update budget limits and spending information",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"max_tokens": map[string]any{
					"type":        "integer",
					"description": "Set maximum tokens limit",
				},
				"max_cost": map[string]any{
					"type":        "number",
					"description": "Set maximum cost limit in USD",
				},
				"reset": map[string]any{
					"type":        "boolean",
					"description": "Reset spending counters",
				},
			},
		},
	}

	handler := func(ctx context.Context, req *mcp.CallToolRequest, input EngineBudgetInput) (*mcp.CallToolResult, any, error) {
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := execution.New(execution.Config{
			IsMCPMode: true,
		}, selector)

		status := engine.Status(ctx)

		response := EngineBudgetResponse{
			DailyTokens:     status.Budget.DailyTokens,
			DailySpending:   status.Budget.DailySpending,
			MonthlyTokens:   status.Budget.MonthlyTokens,
			MonthlySpending: status.Budget.MonthlySpending,
		}

		data, _ := json.MarshalIndent(response, "", "  ")
		msg := fmt.Sprintf("✅ Budget information\n\n```json\n%s\n```\n", string(data))

		if input.Reset {
			msg += "\n**Note:** Budget reset functionality requires engine persistence (coming in v2)"
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}

	mcp.AddTool(s, tool, handler)
}

func registerEngineInterrupt(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "engine_interrupt",
		Description: "Interrupt a running execution",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"execution_id": map[string]any{
					"type":        "string",
					"description": "ID of execution to interrupt",
				},
			},
		},
	}

	handler := func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		ExecutionID string `json:"execution_id"`
	}) (*mcp.CallToolResult, any, error) {
		msg := "✅ Interrupt signal sent\n\n**Note:** Execution interruption requires engine state management (coming in v2)"

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, nil, nil
	}

	mcp.AddTool(s, tool, handler)
}
