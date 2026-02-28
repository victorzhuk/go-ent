package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/spec"
	"github.com/victorzhuk/go-ent/internal/spec/storage"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type RegistryStatusInput struct{}

type RegistryStatusResponse struct {
	Stats RegistryStats `json:"stats"`
}

type RegistryStats struct {
	TotalChanges      int               `json:"total_changes"`
	TotalTasks        int               `json:"total_tasks"`
	CompletedTasks    int               `json:"completed_tasks"`
	CompletionPercent float64           `json:"completion_percent"`
	ChangesByStatus   map[string]int    `json:"changes_by_status"`
	TasksByStatus     map[string]int    `json:"tasks_by_status"`
	ActiveChanges     []ChangeQuickInfo `json:"active_changes"`
}

type ChangeQuickInfo struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Status    string  `json:"status"`
	Progress  float64 `json:"progress"`
	TaskCount int     `json:"task_count"`
}

func registerRegistryStatus(s *mcp.Server, toolRegistry *ToolRegistry, store *storage.BoltStore) {
	tool := &mcp.Tool{
		Name:        "registry_status",
		Description: "Get aggregated stats from the OpenSpec registry",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}

	mcp.AddTool(s, tool, registryStatusHandler(store))
	toolRegistry.Register("registry_status", tool.Description, "registry")
}

func registryStatusHandler(store *storage.BoltStore) func(ctx context.Context, req *mcp.CallToolRequest, input RegistryStatusInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input RegistryStatusInput) (*mcp.CallToolResult, any, error) {
		changes, err := store.ListAllChanges()
		if err != nil {
			return nil, nil, fmt.Errorf("list changes: %w", err)
		}

		tasks, err := store.ListTasks("", "")
		if err != nil {
			return nil, nil, fmt.Errorf("list tasks: %w", err)
		}

		totalChanges := len(changes)
		totalTasks := len(tasks)
		completedTasks := 0
		completionPercent := float64(0)

		changesByStatus := make(map[string]int)
		tasksByStatus := make(map[string]int)
		activeChanges := []ChangeQuickInfo{}

		for _, change := range changes {
			changesByStatus[string(change.Status)]++

			if change.TaskCount > 0 {
				progress := float64(change.Completed) / float64(change.TaskCount) * 100
				completionPercent += progress
			}
		}

		if totalChanges > 0 {
			completionPercent /= float64(totalChanges)
		}

		for _, task := range tasks {
			tasksByStatus[string(task.Status)]++
			if task.Status == spec.TaskCompleted {
				completedTasks++
			}
		}

		for _, change := range changes {
			if change.Status == spec.StatusActive || change.Status == spec.StatusDraft {
				progress := float64(0)
				if change.TaskCount > 0 {
					progress = float64(change.Completed) / float64(change.TaskCount) * 100
				}
				activeChanges = append(activeChanges, ChangeQuickInfo{
					ID:        change.ID,
					Title:     change.Title,
					Status:    string(change.Status),
					Progress:  progress,
					TaskCount: change.TaskCount,
				})
			}
		}

		stats := RegistryStats{
			TotalChanges:      totalChanges,
			TotalTasks:        totalTasks,
			CompletedTasks:    completedTasks,
			CompletionPercent: completionPercent,
			ChangesByStatus:   changesByStatus,
			TasksByStatus:     tasksByStatus,
			ActiveChanges:     activeChanges,
		}

		var sb strings.Builder
		sb.WriteString("# Registry Status\n\n")

		sb.WriteString("## Overview\n\n")
		fmt.Fprintf(&sb, "- Total Changes: %d\n", totalChanges)
		fmt.Fprintf(&sb, "- Total Tasks: %d\n", totalTasks)
		fmt.Fprintf(&sb, "- Completed Tasks: %d\n", completedTasks)
		fmt.Fprintf(&sb, "- Overall Completion: %.1f%%\n\n", completionPercent)

		sb.WriteString("## Changes by Status\n\n")
		caser := cases.Title(language.English)
		statuses := []spec.ChangeStatus{spec.StatusDraft, spec.StatusActive, spec.StatusApproved, spec.StatusArchived}
		for _, status := range statuses {
			count := changesByStatus[string(status)]
			if count > 0 {
				fmt.Fprintf(&sb, "- %s: %d\n", caser.String(string(status)), count)
			}
		}
		sb.WriteString("\n")

		sb.WriteString("## Tasks by Status\n\n")
		taskStatuses := []spec.TaskStatus{spec.TaskPending, spec.TaskInProgress, spec.TaskCompleted}
		for _, status := range taskStatuses {
			count := tasksByStatus[string(status)]
			if count > 0 {
				icon := "⏳"
				switch status {
				case spec.TaskCompleted:
					icon = "✅"
				case spec.TaskInProgress:
					icon = "🔄"
				}
				fmt.Fprintf(&sb, "- %s %s: %d\n", icon, strings.ReplaceAll(string(status), "_", " "), count)
			}
		}
		sb.WriteString("\n")

		if len(activeChanges) > 0 {
			sb.WriteString("## Active Changes\n\n")
			for i, change := range activeChanges {
				fmt.Fprintf(&sb, "%d. **%s** (%s)\n", i+1, change.ID, change.Status)
				fmt.Fprintf(&sb, "   %s\n   Progress: %.0f%% (%d/%d tasks)\n\n",
					change.Title, change.Progress, int(change.Progress*float64(change.TaskCount)/100), change.TaskCount)
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, RegistryStatusResponse{Stats: stats}, nil
	}
}
