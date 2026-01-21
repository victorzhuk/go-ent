package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/execution"
	"github.com/victorzhuk/go-ent/internal/router"
	"github.com/victorzhuk/go-ent/internal/worker"
)

type WorkerSpawnInput struct {
	Provider string   `json:"provider"`
	Task     string   `json:"task"`
	Method   string   `json:"method,omitempty"`
	Files    []string `json:"files,omitempty"`
	Timeout  int      `json:"timeout,omitempty"`
}

type WorkerSpawnResponse struct {
	WorkerID string `json:"worker_id"`
	Status   string `json:"status"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Message  string `json:"message"`
}

type WorkerPromptInput struct {
	WorkerID     string   `json:"worker_id"`
	Prompt       string   `json:"prompt"`
	ContextFiles []string `json:"context_files,omitempty"`
	Tools        []string `json:"tools,omitempty"`
	Stream       bool     `json:"stream,omitempty"`
}

type WorkerPromptResponse struct {
	PromptID string `json:"prompt_id"`
	Status   string `json:"status"`
	Result   string `json:"result,omitempty"`
}

type WorkerStatusInput struct {
	WorkerID string `json:"worker_id"`
}

type WorkerStatusResponse struct {
	WorkerID         string `json:"worker_id"`
	Status           string `json:"status"`
	Provider         string `json:"provider"`
	Model            string `json:"model"`
	Method           string `json:"method"`
	Task             string `json:"task,omitempty"`
	PromptID         string `json:"prompt_id,omitempty"`
	StartedAt        string `json:"started_at"`
	LastOutputAt     string `json:"last_output_at,omitempty"`
	Health           string `json:"health"`
	HealthCheckCount int    `json:"health_check_count,omitempty"`
	RetryCount       int    `json:"retry_count,omitempty"`
}

type WorkerOutputInput struct {
	WorkerID string `json:"worker_id"`
	Since    string `json:"since,omitempty"`
	Filter   string `json:"filter,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

type WorkerOutputResponse struct {
	WorkerID    string `json:"worker_id"`
	Output      string `json:"output"`
	LineCount   int    `json:"line_count"`
	LastUpdated string `json:"last_updated"`
	Truncated   bool   `json:"truncated"`
}

type WorkerCancelInput struct {
	WorkerID string `json:"worker_id"`
	Reason   string `json:"reason,omitempty"`
}

type WorkerCancelResponse struct {
	WorkerID    string `json:"worker_id"`
	Status      string `json:"status"`
	CancelledAt string `json:"cancelled_at"`
	Reason      string `json:"reason,omitempty"`
}

type WorkerListInput struct {
	Status   string `json:"status,omitempty"`
	Provider string `json:"provider,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

type WorkerInfo struct {
	WorkerID  string `json:"worker_id"`
	Status    string `json:"status"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	Task      string `json:"task,omitempty"`
	StartedAt string `json:"started_at"`
	Health    string `json:"health"`
}

type WorkerListResponse struct {
	Workers  []WorkerInfo `json:"workers"`
	Total    int          `json:"total"`
	Filtered int          `json:"filtered"`
}

type ProviderListInput struct {
	ActiveOnly bool   `json:"active_only,omitempty"`
	Method     string `json:"method,omitempty"`
}

type ProviderInfo struct {
	ID           string   `json:"id"`
	Method       string   `json:"method"`
	ProviderName string   `json:"provider_name"`
	Model        string   `json:"model"`
	BestFor      []string `json:"best_for,omitempty"`
	Cost         string   `json:"cost,omitempty"`
	ContextLimit int      `json:"context_limit,omitempty"`
	Health       string   `json:"health"`
}

type ProviderListResponse struct {
	Providers []ProviderInfo `json:"providers"`
	Total     int            `json:"total"`
	Active    int            `json:"active"`
}

type ProviderRecommendInput struct {
	Task        string   `json:"task"`
	Files       []string `json:"files,omitempty"`
	ContextSize int      `json:"context_size,omitempty"`
	Complexity  string   `json:"complexity,omitempty"`
	Priority    string   `json:"priority,omitempty"`
}

type ProviderAlternative struct {
	Provider string  `json:"provider"`
	Model    string  `json:"model"`
	Cost     float64 `json:"cost"`
	Reason   string  `json:"reason"`
}

type ProviderRecommendResponse struct {
	RecommendedProvider string                 `json:"recommended_provider"`
	RecommendedModel    string                 `json:"recommended_model"`
	RecommendedMethod   string                 `json:"recommended_method"`
	Reason              string                 `json:"reason"`
	EstimatedCost       float64                `json:"estimated_cost"`
	Alternatives        []ProviderAlternative  `json:"alternatives"`
	TaskCharacteristics map[string]interface{} `json:"task_characteristics"`
}

func registerWorkerSpawn(s *mcp.Server, manager *worker.WorkerManager, providerConfig *config.ProvidersConfig) {
	tool := &mcp.Tool{
		Name:        "worker_spawn",
		Description: "Spawn an OpenCode worker with specified provider",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"provider": map[string]any{
					"type":        "string",
					"enum":        []any{"glm", "kimi", "deepseek", "haiku"},
					"description": "OpenCode provider to use",
				},
				"task": map[string]any{
					"type":        "string",
					"description": "Task description or prompt for the worker",
				},
				"method": map[string]any{
					"type":        "string",
					"enum":        []any{"acp", "cli", "api"},
					"description": "Communication method (defaults from router)",
				},
				"files": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "List of files to include in context",
				},
				"timeout": map[string]any{
					"type":        "number",
					"description": "Worker timeout in seconds",
				},
			},
			"required": []string{"provider", "task"},
		},
	}

	baseHandler := makeWorkerSpawnHandler(manager, providerConfig)
	handler := WithMetrics[WorkerSpawnInput, any]("worker_spawn", baseHandler)
	mcp.AddTool(s, tool, handler)
}

