package xdg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
