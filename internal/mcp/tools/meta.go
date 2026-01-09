package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerMetaTools registers tool discovery and management tools.
func registerMetaTools(s *mcp.Server, registry *ToolRegistry) {
	registerToolFind(s, registry)
	registerToolDescribe(s, registry)
	registerToolLoad(s, registry)
	registerToolActive(s, registry)
}

// ToolFindInput defines input for tool_find.
type ToolFindInput struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

func registerToolFind(s *mcp.Server, registry *ToolRegistry) {
	tool := &mcp.Tool{
		Name:        "tool_find",
		Description: "Search for tools by query using TF-IDF relevance scoring",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search query (keywords or description)",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Maximum number of results (default: 10)",
				},
			},
			"required": []string{"query"},
		},
	}

	handler := makeToolFindHandler(registry)
	mcp.AddTool(s, tool, handler)
}

func makeToolFindHandler(registry *ToolRegistry) func(context.Context, *mcp.CallToolRequest, ToolFindInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ToolFindInput) (*mcp.CallToolResult, any, error) {
		if input.Query == "" {
			return nil, nil, fmt.Errorf("query is required")
		}

		limit := input.Limit
		if limit <= 0 {
			limit = 10
		}

		results := registry.Find(input.Query, limit)

		var output strings.Builder
		output.WriteString(fmt.Sprintf("Found %d tools matching '%s':\n\n", len(results), input.Query))

		for i, meta := range results {
			active := ""
			if registry.IsActive(meta.Name) {
				active = " ✓ (active)"
			}
			output.WriteString(fmt.Sprintf("%d. **%s**%s\n", i+1, meta.Name, active))
			output.WriteString(fmt.Sprintf("   %s\n\n", meta.Description))
		}

		if len(results) == 0 {
			output.WriteString("No tools found. Try a different query.\n")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: output.String()}},
		}, results, nil
	}
}

// ToolDescribeInput defines input for tool_describe.
type ToolDescribeInput struct {
	Name string `json:"name"`
}

func registerToolDescribe(s *mcp.Server, registry *ToolRegistry) {
	tool := &mcp.Tool{
		Name:        "tool_describe",
		Description: "Get detailed information about a specific tool",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Tool name",
				},
			},
			"required": []string{"name"},
		},
	}

	handler := makeToolDescribeHandler(registry)
	mcp.AddTool(s, tool, handler)
}

func makeToolDescribeHandler(registry *ToolRegistry) func(context.Context, *mcp.CallToolRequest, ToolDescribeInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ToolDescribeInput) (*mcp.CallToolResult, any, error) {
		if input.Name == "" {
			return nil, nil, fmt.Errorf("name is required")
		}

		meta, err := registry.Describe(input.Name)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error: %v", err),
				}},
			}, nil, nil
		}

		schema, _ := json.MarshalIndent(meta.InputSchema, "", "  ")

		var output strings.Builder
		output.WriteString(fmt.Sprintf("# %s\n\n", meta.Name))
		output.WriteString(fmt.Sprintf("**Description:** %s\n\n", meta.Description))

		active := "No"
		if registry.IsActive(meta.Name) {
			active = "Yes ✓"
		}
		output.WriteString(fmt.Sprintf("**Active:** %s\n\n", active))

		if meta.Category != "" {
			output.WriteString(fmt.Sprintf("**Category:** %s\n\n", meta.Category))
		}

		if len(meta.Keywords) > 0 {
			output.WriteString(fmt.Sprintf("**Keywords:** %s\n\n", strings.Join(meta.Keywords, ", ")))
		}

		output.WriteString(fmt.Sprintf("**Input Schema:**\n```json\n%s\n```\n", string(schema)))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: output.String()}},
		}, meta, nil
	}
}

// ToolLoadInput defines input for tool_load.
type ToolLoadInput struct {
	Names []string `json:"names"`
}

func registerToolLoad(s *mcp.Server, registry *ToolRegistry) {
	tool := &mcp.Tool{
		Name:        "tool_load",
		Description: "Load (activate) one or more tools into the active set",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"names": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Tool names to load",
				},
			},
			"required": []string{"names"},
		},
	}

	handler := makeToolLoadHandler(registry)
	mcp.AddTool(s, tool, handler)
}

func makeToolLoadHandler(registry *ToolRegistry) func(context.Context, *mcp.CallToolRequest, ToolLoadInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ToolLoadInput) (*mcp.CallToolResult, any, error) {
		if len(input.Names) == 0 {
			return nil, nil, fmt.Errorf("at least one tool name required")
		}

		if err := registry.Load(input.Names); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error loading tools: %v", err),
				}},
			}, nil, nil
		}

		var output strings.Builder
		output.WriteString(fmt.Sprintf("✅ Loaded %d tool(s):\n\n", len(input.Names)))
		for _, name := range input.Names {
			output.WriteString(fmt.Sprintf("- %s\n", name))
		}

		output.WriteString(fmt.Sprintf("\nTotal active tools: %d\n", len(registry.Active())))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: output.String()}},
		}, registry.Active(), nil
	}
}

// ToolActiveInput defines input for tool_active (no parameters).
type ToolActiveInput struct{}

func registerToolActive(s *mcp.Server, registry *ToolRegistry) {
	tool := &mcp.Tool{
		Name:        "tool_active",
		Description: "List currently active (loaded) tools",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}

	handler := makeToolActiveHandler(registry)
	mcp.AddTool(s, tool, handler)
}

func makeToolActiveHandler(registry *ToolRegistry) func(context.Context, *mcp.CallToolRequest, ToolActiveInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ToolActiveInput) (*mcp.CallToolResult, any, error) {
		active := registry.Active()

		var output strings.Builder
		output.WriteString(fmt.Sprintf("Currently active tools (%d):\n\n", len(active)))

		for i, name := range active {
			meta, err := registry.Describe(name)
			if err == nil {
				output.WriteString(fmt.Sprintf("%d. **%s**\n   %s\n\n", i+1, name, meta.Description))
			} else {
				output.WriteString(fmt.Sprintf("%d. **%s**\n\n", i+1, name))
			}
		}

		if len(active) == 0 {
			output.WriteString("No tools are currently active.\n")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: output.String()}},
		}, active, nil
	}
}
