package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadToolRuntimeConfig(t *testing.T) {
	t.Run("errors when config missing", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		cfg, err := LoadToolRuntimeConfig(tmpDir, "claude")
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "ent config init")
		assert.True(t, errors.Is(err, os.ErrNotExist))
	})

	t.Run("loads claude config from .claude/ent.yaml", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		cfgDir := filepath.Join(tmpDir, ".claude")
		require.NoError(t, os.MkdirAll(cfgDir, 0o750))

		yamlContent := `claude:
  fast: custom-haiku
  main: custom-sonnet
  heavy: custom-opus
`
		cfgPath := filepath.Join(cfgDir, "ent.yaml")
		require.NoError(t, os.WriteFile(cfgPath, []byte(yamlContent), 0o600))

		cfg, err := LoadToolRuntimeConfig(tmpDir, "claude")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "custom-haiku", cfg.Claude.Fast)
		assert.Equal(t, "custom-sonnet", cfg.Claude.Main)
		assert.Equal(t, "custom-opus", cfg.Claude.Heavy)
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
		verify  func(t *testing.T, cfg *ToolRuntimeConfig)
		wantErr bool
	}{
		{
			name:  "claude.fast",
			key:   "claude.fast",
			value: "my-fast",
			verify: func(t *testing.T, cfg *ToolRuntimeConfig) {
				t.Helper()
				assert.Equal(t, "my-fast", cfg.Claude.Fast)
			},
		},
		{
			name:  "claude.main",
			key:   "claude.main",
			value: "my-main",
			verify: func(t *testing.T, cfg *ToolRuntimeConfig) {
				t.Helper()
				assert.Equal(t, "my-main", cfg.Claude.Main)
			},
		},
		{
			name:  "claude.heavy",
			key:   "claude.heavy",
			value: "my-heavy",
			verify: func(t *testing.T, cfg *ToolRuntimeConfig) {
				t.Helper()
				assert.Equal(t, "my-heavy", cfg.Claude.Heavy)
			},
		},
		{
			name:  "claude.agents.coder",
			key:   "claude.agents.coder",
			value: "heavy",
			verify: func(t *testing.T, cfg *ToolRuntimeConfig) {
				t.Helper()
				require.NotNil(t, cfg.Claude.Agents)
				assert.Equal(t, "heavy", cfg.Claude.Agents["coder"])
			},
		},
		{
			name:  "opencode.fast",
			key:   "opencode.fast",
			value: "fast-model",
			verify: func(t *testing.T, cfg *ToolRuntimeConfig) {
				t.Helper()
				assert.Equal(t, "fast-model", cfg.OpenCode.Fast)
			},
		},
		{
			name:  "opencode.agents.scout",
			key:   "opencode.agents.scout",
			value: "main",
			verify: func(t *testing.T, cfg *ToolRuntimeConfig) {
				t.Helper()
				require.NotNil(t, cfg.OpenCode.Agents)
				assert.Equal(t, "main", cfg.OpenCode.Agents["scout"])
			},
		},
		{
			name:    "unknown key errors",
			key:     "budget.daily",
			value:   "50",
			wantErr: true,
		},
		{
			name:    "claude.sonnet errors (old key)",
			key:     "claude.sonnet",
			value:   "some-model",
			wantErr: true,
		},
		{
			name:    "claude.agents.<name> with invalid tier errors",
			key:     "claude.agents.coder",
			value:   "typo-tier",
			wantErr: true,
		},
		{
			name:    "opencode.agents.<name> with invalid tier errors",
			key:     "opencode.agents.scout",
			value:   "not-a-tier",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &ToolRuntimeConfig{}
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

func TestValidateForRuntime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *ToolRuntimeConfig
		runtime string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid claude config",
			cfg: &ToolRuntimeConfig{
				Claude: ModelTiers{Fast: "haiku", Main: "sonnet", Heavy: "opus"},
			},
			runtime: "claude",
			wantErr: false,
		},
		{
			name: "claude missing fast",
			cfg: &ToolRuntimeConfig{
				Claude: ModelTiers{Main: "sonnet", Heavy: "opus"},
			},
			runtime: "claude",
			wantErr: true,
			errMsg:  "claude.fast",
		},
		{
			name: "claude all missing",
			cfg: &ToolRuntimeConfig{
				Claude: ModelTiers{},
			},
			runtime: "claude",
			wantErr: true,
			errMsg:  "claude.fast",
		},
		{
			name: "valid opencode config",
			cfg: &ToolRuntimeConfig{
				OpenCode: ModelTiers{Fast: "fast-m", Main: "main-m", Heavy: "heavy-m"},
			},
			runtime: "opencode",
			wantErr: false,
		},
		{
			name:    "unknown runtime errors",
			cfg:     &ToolRuntimeConfig{},
			runtime: "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateForRuntime(tt.cfg, tt.runtime)
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
