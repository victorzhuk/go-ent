package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/domain"
)

type AgentDelegateInput struct {
	FromRole string `json:"from_role"`
	ToRole   string `json:"to_role"`
	Task     string `json:"task"`
	Reason   string `json:"reason,omitempty"`
}

type AgentDelegateResponse struct {
	FromRole     string `json:"from_role"`
	ToRole       string `json:"to_role"`
	Task         string `json:"task"`
	Reason       string `json:"reason"`
	DelegationID string `json:"delegation_id"`
}

func registerAgentDelegate(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "agent_delegate",
		Description: "Delegate a task from one agent role to another (multi-agent workflow)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"from_role": map[string]any{
					"type":        "string",
					"description": "Source agent role (product, architect, senior, developer, reviewer, ops)",
				},
				"to_role": map[string]any{
					"type":        "string",
					"description": "Target agent role to delegate to",
				},
				"task": map[string]any{
					"type":        "string",
					"description": "Task description to delegate",
				},
				"reason": map[string]any{
					"type":        "string",
					"description": "Reason for delegation (optional)",
				},
			},
			"required": []string{"from_role", "to_role", "task"},
		},
	}

	mcp.AddTool(s, tool, agentDelegateHandler)
}

func agentDelegateHandler(ctx context.Context, req *mcp.CallToolRequest, input AgentDelegateInput) (*mcp.CallToolResult, any, error) {
	fromRole := domain.AgentRole(input.FromRole)
	if !fromRole.Valid() {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("❌ Invalid source role: %s\n\nValid roles: product, architect, senior, developer, reviewer, ops", input.FromRole),
			}},
		}, nil, nil
	}

	toRole := domain.AgentRole(input.ToRole)
	if !toRole.Valid() {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("❌ Invalid target role: %s\n\nValid roles: product, architect, senior, developer, reviewer, ops", input.ToRole),
			}},
		}, nil, nil
	}

	if fromRole == toRole {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: "❌ Cannot delegate to the same role",
			}},
		}, nil, nil
	}

	if input.Task == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: "❌ Task description is required",
			}},
		}, nil, nil
	}

	reason := input.Reason
	if reason == "" {
		reason = fmt.Sprintf("%s requires expertise from %s", fromRole, toRole)
	}

	delegationID := fmt.Sprintf("%s-to-%s", fromRole, toRole)

	response := AgentDelegateResponse{
		FromRole:     fromRole.String(),
		ToRole:       toRole.String(),
		Task:         input.Task,
		Reason:       reason,
		DelegationID: delegationID,
	}

	msg := fmt.Sprintf("✅ **Task Delegated**\n\n")
	msg += fmt.Sprintf("**From**: %s agent\n", fromRole)
	msg += fmt.Sprintf("**To**: %s agent\n\n", toRole)
	msg += fmt.Sprintf("**Task**: %s\n\n", input.Task)
	msg += fmt.Sprintf("**Reason**: %s\n\n", reason)
	msg += fmt.Sprintf("**Delegation ID**: `%s`\n\n", delegationID)
	msg += fmt.Sprintf("---\n\n")
	msg += fmt.Sprintf("**Next Steps**:\n")
	msg += fmt.Sprintf("- The %s agent will handle this task\n", toRole)
	msg += fmt.Sprintf("- Use `agent_execute` to configure the agent\n")

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, response, nil
}
