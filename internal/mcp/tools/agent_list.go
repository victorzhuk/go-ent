package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent"
)

type AgentListInput struct {
	Role       string `json:"role,omitempty"`       // Filter by role
	Complexity string `json:"complexity,omitempty"` // Filter by complexity
}

type AgentListResponse struct {
	Agents []AgentSummary `json:"agents"`
	Total  int            `json:"total"`
}

type AgentSummary struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Role        string   `json:"role,omitempty"`
	Complexity  string   `json:"complexity,omitempty"`
	Model       string   `json:"model"`
	Skills      []string `json:"skills,omitempty"`
}

func registerAgentList(s *mcp.Server, agentRegistry *agent.Registry) {
	tool := &mcp.Tool{
		Name:        "agent_list",
		Description: "List all available agents with their metadata",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"role": map[string]any{
					"type":        "string",
					"description": "Filter by role (planning, execution, debugging, validation, research)",
				},
				"complexity": map[string]any{
					"type":        "string",
					"description": "Filter by complexity (standard, fast, heavy)",
				},
			},
		},
	}

	mcp.AddTool(s, tool, agentListHandler(agentRegistry))
}

func agentListHandler(agentRegistry *agent.Registry) func(ctx context.Context, req *mcp.CallToolRequest, input AgentListInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AgentListInput) (*mcp.CallToolResult, any, error) {
		allAgents := agentRegistry.All()

		// Filter agents
		var filtered []agent.AgentMeta
		for _, a := range allAgents {
			if input.Role != "" && a.Role != input.Role {
				continue
			}
			if input.Complexity != "" && a.Complexity != input.Complexity {
				continue
			}
			filtered = append(filtered, a)
		}

		if len(filtered) == 0 {
			msg := "No agents found"
			if input.Role != "" || input.Complexity != "" {
				msg = "No agents found matching filters"
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: msg}},
			}, AgentListResponse{Agents: []AgentSummary{}, Total: 0}, nil
		}

		// Format output
		var sb strings.Builder
		sb.WriteString("# Available Agents\n\n")
		if input.Role != "" || input.Complexity != "" {
			sb.WriteString("*Filtered*\n\n")
		}
		sb.WriteString(fmt.Sprintf("Found %d agent(s):\n\n", len(filtered)))

		agents := make([]AgentSummary, 0, len(filtered))

		for i, a := range filtered {
			sb.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, a.Name))
			sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", a.Description))

			if a.Role != "" {
				sb.WriteString(fmt.Sprintf("**Role**: %s\n\n", a.Role))
			}

			if a.Complexity != "" {
				sb.WriteString(fmt.Sprintf("**Complexity**: %s\n\n", a.Complexity))
			}

			sb.WriteString(fmt.Sprintf("**Model**: %s\n\n", a.Model))

			if len(a.Skills) > 0 {
				sb.WriteString(fmt.Sprintf("**Skills**: %s\n\n", strings.Join(a.Skills, ", ")))
			}

			if len(a.Dependencies) > 0 {
				sb.WriteString(fmt.Sprintf("**Dependencies**: %s\n\n", strings.Join(a.Dependencies, ", ")))
			}

			agents = append(agents, AgentSummary{
				Name:        a.Name,
				Description: a.Description,
				Role:        a.Role,
				Complexity:  a.Complexity,
				Model:       a.Model,
				Skills:      a.Skills,
			})
		}

		response := AgentListResponse{
			Agents: agents,
			Total:  len(filtered),
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, response, nil
	}
}
