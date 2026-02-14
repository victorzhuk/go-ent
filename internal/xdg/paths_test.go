package xdg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDir(t *testing.T) {
	t.Run("returns path containing go-ent", func(t *testing.T) {
		dir := ConfigDir()
		assert.True(t, strings.HasSuffix(dir, "go-ent"))
	})

	t.Run("respects XDG_CONFIG_HOME", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/tmp/test-config")
		dir := ConfigDir()
		assert.Equal(t, "/tmp/test-config/go-ent", dir)
	})
}

func TestDataDir(t *testing.T) {
	t.Run("returns path containing go-ent", func(t *testing.T) {
		dir := DataDir()
		assert.True(t, strings.HasSuffix(dir, "go-ent"))
	})

	t.Run("respects XDG_DATA_HOME", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "/tmp/test-data")
		dir := DataDir()
		assert.Equal(t, "/tmp/test-data/go-ent", dir)
	})

	t.Run("defaults to ~/.local/share/go-ent", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "")
		dir := DataDir()
		home, _ := os.UserHomeDir()
		assert.Equal(t, filepath.Join(home, ".local", "share", "go-ent"), dir)
	})
}

func TestCacheDir(t *testing.T) {
	t.Run("returns path containing go-ent", func(t *testing.T) {
		dir := CacheDir()
		assert.True(t, strings.HasSuffix(dir, "go-ent"))
	})

	t.Run("respects XDG_CACHE_HOME", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", "/tmp/test-cache")
		dir := CacheDir()
		assert.Equal(t, "/tmp/test-cache/go-ent", dir)
	})
}

func TestLegacyDir(t *testing.T) {
	dir := LegacyDir()
	home, _ := os.UserHomeDir()
	assert.Equal(t, filepath.Join(home, ".go-ent"), dir)
}

func TestMigrateIfNeeded(t *testing.T) {
	t.Run("no-op when legacy dir missing", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		err := MigrateIfNeeded()
		assert.NoError(t, err)
	})

	t.Run("migrates models.yaml", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("HOME", tmp)
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "config"))
		t.Setenv("XDG_DATA_HOME", filepath.Join(tmp, "data"))

		legacyDir := filepath.Join(tmp, ".go-ent")
		require.NoError(t, os.MkdirAll(legacyDir, 0o750))
		require.NoError(t, os.WriteFile(
			filepath.Join(legacyDir, "models.yaml"),
			[]byte("version: \"1\""),
			0o600,
		))

		err := MigrateIfNeeded()
		require.NoError(t, err)

		data, err := os.ReadFile(filepath.Join(tmp, "config", "go-ent", "models.yaml"))
		require.NoError(t, err)
		assert.Equal(t, "version: \"1\"", string(data))
	})

	t.Run("does not overwrite existing", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("HOME", tmp)
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "config"))
		t.Setenv("XDG_DATA_HOME", filepath.Join(tmp, "data"))

		legacyDir := filepath.Join(tmp, ".go-ent")
		require.NoError(t, os.MkdirAll(legacyDir, 0o750))
		require.NoError(t, os.WriteFile(
			filepath.Join(legacyDir, "models.yaml"),
			[]byte("old"),
			0o600,
		))

		xdgDir := filepath.Join(tmp, "config", "go-ent")
		require.NoError(t, os.MkdirAll(xdgDir, 0o750))
		require.NoError(t, os.WriteFile(
			filepath.Join(xdgDir, "models.yaml"),
			[]byte("new"),
			0o600,
		))

		err := MigrateIfNeeded()
		require.NoError(t, err)

		data, err := os.ReadFile(filepath.Join(xdgDir, "models.yaml"))
		require.NoError(t, err)
		assert.Equal(t, "new", string(data))
	})

	t.Run("migrates template directory", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("HOME", tmp)
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "config"))
		t.Setenv("XDG_DATA_HOME", filepath.Join(tmp, "data"))

		tplDir := filepath.Join(tmp, ".go-ent", "templates", "skills", "my-tpl")
		require.NoError(t, os.MkdirAll(tplDir, 0o750))
		require.NoError(t, os.WriteFile(
			filepath.Join(tplDir, "SKILL.md"),
			[]byte("# skill"),
			0o600,
		))

		err := MigrateIfNeeded()
		require.NoError(t, err)

		data, err := os.ReadFile(filepath.Join(tmp, "data", "go-ent", "templates", "skills", "my-tpl", "SKILL.md"))
		require.NoError(t, err)
		assert.Equal(t, "# skill", string(data))
	})
}
