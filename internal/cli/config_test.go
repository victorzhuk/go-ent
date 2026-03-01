package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/cli"
)

func TestConfigInit(t *testing.T) {
	t.Run("creates claude ent.yaml by default", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := cli.NewRootCmd()
		cmd.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd.Execute())

		cfgPath := filepath.Join(tmpDir, ".claude", "ent.yaml")
		assert.FileExists(t, cfgPath)

		data, err := os.ReadFile(cfgPath) //nolint:gosec // test file
		require.NoError(t, err)
		assert.Contains(t, string(data), "models:")
		assert.Contains(t, string(data), "fast: haiku")
		assert.Contains(t, string(data), "main: sonnet")
		assert.Contains(t, string(data), "heavy: opus")
	})

	t.Run("creates opencode ent.yaml with --runtime=opencode", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := cli.NewRootCmd()
		cmd.SetArgs([]string{"config", "init", "--runtime=opencode", tmpDir})
		require.NoError(t, cmd.Execute())

		cfgPath := filepath.Join(tmpDir, ".opencode", "ent.yaml")
		assert.FileExists(t, cfgPath)

		data, err := os.ReadFile(cfgPath) //nolint:gosec // test file
		require.NoError(t, err)
		assert.Contains(t, string(data), "models:")
		assert.Contains(t, string(data), "anthropic/claude-sonnet-4-6")
	})

	t.Run("fails if config already exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd1.Execute())

		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "init", tmpDir})
		err := cmd2.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("creates directory with correct permissions", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := cli.NewRootCmd()
		cmd.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd.Execute())

		cfgDir := filepath.Join(tmpDir, ".claude")
		info, err := os.Stat(cfgDir)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o750), info.Mode().Perm())

		cfgPath := filepath.Join(cfgDir, "ent.yaml")
		info, err = os.Stat(cfgPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})
}

func TestConfigShow(t *testing.T) {
	t.Run("errors when no config file exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := cli.NewRootCmd()
		cmd.SetArgs([]string{"config", "show", tmpDir})
		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ent config init")
	})

	t.Run("shows existing config", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd1.Execute())

		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "show", tmpDir})
		require.NoError(t, cmd2.Execute())
	})
}

func TestConfigSet(t *testing.T) {
	t.Run("sets models.main tier", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd1.Execute())

		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "set", "models.main", "claude-sonnet-4-6-20260101", tmpDir})
		require.NoError(t, cmd2.Execute())

		data, err := os.ReadFile(filepath.Join(tmpDir, ".claude", "ent.yaml")) //nolint:gosec // test file
		require.NoError(t, err)
		assert.Contains(t, string(data), "claude-sonnet-4-6-20260101")
	})

	t.Run("sets per-agent override", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd1.Execute())

		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "set", "models.agents.coder", "heavy", tmpDir})
		require.NoError(t, cmd2.Execute())

		data, err := os.ReadFile(filepath.Join(tmpDir, ".claude", "ent.yaml")) //nolint:gosec // test file
		require.NoError(t, err)
		assert.Contains(t, string(data), "agents:")
		assert.Contains(t, string(data), "coder: heavy")
	})

	t.Run("creates config when no file exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := cli.NewRootCmd()
		cmd.SetArgs([]string{"config", "set", "models.main", "claude-sonnet-x", tmpDir})
		require.NoError(t, cmd.Execute())

		data, err := os.ReadFile(filepath.Join(tmpDir, ".claude", "ent.yaml")) //nolint:gosec // test file
		require.NoError(t, err)
		assert.Contains(t, string(data), "claude-sonnet-x")
	})

	t.Run("rejects unknown key", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd1.Execute())

		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "set", "budget.daily", "50", tmpDir})
		err := cmd2.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown config key")
	})

	t.Run("rejects old claude.main key", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd1.Execute())

		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "set", "claude.main", "some-model", tmpDir})
		err := cmd2.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown config key")
	})
}
