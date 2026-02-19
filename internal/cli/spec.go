package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/openspec"
)

func newSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec",
		Short: "Manage OpenSpec specifications (wrapper for openspec CLI)",
		Long:  "Thin wrapper around openspec CLI commands. For full functionality, use openspec directly.",
	}

	cmd.AddCommand(newSpecInitCmd())
	cmd.AddCommand(newSpecListCmd())
	cmd.AddCommand(newSpecShowCmd())

	return cmd
}

func newSpecInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize spec directory",
		Long:  "Create the .spec directory in the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			specDir := filepath.Join(cwd, ".spec")
			if err := os.MkdirAll(specDir, 0o750); err != nil {
				return fmt.Errorf("create spec directory: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Initialized spec directory at %s\n", specDir)
			return nil
		},
	}
}

func newSpecShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show a spec or change by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			client := openspec.New(cwd)
			data, err := client.Show(context.Background(), "", args[0])
			if err != nil {
				return fmt.Errorf("show %s: %w", args[0], err)
			}

			fmt.Println(string(data))
			return nil
		},
	}
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
