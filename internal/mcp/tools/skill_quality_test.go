package tools

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

func setupTestSkillRegistry(t *testing.T, skillsDir string) *skill.Registry {
	t.Helper()

	registry := skill.NewRegistry()

	err := os.MkdirAll(skillsDir, 0750)
	require.NoError(t, err)

	skillContent := `---
name: test_skill_1
description: Test skill 1
triggers:
  - test
  - example
tags:
  - test
---

# Test Skill 1

This is a test skill with comprehensive documentation.
`

	err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(skillContent), 0600)
	require.NoError(t, err)

	skillContent2 := `---
name: test_skill_2
description: Test skill 2
triggers:
  - another
tags:
  - example
---

# Test Skill 2

Another test skill.
`

	skill2Dir := filepath.Join(skillsDir, "skill2")
	err = os.MkdirAll(skill2Dir, 0750)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(skillContent2), 0600)
	require.NoError(t, err)

	err = registry.Load(skillsDir)
	require.NoError(t, err)

	return registry
}

func TestRegisterSkillQuality(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	skillsDir := t.TempDir()
	registry := setupTestSkillRegistry(t, skillsDir)

	registerSkillQuality(s, registry)
}

func TestSkillQualityHandler(t *testing.T) {
	t.Parallel()

	skillsDir := t.TempDir()
	registry := setupTestSkillRegistry(t, skillsDir)

	t.Run("quality report with multiple skills", func(t *testing.T) {
		t.Parallel()

		input := SkillQualityInput{}
		ctx := context.Background()

		handler := skillQualityHandler(registry)
		result, output, err := handler(ctx, &mcp.CallToolRequest{}, input)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Content, 1)

		qualityOutput, ok := output.(SkillQualityOutput)
		require.True(t, ok)

		assert.Greater(t, len(qualityOutput.Skills), 0, "should have skills")
		assert.Greater(t, qualityOutput.AvgScore, 0.0, "avg score should be positive")
		assert.Empty(t, qualityOutput.BelowThresh, "no threshold set")

		for _, skill := range qualityOutput.Skills {
			assert.NotEmpty(t, skill.Name)
			assert.GreaterOrEqual(t, skill.Score, 0.0)
			assert.LessOrEqual(t, skill.Score, 100.0)
		}

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		assert.Contains(t, textContent.Text, "Skill Quality Report")
		assert.Contains(t, textContent.Text, "Average Score")
	})

	t.Run("quality report with threshold filter", func(t *testing.T) {
		t.Parallel()

		input := SkillQualityInput{Threshold: 95.0}
		ctx := context.Background()

		handler := skillQualityHandler(registry)
		result, output, err := handler(ctx, &mcp.CallToolRequest{}, input)

		require.NoError(t, err)
		require.NotNil(t, result)

		qualityOutput, ok := output.(SkillQualityOutput)
		require.True(t, ok)

		assert.NotEmpty(t, qualityOutput.BelowThresh, "should have skills below threshold")
		assert.Greater(t, qualityOutput.AvgScore, 0.0)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		assert.Contains(t, textContent.Text, "Threshold Filter")
		assert.Contains(t, textContent.Text, "Below Threshold")
	})

	t.Run("quality report with no threshold", func(t *testing.T) {
		t.Parallel()

		input := SkillQualityInput{Threshold: 0}
		ctx := context.Background()

		handler := skillQualityHandler(registry)
		result, output, err := handler(ctx, &mcp.CallToolRequest{}, input)

		require.NoError(t, err)
		require.NotNil(t, result)

		qualityOutput, ok := output.(SkillQualityOutput)
		require.True(t, ok)

		assert.Empty(t, qualityOutput.BelowThresh, "no threshold set, should be empty")

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		assert.NotContains(t, textContent.Text, "Threshold Filter")
	})

	t.Run("empty registry", func(t *testing.T) {
		t.Parallel()

		emptyRegistry := skill.NewRegistry()
		ctx := context.Background()

		handler := skillQualityHandler(emptyRegistry)
		result, output, err := handler(ctx, &mcp.CallToolRequest{}, SkillQualityInput{})

		require.NoError(t, err)
		require.NotNil(t, result)

		qualityOutput, ok := output.(SkillQualityOutput)
		require.True(t, ok)

		assert.Empty(t, qualityOutput.Skills)
		assert.Equal(t, 0.0, qualityOutput.AvgScore)
		assert.Empty(t, qualityOutput.BelowThresh)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		assert.Contains(t, textContent.Text, "No skills loaded")
	})

	t.Run("output formatting", func(t *testing.T) {
		t.Parallel()

		input := SkillQualityInput{Threshold: 50.0}
		ctx := context.Background()

		handler := skillQualityHandler(registry)
		result, _, err := handler(ctx, &mcp.CallToolRequest{}, input)

		require.NoError(t, err)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		text := textContent.Text
		assert.Contains(t, text, "# Skill Quality Report")
		assert.Contains(t, text, "**Average Score**")
		assert.Contains(t, text, "/100")
		assert.Contains(t, text, "## Scores")
		assert.Contains(t, text, "**Total Skills**")
	})

	t.Run("zero threshold behaves like no threshold", func(t *testing.T) {
		t.Parallel()

		input := SkillQualityInput{Threshold: 0}
		ctx := context.Background()

		handler := skillQualityHandler(registry)
		_, output, err := handler(ctx, &mcp.CallToolRequest{}, input)

		require.NoError(t, err)

		qualityOutput, ok := output.(SkillQualityOutput)
		require.True(t, ok)

		assert.Empty(t, qualityOutput.BelowThresh, "zero threshold should not filter")
	})
}
