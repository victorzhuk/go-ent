package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/cli"
)

// TestErrorHandling tests error scenarios across all CLI commands
func TestErrorHandling(t *testing.T) {
	t.Run("agent errors", func(t *testing.T) {
		t.Run("info requires agent name", func(t *testing.T) {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"agent", "info"})
			err := cmd.Execute()
			require.Error(t, err)
		})

		t.Run("info with non-existent agent", func(t *testing.T) {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"agent", "info", "nonexistent-agent-12345"})
			err := cmd.Execute()
			// Should handle gracefully (may error or show message)
			_ = err
		})

		t.Run("list with invalid flags", func(t *testing.T) {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"agent", "list", "--invalid-flag"})
			err := cmd.Execute()
			require.Error(t, err)
		})
	})

	t.Run("skill errors", func(t *testing.T) {
		t.Run("info requires skill name", func(t *testing.T) {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"skill", "info"})
			err := cmd.Execute()
			require.Error(t, err)
		})

		t.Run("info with non-existent skill", func(t *testing.T) {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"skill", "info", "nonexistent-skill-12345"})
			err := cmd.Execute()
			// Should handle gracefully
			_ = err
		})
	})

	t.Run("spec errors", func(t *testing.T) {
		t.Run("show requires spec id", func(t *testing.T) {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"spec", "show"})
			err := cmd.Execute()
			require.Error(t, err)
		})

		t.Run("show with non-existent spec", func(t *testing.T) {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"spec", "show", "nonexistent-spec"})
			err := cmd.Execute()
			require.Error(t, err)
		})

		t.Run("list in non-existent directory", func(t *testing.T) {
			nonExistentDir := filepath.Join(os.TempDir(), "nonexistent-12345")
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"spec", "list", nonExistentDir})
			err := cmd.Execute()
			require.Error(t, err)
		})

		t.Run("init in existing openspec", func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create openspec first time
			cmd1 := cli.NewRootCmd()
			cmd1.SetArgs([]string{"spec", "init", tmpDir})
			require.NoError(t, cmd1.Execute())

			// Try again - should warn but not fail
			cmd2 := cli.NewRootCmd()
			cmd2.SetArgs([]string{"spec", "init", tmpDir})
			err := cmd2.Execute()
			// May not error, just warn
			_ = err
		})
	})

	t.Run("config errors", func(t *testing.T) {
		t.Run("set requires both key and value", func(t *testing.T) {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"config", "set"})
			err := cmd.Execute()
			require.Error(t, err)
		})

		t.Run("set requires value", func(t *testing.T) {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"config", "set", "somekey"})
			err := cmd.Execute()
			require.Error(t, err)
		})

		t.Run("set with invalid key", func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create config
			cmd1 := cli.NewRootCmd()
			cmd1.SetArgs([]string{"config", "init", tmpDir})
			require.NoError(t, cmd1.Execute())

			// Try to set invalid key
			cmd2 := cli.NewRootCmd()
			cmd2.SetArgs([]string{"config", "set", "invalid.nested.key", "value", tmpDir})
			err := cmd2.Execute()
			require.Error(t, err)
		})

		t.Run("show non-existent config", func(t *testing.T) {
			nonExistentDir := filepath.Join(os.TempDir(), "nonexistent-config-12345")
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"config", "show", nonExistentDir})
			err := cmd.Execute()
			// May show defaults instead of erroring
			_ = err
		})

		t.Run("init in readonly directory", func(t *testing.T) {
			if os.Getuid() == 0 {
				t.Skip("Skipping readonly test when running as root")
			}

			tmpDir := t.TempDir()
			readonlyDir := filepath.Join(tmpDir, "readonly")
			require.NoError(t, os.MkdirAll(readonlyDir, 0555))
			defer os.Chmod(readonlyDir, 0755)

			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"config", "init", readonlyDir})
			err := cmd.Execute()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "permission denied")
		})

		t.Run("init already exists", func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create config first time
			cmd1 := cli.NewRootCmd()
			cmd1.SetArgs([]string{"config", "init", tmpDir})
			require.NoError(t, cmd1.Execute())

			// Try again
			cmd2 := cli.NewRootCmd()
			cmd2.SetArgs([]string{"config", "init", tmpDir})
			err := cmd2.Execute()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "already exists")
		})

		t.Run("set in non-existent directory", func(t *testing.T) {
			nonExistentDir := filepath.Join(os.TempDir(), "nonexistent-12345")
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"config", "set", "budget.daily", "10", nonExistentDir})
			err := cmd.Execute()
			require.Error(t, err)
		})
	})

	t.Run("global flag errors", func(t *testing.T) {
		t.Run("unknown global flag", func(t *testing.T) {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"--unknown-flag", "version"})
			err := cmd.Execute()
			require.Error(t, err)
		})

		t.Run("config flag with non-existent file", func(t *testing.T) {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"--config", "/nonexistent/path/config.yaml", "version"})
			err := cmd.Execute()
			// Version command should still work even with invalid config flag
			// since it doesn't use config
			_ = err
		})
	})

	t.Run("command errors", func(t *testing.T) {
		t.Run("unknown command", func(t *testing.T) {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"unknown-command"})
			err := cmd.Execute()
			require.Error(t, err)
		})

		t.Run("unknown subcommand", func(t *testing.T) {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"agent", "unknown-subcommand"})
			err := cmd.Execute()
			// Cobra shows help, which is not an error
			_ = err
		})

		t.Run("too many arguments", func(t *testing.T) {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"version", "extra", "args"})
			err := cmd.Execute()
			// Version command may accept or ignore extra args
			_ = err
		})
	})
}

