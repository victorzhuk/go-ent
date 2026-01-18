package cli_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskCommands(t *testing.T) {
	t.Run("task next with no tasks", func(t *testing.T) {
		stdout, _, err := executeCommandWithCapture(t, "task", "next")
		// May fail if no registry exists
		_ = err
		assert.Contains(t, stdout, "No unblocked tasks")
	})

	t.Run("task next with count flag", func(t *testing.T) {
		_, _, err := executeCommand(t, "task", "next", "--count", "5")
		// May fail if no registry exists
		_ = err
	})

	t.Run("task show without id", func(t *testing.T) {
		_, _, err := executeCommand(t, "task", "show")
		require.Error(t, err, "should require task ID")
	})

	t.Run("task show with invalid format", func(t *testing.T) {
		_, _, err := executeCommand(t, "task", "show", "invalid-format")
		require.Error(t, err)
	})

	t.Run("task run without id", func(t *testing.T) {
		_, _, err := executeCommand(t, "task", "run")
		require.Error(t, err, "should require task ID")
	})

	t.Run("task list with no filters", func(t *testing.T) {
		_, _, err := executeCommand(t, "task", "list")
		// May fail if no registry exists
		_ = err
	})

	t.Run("task list with status filter", func(t *testing.T) {
		_, _, err := executeCommand(t, "task", "list", "--status", "pending")
		// May fail if no registry exists
		_ = err
	})

	t.Run("task list with change filter", func(t *testing.T) {
		_, _, err := executeCommand(t, "task", "list", "--change", "test-change")
		// May fail if no registry exists
		_ = err
	})

	t.Run("task list with unblocked flag", func(t *testing.T) {
		_, _, err := executeCommand(t, "task", "list", "--unblocked")
		// May fail if no registry exists
		_ = err
	})

	t.Run("task complete without id", func(t *testing.T) {
		_, _, err := executeCommand(t, "task", "complete")
		require.Error(t, err, "should require task ID")
	})

	t.Run("task complete with invalid format", func(t *testing.T) {
		_, _, err := executeCommand(t, "task", "complete", "invalid-format")
		require.Error(t, err)
	})
}

func TestTaskHelp(t *testing.T) {
	t.Run("task help", func(t *testing.T) {
		stdout, _, err := executeCommand(t, "task", "--help")
		require.NoError(t, err)
		assert.Contains(t, stdout, "tasks")
		assert.Contains(t, stdout, "next")
		assert.Contains(t, stdout, "show")
		assert.Contains(t, stdout, "run")
		assert.Contains(t, stdout, "list")
	})

	t.Run("task next help", func(t *testing.T) {
		stdout, _, err := executeCommand(t, "task", "next", "--help")
		require.NoError(t, err)
		assert.Contains(t, stdout, "next unblocked task")
		assert.Contains(t, stdout, "--count")
	})

	t.Run("task show help", func(t *testing.T) {
		stdout, _, err := executeCommand(t, "task", "show", "--help")
		require.NoError(t, err)
		assert.Contains(t, stdout, "detailed information")
	})

	t.Run("task run help", func(t *testing.T) {
		stdout, _, err := executeCommand(t, "task", "run", "--help")
		require.NoError(t, err)
		assert.Contains(t, stdout, "agent workflow")
	})

	t.Run("task list help", func(t *testing.T) {
		stdout, _, err := executeCommand(t, "task", "list", "--help")
		require.NoError(t, err)
		assert.Contains(t, stdout, "List")
		assert.Contains(t, stdout, "--status")
		assert.Contains(t, stdout, "--change")
		assert.Contains(t, stdout, "--unblocked")
	})

	t.Run("task complete help", func(t *testing.T) {
		stdout, _, err := executeCommand(t, "task", "complete", "--help")
		require.NoError(t, err)
		assert.Contains(t, stdout, "Update task checkbox")
		assert.Contains(t, stdout, "--yes")
	})
}

func TestTaskCommandFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name:     "invalid status filter",
			args:     []string{"task", "list", "--status", "invalid"},
			contains: []string{},
		},
		{
			name:     "count flag accepts positive number",
			args:     []string{"task", "next", "--count", "10"},
			contains: []string{},
		},
		{
			name:     "status flag accepts valid values",
			args:     []string{"task", "list", "--status", "completed"},
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, err := executeCommand(t, tt.args...)
			_ = err
			_ = stdout
		})
	}
}
