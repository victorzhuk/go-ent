package agent

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/generator"
)

func newInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <agent-name>",
		Short: "Display detailed information about an agent",
		Long: `Display detailed information about an agent definition.

Shows the agent's name, description, model tier, role, skills, dependencies,
and other configuration details.

Examples:
  ent agent info planner
  ent agent info debugger-fast
`,
		Args: cobra.ExactArgs(1),
		RunE: runInfo,
	}

	return cmd
}

func runInfo(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Load agent from meta format
	metaAgent, _, err := generator.LoadAgentMetaSource("agents/meta", name)
	if err != nil {
		return fmt.Errorf("load agent %s: %w", name, err)
	}

	// Display agent information
	fmt.Printf("Agent: %s\n", metaAgent.Name)
	fmt.Printf("Description: %s\n", metaAgent.Description)
	fmt.Printf("\n")

	// Model information
	modelDisplay := metaAgent.Model
	switch metaAgent.Model {
	case "fast":
		modelDisplay = "fast (haiku/quick tasks)"
	case "main":
		modelDisplay = "main (sonnet/standard tasks)"
	case "heavy":
		modelDisplay = "heavy (opus/complex tasks)"
	}
	fmt.Printf("Model: %s\n", modelDisplay)

	// Role
	if metaAgent.Role != "" {
		fmt.Printf("Role: %s\n", metaAgent.Role)
	}

	// Complexity
	if metaAgent.Complexity != "" {
		fmt.Printf("Complexity: %s\n", metaAgent.Complexity)
	}

	// Color
	if metaAgent.Color != "" {
		fmt.Printf("Color: %s\n", metaAgent.Color)
	}
	fmt.Printf("\n")

	// Skills
	if len(metaAgent.Skills) > 0 {
		fmt.Printf("Skills:\n")
		for _, skill := range metaAgent.Skills {
			fmt.Printf("  • %s\n", skill)
		}
		fmt.Printf("\n")
	}

	// Tool presets
	if len(metaAgent.ToolPresets) > 0 {
		fmt.Printf("Tool Presets:\n")
		for _, preset := range metaAgent.ToolPresets {
			fmt.Printf("  • %s\n", preset)
		}
		fmt.Printf("\n")
	}

	// Disallowed tool presets
	if len(metaAgent.DisallowedToolPresets) > 0 {
		fmt.Printf("Disallowed Tool Presets:\n")
		for _, preset := range metaAgent.DisallowedToolPresets {
			fmt.Printf("  • %s\n", preset)
		}
		fmt.Printf("\n")
	}

	// Dependencies
	if len(metaAgent.Dependencies) > 0 {
		fmt.Printf("Dependencies:\n")
		for _, dep := range metaAgent.Dependencies {
			fmt.Printf("  • %s\n", dep)
		}
		fmt.Printf("\n")
	}

	// Prompts
	if metaAgent.Prompts.Main != "" || len(metaAgent.Prompts.Shared) > 0 {
		fmt.Printf("Prompts:\n")
		if metaAgent.Prompts.Main != "" {
			fmt.Printf("  Main: %s\n", metaAgent.Prompts.Main)
		}
		if len(metaAgent.Prompts.Shared) > 0 {
			fmt.Printf("  Shared: %s\n", strings.Join(metaAgent.Prompts.Shared, ", "))
		}
		fmt.Printf("\n")
	}

	return nil
}
