package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/spec"
	"github.com/victorzhuk/go-ent/internal/spec/storage"
)

type RegistryNextTaskInput struct {
	ChangeID string `json:"change_id,omitempty"`
}

type RegistryNextTaskResponse struct {
	TaskID    string   `json:"task_id"`
	ChangeID  string   `json:"change_id"`
	TaskNum   string   `json:"task_num"`
	Content   string   `json:"content"`
	Priority  int      `json:"priority"`
	DependsOn []string `json:"depends_on"`
}

func registerRegistryNextTask(s *mcp.Server, toolRegistry *ToolRegistry, store *storage.BoltStore) {
	tool := &mcp.Tool{
		Name:        "registry_next_task",
		Description: "Find next unblocked task that can be worked on",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"change_id": map[string]any{
					"type":        "string",
					"description": "Optional change ID to limit search to a specific change",
				},
			},
		},
	}

	mcp.AddTool(s, tool, registryNextTaskHandler(store))
	toolRegistry.Register("registry_next_task", tool.Description, "registry")
}

func registryNextTaskHandler(store *storage.BoltStore) func(ctx context.Context, req *mcp.CallToolRequest, input RegistryNextTaskInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input RegistryNextTaskInput) (*mcp.CallToolResult, any, error) {
		tasks, err := store.ListTasks(input.ChangeID, string(spec.TaskPending))
		if err != nil {
			return nil, nil, fmt.Errorf("list pending tasks: %w", err)
		}

		if len(tasks) == 0 {
			msg := "# Next Task\n\nNo unblocked tasks found"
			if input.ChangeID != "" {
				msg = fmt.Sprintf("# Next Task\n\nNo unblocked tasks found in change: %s", input.ChangeID)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: msg}},
			}, nil, nil
		}

		var candidates []*spec.Task

		for _, task := range tasks {
			deps, err := store.GetDeps(task.ChangeID, task.TaskNum)
			if err != nil {
				return nil, nil, fmt.Errorf("get deps for task %s/%s: %w", task.ChangeID, task.TaskNum, err)
			}

			if deps == nil || !deps.IsBlocked {
				candidates = append(candidates, task)
			}
		}

		if len(candidates) == 0 {
			msg := "# Next Task\n\nNo unblocked tasks found - all pending tasks have incomplete dependencies"
			if input.ChangeID != "" {
				msg = fmt.Sprintf("# Next Task\n\nNo unblocked tasks found in change %s - all pending tasks have incomplete dependencies", input.ChangeID)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: msg}},
			}, nil, nil
		}

		nextTask := candidates[0]
		change, err := store.GetChange(nextTask.ChangeID)
		if err != nil {
			return nil, nil, fmt.Errorf("get change %s: %w", nextTask.ChangeID, err)
		}

		deps, _ := store.GetDeps(nextTask.ChangeID, nextTask.TaskNum)

		var content string
		content += "# Next Task to Work On\n\n"
		content += fmt.Sprintf("## %s - %s\n\n", nextTask.ChangeID, nextTask.TaskNum)

		if change != nil {
			content += fmt.Sprintf("**Change**: %s\n\n", change.Title)
		}

		content += fmt.Sprintf("**Status**: %s\n\n", nextTask.Status)
		content += fmt.Sprintf("**Content**: %s\n\n", nextTask.Content)

		if nextTask.Priority != 0 {
			content += fmt.Sprintf("**Priority**: %d\n\n", nextTask.Priority)
		}

		if deps != nil && len(deps.DependsOn) > 0 {
			content += fmt.Sprintf("**Depends On**: %s\n\n", formatTaskNums(deps.DependsOn))
		}

		return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: content}},
			}, RegistryNextTaskResponse{
				TaskID:    nextTask.ID,
				ChangeID:  nextTask.ChangeID,
				TaskNum:   nextTask.TaskNum,
				Content:   nextTask.Content,
				Priority:  nextTask.Priority,
				DependsOn: nextTask.DependsOn,
			}, nil
	}
}
