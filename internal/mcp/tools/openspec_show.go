package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/openspec"
)

type OpenSpecShowInput struct {
	Name string `json:"name"` // Change or spec name
	Type string `json:"type"` // "change" or "spec" (optional)
}

func registerOpenSpecShow(s *mcp.Server, client *openspec.Client) {
	tool := &mcp.Tool{
		Name:        "openspec_show",
		Description: "Show detailed information about a change or spec",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Name of the change or spec to show",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "Type of item: 'change' or 'spec' (optional, auto-detected)",
					"enum":        []string{"change", "spec"},
				},
			},
			"required": []string{"name"},
		},
	}

	mcp.AddTool(s, tool, openspecShowHandler(client))
}

func openspecShowHandler(client *openspec.Client) func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecShowInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecShowInput) (*mcp.CallToolResult, any, error) {
		data, err := client.Show(ctx, input.Type, input.Name)
		if err != nil {
			return nil, nil, fmt.Errorf("show %s: %w", input.Name, err)
		}

		result, err := openspec.ParseShow(data)
		if err != nil {
			return nil, nil, fmt.Errorf("parse show: %w", err)
		}

		// Format as pretty JSON for display
		prettyJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("format result: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(prettyJSON)}},
		}, result, nil
	}
}
