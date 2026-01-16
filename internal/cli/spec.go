package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/spec"
)

func newSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec",
		Short: "Manage OpenSpec specifications",
		Long:  "Initialize, list, and inspect OpenSpec specifications, changes, and tasks",
	}

	cmd.AddCommand(newSpecInitCmd())
	cmd.AddCommand(newSpecListCmd())
	cmd.AddCommand(newSpecShowCmd())

	return cmd
}

func newSpecInitCmd() *cobra.Command {
	var (
		name        string
		module      string
		description string
	)

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize openspec folder in a project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			store := spec.NewStore(path)

			exists, err := store.Exists()
			if err != nil {
				return fmt.Errorf("check openspec folder: %w", err)
			}

			if exists {
				fmt.Printf("openspec folder already exists at %s\n", store.SpecPath())
				return nil
			}

			project := spec.Project{
				Name:        name,
				Module:      module,
				Description: description,
			}

			if err := store.Init(project); err != nil {
				return fmt.Errorf("initialize: %w", err)
			}

			fmt.Printf("✅ Initialized openspec at %s\n", store.SpecPath())
			fmt.Println("\nNext steps:")
			fmt.Println("  1. Create specs: openspec/specs/{name}/spec.md")
			fmt.Println("  2. Create changes: openspec/changes/{id}/proposal.md")
			fmt.Println("  3. Run: go-ent spec list spec")

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Project name")
	cmd.Flags().StringVar(&module, "module", "", "Go module path")
	cmd.Flags().StringVar(&description, "description", "", "Project description")

	return cmd
}

func newSpecListCmd() *cobra.Command {
	var (
		status string
		format string
	)

	cmd := &cobra.Command{
		Use:   "list <type>",
		Short: "List specs, changes, or tasks",
		Long:  "Type must be one of: spec, change, task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			itemType := args[0]

			if itemType != "spec" && itemType != "change" && itemType != "task" {
				return fmt.Errorf("invalid type: %s. Must be spec, change, or task", itemType)
			}

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get current directory: %w", err)
			}

			store := spec.NewStore(cwd)

			exists, err := store.Exists()
			if err != nil {
				return fmt.Errorf("check openspec folder: %w", err)
			}

			if !exists {
				return fmt.Errorf("no openspec folder found. Run 'go-ent spec init' first")
			}

			var items []spec.ListItem

			switch itemType {
			case "spec":
				items, err = store.ListSpecs()
			case "change":
				items, err = store.ListChanges(status)
			case "task":
				items, err = store.ListTasks()
			}

			if err != nil {
				return fmt.Errorf("list %s: %w", itemType, err)
			}

			if len(items) == 0 {
				fmt.Printf("No %s found\n", itemType)
				return nil
			}

			switch format {
			case "table":
				return printSpecTable(items, itemType)
			case "detailed":
				return printSpecDetailed(items)
			default:
				return fmt.Errorf("unknown format: %s", format)
			}
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status (for changes: active, archived)")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, detailed)")

	return cmd
}

func newSpecShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <type> <id>",
		Short: "Show detailed content of a spec, change, or task",
		Long:  "Type must be one of: spec, change, task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			itemType := args[0]
			id := args[1]

			if itemType != "spec" && itemType != "change" && itemType != "task" {
				return fmt.Errorf("invalid type: %s. Must be spec, change, or task", itemType)
			}

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get current directory: %w", err)
			}

			store := spec.NewStore(cwd)

			exists, err := store.Exists()
			if err != nil {
				return fmt.Errorf("check openspec folder: %w", err)
			}

			if !exists {
				return fmt.Errorf("no openspec folder found. Run 'go-ent spec init' first")
			}

			var path string
			switch itemType {
			case "spec":
				path = fmt.Sprintf("specs/%s/spec.md", id)
			case "change":
				path = fmt.Sprintf("changes/%s/proposal.md", id)
			case "task":
				path = fmt.Sprintf("tasks/%s.md", id)
			}

			content, err := store.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read %s: %w", path, err)
			}

			fmt.Printf("# %s: %s\n\n", itemType, id)
			fmt.Println(content)

			return nil
		},
	}
}

func printSpecTable(items []spec.ListItem, itemType string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_ = w.Flush() // intentionally ignore error in defer

	if itemType == "change" {
		_, _ = fmt.Fprintln(w, "ID\tSTATUS\tNAME")
		_, _ = fmt.Fprintln(w, "--\t------\t----")
		for _, item := range items {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", item.ID, item.Status, item.Name)
		}
	} else {
		_, _ = fmt.Fprintln(w, "ID\tNAME")
		_, _ = fmt.Fprintln(w, "--\t----")
		for _, item := range items {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", item.ID, item.Name)
		}
	}

	return nil
}

func printSpecDetailed(items []spec.ListItem) error {
	for i, item := range items {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("# %s\n\n", item.ID)
		if item.Name != "" {
			fmt.Printf("**Name**: %s\n", item.Name)
		}
		if item.Type != "" {
			fmt.Printf("**Type**: %s\n", item.Type)
		}
		if item.Status != "" {
			fmt.Printf("**Status**: %s\n", item.Status)
		}
		if item.Description != "" {
			fmt.Printf("\n%s\n", item.Description)
		}
		if item.Path != "" {
			fmt.Printf("\n**Path**: %s\n", item.Path)
		}
	}

	return nil
}
