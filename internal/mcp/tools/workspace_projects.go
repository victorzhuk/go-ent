package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/workspace"
)

type WorkspaceProjectsInput struct {
	Project string `json:"project"`
}

func registerWorkspaceProjects(s *mcp.Server, toolRegistry *ToolRegistry) {
	tool := &mcp.Tool{
		Name:        "workspace_projects",
		Description: "List sibling projects in the workspace and their specs",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project": map[string]any{
					"type":        "string",
					"description": "Project name to get details for (optional — omit to list all)",
				},
			},
		},
	}

	mcp.AddTool(s, tool, workspaceProjectsHandler())
	toolRegistry.Register("workspace_projects", tool.Description, "workspace")
}

func workspaceProjectsHandler() func(ctx context.Context, req *mcp.CallToolRequest, input WorkspaceProjectsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input WorkspaceProjectsInput) (*mcp.CallToolResult, any, error) {
		ws, err := workspace.DetectAndResolve(".")
		if err != nil {
			return nil, nil, fmt.Errorf("detect workspace: %w", err)
		}
		if ws == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "No workspace configured for current project"}},
			}, nil, nil
		}

		var b strings.Builder
		b.WriteString(fmt.Sprintf("Workspace: %s\n\n", ws.Name))

		if input.Project != "" {
			for _, p := range ws.Projects {
				if p.Name == input.Project {
					b.WriteString(fmt.Sprintf("Project: %s\n", p.Name))
					b.WriteString(fmt.Sprintf("Path: %s\n", p.Path))
					if p.Description != "" {
						b.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
					}
					break
				}
			}
		} else {
			b.WriteString("Projects:\n")
			for _, p := range ws.Projects {
				desc := ""
				if p.Description != "" {
					desc = " — " + p.Description
				}
				b.WriteString(fmt.Sprintf("- %s (%s)%s\n", p.Name, p.Path, desc))
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: b.String()}},
		}, nil, nil
	}
}
