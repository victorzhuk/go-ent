package tools

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/domain"
)

type AgentListInput struct{}

type AgentInfo struct {
	Role         string   `json:"role"`
	Description  string   `json:"description"`
	Model        string   `json:"model"`
	Capabilities []string `json:"capabilities"`
}

func registerAgentList(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "agent_list",
		Description: "List all available agent roles with their capabilities and models",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}

	mcp.AddTool(s, tool, agentListHandler)
}

func agentListHandler(ctx context.Context, req *mcp.CallToolRequest, input AgentListInput) (*mcp.CallToolResult, any, error) {
	agents := []AgentInfo{
		{
			Role:        domain.AgentRoleProduct.String(),
			Description: "User needs, requirements, and product decisions. Focuses on understanding what needs to be built and why.",
			Model:       "opus",
			Capabilities: []string{
				"Requirements analysis",
				"User story definition",
				"Feature prioritization",
				"Product vision",
			},
		},
		{
			Role:        domain.AgentRoleArchitect.String(),
			Description: "System design, architecture, and technical decisions. Responsible for high-level design, technology choices, and architectural patterns.",
			Model:       "opus",
			Capabilities: []string{
				"System architecture",
				"Technology selection",
				"Design patterns",
				"Architectural trade-offs",
			},
		},
		{
			Role:        domain.AgentRoleSenior.String(),
			Description: "Complex implementation, debugging, and code review. Takes on challenging technical problems requiring deep expertise.",
			Model:       "opus/sonnet",
			Capabilities: []string{
				"Complex implementation",
				"Debugging",
				"Performance optimization",
				"Code review",
			},
		},
		{
			Role:        domain.AgentRoleDeveloper.String(),
			Description: "Standard implementation and testing. Executes well-defined tasks and writes tests for new functionality.",
			Model:       "sonnet/haiku",
			Capabilities: []string{
				"Code implementation",
				"Unit testing",
				"Bug fixes",
				"Feature development",
			},
		},
		{
			Role:        domain.AgentRoleReviewer.String(),
			Description: "Code quality and standards enforcement. Reviews code for correctness, style, security, and best practices.",
			Model:       "opus",
			Capabilities: []string{
				"Code review",
				"Quality assurance",
				"Security analysis",
				"Best practices enforcement",
			},
		},
		{
			Role:        domain.AgentRoleOps.String(),
			Description: "Deployment, monitoring, and production issues. Manages infrastructure, observability, and operational concerns.",
			Model:       "sonnet",
			Capabilities: []string{
				"Deployment automation",
				"Infrastructure management",
				"Monitoring setup",
				"Production support",
			},
		},
	}

	var sb strings.Builder
	sb.WriteString("# Available Agent Roles\n\n")
	sb.WriteString("The following specialized agents are available for task execution:\n\n")

	for _, agent := range agents {
		sb.WriteString("## ")
		sb.WriteString(agent.Role)
		sb.WriteString("\n\n")

		sb.WriteString("**Model**: ")
		sb.WriteString(agent.Model)
		sb.WriteString("\n\n")

		sb.WriteString("**Description**: ")
		sb.WriteString(agent.Description)
		sb.WriteString("\n\n")

		sb.WriteString("**Capabilities**:\n")
		for _, cap := range agent.Capabilities {
			sb.WriteString("- ")
			sb.WriteString(cap)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n\n")
	sb.WriteString("**Usage**: Use `agent_execute` to automatically select the optimal agent based on task complexity and type.\n")

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
	}, agents, nil
}
