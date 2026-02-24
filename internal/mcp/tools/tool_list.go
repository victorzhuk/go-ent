package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

func registerToolList(s *mcp.Server, toolRegistry *ToolRegistry) {
	tool := &mcp.Tool{
		Name:        "tool_list",
		Description: "List all available MCP tools",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}

	mcp.AddTool(s, tool, toolListHandler(toolRegistry))
	toolRegistry.Register("tool_list", tool.Description, "discovery")
}

func toolListHandler(toolRegistry *ToolRegistry) func(ctx context.Context, req *mcp.CallToolRequest, input ToolListInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ToolListInput) (*mcp.CallToolResult, any, error) {
		tools := toolRegistry.All()

		var sb strings.Builder
		sb.WriteString("# Available MCP Tools\n\n")
		fmt.Fprintf(&sb, "Total: %d tools\n\n", len(tools))

		// Group by category
		categories := make(map[string][]ToolSummary)
		for _, t := range tools {
			categories[t.Category] = append(categories[t.Category], t)
		}

		caser := cases.Title(language.English)
		for cat, catTools := range categories {
			fmt.Fprintf(&sb, "## %s (%d)\n\n", caser.String(cat), len(catTools))
			for _, t := range catTools {
				fmt.Fprintf(&sb, "- **%s**: %s\n", t.Name, t.Description)
			}
			sb.WriteString("\n")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, ToolListResponse{Tools: tools, Total: len(tools)}, nil
	}
}
