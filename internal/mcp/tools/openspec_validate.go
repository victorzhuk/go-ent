package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/openspec"
)

type OpenSpecValidateInput struct {
	Scope string `json:"scope"` // "all", "changes", "specs", or empty
}

func registerOpenSpecValidate(s *mcp.Server, client *openspec.Client) {
	tool := &mcp.Tool{
		Name:        "openspec_validate",
		Description: "Validate OpenSpec changes and specs for correctness",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"scope": map[string]any{
					"type":        "string",
					"description": "Validation scope: 'all', 'changes', or 'specs'",
					"enum":        []string{"all", "changes", "specs"},
				},
			},
		},
	}

	mcp.AddTool(s, tool, openspecValidateHandler(client))
}

func openspecValidateHandler(client *openspec.Client) func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecValidateInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecValidateInput) (*mcp.CallToolResult, any, error) {
		scope := input.Scope
		if scope == "" {
			scope = "all"
		}

		data, err := client.Validate(ctx, scope)
		if err != nil {
			return nil, nil, fmt.Errorf("validate: %w", err)
		}

		result, err := openspec.ParseValidate(data)
		if err != nil {
			return nil, nil, fmt.Errorf("parse validate: %w", err)
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
