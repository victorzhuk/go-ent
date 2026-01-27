package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/domain"
	"github.com/victorzhuk/go-ent/internal/execution"
	"github.com/victorzhuk/go-ent/internal/skill"
)

// EngineExecuteInput defines the input for engine execution.
type EngineExecuteInput struct {
	Path           string                 `json:"path"`
	Task           string                 `json:"task"`
	TaskType       string                 `json:"task_type,omitempty"`
	Files          []string               `json:"files,omitempty"`
	Strategy       string                 `json:"strategy,omitempty"`
	ForceAgent     string                 `json:"force_agent,omitempty"`
	ForceModel     string                 `json:"force_model,omitempty"`
	ForceRuntime   string                 `json:"force_runtime,omitempty"`
	ForceSummarize bool                   `json:"force_summarize,omitempty"`
	MaxTokens      int                    `json:"max_tokens,omitempty"`
	MaxCost        float64                `json:"max_cost,omitempty"`
	Context        map[string]interface{} `json:"context,omitempty"`
}

// EngineExecuteResponse contains execution results.
type EngineExecuteResponse struct {
	ExecutionID string   `json:"execution_id"`
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
	AvailableRuntimes      []string `json:"available_runtimes"`
	AvailableStrategies    []string `json:"available_strategies"`
	PreferredRuntime       string   `json:"preferred_runtime"`
	CurrentExecutionID     string   `json:"current_execution_id,omitempty"`
	CurrentExecutionStatus string   `json:"current_execution_status,omitempty"`
	DailySpending          float64  `json:"daily_spending"`
	MonthlySpending        float64  `json:"monthly_spending"`
	IsMCPMode              bool     `json:"is_mcp_mode"`
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

// EngineListInput defines input for listing executions.
type EngineListInput struct {
	Status string `json:"status,omitempty"`
	Sort   string `json:"sort,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// ExecutionEntry represents an execution summary.
type ExecutionEntry struct {
	ExecutionID   string `json:"execution_id"`
	Status        string `json:"status"`
	Description   string `json:"description"`
	TaskType      string `json:"task_type"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	FileSizeBytes int64  `json:"file_size_bytes"`
}

// EngineListResponse contains list of executions.
type EngineListResponse struct {
	Executions []ExecutionEntry `json:"executions"`
	Total      int              `json:"total"`
	Offset     int              `json:"offset"`
	Limit      int              `json:"limit"`
}

// EngineFindInput defines input for finding an execution by ID.
type EngineFindInput struct {
	ExecutionID string `json:"execution_id"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

// ExecutionDetail represents detailed execution information.
type ExecutionDetail struct {
	ExecutionID   string                 `json:"execution_id"`
	Status        string                 `json:"status"`
	Description   string                 `json:"description"`
	TaskType      string                 `json:"task_type"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
	StartedAt     string                 `json:"started_at,omitempty"`
	CompletedAt   string                 `json:"completed_at,omitempty"`
	InterruptedAt string                 `json:"interrupted_at,omitempty"`
	Result        *ExecutionResultDetail `json:"result,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionResultDetail represents execution result information.
type ExecutionResultDetail struct {
	Success     bool     `json:"success"`
	Output      string   `json:"output,omitempty"`
	Error       string   `json:"error,omitempty"`
	TokensIn    int      `json:"tokens_in,omitempty"`
	TokensOut   int      `json:"tokens_out,omitempty"`
	Cost        float64  `json:"cost,omitempty"`
	Duration    int64    `json:"duration_ms,omitempty"`
	Adjustments []string `json:"adjustments,omitempty"`
}

// EngineFindResponse contains execution details.
type EngineFindResponse struct {
	Execution *ExecutionDetail `json:"execution"`
}

// EngineHistoryInput defines input for querying execution history.
type EngineHistoryInput struct {
	ExecutionID string `json:"execution_id"`
	ResultType  string `json:"result_type,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	IncludeConv bool   `json:"include_conv,omitempty"`
}

// TimelineEntry represents a step in the execution timeline.
type TimelineEntry struct {
	StepNumber int    `json:"step_number"`
	Timestamp  string `json:"timestamp"`
	Type       string `json:"type"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
	Duration   int64  `json:"duration_ms"`
	Success    bool   `json:"success"`
}

// ConversationMessage represents a message in the conversation history.
type ConversationMessage struct {
	Role      string `json:"role"`
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
}

// EngineHistoryResponse contains execution history.
type EngineHistoryResponse struct {
	ExecutionID  string                `json:"execution_id"`
	Status       string                `json:"status"`
	Description  string                `json:"description"`
	TaskType     string                `json:"task_type"`
	CreatedAt    string                `json:"created_at"`
	UpdatedAt    string                `json:"updated_at"`
	StartedAt    string                `json:"started_at,omitempty"`
	CompletedAt  string                `json:"completed_at,omitempty"`
	Duration     int64                 `json:"duration_ms,omitempty"`
	Timeline     []TimelineEntry       `json:"timeline"`
	Conversation []ConversationMessage `json:"conversation,omitempty"`
	TokensIn     int                   `json:"tokens_in,omitempty"`
	TokensOut    int                   `json:"tokens_out,omitempty"`
	Cost         float64               `json:"cost,omitempty"`
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
				"force_summarize": map[string]any{
					"type":        "boolean",
					"description": "Force summarization regardless of thresholds",
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

	handler := WithMetrics[EngineExecuteInput, any]("engine_execute", func(ctx context.Context, req *mcp.CallToolRequest, input EngineExecuteInput) (*mcp.CallToolResult, any, error) {
		if input.Path == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "❌ Error: Path is required"}},
			}, nil, fmt.Errorf("path is required")
		}
		if input.Task == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "❌ Error: Task description is required"}},
			}, nil, fmt.Errorf("task is required")
		}

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := execution.New(execution.Config{
			IsMCPMode:            true,
			EnableSummarization:  true,
			EnableAutoCheckpoint: true,
		}, selector)

		task := &execution.Task{
			Description: input.Task,
			Type:        input.TaskType,
			Context: &execution.TaskContext{
				ProjectPath: input.Path,
				Files:       input.Files,
			},
		}

		if input.ForceAgent != "" {
			task.ForceAgent = domain.AgentRole(input.ForceAgent)
		}
		if input.ForceModel != "" {
			task.ForceModel = input.ForceModel
		}
		if input.ForceRuntime != "" {
			task.ForceRuntime = domain.Runtime(input.ForceRuntime)
		}
		if input.Strategy != "" {
			task.ForceStrategy = domain.ExecutionStrategy(input.Strategy)
		}
		if input.MaxTokens > 0 || input.MaxCost > 0 {
			task.Budget = &execution.BudgetLimit{
				MaxTokens: input.MaxTokens,
				MaxCost:   input.MaxCost,
			}
		}

		result, err := engine.Execute(ctx, task)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("❌ Execution failed: %v", err),
				}},
			}, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("✅ Execution completed\n\n**Success:** %v\n**Output:** %s\n**Tokens:** %d in, %d out\n**Cost:** $%.4f",
					result.Success,
					result.Output,
					result.TokensIn, result.TokensOut,
					result.Cost,
				),
			}},
		}, nil, nil
	})
	mcp.AddTool(s, tool, handler)
}

func registerEngineSummarize(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "engine_summarize",
		Description: "Manually trigger context summarization for a running or completed execution. Forces summarization regardless of thresholds.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"execution_id": map[string]any{
					"type":        "string",
					"description": "ID of execution to summarize context for",
				},
				"force": map[string]any{
					"type":        "boolean",
					"description": "Force summarization even if thresholds are not met (default: true for manual trigger)",
				},
			},
			"required": []string{"execution_id"},
		},
	}

	handler := func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		ExecutionID string `json:"execution_id"`
		Force       bool   `json:"force"`
	}) (*mcp.CallToolResult, any, error) {
		if input.ExecutionID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "❌ Error: Execution ID is required\n\nUse `engine_list` to see available executions.",
				}},
			}, nil, fmt.Errorf("execution_id is required")
		}

		state, err := execution.LoadState(input.ExecutionID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("❌ Execution not found: %s\n\nUse `engine_list` to see available executions.", input.ExecutionID),
				}},
			}, nil, fmt.Errorf("execution %s not found: %w", input.ExecutionID, err)
		}

		if state.Context == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "❌ Error: Execution has no context to summarize",
				}},
			}, nil, fmt.Errorf("execution %s has no context", input.ExecutionID)
		}

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := execution.New(execution.Config{
			IsMCPMode:           true,
			EnableSummarization: true,
		}, selector)

		task := &execution.Task{
			Context: state.Context,
		}

		summarized := engine.TriggerSummarization(ctx, task, input.ExecutionID, "claude-3-5-sonnet", state, input.Force)

		if !summarized {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("ℹ️  Summarization not performed for execution %s. Context was already within thresholds and force=false.", input.ExecutionID),
				}},
			}, nil, nil
		}

		summaryType := "automatic"
		if input.Force {
			summaryType = "manual"
		}

		msg := fmt.Sprintf("✅ %s summarization completed for execution `%s`\n\n**Files before:** %s\n**Files after:** %s\n**Summarization count:** %d\n\nUse `engine_find` with `include_full=true` to see updated context.",
			summaryType, input.ExecutionID,
			state.Metadata["summarization_original_files"],
			state.Metadata["summarization_files_after"],
			task.Context.SummarizationCount,
		)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, nil, nil
	}

	mcp.AddTool(s, tool, handler)
}

func registerEngineInterrupt(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "engine_interrupt",
		Description: "Interrupt a running execution by ID. Gracefully stops execution and saves state before stopping.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"execution_id": map[string]any{
					"type":        "string",
					"description": "ID of execution to interrupt",
				},
			},
			"required": []string{"execution_id"},
		},
	}

	handler := WithMetrics[EngineInterruptInput, EngineInterruptResponse]("engine_interrupt", makeEngineInterruptHandler())
	mcp.AddTool(s, tool, handler)
}

func registerEngineResume(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "engine_resume",
		Description: "Resume a previously interrupted, failed, or pending execution by ID. Loads saved state and continues execution from where it left off.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"execution_id": map[string]any{
					"type":        "string",
					"description": "ID of execution to resume",
				},
			},
			"required": []string{"execution_id"},
		},
	}

	handler := WithMetrics[EngineResumeInput, EngineResumeResponse]("engine_resume", makeEngineResumeHandler())
	mcp.AddTool(s, tool, handler)
}

type EngineInterruptInput struct {
	ExecutionID string `json:"execution_id"`
}

type EngineInterruptResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ExecutionID string `json:"execution_id"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}

func makeEngineInterruptHandler() func(context.Context, *mcp.CallToolRequest, EngineInterruptInput) (*mcp.CallToolResult, EngineInterruptResponse, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input EngineInterruptInput) (*mcp.CallToolResult, EngineInterruptResponse, error) {
		if input.ExecutionID == "" {
			return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{
						Text: "❌ Error: Execution ID is required\n\nUse `engine_list` to see available executions.",
					}},
				}, EngineInterruptResponse{
					Success: false,
					Error:   "execution_id is required",
				}, fmt.Errorf("execution_id is required")
		}

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := execution.New(execution.Config{
			IsMCPMode:            true,
			EnableAutoCheckpoint: true,
		}, selector)

		if err := engine.Interrupt(ctx, input.ExecutionID); err != nil {
			return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("❌ Failed to interrupt execution `%s`: %v\n\nUse `engine_find` to check execution status.", input.ExecutionID, err),
					}},
				}, EngineInterruptResponse{
					Success:     false,
					ExecutionID: input.ExecutionID,
					Error:       err.Error(),
				}, err
		}

		state, loadErr := execution.LoadState(input.ExecutionID)
		if loadErr != nil {
			return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("⚠️  Execution `%s` interrupted but failed to load state: %v", input.ExecutionID, loadErr),
					}},
				}, EngineInterruptResponse{
					Success:     true,
					ExecutionID: input.ExecutionID,
					Status:      "interrupted",
					Message:     "interrupted (state load error)",
				}, nil
		}

		return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("✅ Execution `%s` interrupted successfully\n\n**Status:** %s\n**Updated at:** %s\n\nThe execution has been gracefully stopped and its state has been saved. You can resume it later using `engine_resume`.",
						input.ExecutionID,
						state.Status,
						state.UpdatedAt.Format("2006-01-02 15:04:05"),
					),
				}},
			}, EngineInterruptResponse{
				Success:     true,
				ExecutionID: input.ExecutionID,
				Status:      state.Status,
				Message:     "interrupted successfully",
			}, nil
	}
}

