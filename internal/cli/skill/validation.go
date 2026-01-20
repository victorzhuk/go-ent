package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	skillpkg "github.com/victorzhuk/go-ent/internal/skill"
)

// Validation utilities for generated skill files.

// ValidateGeneratedSkill validates a generated skill file using the skill parser and validator.
// Returns an error if validation fails, including line numbers for each issue.
func ValidateGeneratedSkill(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("skill file not found: %s", absPath)
	}

	parser := skillpkg.NewParser()
	validator := skillpkg.NewValidator()

	meta, err := parser.ParseSkillFile(absPath)
	if err != nil {
		return fmt.Errorf("parse skill file: %w", err)
	}

	content, err := os.ReadFile(absPath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("read skill file: %w", err)
	}

	result := validator.Validate(meta, string(content))

	if result.ErrorCount() > 0 || len(result.Issues) > 0 {
		var errors, warnings, info []string

		for _, issue := range result.Issues {
			loc := issue.Rule
			if issue.Line > 0 {
				loc = fmt.Sprintf("%s:%d", loc, issue.Line)
			}

			switch issue.Severity {
			case skillpkg.SeverityError:
				errors = append(errors, fmt.Sprintf("  [%s] %s: %s", issue.Severity, loc, issue.Message))
			case skillpkg.SeverityWarning:
				warnings = append(warnings, fmt.Sprintf("  ⚠️  [%s] %s: %s", issue.Severity, loc, issue.Message))
			case skillpkg.SeverityInfo:
				info = append(info, fmt.Sprintf("  ℹ️  [%s] %s: %s", issue.Severity, loc, issue.Message))
			}
		}

		if len(errors) == 0 && len(warnings) == 0 && len(info) == 0 {
			return nil
		}

		var parts []string
		if len(errors) > 0 {
			parts = append(parts, fmt.Sprintf("validation failed for skill '%s':\n  ERRORS:\n%s", meta.Name, strings.Join(errors, "\n")))
		}
		if len(warnings) > 0 {
			parts = append(parts, fmt.Sprintf("\n  WARNINGS:\n%s", strings.Join(warnings, "\n")))
		}
		if len(info) > 0 {
			parts = append(parts, fmt.Sprintf("\n  INFO:\n%s", strings.Join(info, "\n")))
		}

		if len(errors) > 0 {
			return fmt.Errorf("%s", strings.Join(parts, ""))
		}

		fmt.Printf("%s\n", strings.Join(parts, ""))
	}

	return nil
}
