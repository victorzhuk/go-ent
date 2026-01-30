package agent

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/genconfig"
	"github.com/victorzhuk/go-ent/internal/generator"
)

var (
	toolsFlag []string
	nameFlag  string
)

func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate agents for configured tools",
		Long: `Generate agent output for configured tools.

This command reads agent sources from pkg/agents/ (embedded in the binary)
and generates tool-specific output in .claude/agents/ or .opencode/agents/.

Examples:
  ent agent generate                    # Generate all agents for all configured tools
  ent agent generate --tools=claude     # Generate for Claude only
  ent agent generate --name=coder       # Generate specific agent
`,
		RunE: runGenerate,
	}

	cmd.Flags().StringSliceVar(&toolsFlag, "tools", nil, "Override tools from config (claude,opencode)")
	cmd.Flags().StringVar(&nameFlag, "name", "", "Generate specific agent by name")

	return cmd
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := genconfig.Load("ent.yaml")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Override tools if specified
	tools := cfg.Tools
	if len(toolsFlag) > 0 {
		tools = toolsFlag
	}

	// Build targets
	var targets []generator.Target
	for _, tool := range tools {
		switch tool {
		case "claude":
			targets = append(targets, generator.NewClaudeTarget(".claude/agents/ent"))
		case "opencode":
			targets = append(targets, generator.NewOpenCodeTarget(".opencode/agents/ent"))
		case "openspec":
			// OpenSpec is a workflow tool, not an agent generation target - skip
			continue
		default:
			return fmt.Errorf("unknown tool: %s", tool)
		}
	}

	if len(targets) == 0 {
		return fmt.Errorf("no valid tools configured")
	}

	// Run generator using meta format
	// Meta format is in agents/meta/ subdirectory
	gen := generator.New("agents/meta", cfg, targets...)

	if nameFlag != "" {
		// Generate specific agent
		if err := gen.GenerateAgent(nameFlag); err != nil {
			return fmt.Errorf("generate agent %s: %w", nameFlag, err)
		}
	} else {
		// Generate all agents
		if err := gen.GenerateAll(); err != nil {
			return fmt.Errorf("generate: %w", err)
		}
	}

	fmt.Println("\n✅ Agent generation complete!")
	return nil
}