type EngineResumeInput struct {
	ExecutionID string `json:"execution_id"`
}

type EngineResumeResponse struct {
	Success     bool    `json:"success"`
	Message     string  `json:"message"`
	ExecutionID string  `json:"execution_id"`
	Status      string  `json:"status"`
	Output      string  `json:"output,omitempty"`
	TokensIn    int     `json:"tokens_in,omitempty"`
	TokensOut   int     `json:"tokens_out,omitempty"`
	Cost        float64 `json:"cost,omitempty"`
	Error       string  `json:"error,omitempty"`
}

func makeEngineResumeHandler() func(context.Context, *mcp.CallToolRequest, EngineResumeInput) (*mcp.CallToolResult, EngineResumeResponse, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input EngineResumeInput) (*mcp.CallToolResult, EngineResumeResponse, error) {
		if input.ExecutionID == "" {
			return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{
						Text: "❌ Error: Execution ID is required\n\nUse `engine_list` to see available executions.",
					}},
				}, EngineResumeResponse{
					Success: false,
					Error:   "execution_id is required",
				}, fmt.Errorf("execution_id is required")
		}

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := execution.New(execution.Config{
			IsMCPMode:            true,
			EnableSummarization:  true,
			EnableAutoCheckpoint: true,
		}, selector)

		result, err := engine.ResumeExecution(ctx, input.ExecutionID)
		if err != nil {
			return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("❌ Failed to resume execution `%s`: %v\n\nUse `engine_find` to check execution status.", input.ExecutionID, err),
					}},
				}, EngineResumeResponse{
					Success:     false,
					ExecutionID: input.ExecutionID,
					Error:       err.Error(),
				}, err
		}

		return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("✅ Execution `%s` resumed successfully\n\n**Success:** %v\n**Output:** %s\n**Tokens:** %d in, %d out\n**Cost:** $%.4f",
						input.ExecutionID,
						result.Success,
						result.Output,
						result.TokensIn, result.TokensOut,
						result.Cost,
					),
				}},
			}, EngineResumeResponse{
				Success:     true,
				ExecutionID: input.ExecutionID,
				Status:      "completed",
				Output:      result.Output,
				TokensIn:    result.TokensIn,
				TokensOut:   result.TokensOut,
				Cost:        result.Cost,
				Message:     "resumed successfully",
			}, nil
	}
}
