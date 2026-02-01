package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/openspec"
)

type OpenSpecStatusInput struct {
	Change string `json:"change"` // Change name (optional)
}

func registerOpenSpecStatus(s *mcp.Server, toolRegistry *ToolRegistry, client *openspec.Client) {
	tool := &mcp.Tool{
		Name:        "openspec_status",
		Description: "Display artifact completion status for a change",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"change": map[string]any{
					"type":        "string",
					"description": "Name of the change (optional, prompts if not provided)",
				},
			},
		},
	}

	mcp.AddTool(s, tool, openspecStatusHandler(client))
	toolRegistry.Register("openspec_status", tool.Description, "openspec")
}

func openspecStatusHandler(client *openspec.Client) func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecStatusInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecStatusInput) (*mcp.CallToolResult, any, error) {
		data, err := client.Status(ctx, input.Change)
		if err != nil {
			return nil, nil, fmt.Errorf("status: %w", err)
		}

		result, err := openspec.ParseStatus(data)
		if err != nil {
			return nil, nil, fmt.Errorf("parse status: %w", err)
		}

		// Format as pretty JSON
		prettyJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("format result: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(prettyJSON)}},
		}, result, nil
	}
}
