package tools

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/hooks"
	"github.com/victorzhuk/go-ent/internal/openspec"
)

type OpenSpecArchiveInput struct {
	Change string `json:"change"` // Change name to archive
}

func registerOpenSpecArchive(s *mcp.Server, toolRegistry *ToolRegistry, client *openspec.Client, hookRegistry *hooks.Registry) {
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

	mcp.AddTool(s, tool, openspecArchiveHandler(client, hookRegistry))
	toolRegistry.Register("openspec_archive", tool.Description, "openspec")
}

func openspecArchiveHandler(client *openspec.Client, hookRegistry *hooks.Registry) func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecArchiveInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecArchiveInput) (*mcp.CallToolResult, any, error) {
		// Trigger beforeArchive hook
		if hookRegistry != nil {
			openspecHooks := hookRegistry.GetOpenSpecHooks()
			if err := hookRegistry.Executor().RunOpenSpecHook(ctx, openspecHooks.BeforeArchive, "before_archive", map[string]string{
				"CHANGE_ID": input.Change,
			}); err != nil {
				// Log but don't fail the operation
				slog.Warn("beforeArchive hook failed", "error", err)
			}
		}

		if err := client.Archive(ctx, input.Change); err != nil {
			return nil, nil, fmt.Errorf("archive %s: %w", input.Change, err)
		}

		// Trigger afterArchive hook
		if hookRegistry != nil {
			openspecHooks := hookRegistry.GetOpenSpecHooks()
			if err := hookRegistry.Executor().RunOpenSpecHook(ctx, openspecHooks.AfterArchive, "after_archive", map[string]string{
				"CHANGE_ID": input.Change,
			}); err != nil {
				// Log but don't fail the operation
				slog.Warn("afterArchive hook failed", "error", err)
			}
		}

		msg := fmt.Sprintf("✓ Archived change: %s\n\nThe change has been moved to archive and main specs have been updated with delta specs.",
			input.Change)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, map[string]string{"change": input.Change, "status": "archived"}, nil
	}
}
