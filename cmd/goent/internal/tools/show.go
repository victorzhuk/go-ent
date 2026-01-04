package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/spec"
)

type SpecShowInput struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func registerShow(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "goent_spec_show",
		Description: "Show detailed content of a spec, change, or task",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"type": map[string]any{
					"type":        "string",
					"description": "Type of item to show: 'spec', 'change', or 'task'",
					"enum":        []string{"spec", "change", "task"},
				},
				"id": map[string]any{
					"type":        "string",
					"description": "Identifier of the item to show",
				},
			},
			"required": []string{"type", "id"},
		},
	}

	mcp.AddTool(s, tool, specShowHandler)
}

func specShowHandler(ctx context.Context, req *mcp.CallToolRequest, input SpecShowInput) (*mcp.CallToolResult, any, error) {
	if input.Type == "" || input.ID == "" {
		return nil, nil, fmt.Errorf("type and id are required")
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
			Content: []mcp.Content{&mcp.TextContent{Text: "No openspec folder found. Run goent_spec_init first."}},
		}, nil, nil
	}

	var path string
	switch input.Type {
	case "spec":
		path = fmt.Sprintf("specs/%s/spec.md", input.ID)
	case "change":
		path = fmt.Sprintf("changes/%s/proposal.md", input.ID)
	case "task":
		path = fmt.Sprintf("tasks/%s.md", input.ID)
	default:
		return nil, nil, fmt.Errorf("invalid type: %s. Must be spec, change, or task", input.Type)
	}

	content, err := store.ReadFile(path)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error reading %s: %v", path, err)}},
		}, nil, nil
	}

	msg := fmt.Sprintf("# %s: %s\n\n```markdown\n%s\n```", input.Type, input.ID, content)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}
