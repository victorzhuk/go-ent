package agent

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage agents",
		Long:  `List, inspect, and validate agent definitions.`,
	}

	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newInfoCmd())

	return cmd
}
