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
version: "not.semver"
---

# Test Skill

<role>
</role>

<instructions>
</instructions>

<constraints>
- Test
</constraints>

<examples>
<example>
<input>test</input>
<output>test</output>
</example>
</examples>

<output_format>
Test output
</output_format>
`

		require.NoError(t, os.WriteFile(skillPath, []byte(invalidSkill), 0644))

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
description: A test skill
version: "1.0.0"
author: Test
tags: ["test"]
structure_version: "v2"
---

<role>Test role</role>
<instructions>Test instructions</instructions>
<constraints>- Test</constraints>
<examples>
<example>
<input>test</input>
<output>test</output>
</example>
</examples>
<output_format>Test</output_format>
`

		require.NoError(t, os.WriteFile(skillPath, []byte(invalidNameSkill), 0644))

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
			if issue.Rule == "SK002" {
				nameError = &issue
				break
			}
		}

		require.NotNil(t, nameError, "should have SK002 error for invalid name")
		assert.NotEmpty(t, nameError.Suggestion, "name error should have suggestion")
		assert.NotEmpty(t, nameError.Example, "name error should have example")
		assert.Contains(t, nameError.Message, "invalid name format")
		assert.Contains(t, strings.ToLower(nameError.Suggestion), "lowercase")
	})

	t.Run("missing description shows helpful error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")

		missingDescSkill := `---
name: test-skill
version: "1.0.0"
author: Test
tags: ["test"]
structure_version: "v2"
---

<role>Test role</role>
<instructions>Test instructions</instructions>
<constraints>- Test</constraints>
<examples>
<example>
<input>test</input>
<output>test</output>
</example>
</examples>
<output_format>Test</output_format>
`

		require.NoError(t, os.WriteFile(skillPath, []byte(missingDescSkill), 0644))

		parser := skillpkg.NewParser()
		validator := skillpkg.NewValidator()

		meta, err := parser.ParseSkillFile(skillPath)
		require.NoError(t, err)

		content, err := os.ReadFile(skillPath)
		require.NoError(t, err)

		result := validator.Validate(meta, string(content))

		assert.False(t, result.Valid)

		var descError *skillpkg.ValidationIssue
		for _, issue := range result.Issues {
			if issue.Rule == "SK003" && strings.Contains(issue.Message, "description") {
				descError = &issue
				break
			}
		}

		require.NotNil(t, descError, "should have SK003 error for missing description")
		assert.NotEmpty(t, descError.Suggestion, "description error should have suggestion")
		assert.NotEmpty(t, descError.Example, "description error should have example")
	})

	t.Run("empty role section shows helpful error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")

		emptyRoleSkill := `---
name: test-skill
description: A test skill with empty role
version: "1.0.0"
author: Test
tags: ["test"]
structure_version: "v2"
---

<role>
</role>

<instructions>Test instructions</instructions>
<constraints>- Test</constraints>
<examples>
<example>
<input>test</input>
<output>test</output>
</example>
</examples>
<output_format>Test</output_format>
`

		require.NoError(t, os.WriteFile(skillPath, []byte(emptyRoleSkill), 0644))

		parser := skillpkg.NewParser()
		validator := skillpkg.NewValidator()

		meta, err := parser.ParseSkillFile(skillPath)
		require.NoError(t, err)

		content, err := os.ReadFile(skillPath)
		require.NoError(t, err)

		result := validator.Validate(meta, string(content))

		var roleError *skillpkg.ValidationIssue
		for _, issue := range result.Issues {
			if issue.Rule == "SK005" && strings.Contains(issue.Message, "empty") {
				roleError = &issue
				break
			}
		}

		require.NotNil(t, roleError, "should have SK005 error for empty role")
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
version: "1.0.0"
author: Test
tags: ["test"]
structure_version: "v2"
---

<role></role>
<instructions></instructions>
<constraints>- Test</constraints>
<examples>
<example>
<input>test</input>
<output>test</output>
</example>
</examples>
<output_format>Test</output_format>
`

		require.NoError(t, os.WriteFile(skillPath, []byte(invalidSkill), 0644))

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
version: "1.0.0"
author: Test
tags: ["test"]
structure_version: "v2"
---

<role></role>
<instructions></instructions>
<constraints>- Test</constraints>
<examples>
<example>
<input>test</input>
<output>test</output>
</example>
</examples>
<output_format>Test</output_format>
`

		require.NoError(t, os.WriteFile(skillPath, []byte(invalidSkill), 0644))

		parser := skillpkg.NewParser()
		validator := skillpkg.NewValidator()

		meta, err := parser.ParseSkillFile(skillPath)
		require.NoError(t, err)

		content, err := os.ReadFile(skillPath)
		require.NoError(t, err)

		result := validator.Validate(meta, string(content))

		assert.Greater(t, len(result.Issues), 0, "should have validation issues")

		for _, issue := range result.Issues {
			if issue.Rule == "frontmatter" || issue.Rule == "SK002" ||
				issue.Rule == "SK003" || issue.Rule == "SK005" {
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

		invalidVersionSkill := `---
name: test-skill
description: A test skill
version: "not-a-version"
author: Test
tags: ["test"]
structure_version: "v2"
---

<role>Test role</role>
<instructions>Test instructions</instructions>
<constraints>- Test</constraints>
<examples>
<example>
<input>test</input>
<output>test</output>
</example>
</examples>
<output_format>Test</output_format>
`

		require.NoError(t, os.WriteFile(skillPath, []byte(invalidVersionSkill), 0644))

		parser := skillpkg.NewParser()
		validator := skillpkg.NewValidator()

		meta, err := parser.ParseSkillFile(skillPath)
		require.NoError(t, err)

		content, err := os.ReadFile(skillPath)
		require.NoError(t, err)

		result := validator.Validate(meta, string(content))

		assert.False(t, result.Valid)

		var versionError *skillpkg.ValidationIssue
		for _, issue := range result.Issues {
			if issue.Rule == "version" {
				versionError = &issue
				break
			}
		}

		require.NotNil(t, versionError, "should have version error")
		assert.Contains(t, versionError.Message, "invalid semantic version")
		assert.Contains(t, versionError.Message, "v1.0.0")
	})

	t.Run("complete invalid skill with all major issues", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")

		completelyInvalidSkill := `---
name: Bad@Name
description: Too short
version: invalid
---

<role></role>
<instructions></instructions>
<constraints>- Test</constraints>
<examples>
<example>
<input>test</input>
<output>test</output>
</example>
</examples>
<output_format>Test</output_format>
`

		require.NoError(t, os.WriteFile(skillPath, []byte(completelyInvalidSkill), 0644))

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
