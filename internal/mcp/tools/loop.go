package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/spec"
)

type LoopGetInput struct {
	Path string `json:"path"`
}

type LoopSetInput struct {
	Path       string `json:"path"`
	Task       string `json:"task,omitempty"`
	Iteration  *int   `json:"iteration,omitempty"`
	MaxIter    *int   `json:"max_iterations,omitempty"`
	LastError  string `json:"last_error,omitempty"`
	Adjustment string `json:"adjustment,omitempty"`
	Status     string `json:"status,omitempty"`
}

type LoopStartInput struct {
	Path    string `json:"path"`
	Task    string `json:"task"`
	MaxIter int    `json:"max_iterations,omitempty"`
}

type LoopCancelInput struct {
	Path string `json:"path"`
}

func registerLoop(s *mcp.Server) {
	startTool := &mcp.Tool{
		Name:        "loop_start",
		Description: "Start autonomous loop with self-correction",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":           map[string]any{"type": "string", "description": "Path to project directory"},
				"task":           map[string]any{"type": "string", "description": "Task to work on"},
				"max_iterations": map[string]any{"type": "integer", "description": "Maximum number of iterations", "default": 10},
			},
			"required": []string{"path", "task"},
		},
	}
	mcp.AddTool(s, startTool, loopStartHandler)

	getTool := &mcp.Tool{
		Name:        "loop_get",
		Description: "Get current loop state",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string", "description": "Path to project directory"},
			},
			"required": []string{"path"},
		},
	}
	mcp.AddTool(s, getTool, loopGetHandler)

	setTool := &mcp.Tool{
		Name:        "loop_set",
		Description: "Update loop state (iteration, error, adjustment, status)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":           map[string]any{"type": "string", "description": "Path to project directory"},
				"task":           map[string]any{"type": "string", "description": "Current task"},
				"iteration":      map[string]any{"type": "integer", "description": "Current iteration number"},
				"max_iterations": map[string]any{"type": "integer", "description": "Maximum iterations"},
				"last_error":     map[string]any{"type": "string", "description": "Last error encountered"},
				"adjustment":     map[string]any{"type": "string", "description": "Adjustment to make"},
				"status":         map[string]any{"type": "string", "description": "Loop status", "enum": []string{"running", "completed", "failed", "cancelled"}},
			},
			"required": []string{"path"},
		},
	}
	mcp.AddTool(s, setTool, loopSetHandler)

	cancelTool := &mcp.Tool{
		Name:        "loop_cancel",
		Description: "Cancel running loop",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string", "description": "Path to project directory"},
			},
			"required": []string{"path"},
		},
	}
	mcp.AddTool(s, cancelTool, loopCancelHandler)
}

func loopStartHandler(ctx context.Context, req *mcp.CallToolRequest, input LoopStartInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}
	if input.Task == "" {
		return nil, nil, fmt.Errorf("task is required")
	}

	maxIter := input.MaxIter
	if maxIter <= 0 {
		maxIter = 10
	}

	store := spec.NewStore(input.Path)

	if store.LoopExists() {
		existing, err := store.LoadLoop()
		if err == nil && existing.Status == spec.LoopStatusRunning {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Loop already running. Use loop_cancel to stop it first.",
				}},
			}, nil, nil
		}
	}

	loop := spec.NewLoopState(input.Task, maxIter)

	if err := store.SaveLoop(loop); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error starting loop: %v", err)}},
		}, nil, nil
	}

	msg := "✅ Loop started\n\n"
	msg += fmt.Sprintf("Task: %s\n", loop.Task)
	msg += fmt.Sprintf("Max iterations: %d\n", loop.MaxIter)
	msg += fmt.Sprintf("Status: %s\n", loop.Status)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}

func loopGetHandler(ctx context.Context, req *mcp.CallToolRequest, input LoopGetInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	store := spec.NewStore(input.Path)

	if !store.LoopExists() {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No loop state found"}},
		}, nil, nil
	}

	loop, err := store.LoadLoop()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error loading loop: %v", err)}},
		}, nil, nil
	}

	data, err := json.MarshalIndent(loop, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error formatting loop: %v", err)}},
		}, nil, nil
	}

	msg := fmt.Sprintf("Loop State:\n\n```json\n%s\n```", string(data))

	if !loop.ShouldContinue() {
		if loop.Status == spec.LoopStatusRunning {
			msg += "\n\n⚠️  Max iterations reached"
		} else {
			msg += fmt.Sprintf("\n\n⏹️  Loop ended: %s", loop.Status)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}

func loopSetHandler(ctx context.Context, req *mcp.CallToolRequest, input LoopSetInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	store := spec.NewStore(input.Path)

	if !store.LoopExists() {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No loop state found. Use loop_start first."}},
		}, nil, nil
	}

	loop, err := store.LoadLoop()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error loading loop: %v", err)}},
		}, nil, nil
	}

	changed := false

	if input.Task != "" {
		loop.Task = input.Task
		changed = true
	}

	if input.Iteration != nil {
		loop.Iteration = *input.Iteration
		changed = true
	}

	if input.MaxIter != nil {
		loop.MaxIter = *input.MaxIter
		changed = true
	}

	if input.LastError != "" {
		loop.RecordError(input.LastError)
		changed = true
	}

	if input.Adjustment != "" {
		loop.RecordAdjustment(input.Adjustment)
		changed = true
	}

	if input.Status != "" {
		loop.Status = spec.LoopStatus(input.Status)
		changed = true
	}

	if changed {
		if err := store.SaveLoop(loop); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error saving loop: %v", err)}},
			}, nil, nil
		}
	}

	msg := "✅ Loop state updated\n\n"
	msg += fmt.Sprintf("Iteration: %d/%d\n", loop.Iteration, loop.MaxIter)
	msg += fmt.Sprintf("Status: %s\n", loop.Status)

	if loop.LastError != "" {
		msg += fmt.Sprintf("Last error: %s\n", loop.LastError)
	}

	if len(loop.Adjustments) > 0 {
		msg += fmt.Sprintf("Adjustments: %d\n", len(loop.Adjustments))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}

func loopCancelHandler(ctx context.Context, req *mcp.CallToolRequest, input LoopCancelInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	store := spec.NewStore(input.Path)

	if !store.LoopExists() {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No loop running"}},
		}, nil, nil
	}

	loop, err := store.LoadLoop()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error loading loop: %v", err)}},
		}, nil, nil
	}

	loop.Cancel()

	if err := store.SaveLoop(loop); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error cancelling loop: %v", err)}},
		}, nil, nil
	}

	msg := "✅ Loop cancelled\n\n"
	msg += fmt.Sprintf("Task: %s\n", loop.Task)
	msg += fmt.Sprintf("Completed iterations: %d/%d\n", loop.Iteration, loop.MaxIter)
	msg += fmt.Sprintf("Total adjustments: %d\n", len(loop.Adjustments))

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}