func registerWorkerPrompt(s *mcp.Server, manager *worker.WorkerManager) {
	tool := &mcp.Tool{
		Name:        "worker_prompt",
		Description: "Send task to ACP worker",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"worker_id": map[string]any{
					"type":        "string",
					"description": "Worker ID to send prompt to",
				},
				"prompt": map[string]any{
					"type":        "string",
					"description": "Task description or prompt text",
				},
				"context_files": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Files to include in context",
				},
				"tools": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Tools available to worker",
				},
				"stream": map[string]any{
					"type":        "boolean",
					"description": "Return streaming updates",
				},
			},
			"required": []string{"worker_id", "prompt"},
		},
	}

	baseHandler := makeWorkerPromptHandler(manager)
	handler := WithMetrics[WorkerPromptInput, any]("worker_prompt", baseHandler)
	mcp.AddTool(s, tool, handler)
}

func registerWorkerStatus(s *mcp.Server, manager *worker.WorkerManager) {
	tool := &mcp.Tool{
		Name:        "worker_status",
		Description: "Check worker status and health",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"worker_id": map[string]any{
					"type":        "string",
					"description": "Worker ID to check status",
				},
			},
			"required": []string{"worker_id"},
		},
	}

	baseHandler := makeWorkerStatusHandler(manager)
	handler := WithMetrics[WorkerStatusInput, any]("worker_status", baseHandler)
	mcp.AddTool(s, tool, handler)
}

func registerWorkerOutput(s *mcp.Server, manager *worker.WorkerManager) {
	tool := &mcp.Tool{
		Name:        "worker_output",
		Description: "Retrieve worker output (optionally filtered)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"worker_id": map[string]any{
					"type":        "string",
					"description": "Worker to get output from",
				},
				"since": map[string]any{
					"type":        "string",
					"format":      "date-time",
					"description": "Get output since timestamp (RFC3339)",
				},
				"filter": map[string]any{
					"type":        "string",
					"description": "Regex pattern to filter lines",
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "Max number of lines to return",
				},
			},
			"required": []string{"worker_id"},
		},
	}

	baseHandler := makeWorkerOutputHandler(manager)
	handler := WithMetrics[WorkerOutputInput, any]("worker_output", baseHandler)
	mcp.AddTool(s, tool, handler)
}

func registerWorkerCancel(s *mcp.Server, manager *worker.WorkerManager) {
	tool := &mcp.Tool{
		Name:        "worker_cancel",
		Description: "Cancel a running worker",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"worker_id": map[string]any{
					"type":        "string",
					"description": "Worker ID to cancel",
				},
				"reason": map[string]any{
					"type":        "string",
					"description": "Reason for cancellation (optional)",
				},
			},
			"required": []string{"worker_id"},
		},
	}

	baseHandler := makeWorkerCancelHandler(manager)
	handler := WithMetrics[WorkerCancelInput, any]("worker_cancel", baseHandler)
	mcp.AddTool(s, tool, handler)
}

