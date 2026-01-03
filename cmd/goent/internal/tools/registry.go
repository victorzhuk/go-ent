package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/cmd/goent/internal/spec"
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
		Name:        "goent_registry_list",
		Description: "List tasks from the OpenSpec registry with optional filters. Shows aggregated view across all changes.",
	}
	mcp.AddTool(s, listTool, registryListHandler)

	nextTool := &mcp.Tool{
		Name:        "goent_registry_next",
		Description: "Get the next recommended task(s) based on priority, dependencies, and status. Returns unblocked, highest priority pending tasks.",
	}
	mcp.AddTool(s, nextTool, registryNextHandler)

	updateTool := &mcp.Tool{
		Name:        "goent_registry_update",
		Description: "Update task status, priority, or assignment in the registry. Automatically updates blocked_by for dependent tasks.",
	}
	mcp.AddTool(s, updateTool, registryUpdateHandler)

	depsTool := &mcp.Tool{
		Name:        "goent_registry_deps",
		Description: "Manage task dependencies. Supports cross-change dependencies. Detects cycles.",
	}
	mcp.AddTool(s, depsTool, registryDepsHandler)

	syncTool := &mcp.Tool{
		Name:        "goent_registry_sync",
		Description: "Synchronize registry from tasks.md files. Rebuilds registry from source change proposals.",
	}
	mcp.AddTool(s, syncTool, registrySyncHandler)

	initTool := &mcp.Tool{
		Name:        "goent_registry_init",
		Description: "Initialize an empty registry.yaml file. Use this before first sync or to reset the registry.",
	}
	mcp.AddTool(s, initTool, registryInitHandler)
}

func registryListHandler(ctx context.Context, req *mcp.CallToolRequest, input RegistryListInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	store := spec.NewStore(input.Path)
	regStore := spec.NewRegistryStore(store)

	if !regStore.Exists() {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Registry not found. Run goent_registry_sync to initialize from tasks.md files."}},
		}, nil, nil
	}

	filter := spec.TaskFilter{
		ChangeID:  input.ChangeID,
		Status:    spec.RegistryTaskStatus(input.Status),
		Priority:  spec.TaskPriority(input.Priority),
		Assignee:  input.Assignee,
		Unblocked: input.Unblocked,
		Limit:     input.Limit,
	}

	tasks, err := regStore.ListTasks(filter)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error listing tasks: %v", err)}},
		}, nil, nil
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

	store := spec.NewStore(input.Path)
	regStore := spec.NewRegistryStore(store)

	if !regStore.Exists() {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Registry not found. Run goent_registry_sync first."}},
		}, nil, nil
	}

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

	taskID, err := parseTaskID(input.TaskID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid task_id format: %v", err)}},
		}, nil, nil
	}

	store := spec.NewStore(input.Path)
	regStore := spec.NewRegistryStore(store)

	if !regStore.Exists() {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Registry not found. Run goent_registry_sync first."}},
		}, nil, nil
	}

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

	taskID, err := parseTaskID(input.TaskID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid task_id format: %v", err)}},
		}, nil, nil
	}

	store := spec.NewStore(input.Path)
	regStore := spec.NewRegistryStore(store)

	if !regStore.Exists() {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Registry not found. Run goent_registry_sync first."}},
		}, nil, nil
	}

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

	store := spec.NewStore(input.Path)
	regStore := spec.NewRegistryStore(store)

	result, err := regStore.RebuildFromSource()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error syncing registry: %v", err)}},
		}, nil, nil
	}

	msg := "✅ Registry synced from tasks.md files\n\n"
	data, _ := json.MarshalIndent(result, "", "  ")
	msg += string(data)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}

func registryInitHandler(ctx context.Context, req *mcp.CallToolRequest, input RegistryInitInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	store := spec.NewStore(input.Path)
	regStore := spec.NewRegistryStore(store)

	if regStore.Exists() {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Registry already exists. Use goent_registry_sync to update it."}},
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
