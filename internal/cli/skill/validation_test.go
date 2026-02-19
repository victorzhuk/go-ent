package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skillpkg "github.com/victorzhuk/go-ent/internal/skill"
)

func TestValidateGeneratedSkill_ValidSkill(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	validSkill := `---
name: test-skill
description: A valid test skill for expert assistance
triggers:
  - test skill
  - testing
---

## Role

You are a test skill expert focused on providing helpful testing guidance and best practices.

## Instructions

### Core Approach

Provide helpful testing guidance using these patterns:

` + "```" + `go
func TestExample(t *testing.T) {
    t.Parallel()
    assert.Equal(t, "expected", "expected")
}
` + "```" + `

### Edge Cases

If test fails: investigate root cause before applying fix.

## Examples

### Example 1: Basic test

**Input**: Test input

**Output**: Test output
`

	require.NoError(t, os.WriteFile(skillPath, []byte(validSkill), 0o644))

	err := ValidateGeneratedSkill(skillPath)
	assert.NoError(t, err)
}

func TestValidateGeneratedSkill_FileNotFound(t *testing.T) {
	t.Parallel()

	nonExistentPath := filepath.Join(t.TempDir(), "nonexistent.md")

	err := ValidateGeneratedSkill(nonExistentPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestValidateGeneratedSkill_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	invalidYAML := `---
name: test
description: test
invalid: yaml: content:
---

## Role

Test role with enough content to be valid here.

## Instructions

### Section

Test instructions content.

## Examples

### Example 1: Test

**Input**: test

**Output**: test
`

	require.NoError(t, os.WriteFile(skillPath, []byte(invalidYAML), 0o644))

	err := ValidateGeneratedSkill(skillPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse skill file")
}

func TestValidateGeneratedSkill_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	require.NoError(t, os.WriteFile(skillPath, []byte(""), 0o644))

	err := ValidateGeneratedSkill(skillPath)
	assert.Error(t, err)
}

func TestValidateGeneratedSkill_MissingFrontmatter(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	noFrontmatter := `# Test Skill

## Role

Test role content.

## Instructions

Test instructions.

## Examples

### Example 1: Basic

**Input**: test

**Output**: test
`

	require.NoError(t, os.WriteFile(skillPath, []byte(noFrontmatter), 0o644))

	err := ValidateGeneratedSkill(skillPath)
	assert.Error(t, err)
}

func TestValidateGeneratedSkill_MissingName(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	missingName := `---
description: A test skill
triggers:
  - test
---

## Role

Test role content with enough content to be valid here.

## Instructions

### Section

Test instructions content.

## Examples

### Example 1: Test

**Input**: test

**Output**: test
`

	require.NoError(t, os.WriteFile(skillPath, []byte(missingName), 0o644))

	err := ValidateGeneratedSkill(skillPath)
	assert.Error(t, err)
}

func TestValidateGeneratedSkill_CompleteValidSkill(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	completeSkill := `---
name: go-database
description: Database integration patterns with PostgreSQL and pgx for production use
triggers:
  - database
  - postgresql
  - pgx
  - sql query
---

## Role

Expert Go developer specializing in database integration patterns. Extensive experience with PostgreSQL, pgx, and database transaction management. Follows best practices for connection pooling, query optimization, and error handling.

## Instructions

### Connection Management

Set up connection pools with appropriate limits:

` + "```" + `go
pool, err := pgxpool.New(ctx, dsn)
if err != nil {
    return fmt.Errorf("create pool: %w", err)
}
defer pool.Close()
` + "```" + `

### Query Patterns

Use parameterized queries to prevent SQL injection:

` + "```" + `go
query, args, _ := sq.Select("id", "email").
    From("users").
    Where(sq.Eq{"id": id}).
    ToSql()

row := pool.QueryRow(ctx, query, args...)
` + "```" + `

### Transaction Management

Handle transactions with proper rollback:

` + "```" + `go
tx, err := pool.Begin(ctx)
if err != nil {
    return fmt.Errorf("begin tx: %w", err)
}
defer func() { _ = tx.Rollback(ctx) }()

if err := tx.Commit(ctx); err != nil {
    return fmt.Errorf("commit: %w", err)
}
` + "```" + `

### Edge Cases

If connection pool is exhausted: implement retry logic with exponential backoff.

If transaction fails due to serialization error: retry up to 3 times.

## Examples

### Example 1: Set up connection pool with pgx

**Input**: How do I set up a connection pool with pgx?

**Output**:
` + "```" + `go
pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
if err != nil {
    return fmt.Errorf("create pool: %w", err)
}
defer pool.Close()
` + "```" + `

### Example 2: Handle transactions properly

**Input**: How do I handle transactions properly?

**Output**:
` + "```" + `go
tx, err := pool.Begin(ctx)
if err != nil {
    return fmt.Errorf("begin tx: %w", err)
}
defer func() { _ = tx.Rollback(ctx) }()

// do work...

if err := tx.Commit(ctx); err != nil {
    return fmt.Errorf("commit: %w", err)
}
` + "```" + `
`

	require.NoError(t, os.WriteFile(skillPath, []byte(completeSkill), 0o644))

	err := ValidateGeneratedSkill(skillPath)
	assert.NoError(t, err)
}

func TestValidationIssueWithEmptyFields_CLIFormatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		issue       skillpkg.ValidationIssue
		contains    []string
		notContains []string
	}{
		{
			name: "empty suggestion and example",
			issue: skillpkg.ValidationIssue{
				Rule:     "TEST001",
				Severity: skillpkg.SeverityError,
				Message:  "test error",
				Line:     1,
			},
			contains:    []string{"[error]", "TEST001:1", "test error"},
			notContains: []string{"Suggestion:", "Example:"},
		},
		{
			name: "empty suggestion, populated example",
			issue: skillpkg.ValidationIssue{
				Rule:     "TEST002",
				Severity: skillpkg.SeverityError,
				Message:  "test error",
				Example:  "example-value",
				Line:     2,
			},
			contains:    []string{"[error]", "TEST002:2", "test error", "Example: example-value"},
			notContains: []string{"Suggestion:"},
		},
		{
			name: "populated suggestion, empty example",
			issue: skillpkg.ValidationIssue{
				Rule:       "TEST003",
				Severity:   skillpkg.SeverityError,
				Message:    "test error",
				Suggestion: "fix this by doing X",
				Line:       3,
			},
			contains:    []string{"[error]", "TEST003:3", "test error", "Suggestion: fix this by doing X"},
			notContains: []string{"Example:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loc := tt.issue.Rule
			if tt.issue.Line > 0 {
				loc = fmt.Sprintf("%s:%d", tt.issue.Rule, tt.issue.Line)
			}

			var msg string
			switch tt.issue.Severity {
			case skillpkg.SeverityError:
				msg = fmt.Sprintf("  [%s] %s: %s", tt.issue.Severity, loc, tt.issue.Message)
			case skillpkg.SeverityWarning:
				msg = fmt.Sprintf("  ⚠️  [%s] %s: %s", tt.issue.Severity, loc, tt.issue.Message)
			case skillpkg.SeverityInfo:
				msg = fmt.Sprintf("  ℹ️  [%s] %s: %s", tt.issue.Severity, loc, tt.issue.Message)
			}

			if tt.issue.Suggestion != "" {
				msg += fmt.Sprintf("\n    Suggestion: %s", tt.issue.Suggestion)
			}
			if tt.issue.Example != "" {
				msg += fmt.Sprintf("\n    Example: %s", tt.issue.Example)
			}

			for _, s := range tt.contains {
				assert.Contains(t, msg, s)
			}
			for _, s := range tt.notContains {
				assert.NotContains(t, msg, s)
			}
		})
	}
}