func registerWorkerList(s *mcp.Server, manager *worker.WorkerManager) {
	tool := &mcp.Tool{
		Name:        "worker_list",
		Description: "List all workers (optionally filtered)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{
					"type":        "string",
					"enum":        []any{"idle", "running", "completed", "failed"},
					"description": "Filter by worker status",
				},
				"provider": map[string]any{
					"type":        "string",
					"description": "Filter by provider (e.g., glm, kimi, deepseek, haiku)",
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "Maximum number of workers to return",
				},
			},
		},
	}

	baseHandler := makeWorkerListHandler(manager)
	handler := WithMetrics[WorkerListInput, any]("worker_list", baseHandler)
	mcp.AddTool(s, tool, handler)
}

func registerProviderList(s *mcp.Server, providerConfig *config.ProvidersConfig) {
	tool := &mcp.Tool{
		Name:        "provider_list",
		Description: "List configured providers with capabilities",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"active_only": map[string]any{
					"type":        "boolean",
					"description": "Only show active/healthy providers",
				},
				"method": map[string]any{
					"type":        "string",
					"enum":        []any{"acp", "cli", "api"},
					"description": "Filter by communication method",
				},
			},
		},
	}

	baseHandler := makeProviderListHandler(providerConfig)
	handler := WithMetrics[ProviderListInput, any]("provider_list", baseHandler)
	mcp.AddTool(s, tool, handler)
}

func makeWorkerPromptHandler(manager *worker.WorkerManager) func(context.Context, *mcp.CallToolRequest, WorkerPromptInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input WorkerPromptInput) (*mcp.CallToolResult, any, error) {
		if input.WorkerID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Error: worker_id is required",
				}},
			}, nil, fmt.Errorf("worker_id is required")
		}

		if input.Prompt == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Error: prompt is required",
				}},
			}, nil, fmt.Errorf("prompt is required")
		}

		promptReq := worker.PromptRequest{
			WorkerID:     input.WorkerID,
			Prompt:       input.Prompt,
			ContextFiles: input.ContextFiles,
			Tools:        input.Tools,
		}

		response, err := manager.SendPrompt(ctx, promptReq)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Failed to send prompt: %v", err),
				}},
			}, nil, fmt.Errorf("send prompt: %w", err)
		}

		result := WorkerPromptResponse{
			PromptID: response.PromptID,
			Status:   response.Status,
		}

		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
			}, nil, fmt.Errorf("marshal response: %w", err)
		}

		msg := fmt.Sprintf("✅ Prompt Sent to Worker\n\n```json\n%s\n```\n\n", string(data))
		msg += "**Prompt Details:**\n"
		msg += fmt.Sprintf("- Prompt ID: `%s`\n", result.PromptID)
		msg += fmt.Sprintf("- Status: %s\n", result.Status)
		msg += fmt.Sprintf("- Worker ID: `%s`\n", input.WorkerID)
		if len(input.ContextFiles) > 0 {
			msg += fmt.Sprintf("- Context Files: %d\n", len(input.ContextFiles))
		}
		if len(input.Tools) > 0 {
			msg += fmt.Sprintf("- Tools: %d\n", len(input.Tools))
		}
		msg += "\nUse `worker_status` to monitor progress.\n"

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, result, nil
	}
}

