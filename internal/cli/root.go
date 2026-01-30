package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/cli/agent"
	"github.com/victorzhuk/go-ent/internal/cli/skill"
	"github.com/victorzhuk/go-ent/internal/version"
)

var verbose bool

// NewRootCmd creates the root command for the go-ent CLI.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ent",
		Short: "Enterprise Go development toolkit",
		Long: `ent is an enterprise Go development toolkit with multi-agent workflows,
spec-driven development, and intelligent task execution.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	cmd.AddCommand(newVersionCmd())
	// TODO: Phase 5 - Re-enable after ACP integration
	// cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(agent.NewCmd()) // NEW: agent subcommand
	cmd.AddCommand(skill.NewCmd()) // UPDATED: skill subcommand with generate
	cmd.AddCommand(newSpecCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newModelCmd())

	// Backward compatibility aliases
	// These allow users to use old commands while we migrate
	cmd.AddCommand(newGenerateAlias())

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			v := version.Get()
			fmt.Printf("ent %s\n", version.String())
			fmt.Printf("  go: %s\n", v.GoVersion)
			if v.VCSRef != "unknown" && v.VCSRef != "" {
				fmt.Printf("  ref: %s\n", v.VCSRef)
			}
		},
	}
}

// TODO: Phase 5 - Implement using ACP client
// func newRunCmd() *cobra.Command { ... }

// newGenerateAlias creates a backward-compatible alias for the old "ent generate" command
// It delegates to "ent agent generate" for a seamless migration
func newGenerateAlias() *cobra.Command {
	return &cobra.Command{
		Use:    "generate",
		Short:  "Generate agents (alias for 'agent generate')",
		Hidden: true, // Hide from main help
		RunE: func(cmd *cobra.Command, args []string) error {
			// Delegate to agent generate
			agentCmd := agent.NewCmd()
			for _, subcmd := range agentCmd.Commands() {
				if subcmd.Name() == "generate" {
					return subcmd.RunE(subcmd, args)
				}
			}
			return fmt.Errorf("agent generate command not found")
		},
	}
}

// Execute runs the root command.
func Execute() error {
	return NewRootCmd().Execute()
}
