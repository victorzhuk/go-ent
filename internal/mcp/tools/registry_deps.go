package tools

import (
	"context"
	"fmt"
	"sort"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/spec"
)

type RegistryDepsInput struct {
	ChangeID string `json:"change_id"`
	TaskNum  string `json:"task_num,omitempty"`
}

type RegistryDepsResponse struct {
	ChangeID      string               `json:"change_id"`
	Dependencies  []DependencyRelation `json:"dependencies"`
	Dependents    []DependencyRelation `json:"dependents"`
	BlockingTasks []DependencyRelation `json:"blocking_tasks"`
	BlockedBy     []DependencyRelation `json:"blocked_by"`
}

type DependencyRelation struct {
	ChangeID string          `json:"change_id"`
	TaskNum  string          `json:"task_num"`
	Content  string          `json:"content"`
	Status   spec.TaskStatus `json:"status"`
}

func registerRegistryDeps(s *mcp.Server, toolRegistry *ToolRegistry, store *spec.BoltStore) {
	tool := &mcp.Tool{
		Name:        "registry_deps",
		Description: "Show dependency graph for a task or change",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"change_id": map[string]any{
					"type":        "string",
					"description": "Change ID",
				},
				"task_num": map[string]any{
					"type":        "string",
					"description": "Optional task number to show specific task dependencies",
				},
			},
			"required": []string{"change_id"},
		},
	}

	mcp.AddTool(s, tool, registryDepsHandler(store))
	toolRegistry.Register("registry_deps", tool.Description, "registry")
}

func registryDepsHandler(store *spec.BoltStore) func(ctx context.Context, req *mcp.CallToolRequest, input RegistryDepsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input RegistryDepsInput) (*mcp.CallToolResult, any, error) {
		if input.ChangeID == "" {
			return nil, nil, fmt.Errorf("change_id is required")
		}

		change, err := store.GetChange(input.ChangeID)
		if err != nil {
			return nil, nil, fmt.Errorf("get change %s: %w", input.ChangeID, err)
		}

		if change == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("# Dependency Graph\n\nChange not found: %s", input.ChangeID)}},
			}, nil, nil
		}

		var content string

		if input.TaskNum != "" {
			task, err := store.GetTask(input.ChangeID, input.TaskNum)
			if err != nil {
				return nil, nil, fmt.Errorf("get task %s/%s: %w", input.ChangeID, input.TaskNum, err)
			}

			if task == nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("# Dependency Graph\n\nTask not found: %s/%s", input.ChangeID, input.TaskNum)}},
				}, nil, nil
			}

			resp, err := buildTaskDeps(store, task)
			if err != nil {
				return nil, nil, fmt.Errorf("build task deps: %w", err)
			}

			content = formatTaskDeps(change, task, resp)
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: content}},
			}, resp, nil
		}

		resp, err := buildChangeDeps(store, input.ChangeID)
		if err != nil {
			return nil, nil, fmt.Errorf("build change deps: %w", err)
		}

		content = formatChangeDeps(change, resp)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: content}},
		}, resp, nil
	}
}

