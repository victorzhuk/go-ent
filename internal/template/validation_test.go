package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/skill"
)

func TestGoBasicTemplate_ValidateStrictMode(t *testing.T) {
	t.Parallel()

	templatePath := filepath.Join("..", "..", "pkg", "templates", "go-basic", "template.md")

	templateContent, err := os.ReadFile(templatePath)
	require.NoError(t, err, "template file should exist")

	data := map[string]string{
		"SKILL_NAME":  "go-payment",
		"DESCRIPTION": "Go payment processing patterns for secure transactions",
	}

	generatedContent, err := ReplacePlaceholders(string(templateContent), data)
	require.NoError(t, err, "placeholder replacement should succeed")

	assert.NotContains(t, generatedContent, "${SKILL_NAME}", "all placeholders should be replaced")
	assert.NotContains(t, generatedContent, "${DESCRIPTION}", "all placeholders should be replaced")

	parser := skill.NewParser()
	tempFile := filepath.Join(t.TempDir(), "generated-skill.md")
	err = os.WriteFile(tempFile, []byte(generatedContent), 0o644)
	require.NoError(t, err, "should write generated skill to temp file")

	meta, err := parser.ParseSkillFile(tempFile)
	require.NoError(t, err, "generated skill should parse successfully")

	assert.Equal(t, "go-payment", meta.Name)
	assert.Equal(t, "Go payment processing patterns for secure transactions", meta.Description)
	assert.Equal(t, "v4", meta.StructureVersion)

	validator := skill.NewValidator()
	result := validator.ValidateStrict(meta, generatedContent)

	t.Logf("Valid: %t", result.Valid)
	t.Logf("Total Issues: %d", len(result.Issues))
	t.Logf("Errors: %d", result.ErrorCount())
	t.Logf("Warnings: %d", result.WarningCount())

	if len(result.Issues) > 0 {
		t.Logf("Issues:")
		for _, issue := range result.Issues {
			t.Logf("  %s", issue.String())
		}
	}

	assert.True(t, result.Valid, "validation should pass in strict mode")
	assert.Equal(t, 0, result.ErrorCount(), "no errors should be present")
}

func TestGoBasicTemplate_StructuralValidation(t *testing.T) {
	t.Parallel()

	templatePath := filepath.Join("..", "..", "pkg", "templates", "go-basic", "template.md")

	templateContent, err := os.ReadFile(templatePath)
	require.NoError(t, err)

	data := map[string]string{
		"SKILL_NAME":  "go-payment",
		"DESCRIPTION": "Go payment processing patterns",
	}

	generatedContent, err := ReplacePlaceholders(string(templateContent), data)
	require.NoError(t, err)

	assert.Contains(t, generatedContent, "## Role", "should contain Role section")
	assert.Contains(t, generatedContent, "## Instructions", "should contain Instructions section")
	assert.Contains(t, generatedContent, "## Examples", "should contain Examples section")
	assert.Contains(t, generatedContent, "### Example", "should contain at least one example")
	assert.Contains(t, generatedContent, "**Input**", "should contain Input marker")
	assert.Contains(t, generatedContent, "**Output**", "should contain Output marker")
}

func countOccurrences(s, substr string) int {
	count := 0
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			count++
		}
	}
	return count
}

func TestGoCompleteTemplate_ValidateStrictMode(t *testing.T) {
	t.Parallel()

	templatePath := filepath.Join("..", "..", "pkg", "templates", "go-complete", "template.md")

	templateContent, err := os.ReadFile(templatePath)
	require.NoError(t, err, "template file should exist")

	data := map[string]string{
		"SKILL_NAME":  "go-api-service",
		"DESCRIPTION": "Comprehensive Go API service implementation patterns with best practices",
	}

	generatedContent, err := ReplacePlaceholders(string(templateContent), data)
	require.NoError(t, err, "placeholder replacement should succeed")

	assert.NotContains(t, generatedContent, "${SKILL_NAME}", "all placeholders should be replaced")
	assert.NotContains(t, generatedContent, "${DESCRIPTION}", "all placeholders should be replaced")

	parser := skill.NewParser()
	tempFile := filepath.Join(t.TempDir(), "generated-skill.md")
	err = os.WriteFile(tempFile, []byte(generatedContent), 0o644)
	require.NoError(t, err, "should write generated skill to temp file")

	meta, err := parser.ParseSkillFile(tempFile)
	require.NoError(t, err, "generated skill should parse successfully")

	assert.Equal(t, "go-api-service", meta.Name)
	assert.Equal(t, "Comprehensive Go API service implementation patterns with best practices", meta.Description)
	assert.Equal(t, "v4", meta.StructureVersion)

	validator := skill.NewValidator()
	result := validator.ValidateStrict(meta, generatedContent)

	t.Logf("Valid: %t", result.Valid)
	t.Logf("Total Issues: %d", len(result.Issues))
	t.Logf("Errors: %d", result.ErrorCount())
	t.Logf("Warnings: %d", result.WarningCount())

	if len(result.Issues) > 0 {
		t.Logf("Issues:")
		for _, issue := range result.Issues {
			t.Logf("  %s", issue.String())
		}
	}

	assert.True(t, result.Valid, "validation should pass in strict mode")
	assert.Equal(t, 0, result.ErrorCount(), "no errors should be present")
}

func TestGoCompleteTemplate_StructuralValidation(t *testing.T) {
	t.Parallel()

	templatePath := filepath.Join("..", "..", "pkg", "templates", "go-complete", "template.md")

	templateContent, err := os.ReadFile(templatePath)
	require.NoError(t, err)

	data := map[string]string{
		"SKILL_NAME":  "go-api-service",
		"DESCRIPTION": "Go API service implementation patterns",
	}

	generatedContent, err := ReplacePlaceholders(string(templateContent), data)
	require.NoError(t, err)

	assert.Contains(t, generatedContent, "## Role", "should contain Role section")
	assert.Contains(t, generatedContent, "## Instructions", "should contain Instructions section")
	assert.Contains(t, generatedContent, "## Examples", "should contain Examples section")

	exampleCount := countOccurrences(generatedContent, "### Example")
	assert.GreaterOrEqual(t, exampleCount, 2, "should contain at least 2 examples")

	edgeCaseCount := countEdgeCases(generatedContent)
	assert.GreaterOrEqual(t, edgeCaseCount, 3, "should contain at least 3 edge cases")
}

func countEdgeCases(content string) int {
	startIdx := strings.Index(content, "## Instructions")
	if startIdx == -1 {
		return 0
	}
	instructionsContent := content[startIdx:]
	count := strings.Count(instructionsContent, "\nIf ")
	count += strings.Count(instructionsContent, "\n  If ")
	count += strings.Count(instructionsContent, "\n    If ")
	return count
}