func makeWorkerStatusHandler(manager *worker.WorkerManager) func(context.Context, *mcp.CallToolRequest, WorkerStatusInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input WorkerStatusInput) (*mcp.CallToolResult, any, error) {
		if input.WorkerID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Error: worker_id is required",
				}},
			}, nil, fmt.Errorf("worker_id is required")
		}

		w := manager.Get(input.WorkerID)
		if w == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error: worker %s not found", input.WorkerID),
				}},
			}, nil, fmt.Errorf("worker %s not found", input.WorkerID)
		}

		w.Mutex.Lock()
		defer w.Mutex.Unlock()

		taskDesc := ""
		if w.Task != nil {
			taskDesc = w.Task.Description
		}

		promptID := ""

		lastOutputAt := ""
		if !w.LastOutputTime.IsZero() {
			lastOutputAt = w.LastOutputTime.Format(time.RFC3339)
		}

		response := WorkerStatusResponse{
			WorkerID:         w.ID,
			Status:           w.Status.String(),
			Provider:         w.Provider,
			Model:            w.Model,
			Method:           string(w.Method),
			Task:             taskDesc,
			PromptID:         promptID,
			StartedAt:        w.StartedAt.Format(time.RFC3339),
			LastOutputAt:     lastOutputAt,
			Health:           w.Health.String(),
			HealthCheckCount: w.HealthCheckCount,
			RetryCount:       w.RetryCount,
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
			}, nil, fmt.Errorf("marshal response: %w", err)
		}

		msg := fmt.Sprintf("✅ Worker Status\n\n```json\n%s\n```\n\n", string(data))
		msg += "**Status Details:**\n"
		msg += fmt.Sprintf("- Worker ID: `%s`\n", response.WorkerID)
		msg += fmt.Sprintf("- Status: %s\n", response.Status)
		msg += fmt.Sprintf("- Health: %s\n", response.Health)
		msg += fmt.Sprintf("- Provider: %s\n", response.Provider)
		msg += fmt.Sprintf("- Model: %s\n", response.Model)
		if response.Method != "" {
			msg += fmt.Sprintf("- Method: %s\n", response.Method)
		}
		if response.Task != "" {
			msg += fmt.Sprintf("- Task: %s\n", response.Task)
		}
		if response.PromptID != "" {
			msg += fmt.Sprintf("- Active Prompt: `%s`\n", response.PromptID)
		}
		msg += fmt.Sprintf("- Started: %s\n", response.StartedAt)
		if response.LastOutputAt != "" {
			msg += fmt.Sprintf("- Last Output: %s\n", response.LastOutputAt)
		}
		if w.HealthCheckCount > 0 {
			msg += fmt.Sprintf("- Health Checks: %d\n", w.HealthCheckCount)
		}
		if w.RetryCount > 0 {
			msg += fmt.Sprintf("- Retries: %d\n", w.RetryCount)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}
}

func makeWorkerOutputHandler(manager *worker.WorkerManager) func(context.Context, *mcp.CallToolRequest, WorkerOutputInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input WorkerOutputInput) (*mcp.CallToolResult, any, error) {
		if input.WorkerID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Error: worker_id is required",
				}},
			}, nil, fmt.Errorf("worker_id is required")
		}

		outputReq := worker.WorkerOutputRequest{
			WorkerID: input.WorkerID,
			Filter:   input.Filter,
			Limit:    input.Limit,
		}

		if input.Since != "" {
			sinceTime, err := time.Parse(time.RFC3339, input.Since)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Error: invalid since timestamp format: %v", err),
					}},
				}, nil, fmt.Errorf("parse since timestamp: %w", err)
			}
			outputReq.Since = sinceTime
		}

		outputResp, err := manager.GetOutput(outputReq)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Failed to get worker output: %v", err),
				}},
			}, nil, fmt.Errorf("get output: %w", err)
		}

		response := WorkerOutputResponse{
			WorkerID:    outputResp.WorkerID,
			Output:      outputResp.Output,
			LineCount:   outputResp.LineCount,
			LastUpdated: outputResp.LastUpdated.Format(time.RFC3339),
			Truncated:   outputResp.Truncated,
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
			}, nil, fmt.Errorf("marshal response: %w", err)
		}

		msg := fmt.Sprintf("✅ Worker Output\n\n```json\n%s\n```\n\n", string(data))
		msg += "**Output Details:**\n"
		msg += fmt.Sprintf("- Worker ID: `%s`\n", response.WorkerID)
		msg += fmt.Sprintf("- Lines: %d\n", response.LineCount)
		if response.LastUpdated != "" {
			msg += fmt.Sprintf("- Last Updated: %s\n", response.LastUpdated)
		}
		if response.Truncated {
			msg += fmt.Sprintf("- **Truncated** (limit: %d)\n", input.Limit)
		}
		if input.Filter != "" {
			msg += fmt.Sprintf("- Filter: `%s`\n", input.Filter)
		}
		msg += "\n"

		if response.Output != "" {
			msg += "```text\n" + response.Output + "\n```\n"
		} else {
			msg += "*No output available*\n"
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}
}

