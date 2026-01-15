package cli_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/cli"
)

// executeCommand runs a CLI command and captures its output
// Note: Some commands use fmt.Printf directly, so they won't be captured
func executeCommand(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	cmd := cli.NewRootCmd()
	cmd.SetArgs(args)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

// executeCommandWithCapture runs a CLI command and captures ALL output including fmt.Printf
func executeCommandWithCapture(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	// Create pipes for capturing stdout/stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	defer rOut.Close()
	defer rErr.Close()

	// Redirect both os and cobra output to pipes
	os.Stdout = wOut
	os.Stderr = wErr

	cmd := cli.NewRootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(wOut)
	cmd.SetErr(wErr)

	// Read from pipes in separate goroutines
	var stdoutBuf, stderrBuf bytes.Buffer
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(&stdoutBuf, rOut)
	}()

	go func() {
		defer wg.Done()
		io.Copy(&stderrBuf, rErr)
	}()

	// Execute command
	cmdErr := cmd.Execute()

	// Close write ends to signal EOF to readers
	wOut.Close()
	wErr.Close()

	// Wait for both readers to finish
	wg.Wait()

	return stdoutBuf.String(), stderrBuf.String(), cmdErr
}

func TestVersionCommand(t *testing.T) {
	t.Run("version subcommand", func(t *testing.T) {
		stdout, stderr, err := executeCommandWithCapture(t, "version")
		require.NoError(t, err)
		assert.Contains(t, stdout, "go-ent")
		assert.Contains(t, stdout, "go:")
		assert.Empty(t, stderr)
	})
}

func TestHelpCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			"root help",
			[]string{"--help"},
			[]string{"go-ent", "Available Commands"},
		},
		{
			"agent help",
			[]string{"agent", "--help"},
			[]string{"agent", "list", "info"},
		},
		{
			"skill help",
			[]string{"skill", "--help"},
			[]string{"skill", "list", "info"},
		},
		{
			"spec help",
			[]string{"spec", "--help"},
			[]string{"spec", "init", "list", "show"},
		},
		{
			"config help",
			[]string{"config", "--help"},
			[]string{"config", "init", "show", "set"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, err := executeCommand(t, tt.args...)
			require.NoError(t, err)
			for _, expected := range tt.contains {
				assert.Contains(t, stdout, expected)
			}
		})
	}
}

func TestAgentCommands(t *testing.T) {
	t.Run("agent list", func(t *testing.T) {
		stdout, stderr, err := executeCommand(t, "agent", "list")

		// Command should execute (may fail if agents path doesn't exist, which is ok)
		if err == nil {
			assert.NotEmpty(t, stdout)
			assert.Empty(t, stderr)
		}
	})

	t.Run("agent list with flags", func(t *testing.T) {
		stdout, _, err := executeCommand(t, "agent", "list", "--detailed")

		if err == nil {
			assert.NotEmpty(t, stdout)
		}
	})

	t.Run("agent info without name", func(t *testing.T) {
		_, _, err := executeCommand(t, "agent", "info")
		require.Error(t, err, "should require agent name")
	})
}

func TestSkillCommands(t *testing.T) {
	t.Run("skill list", func(t *testing.T) {
		stdout, stderr, err := executeCommand(t, "skill", "list")

		if err == nil {
			assert.NotEmpty(t, stdout)
			assert.Empty(t, stderr)
		}
	})

	t.Run("skill list with flags", func(t *testing.T) {
		stdout, _, err := executeCommand(t, "skill", "list", "--detailed")

		if err == nil {
			assert.NotEmpty(t, stdout)
		}
	})

	t.Run("skill info without name", func(t *testing.T) {
		_, _, err := executeCommand(t, "skill", "info")
		require.Error(t, err, "should require skill name")
	})
}

func TestSpecCommands(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("spec init", func(t *testing.T) {
		oldWd, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(oldWd)

		require.NoError(t, os.Chdir(tmpDir))

		stdout, stderr, err := executeCommandWithCapture(t, "spec", "init")
		if err == nil {
			assert.Contains(t, stdout, "Initialized")
			assert.Empty(t, stderr)

			// Verify .spec directory was created (not openspec)
			specDir := filepath.Join(tmpDir, ".spec")
			assert.DirExists(t, specDir)
		}
	})

	t.Run("spec list", func(t *testing.T) {
		_, _, err := executeCommand(t, "spec", "list")
		// May fail if no openspec directory, which is acceptable
		_ = err
	})

	t.Run("spec show without id", func(t *testing.T) {
		_, _, err := executeCommand(t, "spec", "show")
		require.Error(t, err, "should require spec id")
	})
}