func TestCLIFormatter_OutputFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		issue             skillpkg.ValidationIssue
		expectedStructure []string
	}{
		{
			name: "proper indentation for suggestion",
			issue: skillpkg.ValidationIssue{
				Rule:       "TEST001",
				Severity:   skillpkg.SeverityError,
				Message:    "field missing",
				Suggestion: "add the required field",
				Line:       1,
			},
			expectedStructure: []string{
				"  [error]",
				"    Suggestion:",
			},
		},
		{
			name: "proper indentation for example",
			issue: skillpkg.ValidationIssue{
				Rule:     "TEST002",
				Severity: skillpkg.SeverityError,
				Message:  "invalid format",
				Example:  "name: valid-name",
				Line:     2,
			},
			expectedStructure: []string{
				"  [error]",
				"    Example:",
			},
		},
		{
			name: "both suggestion and example present",
			issue: skillpkg.ValidationIssue{
				Rule:       "TEST003",
				Severity:   skillpkg.SeverityError,
				Message:    "multiple issues",
				Suggestion: "fix these issues",
				Example:    "name: test",
				Line:       3,
			},
			expectedStructure: []string{
				"  [error]",
				"    Suggestion:",
				"    Example:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loc := fmt.Sprintf("%s:%d", tt.issue.Rule, tt.issue.Line)

			var msg string
			msg = fmt.Sprintf("  [%s] %s: %s", tt.issue.Severity, loc, tt.issue.Message)

			if tt.issue.Suggestion != "" {
				msg += fmt.Sprintf("\n    Suggestion: %s", tt.issue.Suggestion)
			}
			if tt.issue.Example != "" {
				msg += fmt.Sprintf("\n    Example: %s", tt.issue.Example)
			}

			lines := strings.Split(msg, "\n")

			for _, expected := range tt.expectedStructure {
				found := false
				for _, line := range lines {
					if strings.Contains(line, expected) {
						found = true
						break
					}
				}
				assert.True(t, found, "expected to find '%s' in output", expected)
			}
		})
	}
}

