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

func registerAgentInfo(s *mcp.Server, toolRegistry *ToolRegistry, agentRegistry *agent.Registry) {
	tool := &mcp.Tool{
		Name:        "agent_info",
		Description: "Get detailed information about a specific agent",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Name of agent to get info for",
				},
			},
			"required": []string{"name"},
		},
	}

	mcp.AddTool(s, tool, agentInfoHandler(agentRegistry))
	toolRegistry.Register("agent_info", tool.Description, "agents")
}

func agentInfoHandler(agentRegistry *agent.Registry) func(ctx context.Context, req *mcp.CallToolRequest, input AgentInfoInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AgentInfoInput) (*mcp.CallToolResult, any, error) {
		meta, ok := agentRegistry.Get(input.Name)
		if !ok {
			return nil, nil, fmt.Errorf("agent not found: %s", input.Name)
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "# Agent: %s\n\n", meta.Name)
		fmt.Fprintf(&sb, "**Description**: %s\n\n", meta.Description)
		fmt.Fprintf(&sb, "**Model**: %s\n\n", meta.Model)

		if meta.Role != "" {
			fmt.Fprintf(&sb, "**Role**: %s\n\n", meta.Role)
		}

		if meta.Complexity != "" {
			fmt.Fprintf(&sb, "**Complexity**: %s\n\n", meta.Complexity)
		}

		if meta.Color != "" {
			fmt.Fprintf(&sb, "**Color**: %s\n\n", meta.Color)
		}

		if len(meta.Skills) > 0 {
			sb.WriteString("## Skills\n\n")
			for _, skill := range meta.Skills {
				fmt.Fprintf(&sb, "- %s\n", skill)
			}
			sb.WriteString("\n")
		}

		if len(meta.ToolPresets) > 0 {
			sb.WriteString("## Tool Presets\n\n")
			for _, preset := range meta.ToolPresets {
				fmt.Fprintf(&sb, "- %s\n", preset)
			}
			sb.WriteString("\n")
		}

		if len(meta.DisallowedToolPresets) > 0 {
			sb.WriteString("## Disallowed Tool Presets\n\n")
			for _, preset := range meta.DisallowedToolPresets {
				fmt.Fprintf(&sb, "- %s\n", preset)
			}
			sb.WriteString("\n")
		}

		if len(meta.Dependencies) > 0 {
			sb.WriteString("## Dependencies\n\n")
			for _, dep := range meta.Dependencies {
				fmt.Fprintf(&sb, "- %s\n", dep)
			}
			sb.WriteString("\n")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, meta, nil
	}
}
