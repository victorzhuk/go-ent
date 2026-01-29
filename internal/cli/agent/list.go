package agent

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/generator"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available agents",
		Long: `List all available agent definitions from embedded sources.

This command shows all agents that can be generated, along with their
descriptions and configured models.

Examples:
  ent agent list
`,
		RunE: runList,
	}

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	// List agents from embedded sources
	agents, err := generator.ListAgents("agents")
	if err != nil {
		return fmt.Errorf("list agents: %w", err)
	}

	if len(agents) == 0 {
		fmt.Println("No agents found")
		return nil
	}

	fmt.Printf("Available agents (%d):\n\n", len(agents))

	for _, name := range agents {
		// Load agent to get description
		agent, _, err := generator.LoadAgentSource("agents", name)
		if err != nil {
			fmt.Printf("  • %s (error loading: %v)\n", name, err)
			continue
		}

		// Format output
		nameDisplay := filepath.Base(name)
		if len(nameDisplay) > 20 {
			nameDisplay = nameDisplay[:17] + "..."
		}

		modelInfo := fmt.Sprintf("claude:%s opencode:%s",
			agent.Model.Claude,
			agent.Model.OpenCode)

		// Truncate description if too long
		desc := agent.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}

		fmt.Printf("  %-20s %-35s %s\n", nameDisplay, modelInfo, desc)
	}

	fmt.Println()
	return nil
}
