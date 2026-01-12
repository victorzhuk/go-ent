package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/marketplace"
)

type PluginSearchInput struct {
	Query    string `json:"query,omitempty"`
	Category string `json:"category,omitempty"`
	Author   string `json:"author,omitempty"`
	SortBy   string `json:"sort_by,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

type PluginSearchResponse struct {
	Plugins []marketplace.PluginInfo `json:"plugins"`
	Total   int                      `json:"total"`
	Query   string                   `json:"query"`
}

type PluginSearchResult struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Category    string   `json:"category"`
	Downloads   int      `json:"downloads"`
	Rating      float64  `json:"rating"`
	Tags        []string `json:"tags"`
	Skills      int      `json:"skills"`
	Agents      int      `json:"agents"`
	Rules       int      `json:"rules"`
}

func registerPluginSearch(s *mcp.Server, searcher *marketplace.Searcher) {
	tool := &mcp.Tool{
		Name:        "plugin_search",
		Description: "Search for plugins in the marketplace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search query string",
				},
				"category": map[string]any{
					"type":        "string",
					"description": "Filter by category (skills, agents, rules, development, testing, devops, enterprise, utilities)",
				},
				"author": map[string]any{
					"type":        "string",
					"description": "Filter by author name",
				},
				"sort_by": map[string]any{
					"type":        "string",
					"description": "Sort by field (downloads, rating, name)",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Maximum number of results to return",
				},
			},
		},
	}

	mcp.AddTool(s, tool, pluginSearchHandler(searcher))
}

func pluginSearchHandler(searcher *marketplace.Searcher) func(ctx context.Context, req *mcp.CallToolRequest, input PluginSearchInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input PluginSearchInput) (*mcp.CallToolResult, any, error) {
		opts := marketplace.SearchOptions{
			Category: input.Category,
			Author:   input.Author,
			SortBy:   input.SortBy,
			Limit:    input.Limit,
		}

		plugins, err := searcher.Search(ctx, input.Query, opts)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Search failed: %s", err.Error())}},
			}, nil, err
		}

		if len(plugins) == 0 {
			msg := "No plugins found"
			if input.Query != "" {
				msg = fmt.Sprintf("No plugins found matching: %s", input.Query)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: msg}},
			}, PluginSearchResponse{Plugins: []marketplace.PluginInfo{}, Total: 0, Query: input.Query}, nil
		}

		var sb strings.Builder
		sb.WriteString("# Search Results\n\n")

		if input.Query != "" {
			sb.WriteString(fmt.Sprintf("**Query**: %s\n\n", input.Query))
		}

		if input.Category != "" {
			sb.WriteString(fmt.Sprintf("**Category**: %s\n", input.Category))
		}
		if input.Author != "" {
			sb.WriteString(fmt.Sprintf("**Author**: %s\n", input.Author))
		}
		if input.SortBy != "" {
			sb.WriteString(fmt.Sprintf("**Sorted by**: %s\n", input.SortBy))
		}
		if len(plugins) > 0 {
			sb.WriteString(fmt.Sprintf("**Total Results**: %d\n\n", len(plugins)))
		}

		results := make([]PluginSearchResult, 0, len(plugins))

		for i, p := range plugins {
			sb.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, p.Name))
			sb.WriteString(fmt.Sprintf("**Version**: %s\n", p.Version))
			sb.WriteString(fmt.Sprintf("**Description**: %s\n", p.Description))
			sb.WriteString(fmt.Sprintf("**Author**: %s\n", p.Author))

			if p.Category != "" {
				sb.WriteString(fmt.Sprintf("**Category**: %s\n", p.Category))
			}

			sb.WriteString(fmt.Sprintf("**Downloads**: %d\n", p.Downloads))
			sb.WriteString(fmt.Sprintf("**Rating**: %.1f\n", p.Rating))

			if p.Skills > 0 || p.Agents > 0 || p.Rules > 0 {
				sb.WriteString("**Includes**: ")
				var includes []string
				if p.Skills > 0 {
					includes = append(includes, fmt.Sprintf("%d skills", p.Skills))
				}
				if p.Agents > 0 {
					includes = append(includes, fmt.Sprintf("%d agents", p.Agents))
				}
				if p.Rules > 0 {
					includes = append(includes, fmt.Sprintf("%d rules", p.Rules))
				}
				sb.WriteString(strings.Join(includes, ", "))
				sb.WriteString("\n")
			}

			if len(p.Tags) > 0 {
				sb.WriteString("**Tags**: ")
				for j, tag := range p.Tags {
					if j > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(fmt.Sprintf("`%s`", tag))
				}
				sb.WriteString("\n")
			}

			sb.WriteString("\n")

			results = append(results, PluginSearchResult{
				Name:        p.Name,
				Version:     p.Version,
				Description: p.Description,
				Author:      p.Author,
				Category:    p.Category,
				Downloads:   p.Downloads,
				Rating:      p.Rating,
				Tags:        p.Tags,
				Skills:      p.Skills,
				Agents:      p.Agents,
				Rules:       p.Rules,
			})
		}

		sb.WriteString("---\n\n")
		sb.WriteString("**Installation**: Use `plugin_install` with a plugin name to install.\n")

		response := PluginSearchResponse{
			Plugins: plugins,
			Total:   len(plugins),
			Query:   input.Query,
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, response, nil
	}
}