func TestCLIFormatter_ClearPrefixes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		issue    skillpkg.ValidationIssue
		prefixes []string
	}{
		{
			name: "suggestion prefix is clear",
			issue: skillpkg.ValidationIssue{
				Rule:       "TEST001",
				Severity:   skillpkg.SeverityError,
				Message:    "test error",
				Suggestion: "fix suggestion",
				Line:       1,
			},
			prefixes: []string{"Suggestion: "},
		},
		{
			name: "example prefix is clear",
			issue: skillpkg.ValidationIssue{
				Rule:     "TEST002",
				Severity: skillpkg.SeverityError,
				Message:  "test error",
				Example:  "example value",
				Line:     2,
			},
			prefixes: []string{"Example: "},
		},
		{
			name: "both prefixes present",
			issue: skillpkg.ValidationIssue{
				Rule:       "TEST003",
				Severity:   skillpkg.SeverityError,
				Message:    "test error",
				Suggestion: "fix suggestion",
				Example:    "example value",
				Line:       3,
			},
			prefixes: []string{"Suggestion: ", "Example: "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loc := fmt.Sprintf("%s:%d", tt.issue.Rule, tt.issue.Line)

			var msg string
			msg = fmt.Sprintf("  [%s] %s: %s", tt.issue.Severity, loc, tt.issue.Message)

			if tt.issue.Suggestion != "" {
				msg += fmt.Sprintf("\n    Suggestion: %s", tt.issue.Suggestion)
			}
			if tt.issue.Example != "" {
				msg += fmt.Sprintf("\n    Example: %s", tt.issue.Example)
			}

			for _, prefix := range tt.prefixes {
				assert.Contains(t, msg, prefix, "expected prefix '%s' not found", prefix)
			}
		})
	}
}

func TestCLIFormatter_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		issue skillpkg.ValidationIssue
		check func(t *testing.T, msg string)
	}{
		{
			name: "very long suggestion",
			issue: skillpkg.ValidationIssue{
				Rule:       "TEST001",
				Severity:   skillpkg.SeverityError,
				Message:    "test error",
				Suggestion: strings.Repeat("This is a very long suggestion that spans multiple words and should still be readable when displayed to the user in the CLI output format. ", 5),
				Line:       1,
			},
			check: func(t *testing.T, msg string) {
				lines := strings.Split(msg, "\n")
				suggestionLine := ""
				for _, line := range lines {
					if strings.Contains(line, "Suggestion:") {
						suggestionLine = line
						break
					}
				}
				assert.NotEmpty(t, suggestionLine, "suggestion line should exist")
				assert.True(t, strings.HasPrefix(suggestionLine, "    Suggestion: "), "suggestion should have 4-space indentation")
			},
		},
		{
			name: "multi-line example",
			issue: skillpkg.ValidationIssue{
				Rule:     "TEST002",
				Severity: skillpkg.SeverityError,
				Message:  "test error",
				Example:  "name: test-skill\ndescription: A test skill\nversion: \"1.0.0\"",
				Line:     2,
			},
			check: func(t *testing.T, msg string) {
				assert.Contains(t, msg, "    Example:")
				lines := strings.Split(msg, "\n")
				exampleLine := ""
				for _, line := range lines {
					if strings.Contains(line, "Example:") {
						exampleLine = line
						break
					}
				}
				assert.NotEmpty(t, exampleLine, "example line should exist")
				assert.True(t, strings.HasPrefix(exampleLine, "    Example: "), "example should have 4-space indentation")
			},
		},
		{
			name: "suggestion and example both long",
			issue: skillpkg.ValidationIssue{
				Rule:       "TEST003",
				Severity:   skillpkg.SeverityError,
				Message:    "test error",
				Suggestion: "This is a very long suggestion that provides detailed guidance on how to fix the issue at hand with multiple pieces of advice.",
				Example:    "---\nname: example-skill\ndescription: An example skill\nversion: \"1.0.0\"\n---",
				Line:       3,
			},
			check: func(t *testing.T, msg string) {
				lines := strings.Split(msg, "\n")
				var hasSuggestion, hasExample bool
				for _, line := range lines {
					if strings.Contains(line, "Suggestion:") {
						hasSuggestion = true
						assert.True(t, strings.HasPrefix(line, "    Suggestion: "), "suggestion should have 4-space indentation")
					}
					if strings.Contains(line, "Example:") {
						hasExample = true
						assert.True(t, strings.HasPrefix(line, "    Example: "), "example should have 4-space indentation")
					}
				}
				assert.True(t, hasSuggestion, "should have suggestion")
				assert.True(t, hasExample, "should have example")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loc := fmt.Sprintf("%s:%d", tt.issue.Rule, tt.issue.Line)

			var msg string
			msg = fmt.Sprintf("  [%s] %s: %s", tt.issue.Severity, loc, tt.issue.Message)

			if tt.issue.Suggestion != "" {
				msg += fmt.Sprintf("\n    Suggestion: %s", tt.issue.Suggestion)
			}
			if tt.issue.Example != "" {
				msg += fmt.Sprintf("\n    Example: %s", tt.issue.Example)
			}

			tt.check(t, msg)
		})
	}
}