func makeWorkerSpawnHandler(manager *worker.WorkerManager, providerConfig *config.ProvidersConfig) func(context.Context, *mcp.CallToolRequest, WorkerSpawnInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input WorkerSpawnInput) (*mcp.CallToolResult, any, error) {
		if input.Provider == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Error: provider is required",
				}},
			}, nil, fmt.Errorf("provider is required")
		}

		if input.Task == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Error: task is required",
				}},
			}, nil, fmt.Errorf("task is required")
		}

		validProviders := map[string]bool{
			"glm":      true,
			"kimi":     true,
			"deepseek": true,
			"haiku":    true,
		}

		if !validProviders[input.Provider] {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error: invalid provider '%s'. Must be one of: glm, kimi, deepseek, haiku", input.Provider),
				}},
			}, nil, fmt.Errorf("invalid provider: %s", input.Provider)
		}

		method := config.CommunicationMethod(input.Method)
		if input.Method != "" && !method.Valid() {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error: invalid method '%s'. Must be one of: acp, cli, api", input.Method),
				}},
			}, nil, fmt.Errorf("invalid method: %s", input.Method)
		}

		if method == "" && providerConfig != nil {
			if provider, exists := providerConfig.Providers[input.Provider]; exists {
				method = provider.Method
			}
		}

		if method == "" {
			method = config.MethodCLI
		}

		timeout := time.Duration(input.Timeout) * time.Second
		if timeout == 0 {
			if providerConfig != nil && providerConfig.Health != nil {
				timeout = time.Duration(providerConfig.Health.WorkerTimeout) * time.Second
			} else {
				timeout = 5 * time.Minute
			}
		}

		task := execution.NewTask(input.Task)
		if len(input.Files) > 0 {
			task.Context = &execution.TaskContext{
				Files: input.Files,
			}
		}

		model := ""
		if providerConfig != nil {
			if provider, exists := providerConfig.Providers[input.Provider]; exists {
				model = provider.Model
			}
		}
		if model == "" {
			model = input.Provider
		}

		openCodeConfigPath := ""
		if providerConfig != nil {
			openCodeConfigPath = providerConfig.OpenCodeConfigPath
		}

		spawnReq := worker.SpawnRequest{
			Provider:           input.Provider,
			Model:              model,
			Method:             method,
			Task:               task,
			Timeout:            timeout,
			OpenCodeConfigPath: openCodeConfigPath,
			Metadata: map[string]interface{}{
				"files": input.Files,
			},
		}

		workerID, err := manager.Spawn(ctx, spawnReq)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Failed to spawn worker: %v", err),
				}},
			}, nil, fmt.Errorf("spawn worker: %w", err)
		}

		status, _ := manager.GetStatus(workerID)

		response := WorkerSpawnResponse{
			WorkerID: workerID,
			Status:   status.String(),
			Provider: input.Provider,
			Model:    model,
			Message:  fmt.Sprintf("Worker %s spawned successfully with provider %s and model %s", workerID, input.Provider, model),
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
			}, nil, fmt.Errorf("marshal response: %w", err)
		}

		msg := fmt.Sprintf("✅ OpenCode Worker Spawned\n\n```json\n%s\n```\n\n", string(data))
		msg += "**Next Steps:**\n"
		msg += fmt.Sprintf("- Worker ID: `%s`\n", workerID)
		msg += fmt.Sprintf("- Status: %s\n", status)
		msg += fmt.Sprintf("- Provider: %s\n", input.Provider)
		msg += fmt.Sprintf("- Model: %s\n", model)
		msg += fmt.Sprintf("- Method: %s\n", method)
		msg += "\nUse `worker_status` or `worker_list` to monitor progress.\n"

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}
}

