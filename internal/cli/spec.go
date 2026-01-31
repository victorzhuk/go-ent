package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/openspec"
)

func newSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec",
		Short: "Manage OpenSpec specifications (wrapper for openspec CLI)",
		Long:  "Thin wrapper around openspec CLI commands. For full functionality, use openspec directly.",
	}

	cmd.AddCommand(newSpecListCmd())

	return cmd
}

func newSpecListCmd() *cobra.Command {
	var what string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List OpenSpec changes or specs",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			client := openspec.New(cwd)
			data, err := client.List(context.Background(), what)
			if err != nil {
				return fmt.Errorf("list %s: %w", what, err)
			}

			// Just print the raw output for simplicity
			fmt.Println(string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&what, "type", "changes", "What to list: changes or specs")

	return cmd
}
