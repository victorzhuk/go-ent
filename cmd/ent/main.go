package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/genconfig"
	"github.com/victorzhuk/go-ent/internal/generator"
	"github.com/victorzhuk/go-ent/internal/genspec"
)

var (
	toolsFlag []string
	forceFlag bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ent",
		Short: "Go-Ent generator tool for multi-tool agent output",
	}

	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(generateCmd())
	rootCmd.AddCommand(validateCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize project with selected tools",
		RunE:  runInit,
	}

	cmd.Flags().StringSliceVar(&toolsFlag, "tools", []string{"claude"}, "Tools to generate for (claude,opencode,openspec)")
	cmd.Flags().BoolVar(&forceFlag, "force", false, "Overwrite existing config")

	return cmd
}

func generateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate agent output for configured tools",
		RunE:  runGenerate,
	}

	cmd.Flags().StringSliceVar(&toolsFlag, "tools", nil, "Override tools from config")

	return cmd
}

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate generated files against tool specs",
		RunE:  runValidate,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := "ent.yaml"

	// Check if config exists
	if !forceFlag {
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("ent.yaml already exists, use --force to overwrite")
		}
	}

	// Create default config
	cfg := genconfig.Default()
	cfg.Tools = toolsFlag

	// Create tool directories
	for _, tool := range toolsFlag {
		var dir string
		switch tool {
		case "claude":
			dir = ".claude/agents"
		case "opencode":
			dir = ".opencode/agents"
		case "openspec":
			dir = "openspec"
		default:
			return fmt.Errorf("unknown tool: %s", tool)
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create %s directory: %w", dir, err)
		}
		fmt.Printf("Created %s/\n", dir)
	}

	// Save config
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("Created %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Customize src/agents/*.yaml for your project")
	fmt.Println("  2. Run 'ent generate' to build agent files")
	fmt.Println("  3. Restart your code tool to load agents")

	return nil
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := genconfig.Load("ent.yaml")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Override tools if specified
	tools := cfg.Tools
	if len(toolsFlag) > 0 {
		tools = toolsFlag
	}

	// Build targets
	var targets []generator.Target
	for _, tool := range tools {
		switch tool {
		case "claude":
			targets = append(targets, generator.NewClaudeTarget(".claude/agents"))
		case "opencode":
			targets = append(targets, generator.NewOpenCodeTarget(".opencode/agents"))
		case "openspec":
			// OpenSpec is a workflow tool, not an agent generation target - skip
			continue
		default:
			return fmt.Errorf("unknown tool: %s", tool)
		}
	}

	// Run generator
	gen := generator.New("src", targets...)
	if err := gen.GenerateAll(); err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	fmt.Println("\nGeneration complete!")
	return nil
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := genconfig.Load("ent.yaml")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Validate each tool's output
	hasErrors := false
	for _, tool := range cfg.Tools {
		var dir string
		switch tool {
		case "claude":
			dir = ".claude/agents"
		case "opencode":
			dir = ".opencode/agents"
		case "openspec":
			// OpenSpec has no agents to validate - skip
			continue
		default:
			fmt.Printf("Skipping unknown tool: %s\n", tool)
			continue
		}

		if err := validateTool(tool, dir); err != nil {
			return err
		}

		// Check results
		validator, err := genspec.NewValidator(tool)
		if err != nil {
			return fmt.Errorf("create validator for %s: %w", tool, err)
		}

		results, err := validator.ValidateDirectory(dir)
		if err != nil {
			return fmt.Errorf("validate %s: %w", tool, err)
		}

		// Print results
		fmt.Printf("\n%s (%s/)\n", tool, dir)
		fmt.Println(strings.Repeat("=", 50))

		for _, result := range results {
			filename := filepath.Base(result.File)
			if result.IsValid() {
				fmt.Printf("✓ %s\n", filename)
			} else {
				hasErrors = true
				fmt.Printf("✗ %s\n", filename)
				for _, err := range result.Errors {
					if err.Field == "" {
						fmt.Printf("    %s\n", err.Message)
					} else {
						fmt.Printf("    %s: %s\n", err.Field, err.Message)
					}
				}
			}
		}
	}

	if hasErrors {
		fmt.Println("\nValidation failed - see errors above")
		return fmt.Errorf("validation failed")
	}

	fmt.Println("\n✓ All files valid!")
	return nil
}

func validateTool(tool, dir string) error {
	// Check directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("%s directory not found: %s (run 'ent generate' first)", tool, dir)
	}
	return nil
}