func makeWorkerCancelHandler(manager *worker.WorkerManager) func(context.Context, *mcp.CallToolRequest, WorkerCancelInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input WorkerCancelInput) (*mcp.CallToolResult, any, error) {
		if input.WorkerID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Error: worker_id is required",
				}},
			}, nil, fmt.Errorf("worker_id is required")
		}

		w := manager.Get(input.WorkerID)
		if w == nil {
			response := WorkerCancelResponse{
				WorkerID: input.WorkerID,
				Status:   "not_found",
			}

			data, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Error formatting response: %v", err),
					}},
				}, nil, fmt.Errorf("marshal response: %w", err)
			}

			msg := fmt.Sprintf("❌ Worker Not Found\n\n```json\n%s\n```", string(data))
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: msg}},
			}, response, nil
		}

		w.Mutex.Lock()
		previousStatus := w.Status
		w.Mutex.Unlock()

		if previousStatus == worker.StatusIdle ||
			previousStatus == worker.StatusCompleted ||
			previousStatus == worker.StatusFailed ||
			previousStatus == worker.StatusCancelled {

			response := WorkerCancelResponse{
				WorkerID:    input.WorkerID,
				Status:      "already_stopped",
				CancelledAt: time.Now().Format(time.RFC3339),
			}

			data, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Error formatting response: %v", err),
					}},
				}, nil, fmt.Errorf("marshal response: %w", err)
			}

			msg := fmt.Sprintf("ℹ️ Worker Already Stopped\n\n```json\n%s\n```\n\n", string(data))
			msg += fmt.Sprintf("- Worker ID: `%s`\n", response.WorkerID)
			msg += fmt.Sprintf("- Previous Status: %s\n", previousStatus)
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: msg}},
			}, response, nil
		}

		if err := w.Stop(); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Failed to cancel worker: %v", err),
				}},
			}, nil, fmt.Errorf("cancel worker: %w", err)
		}

		response := WorkerCancelResponse{
			WorkerID:    input.WorkerID,
			Status:      "cancelled",
			CancelledAt: time.Now().Format(time.RFC3339),
			Reason:      input.Reason,
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
			}, nil, fmt.Errorf("marshal response: %w", err)
		}

		msg := fmt.Sprintf("✅ Worker Cancelled\n\n```json\n%s\n```\n\n", string(data))
		msg += "**Cancellation Details:**\n"
		msg += fmt.Sprintf("- Worker ID: `%s`\n", response.WorkerID)
		msg += fmt.Sprintf("- Status: %s\n", response.Status)
		msg += fmt.Sprintf("- Cancelled At: %s\n", response.CancelledAt)
		if response.Reason != "" {
			msg += fmt.Sprintf("- Reason: %s\n", response.Reason)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}
}

func makeWorkerListHandler(manager *worker.WorkerManager) func(context.Context, *mcp.CallToolRequest, WorkerListInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input WorkerListInput) (*mcp.CallToolResult, any, error) {
		workers := manager.List()
		total := len(workers)

		var statusFilter worker.WorkerStatus
		if input.Status != "" {
			statusFilter = worker.WorkerStatus(input.Status)
			if !statusFilter.Valid() {
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Error: invalid status '%s'. Must be one of: idle, running, completed, failed", input.Status),
					}},
				}, nil, fmt.Errorf("invalid status: %s", input.Status)
			}
		}

		var result []WorkerInfo
		for _, w := range workers {
			w.Mutex.Lock()

			if input.Status != "" && w.Status != statusFilter {
				w.Mutex.Unlock()
				continue
			}

			if input.Provider != "" && w.Provider != input.Provider {
				w.Mutex.Unlock()
				continue
			}

			taskDesc := ""
			if w.Task != nil {
				taskDesc = w.Task.Description
			}

			workerInfo := WorkerInfo{
				WorkerID:  w.ID,
				Status:    w.Status.String(),
				Provider:  w.Provider,
				Model:     w.Model,
				Task:      taskDesc,
				StartedAt: w.StartedAt.Format(time.RFC3339),
				Health:    w.Health.String(),
			}
			result = append(result, workerInfo)
			w.Mutex.Unlock()
		}

		filtered := len(result)

		if input.Limit > 0 && len(result) > input.Limit {
			result = result[:input.Limit]
		}

		response := WorkerListResponse{
			Workers:  result,
			Total:    total,
			Filtered: filtered,
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
			}, nil, fmt.Errorf("marshal response: %w", err)
		}

		msg := fmt.Sprintf("✅ Worker List\n\n```json\n%s\n```\n\n", string(data))
		msg += "**Summary:**\n"
		msg += fmt.Sprintf("- Total Workers: %d\n", total)
		msg += fmt.Sprintf("- Filtered Workers: %d\n", filtered)
		if input.Limit > 0 && len(result) == input.Limit {
			msg += fmt.Sprintf("- Limited to: %d\n", input.Limit)
		}
		if input.Status != "" {
			msg += fmt.Sprintf("- Status Filter: %s\n", input.Status)
		}
		if input.Provider != "" {
			msg += fmt.Sprintf("- Provider Filter: %s\n", input.Provider)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}
}

