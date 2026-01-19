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

	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("read skill file: %w", err)
	}

	result := validator.Validate(meta, string(content))

	if !result.Valid || result.ErrorCount() > 0 {
		var issues []string
		for _, issue := range result.Issues {
			if issue.Severity != skillpkg.SeverityError {
				continue
			}
			loc := issue.Rule
			if issue.Line > 0 {
				loc = fmt.Sprintf("%s:%d", loc, issue.Line)
			}
			issues = append(issues, fmt.Sprintf("  [%s] %s: %s", issue.Severity, loc, issue.Message))
		}

		if len(issues) == 0 {
			return nil
		}

		return fmt.Errorf("validation failed for skill '%s':\n%s", meta.Name, strings.Join(issues, "\n"))
	}

	return nil
}
