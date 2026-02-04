package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/hooks"
	"github.com/victorzhuk/go-ent/internal/openspec"
)

type OpenSpecNewChangeInput struct {
	Name string `json:"name"` // Change name
}

func registerOpenSpecNewChange(s *mcp.Server, toolRegistry *ToolRegistry, client *openspec.Client, hookRegistry *hooks.Registry) {
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

	mcp.AddTool(s, tool, openspecNewChangeHandler(client, hookRegistry))
	toolRegistry.Register("openspec_new_change", tool.Description, "openspec")
}

func openspecNewChangeHandler(client *openspec.Client, hookRegistry *hooks.Registry) func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecNewChangeInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecNewChangeInput) (*mcp.CallToolResult, any, error) {
		if err := client.NewChange(ctx, input.Name); err != nil {
			return nil, nil, fmt.Errorf("create change %s: %w", input.Name, err)
		}

		// Trigger onChangeCreated hook
		if hookRegistry != nil {
			openspecHooks := hookRegistry.GetOpenSpecHooks()
			if err := hookRegistry.Executor().RunOpenSpecHook(ctx, openspecHooks.OnChangeCreated, "change_created", map[string]string{
				"CHANGE_ID": input.Name,
			}); err != nil {
				// Log but don't fail the operation
				fmt.Printf("Warning: onChangeCreated hook failed: %v\n", err)
			}
		}

		msg := fmt.Sprintf("✓ Created new change: %s\n\nNext steps:\n1. Edit openspec/changes/%s/proposal.md\n2. Create tasks in openspec/changes/%s/tasks.md\n3. Use openspec_status to track progress",
			input.Name, input.Name, input.Name)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, map[string]string{"name": input.Name, "status": "created"}, nil
	}
}
