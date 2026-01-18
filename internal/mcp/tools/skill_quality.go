package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/skill"
)

type SkillQualityInput struct {
	Threshold float64 `json:"threshold,omitempty"`
}

type SkillQualityOutput struct {
	Skills      []SkillScore `json:"skills"`
	AvgScore    float64      `json:"avg_score"`
	BelowThresh []string     `json:"below_threshold,omitempty"`
}

type SkillScore struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

func registerSkillQuality(s *mcp.Server, skillRegistry *skill.Registry) {
	tool := &mcp.Tool{
		Name:        "skill_quality",
		Description: "Get quality scores for all skills",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"threshold": map[string]any{
					"type":        "number",
					"description": "Optional threshold to filter skills below this score (0-100)",
				},
			},
		},
	}

	mcp.AddTool(s, tool, skillQualityHandler(skillRegistry))
}

func skillQualityHandler(skillRegistry *skill.Registry) func(ctx context.Context, req *mcp.CallToolRequest, input SkillQualityInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SkillQualityInput) (*mcp.CallToolResult, any, error) {
		report := skillRegistry.GetQualityReport()

		skills := make([]SkillScore, 0, len(report))
		totalScore := 0.0
		belowThresh := make([]string, 0)

		for name, score := range report {
			skills = append(skills, SkillScore{
				Name:  name,
				Score: score,
			})
			totalScore += score

			if input.Threshold > 0 && score < input.Threshold {
				belowThresh = append(belowThresh, name)
			}
		}

		avgScore := 0.0
		if len(skills) > 0 {
			avgScore = totalScore / float64(len(skills))
		}

		output := SkillQualityOutput{
			Skills:      skills,
			AvgScore:    avgScore,
			BelowThresh: belowThresh,
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: formatQualityOutput(input.Threshold, output)}},
		}, output, nil
	}
}

func formatQualityOutput(threshold float64, output SkillQualityOutput) string {
	var sb strings.Builder

	sb.WriteString("# Skill Quality Report\n\n")
	sb.WriteString(fmt.Sprintf("**Average Score**: %.1f/100\n\n", output.AvgScore))

	if threshold > 0 {
		sb.WriteString(fmt.Sprintf("**Threshold Filter**: %.1f\n\n", threshold))
		if len(output.BelowThresh) > 0 {
			sb.WriteString(fmt.Sprintf("**Below Threshold** (%d): %s\n\n", len(output.BelowThresh), strings.Join(output.BelowThresh, ", ")))
		} else {
			sb.WriteString("**Below Threshold**: None\n\n")
		}
	}

	if len(output.Skills) == 0 {
		sb.WriteString("No skills loaded.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("**Total Skills**: %d\n\n", len(output.Skills)))
	sb.WriteString("## Scores\n\n")

	for _, skill := range output.Skills {
		sb.WriteString(fmt.Sprintf("- **%s**: %.1f/100\n", skill.Name, skill.Score))
	}

	return sb.String()
}
