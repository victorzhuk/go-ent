package skill

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunNonInteractive(t *testing.T) {
	t.Run("requires template flag", func(t *testing.T) {
		_, err := runNonInteractive("test", "", "desc", "cat", "author", "tags")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--template flag is required")
	})

	t.Run("requires description flag", func(t *testing.T) {
		_, err := runNonInteractive("test", "tpl", "", "cat", "author", "tags")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--description flag is required")
	})

	t.Run("requires category if cannot detect", func(t *testing.T) {
		_, err := runNonInteractive("myskill", "tpl", "desc", "", "author", "tags")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--category flag required")
	})

	t.Run("detects category from name", func(t *testing.T) {
		cfg, err := runNonInteractive("go-payment", "tpl", "desc", "", "author", "tags")
		require.NoError(t, err)
		assert.Equal(t, "go", cfg.Category)
	})

	t.Run("detects TypeScript category", func(t *testing.T) {
		cfg, err := runNonInteractive("typescript-api", "tpl", "desc", "", "author", "tags")
		require.NoError(t, err)
		assert.Equal(t, "typescript", cfg.Category)
	})

	t.Run("detects database category", func(t *testing.T) {
		cfg, err := runNonInteractive("db-connection", "tpl", "desc", "", "author", "tags")
		require.NoError(t, err)
		assert.Equal(t, "database", cfg.Category)
	})

	t.Run("uses provided category", func(t *testing.T) {
		cfg, err := runNonInteractive("myskill", "tpl", "desc", "custom", "author", "tags")
		require.NoError(t, err)
		assert.Equal(t, "custom", cfg.Category)
	})

	t.Run("creates correct output path", func(t *testing.T) {
		cfg, err := runNonInteractive("go-payment", "tpl", "Test skill", "go", "author", "tags")
		require.NoError(t, err)
		expectedPath := "plugins/go-ent/skills/go/go-payment/SKILL.md"
		assert.Equal(t, expectedPath, cfg.OutputPath)
	})

	t.Run("returns valid config", func(t *testing.T) {
		cfg, err := runNonInteractive("go-test", "test-template", "Test description", "go", "author", "tags")
		require.NoError(t, err)
		assert.Equal(t, "go-test", cfg.Name)
		assert.Equal(t, "test-template", cfg.TemplateName)
		assert.Equal(t, "Test description", cfg.Description)
		assert.Equal(t, "go", cfg.Category)
		assert.Equal(t, "plugins/go-ent/skills/go/go-test/SKILL.md", cfg.OutputPath)
	})

	t.Run("handles uppercase category detection", func(t *testing.T) {
		cfg, err := runNonInteractive("GO-PAYMENT", "tpl", "desc", "", "author", "tags")
		require.NoError(t, err)
		assert.Equal(t, "go", cfg.Category)
	})
}
