package tools

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/spec"
)

type StateSyncInput struct {
	Path   string `json:"path"`
	DryRun bool   `json:"dry_run,omitempty"`
}

type StateShowInput struct {
	Path     string `json:"path"`
	ChangeID string `json:"change_id,omitempty"`
}

func registerStateSync(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "state_sync",
		Description: "Sync tasks.md to BoltDB registry and generate state.md files",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Path to project directory",
				},
				"dry_run": map[string]any{
					"type":        "boolean",
					"description": "Preview changes without saving",
					"default":     false,
				},
			},
			"required": []string{"path"},
		},
	}
	mcp.AddTool(s, tool, stateSyncHandler)
}

func registerStateShow(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "state_show",
		Description: "Display current state (quick view for /status)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Path to project directory",
				},
				"change_id": map[string]any{
					"type":        "string",
					"description": "Optional change ID to show specific change state",
				},
			},
			"required": []string{"path"},
		},
	}
	mcp.AddTool(s, tool, stateShowHandler)
}

func stateSyncHandler(ctx context.Context, req *mcp.CallToolRequest, input StateSyncInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: path is required"}},
		}, nil, nil
	}

	store := spec.NewStore(input.Path)

	boltPath := filepath.Join(input.Path, "openspec", "registry.db")
	bolt, err := spec.NewBoltStore(boltPath)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating bolt store: %v", err)}},
		}, nil, nil
	}
	defer func() { _ = bolt.Close() }()

	stateStore := spec.NewStateStore(store, bolt)

	if err := stateStore.SyncFromTasksMd(); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error syncing from tasks.md: %v", err)}},
		}, nil, nil
	}

	changes, err := bolt.ListChanges()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error listing changes: %v", err)}},
		}, nil, nil
	}

	for _, change := range changes {
		changeStatePath := filepath.Join(input.Path, "openspec", "changes", change.ID, "state.md")
		if err := stateStore.WriteChangeStateMd(change.ID, changeStatePath); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error writing state for %s: %v", change.ID, err)}},
			}, nil, nil
		}
	}

	rootStatePath := filepath.Join(input.Path, "openspec", "state.md")
	if err := stateStore.WriteRootStateMd(rootStatePath); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error writing root state: %v", err)}},
		}, nil, nil
	}

	result := fmt.Sprintf("✓ Synced %d changes\n", len(changes))
	result += "✓ Generated state.md files\n"
	result += "✓ BoltDB registry updated\n"

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil, nil
}

func stateShowHandler(ctx context.Context, req *mcp.CallToolRequest, input StateShowInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: path is required"}},
		}, nil, nil
	}

	store := spec.NewStore(input.Path)

	boltPath := filepath.Join(input.Path, "openspec", "registry.db")
	bolt, err := spec.NewBoltStore(boltPath)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating bolt store: %v", err)}},
		}, nil, nil
	}
	defer func() { _ = bolt.Close() }()

	stateStore := spec.NewStateStore(store, bolt)

	if input.ChangeID != "" {
		state, err := stateStore.GenerateChangeState(input.ChangeID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error generating state: %v", err)}},
			}, nil, nil
		}

		result := fmt.Sprintf("# %s\n\n", input.ChangeID)
		result += fmt.Sprintf("Progress: %d/%d (%d%%)\n\n",
			state.Progress.Completed, state.Progress.Total, state.Progress.Percent)

		if state.CurrentTask != nil {
			result += fmt.Sprintf("Current: T%s - %s\n\n",
				state.CurrentTask.ID.TaskNum, state.CurrentTask.Content)
		} else {
			result += "Current: None\n\n"
		}

		if len(state.Blockers) > 0 {
			result += fmt.Sprintf("Blockers: %d\n", len(state.Blockers))
		} else {
			result += "Blockers: None\n"
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	}

	rootState, err := stateStore.GenerateRootState()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error generating root state: %v", err)}},
		}, nil, nil
	}

	result := "# OpenSpec State\n\n"
	result += fmt.Sprintf("%d active changes\n\n", len(rootState.ActiveChanges))

	for _, change := range rootState.ActiveChanges {
		percent := 0
		if change.Total > 0 {
			percent = (change.Completed * 100) / change.Total
		}
		result += fmt.Sprintf("- %s: %d%% (%d/%d)\n", change.ID, percent, change.Completed, change.Total)
	}

	if len(rootState.RecommendedTasks) > 0 {
		result += "\n## Next Recommended\n"
		for i, task := range rootState.RecommendedTasks {
			result += fmt.Sprintf("%d. %s - %s\n", i+1, task.ID.String(), task.Content)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil, nil
}
