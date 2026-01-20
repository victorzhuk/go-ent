package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent/background"
)

type AgentBgKillInput struct {
	AgentID string `json:"agent_id"`
}

type AgentBgKillResponse struct {
	AgentID string `json:"agent_id"`
	Role    string `json:"role"`
	Model   string `json:"model"`
	Task    string `json:"task"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func registerAgentBgKill(s *mcp.Server, manager *background.Manager) {
	tool := &mcp.Tool{
		Name:        "go_ent_agent_kill",
		Description: "Terminate a running background agent",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"agent_id": map[string]any{
					"type":        "string",
					"description": "ID of the agent to terminate",
				},
			},
			"required": []string{"agent_id"},
		},
	}

	handler := makeAgentBgKillHandler(manager)
	mcp.AddTool(s, tool, handler)
}

func makeAgentBgKillHandler(manager *background.Manager) func(context.Context, *mcp.CallToolRequest, AgentBgKillInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AgentBgKillInput) (*mcp.CallToolResult, any, error) {
		if input.AgentID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Error: agent_id is required",
				}},
			}, nil, fmt.Errorf("agent_id is required")
		}

		err := manager.Kill(ctx, input.AgentID)
		if err != nil {
			if err == background.ErrAgentNotFound {
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("Agent not found: %s", input.AgentID),
					}},
				}, nil, fmt.Errorf("agent not found: %s", input.AgentID)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Failed to kill agent: %v", err),
				}},
			}, nil, fmt.Errorf("kill agent: %w", err)
		}

		agent, getErr := manager.Get(input.AgentID)
		if getErr != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Agent killed but failed to retrieve details: %v", getErr),
				}},
			}, nil, fmt.Errorf("get agent: %w", getErr)
		}

		response := AgentBgKillResponse{
			AgentID: agent.ID,
			Role:    agent.Role,
			Model:   agent.Model,
			Task:    agent.Task,
			Status:  string(agent.Status),
			Message: fmt.Sprintf("Agent %s terminated successfully", agent.ID),
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
			}, nil, fmt.Errorf("marshal response: %w", err)
		}

		msg := buildAgentBgKillMessage(agent, string(data))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}
}

func buildAgentBgKillMessage(agent *background.Agent, data string) string {
	msg := "🛑 Background Agent Terminated\n\n"
	msg += fmt.Sprintf("**Agent ID**: `%s`\n\n", agent.ID)
	msg += fmt.Sprintf("**Status**: %s\n", agent.Status)
	msg += fmt.Sprintf("**Role**: %s\n", agent.Role)
	msg += fmt.Sprintf("**Model**: %s\n", agent.Model)
	msg += fmt.Sprintf("**Task**: %s\n\n", agent.Task)
	msg += "---\n\n"
	msg += "**Full Details**:\n\n"
	msg += fmt.Sprintf("```json\n%s\n```", data)

	return msg
}