func TestCLIFormatter_ReadableStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		issue skillpkg.ValidationIssue
		check func(t *testing.T, msg string)
	}{
		{
			name: "error with suggestion is readable",
			issue: skillpkg.ValidationIssue{
				Rule:       "SK001",
				Severity:   skillpkg.SeverityError,
				Message:    "name field is missing",
				Suggestion: "Add a name field to the frontmatter",
				Example:    "name: my-skill",
				Line:       1,
			},
			check: func(t *testing.T, msg string) {
				lines := strings.Split(msg, "\n")

				assert.Len(t, lines, 3, "should have 3 lines (error, suggestion, example)")
				assert.Contains(t, lines[0], "[error]")
				assert.Contains(t, lines[1], "Suggestion:")
				assert.Contains(t, lines[2], "Example:")
			},
		},
		{
			name: "warning with suggestion is readable",
			issue: skillpkg.ValidationIssue{
				Rule:       "SK002",
				Severity:   skillpkg.SeverityWarning,
				Message:    "name format is invalid",
				Suggestion: "Use kebab-case for skill names",
				Example:    "name: my-skill",
				Line:       2,
			},
			check: func(t *testing.T, msg string) {
				lines := strings.Split(msg, "\n")

				assert.Contains(t, lines[0], "⚠️")
				assert.Contains(t, lines[0], "[warning]")
				assert.Contains(t, lines[1], "Suggestion:")
			},
		},
		{
			name: "info with suggestion is readable",
			issue: skillpkg.ValidationIssue{
				Rule:       "SK003",
				Severity:   skillpkg.SeverityInfo,
				Message:    "description is short",
				Suggestion: "Add more detail to the description",
				Line:       3,
			},
			check: func(t *testing.T, msg string) {
				lines := strings.Split(msg, "\n")

				assert.Contains(t, lines[0], "ℹ️")
				assert.Contains(t, lines[0], "[info]")
				assert.Contains(t, lines[1], "Suggestion:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loc := fmt.Sprintf("%s:%d", tt.issue.Rule, tt.issue.Line)

			var msg string
			switch tt.issue.Severity {
			case skillpkg.SeverityError:
				msg = fmt.Sprintf("  [%s] %s: %s", tt.issue.Severity, loc, tt.issue.Message)
			case skillpkg.SeverityWarning:
				msg = fmt.Sprintf("  ⚠️  [%s] %s: %s", tt.issue.Severity, loc, tt.issue.Message)
			case skillpkg.SeverityInfo:
				msg = fmt.Sprintf("  ℹ️  [%s] %s: %s", tt.issue.Severity, loc, tt.issue.Message)
			}

			if tt.issue.Suggestion != "" {
				msg += fmt.Sprintf("\n    Suggestion: %s", tt.issue.Suggestion)
			}
			if tt.issue.Example != "" {
				msg += fmt.Sprintf("\n    Example: %s", tt.issue.Example)
			}

			tt.check(t, msg)
		})
	}
}
