package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/spec"
)

type RegistryGetChangeInput struct {
	ChangeID string `json:"change_id"`
}

type RegistryGetChangeResponse struct {
	Change  *ChangeDetail     `json:"change"`
	Tasks   []TaskWithDetails `json:"tasks"`
	Summary TaskSummaryStats  `json:"summary"`
}

type ChangeDetail struct {
	ID         string            `json:"id"`
	Title      string            `json:"title"`
	Status     spec.ChangeStatus `json:"status"`
	TaskCount  int               `json:"task_count"`
	Completed  int               `json:"completed"`
	InProgress int               `json:"in_progress"`
	Blocked    int               `json:"blocked"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
}

type TaskWithDetails struct {
	ID        string          `json:"id"`
	ChangeID  string          `json:"change_id"`
	TaskNum   string          `json:"task_num"`
	Content   string          `json:"content"`
	Status    spec.TaskStatus `json:"status"`
	Priority  int             `json:"priority"`
	DependsOn []string        `json:"depends_on"`
	SyncedAt  string          `json:"synced_at"`
}

type TaskSummaryStats struct {
	Total      int `json:"total"`
	Pending    int `json:"pending"`
	InProgress int `json:"in_progress"`
	Completed  int `json:"completed"`
}

func registerRegistryGetChange(s *mcp.Server, toolRegistry *ToolRegistry, store *spec.BoltStore) {
	tool := &mcp.Tool{
		Name:        "registry_get_change",
		Description: "Get detailed change info with all tasks",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"change_id": map[string]any{
					"type":        "string",
					"description": "Change ID to fetch",
				},
			},
			"required": []string{"change_id"},
		},
	}

	mcp.AddTool(s, tool, registryGetChangeHandler(store))
	toolRegistry.Register("registry_get_change", tool.Description, "registry")
}

func registryGetChangeHandler(store *spec.BoltStore) func(ctx context.Context, req *mcp.CallToolRequest, input RegistryGetChangeInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input RegistryGetChangeInput) (*mcp.CallToolResult, any, error) {
		if input.ChangeID == "" {
			return nil, nil, fmt.Errorf("change_id is required")
		}

		change, err := store.GetChange(input.ChangeID)
		if err != nil {
			return nil, nil, fmt.Errorf("get change %s: %w", input.ChangeID, err)
		}

		if change == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("# Change Not Found\n\nNo change found with ID: %s", input.ChangeID)}},
			}, nil, nil
		}

		tasks, err := store.GetTasksByChange(input.ChangeID)
		if err != nil {
			return nil, nil, fmt.Errorf("get tasks for change %s: %w", input.ChangeID, err)
		}

		var summaryStats TaskSummaryStats
		taskDetails := make([]TaskWithDetails, 0, len(tasks))

		for _, task := range tasks {
			summaryStats.Total++
			switch task.Status {
			case spec.TaskPending:
				summaryStats.Pending++
			case spec.TaskInProgress:
				summaryStats.InProgress++
			case spec.TaskCompleted:
				summaryStats.Completed++
			}

			taskDetails = append(taskDetails, TaskWithDetails{
				ID:        task.ID,
				ChangeID:  task.ChangeID,
				TaskNum:   task.TaskNum,
				Content:   task.Content,
				Status:    task.Status,
				Priority:  task.Priority,
				DependsOn: task.DependsOn,
				SyncedAt:  task.SyncedAt.Format("2006-01-02 15:04"),
			})
		}

		changeDetail := &ChangeDetail{
			ID:         change.ID,
			Title:      change.Title,
			Status:     change.Status,
			TaskCount:  change.TaskCount,
			Completed:  change.Completed,
			InProgress: change.InProgress,
			Blocked:    change.Blocked,
			CreatedAt:  change.CreatedAt.Format("2006-01-02 15:04"),
			UpdatedAt:  change.UpdatedAt.Format("2006-01-02 15:04"),
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Change: %s\n\n", change.ID))
		sb.WriteString(fmt.Sprintf("**Title**: %s\n\n", change.Title))
		sb.WriteString(fmt.Sprintf("**Status**: %s\n\n", change.Status))

		progress := float64(0)
		if change.TaskCount > 0 {
			progress = float64(change.Completed) / float64(change.TaskCount) * 100
		}
		sb.WriteString(fmt.Sprintf("**Progress**: %d/%d tasks (%.0f%%)\n\n", change.Completed, change.TaskCount, progress))

		sb.WriteString("## Summary\n\n")
		sb.WriteString(fmt.Sprintf("- Total Tasks: %d\n", summaryStats.Total))
		sb.WriteString(fmt.Sprintf("- Pending: %d\n", summaryStats.Pending))
		sb.WriteString(fmt.Sprintf("- In Progress: %d\n", summaryStats.InProgress))
		sb.WriteString(fmt.Sprintf("- Completed: %d\n\n", summaryStats.Completed))

		sb.WriteString("## Details\n\n")
		sb.WriteString(fmt.Sprintf("- Created: %s\n", change.CreatedAt.Format("2006-01-02 15:04")))
		sb.WriteString(fmt.Sprintf("- Updated: %s\n\n", change.UpdatedAt.Format("2006-01-02 15:04")))

		sb.WriteString("## Tasks\n\n")

		for i, task := range tasks {
			statusIcon := "⏳"
			switch task.Status {
			case spec.TaskCompleted:
				statusIcon = "✅"
			case spec.TaskInProgress:
				statusIcon = "🔄"
			}

			sb.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, task.TaskNum))
			sb.WriteString(fmt.Sprintf("**Status**: %s %s\n\n", statusIcon, task.Status))
			sb.WriteString(fmt.Sprintf("**Content**: %s\n\n", task.Content))

			if len(task.DependsOn) > 0 {
				sb.WriteString(fmt.Sprintf("**Depends On**: %s\n\n", strings.Join(task.DependsOn, ", ")))
			}
		}

		return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
			}, RegistryGetChangeResponse{
				Change:  changeDetail,
				Tasks:   taskDetails,
				Summary: summaryStats,
			}, nil
	}
}
