package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skillpkg "github.com/victorzhuk/go-ent/internal/skill"
)

func TestInvalidSkill_3_2_3_HelpfulErrors(t *testing.T) {
	t.Parallel()

	t.Run("skill with multiple validation errors shows suggestions and examples", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")

		invalidSkill := `---
name: Invalid@Name
description: Short
triggers:
  - test
---

## Role

Short.

## Instructions

Too short.

## Examples

No proper example.
`

		require.NoError(t, os.WriteFile(skillPath, []byte(invalidSkill), 0o644))

		parser := skillpkg.NewParser()
		validator := skillpkg.NewValidator()

		meta, err := parser.ParseSkillFile(skillPath)
		require.NoError(t, err, "should parse file")

		content, err := os.ReadFile(skillPath)
		require.NoError(t, err)

		result := validator.Validate(meta, string(content))

		assert.False(t, result.Valid, "skill should be invalid")
		assert.Greater(t, result.ErrorCount(), 0, "should have errors")

		foundSuggestions := 0
		foundExamples := 0

		for _, issue := range result.Issues {
			if issue.Suggestion != "" {
				foundSuggestions++
			}
			if issue.Example != "" {
				foundExamples++
			}
		}

		assert.Greater(t, foundSuggestions, 0, "should have at least one suggestion")
		assert.Greater(t, foundExamples, 0, "should have at least one example")
	})

	t.Run("invalid name format shows helpful error with suggestion", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")

		invalidNameSkill := `---
name: Invalid@Name#With$Symbols
description: A test skill for expert guidance
triggers:
  - test
---

## Role

Expert test assistant for A test skill for expert guidance with focus on quality and best practices.

## Instructions

### Core Approach

Provide helpful guidance using standard patterns:

` + "```" + `go
func Process() error {
    return nil
}
` + "```" + `

### Edge Cases

If unclear: ask clarifying questions.

## Examples

### Example 1: Basic usage

**Input**: test

**Output**: test response
`

		require.NoError(t, os.WriteFile(skillPath, []byte(invalidNameSkill), 0o644))

		parser := skillpkg.NewParser()
		validator := skillpkg.NewValidator()

		meta, err := parser.ParseSkillFile(skillPath)
		require.NoError(t, err)

		content, err := os.ReadFile(skillPath)
		require.NoError(t, err)

		result := validator.Validate(meta, string(content))

		assert.False(t, result.Valid)

		var nameError *skillpkg.ValidationIssue
		for _, issue := range result.Issues {
			if issue.Rule == "name-format" {
				nameError = &issue
				break
			}
		}

		require.NotNil(t, nameError, "should have name-format error for invalid name")
		assert.NotEmpty(t, nameError.Suggestion, "name error should have suggestion")
		assert.NotEmpty(t, nameError.Example, "name error should have example")
		assert.Contains(t, nameError.Message, "invalid name format")
	})

	t.Run("missing description shows helpful error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")

		// The v4 parser itself validates description presence, so parse fails before validator runs.
		// We verify the error message mentions "description" to ensure it's a helpful error.
		missingDescSkill := `---
name: test-skill
triggers:
  - test
---

## Role

Expert test assistant with focus on quality and best practices.

## Instructions

### Core Approach

Provide helpful guidance using standard patterns.

## Examples

### Example 1: Basic usage

**Input**: test

**Output**: test response
`

		require.NoError(t, os.WriteFile(skillPath, []byte(missingDescSkill), 0o644))

		parser := skillpkg.NewParser()

		_, err := parser.ParseSkillFile(skillPath)
		require.Error(t, err, "skill with missing description should fail to parse")
		assert.Contains(t, err.Error(), "description", "parse error should mention description")
	})

	t.Run("empty role section shows helpful error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")

		emptyRoleSkill := `---
name: test-skill
description: A test skill with empty role section
triggers:
  - test
---

## Role

## Instructions

### Core Approach

Provide helpful guidance using standard patterns.

## Examples

### Example 1: Basic usage

**Input**: test

**Output**: test response
`

		require.NoError(t, os.WriteFile(skillPath, []byte(emptyRoleSkill), 0o644))

		parser := skillpkg.NewParser()
		validator := skillpkg.NewValidator()

		meta, err := parser.ParseSkillFile(skillPath)
		require.NoError(t, err)

		content, err := os.ReadFile(skillPath)
		require.NoError(t, err)

		result := validator.Validate(meta, string(content))

		var roleError *skillpkg.ValidationIssue
		for _, issue := range result.Issues {
			if issue.Rule == "role-section" {
				roleError = &issue
				break
			}
		}

		require.NotNil(t, roleError, "should have role-section error for empty role")
		assert.NotEmpty(t, roleError.Suggestion, "role error should have suggestion")
		assert.NotEmpty(t, roleError.Example, "role error should have example")
		assert.Contains(t, strings.ToLower(roleError.Suggestion), "role")
	})

	t.Run("CLI formatter displays errors with suggestions", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")

		invalidSkill := `---
name: Invalid@Name
description: Short
triggers:
  - test
---

## Role

## Instructions

## Examples
`

		require.NoError(t, os.WriteFile(skillPath, []byte(invalidSkill), 0o644))

		err := ValidateGeneratedSkill(skillPath)

		assert.Error(t, err, "validation should fail")

		errorMsg := err.Error()

		assert.Contains(t, errorMsg, "Suggestion:", "error message should contain suggestion prefix")
		assert.Contains(t, errorMsg, "Example:", "error message should contain example prefix")
	})

	t.Run("each error has suggestion and example fields", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")

		invalidSkill := `---
name: Invalid@Name
description: Short
triggers:
  - test
---

## Role

## Instructions

## Examples
`

		require.NoError(t, os.WriteFile(skillPath, []byte(invalidSkill), 0o644))

		parser := skillpkg.NewParser()
		validator := skillpkg.NewValidator()

		meta, err := parser.ParseSkillFile(skillPath)
		require.NoError(t, err)

		content, err := os.ReadFile(skillPath)
		require.NoError(t, err)

		result := validator.Validate(meta, string(content))

		assert.Greater(t, len(result.Issues), 0, "should have validation issues")

		for _, issue := range result.Issues {
			if issue.Rule == "name-format" || issue.Rule == "frontmatter" ||
				issue.Rule == "role-section" || issue.Rule == "instructions-section" {
				assert.NotEmpty(t, issue.Suggestion,
					"error %s should have suggestion", issue.Rule)
				assert.NotEmpty(t, issue.Example,
					"error %s should have example", issue.Rule)
			}
		}
	})

	t.Run("invalid semantic version shows helpful error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")

		// v4 format doesn't have a version field in frontmatter - it only has name, description, triggers
		// So "invalid version" isn't directly testable with v4 parser.
		// Instead we test that a skill with empty triggers fails validation.
		// Note: triggers: key must be present for v4 detection to work; empty list is tested by validator.
		missingTriggersSkill := `---
name: test-skill
description: A test skill for expert guidance
triggers: []
---

## Role

Expert test assistant for A test skill with focus on best practices.

## Instructions

### Core Approach

Provide helpful guidance using standard patterns.

## Examples

### Example 1: Basic usage

**Input**: test

**Output**: test response
`

		require.NoError(t, os.WriteFile(skillPath, []byte(missingTriggersSkill), 0o644))

		parser := skillpkg.NewParser()
		validator := skillpkg.NewValidator()

		meta, err := parser.ParseSkillFile(skillPath)
		require.NoError(t, err)

		content, err := os.ReadFile(skillPath)
		require.NoError(t, err)

		result := validator.Validate(meta, string(content))

		assert.False(t, result.Valid)

		var triggersError *skillpkg.ValidationIssue
		for _, issue := range result.Issues {
			if issue.Rule == "frontmatter" && strings.Contains(issue.Message, "triggers") {
				triggersError = &issue
				break
			}
		}

		require.NotNil(t, triggersError, "should have frontmatter error for missing triggers")
	})

	t.Run("complete invalid skill with all major issues", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")

		completelyInvalidSkill := `---
name: Bad@Name
description: Too short
triggers:
  - test
---

## Role

## Instructions

## Examples
`

		require.NoError(t, os.WriteFile(skillPath, []byte(completelyInvalidSkill), 0o644))

		err := ValidateGeneratedSkill(skillPath)

		assert.Error(t, err)

		errorMsg := err.Error()

		assert.Contains(t, errorMsg, "validation failed")
		assert.Contains(t, errorMsg, "ERRORS:")
		assert.Contains(t, errorMsg, "Suggestion:")
		assert.Contains(t, errorMsg, "Example:")

		assert.Contains(t, errorMsg, "Bad@Name", "should mention invalid name")
	})
}