func buildTaskDeps(store *spec.BoltStore, task *spec.Task) (*RegistryDepsResponse, error) {
	deps, err := store.GetDeps(task.ChangeID, task.TaskNum)
	if err != nil {
		return nil, fmt.Errorf("get deps: %w", err)
	}

	resp := &RegistryDepsResponse{
		ChangeID: task.ChangeID,
	}

	if deps != nil {
		for _, depNum := range deps.DependsOn {
			depTask, err := store.GetTask(task.ChangeID, depNum)
			if err != nil {
				return nil, fmt.Errorf("get dep task %s: %w", depNum, err)
			}
			if depTask != nil {
				resp.Dependencies = append(resp.Dependencies, DependencyRelation{
					ChangeID: depTask.ChangeID,
					TaskNum:  depTask.TaskNum,
					Content:  depTask.Content,
					Status:   depTask.Status,
				})
			}
		}

		for _, depID := range deps.DependedBy {
			depTask, err := getTaskByID(store, depID)
			if err != nil {
				return nil, fmt.Errorf("get dependent task %s: %w", depID, err)
			}
			if depTask != nil {
				resp.Dependents = append(resp.Dependents, DependencyRelation{
					ChangeID: depTask.ChangeID,
					TaskNum:  depTask.TaskNum,
					Content:  depTask.Content,
					Status:   depTask.Status,
				})
			}
		}

		for _, blockingID := range deps.BlockingTasks {
			blockTask, err := getTaskByID(store, blockingID)
			if err != nil {
				return nil, fmt.Errorf("get blocking task %s: %w", blockingID, err)
			}
			if blockTask != nil {
				resp.BlockingTasks = append(resp.BlockingTasks, DependencyRelation{
					ChangeID: blockTask.ChangeID,
					TaskNum:  blockTask.TaskNum,
					Content:  blockTask.Content,
					Status:   blockTask.Status,
				})
			}
		}
	}

	return resp, nil
}

func buildChangeDeps(store *spec.BoltStore, changeID string) (*RegistryDepsResponse, error) {
	tasks, err := store.GetTasksByChange(changeID)
	if err != nil {
		return nil, fmt.Errorf("get tasks: %w", err)
	}

	resp := &RegistryDepsResponse{
		ChangeID: changeID,
	}

	for _, task := range tasks {
		if task.Status == spec.TaskPending {
			deps, err := store.GetDeps(task.ChangeID, task.TaskNum)
			if err != nil {
				return nil, fmt.Errorf("get deps for %s/%s: %w", task.ChangeID, task.TaskNum, err)
			}

			if deps != nil && deps.IsBlocked {
				for _, blockingID := range deps.BlockingTasks {
					blockTask, err := getTaskByID(store, blockingID)
					if err != nil {
						return nil, fmt.Errorf("get blocking task %s: %w", blockingID, err)
					}
					if blockTask != nil && !containsDependencyRelation(resp.BlockingTasks, blockTask.TaskNum) {
						resp.BlockingTasks = append(resp.BlockingTasks, DependencyRelation{
							ChangeID: blockTask.ChangeID,
							TaskNum:  blockTask.TaskNum,
							Content:  blockTask.Content,
							Status:   blockTask.Status,
						})
					}
				}
				if !containsDependencyRelation(resp.BlockedBy, task.TaskNum) {
					resp.BlockedBy = append(resp.BlockedBy, DependencyRelation{
						ChangeID: task.ChangeID,
						TaskNum:  task.TaskNum,
						Content:  task.Content,
						Status:   task.Status,
					})
				}
			}
		}
	}

	return resp, nil
}

