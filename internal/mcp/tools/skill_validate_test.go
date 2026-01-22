package tools

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/skill"
)

func TestRegisterSkillValidate(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := skill.NewRegistry()

	registerSkillValidate(s, registry)
}

func TestSkillValidateHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T) *skill.Registry
		input       SkillValidateInput
		expectError bool
		checkOutput func(t *testing.T, result *mcp.CallToolResult, data any)
	}{
		{
			name: "validate single valid skill",
			setup: func(t *testing.T) *skill.Registry {
				registry := skill.NewRegistry()
				tmpDir := t.TempDir()

				content := `---
name: test-skill
description: A test skill
version: v1
triggers:
  - test
---

# Role

Test role.

# Instructions

Test instructions.
`
				skillPath := filepath.Join(tmpDir, "SKILL.md")
				err := os.WriteFile(skillPath, []byte(content), 0600)
				require.NoError(t, err)

				err = registry.Load(tmpDir)
				require.NoError(t, err)

				return registry
			},
			input: SkillValidateInput{
				Name: "test-skill",
			},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				require.NotNil(t, result)
				require.Len(t, result.Content, 1)

				textContent, ok := result.Content[0].(*mcp.TextContent)
				require.True(t, ok, "Expected TextContent")

				assert.Contains(t, textContent.Text, "Validation: test-skill")
				assert.Contains(t, textContent.Text, "✓ VALID")
				assert.Contains(t, textContent.Text, "Quality Score")

				output, ok := data.(SkillValidateOutput)
				require.True(t, ok, "Expected SkillValidateOutput")
				assert.True(t, output.Valid)
				assert.Greater(t, output.Score, 0.0)
			},
		},
		{
			name: "validate all skills",
			setup: func(t *testing.T) *skill.Registry {
				registry := skill.NewRegistry()
				tmpDir := t.TempDir()

				content1 := `---
name: skill-one
description: First skill
version: v1
---

# Instructions
Test 1.
`
				content2 := `---
name: skill-two
description: Second skill
version: v1
---

# Instructions
Test 2.
`
				skillDir1 := filepath.Join(tmpDir, "skill1")
				err := os.MkdirAll(skillDir1, 0750)
				require.NoError(t, err)

				skillPath1 := filepath.Join(skillDir1, "SKILL.md")
				err = os.WriteFile(skillPath1, []byte(content1), 0600)
				require.NoError(t, err)

				skillDir2 := filepath.Join(tmpDir, "skill2")
				err = os.MkdirAll(skillDir2, 0750)
				require.NoError(t, err)

				skillPath2 := filepath.Join(skillDir2, "SKILL.md")
				err = os.WriteFile(skillPath2, []byte(content2), 0600)
				require.NoError(t, err)

				err = registry.Load(tmpDir)
				require.NoError(t, err)

				return registry
			},
			input: SkillValidateInput{
				Name: "",
			},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				require.NotNil(t, result)

				textContent, ok := result.Content[0].(*mcp.TextContent)
				require.True(t, ok, "Expected TextContent")

				assert.Contains(t, textContent.Text, "Validation: All Skills")

				output, ok := data.(SkillValidateOutput)
				require.True(t, ok, "Expected SkillValidateOutput")
				assert.True(t, output.Valid)
				assert.Greater(t, output.Score, 0.0)
			},
		},
		{
			name: "skill not found",
			setup: func(t *testing.T) *skill.Registry {
				registry := skill.NewRegistry()
				return registry
			},
			input: SkillValidateInput{
				Name: "nonexistent-skill",
			},
			expectError: true,
		},
		{
			name: "strict mode valid",
			setup: func(t *testing.T) *skill.Registry {
				registry := skill.NewRegistry()
				tmpDir := t.TempDir()

				content := `---
name: test-skill
description: A test skill
version: v1
---

# Role

Test role.

# Instructions

Test instructions.
`
				skillPath := filepath.Join(tmpDir, "SKILL.md")
				err := os.WriteFile(skillPath, []byte(content), 0600)
				require.NoError(t, err)

				err = registry.Load(tmpDir)
				require.NoError(t, err)

				return registry
			},
			input: SkillValidateInput{
				Name:   "test-skill",
				Strict: true,
			},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				require.NotNil(t, result)

				output, ok := data.(SkillValidateOutput)
				require.True(t, ok, "Expected SkillValidateOutput")
				assert.True(t, output.Valid, "Valid skill should pass strict mode")
			},
		},
		{
			name: "strict mode with warnings",
			setup: func(t *testing.T) *skill.Registry {
				registry := skill.NewRegistry()
				tmpDir := t.TempDir()

				content := `---
name: test-skill
description: A test skill
version: v2
---

# Instructions

Test instructions.
`
				skillPath := filepath.Join(tmpDir, "SKILL.md")
				err := os.WriteFile(skillPath, []byte(content), 0600)
				require.NoError(t, err)

				err = registry.Load(tmpDir)
				require.NoError(t, err)

				return registry
			},
			input: SkillValidateInput{
				Name:   "test-skill",
				Strict: true,
			},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				require.NotNil(t, result)

				output, ok := data.(SkillValidateOutput)
				require.True(t, ok, "Expected SkillValidateOutput")

				if len(output.Issues) > 0 {
					assert.False(t, output.Valid, "Skill with issues should fail strict mode")
					assert.Greater(t, len(output.Issues), 0, "Should have issues")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_ = mcp.NewServer(
				&mcp.Implementation{Name: "test", Version: "1.0.0"},
				nil,
			)
			registry := tt.setup(t)

			handler := skillValidateHandler(registry)
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, data, err := handler(ctx, req, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.checkOutput != nil {
				tt.checkOutput(t, result, data)
			}
		})
	}
}

