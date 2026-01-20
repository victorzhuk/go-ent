package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent/background"
)

type AgentBgListInput struct {
	Status string `json:"status,omitempty"`
}

type AgentBgListResponse struct {
	Agents       []AgentBgStatusResponse `json:"agents"`
	TotalCount   int                     `json:"total_count"`
	StatusFilter string                  `json:"status_filter,omitempty"`
	Counts       map[string]int          `json:"counts"`
}

func registerAgentBgList(s *mcp.Server, manager *background.Manager) {
	validStatuses := []string{
		string(background.StatusPending),
		string(background.StatusRunning),
		string(background.StatusCompleted),
		string(background.StatusFailed),
		string(background.StatusKilled),
	}

	tool := &mcp.Tool{
		Name:        "go_ent_agent_list",
		Description: "List all background agents with optional status filtering. Shows agent details including status, role, model, task, and timestamps.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{
					"type":        "string",
					"description": fmt.Sprintf("Filter by agent status. Valid values: %s. Leave empty to list all agents.", strings.Join(validStatuses, ", ")),
					"enum":        validStatuses,
				},
			},
		},
	}

	handler := makeAgentBgListHandler(manager)
	mcp.AddTool(s, tool, handler)
}

func makeAgentBgListHandler(manager *background.Manager) func(context.Context, *mcp.CallToolRequest, AgentBgListInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AgentBgListInput) (*mcp.CallToolResult, any, error) {
		statusFilter := input.Status

		if statusFilter != "" && !background.Status(statusFilter).Valid() {
			validStatuses := []string{
				string(background.StatusPending),
				string(background.StatusRunning),
				string(background.StatusCompleted),
				string(background.StatusFailed),
				string(background.StatusKilled),
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error: invalid status '%s'. Valid values: %s", statusFilter, strings.Join(validStatuses, ", ")),
				}},
			}, nil, fmt.Errorf("invalid status: %s", statusFilter)
		}

		agents := manager.List(background.Status(statusFilter))

		response := AgentBgListResponse{
			Agents:       make([]AgentBgStatusResponse, 0, len(agents)),
			TotalCount:   len(agents),
			StatusFilter: statusFilter,
			Counts:       make(map[string]int),
		}

		for _, agent := range agents {
			snap := agent.GetSnapshot()
			agentResp := buildAgentBgStatusResponse(snap, agent)
			response.Agents = append(response.Agents, agentResp)

			response.Counts[agentResp.Status]++
		}

		if statusFilter == "" {
			response.Counts["pending"] = manager.CountByStatus(background.StatusPending)
			response.Counts["running"] = manager.CountByStatus(background.StatusRunning)
			response.Counts["completed"] = manager.CountByStatus(background.StatusCompleted)
			response.Counts["failed"] = manager.CountByStatus(background.StatusFailed)
			response.Counts["killed"] = manager.CountByStatus(background.StatusKilled)
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
			}, nil, fmt.Errorf("marshal response: %w", err)
		}

		msg := buildAgentBgListMessage(response, string(data))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}
}

func buildAgentBgListMessage(response AgentBgListResponse, data string) string {
	var sb strings.Builder

	sb.WriteString("# Background Agents\n\n")

	if response.StatusFilter != "" {
		sb.WriteString(fmt.Sprintf("Showing %d agent(s) with status: **%s**\n\n", response.TotalCount, response.StatusFilter))
	} else {
		sb.WriteString(fmt.Sprintf("Showing %d total agent(s)\n\n", response.TotalCount))
	}

	if len(response.Counts) > 0 {
		sb.WriteString("## Summary\n\n")
		if response.StatusFilter != "" {
			sb.WriteString(fmt.Sprintf("- **%s**: %d\n", response.StatusFilter, response.Counts[response.StatusFilter]))
		} else {
			sb.WriteString("- **pending**: ")
			sb.WriteString(fmt.Sprintf("%d\n", response.Counts["pending"]))
			sb.WriteString("- **running**: ")
			sb.WriteString(fmt.Sprintf("%d\n", response.Counts["running"]))
			sb.WriteString("- **completed**: ")
			sb.WriteString(fmt.Sprintf("%d\n", response.Counts["completed"]))
			sb.WriteString("- **failed**: ")
			sb.WriteString(fmt.Sprintf("%d\n", response.Counts["failed"]))
			sb.WriteString("- **killed**: ")
			sb.WriteString(fmt.Sprintf("%d\n", response.Counts["killed"]))
		}
		sb.WriteString("\n")
	}

	if len(response.Agents) > 0 {
		sb.WriteString("## Agents\n\n")
		for _, agent := range response.Agents {
			var statusIcon string
			switch agent.Status {
			case string(background.StatusPending):
				statusIcon = "⏳"
			case string(background.StatusRunning):
				statusIcon = "▶️"
			case string(background.StatusCompleted):
				statusIcon = "✅"
			case string(background.StatusFailed):
				statusIcon = "❌"
			case string(background.StatusKilled):
				statusIcon = "🛑"
			default:
				statusIcon = "❓"
			}

			sb.WriteString(fmt.Sprintf("### %s %s\n\n", statusIcon, agent.ID))
			sb.WriteString(fmt.Sprintf("- **Status**: %s\n", agent.Status))
			sb.WriteString(fmt.Sprintf("- **Role**: %s\n", agent.Role))
			sb.WriteString(fmt.Sprintf("- **Model**: %s\n", agent.Model))
			sb.WriteString(fmt.Sprintf("- **Duration**: %s\n", agent.Duration))
			sb.WriteString(fmt.Sprintf("- **Created**: %s\n", agent.CreatedAt))

			if agent.StartedAt != "" {
				sb.WriteString(fmt.Sprintf("- **Started**: %s\n", agent.StartedAt))
			}

			if agent.CompletedAt != "" {
				sb.WriteString(fmt.Sprintf("- **Completed**: %s\n", agent.CompletedAt))
			}

			sb.WriteString("\n**Task**:\n")
			sb.WriteString(agent.Task)
			sb.WriteString("\n\n")

			if agent.Output != "" {
				sb.WriteString("**Output**:\n```\n")
				sb.WriteString(agent.Output)
				sb.WriteString("\n```\n\n")
			}

			if agent.Error != "" {
				sb.WriteString("**Error**:\n```\n")
				sb.WriteString(agent.Error)
				sb.WriteString("\n```\n\n")
			}

			sb.WriteString("---\n\n")
		}
	} else {
		sb.WriteString("No agents found.\n\n")
	}

	sb.WriteString("**Full Details**:\n\n")
	sb.WriteString(fmt.Sprintf("```json\n%s\n```", data))

	return sb.String()
}
