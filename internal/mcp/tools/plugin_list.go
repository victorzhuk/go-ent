package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/plugin"
)

type PluginListInput struct {
	Filter string `json:"filter,omitempty"`
}

type PluginListResponse struct {
	Plugins []plugin.PluginInfo `json:"plugins"`
	Total   int                 `json:"total"`
}

type PluginSummary struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Enabled     bool   `json:"enabled"`
}

func registerPluginList(s *mcp.Server, pluginManager *plugin.Manager) {
	tool := &mcp.Tool{
		Name:        "plugin_list",
		Description: "List all installed plugins with their metadata",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"filter": map[string]any{
					"type":        "string",
					"description": "Optional filter by plugin name (substring match)",
				},
			},
		},
	}

	mcp.AddTool(s, tool, pluginListHandler(pluginManager))
}

func pluginListHandler(pluginManager *plugin.Manager) func(ctx context.Context, req *mcp.CallToolRequest, input PluginListInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input PluginListInput) (*mcp.CallToolResult, any, error) {
		allPlugins := pluginManager.List()

		var filtered []plugin.PluginInfo
		if input.Filter != "" {
			filterLower := strings.ToLower(input.Filter)
			for _, p := range allPlugins {
				if strings.Contains(strings.ToLower(p.Name), filterLower) {
					filtered = append(filtered, p)
				}
			}
		} else {
			filtered = allPlugins
		}

		if len(filtered) == 0 {
			msg := "No plugins found"
			if input.Filter != "" {
				msg = fmt.Sprintf("No plugins found matching filter: %s", input.Filter)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: msg}},
			}, PluginListResponse{Plugins: []plugin.PluginInfo{}, Total: 0}, nil
		}

		var sb strings.Builder
		sb.WriteString("# Installed Plugins\n\n")
		if input.Filter != "" {
			sb.WriteString(fmt.Sprintf("*Filtered by: %s*\n\n", input.Filter))
		}
		sb.WriteString(fmt.Sprintf("Found %d plugin(s):\n\n", len(filtered)))

		plugins := make([]plugin.PluginInfo, 0, len(filtered))

		for i, p := range filtered {
			sb.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, p.Name))
			sb.WriteString(fmt.Sprintf("**Version**: %s\n\n", p.Version))
			sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", p.Description))
			sb.WriteString(fmt.Sprintf("**Author**: %s\n\n", p.Author))
			sb.WriteString(fmt.Sprintf("**Status**: %s\n\n", getStatusBadge(p.Enabled)))

			if p.Skills > 0 {
				sb.WriteString(fmt.Sprintf("**Skills**: %d\n", p.Skills))
			}
			if p.Agents > 0 {
				sb.WriteString(fmt.Sprintf("**Agents**: %d\n", p.Agents))
			}
			if p.Rules > 0 {
				sb.WriteString(fmt.Sprintf("**Rules**: %d\n", p.Rules))
			}
			sb.WriteString("\n")

			plugins = append(plugins, p)
		}

		sb.WriteString("---\n\n")
		sb.WriteString("**Usage**: Use `plugin_info` with a plugin name to view full details.\n")

		response := PluginListResponse{
			Plugins: plugins,
			Total:   len(filtered),
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, response, nil
	}
}

func getStatusBadge(enabled bool) string {
	if enabled {
		return "✅ Enabled"
	}
	return "❌ Disabled"
}
