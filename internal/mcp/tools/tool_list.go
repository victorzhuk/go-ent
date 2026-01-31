package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ToolListInput struct{}

type ToolListResponse struct {
	Tools []ToolSummary `json:"tools"`
	Total int           `json:"total"`
}

type ToolSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

func registerToolList(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "tool_list",
		Description: "List all available MCP tools",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}

	mcp.AddTool(s, tool, toolListHandler(s))
}

func toolListHandler(server *mcp.Server) func(ctx context.Context, req *mcp.CallToolRequest, input ToolListInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ToolListInput) (*mcp.CallToolResult, any, error) {
		// Get all registered tools
		// Note: This is a simplified implementation
		// In a real implementation, we'd need to access the server's tool registry

		// For now, hardcode the known tools
		tools := []ToolSummary{
			// Skill tools
			{"skill_list", "List all available skills with their metadata", "skills"},
			{"skill_info", "Get detailed information about a specific skill", "skills"},
			{"skill_validate", "Validate skill format and metadata", "skills"},

			// OpenSpec tools
			{"openspec_list", "List OpenSpec changes or specs", "openspec"},
			{"openspec_show", "Show detailed information about a change or spec", "openspec"},
			{"openspec_new_change", "Create a new OpenSpec change proposal", "openspec"},
			{"openspec_archive", "Archive a completed change", "openspec"},
			{"openspec_validate", "Validate OpenSpec changes and specs", "openspec"},
			{"openspec_status", "Display artifact completion status", "openspec"},
			{"openspec_instructions", "Get enriched instructions for artifacts", "openspec"},

			// Agent tools
			{"agent_list", "List all available agents", "agents"},
			{"agent_info", "Get detailed information about a specific agent", "agents"},
			{"agent_generate", "Generate agent configurations", "agents"},

			// Config tools
			{"config_show", "Show ent.yaml configuration", "config"},
			{"config_set", "Update ent.yaml configuration", "config"},

			// Discovery tools
			{"tool_list", "List all available MCP tools", "discovery"},
			{"skill_match", "Find skills matching a query", "discovery"},

			// Generate tool
			{"generate", "Generate project scaffolding", "generation"},
		}

		var sb strings.Builder
		sb.WriteString("# Available MCP Tools\n\n")
		sb.WriteString(fmt.Sprintf("Total: %d tools\n\n", len(tools)))

		// Group by category
		categories := make(map[string][]ToolSummary)
		for _, t := range tools {
			categories[t.Category] = append(categories[t.Category], t)
		}

		for cat, catTools := range categories {
			sb.WriteString(fmt.Sprintf("## %s (%d)\n\n", strings.Title(cat), len(catTools)))
			for _, t := range catTools {
				sb.WriteString(fmt.Sprintf("- **%s**: %s\n", t.Name, t.Description))
			}
			sb.WriteString("\n")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, ToolListResponse{Tools: tools, Total: len(tools)}, nil
	}
}
