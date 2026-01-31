package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/openspec"
)

type OpenSpecNewChangeInput struct {
	Name string `json:"name"` // Change name
}

func registerOpenSpecNewChange(s *mcp.Server, client *openspec.Client) {
	tool := &mcp.Tool{
		Name:        "openspec_new_change",
		Description: "Create a new OpenSpec change proposal",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Name for the new change (kebab-case recommended)",
				},
			},
			"required": []string{"name"},
		},
	}

	mcp.AddTool(s, tool, openspecNewChangeHandler(client))
}

func openspecNewChangeHandler(client *openspec.Client) func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecNewChangeInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecNewChangeInput) (*mcp.CallToolResult, any, error) {
		if err := client.NewChange(ctx, input.Name); err != nil {
			return nil, nil, fmt.Errorf("create change %s: %w", input.Name, err)
		}

		msg := fmt.Sprintf("✓ Created new change: %s\n\nNext steps:\n1. Edit openspec/changes/%s/proposal.md\n2. Create tasks in openspec/changes/%s/tasks.md\n3. Use openspec_status to track progress",
			input.Name, input.Name, input.Name)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, map[string]string{"name": input.Name, "status": "created"}, nil
	}
}
