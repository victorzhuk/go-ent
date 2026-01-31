package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/openspec"
)

type OpenSpecListInput struct {
	Type string `json:"type"` // "changes" or "specs"
}

func registerOpenSpecList(s *mcp.Server, client *openspec.Client) {
	tool := &mcp.Tool{
		Name:        "openspec_list",
		Description: "List OpenSpec changes or specs with their status and progress",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"type": map[string]any{
					"type":        "string",
					"description": "What to list: 'changes' (default) or 'specs'",
					"enum":        []string{"changes", "specs"},
				},
			},
		},
	}

	mcp.AddTool(s, tool, openspecListHandler(client))
}

func openspecListHandler(client *openspec.Client) func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecListInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input OpenSpecListInput) (*mcp.CallToolResult, any, error) {
		// Default to changes
		listType := input.Type
		if listType == "" {
			listType = "changes"
		}

		data, err := client.List(ctx, listType)
		if err != nil {
			return nil, nil, fmt.Errorf("list %s: %w", listType, err)
		}

		items, err := openspec.ParseList(data)
		if err != nil {
			return nil, nil, fmt.Errorf("parse list: %w", err)
		}

		// Format output
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# OpenSpec %s\n\n", strings.Title(listType)))
		sb.WriteString(fmt.Sprintf("Found %d %s:\n\n", len(items), listType))

		for i, item := range items {
			sb.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, item.Name))

			if item.Status != "" {
				sb.WriteString(fmt.Sprintf("**Status**: %s\n\n", item.Status))
			}

			if item.TotalTasks > 0 {
				progress := float64(item.CompletedTasks) / float64(item.TotalTasks) * 100
				sb.WriteString(fmt.Sprintf("**Progress**: %d/%d tasks (%.0f%%)\n\n",
					item.CompletedTasks, item.TotalTasks, progress))
			}

			if item.Description != "" {
				sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", item.Description))
			}

			if item.LastModified != "" {
				sb.WriteString(fmt.Sprintf("**Last Modified**: %s\n\n", item.LastModified))
			}
		}

		if len(items) == 0 {
			sb.WriteString(fmt.Sprintf("No %s found.\n", listType))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, items, nil
	}
}
