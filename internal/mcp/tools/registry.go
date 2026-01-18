package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/spec"
)

type RegistryListInput struct {
	Path      string `json:"path"`
	ChangeID  string `json:"change_id,omitempty"`
	Status    string `json:"status,omitempty"`
	Priority  string `json:"priority,omitempty"`
	Assignee  string `json:"assignee,omitempty"`
	Unblocked bool   `json:"unblocked,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type RegistryNextInput struct {
	Path     string `json:"path"`
	ChangeID string `json:"change_id,omitempty"`
	Count    int    `json:"count,omitempty"`
}

type RegistryUpdateInput struct {
	Path     string `json:"path"`
	TaskID   string `json:"task_id"`
	Status   string `json:"status,omitempty"`
	Priority string `json:"priority,omitempty"`
	Assignee string `json:"assignee,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

type RegistryDepsInput struct {
	Path      string `json:"path"`
	TaskID    string `json:"task_id"`
	Operation string `json:"operation"`
	DependsOn string `json:"depends_on,omitempty"`
}

type RegistrySyncInput struct {
	Path   string `json:"path"`
	DryRun bool   `json:"dry_run,omitempty"`
	Force  bool   `json:"force,omitempty"`
}

type RegistryInitInput struct {
	Path string `json:"path"`
}

func registerRegistry(s *mcp.Server) {
	listTool := &mcp.Tool{
		Name:        "registry_list",
		Description: "List tasks from the OpenSpec registry with optional filters. Shows aggregated view across all changes.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":      map[string]any{"type": "string", "description": "Path to project directory"},
				"change_id": map[string]any{"type": "string", "description": "Filter by change ID"},
				"status":    map[string]any{"type": "string", "description": "Filter by status (pending, in_progress, completed)"},
				"priority":  map[string]any{"type": "string", "description": "Filter by priority (critical, high, medium, low)"},
				"assignee":  map[string]any{"type": "string", "description": "Filter by assignee"},
				"unblocked": map[string]any{"type": "boolean", "description": "Show only unblocked tasks"},
				"limit":     map[string]any{"type": "integer", "description": "Maximum number of tasks to return"},
			},
			"required": []string{"path"},
		},
	}
	mcp.AddTool(s, listTool, registryListHandler)

	nextTool := &mcp.Tool{
		Name:        "registry_next",
		Description: "Get the next recommended task(s) based on priority, dependencies, and status. Returns unblocked, highest priority pending tasks.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":      map[string]any{"type": "string", "description": "Path to project directory"},
				"change_id": map[string]any{"type": "string", "description": "Filter by change ID"},
				"count":     map[string]any{"type": "integer", "description": "Number of tasks to return", "default": 1},
			},
			"required": []string{"path"},
		},
	}
	mcp.AddTool(s, nextTool, registryNextHandler)

	updateTool := &mcp.Tool{
		Name:        "registry_update",
		Description: "Update task status, priority, or assignment in the registry. Automatically updates blocked_by for dependent tasks.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":     map[string]any{"type": "string", "description": "Path to project directory"},
				"task_id":  map[string]any{"type": "string", "description": "Task ID to update"},
				"status":   map[string]any{"type": "string", "description": "New status", "enum": []string{"pending", "in_progress", "completed"}},
				"priority": map[string]any{"type": "string", "description": "New priority", "enum": []string{"critical", "high", "medium", "low"}},
				"assignee": map[string]any{"type": "string", "description": "New assignee"},
				"notes":    map[string]any{"type": "string", "description": "Additional notes"},
			},
			"required": []string{"path", "task_id"},
		},
	}
	mcp.AddTool(s, updateTool, registryUpdateHandler)

	depsTool := &mcp.Tool{
		Name:        "registry_deps",
		Description: "Manage task dependencies. Supports cross-change dependencies. Detects cycles.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":       map[string]any{"type": "string", "description": "Path to project directory"},
				"task_id":    map[string]any{"type": "string", "description": "Task ID"},
				"operation":  map[string]any{"type": "string", "description": "Operation to perform", "enum": []string{"add", "remove", "list"}},
				"depends_on": map[string]any{"type": "string", "description": "Dependency task ID (required for add/remove)"},
			},
			"required": []string{"path", "task_id", "operation"},
		},
	}
	mcp.AddTool(s, depsTool, registryDepsHandler)

	syncTool := &mcp.Tool{
		Name:        "registry_sync",
		Description: "Synchronize registry from tasks.md files. Rebuilds registry from source change proposals.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":    map[string]any{"type": "string", "description": "Path to project directory"},
				"dry_run": map[string]any{"type": "boolean", "description": "Preview changes without saving", "default": false},
				"force":   map[string]any{"type": "boolean", "description": "Force sync even if registry has local changes", "default": false},
			},
			"required": []string{"path"},
		},
	}
	mcp.AddTool(s, syncTool, registrySyncHandler)

	initTool := &mcp.Tool{
		Name:        "registry_init",
		Description: "Initialize an empty registry.yaml file. Use this before first sync or to reset the registry.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string", "description": "Path to project directory"},
			},
			"required": []string{"path"},
		},
	}
	mcp.AddTool(s, initTool, registryInitHandler)
}

