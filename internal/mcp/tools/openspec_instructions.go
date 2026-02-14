package tools

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/openspec"
	"github.com/victorzhuk/go-ent/internal/workspace"
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

		wsCtx, err := workspaceContext()
		if err != nil {
			slog.Warn("workspace context unavailable", "error", err)
		}
		if wsCtx != "" {
			instructions = instructions + "\n\n" + wsCtx
		}

		return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: instructions}},
			}, map[string]string{
				"change":   input.Change,
				"artifact": input.Artifact,
			}, nil
	}
}

func workspaceContext() (string, error) {
	ws, err := workspace.DetectAndResolve(".")
	if err != nil {
		return "", fmt.Errorf("detect workspace: %w", err)
	}
	if ws == nil {
		return "", nil
	}

	prompt, err := workspace.GenerateContextPrompt(ws)
	if err != nil {
		return "", fmt.Errorf("generate workspace context: %w", err)
	}

	return prompt, nil
}
