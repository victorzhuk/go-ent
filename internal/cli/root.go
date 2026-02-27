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

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(agent.NewCmd())
	cmd.AddCommand(skill.NewCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newWorkspaceCmd())

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

// Execute runs the root command.
func Execute() error {
	return NewRootCmd().Execute()
}