func registryListHandler(ctx context.Context, req *mcp.CallToolRequest, input RegistryListInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	// Check if registry.db exists before creating BoltDB
	registryPath := filepath.Join(input.Path, "openspec", "registry.db")
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		// Fallback to parsing tasks.md
		tasks, err := parseTasksFromSource(input.Path, input.ChangeID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error parsing tasks.md: %v", err)}},
			}, nil, nil
		}

		filtered := filterTasks(tasks, spec.TaskFilter{
			ChangeID:  input.ChangeID,
			Status:    spec.RegistryTaskStatus(input.Status),
			Priority:  spec.TaskPriority(input.Priority),
			Assignee:  input.Assignee,
			Unblocked: input.Unblocked,
		})

		if input.Limit > 0 && len(filtered) > input.Limit {
			filtered = filtered[:input.Limit]
		}

		summary := buildSummary(filtered)

		output := map[string]interface{}{
			"total":    len(tasks),
			"filtered": len(filtered),
			"tasks":    filtered,
			"note":     "Showing tasks from tasks.md files (registry.db not found)",
		}

		if summary != nil {
			output["summary"] = summary
		}

		data, _ := json.MarshalIndent(output, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}

	// Use BoltDB registry
	store := spec.NewStore(input.Path)
	regStore, err := spec.NewRegistryStore(store)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating registry store: %v", err)}},
		}, nil, nil
	}
	defer func() { _ = regStore.Close() }()

	filter := spec.TaskFilter{
		ChangeID:  input.ChangeID,
		Status:    spec.RegistryTaskStatus(input.Status),
		Priority:  spec.TaskPriority(input.Priority),
		Assignee:  input.Assignee,
		Unblocked: input.Unblocked,
	}

	tasks, err := regStore.ListTasks(filter)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error listing tasks: %v", err)}},
		}, nil, nil
	}

	if input.Limit > 0 && len(tasks) > input.Limit {
		tasks = tasks[:input.Limit]
	}

	stats, _ := regStore.Stats()

	output := map[string]interface{}{
		"total":    len(tasks),
		"filtered": len(tasks),
		"tasks":    tasks,
	}

	if stats != nil {
		output["summary"] = map[string]interface{}{
			"by_status":   stats.ByStatus,
			"by_priority": stats.ByPriority,
			"by_change":   stats.ByChange,
		}
	}

	data, _ := json.MarshalIndent(output, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func registryNextHandler(ctx context.Context, req *mcp.CallToolRequest, input RegistryNextInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	// Check if registry.db exists before creating BoltDB
	registryPath := filepath.Join(input.Path, "openspec", "registry.db")
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		// Fallback to parsing tasks.md
		tasks, err := parseTasksFromSource(input.Path, input.ChangeID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error parsing tasks.md: %v", err)}},
			}, nil, nil
		}

		count := input.Count
		if count <= 0 {
			count = 1
		}

		next := findNextTasks(tasks, count)

		data, _ := json.MarshalIndent(next, "", "  ")
		text := string(data)
		if len(text) > 0 {
			text += "\n\n"
		}
		text += "Note: Showing tasks from tasks.md files (registry.db not found)"

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	}

	// Use BoltDB registry
	store := spec.NewStore(input.Path)
	regStore, err := spec.NewRegistryStore(store)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating registry store: %v", err)}},
		}, nil, nil
	}
	defer func() { _ = regStore.Close() }()

	count := input.Count
	if count <= 0 {
		count = 1
	}

	result, err := regStore.NextTask(count)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error finding next task: %v", err)}},
		}, nil, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func registryUpdateHandler(ctx context.Context, req *mcp.CallToolRequest, input RegistryUpdateInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}
	if input.TaskID == "" {
		return nil, nil, fmt.Errorf("task_id is required")
	}

	// Check if registry.db exists before creating BoltDB
	registryPath := filepath.Join(input.Path, "openspec", "registry.db")
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Registry not found. Run registry_sync first."}},
		}, nil, nil
	}

	taskID, err := parseTaskID(input.TaskID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid task_id format: %v", err)}},
		}, nil, nil
	}

	store := spec.NewStore(input.Path)
	regStore, err := spec.NewRegistryStore(store)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating registry store: %v", err)}},
		}, nil, nil
	}
	defer func() { _ = regStore.Close() }()

	updates := spec.TaskUpdate{}
	if input.Status != "" {
		status := spec.RegistryTaskStatus(input.Status)
		updates.Status = &status
	}
	if input.Priority != "" {
		priority := spec.TaskPriority(input.Priority)
		updates.Priority = &priority
	}
	if input.Assignee != "" {
		updates.Assignee = &input.Assignee
	}
	if input.Notes != "" {
		updates.Notes = &input.Notes
	}

	if err := regStore.UpdateTask(taskID, updates); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error updating task: %v", err)}},
		}, nil, nil
	}

	task, _ := regStore.GetTask(taskID)
	msg := fmt.Sprintf("✅ Updated task %s\n\n", taskID.String())
	if task != nil {
		data, _ := json.MarshalIndent(task, "", "  ")
		msg += string(data)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}

