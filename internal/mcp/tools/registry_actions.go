package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/hooks"
	"github.com/victorzhuk/go-ent/internal/spec"
)

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimLeft(s, cutset string) string {
	for len(s) > 0 {
		found := false
		for _, c := range cutset {
			if len(s) > 0 && rune(s[0]) == c {
				s = s[1:]
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	return s
}

type RegistryMarkDoneInput struct {
	ChangeID string `json:"change_id"`
	TaskNum  string `json:"task_num"`
}

type RegistryMarkDoneResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func registerRegistryMarkDone(s *mcp.Server, toolRegistry *ToolRegistry, cwd string, hookRegistry *hooks.Registry) {
	tool := &mcp.Tool{
		Name:        "registry_mark_done",
		Description: "Mark a task as completed by checking the checkbox in the tasks.md file",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"change_id": map[string]any{
					"type":        "string",
					"description": "The change ID (e.g., 'add-user-auth')",
				},
				"task_num": map[string]any{
					"type":        "string",
					"description": "The task number (e.g., '1.2')",
				},
			},
			"required": []string{"change_id", "task_num"},
		},
	}

	mcp.AddTool(s, tool, registryMarkDoneHandler(cwd, hookRegistry))
	toolRegistry.Register("registry_mark_done", tool.Description, "registry")
}

func registryMarkDoneHandler(cwd string, hookRegistry *hooks.Registry) func(ctx context.Context, req *mcp.CallToolRequest, input RegistryMarkDoneInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input RegistryMarkDoneInput) (*mcp.CallToolResult, any, error) {
		tasksPath := filepath.Join(cwd, "openspec", "changes", input.ChangeID, "tasks.md")

		// #nosec G304 - tasksPath is constructed from validated inputs
		data, err := os.ReadFile(tasksPath)
		if err != nil {
			return nil, nil, fmt.Errorf("read tasks.md: %w", err)
		}

		lines := splitLines(string(data))
		found := false

		for i, line := range lines {
			if strings.HasPrefix(line, "- [ ]") || strings.HasPrefix(line, "- [x]") {
				content := trimLeft(line[2:], " ")
				if len(content) >= 2 && strings.HasPrefix(content, "**") {
					remaining := content[2:]
					if idx := strings.Index(remaining, "**"); idx != -1 {
						taskNum := remaining[:idx]
						if taskNum == input.TaskNum {
							if strings.HasPrefix(line, "- [ ]") {
								lines[i] = strings.Replace(line, "- [ ]", "- [x]", 1)
								found = true
							} else {
								return &mcp.CallToolResult{
									Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("# Mark Task Done\n\nTask %s/%s is already marked as completed.", input.ChangeID, input.TaskNum)}},
								}, RegistryMarkDoneResponse{Success: false, Message: "Task already completed"}, nil
							}
						}
					}
				}
			}
		}

		if !found {
			return nil, nil, fmt.Errorf("task %s/%s not found", input.ChangeID, input.TaskNum)
		}

		content := strings.Join(lines, "\n")
		if err := os.WriteFile(tasksPath, []byte(content), 0o600); err != nil {
			return nil, nil, fmt.Errorf("write tasks.md: %w", err)
		}

		// Trigger onTaskCompleted hook
		if hookRegistry != nil {
			openspecHooks := hookRegistry.GetOpenSpecHooks()
			if err := hookRegistry.Executor().RunOpenSpecHook(ctx, openspecHooks.OnTaskCompleted, "task_completed", map[string]string{
				"CHANGE_ID": input.ChangeID,
				"TASK_NUM":  input.TaskNum,
			}); err != nil {
				// Log but don't fail the operation
				fmt.Printf("Warning: onTaskCompleted hook failed: %v\n", err)
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("# Task Marked Complete\n\n✅ Task %s/%s has been marked as done.\n\nRun `registry_sync` to update the BoltDB cache.", input.ChangeID, input.TaskNum)}},
		}, RegistryMarkDoneResponse{Success: true, Message: "Task marked complete"}, nil
	}
}

