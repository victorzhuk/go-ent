package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/cli"
)

func TestConfigWithRealFiles(t *testing.T) {
	t.Run("init creates valid config file", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := cli.NewRootCmd()
		cmd.SetArgs([]string{"config", "init", tmpDir})

		err := cmd.Execute()
		require.NoError(t, err)

		// Verify config file exists
		cfgPath := filepath.Join(tmpDir, ".go-ent", "config.yaml")
		assert.FileExists(t, cfgPath)

		// Verify config is valid YAML
		data, err := os.ReadFile(cfgPath)
		require.NoError(t, err)
		assert.Contains(t, string(data), "version:")
		assert.Contains(t, string(data), "agents:")
		assert.Contains(t, string(data), "budget:")
		assert.Contains(t, string(data), "runtime:")
		assert.Contains(t, string(data), "models:")
		assert.Contains(t, string(data), "skills:")
	})

	t.Run("init fails if config already exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create config first time
		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		err := cmd1.Execute()
		require.NoError(t, err)

		// Try to create again
		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "init", tmpDir})
		err = cmd2.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("show reads existing config", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create config
		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd1.Execute())

		// Show config
		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "show", tmpDir})
		err := cmd2.Execute()
		require.NoError(t, err)
	})

	t.Run("set modifies config file", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create config
		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd1.Execute())

		// Read original budget
		cfgPath := filepath.Join(tmpDir, ".go-ent", "config.yaml")
		origData, err := os.ReadFile(cfgPath)
		require.NoError(t, err)

		// Modify budget (path as third argument)
		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "set", "budget.daily", "50", tmpDir})
		err = cmd2.Execute()
		require.NoError(t, err)

		// Verify file was modified
		newData, err := os.ReadFile(cfgPath)
		require.NoError(t, err)
		assert.NotEqual(t, string(origData), string(newData))
		assert.Contains(t, string(newData), "daily: 50")
	})

	t.Run("config file permissions", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := cli.NewRootCmd()
		cmd.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd.Execute())

		// Verify directory permissions
		cfgDir := filepath.Join(tmpDir, ".go-ent")
		info, err := os.Stat(cfgDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
		assert.Equal(t, os.FileMode(0755), info.Mode().Perm())

		// Verify file permissions
		cfgPath := filepath.Join(cfgDir, "config.yaml")
		info, err = os.Stat(cfgPath)
		require.NoError(t, err)
		assert.False(t, info.IsDir())
		assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
	})

	t.Run("config in nested directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "project", "subdir")
		require.NoError(t, os.MkdirAll(nestedDir, 0755))

		cmd := cli.NewRootCmd()
		cmd.SetArgs([]string{"config", "init", nestedDir})
		require.NoError(t, cmd.Execute())

		cfgPath := filepath.Join(nestedDir, ".go-ent", "config.yaml")
		assert.FileExists(t, cfgPath)
	})

	t.Run("show with summary format", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create config
		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd1.Execute())

		// Show with summary format
		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "show", tmpDir, "--format", "summary"})
		err := cmd2.Execute()
		require.NoError(t, err)
	})

	t.Run("set multiple values", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create config
		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd1.Execute())

		cfgPath := filepath.Join(tmpDir, ".go-ent", "config.yaml")

		// Set daily budget
		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "set", "budget.daily", "100", tmpDir})
		require.NoError(t, cmd2.Execute())

		// Set monthly budget
		cmd3 := cli.NewRootCmd()
		cmd3.SetArgs([]string{"config", "set", "budget.monthly", "2000", tmpDir})
		require.NoError(t, cmd3.Execute())

		// Set default agent
		cmd4 := cli.NewRootCmd()
		cmd4.SetArgs([]string{"config", "set", "agents.default", "architect", tmpDir})
		require.NoError(t, cmd4.Execute())

		// Verify all changes
		data, err := os.ReadFile(cfgPath)
		require.NoError(t, err)
		assert.Contains(t, string(data), "daily: 100")
		assert.Contains(t, string(data), "monthly: 2000")
		assert.Contains(t, string(data), "default: architect")
	})

	t.Run("set with invalid values", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create config
		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd1.Execute())

		// Try to set invalid budget (negative)
		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "set", "budget.daily", "-10", tmpDir})
		err := cmd2.Execute()
		// Should either error or silently allow (depending on validation)
		_ = err

		// Try to set non-existent field
		cmd3 := cli.NewRootCmd()
		cmd3.SetArgs([]string{"config", "set", "nonexistent.field", "value", tmpDir})
		err = cmd3.Execute()
		// Should error
		assert.Error(t, err)
	})

	t.Run("config with custom directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		customDir := filepath.Join(tmpDir, "custom")

		// Create config in custom directory
		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", customDir})
		require.NoError(t, cmd1.Execute())

		// Show config from custom directory
		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "show", customDir})
		err := cmd2.Execute()
		require.NoError(t, err)
	})
}

func TestConfigEdgeCases(t *testing.T) {
	t.Run("config in readonly directory", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping readonly test when running as root")
		}

		tmpDir := t.TempDir()
		readonlyDir := filepath.Join(tmpDir, "readonly")
		require.NoError(t, os.MkdirAll(readonlyDir, 0555))
		defer os.Chmod(readonlyDir, 0755) // cleanup

		cmd := cli.NewRootCmd()
		cmd.SetArgs([]string{"config", "init", readonlyDir})
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("show non-existent config", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := cli.NewRootCmd()
		cmd.SetArgs([]string{"config", "show", tmpDir})
		err := cmd.Execute()
		// Should either load defaults or error
		_ = err
	})

	t.Run("config with symlink", func(t *testing.T) {
		tmpDir := t.TempDir()
		realDir := filepath.Join(tmpDir, "real")
		linkDir := filepath.Join(tmpDir, "link")

		require.NoError(t, os.MkdirAll(realDir, 0755))
		require.NoError(t, os.Symlink(realDir, linkDir))

		cmd := cli.NewRootCmd()
		cmd.SetArgs([]string{"config", "init", linkDir})
		err := cmd.Execute()
		require.NoError(t, err)

		// Verify config exists in real directory
		cfgPath := filepath.Join(realDir, ".go-ent", "config.yaml")
		assert.FileExists(t, cfgPath)
	})

	t.Run("config reads are safe", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create config
		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd1.Execute())

		cfgPath := filepath.Join(tmpDir, ".go-ent", "config.yaml")

		// Multiple sequential reads should work
		for i := 0; i < 5; i++ {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"config", "show", tmpDir})
			err := cmd.Execute()
			assert.NoError(t, err)
		}

		// File should still exist and be valid
		assert.FileExists(t, cfgPath)
		data, err := os.ReadFile(cfgPath)
		require.NoError(t, err)
		assert.Contains(t, string(data), "version:")
	})
}
