package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/config"
	"gopkg.in/yaml.v3"
)

type ConfigSetInput struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	ProjectDir string `json:"project_dir,omitempty"`
	Runtime    string `json:"runtime,omitempty"`
}

func registerConfigSet(s *mcp.Server, toolRegistry *ToolRegistry) {
	tool := &mcp.Tool{
		Name:        "config_set",
		Description: "Update a model alias in ent runtime configuration",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"key": map[string]any{
					"type":        "string",
					"description": "Config key (e.g. claude.sonnet, opencode.fast)",
				},
				"value": map[string]any{
					"type":        "string",
					"description": "Model ID to set",
				},
				"project_dir": map[string]any{
					"type":        "string",
					"description": "Project directory (defaults to current directory)",
				},
				"runtime": map[string]any{
					"type":        "string",
					"description": "Runtime (claude|opencode, auto-detected if omitted)",
				},
			},
			"required": []string{"key", "value"},
		},
	}

	mcp.AddTool(s, tool, configSetHandler())
	toolRegistry.Register("config_set", tool.Description, "config")
}

func configSetHandler() func(ctx context.Context, req *mcp.CallToolRequest, input ConfigSetInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ConfigSetInput) (*mcp.CallToolResult, any, error) {
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

		cfg, err := config.LoadToolRuntimeConfig(projectDir, rt)
		if err != nil {
			return nil, nil, fmt.Errorf("load config: %w", err)
		}

		if err := mcpApplyConfigKey(cfg, input.Key, input.Value); err != nil {
			return nil, nil, err
		}

		cfgPath := mcpRuntimeConfigPath(rt, projectDir)
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0o750); err != nil {
			return nil, nil, fmt.Errorf("create config directory: %w", err)
		}

		data, err := yaml.Marshal(cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal config: %w", err)
		}

		if err := os.WriteFile(cfgPath, data, 0o600); err != nil {
			return nil, nil, fmt.Errorf("write config: %w", err)
		}

		msg := fmt.Sprintf("Updated %s = %s", input.Key, input.Value)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, nil, nil
	}
}

func mcpApplyConfigKey(cfg *config.ToolRuntimeConfig, key, value string) error {
	switch key {
	case "claude.sonnet":
		cfg.Claude.Sonnet = value
	case "claude.opus":
		cfg.Claude.Opus = value
	case "claude.haiku":
		cfg.Claude.Haiku = value
	case "opencode.fast":
		cfg.OpenCode.Fast = value
	case "opencode.main":
		cfg.OpenCode.Main = value
	case "opencode.heavy":
		cfg.OpenCode.Heavy = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

func mcpRuntimeConfigPath(runtime, projectDir string) string {
	switch runtime {
	case "opencode":
		return filepath.Join(projectDir, ".opencode", "ent.yaml")
	default:
		return filepath.Join(projectDir, ".claude", "ent.yaml")
	}
}
