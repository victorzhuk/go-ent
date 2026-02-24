package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/skill"
)

type SkillValidateInput struct {
	Name   string `json:"name,omitempty"`
	Strict bool   `json:"strict,omitempty"`
}

type SkillValidateOutput struct {
	Valid  bool                    `json:"valid"`
	Issues []skill.ValidationIssue `json:"issues"`
}

func registerSkillValidate(s *mcp.Server, toolRegistry *ToolRegistry, skillRegistry *skill.Registry) {
	tool := &mcp.Tool{
		Name:        "skill_validate",
		Description: "Validate skill structure and content quality",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Optional skill name to validate (empty validates all skills)",
				},
				"strict": map[string]any{
					"type":        "boolean",
					"description": "Enable strict validation mode (all issues treated as errors)",
				},
			},
		},
	}

	mcp.AddTool(s, tool, skillValidateHandler(skillRegistry))
	toolRegistry.Register("skill_validate", tool.Description, "skills")
}

func skillValidateHandler(skillRegistry *skill.Registry) func(ctx context.Context, req *mcp.CallToolRequest, input SkillValidateInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SkillValidateInput) (*mcp.CallToolResult, any, error) {
		var result *skill.ValidationResult
		var err error

		if input.Name != "" {
			result, err = skillRegistry.ValidateSkill(input.Name)
			if err != nil {
				return nil, nil, fmt.Errorf("validate skill %s: %w", input.Name, err)
			}
		}

		output := SkillValidateOutput{
			Valid:  result.Valid,
			Issues: result.Issues,
		}

		if input.Strict {
			output.Valid = len(output.Issues) == 0
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: formatValidationOutput(input.Name, output)}},
		}, output, nil
	}
}

func formatValidationOutput(skillName string, output SkillValidateOutput) string {
	var sb strings.Builder

	if skillName != "" {
		fmt.Fprintf(&sb, "# Validation: %s\n\n", skillName)
	} else {
		sb.WriteString("# Validation: All Skills\n\n")
	}

	status := "✓ VALID"
	if !output.Valid {
		status = "✗ INVALID"
	}
	fmt.Fprintf(&sb, "**Status**: %s\n\n", status)

	if len(output.Issues) == 0 {
		sb.WriteString("No issues found.\n")
		return sb.String()
	}

	errors := 0
	warnings := 0
	infos := 0

	for _, issue := range output.Issues {
		switch issue.Severity {
		case skill.SeverityError:
			errors++
		case skill.SeverityWarning:
			warnings++
		case skill.SeverityInfo:
			infos++
		}
	}

	fmt.Fprintf(&sb, "**Issues**: %d total (%d errors, %d warnings, %d info)\n\n", len(output.Issues), errors, warnings, infos)

	sb.WriteString("## Issues\n\n")

	for i, issue := range output.Issues {
		fmt.Fprintf(&sb, "### %d. [%s] %s\n\n", i+1, issue.Severity, issue.Rule)

		if issue.Line > 0 {
			fmt.Fprintf(&sb, "**Line**: %d\n\n", issue.Line)
		}

		fmt.Fprintf(&sb, "**Message**: %s\n\n", issue.Message)

		if issue.Suggestion != "" {
			fmt.Fprintf(&sb, "**Suggestion**: %s\n\n", issue.Suggestion)
		}

		if issue.Example != "" {
			fmt.Fprintf(&sb, "**Example**: %s\n\n", issue.Example)
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}
