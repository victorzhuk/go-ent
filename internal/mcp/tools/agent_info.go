package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent"
)

type AgentInfoInput struct {
	Name string `json:"name"` // Agent name
}

func registerAgentInfo(s *mcp.Server, agentRegistry *agent.Registry) {
	tool := &mcp.Tool{
		Name:        "agent_info",
		Description: "Get detailed information about a specific agent",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Name of the agent to get info for",
				},
			},
			"required": []string{"name"},
		},
	}

	mcp.AddTool(s, tool, agentInfoHandler(agentRegistry))
}

func agentInfoHandler(agentRegistry *agent.Registry) func(ctx context.Context, req *mcp.CallToolRequest, input AgentInfoInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AgentInfoInput) (*mcp.CallToolResult, any, error) {
		meta, ok := agentRegistry.Get(input.Name)
		if !ok {
			return nil, nil, fmt.Errorf("agent not found: %s", input.Name)
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Agent: %s\n\n", meta.Name))
		sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", meta.Description))
		sb.WriteString(fmt.Sprintf("**Model**: %s\n\n", meta.Model))

		if meta.Role != "" {
			sb.WriteString(fmt.Sprintf("**Role**: %s\n\n", meta.Role))
		}

		if meta.Complexity != "" {
			sb.WriteString(fmt.Sprintf("**Complexity**: %s\n\n", meta.Complexity))
		}

		if meta.Color != "" {
			sb.WriteString(fmt.Sprintf("**Color**: %s\n\n", meta.Color))
		}

		if len(meta.Skills) > 0 {
			sb.WriteString("## Skills\n\n")
			for _, skill := range meta.Skills {
				sb.WriteString(fmt.Sprintf("- %s\n", skill))
			}
			sb.WriteString("\n")
		}

		if len(meta.ToolPresets) > 0 {
			sb.WriteString("## Tool Presets\n\n")
			for _, preset := range meta.ToolPresets {
				sb.WriteString(fmt.Sprintf("- %s\n", preset))
			}
			sb.WriteString("\n")
		}

		if len(meta.DisallowedToolPresets) > 0 {
			sb.WriteString("## Disallowed Tool Presets\n\n")
			for _, preset := range meta.DisallowedToolPresets {
				sb.WriteString(fmt.Sprintf("- %s\n", preset))
			}
			sb.WriteString("\n")
		}

		if len(meta.Dependencies) > 0 {
			sb.WriteString("## Dependencies\n\n")
			for _, dep := range meta.Dependencies {
				sb.WriteString(fmt.Sprintf("- %s\n", dep))
			}
			sb.WriteString("\n")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, meta, nil
	}
}
