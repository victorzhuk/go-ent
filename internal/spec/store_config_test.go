package spec

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/domain"
)

func TestStore_ConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	expected := filepath.Join(tmpDir, ".go-ent", "config.yaml")
	assert.Equal(t, expected, store.ConfigPath())
}

func TestStore_AgentsPath(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	expected := filepath.Join(tmpDir, "plugins", "go-ent", "agents")
	assert.Equal(t, expected, store.AgentsPath())
}

func TestStore_SkillsPath(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	expected := filepath.Join(tmpDir, "plugins", "go-ent", "skills")
	assert.Equal(t, expected, store.SkillsPath())
}

func TestStore_LoadConfig(t *testing.T) {
	t.Run("returns default config when file missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewStore(tmpDir)

		cfg, err := store.LoadConfig()
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "1.0", cfg.Version)
		assert.Equal(t, domain.AgentRoleSenior, cfg.Agents.Default)
	})

	t.Run("loads existing config", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewStore(tmpDir)

		cfgDir := filepath.Join(tmpDir, ".go-ent")
		require.NoError(t, os.MkdirAll(cfgDir, 0755))

		yamlContent := `version: "1.0"

agents:
  default: architect
  roles:
    architect:
      model: opus
      skills: [go-arch]
      budget_limit: 5.0

runtime:
  preferred: claude-code

budget:
  daily: 20.0
  monthly: 400.0
  per_task: 3.0

models:
  opus: claude-opus-4-5-20251101
`
		cfgPath := filepath.Join(cfgDir, "config.yaml")
		require.NoError(t, os.WriteFile(cfgPath, []byte(yamlContent), 0644))

		cfg, err := store.LoadConfig()
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, domain.AgentRoleArchitect, cfg.Agents.Default)
		assert.Equal(t, 20.0, cfg.Budget.Daily)
		assert.Equal(t, 400.0, cfg.Budget.Monthly)
	})
}

func TestStore_SaveConfig(t *testing.T) {
	t.Run("creates config directory and file", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewStore(tmpDir)

		cfg := config.DefaultConfig()
		cfg.Budget.Daily = 25.0

		err := store.SaveConfig(cfg)
		require.NoError(t, err)

		cfgPath := store.ConfigPath()
		assert.FileExists(t, cfgPath)

		data, err := os.ReadFile(cfgPath)
		require.NoError(t, err)
		assert.Contains(t, string(data), "daily: 25")
	})

	t.Run("overwrites existing config", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewStore(tmpDir)

		cfg1 := config.DefaultConfig()
		cfg1.Budget.Daily = 10.0
		require.NoError(t, store.SaveConfig(cfg1))

		cfg2 := config.DefaultConfig()
		cfg2.Budget.Daily = 30.0
		require.NoError(t, store.SaveConfig(cfg2))

		loaded, err := store.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, 30.0, loaded.Budget.Daily)
	})

	t.Run("round-trip save and load", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewStore(tmpDir)

		original := config.DefaultConfig()
		original.Agents.Default = domain.AgentRoleArchitect
		original.Budget.Daily = 15.0
		original.Budget.Monthly = 300.0
		original.Runtime.Preferred = domain.RuntimeCLI

		require.NoError(t, store.SaveConfig(original))

		loaded, err := store.LoadConfig()
		require.NoError(t, err)

		assert.Equal(t, original.Version, loaded.Version)
		assert.Equal(t, original.Agents.Default, loaded.Agents.Default)
		assert.Equal(t, original.Budget.Daily, loaded.Budget.Daily)
		assert.Equal(t, original.Budget.Monthly, loaded.Budget.Monthly)
		assert.Equal(t, original.Runtime.Preferred, loaded.Runtime.Preferred)
	})
}
