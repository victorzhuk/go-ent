package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/openspec"
)

type OpenSpecArchiveInput struct {
	Change string `json:"change"` // Change name to archive
}

func registerOpenSpecArchive(s *mcp.Server, toolRegistry *ToolRegistry, client *openspec.Client) {
	tool := &mcp.Tool{
		Name:        "openspec_archive",
		Description: "Archive a completed change and update main specs",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"change": map[string]any{
					"type":        "string",
					"description": "Name of the completed change to archive",
				},
			},
			"required": []string{"change"},
		},
	}

	mcp.AddTool(s, tool, openspecArchiveHandler(client))
	toolRegistry.Register("openspec_archive", tool.Description, "openspec")
}

func openspecArchiveHandler(client *openspec.Client) func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecArchiveInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecArchiveInput) (*mcp.CallToolResult, any, error) {
		if err := client.Archive(ctx, input.Change); err != nil {
			return nil, nil, fmt.Errorf("archive %s: %w", input.Change, err)
		}

		msg := fmt.Sprintf("✓ Archived change: %s\n\nThe change has been moved to archive and main specs have been updated with delta specs.",
			input.Change)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, map[string]string{"change": input.Change, "status": "archived"}, nil
	}
}
