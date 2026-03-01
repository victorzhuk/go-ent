package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/config"
	"gopkg.in/yaml.v3"
)

type ConfigShowInput struct {
	ProjectDir string `json:"project_dir,omitempty"`
	Runtime    string `json:"runtime,omitempty"`
}

func registerConfigShow(s *mcp.Server, toolRegistry *ToolRegistry) {
	tool := &mcp.Tool{
		Name:        "config_show",
		Description: "Show ent runtime configuration (model aliases)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_dir": map[string]any{
					"type":        "string",
					"description": "Project directory (defaults to current directory)",
				},
				"runtime": map[string]any{
					"type":        "string",
					"description": "Runtime to show config for (claude|opencode, auto-detected if omitted)",
				},
			},
		},
	}

	mcp.AddTool(s, tool, configShowHandler())
	toolRegistry.Register("config_show", tool.Description, "config")
}

func configShowHandler() func(ctx context.Context, req *mcp.CallToolRequest, input ConfigShowInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ConfigShowInput) (*mcp.CallToolResult, any, error) {
		projectDir := input.ProjectDir
		if projectDir == "" {
			projectDir = "."
		}

		rt := input.Runtime
		if rt == "" {
			rt = config.DetectRuntime(projectDir)
		}
		if rt == "" {
			rt = "claude"
		}

		cfg, err := config.LoadRuntimeConfig(projectDir, rt)
		if err != nil {
			return nil, nil, fmt.Errorf("load config: %w", err)
		}

		output, err := yaml.Marshal(cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("format yaml: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(output)}},
		}, cfg, nil
	}
}
