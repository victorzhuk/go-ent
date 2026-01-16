package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/domain"
	"gopkg.in/yaml.v3"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Initialize, view, and modify go-ent configuration",
	}

	cmd.AddCommand(newConfigInitCmd())
	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigSetCmd())

	return cmd
}

func newConfigInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize configuration file",
		Long:  "Create .go-ent/config.yaml with default configuration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot := "."
			if len(args) > 0 {
				projectRoot = args[0]
			}

			cfgPath := filepath.Join(projectRoot, ".go-ent", "config.yaml")

			// Check if config already exists
			if _, err := os.Stat(cfgPath); err == nil {
				return fmt.Errorf("config file already exists at %s", cfgPath)
			}

			// Create .go-ent directory
			cfgDir := filepath.Join(projectRoot, ".go-ent")
			if err := os.MkdirAll(cfgDir, 0750); err != nil {
				return fmt.Errorf("create config directory: %w", err)
			}

			// Generate default config
			cfg := config.DefaultConfig()

			// Marshal to YAML
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("marshal config: %w", err)
			}

			// Write config file
			if err := os.WriteFile(cfgPath, data, 0600); err != nil {
				return fmt.Errorf("write config file: %w", err)
			}

			fmt.Printf("✅ Created config file at %s\n", cfgPath)
			fmt.Println("\nDefault configuration:")
			fmt.Printf("  - Budget: $%.2f/day, $%.2f/month, $%.2f/task\n",
				cfg.Budget.Daily, cfg.Budget.Monthly, cfg.Budget.PerTask)
			fmt.Printf("  - Runtime: %s (fallback: %v)\n",
				cfg.Runtime.Preferred, cfg.Runtime.Fallback)
			fmt.Printf("  - Default agent: %s\n", cfg.Agents.Default)
			fmt.Printf("  - Models: opus, sonnet, haiku\n")
			fmt.Printf("  - Skills: %d enabled\n", len(cfg.Skills.Enabled))

			return nil
		},
	}
}

func newConfigShowCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "show [path]",
		Short: "Show current configuration",
		Long:  "Display the current configuration (from file or defaults)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot := "."
			if len(args) > 0 {
				projectRoot = args[0]
			}

			cfg, err := config.Load(projectRoot)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			switch format {
			case "yaml":
				return printConfigYAML(cfg)
			case "summary":
				return printConfigSummary(cfg)
			default:
				return fmt.Errorf("unknown format: %s", format)
			}
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "yaml", "Output format (yaml, summary)")

	return cmd
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value> [path]",
		Short: "Set a configuration value",
		Long: `Set a configuration value using dot notation.

Examples:
  go-ent config set budget.daily 15.0
  go-ent config set agents.default developer
  go-ent config set runtime.preferred cli

Supported keys:
  - budget.daily, budget.monthly, budget.per_task
  - agents.default
  - runtime.preferred
  - models.<name> (e.g., models.opus)`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			projectRoot := "."
			if len(args) > 2 {
				projectRoot = args[2]
			}

			cfg, err := config.Load(projectRoot)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			// Update the specified key
			if err := updateConfigValue(cfg, key, value); err != nil {
				return err
			}

			// Validate updated config
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("validate config: %w", err)
			}

			// Save updated config
			cfgPath := filepath.Join(projectRoot, ".go-ent", "config.yaml")
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("marshal config: %w", err)
			}

			if err := os.WriteFile(cfgPath, data, 0600); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			fmt.Printf("✅ Updated %s = %s\n", key, value)

			return nil
		},
	}
}

func printConfigYAML(cfg *config.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func printConfigSummary(cfg *config.Config) error {
	fmt.Printf("# go-ent Configuration\n\n")
	fmt.Printf("**Version**: %s\n\n", cfg.Version)

	fmt.Printf("## Budget\n\n")
	fmt.Printf("- Daily: $%.2f\n", cfg.Budget.Daily)
	fmt.Printf("- Monthly: $%.2f\n", cfg.Budget.Monthly)
	fmt.Printf("- Per Task: $%.2f\n", cfg.Budget.PerTask)
	fmt.Printf("- Tracking: %v\n\n", cfg.Budget.Tracking)

	fmt.Printf("## Runtime\n\n")
	fmt.Printf("- Preferred: %s\n", cfg.Runtime.Preferred)
	if len(cfg.Runtime.Fallback) > 0 {
		fmt.Printf("- Fallback: %v\n", cfg.Runtime.Fallback)
	}
	fmt.Println()

	fmt.Printf("## Agents\n\n")
	fmt.Printf("- Default: %s\n", cfg.Agents.Default)
	fmt.Printf("- Roles configured: %d\n", len(cfg.Agents.Roles))
	fmt.Printf("- Auto-delegation: %v\n\n", cfg.Agents.Delegation.Auto)

	fmt.Printf("## Models\n\n")
	for name, id := range cfg.Models {
		fmt.Printf("- %s: %s\n", name, id)
	}
	fmt.Println()

	fmt.Printf("## Skills\n\n")
	fmt.Printf("- Enabled: %d\n", len(cfg.Skills.Enabled))
	if cfg.Skills.CustomDir != "" {
		fmt.Printf("- Custom Directory: %s\n", cfg.Skills.CustomDir)
	}

	return nil
}

func updateConfigValue(cfg *config.Config, key, value string) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid key format: %s (use dot notation, e.g., budget.daily)", key)
	}

	section := parts[0]
	field := parts[1]

	switch section {
	case "budget":
		return updateBudgetValue(cfg, field, value)
	case "agents":
		return updateAgentsValue(cfg, field, value)
	case "runtime":
		return updateRuntimeValue(cfg, field, value)
	case "models":
		if len(parts) == 2 {
			cfg.Models[field] = value
			return nil
		}
		return fmt.Errorf("invalid models key: %s", key)
	default:
		return fmt.Errorf("unknown config section: %s", section)
	}
}

func updateBudgetValue(cfg *config.Config, field, value string) error {
	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid float value for budget.%s: %w", field, err)
	}

	switch field {
	case "daily":
		cfg.Budget.Daily = floatVal
	case "monthly":
		cfg.Budget.Monthly = floatVal
	case "per_task":
		cfg.Budget.PerTask = floatVal
	case "tracking":
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool value for budget.tracking: %w", err)
		}
		cfg.Budget.Tracking = boolVal
	default:
		return fmt.Errorf("unknown budget field: %s", field)
	}

	return nil
}

func updateAgentsValue(cfg *config.Config, field, value string) error {
	switch field {
	case "default":
		role := domain.AgentRole(value)
		if !role.Valid() {
			return fmt.Errorf("invalid agent role: %s", value)
		}
		cfg.Agents.Default = role
	default:
		return fmt.Errorf("unknown agents field: %s (only 'default' is supported via set)", field)
	}

	return nil
}

func updateRuntimeValue(cfg *config.Config, field, value string) error {
	switch field {
	case "preferred":
		runtime := domain.Runtime(value)
		if !runtime.Valid() {
			return fmt.Errorf("invalid runtime: %s", value)
		}
		cfg.Runtime.Preferred = runtime
	default:
		return fmt.Errorf("unknown runtime field: %s", field)
	}

	return nil
}
