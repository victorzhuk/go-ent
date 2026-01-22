package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgradeToLevel_LoadExtended(t *testing.T) {
	t.Run("loads full body content", func(t *testing.T) {
		skillContent := `---
name: test-skill
description: A test skill for Level 3 loading
version: "1.0.0"
author: Test Author
tags:
  - test
triggers:
  - keywords: ["test"]
    weight: 0.8
---
<role>
You are a test assistant.
</role>

<instructions>
Follow these instructions.
</instructions>

<constraints>
No constraints.
</constraints>

<examples>
<example>
<input>test</input>
<output>result</output>
</example>
</examples>
`

		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")
		require.NoError(t, os.WriteFile(skillPath, []byte(skillContent), 0o644))

		parser := NewParser()
		meta, err := parser.ParseSkillFile(skillPath)
		require.NoError(t, err)

		assert.Equal(t, LoadMetadata, meta.LoadLevel)
		assert.Nil(t, meta.Core)
		assert.Nil(t, meta.Full)

		err = parser.UpgradeToLevel(meta, LoadCore)
		require.NoError(t, err)

		assert.Equal(t, LoadCore, meta.LoadLevel)
		assert.NotNil(t, meta.Core)
		assert.Equal(t, "You are a test assistant.", meta.Core.Role)
		assert.Nil(t, meta.Full)

		err = parser.UpgradeToLevel(meta, LoadExtended)
		require.NoError(t, err)

		assert.Equal(t, LoadExtended, meta.LoadLevel)
		assert.NotNil(t, meta.Full)
		assert.NotEmpty(t, meta.Full.Body)
		assert.Contains(t, meta.Full.Body, "<role>")
		assert.Contains(t, meta.Full.Body, "You are a test assistant.")
		assert.Nil(t, meta.Full.References)
		assert.Nil(t, meta.Full.Scripts)
	})

	t.Run("ensures Core is loaded before Extended", func(t *testing.T) {
		skillContent := `---
name: test-skill
description: Test
---
<role>Role</role>
<instructions>Instructions</instructions>
<constraints>Constraints</constraints>
<examples></examples>
`

		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")
		require.NoError(t, os.WriteFile(skillPath, []byte(skillContent), 0o644))

		parser := NewParser()
		meta, err := parser.ParseSkillFile(skillPath)
		require.NoError(t, err)

		assert.Equal(t, LoadMetadata, meta.LoadLevel)
		assert.Nil(t, meta.Core)

		err = parser.UpgradeToLevel(meta, LoadExtended)
		require.NoError(t, err)

		assert.Equal(t, LoadExtended, meta.LoadLevel)
		assert.NotNil(t, meta.Core, "Core should be loaded when upgrading to Extended")
		assert.NotNil(t, meta.Full)
	})

	t.Run("handles multiple upgrades gracefully", func(t *testing.T) {
		skillContent := `---
name: test-skill
description: Test
---
<role>Role</role>
<instructions>Instructions</instructions>
<constraints>Constraints</constraints>
<examples></examples>
`

		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")
		require.NoError(t, os.WriteFile(skillPath, []byte(skillContent), 0o644))

		parser := NewParser()
		meta, err := parser.ParseSkillFile(skillPath)
		require.NoError(t, err)

		err = parser.UpgradeToLevel(meta, LoadCore)
		require.NoError(t, err)
		assert.Equal(t, LoadCore, meta.LoadLevel)

		err = parser.UpgradeToLevel(meta, LoadCore)
		require.NoError(t, err, "Re-upgrading to same level should not error")
		assert.Equal(t, LoadCore, meta.LoadLevel)

		err = parser.UpgradeToLevel(meta, LoadExtended)
		require.NoError(t, err)
		assert.Equal(t, LoadExtended, meta.LoadLevel)

		err = parser.UpgradeToLevel(meta, LoadExtended)
		require.NoError(t, err, "Re-upgrading to same level should not error")
		assert.Equal(t, LoadExtended, meta.LoadLevel)
	})
}