func makeProviderListHandler(providerConfig *config.ProvidersConfig) func(context.Context, *mcp.CallToolRequest, ProviderListInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ProviderListInput) (*mcp.CallToolResult, any, error) {
		if providerConfig == nil || len(providerConfig.Providers) == 0 {
			response := ProviderListResponse{
				Providers: []ProviderInfo{},
				Total:     0,
				Active:    0,
			}

			data, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Error formatting response: %v", err),
					}},
				}, nil, fmt.Errorf("marshal response: %w", err)
			}

			msg := fmt.Sprintf("ℹ️ No Providers Configured\n\n```json\n%s\n```\n", string(data))
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: msg}},
			}, response, nil
		}

		var methodFilter config.CommunicationMethod
		if input.Method != "" {
			methodFilter = config.CommunicationMethod(input.Method)
			if !methodFilter.Valid() {
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Error: invalid method '%s'. Must be one of: acp, cli, api", input.Method),
					}},
				}, nil, fmt.Errorf("invalid method: %s", input.Method)
			}
		}

		var result []ProviderInfo
		total := len(providerConfig.Providers)

		for name, provider := range providerConfig.Providers {
			if input.Method != "" && provider.Method != methodFilter {
				continue
			}

			health := "unknown"
			if input.ActiveOnly && health != "healthy" {
				continue
			}

			cost := ""
			if provider.Cost != nil {
				if provider.Cost.Per1kTokens > 0 {
					cost = fmt.Sprintf("$%.4f/1k tokens", provider.Cost.Per1kTokens)
				}
				if provider.Cost.PerHour > 0 {
					if cost != "" {
						cost += ", "
					}
					cost += fmt.Sprintf("$%.2f/hour", provider.Cost.PerHour)
				}
			}

			providerInfo := ProviderInfo{
				ID:           name,
				Method:       provider.Method.String(),
				ProviderName: provider.Provider,
				Model:        provider.Model,
				BestFor:      provider.BestFor,
				Cost:         cost,
				ContextLimit: provider.ContextLimit,
				Health:       health,
			}

			result = append(result, providerInfo)
		}

		active := len(result)

		response := ProviderListResponse{
			Providers: result,
			Total:     total,
			Active:    active,
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
			}, nil, fmt.Errorf("marshal response: %w", err)
		}

		msg := fmt.Sprintf("✅ Providers List\n\n```json\n%s\n```\n\n", string(data))
		msg += "**Summary:**\n"
		msg += fmt.Sprintf("- Total Providers: %d\n", total)
		msg += fmt.Sprintf("- Active Providers: %d\n", active)
		if input.ActiveOnly {
			msg += fmt.Sprintf("- Filter: active_only=%t\n", input.ActiveOnly)
		}
		if input.Method != "" {
			msg += fmt.Sprintf("- Filter: method=%s\n", input.Method)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}
}

func registerProviderRecommend(s *mcp.Server, providerConfig *config.ProvidersConfig) {
	tool := &mcp.Tool{
		Name:        "provider_recommend",
		Description: "Get optimal provider/model for a task",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"task": map[string]any{
					"type":        "string",
					"description": "Task description or prompt",
				},
				"files": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "List of files for task",
				},
				"context_size": map[string]any{
					"type":        "number",
					"description": "Estimated token count",
				},
				"complexity": map[string]any{
					"type":        "string",
					"enum":        []any{"simple", "medium", "complex"},
					"description": "Task complexity",
				},
				"priority": map[string]any{
					"type":        "string",
					"enum":        []any{"low", "medium", "high"},
					"description": "Task priority",
				},
			},
			"required": []string{"task"},
		},
	}

	baseHandler := makeProviderRecommendHandler(providerConfig)
	handler := WithMetrics[ProviderRecommendInput, any]("provider_recommend", baseHandler)
	mcp.AddTool(s, tool, handler)
}

