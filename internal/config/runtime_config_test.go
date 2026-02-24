package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadToolRuntimeConfig(t *testing.T) {
	t.Run("returns default when config missing", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		cfg, err := LoadToolRuntimeConfig(tmpDir, "claude")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "claude-sonnet-4-5-20250929", cfg.Claude.Sonnet)
		assert.Equal(t, "claude-opus-4-5-20251101", cfg.Claude.Opus)
		assert.Equal(t, "claude-haiku-4-5-20251001", cfg.Claude.Haiku)
	})

	t.Run("loads claude config from .claude/ent.yaml", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		cfgDir := filepath.Join(tmpDir, ".claude")
		require.NoError(t, os.MkdirAll(cfgDir, 0o750))

		yamlContent := `claude:
  sonnet: custom-sonnet-model
  opus: custom-opus-model
  haiku: custom-haiku-model
`
		cfgPath := filepath.Join(cfgDir, "ent.yaml")
		require.NoError(t, os.WriteFile(cfgPath, []byte(yamlContent), 0o600))

		cfg, err := LoadToolRuntimeConfig(tmpDir, "claude")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "custom-sonnet-model", cfg.Claude.Sonnet)
		assert.Equal(t, "custom-opus-model", cfg.Claude.Opus)
		assert.Equal(t, "custom-haiku-model", cfg.Claude.Haiku)
	})

	t.Run("loads opencode config from .opencode/ent.yaml", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		cfgDir := filepath.Join(tmpDir, ".opencode")
		require.NoError(t, os.MkdirAll(cfgDir, 0o750))

		yamlContent := `opencode:
  fast: custom-fast-model
  main: custom-main-model
  heavy: custom-heavy-model
`
		cfgPath := filepath.Join(cfgDir, "ent.yaml")
		require.NoError(t, os.WriteFile(cfgPath, []byte(yamlContent), 0o600))

		cfg, err := LoadToolRuntimeConfig(tmpDir, "opencode")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "custom-fast-model", cfg.OpenCode.Fast)
		assert.Equal(t, "custom-main-model", cfg.OpenCode.Main)
		assert.Equal(t, "custom-heavy-model", cfg.OpenCode.Heavy)
	})

	t.Run("returns error for unknown runtime", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		cfg, err := LoadToolRuntimeConfig(tmpDir, "unknown")
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func TestClaudeModelsResolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   ClaudeModels
		alias    string
		expected string
	}{
		{
			name:     "resolve fast alias",
			config:   ClaudeModels{Haiku: "my-haiku"},
			alias:    "fast",
			expected: "my-haiku",
		},
		{
			name:     "resolve haiku alias",
			config:   ClaudeModels{Haiku: "my-haiku"},
			alias:    "haiku",
			expected: "my-haiku",
		},
		{
			name:     "resolve main alias",
			config:   ClaudeModels{Sonnet: "my-sonnet"},
			alias:    "main",
			expected: "my-sonnet",
		},
		{
			name:     "resolve sonnet alias",
			config:   ClaudeModels{Sonnet: "my-sonnet"},
			alias:    "sonnet",
			expected: "my-sonnet",
		},
		{
			name:     "resolve heavy alias",
			config:   ClaudeModels{Opus: "my-opus"},
			alias:    "heavy",
			expected: "my-opus",
		},
		{
			name:     "resolve opus alias",
			config:   ClaudeModels{Opus: "my-opus"},
			alias:    "opus",
			expected: "my-opus",
		},
		{
			name:     "return unknown alias as-is",
			config:   ClaudeModels{},
			alias:    "custom-model-id",
			expected: "custom-model-id",
		},
		{
			name:     "use default when haiku empty",
			config:   ClaudeModels{},
			alias:    "haiku",
			expected: "claude-haiku-4-5-20251001",
		},
		{
			name:     "use default when sonnet empty",
			config:   ClaudeModels{},
			alias:    "sonnet",
			expected: "claude-sonnet-4-5-20250929",
		},
		{
			name:     "use default when opus empty",
			config:   ClaudeModels{},
			alias:    "opus",
			expected: "claude-opus-4-5-20251101",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.config.Resolve(tt.alias)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOpenCodeModelsResolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   OpenCodeModels
		alias    string
		expected string
	}{
		{
			name:     "resolve fast alias",
			config:   OpenCodeModels{Fast: "my-fast"},
			alias:    "fast",
			expected: "my-fast",
		},
		{
			name:     "resolve haiku alias",
			config:   OpenCodeModels{Fast: "my-fast"},
			alias:    "haiku",
			expected: "my-fast",
		},
		{
			name:     "resolve main alias",
			config:   OpenCodeModels{Main: "my-main"},
			alias:    "main",
			expected: "my-main",
		},
		{
			name:     "resolve sonnet alias",
			config:   OpenCodeModels{Main: "my-main"},
			alias:    "sonnet",
			expected: "my-main",
		},
		{
			name:     "resolve heavy alias",
			config:   OpenCodeModels{Heavy: "my-heavy"},
			alias:    "heavy",
			expected: "my-heavy",
		},
		{
			name:     "resolve opus alias",
			config:   OpenCodeModels{Heavy: "my-heavy"},
			alias:    "opus",
			expected: "my-heavy",
		},
		{
			name:     "return unknown alias as-is",
			config:   OpenCodeModels{},
			alias:    "custom-model-id",
			expected: "custom-model-id",
		},
		{
			name:     "use default when fast empty",
			config:   OpenCodeModels{},
			alias:    "fast",
			expected: "zai-coding-plan/glm-4.7-flash",
		},
		{
			name:     "use default when main empty",
			config:   OpenCodeModels{},
			alias:    "main",
			expected: "zai-coding-plan/glm-5",
		},
		{
			name:     "use default when heavy empty",
			config:   OpenCodeModels{},
			alias:    "heavy",
			expected: "kimi-for-coding/k2p5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.config.Resolve(tt.alias)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultToolRuntimeConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultToolRuntimeConfig()
	require.NotNil(t, cfg)

	assert.NotEmpty(t, cfg.Claude.Sonnet)
	assert.NotEmpty(t, cfg.Claude.Opus)
	assert.NotEmpty(t, cfg.Claude.Haiku)
	assert.NotEmpty(t, cfg.OpenCode.Fast)
	assert.NotEmpty(t, cfg.OpenCode.Main)
	assert.NotEmpty(t, cfg.OpenCode.Heavy)
}
