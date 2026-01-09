package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type AgentStatusInput struct {
	Path string `json:"path,omitempty"`
}

func registerAgentStatus(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "agent_status",
		Description: "Get agent execution system status and capabilities",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Optional path to project directory",
				},
			},
		},
	}

	mcp.AddTool(s, tool, agentStatusHandler)
}

func agentStatusHandler(ctx context.Context, req *mcp.CallToolRequest, input AgentStatusInput) (*mcp.CallToolResult, any, error) {
	msg := "# Agent Execution Status\n\n"
	msg += "## Available Capabilities\n\n"
	msg += "✅ **Agent Selection** - Automatic role and model selection based on task complexity\n"
	msg += "✅ **Skill Matching** - Context-aware skill activation\n"
	msg += "✅ **Complexity Analysis** - Task complexity evaluation\n\n"

	msg += "## Agent Roles\n\n"
	msg += "- `architect` - System design and architecture (Opus)\n"
	msg += "- `senior` - Complex implementations (Opus/Sonnet)\n"
	msg += "- `developer` - Standard development tasks (Sonnet/Haiku)\n"
	msg += "- `ops` - DevOps and infrastructure (Sonnet)\n"
	msg += "- `reviewer` - Code review and quality checks (Opus)\n\n"

	msg += "## Model Selection\n\n"
	msg += "Models are automatically selected based on role and complexity:\n"
	msg += "- **Opus**: Architectural and complex tasks\n"
	msg += "- **Sonnet**: Standard development work\n"
	msg += "- **Haiku**: Simple, routine tasks\n\n"

	msg += "## Usage\n\n"
	msg += "Use `agent_execute` to select and configure an agent for a task.\n"
	msg += "The system analyzes task complexity and matches appropriate skills.\n"

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}
