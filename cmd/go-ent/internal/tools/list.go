package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/spec"
)

type SpecListInput struct {
	Type   string `json:"type"`
	Status string `json:"status,omitempty"`
}

func registerList(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "go_ent_spec_list",
		Description: "List specs, changes, or tasks",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"type": map[string]any{
					"type":        "string",
					"description": "Type of items to list: 'spec', 'change', or 'task'",
					"enum":        []string{"spec", "change", "task"},
				},
				"status": map[string]any{
					"type":        "string",
					"description": "Filter by status (for changes: 'active' or 'archived')",
				},
			},
			"required": []string{"type"},
		},
	}

	mcp.AddTool(s, tool, specListHandler)
}

func specListHandler(ctx context.Context, req *mcp.CallToolRequest, input SpecListInput) (*mcp.CallToolResult, any, error) {
	if input.Type == "" {
		return nil, nil, fmt.Errorf("type is required")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error getting current directory: %v", err)}},
		}, nil, nil
	}

	store := spec.NewStore(cwd)

	exists, err := store.Exists()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error checking .spec folder: %v", err)}},
		}, nil, nil
	}

	if !exists {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No openspec folder found. Run go_ent_spec_init first."}},
		}, nil, nil
	}

	var items []spec.ListItem

	switch input.Type {
	case "specs":
		items, err = store.ListSpecs()
	case "changes":
		items, err = store.ListChanges(input.Status)
	case "tasks":
		items, err = store.ListTasks()
	default:
		return nil, nil, fmt.Errorf("invalid type: %s. Must be specs, changes, or tasks", input.Type)
	}

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error listing %s: %v", input.Type, err)}},
		}, nil, nil
	}

	if len(items) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("No %s found", input.Type)}},
		}, nil, nil
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error formatting results: %v", err)}},
		}, nil, nil
	}

	msg := fmt.Sprintf("Found %d %s:\n\n```json\n%s\n```", len(items), input.Type, string(data))

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}
