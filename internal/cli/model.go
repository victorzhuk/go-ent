package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/model"
)

func newModelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Manage model category mappings",
		Long:  `Manage model category mappings for fast, main, and heavy categories across different runtimes.`,
	}

	cmd.AddCommand(newModelListCmd())
	cmd.AddCommand(newModelSetCmd())
	cmd.AddCommand(newModelResetCmd())

	return cmd
}

func newModelListCmd() *cobra.Command {
	var runtime string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List current model mappings",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := model.LoadGlobal()
			project, _ := model.LoadProject(".")
			cfg := model.Merge(global, project)

			if cfg == nil {
				cfg = model.DefaultConfig()
			}

			if runtime != "" {
				mapping, ok := cfg.Runtimes[runtime]
				if !ok {
					return fmt.Errorf("unknown runtime: %s", runtime)
				}
				fmt.Printf("Runtime: %s\n", runtime)
				fmt.Printf("  fast  -> %s\n", mapping.Fast)
				fmt.Printf("  main  -> %s\n", mapping.Main)
				fmt.Printf("  heavy -> %s\n", mapping.Heavy)
			} else {
				for rt, mapping := range cfg.Runtimes {
					fmt.Printf("Runtime: %s\n", rt)
					fmt.Printf("  fast  -> %s\n", mapping.Fast)
					fmt.Printf("  main  -> %s\n", mapping.Main)
					fmt.Printf("  heavy -> %s\n", mapping.Heavy)
					fmt.Println()
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&runtime, "runtime", "", "Filter by runtime (claude or opencode)")

	return cmd
}

func newModelSetCmd() *cobra.Command {
	var runtime string
	var global bool

	cmd := &cobra.Command{
		Use:   "set <category> <model-id>",
		Short: "Set model for a category",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			category := args[0]
			modelID := args[1]

			if !model.IsValid(category) {
				return fmt.Errorf("invalid category: %s (must be fast, main, or heavy)", category)
			}

			if runtime == "" {
				return fmt.Errorf("--runtime flag is required")
			}

			var cfg *model.Config
			var err error

			if global {
				cfg, err = model.LoadGlobal()
			} else {
				cfg, err = model.LoadProject(".")
			}

			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if cfg == nil {
				cfg = model.DefaultConfig()
			}

			mapping, ok := cfg.Runtimes[runtime]
			if !ok {
				mapping = model.Mapping{}
			}

			switch model.Category(category) {
			case model.Fast:
				mapping.Fast = modelID
			case model.Main:
				mapping.Main = modelID
			case model.Heavy:
				mapping.Heavy = modelID
			}

			cfg.Runtimes[runtime] = mapping

			if global {
				err = model.SaveGlobal(cfg)
			} else {
				err = model.SaveProject(".", cfg)
			}

			if err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Printf("Set %s/%s = %s\n", runtime, category, modelID)
			return nil
		},
	}

	cmd.Flags().StringVar(&runtime, "runtime", "", "Runtime to configure (claude or opencode)")
	cmd.Flags().BoolVar(&global, "global", false, "Save to global config (~/.go-ent/models.yaml)")

	return cmd
}

func newModelResetCmd() *cobra.Command {
	var runtime string
	var global bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset model mappings to defaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := model.DefaultConfig()

			if runtime != "" {
				// Only reset specific runtime
				existing, _ := model.LoadGlobal()
				if existing != nil && !global {
					existing, _ = model.LoadProject(".")
				}

				if existing != nil {
					for rt := range existing.Runtimes {
						if rt != runtime {
							cfg.Runtimes[rt] = existing.Runtimes[rt]
						}
					}
				}
			}

			var err error
			if global {
				err = model.SaveGlobal(cfg)
			} else {
				err = model.SaveProject(".", cfg)
			}

			if err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			if runtime != "" {
				fmt.Printf("Reset %s runtime to defaults\n", runtime)
			} else {
				fmt.Println("Reset all runtimes to defaults")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&runtime, "runtime", "", "Runtime to reset (claude or opencode)")
	cmd.Flags().BoolVar(&global, "global", false, "Reset global config (~/.go-ent/models.yaml)")

	return cmd
}
