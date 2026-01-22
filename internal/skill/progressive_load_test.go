package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressiveLoadingLevels(t *testing.T) {
	tests := []struct {
		name        string
		loadLevel   LoadLevel
		hasMetadata bool
		hasCore     bool
		hasFull     bool
	}{
		{"Level 1 - Metadata only", LoadMetadata, true, false, false},
		{"Level 2 - Core content", LoadCore, true, true, false},
		{"Level 3 - Full content", LoadExtended, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parser := NewParser()

			meta, err := parser.ParseSkillFile("testdata/test_skill.md")
			require.NoError(t, err)

			assert.Equal(t, LoadMetadata, meta.LoadLevel)
			assert.NotNil(t, meta.Name)
			assert.True(t, tt.hasMetadata)

			if tt.loadLevel > LoadMetadata {
				err = parser.UpgradeToLevel(meta, tt.loadLevel)
				require.NoError(t, err)
				assert.Equal(t, tt.loadLevel, meta.LoadLevel)
			}

			if tt.hasCore {
				assert.NotNil(t, meta.Core)
				assert.NotEmpty(t, meta.Core.Role)
				assert.NotEmpty(t, meta.Core.Instructions)
			} else {
				assert.Nil(t, meta.Core)
			}

			if tt.hasFull {
				assert.NotNil(t, meta.Full)
				assert.NotEmpty(t, meta.Full.Body)
			} else {
				assert.Nil(t, meta.Full)
			}
		})
	}
}

func TestTokenUsagePerLevel(t *testing.T) {
	t.Parallel()
	parser := NewParser()

	meta, err := parser.ParseSkillFile("testdata/test_skill.md")
	require.NoError(t, err)

	err = parser.UpgradeToLevel(meta, LoadCore)
	require.NoError(t, err)

	level1Tokens := countTokens(fmt.Sprintf("%s %s %v",
		meta.Name, meta.Description, meta.Triggers))
	assert.Less(t, level1Tokens, 200, "Level 1 should be <200 tokens")

	level2Tokens := level1Tokens + countTokens(
		meta.Core.Role+meta.Core.Instructions+
			meta.Core.Constraints+meta.Core.Examples)
	assert.Less(t, level2Tokens, 5000, "Level 2 should be <5k tokens")

	err = parser.UpgradeToLevel(meta, LoadExtended)
	require.NoError(t, err)

	level3Tokens := level2Tokens + countTokens(meta.Full.Body)

	t.Logf("Token usage: Level1=%d, Level2=%d, Level3=%d",
		level1Tokens, level2Tokens, level3Tokens)
}

func TestLazyLoading(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	err := os.MkdirAll(filepath.Join(tmpDir, "skills"), 0o750) // #nosec G301
	require.NoError(t, err)

	testSkillContent := `---
name: test-skill
description: "Test skill for lazy loading"
triggers:
  - keywords:
      - test
      - lazy
---

<role>
Test role for lazy loading verification
</role>

<instructions>
Basic test instructions
</instructions>
`

	skillPath := filepath.Join(tmpDir, "skills", "SKILL.md")
	err = os.WriteFile(skillPath, []byte(testSkillContent), 0o600) // #nosec G306
	require.NoError(t, err)

	registry := NewRegistry()

	err = registry.Load(filepath.Join(tmpDir, "skills"))
	require.NoError(t, err)

	for _, skill := range registry.All() {
		assert.Equal(t, LoadMetadata, skill.LoadLevel)
		assert.Nil(t, skill.Core)
		assert.Nil(t, skill.Full)
	}

	err = registry.UpgradeSkill("test-skill", LoadCore)
	require.NoError(t, err)

	for _, skill := range registry.All() {
		if skill.Name == "test-skill" {
			assert.Equal(t, LoadCore, skill.LoadLevel)
			assert.NotNil(t, skill.Core)
		} else {
			assert.Equal(t, LoadMetadata, skill.LoadLevel)
			assert.Nil(t, skill.Core)
		}
	}
}
