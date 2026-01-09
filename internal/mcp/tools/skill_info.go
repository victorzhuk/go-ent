package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/skill"
)

type SkillInfoInput struct {
	Name string `json:"name"`
}

func registerSkillInfo(s *mcp.Server, skillRegistry *skill.Registry) {
	tool := &mcp.Tool{
		Name:        "skill_info",
		Description: "Get detailed information about a specific skill",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Skill name",
				},
			},
			"required": []string{"name"},
		},
	}

	mcp.AddTool(s, tool, skillInfoHandler(skillRegistry))
}

func skillInfoHandler(skillRegistry *skill.Registry) func(ctx context.Context, req *mcp.CallToolRequest, input SkillInfoInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SkillInfoInput) (*mcp.CallToolResult, any, error) {
		if input.Name == "" {
			return nil, nil, fmt.Errorf("name is required")
		}

		meta, err := skillRegistry.Get(input.Name)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Skill not found: %v", err)}},
			}, nil, nil
		}

		content, err := os.ReadFile(meta.FilePath)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error reading skill file: %v", err)}},
			}, nil, nil
		}

		msg := fmt.Sprintf(`# Skill: %s

## Description
%s

## Triggers
%v

## File Path
%s

## Content
`+"```markdown\n%s\n```", meta.Name, meta.Description, meta.Triggers, meta.FilePath, string(content))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, nil, nil
	}
}