func TestSkillValidateOutputFormatting(t *testing.T) {
	t.Parallel()

	_ = mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := skill.NewRegistry()
	tmpDir := t.TempDir()

	content := `---
name: test-skill
description: A test skill
version: v1
---

# Role

Test role.

# Instructions

Test instructions.
`
	skillPath := filepath.Join(tmpDir, "SKILL.md")
	err := os.WriteFile(skillPath, []byte(content), 0600)
	require.NoError(t, err)

	err = registry.Load(tmpDir)
	require.NoError(t, err)

	handler := skillValidateHandler(registry)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, _, err := handler(ctx, req, SkillValidateInput{Name: "test-skill"})
	require.NoError(t, err)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	outputText := textContent.Text

	assert.Contains(t, outputText, "# Validation:")
	assert.Contains(t, outputText, "**Status**:")
	assert.Contains(t, outputText, "✓ VALID")
	assert.Contains(t, outputText, "**Quality Score**:")
	assert.Contains(t, outputText, "No issues found.")
}

func TestSkillValidateWithIssues(t *testing.T) {
	t.Parallel()

	_ = mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := skill.NewRegistry()
	tmpDir := t.TempDir()

	content := `---
name: test-skill
description: A test skill
version: v1
---

# Instructions

Test instructions without role section.
`
	skillPath := filepath.Join(tmpDir, "SKILL.md")
	err := os.WriteFile(skillPath, []byte(content), 0600)
	require.NoError(t, err)

	err = registry.Load(tmpDir)
	require.NoError(t, err)

	handler := skillValidateHandler(registry)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, data, err := handler(ctx, req, SkillValidateInput{Name: "test-skill"})
	require.NoError(t, err)

	output, ok := data.(SkillValidateOutput)
	require.True(t, ok)

	if len(output.Issues) > 0 {
		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		outputText := textContent.Text
		assert.Contains(t, outputText, "## Issues")
		assert.Contains(t, outputText, "**Issues**:")
	}
}

func TestSkillValidateEmptyRegistry(t *testing.T) {
	t.Parallel()

	_ = mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := skill.NewRegistry()

	handler := skillValidateHandler(registry)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, data, err := handler(ctx, req, SkillValidateInput{})
	require.NoError(t, err)

	output, ok := data.(SkillValidateOutput)
	require.True(t, ok)

	assert.True(t, output.Valid)
	assert.Equal(t, 0.0, output.Score)
	assert.Empty(t, output.Issues)
}

func TestFormatValidationOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		output SkillValidateOutput
	}{
		{
			name:  "valid single skill",
			input: "test-skill",
			output: SkillValidateOutput{
				Valid:  true,
				Score:  85.0,
				Issues: []skill.ValidationIssue{},
			},
		},
		{
			name:  "all skills",
			input: "",
			output: SkillValidateOutput{
				Valid:  true,
				Score:  75.5,
				Issues: []skill.ValidationIssue{},
			},
		},
		{
			name:  "invalid with issues",
			input: "bad-skill",
			output: SkillValidateOutput{
				Valid: false,
				Score: 45.0,
				Issues: []skill.ValidationIssue{
					{
						Rule:     "frontmatter",
						Severity: skill.SeverityError,
						Message:  "missing name field",
						Line:     2,
					},
					{
						Rule:     "xml-tags",
						Severity: skill.SeverityWarning,
						Message:  "missing <role> section",
						Line:     5,
					},
				},
			},
		},
		{
			name:  "issues with suggestion and example",
			input: "test-skill",
			output: SkillValidateOutput{
				Valid: false,
				Score: 50.0,
				Issues: []skill.ValidationIssue{
					{
						Rule:       "SK002",
						Severity:   skill.SeverityError,
						Message:    "invalid name format",
						Suggestion: "use lowercase letters, numbers, and hyphens only",
						Example:    "my-skill-name",
						Line:       2,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			formatted := formatValidationOutput(tt.input, tt.output)

			assert.Contains(t, formatted, "# Validation:")
			assert.Contains(t, formatted, "**Status**:")
			assert.Contains(t, formatted, "Quality Score")

			if tt.output.Valid {
				assert.Contains(t, formatted, "✓ VALID")
			} else {
				assert.Contains(t, formatted, "✗ INVALID")
			}

			if len(tt.output.Issues) == 0 {
				assert.Contains(t, formatted, "No issues found.")
			} else {
				assert.Contains(t, formatted, "**Issues**:")
				assert.Contains(t, formatted, "## Issues")

				for _, issue := range tt.output.Issues {
					assert.Contains(t, formatted, string(issue.Severity))
					assert.Contains(t, formatted, issue.Rule)
					assert.Contains(t, formatted, issue.Message)

					if issue.Suggestion != "" {
						assert.Contains(t, formatted, "**Suggestion**:")
						assert.Contains(t, formatted, issue.Suggestion)
					}

					if issue.Example != "" {
						assert.Contains(t, formatted, "**Example**:")
						assert.Contains(t, formatted, issue.Example)
					}
				}
			}
		})
	}
}