func registryDepsHandler(ctx context.Context, req *mcp.CallToolRequest, input RegistryDepsInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}
	if input.TaskID == "" {
		return nil, nil, fmt.Errorf("task_id is required")
	}

	// Check if registry.db exists before creating BoltDB
	registryPath := filepath.Join(input.Path, "openspec", "registry.db")
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Registry not found. Run registry_sync first."}},
		}, nil, nil
	}

	taskID, err := parseTaskID(input.TaskID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid task_id format: %v", err)}},
		}, nil, nil
	}

	store := spec.NewStore(input.Path)
	regStore, err := spec.NewRegistryStore(store)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating registry store: %v", err)}},
		}, nil, nil
	}
	defer func() { _ = regStore.Close() }()

	switch input.Operation {
	case "show":
		graph, err := regStore.GetDependencyGraph(taskID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error getting dependency graph: %v", err)}},
			}, nil, nil
		}
		data, _ := json.MarshalIndent(graph, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil

	case "add":
		if input.DependsOn == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "depends_on is required for add operation"}},
			}, nil, nil
		}
		depID, err := parseTaskID(input.DependsOn)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid depends_on format: %v", err)}},
			}, nil, nil
		}
		if err := regStore.AddDependency(taskID, depID); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error adding dependency: %v", err)}},
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("✅ Added dependency: %s depends on %s", taskID.String(), depID.String())}},
		}, nil, nil

	case "remove":
		if input.DependsOn == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "depends_on is required for remove operation"}},
			}, nil, nil
		}
		depID, err := parseTaskID(input.DependsOn)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid depends_on format: %v", err)}},
			}, nil, nil
		}
		if err := regStore.RemoveDependency(taskID, depID); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error removing dependency: %v", err)}},
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("✅ Removed dependency: %s no longer depends on %s", taskID.String(), depID.String())}},
		}, nil, nil

	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid operation: %s. Must be show, add, or remove", input.Operation)}},
		}, nil, nil
	}
}

func registrySyncHandler(ctx context.Context, req *mcp.CallToolRequest, input RegistrySyncInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	// Check if registry.db exists before creating BoltDB
	registryPath := filepath.Join(input.Path, "openspec", "registry.db")
	if _, err := os.Stat(registryPath); os.IsNotExist(err) && !input.Force {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Registry not found. Run registry_init first to create the registry.db file."}},
		}, nil, nil
	}

	store := spec.NewStore(input.Path)
	regStore, err := spec.NewRegistryStore(store)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating registry store: %v", err)}},
		}, nil, nil
	}
	defer func() { _ = regStore.Close() }()

	result, err := regStore.RebuildFromSource()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error syncing registry: %v", err)}},
		}, nil, nil
	}

	msg := "✅ Registry synced from tasks.md files\n\n"
	data, _ := json.MarshalIndent(result, "", "  ")
	msg += string(data)

	msg += "\n\n💡 Tip: Run state_sync to generate state.md files from the updated registry"

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}

func registryInitHandler(ctx context.Context, req *mcp.CallToolRequest, input RegistryInitInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	store := spec.NewStore(input.Path)
	regStore, err := spec.NewRegistryStore(store)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating registry store: %v", err)}},
		}, nil, nil
	}
	defer func() { _ = regStore.Close() }()

	if regStore.Exists() {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Registry already exists. Use registry_sync to update it."}},
		}, nil, nil
	}

	if err := regStore.Init(); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error initializing registry: %v", err)}},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("✅ Initialized empty registry at %s", store.RegistryPath())}},
	}, nil, nil
}

