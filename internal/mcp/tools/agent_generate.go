package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/generator"
)

type AgentGenerateInput struct {
	Agent   string   `json:"agent,omitempty"`
	Targets []string `json:"targets,omitempty"`
}

func registerAgentGenerate(s *mcp.Server, toolRegistry *ToolRegistry, srcDir string) {
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
	toolRegistry.Register("agent_generate", tool.Description, "agents")
}

func agentGenerateHandler(srcDir string) func(ctx context.Context, req *mcp.CallToolRequest, input AgentGenerateInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AgentGenerateInput) (*mcp.CallToolResult, any, error) {
		targetNames := input.Targets
		if len(targetNames) == 0 {
			targetNames = []string{"claude", "opencode"}
		}

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

		cfg, err := config.LoadCombinedRuntimeConfig(".", targetNames)
		if err != nil {
			return nil, nil, fmt.Errorf("load runtime config: %w", err)
		}
		for _, name := range targetNames {
			if err := config.ValidateForRuntime(cfg, name); err != nil {
				return nil, nil, fmt.Errorf("invalid %s config: %w", name, err)
			}
		}

		gen := generator.New(srcDir, cfg, targets...)

		var msg string
		if input.Agent != "" {
			if err := gen.GenerateAgent(input.Agent); err != nil {
				return nil, nil, fmt.Errorf("generate agent: %w", err)
			}
			msg = fmt.Sprintf("✓ Generated agent: %s\nTargets: %v", input.Agent, targetNames)
		} else {
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
