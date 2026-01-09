package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/skill"
)

type SkillListInput struct {
	Filter string `json:"filter,omitempty"`
}

type SkillListResponse struct {
	Skills []SkillSummary `json:"skills"`
	Total  int            `json:"total"`
}

type SkillSummary struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Triggers    []string `json:"triggers"`
}

func registerSkillList(s *mcp.Server, skillRegistry *skill.Registry) {
	tool := &mcp.Tool{
		Name:        "skill_list",
		Description: "List all available skills with their metadata",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"filter": map[string]any{
					"type":        "string",
					"description": "Optional filter by skill name (substring match)",
				},
			},
		},
	}

	mcp.AddTool(s, tool, skillListHandler(skillRegistry))
}

func skillListHandler(skillRegistry *skill.Registry) func(ctx context.Context, req *mcp.CallToolRequest, input SkillListInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SkillListInput) (*mcp.CallToolResult, any, error) {
		allSkills := skillRegistry.All()

		var filtered []skill.SkillMeta
		if input.Filter != "" {
			filterLower := strings.ToLower(input.Filter)
			for _, s := range allSkills {
				if strings.Contains(strings.ToLower(s.Name), filterLower) {
					filtered = append(filtered, s)
				}
			}
		} else {
			filtered = allSkills
		}

		if len(filtered) == 0 {
			msg := "No skills found"
			if input.Filter != "" {
				msg = fmt.Sprintf("No skills found matching filter: %s", input.Filter)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: msg}},
			}, SkillListResponse{Skills: []SkillSummary{}, Total: 0}, nil
		}

		var sb strings.Builder
		sb.WriteString("# Available Skills\n\n")
		if input.Filter != "" {
			sb.WriteString(fmt.Sprintf("*Filtered by: %s*\n\n", input.Filter))
		}
		sb.WriteString(fmt.Sprintf("Found %d skill(s):\n\n", len(filtered)))

		skills := make([]SkillSummary, 0, len(filtered))

		for i, s := range filtered {
			sb.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, s.Name))
			sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", s.Description))

			if len(s.Triggers) > 0 {
				sb.WriteString("**Triggers**: ")
				for j, trigger := range s.Triggers {
					if j > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(fmt.Sprintf("`%s`", trigger))
				}
				sb.WriteString("\n\n")
			}

			skills = append(skills, SkillSummary{
				Name:        s.Name,
				Description: s.Description,
				Triggers:    s.Triggers,
			})
		}

		sb.WriteString("---\n\n")
		sb.WriteString("**Usage**: Use `skill_info` with a skill name to view full details.\n")

		response := SkillListResponse{
			Skills: skills,
			Total:  len(filtered),
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, response, nil
	}
}
