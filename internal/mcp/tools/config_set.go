package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"
)

type ConfigSetInput struct {
	Key   string `json:"key"`   // Key path (e.g., "models.fast.claude")
	Value string `json:"value"` // New value
}

func registerConfigSet(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "config_set",
		Description: "Update a value in ent.yaml configuration",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"key": map[string]any{
					"type":        "string",
					"description": "Configuration key path (e.g., 'models.fast.claude')",
				},
				"value": map[string]any{
					"type":        "string",
					"description": "New value to set",
				},
			},
			"required": []string{"key", "value"},
		},
	}

	mcp.AddTool(s, tool, configSetHandler())
}

func configSetHandler() func(ctx context.Context, req *mcp.CallToolRequest, input ConfigSetInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ConfigSetInput) (*mcp.CallToolResult, any, error) {
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

		// TODO: Implement key path setting
		// For now, return an error
		return nil, nil, fmt.Errorf("config_set not yet implemented - please edit ent.yaml manually")
	}
}
