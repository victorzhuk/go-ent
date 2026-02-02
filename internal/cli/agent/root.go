package agent

import "github.com/spf13/cobra"

// NewCmd creates the agent command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage agents",
		Long:  `Manage agent definitions and generation for different tools.`,
	}

	cmd.AddCommand(newAgentCmd())
	cmd.AddCommand(newGenerateCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newListCmd())

	return cmd
}
