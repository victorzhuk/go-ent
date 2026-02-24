package agent

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/config"
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
  ent agent generate                    # Generate all agents for detected runtime
  ent agent generate --tools=claude     # Generate for Claude only
  ent agent generate --name=coder       # Generate specific agent
`,
		RunE: runGenerate,
	}

	cmd.Flags().StringSliceVar(&toolsFlag, "tools", nil, "Target tools (claude,opencode)")
	cmd.Flags().StringVar(&nameFlag, "name", "", "Generate specific agent by name")

	return cmd
}

func runGenerate(cmd *cobra.Command, args []string) error {
	tools := toolsFlag
	if len(tools) == 0 {
		if detected := config.DetectRuntime("."); detected != "" {
			tools = []string{detected}
		}
	}

	if len(tools) == 0 {
		tools = []string{"claude"}
	}

	var targets []generator.Target
	for _, tool := range tools {
		switch tool {
		case "claude":
			targets = append(targets, generator.NewClaudeTarget(".claude/agents/ent"))
		case "opencode":
			targets = append(targets, generator.NewOpenCodeTarget(".opencode/agents/ent"))
		case "openspec":
			continue
		default:
			return fmt.Errorf("unknown tool: %s", tool)
		}
	}

	if len(targets) == 0 {
		return fmt.Errorf("no valid tools configured")
	}

	cfg := config.LoadCombinedRuntimeConfig(".", tools)

	gen := generator.New("agents/meta", cfg, targets...)

	if nameFlag != "" {
		if err := gen.GenerateAgent(nameFlag); err != nil {
			return fmt.Errorf("generate agent %s: %w", nameFlag, err)
		}
	} else {
		if err := gen.GenerateAll(); err != nil {
			return fmt.Errorf("generate: %w", err)
		}
	}

	fmt.Println("\n✅ Agent generation complete!")
	return nil
}
