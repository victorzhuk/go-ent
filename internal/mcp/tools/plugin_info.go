package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/plugin"
)

type PluginInfoInput struct {
	Name string `json:"name" jsonschema:"required"`
}

type PluginInfoResponse struct {
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Author      string         `json:"author"`
	Enabled     bool           `json:"enabled"`
	Skills      []SkillRefInfo `json:"skills"`
	Agents      []AgentRefInfo `json:"agents"`
	Rules       []RuleRefInfo  `json:"rules"`
}

type SkillRefInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type AgentRefInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type RuleRefInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func registerPluginInfo(s *mcp.Server, pluginManager *plugin.Manager) {
	tool := &mcp.Tool{
		Name:        "plugin_info",
		Description: "Get detailed information about an installed plugin",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"name"},
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Plugin name",
				},
			},
		},
	}

	mcp.AddTool(s, tool, pluginInfoHandler(pluginManager))
}

func pluginInfoHandler(pluginManager *plugin.Manager) func(ctx context.Context, req *mcp.CallToolRequest, input PluginInfoInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input PluginInfoInput) (*mcp.CallToolResult, any, error) {
		if input.Name == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Plugin name is required"}},
			}, nil, fmt.Errorf("plugin name is required")
		}

		plugin, err := pluginManager.Get(input.Name)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Plugin not found: %s", input.Name)}},
			}, nil, fmt.Errorf("plugin not found: %s", input.Name)
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Plugin: %s\n\n", plugin.Manifest.Name))
		sb.WriteString(fmt.Sprintf("**Version**: %s\n", plugin.Manifest.Version))
		sb.WriteString(fmt.Sprintf("**Description**: %s\n", plugin.Manifest.Description))
		sb.WriteString(fmt.Sprintf("**Author**: %s\n", plugin.Manifest.Author))
		sb.WriteString(fmt.Sprintf("**Status**: %s\n\n", getStatusBadge(plugin.Enabled)))

		if len(plugin.Manifest.Skills) > 0 {
			sb.WriteString(fmt.Sprintf("## Skills (%d)\n\n", len(plugin.Manifest.Skills)))
			for _, s := range plugin.Manifest.Skills {
				sb.WriteString(fmt.Sprintf("### %s\n\n", s.Name))
				sb.WriteString(fmt.Sprintf("**Path**: `%s`\n\n", s.Path))
			}
		}

		if len(plugin.Manifest.Agents) > 0 {
			sb.WriteString(fmt.Sprintf("## Agents (%d)\n\n", len(plugin.Manifest.Agents)))
			for _, a := range plugin.Manifest.Agents {
				sb.WriteString(fmt.Sprintf("### %s\n\n", a.Name))
				sb.WriteString(fmt.Sprintf("**Path**: `%s`\n\n", a.Path))
			}
		}

		if len(plugin.Manifest.Rules) > 0 {
			sb.WriteString(fmt.Sprintf("## Rules (%d)\n\n", len(plugin.Manifest.Rules)))
			for _, r := range plugin.Manifest.Rules {
				sb.WriteString(fmt.Sprintf("### %s\n\n", r.Name))
				sb.WriteString(fmt.Sprintf("**Path**: `%s`\n\n", r.Path))
			}
		}

		sb.WriteString("---\n\n")

		if !plugin.Enabled {
			sb.WriteString("⚠️  This plugin is currently disabled. Use the plugin manager to enable it.\n\n")
		}

		sb.WriteString("**Available Actions**:\n")
		sb.WriteString("1. Use `plugin_install` to install or update\n")
		sb.WriteString("2. Use `plugin_list` to see all installed plugins\n")
		sb.WriteString("3. Use `plugin_search` to find more plugins in the marketplace\n")

		skills := make([]SkillRefInfo, 0, len(plugin.Manifest.Skills))
		for _, s := range plugin.Manifest.Skills {
			skills = append(skills, SkillRefInfo{Name: s.Name, Path: s.Path})
		}

		agents := make([]AgentRefInfo, 0, len(plugin.Manifest.Agents))
		for _, a := range plugin.Manifest.Agents {
			agents = append(agents, AgentRefInfo{Name: a.Name, Path: a.Path})
		}

		rules := make([]RuleRefInfo, 0, len(plugin.Manifest.Rules))
		for _, r := range plugin.Manifest.Rules {
			rules = append(rules, RuleRefInfo{Name: r.Name, Path: r.Path})
		}

		response := PluginInfoResponse{
			Name:        plugin.Manifest.Name,
			Version:     plugin.Manifest.Version,
			Description: plugin.Manifest.Description,
			Author:      plugin.Manifest.Author,
			Enabled:     plugin.Enabled,
			Skills:      skills,
			Agents:      agents,
			Rules:       rules,
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, response, nil
	}
}