func formatTaskDeps(change *spec.ChangeMetadata, task *spec.Task, resp *RegistryDepsResponse) string {
	var content string
	content += "# Dependency Graph\n\n"
	content += fmt.Sprintf("## Change: %s\n\n", change.ID)
	content += fmt.Sprintf("**Title**: %s\n\n", change.Title)
	content += "---\n\n"
	content += fmt.Sprintf("## Task: %s\n\n", task.TaskNum)
	content += fmt.Sprintf("**Content**: %s\n\n", task.Content)
	content += fmt.Sprintf("**Status**: %s\n\n", task.Status)
	content += "---\n\n"

	if len(resp.Dependencies) > 0 {
		content += "### Dependencies (this task depends on)\n\n"
		for i, dep := range resp.Dependencies {
			statusIcon := "✅"
			switch dep.Status {
			case spec.TaskPending:
				statusIcon = "⏳"
			case spec.TaskInProgress:
				statusIcon = "🔄"
			}
			content += fmt.Sprintf("**%d. %s** %s\n", i+1, dep.TaskNum, statusIcon)
			content += fmt.Sprintf("- Status: %s\n", dep.Status)
			content += fmt.Sprintf("- Content: %s\n\n", dep.Content)
		}
	} else {
		content += "### Dependencies\n\nNo dependencies\n\n"
	}

	if len(resp.Dependents) > 0 {
		content += "### Dependents (tasks waiting on this one)\n\n"
		for i, dep := range resp.Dependents {
			statusIcon := "✅"
			switch dep.Status {
			case spec.TaskPending:
				statusIcon = "⏳"
			case spec.TaskInProgress:
				statusIcon = "🔄"
			}
			content += fmt.Sprintf("**%d. %s** %s\n", i+1, dep.TaskNum, statusIcon)
			content += fmt.Sprintf("- Status: %s\n", dep.Status)
			content += fmt.Sprintf("- Content: %s\n\n", dep.Content)
		}
	} else {
		content += "### Dependents\n\nNo tasks depend on this one\n\n"
	}

	if len(resp.BlockingTasks) > 0 {
		content += "### Blocking Tasks (incomplete dependencies)\n\n"
		for i, dep := range resp.BlockingTasks {
			content += fmt.Sprintf("**%d. %s** ⏳\n", i+1, dep.TaskNum)
			content += fmt.Sprintf("- Status: %s\n", dep.Status)
			content += fmt.Sprintf("- Content: %s\n\n", dep.Content)
		}
	}

	return content
}

func formatChangeDeps(change *spec.ChangeMetadata, resp *RegistryDepsResponse) string {
	var content string
	content += "# Dependency Graph\n\n"
	content += fmt.Sprintf("## Change: %s\n\n", change.ID)
	content += fmt.Sprintf("**Title**: %s\n\n", change.Title)
	content += "---\n\n"

	sort.Slice(resp.BlockingTasks, func(i, j int) bool {
		return resp.BlockingTasks[i].TaskNum < resp.BlockingTasks[j].TaskNum
	})
	sort.Slice(resp.BlockedBy, func(i, j int) bool {
		return resp.BlockedBy[i].TaskNum < resp.BlockedBy[j].TaskNum
	})

	if len(resp.BlockingTasks) > 0 {
		content += "### Blocking Tasks (blocking other pending tasks)\n\n"
		for i, task := range resp.BlockingTasks {
			content += fmt.Sprintf("**%d. %s** - %s\n\n", i+1, task.TaskNum, task.Status)
			content += fmt.Sprintf("%s\n\n", task.Content)
		}
	} else {
		content += "### Blocking Tasks\n\nNo blocking tasks\n\n"
	}

	if len(resp.BlockedBy) > 0 {
		content += "### Blocked Tasks (waiting for dependencies)\n\n"
		for i, task := range resp.BlockedBy {
			content += fmt.Sprintf("**%d. %s** - %s\n\n", i+1, task.TaskNum, task.Status)
			content += fmt.Sprintf("%s\n\n", task.Content)
		}
	} else {
		content += "### Blocked Tasks\n\nNo blocked tasks\n\n"
	}

	if len(resp.BlockingTasks) == 0 && len(resp.BlockedBy) == 0 {
		content += "### Summary\n\nNo dependencies blocking progress. All pending tasks can be worked on.\n\n"
	}

	return content
}

func getTaskByID(store *spec.BoltStore, taskID string) (*spec.Task, error) {
	changeID, taskNum, err := parseTaskID(taskID)
	if err != nil {
		return nil, err
	}
	return store.GetTask(changeID, taskNum)
}

func parseTaskID(taskID string) (string, string, error) {
	for i := len(taskID) - 1; i >= 0; i-- {
		if taskID[i] == ':' {
			return taskID[:i], taskID[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("invalid task id format: %s", taskID)
}

func formatTaskNums(nums []string) string {
	result := ""
	for i, num := range nums {
		if i > 0 {
			result += ", "
		}
		result += num
	}
	return result
}

func containsDependencyRelation(relations []DependencyRelation, taskNum string) bool {
	for _, r := range relations {
		if r.TaskNum == taskNum {
			return true
		}
	}
	return false
}