func parseTaskID(s string) (spec.TaskID, error) {
	parts := splitTaskID(s)
	if len(parts) != 2 {
		return spec.TaskID{}, fmt.Errorf("task_id must be in format change-id/task-num (e.g. add-auth/1.1)")
	}
	return spec.TaskID{
		ChangeID: parts[0],
		TaskNum:  parts[1],
	}, nil
}

func splitTaskID(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func parseTasksFromSource(path string, changeID string) ([]spec.RegistryTask, error) {
	changesPath := filepath.Join(path, "openspec", "changes")
	entries, err := os.ReadDir(changesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []spec.RegistryTask{}, nil
		}
		return nil, fmt.Errorf("read changes dir: %w", err)
	}

	store := spec.NewStore(path)
	var allTasks []spec.RegistryTask

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "archive" {
			continue
		}

		id := entry.Name()
		if changeID != "" && id != changeID {
			continue
		}

		tasksPath := filepath.Join(changesPath, id, "tasks.md")
		if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
			continue
		}

		stateStore := spec.NewStateStore(store, nil)

		tasks, err := stateStore.ParseTasksWithDependencies(id, tasksPath)
		if err != nil {
			continue
		}

		allTasks = append(allTasks, tasks...)
	}

	return allTasks, nil
}

func filterTasks(tasks []spec.RegistryTask, filter spec.TaskFilter) []spec.RegistryTask {
	filtered := make([]spec.RegistryTask, 0)

	for _, task := range tasks {
		if filter.ChangeID != "" && task.ID.ChangeID != filter.ChangeID {
			continue
		}
		if filter.Status != "" && task.Status != filter.Status {
			continue
		}
		if filter.Priority != "" && task.Priority != filter.Priority {
			continue
		}
		if filter.Assignee != "" && task.Assignee != filter.Assignee {
			continue
		}
		if filter.Unblocked {
			hasUnmet := false
			for _, dep := range task.DependsOn {
				depMet := false
				for _, t := range tasks {
					if t.ID == dep && t.Status == spec.RegStatusCompleted {
						depMet = true
						break
					}
				}
				if !depMet {
					hasUnmet = true
					break
				}
			}
			if hasUnmet {
				continue
			}
		}
		filtered = append(filtered, task)
	}

	return filtered
}

func buildSummary(tasks []spec.RegistryTask) map[string]interface{} {
	byStatus := make(map[string]int)
	byPriority := make(map[string]int)
	byChange := make(map[string]int)

	for _, task := range tasks {
		byStatus[string(task.Status)]++
		byPriority[string(task.Priority)]++
		byChange[task.ID.ChangeID]++
	}

	return map[string]interface{}{
		"by_status":   byStatus,
		"by_priority": byPriority,
		"by_change":   byChange,
	}
}

func findNextTasks(tasks []spec.RegistryTask, count int) *spec.NextTaskResult {
	unblocked := filterTasks(tasks, spec.TaskFilter{
		Status:    spec.RegStatusPending,
		Unblocked: true,
	})

	sort.Slice(unblocked, func(i, j int) bool {
		if unblocked[i].Priority != unblocked[j].Priority {
			return priorityValue(unblocked[i].Priority) < priorityValue(unblocked[j].Priority)
		}
		return unblocked[i].ID.String() < unblocked[j].ID.String()
	})

	blockedCount := 0
	for _, task := range tasks {
		hasUnmet := false
		for _, dep := range task.DependsOn {
			depMet := false
			for _, t := range tasks {
				if t.ID == dep && t.Status == spec.RegStatusCompleted {
					depMet = true
					break
				}
			}
			if !depMet {
				hasUnmet = true
				break
			}
		}
		if hasUnmet {
			blockedCount++
		}
	}

	result := &spec.NextTaskResult{
		BlockedCount: blockedCount,
	}

	if len(unblocked) > 0 {
		result.Recommended = &unblocked[0]
		result.Reason = fmt.Sprintf("Highest priority (%s) unblocked task", unblocked[0].Priority)
		if len(unblocked[0].DependsOn) > 0 {
			result.Reason += fmt.Sprintf(". Dependencies completed: %d", len(unblocked[0].DependsOn))
		}

		if len(unblocked) > 1 {
			limit := count
			if len(unblocked)-1 < count {
				limit = len(unblocked) - 1
			}
			result.Alternatives = unblocked[1 : limit+1]
		}
	}

	return result
}

func priorityValue(p spec.TaskPriority) int {
	switch p {
	case spec.PriorityCritical:
		return 1
	case spec.PriorityHigh:
		return 2
	case spec.PriorityMedium:
		return 3
	case spec.PriorityLow:
		return 4
	case spec.PriorityBacklog:
		return 5
	default:
		return 6
	}
}
