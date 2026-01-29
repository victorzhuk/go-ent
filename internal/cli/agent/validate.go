package agent

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/genconfig"
	"github.com/victorzhuk/go-ent/internal/genspec"
)

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate generated agent files",
		Long: `Validate generated agent files against tool specifications.

This command checks that generated agent files in .claude/agents/ and
.opencode/agents/ conform to the tool's expected format.

Examples:
  ent agent validate
`,
		RunE: runValidate,
	}

	return cmd
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
