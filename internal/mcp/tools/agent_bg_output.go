package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent/background"
)

type AgentBgOutputInput struct {
	AgentID       string `json:"agent_id"`
	FilterPattern string `json:"filter_pattern,omitempty"`
}

type AgentBgOutputResponse struct {
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
	Output  string `json:"output"`
	Filter  string `json:"filter,omitempty"`
}

func registerAgentBgOutput(s *mcp.Server, manager *background.Manager) {
	tool := &mcp.Tool{
		Name:        "go_ent_agent_output",
		Description: "Retrieve the output of a background agent with optional regex filtering",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"agent_id": map[string]any{
					"type":        "string",
					"description": "ID of the agent to retrieve output from",
				},
				"filter_pattern": map[string]any{
					"type":        "string",
					"description": "Optional: Regex pattern to filter output (e.g., 'ERROR:.*' for only error lines)",
				},
			},
			"required": []string{"agent_id"},
		},
	}

	handler := makeAgentBgOutputHandler(manager)
	mcp.AddTool(s, tool, handler)
}

func makeAgentBgOutputHandler(manager *background.Manager) func(context.Context, *mcp.CallToolRequest, AgentBgOutputInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AgentBgOutputInput) (*mcp.CallToolResult, any, error) {
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

		output, err := getAgentOutput(snap, input.FilterPattern)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Failed to filter output: %v", err),
				}},
			}, nil, fmt.Errorf("filter output: %w", err)
		}

		response := AgentBgOutputResponse{
			AgentID: snap.ID,
			Status:  string(snap.Status),
			Output:  output,
		}

		if input.FilterPattern != "" {
			response.Filter = input.FilterPattern
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
			}, nil, fmt.Errorf("marshal response: %w", err)
		}

		msg := buildAgentBgOutputMessage(snap, input.FilterPattern, string(data))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, response, nil
	}
}

func getAgentOutput(snap background.Snapshot, pattern string) (string, error) {
	switch snap.Status {
	case background.StatusCompleted, background.StatusFailed:
		if snap.Output == "" {
			return "", nil
		}

		if pattern == "" {
			return snap.Output, nil
		}

		re, err := regexp.Compile(pattern)
		if err != nil {
			return "", fmt.Errorf("invalid regex pattern: %w", err)
		}

		matches := re.FindAllString(snap.Output, -1)
		if len(matches) == 0 {
			return "", nil
		}

		var result string
		for _, match := range matches {
			result += match
		}

		return result, nil

	case background.StatusRunning:
		return "", nil

	default:
		return "", nil
	}
}

func buildAgentBgOutputMessage(snap background.Snapshot, pattern, data string) string {
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

	msg := fmt.Sprintf("# Agent Output %s\n\n", statusIcon)
	msg += fmt.Sprintf("**Agent ID**: `%s`\n\n", snap.ID)
	msg += fmt.Sprintf("**Status**: %s\n", snap.Status)

	if pattern != "" {
		msg += fmt.Sprintf("**Filter Pattern**: `%s`\n\n", pattern)
	} else {
		msg += "\n"
	}

	msg += "## Output\n\n"

	if snap.Status == background.StatusRunning {
		msg += "Agent is still running. Output will be available when the agent completes.\n\n"
	} else if snap.Output == "" {
		msg += "No output available.\n\n"
	} else {
		msg += "```\n"
		if pattern != "" {
			msg += "(Filtered output)\n\n"
		}
		msg += snap.Output
		msg += "\n```\n\n"
	}

	msg += "---\n\n"
	msg += "**Full Details**:\n\n"
	msg += fmt.Sprintf("```json\n%s\n```", data)

	return msg
}