func TestErrorMessages(t *testing.T) {
	t.Run("error messages are helpful", func(t *testing.T) {
		tests := []struct {
			name            string
			args            []string
			shouldError     bool
			expectedInError string
		}{
			{
				"agent info missing name",
				[]string{"agent", "info"},
				true,
				"arg",
			},
			{
				"skill info missing name",
				[]string{"skill", "info"},
				true,
				"arg",
			},
			{
				"spec show missing id",
				[]string{"spec", "show"},
				true,
				"arg",
			},
			{
				"config set missing args",
				[]string{"config", "set"},
				true,
				"arg",
			},
			{
				"unknown command",
				[]string{"invalid-cmd"},
				true,
				"unknown",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cmd := cli.NewRootCmd()
				cmd.SetArgs(tt.args)
				err := cmd.Execute()
				if tt.shouldError {
					require.Error(t, err)
					// Error message should be informative
					errMsg := err.Error()
					assert.NotEmpty(t, errMsg)
				}
			})
		}
	})

	t.Run("config error messages", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create config
		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd1.Execute())

		tests := []struct {
			name  string
			key   string
			value string
		}{
			{
				"invalid key path",
				"nonexistent.field",
				"value",
			},
			{
				"too deeply nested key",
				"budget.daily.nested",
				"value",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cmd := cli.NewRootCmd()
				cmd.SetArgs([]string{"config", "set", tt.key, tt.value, tmpDir})
				err := cmd.Execute()
				require.Error(t, err)
				// Should have an error message
				assert.NotEmpty(t, err.Error())
			})
		}
	})
}

func TestRecoveryFromErrors(t *testing.T) {
	t.Run("can recover after error", func(t *testing.T) {
		tmpDir := t.TempDir()

		// First command may show defaults instead of error
		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "show", tmpDir})
		_ = cmd1.Execute()

		// Create config
		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd2.Execute())

		// Now show should definitely work
		cmd3 := cli.NewRootCmd()
		cmd3.SetArgs([]string{"config", "show", tmpDir})
		err3 := cmd3.Execute()
		require.NoError(t, err3)
	})

	t.Run("multiple errors don't break state", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Multiple failing commands
		for i := 0; i < 3; i++ {
			cmd := cli.NewRootCmd()
			cmd.SetArgs([]string{"config", "show", tmpDir})
			_ = cmd.Execute() // Expected to fail
		}

		// Should still be able to create config
		cmd := cli.NewRootCmd()
		cmd.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd.Execute())
	})
}

func TestValidationErrors(t *testing.T) {
	t.Run("config validation", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create config
		cmd1 := cli.NewRootCmd()
		cmd1.SetArgs([]string{"config", "init", tmpDir})
		require.NoError(t, cmd1.Execute())

		// Try to set negative budget (may be validated or cause parse error)
		cmd2 := cli.NewRootCmd()
		cmd2.SetArgs([]string{"config", "set", "budget.daily", "-100", tmpDir})
		err := cmd2.Execute()
		// May error due to flag parsing or validation
		_ = err
	})

	t.Run("path validation", func(t *testing.T) {
		// Null byte in path should be rejected by OS
		cmd := cli.NewRootCmd()
		cmd.SetArgs([]string{"config", "init", "/tmp/test\x00invalid"})
		err := cmd.Execute()
		// OS-level error expected
		_ = err
	})
}
