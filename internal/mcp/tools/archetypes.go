package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/generation"
)

type ListArchetypesInput struct {
	ProjectRoot string `json:"project_root,omitempty"`
	Filter      string `json:"filter,omitempty"`
}

func registerListArchetypes(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "go_ent_list_archetypes",
		Description: "List available project archetypes (built-in and custom). Archetypes define template sets for different project types.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_root": map[string]any{
					"type":        "string",
					"description": "Project root directory (defaults to current directory)",
				},
				"filter": map[string]any{
					"type":        "string",
					"description": "Optional filter by archetype name (substring match)",
				},
			},
		},
	}

	mcp.AddTool(s, tool, listArchetypesHandler)
}

func listArchetypesHandler(ctx context.Context, req *mcp.CallToolRequest, input ListArchetypesInput) (*mcp.CallToolResult, any, error) {
	projectRoot := input.ProjectRoot
	if projectRoot == "" {
		projectRoot = "."
	}

	// Load config to get custom archetypes
	cfg, err := generation.LoadConfig(projectRoot)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Error loading config: %v", err),
			}},
		}, nil, nil
	}

	// Get all archetypes
	archetypes := generation.ListArchetypes(cfg)

	// Apply filter if specified
	if input.Filter != "" {
		var filtered []generation.ArchetypeMetadata
		for _, arch := range archetypes {
			if contains(arch.Name, input.Filter) || contains(arch.Description, input.Filter) {
				filtered = append(filtered, arch)
			}
		}
		archetypes = filtered
	}

	// Format output
	data, err := json.MarshalIndent(archetypes, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Error formatting archetypes: %v", err),
			}},
		}, nil, nil
	}

	msg := fmt.Sprintf("# Available Archetypes (%d)\n\n", len(archetypes))
	msg += "```json\n" + string(data) + "\n```\n\n"

	msg += "## Usage\n\n"
	msg += "Use archetype names in:\n"
	msg += "- `generation.yaml` defaults\n"
	msg += "- Component-specific overrides\n"
	msg += "- `go_ent_generate` tool `project_type` parameter\n"

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && len(substr) > 0 && s[0:len(substr)] == substr ||
		len(s) > len(substr) && findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
