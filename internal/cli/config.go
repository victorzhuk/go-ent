package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/config"
	"gopkg.in/yaml.v3"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Initialize, view, and modify ent runtime configuration",
	}

	cmd.AddCommand(newConfigInitCmd())
	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigSetCmd())

	return cmd
}

func newConfigInitCmd() *cobra.Command {
	var runtime string

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize ent runtime config",
		Long:  "Create .claude/ent.yaml (or .opencode/ent.yaml) with default model configuration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot := "."
			if len(args) > 0 {
				projectRoot = args[0]
			}

			rt := runtime
			if rt == "" {
				rt = config.DetectRuntime(projectRoot)
			}
			if rt == "" {
				rt = "claude"
			}

			cfgPath, err := config.RuntimeConfigPath(rt, projectRoot)
			if err != nil {
				return err
			}

			if _, err := os.Stat(cfgPath); err == nil {
				return fmt.Errorf("config file already exists at %s", cfgPath)
			}

			if err := os.MkdirAll(filepath.Dir(cfgPath), 0o750); err != nil {
				return fmt.Errorf("create config directory: %w", err)
			}

			cfg := &config.RuntimeConfig{
				Models: config.DefaultModelsForRuntime(rt),
			}
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("marshal config: %w", err)
			}

			if err := os.WriteFile(cfgPath, data, 0o600); err != nil {
				return fmt.Errorf("write config file: %w", err)
			}

			fmt.Printf("Created config file at %s\n", cfgPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&runtime, "runtime", "", "target runtime (claude|opencode)")
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	var runtime string

	cmd := &cobra.Command{
		Use:   "show [path]",
		Short: "Show runtime configuration",
		Long:  "Display model tiers from .claude/ent.yaml or .opencode/ent.yaml",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot := "."
			if len(args) > 0 {
				projectRoot = args[0]
			}

			rt := runtime
			if rt == "" {
				rt = config.DetectRuntime(projectRoot)
			}
			if rt == "" {
				rt = "claude"
			}

			cfg, err := config.LoadRuntimeConfig(projectRoot, rt)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("marshal config: %w", err)
			}

			fmt.Print(string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&runtime, "runtime", "", "target runtime (claude|opencode, auto-detected if omitted)")
	return cmd
}

func newConfigSetCmd() *cobra.Command {
	var runtime string

	return &cobra.Command{
		Use:   "set <key> <value> [path]",
		Short: "Set a model tier or per-agent override",
		Long: `Set a model tier or per-agent override in the runtime config.

Examples:
  ent config set models.main claude-sonnet-4-6-20260101
  ent config set models.agents.coder heavy
  ent config set models.fast anthropic/claude-haiku-4-5

Supported keys:
  - models.fast, models.main, models.heavy
  - models.agents.<name>`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			projectRoot := "."
			if len(args) > 2 {
				projectRoot = args[2]
			}

			rt := runtime
			if rt == "" {
				rt = config.DetectRuntime(projectRoot)
			}
			if rt == "" {
				rt = "claude"
			}

			cfg, err := config.LoadRuntimeConfig(projectRoot, rt)
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("load config: %w", err)
				}
				cfg = &config.RuntimeConfig{}
			}

			if err := config.ApplyKey(cfg, key, value); err != nil {
				return err
			}

			cfgPath, err := config.RuntimeConfigPath(rt, projectRoot)
			if err != nil {
				return err
			}

			if err := os.MkdirAll(filepath.Dir(cfgPath), 0o750); err != nil {
				return fmt.Errorf("create config directory: %w", err)
			}

			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("marshal config: %w", err)
			}

			if err := os.WriteFile(cfgPath, data, 0o600); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			fmt.Printf("Updated %s = %s\n", key, value)
			return nil
		},
	}
}
