package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"
)

type ConfigShowInput struct {
	Key string `json:"key,omitempty"` // Optional key to show specific config value
}

func registerConfigShow(s *mcp.Server, toolRegistry *ToolRegistry) {
	tool := &mcp.Tool{
		Name:        "config_show",
		Description: "Show ent.yaml configuration",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"key": map[string]any{
					"type":        "string",
					"description": "Optional key path (e.g., 'models.fast') to show specific value",
				},
			},
		},
	}

	mcp.AddTool(s, tool, configShowHandler())
	toolRegistry.Register("config_show", tool.Description, "config")
}

func configShowHandler() func(ctx context.Context, req *mcp.CallToolRequest, input ConfigShowInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ConfigShowInput) (*mcp.CallToolResult, any, error) {
		// Read ent.yaml
		data, err := os.ReadFile("ent.yaml")
		if err != nil {
			return nil, nil, fmt.Errorf("read ent.yaml: %w", err)
		}

		// Parse YAML
		var config map[string]any
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, nil, fmt.Errorf("parse ent.yaml: %w", err)
		}

		// If key is specified, extract that value
		// TODO: Implement key path extraction
		// For now, just return the whole config
		result := config

		// Format as YAML
		output, err := yaml.Marshal(result)
		if err != nil {
			return nil, nil, fmt.Errorf("format yaml: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(output)}},
		}, result, nil
	}
}
