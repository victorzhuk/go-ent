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

	templatePath := filepath.Join("..", "..", "plugins", "go-ent", "templates", "skills", "go-basic", "template.md")

	templateContent, err := os.ReadFile(templatePath)
	require.NoError(t, err, "template file should exist")

	data := map[string]string{
		"SKILL_NAME":  "go-payment",
		"DESCRIPTION": "Go payment processing patterns for secure transactions",
		"VERSION":     "1.0.0",
		"AUTHOR":      "go-ent",
		"TAGS":        "go, payment, backend, security",
	}

	generatedContent, err := ReplacePlaceholders(string(templateContent), data)
	require.NoError(t, err, "placeholder replacement should succeed")

	assert.NotContains(t, generatedContent, "${SKILL_NAME}", "all placeholders should be replaced")
	assert.NotContains(t, generatedContent, "${DESCRIPTION}", "all placeholders should be replaced")
	assert.NotContains(t, generatedContent, "${VERSION}", "all placeholders should be replaced")
	assert.NotContains(t, generatedContent, "${AUTHOR}", "all placeholders should be replaced")
	assert.NotContains(t, generatedContent, "${TAGS}", "all placeholders should be replaced")

	parser := skill.NewParser()
	tempFile := filepath.Join(t.TempDir(), "generated-skill.md")
	err = os.WriteFile(tempFile, []byte(generatedContent), 0644)
	require.NoError(t, err, "should write generated skill to temp file")

	meta, err := parser.ParseSkillFile(tempFile)
	require.NoError(t, err, "generated skill should parse successfully")

	assert.Equal(t, "go-payment", meta.Name)
	assert.Equal(t, "Go payment processing patterns for secure transactions", meta.Description)
	assert.Equal(t, "1.0.0", meta.Version)
	assert.Equal(t, "go-ent", meta.Author)
	assert.Equal(t, []string{"go", "payment", "backend", "security"}, meta.Tags)
	assert.Equal(t, "v2", meta.StructureVersion)

	scorer := skill.NewQualityScorer()
	qualityScore := scorer.Score(meta, generatedContent)
	meta.QualityScore = qualityScore

	assert.GreaterOrEqual(t, qualityScore, 90.0, "quality score should be >= 90")

	validator := skill.NewValidator()
	result := validator.ValidateStrict(meta, generatedContent)

	t.Logf("Quality Score: %.2f", qualityScore)
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

	templatePath := filepath.Join("..", "..", "plugins", "go-ent", "templates", "skills", "go-basic", "template.md")

	templateContent, err := os.ReadFile(templatePath)
	require.NoError(t, err)

	data := map[string]string{
		"SKILL_NAME":  "go-payment",
		"DESCRIPTION": "Go payment processing patterns",
		"VERSION":     "1.0.0",
		"AUTHOR":      "go-ent",
		"TAGS":        "go,payment",
	}

	generatedContent, err := ReplacePlaceholders(string(templateContent), data)
	require.NoError(t, err)

	assert.Contains(t, generatedContent, "<role>", "should contain role section")
	assert.Contains(t, generatedContent, "<instructions>", "should contain instructions section")
	assert.Contains(t, generatedContent, "<constraints>", "should contain constraints section")
	assert.Contains(t, generatedContent, "<edge_cases>", "should contain edge_cases section")
	assert.Contains(t, generatedContent, "<examples>", "should contain examples section")
	assert.Contains(t, generatedContent, "<output_format>", "should contain output_format section")

	exampleCount := countOccurrences(generatedContent, "<example>")
	assert.GreaterOrEqual(t, exampleCount, 2, "should contain at least 2 examples")
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

	templatePath := filepath.Join("..", "..", "plugins", "go-ent", "templates", "skills", "go-complete", "template.md")

	templateContent, err := os.ReadFile(templatePath)
	require.NoError(t, err, "template file should exist")

	data := map[string]string{
		"SKILL_NAME":  "go-api-service",
		"DESCRIPTION": "Comprehensive Go API service implementation patterns with best practices",
		"VERSION":     "1.0.0",
		"AUTHOR":      "go-ent",
		"TAGS":        "go,api,backend,web",
	}

	generatedContent, err := ReplacePlaceholders(string(templateContent), data)
	require.NoError(t, err, "placeholder replacement should succeed")

	assert.NotContains(t, generatedContent, "${SKILL_NAME}", "all placeholders should be replaced")
	assert.NotContains(t, generatedContent, "${DESCRIPTION}", "all placeholders should be replaced")
	assert.NotContains(t, generatedContent, "${VERSION}", "all placeholders should be replaced")
	assert.NotContains(t, generatedContent, "${AUTHOR}", "all placeholders should be replaced")
	assert.NotContains(t, generatedContent, "${TAGS}", "all placeholders should be replaced")

	parser := skill.NewParser()
	tempFile := filepath.Join(t.TempDir(), "generated-skill.md")
	err = os.WriteFile(tempFile, []byte(generatedContent), 0644)
	require.NoError(t, err, "should write generated skill to temp file")

	meta, err := parser.ParseSkillFile(tempFile)
	require.NoError(t, err, "generated skill should parse successfully")

	assert.Equal(t, "go-api-service", meta.Name)
	assert.Equal(t, "Comprehensive Go API service implementation patterns with best practices", meta.Description)
	assert.Equal(t, "1.0.0", meta.Version)
	assert.Equal(t, "go-ent", meta.Author)
	assert.Equal(t, []string{"go", "api", "backend", "web"}, meta.Tags)
	assert.Equal(t, "v2", meta.StructureVersion)

	scorer := skill.NewQualityScorer()
	qualityScore := scorer.Score(meta, generatedContent)
	meta.QualityScore = qualityScore

	assert.GreaterOrEqual(t, qualityScore, 90.0, "quality score should be >= 90")

	validator := skill.NewValidator()
	result := validator.ValidateStrict(meta, generatedContent)

	t.Logf("Quality Score: %.2f", qualityScore)
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

	templatePath := filepath.Join("..", "..", "plugins", "go-ent", "templates", "skills", "go-complete", "template.md")

	templateContent, err := os.ReadFile(templatePath)
	require.NoError(t, err)

	data := map[string]string{
		"SKILL_NAME":  "go-api-service",
		"DESCRIPTION": "Go API service implementation patterns",
		"VERSION":     "1.0.0",
		"AUTHOR":      "go-ent",
		"TAGS":        "go, api",
	}

	generatedContent, err := ReplacePlaceholders(string(templateContent), data)
	require.NoError(t, err)

	assert.Contains(t, generatedContent, "<role>", "should contain role section")
	assert.Contains(t, generatedContent, "<instructions>", "should contain instructions section")
	assert.Contains(t, generatedContent, "<constraints>", "should contain constraints section")
	assert.Contains(t, generatedContent, "<edge_cases>", "should contain edge_cases section")
	assert.Contains(t, generatedContent, "<examples>", "should contain examples section")
	assert.Contains(t, generatedContent, "<output_format>", "should contain output_format section")

	exampleCount := countOccurrences(generatedContent, "<example>")
	assert.GreaterOrEqual(t, exampleCount, 3, "should contain at least 3 examples")

	edgeCaseCount := countEdgeCases(generatedContent)
	assert.GreaterOrEqual(t, edgeCaseCount, 5, "should contain at least 5 edge cases")
}

func countEdgeCases(content string) int {
	startIdx := strings.Index(content, "<edge_cases>")
	if startIdx == -1 {
		return 0
	}
	endIdx := strings.Index(content, "</edge_cases>")
	if endIdx == -1 {
		return 0
	}
	edgeCasesContent := content[startIdx:endIdx]
	count := strings.Count(edgeCasesContent, "\nIf")
	count += strings.Count(edgeCasesContent, "\n  If")
	count += strings.Count(edgeCasesContent, "\n    If")
	return count
}
