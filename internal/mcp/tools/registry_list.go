package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/spec"
	"github.com/victorzhuk/go-ent/internal/spec/storage"
)

type RegistryListChangesInput struct{}

type RegistryListChangesResponse struct {
	Changes []ChangeSummary `json:"changes"`
	Total   int             `json:"total"`
}

type ChangeSummary struct {
	ID         string            `json:"id"`
	Title      string            `json:"title"`
	Status     spec.ChangeStatus `json:"status"`
	TaskCount  int               `json:"task_count"`
	Completed  int               `json:"completed"`
	InProgress int               `json:"in_progress"`
	Blocked    int               `json:"blocked"`
	UpdatedAt  string            `json:"updated_at"`
}

func registerRegistryListChanges(s *mcp.Server, toolRegistry *ToolRegistry, store *storage.BoltStore) {
	tool := &mcp.Tool{
		Name:        "registry_list_changes",
		Description: "List all changes from the OpenSpec registry",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}

	mcp.AddTool(s, tool, registryListChangesHandler(store))
	toolRegistry.Register("registry_list_changes", tool.Description, "registry")
}

func registryListChangesHandler(store *storage.BoltStore) func(ctx context.Context, req *mcp.CallToolRequest, input RegistryListChangesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input RegistryListChangesInput) (*mcp.CallToolResult, any, error) {
		changes, err := store.ListAllChanges()
		if err != nil {
			return nil, nil, fmt.Errorf("list changes: %w", err)
		}

		if len(changes) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "# Registry Changes\n\nNo changes found in the registry."}},
			}, RegistryListChangesResponse{Changes: []ChangeSummary{}, Total: 0}, nil
		}

		var sb strings.Builder
		sb.WriteString("# Registry Changes\n\n")
		fmt.Fprintf(&sb, "Found %d change(s):\n\n", len(changes))

		summaryList := make([]ChangeSummary, 0, len(changes))

		for i, change := range changes {
			progress := float64(0)
			if change.TaskCount > 0 {
				progress = float64(change.Completed) / float64(change.TaskCount) * 100
			}

			fmt.Fprintf(&sb, "## %d. %s\n\n", i+1, change.ID)
			fmt.Fprintf(&sb, "**Title**: %s\n\n", change.Title)
			fmt.Fprintf(&sb, "**Status**: %s\n\n", change.Status)
			fmt.Fprintf(&sb, "**Progress**: %d/%d tasks (%.0f%%)\n\n",
				change.Completed, change.TaskCount, progress)

			if change.InProgress > 0 {
				fmt.Fprintf(&sb, "**In Progress**: %d\n\n", change.InProgress)
			}

			if change.Blocked > 0 {
				fmt.Fprintf(&sb, "**Blocked**: %d\n\n", change.Blocked)
			}

			fmt.Fprintf(&sb, "**Updated**: %s\n\n", change.UpdatedAt.Format(spec.DateTimeFormat))
			sb.WriteString("---\n\n")

			summaryList = append(summaryList, ChangeSummary{
				ID:         change.ID,
				Title:      change.Title,
				Status:     change.Status,
				TaskCount:  change.TaskCount,
				Completed:  change.Completed,
				InProgress: change.InProgress,
				Blocked:    change.Blocked,
				UpdatedAt:  change.UpdatedAt.Format(spec.DateTimeFormat),
			})
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, RegistryListChangesResponse{Changes: summaryList, Total: len(changes)}, nil
	}
}

type RegistryListTasksInput struct {
	ChangeID string `json:"change_id,omitempty"`
	Status   string `json:"status,omitempty"`
}

type RegistryListTasksResponse struct {
	Tasks []TaskSummary `json:"tasks"`
	Total int           `json:"total"`
}

type TaskSummary struct {
	ID        string          `json:"id"`
	ChangeID  string          `json:"change_id"`
	TaskNum   string          `json:"task_num"`
	Content   string          `json:"content"`
	Status    spec.TaskStatus `json:"status"`
	Priority  int             `json:"priority"`
	DependsOn []string        `json:"depends_on,omitempty"`
}

func registerRegistryListTasks(s *mcp.Server, toolRegistry *ToolRegistry, store *storage.BoltStore) {
	tool := &mcp.Tool{
		Name:        "registry_list_tasks",
		Description: "List tasks from the OpenSpec registry with optional filtering",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"change_id": map[string]any{
					"type":        "string",
					"description": "Filter by change ID (empty for all changes)",
				},
				"status": map[string]any{
					"type":        "string",
					"description": "Filter by task status (pending, in_progress, completed)",
					"enum":        []string{"pending", "in_progress", "completed"},
				},
			},
		},
	}

	mcp.AddTool(s, tool, registryListTasksHandler(store))
	toolRegistry.Register("registry_list_tasks", tool.Description, "registry")
}

func registryListTasksHandler(store *storage.BoltStore) func(ctx context.Context, req *mcp.CallToolRequest, input RegistryListTasksInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input RegistryListTasksInput) (*mcp.CallToolResult, any, error) {
		tasks, err := store.ListTasks(input.ChangeID, input.Status)
		if err != nil {
			return nil, nil, fmt.Errorf("list tasks: %w", err)
		}

		if len(tasks) == 0 {
			msg := "# Registry Tasks\n\nNo tasks found"
			if input.ChangeID != "" || input.Status != "" {
				msg = fmt.Sprintf("# Registry Tasks\n\nNo tasks found matching filters (change_id=%q, status=%q)", input.ChangeID, input.Status)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: msg}},
			}, RegistryListTasksResponse{Tasks: []TaskSummary{}, Total: 0}, nil
		}

		var sb strings.Builder
		sb.WriteString("# Registry Tasks\n\n")

		if input.ChangeID != "" || input.Status != "" {
			sb.WriteString("*Filtered*\n\n")
			if input.ChangeID != "" {
				fmt.Fprintf(&sb, "- Change ID: %s\n", input.ChangeID)
			}
			if input.Status != "" {
				fmt.Fprintf(&sb, "- Status: %s\n", input.Status)
			}
			sb.WriteString("\n")
		}

		fmt.Fprintf(&sb, "Found %d task(s):\n\n", len(tasks))

		summaryList := make([]TaskSummary, 0, len(tasks))

		for i, task := range tasks {
			fmt.Fprintf(&sb, "## %d. %s - %s\n\n", i+1, task.ChangeID, task.TaskNum)

			fmt.Fprintf(&sb, "**Status**: %s %s\n\n", task.StatusIcon(), task.Status)
			fmt.Fprintf(&sb, "**Content**: %s\n\n", task.Content)

			if task.Priority != 0 {
				fmt.Fprintf(&sb, "**Priority**: %d\n\n", task.Priority)
			}

			if len(task.DependsOn) > 0 {
				fmt.Fprintf(&sb, "**Depends On**: %s\n\n", strings.Join(task.DependsOn, ", "))
			}

			sb.WriteString("---\n\n")

			summaryList = append(summaryList, TaskSummary{
				ID:        task.ID,
				ChangeID:  task.ChangeID,
				TaskNum:   task.TaskNum,
				Content:   task.Content,
				Status:    task.Status,
				Priority:  task.Priority,
				DependsOn: task.DependsOn,
			})
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, RegistryListTasksResponse{Tasks: summaryList, Total: len(tasks)}, nil
	}
}