func makeProviderRecommendHandler(providerConfig *config.ProvidersConfig) func(context.Context, *mcp.CallToolRequest, ProviderRecommendInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ProviderRecommendInput) (*mcp.CallToolResult, any, error) {
		if input.Task == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Error: task is required",
				}},
			}, nil, fmt.Errorf("task is required")
		}

		if providerConfig == nil || len(providerConfig.Providers) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Error: no providers configured",
				}},
			}, nil, fmt.Errorf("no providers configured")
		}

		task := execution.NewTask(input.Task)

		if len(input.Files) > 0 {
			task.Context = &execution.TaskContext{
				Files: input.Files,
			}
		}

		if input.Complexity != "" {
			task.WithMetadata("complexity", input.Complexity)
		}

		if input.Priority != "" {
			task.WithMetadata("priority", input.Priority)
		}

		workerConfig := &worker.Config{
			Providers: make(map[string]worker.ProviderDefinition),
		}

		for name, provider := range providerConfig.Providers {
			workerConfig.Providers[name] = worker.ProviderDefinition(provider)
		}

		r, err := router.NewRouter(workerConfig, nil)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Failed to initialize router: %v", err),
				}},
			}, nil, fmt.Errorf("initialize router: %w", err)
		}

		decision, err := r.Route(ctx, task)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Failed to route task: %v", err),
				}},
			}, nil, fmt.Errorf("route task: %w", err)
		}

		taskChars := map[string]interface{}{
			"complexity": input.Complexity,
			"priority":   input.Priority,
			"file_count": len(input.Files),
		}
		if input.ContextSize > 0 {
			taskChars["context_size"] = input.ContextSize
		} else {
			estimatedTokens := len(input.Files)*2000 + len(input.Task)/4
			taskChars["estimated_tokens"] = estimatedTokens
		}

		alternatives := buildAlternatives(providerConfig, decision, task)

		response := ProviderRecommendResponse{
			RecommendedProvider: decision.Provider,
			RecommendedModel:    decision.Model,
			RecommendedMethod:   decision.Method.String(),
			Reason:              decision.Reason,
			EstimatedCost:       decision.EstimatedCost,
			Alternatives:        alternatives,
			TaskCharacteristics: taskChars,
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
			}, nil, fmt.Errorf("marshal response: %w", err)
		}

		msg := fmt.Sprintf("✅ Provider Recommendation\n\n```json\n%s\n```\n\n", string(data))
		msg += "**Recommendation:**\n"
		msg += fmt.Sprintf("- **Provider:** %s\n", decision.Provider)
		msg += fmt.Sprintf("- **Model:** %s\n", decision.Model)
		msg += fmt.Sprintf("- **Method:** %s\n", decision.Method.String())
		msg += fmt.Sprintf("- **Estimated Cost:** $%.4f\n", decision.EstimatedCost)
		msg += fmt.Sprintf("- **Reason:** %s\n\n", decision.Reason)

		if len(alternatives) > 0 {
			msg += "**Alternative Providers:**\n"
			for i, alt := range alternatives {
				msg += fmt.Sprintf("%d. %s (%s)\n", i+1, alt.Provider, alt.Model)
				msg += fmt.Sprintf("   - Cost: $%.4f\n", alt.Cost)
				msg += fmt.Sprintf("   - Reason: %s\n", alt.Reason)
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}
}

func buildAlternatives(providerConfig *config.ProvidersConfig, decision *router.RoutingDecision, task *execution.Task) []ProviderAlternative {
	var alternatives []ProviderAlternative

	providersConfig := make(map[string]worker.ProviderDefinition)
	for name, def := range providerConfig.Providers {
		providersConfig[name] = worker.ProviderDefinition(def)
	}

	baseCost := 0.01
	if task.Context != nil && len(task.Context.Files) > 0 {
		estimatedTokens := len(task.Context.Files)*2000 + len(task.Description)/4
		if estimatedTokens > 100000 {
			baseCost = 0.05
		} else if estimatedTokens > 50000 {
			baseCost = 0.03
		} else if estimatedTokens > 20000 {
			baseCost = 0.02
		}
	}

	for name, provider := range providersConfig {
		if name == decision.Provider {
			continue
		}

		costMultiplier := 1.0
		if provider.Method == config.MethodACP {
			costMultiplier = 1.5
		} else if provider.Method == config.MethodAPI {
			costMultiplier = 1.0
		} else if provider.Method == config.MethodCLI {
			costMultiplier = 0.5
		}

		if provider.Provider == "anthropic" {
			costMultiplier *= 2.0
		} else if provider.Provider == "moonshot" {
			costMultiplier *= 1.0
		} else if provider.Provider == "deepseek" {
			costMultiplier *= 0.5
		}

		altCost := baseCost * costMultiplier

		reason := ""
		if altCost < decision.EstimatedCost {
			reason = fmt.Sprintf("cheaper option ($%.4f vs $%.4f)", altCost, decision.EstimatedCost)
		} else if provider.ContextLimit > 100000 {
			reason = fmt.Sprintf("larger context window (%d tokens)", provider.ContextLimit)
		} else {
			reason = "alternative provider with different capabilities"
		}

		alternatives = append(alternatives, ProviderAlternative{
			Provider: name,
			Model:    provider.Model,
			Cost:     altCost,
			Reason:   reason,
		})
	}

	if len(alternatives) > 3 {
		alternatives = alternatives[:3]
	}

	return alternatives
}
