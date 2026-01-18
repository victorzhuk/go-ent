package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/victorzhuk/go-ent/internal/skill"
)

func TestSkillValidateAndQualityRegistration(t *testing.T) {
	t.Parallel()

	registry := skill.NewRegistry()

	// Create temporary directory for test skills
	tmpDir := t.TempDir()
	skillsPath := filepath.Join(tmpDir, "skills")
	err := os.Mkdir(skillsPath, 0755)
	assert.NoError(t, err)

	// Register a sample skill
	skillDir := filepath.Join(skillsPath, "test-skill")
	err = os.Mkdir(skillDir, 0755)
	assert.NoError(t, err)

	skillContent := `---
name: test-skill
description: Test skill for validation
---

# Test Skill

## Description
Test skill description

## When to Use
Test conditions

## Triggers
- test
- validation
`

	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644)
	assert.NoError(t, err)

	err = registry.Load(skillsPath)
	assert.NoError(t, err)

	// Test skill_validate handler
	validateHandler := skillValidateHandler(registry)
	ctx := context.Background()

	result, data, err := validateHandler(ctx, nil, SkillValidateInput{Name: "test-skill"})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, data)

	// Verify result content is not empty
	assert.Len(t, result.Content, 1)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "Validation")

	// Test skill_quality handler
	qualityHandler := skillQualityHandler(registry)

	result2, data2, err := qualityHandler(ctx, nil, SkillQualityInput{})
	assert.NoError(t, err)
	assert.NotNil(t, result2)
	assert.NotNil(t, data2)

	// Verify result content is not empty
	assert.Len(t, result2.Content, 1)
	assert.Contains(t, result2.Content[0].(*mcp.TextContent).Text, "Quality Report")
}

func TestSkillValidateAll(t *testing.T) {
	t.Parallel()

	registry := skill.NewRegistry()

	// Create temporary directory for test skills
	tmpDir := t.TempDir()
	skillsPath := filepath.Join(tmpDir, "skills")
	err := os.Mkdir(skillsPath, 0755)
	assert.NoError(t, err)

	// Register multiple skills
	skillDir1 := filepath.Join(skillsPath, "skill-one")
	err = os.Mkdir(skillDir1, 0755)
	assert.NoError(t, err)

	skill1 := `---
name: skill-one
description: First skill
---

# Skill One
## Description
First skill`
	err = os.WriteFile(filepath.Join(skillDir1, "SKILL.md"), []byte(skill1), 0644)
	assert.NoError(t, err)

	skillDir2 := filepath.Join(skillsPath, "skill-two")
	err = os.Mkdir(skillDir2, 0755)
	assert.NoError(t, err)

	skill2 := `---
name: skill-two
description: Second skill
---

# Skill Two
## Description
Second skill`
	err = os.WriteFile(filepath.Join(skillDir2, "SKILL.md"), []byte(skill2), 0644)
	assert.NoError(t, err)

	err = registry.Load(skillsPath)
	assert.NoError(t, err)

	handler := skillValidateHandler(registry)
	ctx := context.Background()

	result, data, err := handler(ctx, nil, SkillValidateInput{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, data)
}

func TestSkillQualityWithThreshold(t *testing.T) {
	t.Parallel()

	registry := skill.NewRegistry()

	// Create temporary directory for test skills
	tmpDir := t.TempDir()
	skillsPath := filepath.Join(tmpDir, "skills")
	err := os.Mkdir(skillsPath, 0755)
	assert.NoError(t, err)

	// Register skills with different quality
	goodDir := filepath.Join(skillsPath, "good-skill")
	err = os.Mkdir(goodDir, 0755)
	assert.NoError(t, err)

	goodSkill := `---
name: good-skill
description: A comprehensive skill
---

# Good Skill

## Description
A comprehensive skill with full documentation

## When to Use
Detailed use case description

## Examples
Example code snippets

## Triggers
- good
`
	err = os.WriteFile(filepath.Join(goodDir, "SKILL.md"), []byte(goodSkill), 0644)
	assert.NoError(t, err)

	minimalDir := filepath.Join(skillsPath, "minimal-skill")
	err = os.Mkdir(minimalDir, 0755)
	assert.NoError(t, err)

	minimalSkill := `---
name: minimal-skill
description: Minimal skill
---

# Minimal
Basic info

## Triggers
- minimal
`
	err = os.WriteFile(filepath.Join(minimalDir, "SKILL.md"), []byte(minimalSkill), 0644)
	assert.NoError(t, err)

	err = registry.Load(skillsPath)
	assert.NoError(t, err)

	handler := skillQualityHandler(registry)
	ctx := context.Background()

	result, data, err := handler(ctx, nil, SkillQualityInput{Threshold: 50.0})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, data)
}
