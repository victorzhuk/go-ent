package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRuntimeConfig(t *testing.T) {
	t.Run("errors when config missing", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		cfg, err := LoadRuntimeConfig(tmpDir, "claude")
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "ent config init")
		assert.True(t, errors.Is(err, os.ErrNotExist))
	})

	t.Run("loads config from .claude/ent.yaml", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		cfgDir := filepath.Join(tmpDir, ".claude")
		require.NoError(t, os.MkdirAll(cfgDir, 0o750))

		yamlContent := `models:
  fast: custom-haiku
  main: custom-sonnet
  heavy: custom-opus
`
		cfgPath := filepath.Join(cfgDir, "ent.yaml")
		require.NoError(t, os.WriteFile(cfgPath, []byte(yamlContent), 0o600))

		cfg, err := LoadRuntimeConfig(tmpDir, "claude")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "custom-haiku", cfg.Models.Fast)
		assert.Equal(t, "custom-sonnet", cfg.Models.Main)
		assert.Equal(t, "custom-opus", cfg.Models.Heavy)
	})

	t.Run("loads config from .opencode/ent.yaml", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		cfgDir := filepath.Join(tmpDir, ".opencode")
		require.NoError(t, os.MkdirAll(cfgDir, 0o750))

		yamlContent := `models:
  fast: custom-fast-model
  main: custom-main-model
  heavy: custom-heavy-model
`
		cfgPath := filepath.Join(cfgDir, "ent.yaml")
		require.NoError(t, os.WriteFile(cfgPath, []byte(yamlContent), 0o600))

		cfg, err := LoadRuntimeConfig(tmpDir, "opencode")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "custom-fast-model", cfg.Models.Fast)
		assert.Equal(t, "custom-main-model", cfg.Models.Main)
		assert.Equal(t, "custom-heavy-model", cfg.Models.Heavy)
	})

	t.Run("returns error for unknown runtime", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		cfg, err := LoadRuntimeConfig(tmpDir, "unknown")
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func TestModelTiersResolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tiers    ModelTiers
		tier     string
		expected string
	}{
		{
			name:     "resolve fast tier",
			tiers:    ModelTiers{Fast: "my-fast-model"},
			tier:     "fast",
			expected: "my-fast-model",
		},
		{
			name:     "resolve main tier",
			tiers:    ModelTiers{Main: "my-main-model"},
			tier:     "main",
			expected: "my-main-model",
		},
		{
			name:     "resolve heavy tier",
			tiers:    ModelTiers{Heavy: "my-heavy-model"},
			tier:     "heavy",
			expected: "my-heavy-model",
		},
		{
			name:     "unknown tier passes through",
			tiers:    ModelTiers{},
			tier:     "custom-model-id",
			expected: "custom-model-id",
		},
		{
			name:     "empty fast returns empty string",
			tiers:    ModelTiers{},
			tier:     "fast",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.tiers.Resolve(tt.tier)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestModelTiersResolveForAgent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tiers       ModelTiers
		agentName   string
		defaultTier string
		expected    string
	}{
		{
			name: "agent override present",
			tiers: ModelTiers{
				Fast:   "fast-val",
				Main:   "main-val",
				Heavy:  "heavy-val",
				Agents: map[string]string{"coder": "heavy"},
			},
			agentName:   "coder",
			defaultTier: "main",
			expected:    "heavy-val",
		},
		{
			name: "agent override absent falls back to default tier",
			tiers: ModelTiers{
				Fast:   "fast-val",
				Main:   "main-val",
				Heavy:  "heavy-val",
				Agents: map[string]string{"scout": "fast"},
			},
			agentName:   "coder",
			defaultTier: "main",
			expected:    "main-val",
		},
		{
			name: "nil agents falls back to default tier",
			tiers: ModelTiers{
				Fast:  "fast-val",
				Main:  "main-val",
				Heavy: "heavy-val",
			},
			agentName:   "coder",
			defaultTier: "heavy",
			expected:    "heavy-val",
		},
		{
			name:        "empty agents map falls back to default tier",
			tiers:       ModelTiers{Main: "main-val", Agents: map[string]string{}},
			agentName:   "coder",
			defaultTier: "main",
			expected:    "main-val",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.tiers.ResolveForAgent(tt.agentName, tt.defaultTier)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		value   string
		verify  func(t *testing.T, cfg *RuntimeConfig)
		wantErr bool
	}{
		{
			name:  "models.fast",
			key:   "models.fast",
			value: "my-fast",
			verify: func(t *testing.T, cfg *RuntimeConfig) {
				t.Helper()
				assert.Equal(t, "my-fast", cfg.Models.Fast)
			},
		},
		{
			name:  "models.main",
			key:   "models.main",
			value: "my-main",
			verify: func(t *testing.T, cfg *RuntimeConfig) {
				t.Helper()
				assert.Equal(t, "my-main", cfg.Models.Main)
			},
		},
		{
			name:  "models.heavy",
			key:   "models.heavy",
			value: "my-heavy",
			verify: func(t *testing.T, cfg *RuntimeConfig) {
				t.Helper()
				assert.Equal(t, "my-heavy", cfg.Models.Heavy)
			},
		},
		{
			name:  "models.agents.coder",
			key:   "models.agents.coder",
			value: "heavy",
			verify: func(t *testing.T, cfg *RuntimeConfig) {
				t.Helper()
				require.NotNil(t, cfg.Models.Agents)
				assert.Equal(t, "heavy", cfg.Models.Agents["coder"])
			},
		},
		{
			name:    "unknown key errors",
			key:     "budget.daily",
			value:   "50",
			wantErr: true,
		},
		{
			name:    "old claude.fast key errors",
			key:     "claude.fast",
			value:   "some-model",
			wantErr: true,
		},
		{
			name:    "models.agents.<name> with invalid tier errors",
			key:     "models.agents.coder",
			value:   "typo-tier",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &RuntimeConfig{}
			err := ApplyKey(cfg, tt.key, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.verify(t, cfg)
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *RuntimeConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with all tiers",
			cfg: &RuntimeConfig{
				Models: ModelTiers{Fast: "haiku", Main: "sonnet", Heavy: "opus"},
			},
			wantErr: false,
		},
		{
			name: "missing fast",
			cfg: &RuntimeConfig{
				Models: ModelTiers{Main: "sonnet", Heavy: "opus"},
			},
			wantErr: true,
			errMsg:  "models.fast",
		},
		{
			name:    "all missing",
			cfg:     &RuntimeConfig{},
			wantErr: true,
			errMsg:  "models.fast",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}
