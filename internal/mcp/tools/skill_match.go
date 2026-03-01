package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/skill"
)

const defaultMatchLimit = 5

type SkillMatchInput struct {
	Query string `json:"query"` // Search query
	Limit int    `json:"limit"` // Max results (default: 5)
}

type SkillMatchResponse struct {
	Matches []SkillMatchResult `json:"matches"`
	Query   string             `json:"query"`
}

type SkillMatchResult struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Triggers    []string `json:"triggers"`
	Score       float64  `json:"score"` // Relevance score (0-1)
}

func registerSkillMatch(s *mcp.Server, toolRegistry *ToolRegistry, skillRegistry *skill.Registry) {
	tool := &mcp.Tool{
		Name:        "skill_match",
		Description: "Find skills matching a query using keyword search",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search query to match against skills",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Maximum number of results (default: 5)",
					"default":     5,
				},
			},
			"required": []string{"query"},
		},
	}

	mcp.AddTool(s, tool, skillMatchHandler(skillRegistry))
	toolRegistry.Register("skill_match", tool.Description, "discovery")
}

func skillMatchHandler(skillRegistry *skill.Registry) func(ctx context.Context, req *mcp.CallToolRequest, input SkillMatchInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SkillMatchInput) (*mcp.CallToolResult, any, error) {
		if input.Limit == 0 {
			input.Limit = defaultMatchLimit
		}

		queryLower := strings.ToLower(input.Query)
		queryWords := strings.Fields(queryLower)

		allSkills := skillRegistry.All()
		var matches []SkillMatchResult

		// Simple keyword matching
		for _, s := range allSkills {
			score := 0.0

			// Check name
			nameLower := strings.ToLower(s.Name)
			for _, word := range queryWords {
				if strings.Contains(nameLower, word) {
					score += 1.0
				}
			}

			// Check description
			descLower := strings.ToLower(s.Description)
			for _, word := range queryWords {
				if strings.Contains(descLower, word) {
					score += 0.5
				}
			}

			// Check triggers
			for _, trigger := range s.Triggers {
				triggerLower := strings.ToLower(trigger)
				for _, word := range queryWords {
					if strings.Contains(triggerLower, word) {
						score += 0.8
					}
				}
			}

			if score > 0 {
				matches = append(matches, SkillMatchResult{
					Name:        s.Name,
					Description: s.Description,
					Triggers:    s.Triggers,
					Score:       score,
				})
			}
		}

		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Score > matches[j].Score
		})

		// Limit results
		if len(matches) > input.Limit {
			matches = matches[:input.Limit]
		}

		// Format output
		var sb strings.Builder
		fmt.Fprintf(&sb, "# Skills matching: %s\n\n", input.Query)
		fmt.Fprintf(&sb, "Found %d match(es):\n\n", len(matches))

		for i, m := range matches {
			fmt.Fprintf(&sb, "## %d. %s (score: %.1f)\n\n", i+1, m.Name, m.Score)
			fmt.Fprintf(&sb, "**Description**: %s\n\n", m.Description)
			if len(m.Triggers) > 0 {
				fmt.Fprintf(&sb, "**Triggers**: %s\n\n", strings.Join(m.Triggers, ", "))
			}
		}

		if len(matches) == 0 {
			sb.WriteString("No matching skills found. Try different keywords.\n")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, SkillMatchResponse{Matches: matches, Query: input.Query}, nil
	}
}
