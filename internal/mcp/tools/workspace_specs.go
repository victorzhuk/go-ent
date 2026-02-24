package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/workspace"
)

type WorkspaceSpecsInput struct {
	SpecID string `json:"spec_id"`
}

func registerWorkspaceSpecs(s *mcp.Server, toolRegistry *ToolRegistry) {
	tool := &mcp.Tool{
		Name:        "workspace_specs",
		Description: "List or read workspace-level shared specs. Without spec_id, lists all workspace specs. With spec_id, reads that spec.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"spec_id": map[string]any{
					"type":        "string",
					"description": "Spec ID to read (optional — omit to list all)",
				},
			},
		},
	}

	mcp.AddTool(s, tool, workspaceSpecsHandler())
	toolRegistry.Register("workspace_specs", tool.Description, "workspace")
}

func workspaceSpecsHandler() func(ctx context.Context, req *mcp.CallToolRequest, input WorkspaceSpecsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input WorkspaceSpecsInput) (*mcp.CallToolResult, any, error) {
		ws, err := workspace.DetectAndResolve(".")
		if err != nil {
			return nil, nil, fmt.Errorf("detect workspace: %w", err)
		}
		if ws == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "No workspace configured for current project"}},
			}, nil, nil
		}

		if input.SpecID != "" {
			return readWorkspaceSpec(ws, input.SpecID)
		}

		return listWorkspaceSpecs(ws)
	}
}

func listWorkspaceSpecs(ws *workspace.Workspace) (*mcp.CallToolResult, any, error) {
	specsDir := filepath.Join(ws.Path, "openspec", "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "No workspace specs found"}},
			}, nil, nil
		}
		return nil, nil, fmt.Errorf("read specs: %w", err)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Workspace: %s\nSpecs:\n\n", ws.Name)

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		specPath := filepath.Join(specsDir, e.Name(), "spec.md")
		title := workspace.ExtractTitle(specPath)

		if title != "" {
			fmt.Fprintf(&b, "- %s: %s\n", e.Name(), title)
		} else {
			fmt.Fprintf(&b, "- %s\n", e.Name())
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: b.String()}},
	}, nil, nil
}

func readWorkspaceSpec(ws *workspace.Workspace, specID string) (*mcp.CallToolResult, any, error) {
	specPath := filepath.Join(ws.Path, "openspec", "specs", specID, "spec.md")
	data, err := os.ReadFile(specPath) // #nosec G304
	if err != nil {
		return nil, nil, fmt.Errorf("read spec %s: %w", specID, err)
	}

	return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, map[string]string{
			"spec_id":   specID,
			"workspace": ws.Name,
		}, nil
}