type RegistryStartTaskInput struct {
	ChangeID string `json:"change_id"`
	TaskNum  string `json:"task_num"`
	Assignee string `json:"assignee,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

type RegistryStartTaskResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	StartedAt string `json:"started_at"`
}

func registerRegistryStartTask(s *mcp.Server, toolRegistry *ToolRegistry, store *spec.BoltStore, hookRegistry *hooks.Registry) {
	tool := &mcp.Tool{
		Name:        "registry_start_task",
		Description: "Set a task as in-progress in runtime state (BoltDB only, not persisted to markdown)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"change_id": map[string]any{
					"type":        "string",
					"description": "The change ID (e.g., 'add-user-auth')",
				},
				"task_num": map[string]any{
					"type":        "string",
					"description": "The task number (e.g., '1.2')",
				},
				"assignee": map[string]any{
					"type":        "string",
					"description": "Optional: who is working on the task",
				},
				"notes": map[string]any{
					"type":        "string",
					"description": "Optional: any notes about the work",
				},
			},
			"required": []string{"change_id", "task_num"},
		},
	}

	mcp.AddTool(s, tool, registryStartTaskHandler(store, hookRegistry))
	toolRegistry.Register("registry_start_task", tool.Description, "registry")
}

func registryStartTaskHandler(store *spec.BoltStore, hookRegistry *hooks.Registry) func(ctx context.Context, req *mcp.CallToolRequest, input RegistryStartTaskInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input RegistryStartTaskInput) (*mcp.CallToolResult, any, error) {
		task, err := store.GetTask(input.ChangeID, input.TaskNum)
		if err != nil {
			return nil, nil, fmt.Errorf("get task: %w", err)
		}
		if task == nil {
			return nil, nil, fmt.Errorf("task %s/%s not found", input.ChangeID, input.TaskNum)
		}

		if task.Status == spec.TaskCompleted {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("# Start Task\n\nTask %s/%s is already completed.", input.ChangeID, input.TaskNum)}},
			}, RegistryStartTaskResponse{Success: false, Message: "Task already completed"}, nil
		}

		state := &spec.RuntimeState{
			TaskID:    task.ID,
			Status:    spec.TaskInProgress,
			Assignee:  input.Assignee,
			StartedAt: time.Now(),
		}

		if input.Notes != "" {
			state.Notes = []string{input.Notes}
		}

		if err := store.PutRuntimeState(state); err != nil {
			return nil, nil, fmt.Errorf("put runtime state: %w", err)
		}

		// Trigger onTaskStarted hook
		if hookRegistry != nil {
			openspecHooks := hookRegistry.GetOpenSpecHooks()
			if err := hookRegistry.Executor().RunOpenSpecHook(ctx, openspecHooks.OnTaskStarted, "task_started", map[string]string{
				"CHANGE_ID": input.ChangeID,
				"TASK_NUM":  input.TaskNum,
			}); err != nil {
				// Log but don't fail the operation
				fmt.Printf("Warning: onTaskStarted hook failed: %v\n", err)
			}
		}

		var content strings.Builder
		content.WriteString("# Task Started\n\n")
		content.WriteString(fmt.Sprintf("**Task**: %s/%s\n\n", input.ChangeID, input.TaskNum))
		content.WriteString("**Status**: In Progress 🔄\n\n")
		content.WriteString(fmt.Sprintf("**Started**: %s\n\n", state.StartedAt.Format("2006-01-02 15:04:05")))

		if input.Assignee != "" {
			content.WriteString(fmt.Sprintf("**Assignee**: %s\n\n", input.Assignee))
		}

		if input.Notes != "" {
			content.WriteString(fmt.Sprintf("**Notes**: %s\n\n", input.Notes))
		}

		content.WriteString("*Runtime state stored in BoltDB. Not persisted to markdown.*\n")

		return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: content.String()}},
			}, RegistryStartTaskResponse{
				Success:   true,
				Message:   "Task started",
				StartedAt: state.StartedAt.Format(time.RFC3339),
			}, nil
	}
}

type RegistrySyncInput struct{}

type RegistrySyncResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ChangesSync int    `json:"changes_synced"`
	TasksSync   int    `json:"tasks_synced"`
	Duration    string `json:"duration"`
}

func registerRegistrySync(s *mcp.Server, toolRegistry *ToolRegistry, store *spec.BoltStore) {
	tool := &mcp.Tool{
		Name:        "registry_sync",
		Description: "Force a full rebuild of BoltDB cache from markdown files",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}

	mcp.AddTool(s, tool, registrySyncHandler(store))
	toolRegistry.Register("registry_sync", tool.Description, "registry")
}

func registrySyncHandler(store *spec.BoltStore) func(ctx context.Context, req *mcp.CallToolRequest, input RegistrySyncInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input RegistrySyncInput) (*mcp.CallToolResult, any, error) {
		start := time.Now()

		if err := store.RebuildFromMarkdown(""); err != nil {
			return nil, nil, fmt.Errorf("rebuild from markdown: %w", err)
		}

		duration := time.Since(start)

		changes, err := store.ListAllChanges()
		if err != nil {
			return nil, nil, fmt.Errorf("list changes: %w", err)
		}

		tasks, err := store.ListTasks("", "")
		if err != nil {
			return nil, nil, fmt.Errorf("list tasks: %w", err)
		}

		var content strings.Builder
		content.WriteString("# Registry Sync Complete\n\n")
		content.WriteString("✅ BoltDB cache rebuilt from markdown files.\n\n")
		content.WriteString(fmt.Sprintf("**Changes synced**: %d\n\n", len(changes)))
		content.WriteString(fmt.Sprintf("**Tasks synced**: %d\n\n", len(tasks)))
		content.WriteString(fmt.Sprintf("**Duration**: %v\n\n", duration))

		return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: content.String()}},
			}, RegistrySyncResponse{
				Success:     true,
				Message:     "Sync complete",
				ChangesSync: len(changes),
				TasksSync:   len(tasks),
				Duration:    duration.String(),
			}, nil
	}
}
