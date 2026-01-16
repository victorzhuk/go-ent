package config

//nolint:gosec // test file with necessary file operations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/victorzhuk/go-ent/internal/domain"
)

func TestLoad(t *testing.T) {
	t.Run("returns default config when file missing", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		cfg, err := Load(tmpDir)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "1.0", cfg.Version)
		assert.Equal(t, domain.AgentRoleSenior, cfg.Agents.Default)
		assert.Equal(t, domain.RuntimeClaudeCode, cfg.Runtime.Preferred)
	})

	t.Run("loads valid config from YAML", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		cfgDir := filepath.Join(tmpDir, ".go-ent")
		require.NoError(t, os.MkdirAll(cfgDir, 0750))

		yamlContent := `version: "1.0"

agents:
  default: architect
  roles:
    architect:
      model: opus
      skills: [go-arch, go-api]
      budget_limit: 5.0

runtime:
  preferred: claude-code
  fallback: [cli]

budget:
  daily: 15.0
  monthly: 300.0
  per_task: 2.0

models:
  opus: claude-opus-4-5-20251101
  sonnet: claude-sonnet-4-5-20251101
`
		cfgPath := filepath.Join(cfgDir, "config.yaml")
		require.NoError(t, os.WriteFile(cfgPath, []byte(yamlContent), 0600))

		cfg, err := Load(tmpDir)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "1.0", cfg.Version)
		assert.Equal(t, domain.AgentRoleArchitect, cfg.Agents.Default)
		assert.Equal(t, domain.RuntimeClaudeCode, cfg.Runtime.Preferred)
		assert.Equal(t, 15.0, cfg.Budget.Daily)
		assert.Equal(t, 300.0, cfg.Budget.Monthly)
		assert.Equal(t, 2.0, cfg.Budget.PerTask)
		assert.Equal(t, "claude-opus-4-5-20251101", cfg.Models["opus"])
	})

	t.Run("returns error on invalid YAML", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		cfgDir := filepath.Join(tmpDir, ".go-ent")
		require.NoError(t, os.MkdirAll(cfgDir, 0750))

		invalidYAML := `version: "1.0"
agents:
  default: architect
  roles:
    - this is invalid
`
		cfgPath := filepath.Join(cfgDir, "config.yaml")
		require.NoError(t, os.WriteFile(cfgPath, []byte(invalidYAML), 0600))

		cfg, err := Load(tmpDir)
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.ErrorIs(t, err, ErrInvalidYAML)
	})

	t.Run("returns error on validation failure", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		cfgDir := filepath.Join(tmpDir, ".go-ent")
		require.NoError(t, os.MkdirAll(cfgDir, 0750))

		invalidConfig := `version: "1.0"

agents:
  default: invalid-role
  roles:
    architect:
      model: opus

runtime:
  preferred: claude-code

budget:
  daily: -10.0

models:
  opus: claude-opus-4-5-20251101
`
		cfgPath := filepath.Join(cfgDir, "config.yaml")
		require.NoError(t, os.WriteFile(cfgPath, []byte(invalidConfig), 0600))

		cfg, err := Load(tmpDir)
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func TestLoadWithEnv(t *testing.T) {
	t.Run("overrides budget daily from env", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		getenv := func(key string) string {
			if key == "GOENT_BUDGET_DAILY" {
				return "25.5"
			}
			return ""
		}

		cfg, err := LoadWithEnv(tmpDir, getenv)
		require.NoError(t, err)
		assert.Equal(t, 25.5, cfg.Budget.Daily)
	})

	t.Run("overrides budget monthly from env", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		getenv := func(key string) string {
			if key == "GOENT_BUDGET_MONTHLY" {
				return "500.0"
			}
			return ""
		}

		cfg, err := LoadWithEnv(tmpDir, getenv)
		require.NoError(t, err)
		assert.Equal(t, 500.0, cfg.Budget.Monthly)
	})

	t.Run("overrides budget per_task from env", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		getenv := func(key string) string {
			if key == "GOENT_BUDGET_PER_TASK" {
				return "3.5"
			}
			return ""
		}

		cfg, err := LoadWithEnv(tmpDir, getenv)
		require.NoError(t, err)
		assert.Equal(t, 3.5, cfg.Budget.PerTask)
	})

	t.Run("overrides runtime preferred from env", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		getenv := func(key string) string {
			if key == "GOENT_RUNTIME_PREFERRED" {
				return "cli"
			}
			return ""
		}

		cfg, err := LoadWithEnv(tmpDir, getenv)
		require.NoError(t, err)
		assert.Equal(t, domain.RuntimeCLI, cfg.Runtime.Preferred)
	})

	t.Run("overrides agents default from env", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		getenv := func(key string) string {
			if key == "GOENT_AGENTS_DEFAULT" {
				return "architect"
			}
			return ""
		}

		cfg, err := LoadWithEnv(tmpDir, getenv)
		require.NoError(t, err)
		assert.Equal(t, domain.AgentRoleArchitect, cfg.Agents.Default)
	})

	t.Run("returns error on invalid budget daily", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		getenv := func(key string) string {
			if key == "GOENT_BUDGET_DAILY" {
				return "invalid"
			}
			return ""
		}

		cfg, err := LoadWithEnv(tmpDir, getenv)
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("returns error on invalid runtime", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		getenv := func(key string) string {
			if key == "GOENT_RUNTIME_PREFERRED" {
				return "invalid-runtime"
			}
			return ""
		}

		cfg, err := LoadWithEnv(tmpDir, getenv)
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("returns error on invalid agent role", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		getenv := func(key string) string {
			if key == "GOENT_AGENTS_DEFAULT" {
				return "invalid-role"
			}
			return ""
		}

		cfg, err := LoadWithEnv(tmpDir, getenv)
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("applies multiple env overrides", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		getenv := func(key string) string {
			switch key {
			case "GOENT_BUDGET_DAILY":
				return "50.0"
			case "GOENT_BUDGET_MONTHLY":
				return "1000.0"
			case "GOENT_RUNTIME_PREFERRED":
				return "cli"
			case "GOENT_AGENTS_DEFAULT":
				return "architect"
			default:
				return ""
			}
		}

		cfg, err := LoadWithEnv(tmpDir, getenv)
		require.NoError(t, err)
		assert.Equal(t, 50.0, cfg.Budget.Daily)
		assert.Equal(t, 1000.0, cfg.Budget.Monthly)
		assert.Equal(t, domain.RuntimeCLI, cfg.Runtime.Preferred)
		assert.Equal(t, domain.AgentRoleArchitect, cfg.Agents.Default)
	})
}
