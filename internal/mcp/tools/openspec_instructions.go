package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/openspec"
)

type OpenSpecInstructionsInput struct {
	Change   string `json:"change"`   // Change name (optional)
	Artifact string `json:"artifact"` // Artifact name like "proposal", "tasks" (optional)
}

func registerOpenSpecInstructions(s *mcp.Server, toolRegistry *ToolRegistry, client *openspec.Client) {
	tool := &mcp.Tool{
		Name:        "openspec_instructions",
		Description: "Get enriched instructions for creating an artifact or applying tasks",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"change": map[string]any{
					"type":        "string",
					"description": "Name of the change (optional)",
				},
				"artifact": map[string]any{
					"type":        "string",
					"description": "Artifact name like 'proposal', 'tasks', etc. (optional)",
				},
			},
		},
	}

	mcp.AddTool(s, tool, openspecInstructionsHandler(client))
	toolRegistry.Register("openspec_instructions", tool.Description, "openspec")
}

func openspecInstructionsHandler(client *openspec.Client) func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecInstructionsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecInstructionsInput) (*mcp.CallToolResult, any, error) {
		instructions, err := client.Instructions(ctx, input.Change, input.Artifact)
		if err != nil {
			return nil, nil, fmt.Errorf("instructions: %w", err)
		}

		return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: instructions}},
			}, map[string]string{
				"change":   input.Change,
				"artifact": input.Artifact,
			}, nil
	}
}
