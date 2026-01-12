package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/plugin"
)

type PluginInstallInput struct {
	Name    string `json:"name" jsonschema:"required"`
	Version string `json:"version,omitempty"`
}

type PluginInstallResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func registerPluginInstall(s *mcp.Server, pluginManager *plugin.Manager) {
	tool := &mcp.Tool{
		Name:        "plugin_install",
		Description: "Install a plugin from the marketplace",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"name"},
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Plugin name to install",
				},
				"version": map[string]any{
					"type":        "string",
					"description": "Plugin version (defaults to latest if not specified)",
				},
			},
		},
	}

	mcp.AddTool(s, tool, pluginInstallHandler(pluginManager))
}

func pluginInstallHandler(pluginManager *plugin.Manager) func(ctx context.Context, req *mcp.CallToolRequest, input PluginInstallInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input PluginInstallInput) (*mcp.CallToolResult, any, error) {
		if input.Name == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Plugin name is required"}},
			}, nil, fmt.Errorf("plugin name is required")
		}

		version := input.Version
		if version == "" {
			version = "latest"
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Installing Plugin: %s\n\n", input.Name))
		sb.WriteString(fmt.Sprintf("**Version**: %s\n\n", version))

		if err := pluginManager.Install(ctx, input.Name, version); err != nil {
			sb.WriteString(fmt.Sprintf("**❌ Failed**: %s\n\n", err.Error()))

			response := PluginInstallResponse{
				Name:    input.Name,
				Version: version,
				Status:  "failed",
				Message: err.Error(),
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
			}, response, nil
		}

		plugin, err := pluginManager.Get(input.Name)
		if err != nil {
			sb.WriteString(fmt.Sprintf("**⚠️  Warning**: Could not retrieve plugin details: %s\n\n", err.Error()))
		} else {
			sb.WriteString("**✅ Successfully Installed**\n\n")
			sb.WriteString(fmt.Sprintf("**Version**: %s\n", plugin.Manifest.Version))
			sb.WriteString(fmt.Sprintf("**Description**: %s\n", plugin.Manifest.Description))
			sb.WriteString(fmt.Sprintf("**Author**: %s\n\n", plugin.Manifest.Author))

			if len(plugin.Manifest.Skills) > 0 {
				sb.WriteString("**Skills**:\n")
				for _, s := range plugin.Manifest.Skills {
					sb.WriteString(fmt.Sprintf("  • %s\n", s.Name))
				}
				sb.WriteString("\n")
			}

			if len(plugin.Manifest.Agents) > 0 {
				sb.WriteString("**Agents**:\n")
				for _, a := range plugin.Manifest.Agents {
					sb.WriteString(fmt.Sprintf("  • %s\n", a.Name))
				}
				sb.WriteString("\n")
			}

			if len(plugin.Manifest.Rules) > 0 {
				sb.WriteString("**Rules**:\n")
				for _, r := range plugin.Manifest.Rules {
					sb.WriteString(fmt.Sprintf("  • %s\n", r.Name))
				}
				sb.WriteString("\n")
			}
		}

		sb.WriteString("---\n\n")
		sb.WriteString("**Next Steps**:\n")
		sb.WriteString("1. Use `plugin_list` to verify installation\n")
		sb.WriteString("2. Use `plugin_info` to view full plugin details\n")

		response := PluginInstallResponse{
			Name:    input.Name,
			Version: version,
			Status:  "success",
			Message: "Plugin installed successfully",
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, response, nil
	}
}
