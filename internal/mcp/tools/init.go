package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/spec"
)

type SpecInitInput struct {
	Path        string            `json:"path"`
	Name        string            `json:"name,omitempty"`
	Module      string            `json:"module,omitempty"`
	Description string            `json:"description,omitempty"`
	Conventions map[string]string `json:"conventions,omitempty"`
}

func registerInit(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "go_ent_spec_init",
		Description: "Initialize openspec folder in a project directory",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Path to the project directory where openspec folder will be created",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Project name",
				},
				"module": map[string]any{
					"type":        "string",
					"description": "Go module path",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Project description",
				},
				"conventions": map[string]any{
					"type":        "object",
					"description": "Project-specific conventions (key-value pairs)",
					"additionalProperties": map[string]any{
						"type": "string",
					},
				},
			},
			"required": []string{"path"},
		},
	}

	mcp.AddTool(s, tool, specInitHandler)
}

func specInitHandler(ctx context.Context, req *mcp.CallToolRequest, input SpecInitInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	store := spec.NewStore(input.Path)

	exists, err := store.Exists()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error checking .spec folder: %v", err)}},
		}, nil, nil
	}

	if exists {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(".spec folder already exists at %s", store.SpecPath())}},
		}, nil, nil
	}

	project := spec.Project{
		Name:        input.Name,
		Module:      input.Module,
		Description: input.Description,
		Conventions: input.Conventions,
	}

	if project.Conventions == nil {
		project.Conventions = make(map[string]string)
	}

	if err := store.Init(project); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error initializing .spec folder: %v", err)}},
		}, nil, nil
	}

	msg := fmt.Sprintf("✅ Initialized .spec folder at %s\n\n", store.SpecPath())
	msg += "Created structure:\n"
	msg += "  - .spec/project.yaml\n"
	msg += "  - .spec/specs/\n"
	msg += "  - .spec/changes/\n"
	msg += "  - .spec/tasks/\n"
	msg += "  - .spec/archive/\n"

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}
