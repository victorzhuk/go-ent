package skill

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/skill"
)

func newLintCmd() *cobra.Command {
	var fix bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "lint [path]",
		Short: "Validate and lint skill files",
		Long: `Validate skill files for format and structure issues.

The lint command checks skill files for common issues such as:
  - Missing or malformed frontmatter
  - Invalid XML section formatting
  - Required field validation
  - Formatting inconsistencies

Use --fix to automatically fix common formatting issues.
Use --json for structured output suitable for CI/CD pipelines.

Exit codes:
  0: all skills pass
  1: validation errors found

Examples:
  # Lint current directory
  ent skill lint

  # Lint specific path
  ent skill lint ./skills

  # Auto-fix issues
  ent skill lint --fix

  # JSON output for CI
  ent skill lint --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("resolve path: %w", err)
			}

			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				return fmt.Errorf("path not found: %s", absPath)
			}

			return runLint(absPath, fix, jsonOutput)
		},
		SilenceErrors: true,
	}

	cmd.Flags().BoolVar(&fix, "fix", false, "automatically fix common issues")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output results in JSON format")

	return cmd
}

func runLint(path string, fix, jsonOutput bool) error {
	registry := skill.NewRegistry()

	if err := registry.Load(path); err != nil {
		return fmt.Errorf("load skills from %s: %w", path, err)
	}

	skills := registry.All()
	if len(skills) == 0 {
		fmt.Println("No skills found")
		return nil
	}

	validator := skill.NewValidator()
	results := make([]LintResult, 0, len(skills))

	for _, s := range skills {
		content, err := os.ReadFile(s.FilePath)
		if err != nil {
			return fmt.Errorf("read skill file %s: %w", s.FilePath, err)
		}

		result := validator.Validate(&s, string(content))
		lintResult := LintResult{
			Skill:   s.Name,
			File:    s.FilePath,
			Valid:   result.ErrorCount() == 0,
			Issues:  result.Issues,
			Fixed:   false,
			Changes: []skill.FixChange{},
		}

		if fix {
			fixer := skill.NewFixer()

			hasFixableFrontmatter := fixer.HasFixableFrontmatter(string(content))
			hasFixableXML := fixer.HasFixableXML(string(content))
			hasFixableValidation := fixer.HasFixableValidationIssues(string(content))

			if hasFixableFrontmatter || hasFixableXML {
				fixResult, err := fixer.FixFile(s.FilePath, string(content))
				if err != nil {
					return fmt.Errorf("fix skill file %s: %w", s.FilePath, err)
				}
				lintResult.Fixed = fixResult.Fixed
				lintResult.Changes = fixResult.Changes

				if fixResult.Fixed {
					content, err = os.ReadFile(s.FilePath)
					if err != nil {
						return fmt.Errorf("re-read skill file %s: %w", s.FilePath, err)
					}
					result = validator.Validate(&s, string(content))
					lintResult.Valid = result.ErrorCount() == 0
					lintResult.Issues = result.Issues
				}
			}

			if hasFixableValidation {
				fixResult, err := fixer.FixValidationFile(s.FilePath, string(content))
				if err != nil {
					return fmt.Errorf("fix validation issues in %s: %w", s.FilePath, err)
				}
				if fixResult.Fixed {
					lintResult.Fixed = true
					lintResult.Changes = append(lintResult.Changes, fixResult.Changes...)

					content, err = os.ReadFile(s.FilePath)
					if err != nil {
						return fmt.Errorf("re-read skill file %s: %w", s.FilePath, err)
					}
					result = validator.Validate(&s, string(content))
					lintResult.Valid = result.ErrorCount() == 0
					lintResult.Issues = result.Issues
				}
			}
		}

		results = append(results, lintResult)
	}

	if jsonOutput {
		if err := printJSONResults(results); err != nil {
			return fmt.Errorf("output json: %w", err)
		}
		return determineExitCode(results)
	}

	if err := printConsoleResults(results); err != nil {
		return err
	}
	return determineExitCode(results)
}

type LintResult struct {
	Skill   string
	File    string
	Valid   bool
	Issues  []skill.ValidationIssue
	Fixed   bool
	Changes []skill.FixChange
}

func printConsoleResults(results []LintResult) error {
	hasErrors := false

	for _, r := range results {
		if r.Valid {
			fmt.Printf("✓ %s: OK\n", r.Skill)
			continue
		}

		hasErrors = true
		fmt.Printf("✗ %s: %d issues\n", r.Skill, len(r.Issues))

		for _, issue := range r.Issues {
			severity := string(issue.Severity)
			loc := issue.Rule
			if issue.Line > 0 {
				loc = fmt.Sprintf("%s:%d", loc, issue.Line)
			}

			fmt.Printf("  [%s] %s: %s\n", severity, loc, issue.Message)

			if issue.Suggestion != "" {
				fmt.Printf("    Suggestion: %s\n", issue.Suggestion)
			}
		}

		if r.Fixed {
			fmt.Printf("  Applied fixes:\n")
			for _, change := range r.Changes {
				fmt.Printf("    ✓ %s\n", change.Message)
			}
		}

		fmt.Println()
	}

	if hasErrors {
		return fmt.Errorf("validation failed")
	}

	return nil
}

func determineExitCode(results []LintResult) error {
	for _, r := range results {
		if !r.Valid {
			return fmt.Errorf("validation errors found")
		}
	}
	return nil
}

func printJSONResults(results []LintResult) error {
	var (
		totalErrors   int
		totalWarnings int
		totalInfo     int
	)

	for _, r := range results {
		for _, issue := range r.Issues {
			switch issue.Severity {
			case skill.SeverityError:
				totalErrors++
			case skill.SeverityWarning:
				totalWarnings++
			case skill.SeverityInfo:
				totalInfo++
			}
		}
	}

	totalFixed := 0
	for _, r := range results {
		if r.Fixed {
			totalFixed++
		}
	}

	output := struct {
		Total    int          `json:"total"`
		Errors   int          `json:"errors"`
		Warnings int          `json:"warnings"`
		Success  int          `json:"success"`
		Fixed    int          `json:"fixed"`
		Results  []LintResult `json:"results"`
	}{
		Total:    len(results),
		Errors:   totalErrors,
		Warnings: totalWarnings,
		Success:  len(results) - countInvalid(results),
		Fixed:    totalFixed,
		Results:  results,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func countInvalid(results []LintResult) int {
	count := 0
	for _, r := range results {
		if !r.Valid {
			count++
		}
	}
	return count
}
