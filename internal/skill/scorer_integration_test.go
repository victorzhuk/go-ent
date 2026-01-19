package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDataPath() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "..", "..", "plugins", "go-ent", "skills")
}

func TestQualityScorer_Integration_WithRealSkills(t *testing.T) {
	t.Parallel()

	skillsPath := getTestDataPath()

	if _, err := os.Stat(skillsPath); os.IsNotExist(err) {
		t.Skip("skills directory not found, skipping integration test")
		return
	}

	scorer := NewQualityScorer()

	entries, err := os.ReadDir(skillsPath)
	require.NoError(t, err)

	skillsLoaded := 0

	for _, entry := range entries {
		if entry.IsDir() {
			skillPath := filepath.Join(skillsPath, entry.Name())
			entries2, err := os.ReadDir(skillPath)
			if err != nil {
				continue
			}

			for _, subEntry := range entries2 {
				if !subEntry.IsDir() {
					continue
				}

				skillFile := filepath.Join(skillPath, subEntry.Name(), "SKILL.md")

				if _, err := os.Stat(skillFile); os.IsNotExist(err) {
					continue
				}

				parser := NewParser()
				meta, err := parser.ParseSkillFile(skillFile)
				if err != nil {
					continue
				}

				content, err := os.ReadFile(skillFile)
				require.NoError(t, err)

				result := scorer.Score(meta, string(content))

				assert.GreaterOrEqual(t, result.Total, 0.0,
					"Total score should be >= 0")
				assert.LessOrEqual(t, result.Total, 120.0,
					"Total score should be <= 120")

				assert.GreaterOrEqual(t, result.Structure.Total, 0.0)
				assert.LessOrEqual(t, result.Structure.Total, 20.0)

				assert.GreaterOrEqual(t, result.Content.Total, 0.0)
				assert.LessOrEqual(t, result.Content.Total, 25.0)

				assert.GreaterOrEqual(t, result.Examples.Total, 0.0)
				assert.LessOrEqual(t, result.Examples.Total, 25.0)

				assert.GreaterOrEqual(t, result.Triggers, 0.0)
				assert.LessOrEqual(t, result.Triggers, 15.0)

				assert.GreaterOrEqual(t, result.Conciseness, 0.0)
				assert.LessOrEqual(t, result.Conciseness, 15.0)

				skillsLoaded++
			}
		}
	}

	if skillsLoaded == 0 {
		t.Skip("no valid skills found in test directory")
	}
}

func TestQualityScorer_Integration_BreakdownSumsToTotal(t *testing.T) {
	t.Parallel()

	scorer := NewQualityScorer()

	content := `<role>Expert Go developer</role>
<instructions>Implement features using clean patterns</instructions>
<constraints>- Follow SOLID
- Avoid globals</constraints>
<examples>
<example>
<input>Refactor code</input>
<output>Clean implementation</output>
</example>
<example>
<input>Add test</input>
<output>Test case</output>
</example>
</examples>
<output_format>Clear output</output_format>
<edge_cases>Handle errors</edge_cases>`

	meta := &SkillMeta{
		Name:        "test-skill",
		Description: "Test description",
		Version:     "1.0.0",
		Tags:        []string{"test"},
	}

	result := scorer.Score(meta, content)

	expectedTotal := result.Structure.Total +
		result.Content.Total +
		result.Examples.Total +
		result.Triggers +
		result.Conciseness

	componentTotal := 20.0 + expectedTotal

	assert.Equal(t, result.Total, componentTotal,
		"Total should equal frontmatter (20) + sum of components")

	assert.GreaterOrEqual(t, result.Structure.Total, 0.0)
	assert.GreaterOrEqual(t, result.Content.Total, 0.0)
	assert.GreaterOrEqual(t, result.Examples.Total, 0.0)
	assert.GreaterOrEqual(t, result.Triggers, 0.0)
	assert.GreaterOrEqual(t, result.Conciseness, 0.0)
}

func TestQualityScorer_Integration_RealSkillExample(t *testing.T) {
	t.Parallel()

	skillsPath := getTestDataPath()

	goCodePath := filepath.Join(skillsPath, "go", "go-code", "SKILL.md")
	if _, err := os.Stat(goCodePath); os.IsNotExist(err) {
		t.Skip("go-code skill not found, skipping real skill test")
		return
	}

	parser := NewParser()
	meta, err := parser.ParseSkillFile(goCodePath)
	require.NoError(t, err)

	content, err := os.ReadFile(goCodePath)
	require.NoError(t, err)

	scorer := NewQualityScorer()
	result := scorer.Score(meta, string(content))

	t.Logf("go-code skill scores:")
	t.Logf("  Total: %.2f", result.Total)
	t.Logf("  Structure: %.2f", result.Structure.Total)
	t.Logf("  Content: %.2f", result.Content.Total)
	t.Logf("  Examples: %.2f", result.Examples.Total)
	t.Logf("  Triggers: %.2f", result.Triggers)
	t.Logf("  Conciseness: %.2f", result.Conciseness)

	assert.Greater(t, result.Structure.Total, 15.0,
		"go-code should have most sections")
	assert.GreaterOrEqual(t, result.Examples.Total, 10.0,
		"go-code should have 3+ examples")
	assert.Greater(t, result.Triggers, 5.0,
		"go-code should have some triggers from description")
}
