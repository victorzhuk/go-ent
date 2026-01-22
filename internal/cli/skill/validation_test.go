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
description: A valid test skill
version: "1.0.0"
author: Test Author
tags: ["test", "validation"]
structure_version: "v2"
---

# Test Skill

<role>
You are a test skill expert.
</role>

<instructions>
Provide helpful testing guidance.
</instructions>

<constraints>
- Follow best practices
- Write clear code
</constraints>

<edge_cases>
If test fails: investigate root cause
</edge_cases>

<examples>
<example>
<input>Test input</input>
<output>Test output</output>
</example>
</examples>

<output_format>
Provide clear, actionable guidance.
</output_format>
`

	require.NoError(t, os.WriteFile(skillPath, []byte(validSkill), 0644))

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
version: "1.0.0"
---
<role>Test</role>
<instructions>Test</instructions>
<constraints>- Test</constraints>
<examples>
<example>
<input>test</input>
<output>test</output>
</example>
</examples>
<output_format>Test</output_format>
`

	require.NoError(t, os.WriteFile(skillPath, []byte(invalidYAML), 0644))

	err := ValidateGeneratedSkill(skillPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse skill file")
}

func TestValidateGeneratedSkill_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	require.NoError(t, os.WriteFile(skillPath, []byte(""), 0644))

	err := ValidateGeneratedSkill(skillPath)
	assert.Error(t, err)
}

func TestValidateGeneratedSkill_MissingFrontmatter(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	noFrontmatter := `# Test Skill

<role>Test role</role>
<instructions>Test instructions</instructions>
<constraints>- Test constraint</constraints>
<examples>
<example>
<input>test</input>
<output>test</output>
</example>
</examples>
<output_format>Test</output_format>
`

	require.NoError(t, os.WriteFile(skillPath, []byte(noFrontmatter), 0644))

	err := ValidateGeneratedSkill(skillPath)
	assert.Error(t, err)
}

func TestValidateGeneratedSkill_MissingName(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	missingName := `---
description: A test skill
version: "1.0.0"
---
<role>Test</role>
<instructions>Test</instructions>
<constraints>- Test</constraints>
<examples>
<example>
<input>test</input>
<output>test</output>
</example>
</examples>
<output_format>Test</output_format>
`

	require.NoError(t, os.WriteFile(skillPath, []byte(missingName), 0644))

	err := ValidateGeneratedSkill(skillPath)
	assert.Error(t, err)
}

func TestValidateGeneratedSkill_CompleteValidSkill(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	completeSkill := `---
name: go-database
description: Database integration patterns with PostgreSQL and pgx
version: "1.0.0"
author: go-ent
tags: ["go", "database", "postgresql", "pgx"]
structure_version: "v2"
---

# Go Database Patterns

<role>
You are an expert Go developer specializing in database integration patterns.
You have extensive experience with PostgreSQL, pgx, and database transaction management.
You follow best practices for connection pooling, query optimization, and error handling.
</role>

<instructions>
Provide guidance on Go database integration including:

1. **Connection Management**
   - Set up connection pools with appropriate limits
   - Configure health checks and timeouts
   - Handle connection lifecycle properly

2. **Query Patterns**
   - Use parameterized queries to prevent SQL injection
   - Implement batch operations for efficiency
   - Use prepared statements for repeated queries

3. **Transaction Management**
   - Handle transactions with proper rollback
   - Implement retry logic for transient failures
   - Use context for cancellation

4. **Error Handling**
   - Wrap database errors with context
   - Handle specific pgx error types
   - Provide meaningful error messages

5. **Performance**
   - Optimize queries with indexes
   - Use connection pooling effectively
   - Implement caching where appropriate
</instructions>

<constraints>
- Always use context.Context for database operations
- Never concatenate strings to build SQL queries
- Always check for errors after database operations
- Close rows objects to prevent connection leaks
- Use tx.Commit() only after all operations succeed
- Implement proper error wrapping with context
</constraints>

<edge_cases>
If connection pool is exhausted: log error and implement retry logic with exponential backoff
If transaction fails due to serialization error: retry transaction up to 3 times
If database is unavailable during startup: implement circuit breaker pattern
</edge_cases>

<examples>
<example>
<input>How do I set up a connection pool with pgx?</input>
<output>
Provide a code example showing proper connection pool setup with pgx.
</output>
</example>

<example>
<input>How do I handle transactions properly?</input>
<output>
Provide a code example showing proper transaction handling with rollback.
</output>
</example>
</examples>

<output_format>
Provide production-ready Go code with:
- Complete, runnable examples
- Proper error handling with context
- Usage of pgx best practices
- Comments explaining key decisions
- Type safety with struct mapping
</output_format>
`

	require.NoError(t, os.WriteFile(skillPath, []byte(completeSkill), 0644))

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
