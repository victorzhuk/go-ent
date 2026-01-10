package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/agent"
)

func newAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage agents",
		Long:  "List and inspect available agent roles and their capabilities",
	}

	cmd.AddCommand(newAgentListCmd())
	cmd.AddCommand(newAgentInfoCmd())

	return cmd
}

func newAgentListCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available agents",
		Long:  "Display a table of all available agent roles with their models and descriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			agentsPath := getAgentsPath()
			registry := agent.NewRegistry()

			if err := registry.Load(agentsPath); err != nil {
				return fmt.Errorf("load agents: %w", err)
			}

			agents := registry.All()
			if len(agents) == 0 {
				fmt.Fprintln(os.Stderr, "No agents found")
				return nil
			}

			switch format {
			case "table":
				return printAgentsTable(agents)
			case "detailed":
				return printAgentsDetailed(agents)
			default:
				return fmt.Errorf("unknown format: %s", format)
			}
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, detailed)")

	return cmd
}

func newAgentInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <name>",
		Short: "Show detailed information about an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			agentsPath := getAgentsPath()
			registry := agent.NewRegistry()

			if err := registry.Load(agentsPath); err != nil {
				return fmt.Errorf("load agents: %w", err)
			}

			meta, err := registry.Get(name)
			if err != nil {
				return fmt.Errorf("agent not found: %s", name)
			}

			return printAgentInfo(meta)
		},
	}
}

func getAgentsPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "plugins/go-ent/agents"
	}

	exeDir := filepath.Dir(exe)
	path := filepath.Join(exeDir, "..", "plugins", "go-ent", "agents")

	if _, err := os.Stat(path); err == nil {
		return path
	}

	return "plugins/go-ent/agents"
}

func printAgentsTable(agents []agent.AgentMeta) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tMODEL\tDESCRIPTION")
	fmt.Fprintln(w, "----\t-----\t-----------")

	for _, a := range agents {
		desc := truncate(a.Description, 60)
		fmt.Fprintf(w, "%s\t%s\t%s\n", a.Name, a.Model, desc)
	}

	return nil
}

func printAgentsDetailed(agents []agent.AgentMeta) error {
	for i, a := range agents {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("# %s\n\n", a.Name)
		fmt.Printf("**Model**: %s\n", a.Model)
		if a.Color != "" {
			fmt.Printf("**Color**: %s\n", a.Color)
		}
		fmt.Printf("\n%s\n", a.Description)

		if len(a.Skills) > 0 {
			fmt.Printf("\n**Skills**: %s\n", strings.Join(a.Skills, ", "))
		}

		if len(a.Tools) > 0 {
			fmt.Printf("**Tools**: %d enabled\n", len(a.Tools))
		}
	}

	return nil
}

func printAgentInfo(meta agent.AgentMeta) error {
	fmt.Printf("# Agent: %s\n\n", meta.Name)
	fmt.Printf("**Model**: %s\n", meta.Model)
	if meta.Color != "" {
		fmt.Printf("**Color**: %s\n", meta.Color)
	}
	fmt.Printf("\n## Description\n\n%s\n", meta.Description)

	if len(meta.Skills) > 0 {
		fmt.Printf("\n## Skills\n\n")
		for _, skill := range meta.Skills {
			fmt.Printf("- %s\n", skill)
		}
	}

	if len(meta.Tools) > 0 {
		fmt.Printf("\n## Tools\n\n")
		for tool, enabled := range meta.Tools {
			status := "disabled"
			if enabled {
				status = "enabled"
			}
			fmt.Printf("- %s: %s\n", tool, status)
		}
	}

	if meta.Content != "" {
		fmt.Printf("\n## Instructions\n\n%s\n", meta.Content)
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
