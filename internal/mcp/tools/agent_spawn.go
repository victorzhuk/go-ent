package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent/background"
)

type AgentSpawnInput struct {
	Task    string `json:"task"`
	Role    string `json:"role,omitempty"`
	Model   string `json:"model,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

type AgentSpawnResponse struct {
	AgentID   string `json:"agent_id"`
	Role      string `json:"role"`
	Model     string `json:"model"`
	Task      string `json:"task"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	Message   string `json:"message"`
}

func registerAgentSpawn(s *mcp.Server, manager *background.Manager) {
	tool := &mcp.Tool{
		Name:        "go_ent_agent_spawn",
		Description: "Spawn a background agent to execute a task asynchronously",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"task": map[string]any{
					"type":        "string",
					"description": "Task description for the agent to execute",
				},
				"role": map[string]any{
					"type":        "string",
					"description": "Optional: Override agent role (architect, senior, developer, ops, reviewer)",
				},
				"model": map[string]any{
					"type":        "string",
					"description": "Optional: Override model selection (opus, sonnet, haiku)",
				},
				"timeout": map[string]any{
					"type":        "integer",
					"description": "Optional: Maximum execution duration in seconds",
				},
			},
			"required": []string{"task"},
		},
	}

	handler := makeAgentSpawnHandler(manager)
	mcp.AddTool(s, tool, handler)
}

func makeAgentSpawnHandler(manager *background.Manager) func(context.Context, *mcp.CallToolRequest, AgentSpawnInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AgentSpawnInput) (*mcp.CallToolResult, any, error) {
		if input.Task == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Error: task is required",
				}},
			}, nil, fmt.Errorf("task is required")
		}

		opts := background.SpawnOpts{
			Role:    input.Role,
			Model:   input.Model,
			Timeout: input.Timeout,
		}

		agent, err := manager.Spawn(ctx, input.Task, opts)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Failed to spawn agent: %v", err),
				}},
			}, nil, fmt.Errorf("spawn agent: %w", err)
		}

		response := AgentSpawnResponse{
			AgentID:   agent.ID,
			Role:      agent.Role,
			Model:     agent.Model,
			Task:      agent.Task,
			Status:    string(agent.Status),
			CreatedAt: agent.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Message:   fmt.Sprintf("Agent %s spawned successfully with role %s and model %s", agent.ID, agent.Role, agent.Model),
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
			}, nil, fmt.Errorf("marshal response: %w", err)
		}

		msg := fmt.Sprintf("✅ Background Agent Spawned\n\n```json\n%s\n```\n\n", string(data))
		msg += "**Next Steps:**\n"
		msg += fmt.Sprintf("- Agent ID: `%s`\n", agent.ID)
		msg += fmt.Sprintf("- Status: %s\n", agent.Status)
		msg += fmt.Sprintf("- Role: %s\n", agent.Role)
		msg += fmt.Sprintf("- Model: %s\n", agent.Model)
		msg += "\nUse `go_ent_agent_status` or `go_ent_agent_list` to monitor progress.\n"

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}
}
