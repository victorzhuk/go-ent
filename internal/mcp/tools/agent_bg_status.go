package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent/background"
)

type AgentBgStatusInput struct {
	AgentID string `json:"agent_id"`
}

type AgentBgStatusResponse struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	Model       string `json:"model"`
	Task        string `json:"task"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	StartedAt   string `json:"started_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
	Duration    string `json:"duration"`
	Output      string `json:"output,omitempty"`
	Error       string `json:"error,omitempty"`
}

func registerAgentBgStatus(s *mcp.Server, manager *background.Manager) {
	tool := &mcp.Tool{
		Name:        "go_ent_agent_status",
		Description: "Check the status, progress, and output of a specific background agent",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"agent_id": map[string]any{
					"type":        "string",
					"description": "ID of the agent to check",
				},
			},
			"required": []string{"agent_id"},
		},
	}

	handler := makeAgentBgStatusHandler(manager)
	mcp.AddTool(s, tool, handler)
}

func makeAgentBgStatusHandler(manager *background.Manager) func(context.Context, *mcp.CallToolRequest, AgentBgStatusInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AgentBgStatusInput) (*mcp.CallToolResult, any, error) {
		if input.AgentID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: "Error: agent_id is required",
				}},
			}, nil, fmt.Errorf("agent_id is required")
		}

		agent, err := manager.Get(input.AgentID)
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
					Text: fmt.Sprintf("Failed to get agent: %v", err),
				}},
			}, nil, fmt.Errorf("get agent: %w", err)
		}

		snap := agent.GetSnapshot()
		response := buildAgentBgStatusResponse(snap, agent)

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
			}, nil, fmt.Errorf("marshal response: %w", err)
		}

		msg := buildAgentBgStatusMessage(snap, string(data))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}
}

func buildAgentBgStatusResponse(snap background.Snapshot, agent *background.Agent) AgentBgStatusResponse {
	response := AgentBgStatusResponse{
		ID:        snap.ID,
		Role:      snap.Role,
		Model:     snap.Model,
		Task:      snap.Task,
		Status:    string(snap.Status),
		CreatedAt: snap.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Duration:  agent.Duration().Round(time.Millisecond).String(),
	}

	if !snap.StartedAt.IsZero() {
		response.StartedAt = snap.StartedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	if !snap.CompletedAt.IsZero() {
		response.CompletedAt = snap.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	if snap.Output != "" {
		response.Output = snap.Output
	}

	if snap.Error != nil {
		response.Error = snap.Error.Error()
	}

	return response
}

func buildAgentBgStatusMessage(snap background.Snapshot, data string) string {
	var statusIcon string
	switch snap.Status {
	case background.StatusPending:
		statusIcon = "⏳"
	case background.StatusRunning:
		statusIcon = "▶️"
	case background.StatusCompleted:
		statusIcon = "✅"
	case background.StatusFailed:
		statusIcon = "❌"
	case background.StatusKilled:
		statusIcon = "🛑"
	default:
		statusIcon = "❓"
	}

	duration := time.Duration(0)
	if !snap.StartedAt.IsZero() {
		if !snap.CompletedAt.IsZero() {
			duration = snap.CompletedAt.Sub(snap.StartedAt)
		} else {
			duration = time.Since(snap.StartedAt)
		}
	}

	msg := fmt.Sprintf("# Background Agent Status %s\n\n", statusIcon)
	msg += fmt.Sprintf("**Agent ID**: `%s`\n\n", snap.ID)
	msg += fmt.Sprintf("**Status**: %s\n", snap.Status)
	msg += fmt.Sprintf("**Role**: %s\n", snap.Role)
	msg += fmt.Sprintf("**Model**: %s\n", snap.Model)
	msg += fmt.Sprintf("**Duration**: %s\n\n", duration.Round(time.Millisecond))

	msg += "## Task\n\n"
	msg += fmt.Sprintf("%s\n\n", snap.Task)

	msg += "## Timeline\n\n"
	msg += fmt.Sprintf("- **Created**: %s\n", snap.CreatedAt.Format("2006-01-02 15:04:05"))
	if !snap.StartedAt.IsZero() {
		msg += fmt.Sprintf("- **Started**: %s\n", snap.StartedAt.Format("2006-01-02 15:04:05"))
	}
	if !snap.CompletedAt.IsZero() {
		msg += fmt.Sprintf("- **Completed**: %s\n", snap.CompletedAt.Format("2006-01-02 15:04:05"))
	}
	msg += "\n"

	if snap.Output != "" {
		msg += "## Output\n\n"
		msg += "```\n"
		msg += snap.Output
		msg += "\n```\n\n"
	}

	if snap.Error != nil {
		msg += "## Error\n\n"
		msg += "```\n"
		msg += snap.Error.Error()
		msg += "\n```\n\n"
	}

	msg += "---\n\n"
	msg += "**Full Details**:\n\n"
	msg += fmt.Sprintf("```json\n%s\n```", data)

	return msg
}