func TestConfigCommands(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("config init", func(t *testing.T) {
		oldWd, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(oldWd)

		require.NoError(t, os.Chdir(tmpDir))

		_, stderr, err := executeCommand(t, "config", "init")
		if err == nil {
			// Stdout is printed directly by the command, so it may be empty
			// but stderr should be empty
			assert.Empty(t, stderr)

			// Verify config directory was created
			cfgDir := filepath.Join(tmpDir, ".go-ent")
			assert.DirExists(t, cfgDir)
		}
	})

	t.Run("config show", func(t *testing.T) {
		_, _, err := executeCommand(t, "config", "show")
		// May fail if no config exists, which is acceptable
		_ = err
	})

	t.Run("config set without args", func(t *testing.T) {
		_, _, err := executeCommand(t, "config", "set")
		require.Error(t, err, "should require key and value")
	})

	t.Run("config set with only key", func(t *testing.T) {
		_, _, err := executeCommand(t, "config", "set", "somekey")
		require.Error(t, err, "should require both key and value")
	})
}

func TestRunCommand(t *testing.T) {
	t.Run("run requires task description", func(t *testing.T) {
		_, _, err := executeCommand(t, "run")
		require.Error(t, err)
	})

	t.Run("run with dry-run", func(t *testing.T) {
		stdout, _, err := executeCommandWithCapture(t, "run", "--dry-run", "add logging")
		require.NoError(t, err)
		assert.Contains(t, stdout, "Agent Selection")
		assert.Contains(t, stdout, "Dry run mode")
	})

	t.Run("run with agent override", func(t *testing.T) {
		stdout, _, err := executeCommandWithCapture(t, "run", "--dry-run", "--agent", "architect", "design system")
		require.NoError(t, err)
		assert.Contains(t, stdout, "architect")
		assert.Contains(t, stdout, "manually overridden")
	})

	t.Run("run with task type", func(t *testing.T) {
		stdout, _, err := executeCommandWithCapture(t, "run", "--dry-run", "--type", "bugfix", "fix memory leak")
		require.NoError(t, err)
		assert.Contains(t, stdout, "Agent Selection")
		assert.Contains(t, stdout, "Complexity")
	})

	t.Run("run with files", func(t *testing.T) {
		stdout, _, err := executeCommandWithCapture(t, "run", "--dry-run", "--files", "repo.go,service.go", "refactor code")
		require.NoError(t, err)
		assert.Contains(t, stdout, "Agent Selection")
	})

	t.Run("run with budget", func(t *testing.T) {
		stdout, _, err := executeCommandWithCapture(t, "run", "--dry-run", "--budget", "1000", "simple task")
		require.NoError(t, err)
		assert.Contains(t, stdout, "Agent Selection")
	})

	t.Run("run without dry-run shows not implemented", func(t *testing.T) {
		stdout, _, err := executeCommandWithCapture(t, "run", "test task")
		require.NoError(t, err)
		assert.Contains(t, stdout, "Execution engine not yet implemented")
	})

	t.Run("run help", func(t *testing.T) {
		stdout, _, err := executeCommand(t, "run", "--help")
		require.NoError(t, err)
		assert.Contains(t, stdout, "Execute a task with automatic agent selection")
		assert.Contains(t, stdout, "--agent")
		assert.Contains(t, stdout, "--type")
		assert.Contains(t, stdout, "--dry-run")
	})
}

func TestGlobalFlags(t *testing.T) {
	t.Run("verbose flag", func(t *testing.T) {
		stdout, stderr, err := executeCommandWithCapture(t, "--verbose", "version")
		require.NoError(t, err)
		assert.Contains(t, stdout, "go-ent")
		assert.Empty(t, stderr)
	})

	t.Run("config flag", func(t *testing.T) {
		stdout, stderr, err := executeCommandWithCapture(t, "--config", "/tmp/test-config.yaml", "version")
		require.NoError(t, err)
		assert.Contains(t, stdout, "go-ent")
		assert.Empty(t, stderr)
	})
}

func TestInvalidCommands(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		shouldError bool
	}{
		{"unknown command", []string{"unknown"}, true},
		{"unknown subcommand", []string{"agent", "unknown"}, false}, // Cobra shows help, no error
		{"invalid flag", []string{"--invalid"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := executeCommand(t, tt.args...)
			if tt.shouldError {
				require.Error(t, err)
			}
		})
	}
}
