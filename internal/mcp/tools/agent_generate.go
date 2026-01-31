package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/genconfig"
	"github.com/victorzhuk/go-ent/internal/generator"
)

type AgentGenerateInput struct {
	Agent   string   `json:"agent,omitempty"`   // Agent name (optional, generates all if empty)
	Targets []string `json:"targets,omitempty"` // Target platforms (claude, opencode), default: both
}

func registerAgentGenerate(s *mcp.Server, srcDir string) {
	tool := &mcp.Tool{
		Name:        "agent_generate",
		Description: "Generate agent configurations for target platforms from meta definitions",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"agent": map[string]any{
					"type":        "string",
					"description": "Agent name to generate (empty generates all agents)",
				},
				"targets": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "string",
						"enum": []string{"claude", "opencode"},
					},
					"description": "Target platforms (default: both)",
				},
			},
		},
	}

	mcp.AddTool(s, tool, agentGenerateHandler(srcDir))
}

func agentGenerateHandler(srcDir string) func(ctx context.Context, req *mcp.CallToolRequest, input AgentGenerateInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AgentGenerateInput) (*mcp.CallToolResult, any, error) {
		// Default targets to both if not specified
		targetNames := input.Targets
		if len(targetNames) == 0 {
			targetNames = []string{"claude", "opencode"}
		}

		// Load config
		cfg, err := genconfig.Load("ent.yaml")
		if err != nil {
			return nil, nil, fmt.Errorf("load config: %w", err)
		}

		// Create targets
		targets := make([]generator.Target, 0, len(targetNames))
		for _, name := range targetNames {
			switch name {
			case "claude":
				targets = append(targets, generator.NewClaudeTarget(".claude/agents"))
			case "opencode":
				targets = append(targets, generator.NewOpenCodeTarget(".opencode/agents"))
			default:
				return nil, nil, fmt.Errorf("unknown target: %s", name)
			}
		}

		// Create generator
		gen := generator.New(srcDir, cfg, targets...)

		// Generate
		var msg string
		if input.Agent != "" {
			// Generate single agent
			if err := gen.GenerateAgent(input.Agent); err != nil {
				return nil, nil, fmt.Errorf("generate agent: %w", err)
			}
			msg = fmt.Sprintf("✓ Generated agent: %s\nTargets: %v", input.Agent, targetNames)
		} else {
			// Generate all agents
			if err := gen.GenerateAll(); err != nil {
				return nil, nil, fmt.Errorf("generate all: %w", err)
			}

			agents, _ := generator.ListAgents(srcDir)
			msg = fmt.Sprintf("✓ Generated %d agent(s)\nTargets: %v", len(agents), targetNames)
		}

		return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: msg}},
			}, map[string]any{
				"agent":   input.Agent,
				"targets": targetNames,
				"status":  "generated",
			}, nil
	}
}
