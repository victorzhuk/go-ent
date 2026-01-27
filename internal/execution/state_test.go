package execution

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victorzhuk/go-ent/internal/domain"
)

func TestNewExecutionState(t *testing.T) {
	t.Run("creates state with default values", func(t *testing.T) {
		task := NewTask("test task").WithType("feature")

		state := NewExecutionState(task)

		assert.NotEmpty(t, state.ID)
		assert.Equal(t, task, state.Task)
		assert.Equal(t, task.Context, state.Context)
		assert.Equal(t, ExecutionStatusPending, state.Status)
		assert.Equal(t, StateVersion, state.Version)
		assert.NotEmpty(t, state.Checksum)
		assert.False(t, state.CreatedAt.IsZero())
		assert.False(t, state.UpdatedAt.IsZero())
		assert.True(t, state.StartedAt.IsZero())
		assert.True(t, state.CompletedAt.IsZero())
		assert.NotNil(t, state.Metadata)
		assert.True(t, state.ValidateChecksum())
	})
}

func TestExecutionState_WithConfig(t *testing.T) {
	t.Run("sets configuration correctly", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)

		cfg := ExecutionConfig{
			Agent:    domain.AgentRoleDeveloper,
			Model:    "claude-3.5-sonnet",
			Runtime:  domain.RuntimeClaudeCode,
			Strategy: domain.ExecutionStrategySingle,
		}

		result := state.WithConfig(cfg)

		assert.Equal(t, cfg.Agent, result.Agent)
		assert.Equal(t, cfg.Model, result.Model)
		assert.Equal(t, cfg.Runtime, result.Runtime)
		assert.Equal(t, cfg.Strategy, result.Strategy)
		assert.True(t, result.ValidateChecksum())
	})
}

func TestExecutionState_Start(t *testing.T) {
	t.Run("starts pending execution", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)

		err := state.Start()

		assert.NoError(t, err)
		assert.Equal(t, ExecutionStatusRunning, state.Status)
		assert.False(t, state.StartedAt.IsZero())
		assert.True(t, state.CompletedAt.IsZero())
		assert.True(t, state.ValidateChecksum())
	})

	t.Run("fails to start non-pending execution", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)
		state.Status = ExecutionStatusRunning

		err := state.Start()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot start execution")
	})
}

func TestExecutionState_Complete(t *testing.T) {
	t.Run("completes running execution", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)
		state.Start()

		result := &Result{
			Success: true,
			Output:  "task completed",
		}

		err := state.Complete(result)

		assert.NoError(t, err)
		assert.Equal(t, ExecutionStatusCompleted, state.Status)
		assert.Equal(t, result, state.Result)
		assert.False(t, state.CompletedAt.IsZero())
		assert.True(t, state.ValidateChecksum())
	})

	t.Run("fails to complete non-running execution", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)
		result := &Result{}

		err := state.Complete(result)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot complete execution")
	})
}

func TestExecutionState_Fail(t *testing.T) {
	t.Run("fails running execution", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)
		state.Start()

		err := state.Fail(assert.AnError)

		assert.NoError(t, err)
		assert.Equal(t, ExecutionStatusFailed, state.Status)
		assert.NotNil(t, state.Result)
		assert.False(t, state.Result.Success)
		assert.Equal(t, assert.AnError.Error(), state.Result.Error)
		assert.False(t, state.CompletedAt.IsZero())
		assert.True(t, state.ValidateChecksum())
	})
}

func TestExecutionState_Interrupt(t *testing.T) {
	t.Run("interrupts running execution", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)
		state.Start()

		err := state.Interrupt()

		assert.NoError(t, err)
		assert.Equal(t, ExecutionStatusInterrupted, state.Status)
		assert.True(t, state.ValidateChecksum())
	})
}

func TestExecutionState_UpdateContext(t *testing.T) {
	t.Run("updates context", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)

		newCtx := &TaskContext{
			ProjectPath: "/new/path",
			ChangeID:    "change-2",
			TaskID:      "task-2",
		}

		err := state.UpdateContext(newCtx)

		assert.NoError(t, err)
		assert.Equal(t, newCtx, state.Context)
		assert.True(t, state.ValidateChecksum())
	})
}

func TestExecutionState_Metadata(t *testing.T) {
	t.Run("sets and gets metadata", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)

		state.SetMetadata("key1", "value1")
		state.SetMetadata("key2", "value2")

		val1, ok1 := state.GetMetadata("key1")
		assert.True(t, ok1)
		assert.Equal(t, "value1", val1)

		val2, ok2 := state.GetMetadata("key2")
		assert.True(t, ok2)
		assert.Equal(t, "value2", val2)

		val3, ok3 := state.GetMetadata("key3")
		assert.False(t, ok3)
		assert.Empty(t, val3)
	})
}

func TestExecutionState_StatusMethods(t *testing.T) {
	t.Run("status methods return correct values", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)

		assert.False(t, state.IsRunning())
		assert.False(t, state.IsCompleted())
		assert.False(t, state.IsFailed())
		assert.False(t, state.IsInterrupted())

		state.Start()
		assert.True(t, state.IsRunning())
		assert.False(t, state.IsCompleted())

		state.Complete(&Result{})
		assert.False(t, state.IsRunning())
		assert.True(t, state.IsCompleted())
	})
}

func TestExecutionState_Duration(t *testing.T) {
	t.Run("calculates duration correctly", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)

		assert.Zero(t, state.Duration())

		now := time.Now()
		state.StartedAt = now.Add(-time.Second)
		assert.GreaterOrEqual(t, state.Duration(), time.Second)

		state.CompletedAt = now.Add(-500 * time.Millisecond)
		assert.Equal(t, 500*time.Millisecond, state.Duration())
	})
}

func TestExecutionState_Checksum(t *testing.T) {
	t.Run("computes valid checksum", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)

		assert.NotEmpty(t, state.Checksum)
		assert.True(t, state.ValidateChecksum())
	})

	t.Run("updates checksum on changes", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)
		originalChecksum := state.Checksum

		state.Status = ExecutionStatusRunning

		assert.NotEqual(t, originalChecksum, state.computeChecksum())
	})
}

func TestExecutionState_Serialization(t *testing.T) {
	t.Run("serializes to JSON", func(t *testing.T) {
		task := NewTask("test task").WithType("feature")
		state := NewExecutionState(task)
		state.Start()

		data, err := state.ToJSON()

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
		assert.Contains(t, string(data), `"id"`)
		assert.Contains(t, string(data), `"status"`)
		assert.Contains(t, string(data), `"version"`)
	})

	t.Run("deserializes from JSON", func(t *testing.T) {
		task := NewTask("test task").WithType("feature")
		original := NewExecutionState(task)
		original.Start()

		data, err := original.ToJSON()
		assert.NoError(t, err)

		restored := &ExecutionState{}
		err = restored.FromJSON(data)

		assert.NoError(t, err)
		assert.Equal(t, original.ID, restored.ID)
		assert.Equal(t, original.Status, restored.Status)
		assert.Equal(t, original.Version, restored.Version)
		assert.Equal(t, original.Task.Description, restored.Task.Description)
		assert.True(t, restored.ValidateChecksum())
	})

	t.Run("preserves all fields through serialize/deserialize", func(t *testing.T) {
		task := NewTask("test task").WithType("feature")
		original := NewExecutionState(task)
		original.Start()
		original.SetMetadata("key1", "value1")

		cfg := ExecutionConfig{
			Agent:    domain.AgentRoleDeveloper,
			Model:    "claude-3.5-sonnet",
			Runtime:  domain.RuntimeClaudeCode,
			Strategy: domain.ExecutionStrategySingle,
		}
		original.WithConfig(cfg)

		data, err := original.ToJSON()
		assert.NoError(t, err)

		restored := &ExecutionState{}
		err = restored.FromJSON(data)
		assert.NoError(t, err)

		assert.Equal(t, original.ID, restored.ID)
		assert.Equal(t, original.Status, restored.Status)
		assert.Equal(t, original.Version, restored.Version)
		assert.Equal(t, original.Checksum, restored.Checksum)
		assert.Equal(t, original.Agent, restored.Agent)
		assert.Equal(t, original.Model, restored.Model)
		assert.Equal(t, original.Runtime, restored.Runtime)
		assert.Equal(t, original.Strategy, restored.Strategy)
		assert.Equal(t, original.StartedAt.Unix(), restored.StartedAt.Unix())
		assert.Equal(t, original.CreatedAt.Unix(), restored.CreatedAt.Unix())
		assert.Equal(t, original.UpdatedAt.Unix(), restored.UpdatedAt.Unix())
		assert.Equal(t, original.Metadata["key1"], restored.Metadata["key1"])
		assert.True(t, restored.ValidateChecksum())
	})

	t.Run("handles result in serialization", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)
		state.Start()

		result := &Result{
			Success:     true,
			Output:      "test output",
			TokensIn:    100,
			TokensOut:   50,
			Cost:        0.15,
			Duration:    5 * time.Second,
			Adjustments: []string{"adjustment1", "adjustment2"},
			Metadata: map[string]interface{}{
				"key": "value",
			},
		}

		err := state.Complete(result)
		assert.NoError(t, err)

		data, err := state.ToJSON()
		assert.NoError(t, err)

		restored := &ExecutionState{}
		err = restored.FromJSON(data)
		assert.NoError(t, err)

		assert.NotNil(t, restored.Result)
		assert.Equal(t, result.Success, restored.Result.Success)
		assert.Equal(t, result.Output, restored.Result.Output)
		assert.Equal(t, result.TokensIn, restored.Result.TokensIn)
		assert.Equal(t, result.TokensOut, restored.Result.TokensOut)
		assert.Equal(t, result.Cost, restored.Result.Cost)
		assert.Equal(t, result.Duration, restored.Result.Duration)
		assert.Equal(t, len(result.Adjustments), len(restored.Result.Adjustments))
		assert.Equal(t, result.Adjustments[0], restored.Result.Adjustments[0])
		assert.True(t, restored.ValidateChecksum())
	})
}

func TestExecutionState_Clone(t *testing.T) {
	t.Run("creates independent copy", func(t *testing.T) {
		task := NewTask("test task").WithType("feature")
		original := NewExecutionState(task)
		original.Start()
		original.SetMetadata("key1", "value1")

		clone := original.Clone()

		assert.Equal(t, original.ID, clone.ID)
		assert.Equal(t, original.Status, clone.Status)
		assert.Equal(t, original.Version, clone.Version)
		assert.Equal(t, original.Checksum, clone.Checksum)

		original.Status = ExecutionStatusCompleted

		assert.NotEqual(t, original.Status, clone.Status)
		assert.Equal(t, ExecutionStatusRunning, clone.Status)
	})

	t.Run("deep copies nested structures", func(t *testing.T) {
		task := NewTask("test task")
		task.Metadata = map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}

		original := NewExecutionState(task)
		original.SetMetadata("meta1", "metaValue1")

		clone := original.Clone()

		original.Task.Metadata["key1"] = "modified"
		original.Metadata["meta1"] = "modified"

		assert.Equal(t, "value1", clone.Task.Metadata["key1"])
		assert.Equal(t, "metaValue1", clone.Metadata["meta1"])
	})

	t.Run("clones context with files", func(t *testing.T) {
		ctx := NewTaskContext("/test/path")
		ctx.WithFiles([]string{"file1.go", "file2.go"})

		task := NewTask("test task").WithContext(ctx)
		original := NewExecutionState(task)

		clone := original.Clone()

		original.Context.Files[0] = "modified.go"

		assert.Equal(t, "file1.go", clone.Context.Files[0])
		assert.Equal(t, 2, len(clone.Context.Files))
	})

	t.Run("clones result with adjustments", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)
		state.Start()

		result := &Result{
			Success:     true,
			Adjustments: []string{"adj1", "adj2", "adj3"},
		}
		state.Complete(result)

		clone := state.Clone()

		clone.Result.Adjustments[0] = "modified"

		assert.Equal(t, "adj1", state.Result.Adjustments[0])
	})
}

func TestExecutionState_VersionHandling(t *testing.T) {
	t.Run("sets correct version", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)

		assert.Equal(t, StateVersion, state.Version)
	})

	t.Run("preserves version through serialization", func(t *testing.T) {
		task := NewTask("test task")
		original := NewExecutionState(task)

		data, _ := original.ToJSON()
		restored := &ExecutionState{}
		restored.FromJSON(data)

		assert.Equal(t, original.Version, restored.Version)
	})
}

func TestExecutionState_ValidateChecksum(t *testing.T) {
	t.Run("returns false for empty checksum", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)
		state.Checksum = ""

		assert.False(t, state.ValidateChecksum())
	})

	t.Run("returns true for valid checksum", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)

		assert.True(t, state.ValidateChecksum())
	})

	t.Run("returns false for modified data", func(t *testing.T) {
		task := NewTask("test task")
		state := NewExecutionState(task)

		state.Status = "modified"

		assert.False(t, state.ValidateChecksum())
	})
}

func TestExecutionState_String(t *testing.T) {
	t.Run("returns formatted string", func(t *testing.T) {
		task := NewTask("test task description")
		state := NewExecutionState(task)

		str := state.String()

		assert.Contains(t, str, "ExecutionState")
		assert.Contains(t, str, state.ID)
		assert.Contains(t, str, state.Status)
		assert.Contains(t, str, "test task description")
	})
}

func TestStorage_EnsureExecutionsDir(t *testing.T) {
	t.Run("creates directory if not exists", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		err := EnsureExecutionsDir()

		assert.NoError(t, err)
		stat, err := os.Stat(testDir)
		assert.NoError(t, err)
		assert.True(t, stat.IsDir())
		assert.Equal(t, os.FileMode(0755), stat.Mode().Perm())
	})

	t.Run("is idempotent", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		err1 := EnsureExecutionsDir()
		err2 := EnsureExecutionsDir()

		assert.NoError(t, err1)
		assert.NoError(t, err2)
	})

	t.Run("handles nested directory creation", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test", "nested", "deep", "path")
		defer os.RemoveAll(filepath.Join(os.TempDir(), "go-ent-test"))

		execDirPath = testDir

		err := EnsureExecutionsDir()

		assert.NoError(t, err)
		stat, err := os.Stat(testDir)
		assert.NoError(t, err)
		assert.True(t, stat.IsDir())
	})
}

func TestStorage_SaveState(t *testing.T) {
	t.Run("saves state to file", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("test task").WithType("feature")
		state := NewExecutionState(task)
		state.Start()

		err := SaveState(state)

		assert.NoError(t, err)

		expectedFile := filepath.Join(testDir, state.ID+".json")
		_, err = os.Stat(expectedFile)
		assert.NoError(t, err)

		data, err := os.ReadFile(expectedFile)
		assert.NoError(t, err)
		assert.Contains(t, string(data), state.ID)
		assert.Contains(t, string(data), `"status"`)
	})

	t.Run("uses atomic write", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("test task")
		state := NewExecutionState(task)

		err := SaveState(state)

		assert.NoError(t, err)

		expectedFile := filepath.Join(testDir, state.ID+".json")
		tempFile := expectedFile + ".tmp"

		_, err = os.Stat(expectedFile)
		assert.NoError(t, err)

		_, err = os.Stat(tempFile)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("returns error for nil state", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		err := SaveState(nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot save nil state")
	})

	t.Run("returns error for state without ID", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("test task")
		state := NewExecutionState(task)
		state.ID = ""

		err := SaveState(state)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state has no ID")
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("test task")
		state := NewExecutionState(task)

		err := SaveState(state)
		assert.NoError(t, err)

		state.Status = ExecutionStatusRunning
		err = SaveState(state)
		assert.NoError(t, err)

		expectedFile := filepath.Join(testDir, state.ID+".json")
		data, err := os.ReadFile(expectedFile)
		assert.NoError(t, err)
		assert.Contains(t, string(data), ExecutionStatusRunning)
	})
}

func TestStorage_LoadState(t *testing.T) {
	t.Run("loads state from file", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("test task").WithType("feature")
		original := NewExecutionState(task)
		original.Start()
		original.SetMetadata("key1", "value1")

		cfg := ExecutionConfig{
			Agent:    domain.AgentRoleDeveloper,
			Model:    "claude-3.5-sonnet",
			Runtime:  domain.RuntimeClaudeCode,
			Strategy: domain.ExecutionStrategySingle,
		}
		original.WithConfig(cfg)

		err := SaveState(original)
		assert.NoError(t, err)

		loaded, err := LoadState(original.ID)

		assert.NoError(t, err)
		assert.NotNil(t, loaded)
		assert.Equal(t, original.ID, loaded.ID)
		assert.Equal(t, original.Status, loaded.Status)
		assert.Equal(t, original.Version, loaded.Version)
		assert.Equal(t, original.Agent, loaded.Agent)
		assert.Equal(t, original.Model, loaded.Model)
		assert.Equal(t, original.Runtime, loaded.Runtime)
		assert.Equal(t, original.Strategy, loaded.Strategy)
		assert.Equal(t, original.Task.Description, loaded.Task.Description)
		assert.Equal(t, "value1", loaded.Metadata["key1"])
		assert.True(t, loaded.ValidateChecksum())
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		_, err := LoadState("nonexistent-id")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error for empty execution ID", func(t *testing.T) {
		_, err := LoadState("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("returns error for corrupted JSON", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		if err := EnsureExecutionsDir(); err != nil {
			t.Fatal(err)
		}

		filename := filepath.Join(testDir, "corrupted-id.json")
		err := os.WriteFile(filename, []byte("invalid json"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		_, err = LoadState("corrupted-id")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "corrupted state")
	})

	t.Run("validates checksum on load", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("test task")
		original := NewExecutionState(task)

		err := SaveState(original)
		assert.NoError(t, err)

		filename := filepath.Join(testDir, original.ID+".json")

		original.Checksum = "invalidchecksum"
		modifiedData, _ := original.ToJSON()
		if err := os.WriteFile(filename, modifiedData, 0644); err != nil {
			t.Fatal(err)
		}

		_, err = LoadState(original.ID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "checksum validation failed")
	})

	t.Run("loads state with result", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("test task")
		original := NewExecutionState(task)
		original.Start()

		result := &Result{
			Success:     true,
			Output:      "test output",
			TokensIn:    100,
			TokensOut:   50,
			Cost:        0.15,
			Duration:    5 * time.Second,
			Adjustments: []string{"adjustment1", "adjustment2"},
		}
		original.Complete(result)

		err := SaveState(original)
		assert.NoError(t, err)

		loaded, err := LoadState(original.ID)

		assert.NoError(t, err)
		assert.NotNil(t, loaded.Result)
		assert.Equal(t, result.Success, loaded.Result.Success)
		assert.Equal(t, result.Output, loaded.Result.Output)
		assert.Equal(t, result.TokensIn, loaded.Result.TokensIn)
		assert.Equal(t, result.TokensOut, loaded.Result.TokensOut)
		assert.Equal(t, result.Cost, loaded.Result.Cost)
		assert.Equal(t, result.Duration, loaded.Result.Duration)
		assert.Equal(t, len(result.Adjustments), len(loaded.Result.Adjustments))
	})
}

func TestStorage_DeleteState(t *testing.T) {
	t.Run("deletes existing state", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("test task")
		state := NewExecutionState(task)

		err := SaveState(state)
		assert.NoError(t, err)

		filename := filepath.Join(testDir, state.ID+".json")
		_, err = os.Stat(filename)
		assert.NoError(t, err)

		err = DeleteState(state.ID)
		assert.NoError(t, err)

		_, err = os.Stat(filename)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("returns error for missing state", func(t *testing.T) {
		err := DeleteState("nonexistent-id")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error for empty execution ID", func(t *testing.T) {
		err := DeleteState("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestStorage_ListExecutions(t *testing.T) {
	t.Run("lists all executions", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task1 := NewTask("task 1")
		state1 := NewExecutionState(task1)

		task2 := NewTask("task 2")
		state2 := NewExecutionState(task2)

		task3 := NewTask("task 3")
		state3 := NewExecutionState(task3)

		err := SaveState(state1)
		assert.NoError(t, err)
		err = SaveState(state2)
		assert.NoError(t, err)
		err = SaveState(state3)
		assert.NoError(t, err)

		executions, err := ListExecutions()

		assert.NoError(t, err)
		assert.Len(t, executions, 3)
		assert.Contains(t, executions, state1.ID)
		assert.Contains(t, executions, state2.ID)
		assert.Contains(t, executions, state3.ID)
	})

	t.Run("returns empty list for no executions", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		executions, err := ListExecutions()

		assert.NoError(t, err)
		assert.Empty(t, executions)
	})

	t.Run("ignores non-json files", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("test task")
		state := NewExecutionState(task)

		err := SaveState(state)
		assert.NoError(t, err)

		nonJSONFile := filepath.Join(testDir, "readme.txt")
		err = os.WriteFile(nonJSONFile, []byte("not json"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		executions, err := ListExecutions()

		assert.NoError(t, err)
		assert.Len(t, executions, 1)
		assert.Contains(t, executions, state.ID)
	})

	t.Run("ignores subdirectories", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("test task")
		state := NewExecutionState(task)

		err := SaveState(state)
		assert.NoError(t, err)

		subdir := filepath.Join(testDir, "subdir")
		err = os.Mkdir(subdir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		executions, err := ListExecutions()

		assert.NoError(t, err)
		assert.Len(t, executions, 1)
		assert.Contains(t, executions, state.ID)
	})
}

func TestStorage_SaveLoadRoundTrip(t *testing.T) {
	t.Run("preserves all state through save/load", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		ctx := NewTaskContext("/test/path")
		ctx.WithFiles([]string{"file1.go", "file2.go"})

		task := NewTask("test task").WithType("feature").WithContext(ctx)
		original := NewExecutionState(task)
		original.Start()

		result := &Result{
			Success:     true,
			Output:      "test output",
			TokensIn:    100,
			TokensOut:   50,
			Cost:        0.15,
			Duration:    5 * time.Second,
			Adjustments: []string{"adjustment1"},
			Metadata: map[string]interface{}{
				"key": "value",
			},
		}
		original.Complete(result)

		cfg := ExecutionConfig{
			Agent:    domain.AgentRoleDeveloper,
			Model:    "claude-3.5-sonnet",
			Runtime:  domain.RuntimeClaudeCode,
			Strategy: domain.ExecutionStrategySingle,
			Budget: &BudgetLimit{
				MaxCost:   10.0,
				MaxTokens: 10000,
			},
			Skills: []string{"go-code", "go-test"},
		}
		original.WithConfig(cfg)

		original.SetMetadata("meta1", "value1")
		original.SetMetadata("meta2", "value2")

		err := SaveState(original)
		assert.NoError(t, err)

		loaded, err := LoadState(original.ID)
		assert.NoError(t, err)

		assert.Equal(t, original.ID, loaded.ID)
		assert.Equal(t, original.Status, loaded.Status)
		assert.Equal(t, original.Version, loaded.Version)
		assert.Equal(t, original.Agent, loaded.Agent)
		assert.Equal(t, original.Model, loaded.Model)
		assert.Equal(t, original.Runtime, loaded.Runtime)
		assert.Equal(t, original.Strategy, loaded.Strategy)
		assert.Equal(t, original.Checksum, loaded.Checksum)
		assert.Equal(t, original.CreatedAt.Unix(), loaded.CreatedAt.Unix())
		assert.Equal(t, original.UpdatedAt.Unix(), loaded.UpdatedAt.Unix())
		assert.Equal(t, original.StartedAt.Unix(), loaded.StartedAt.Unix())
		assert.Equal(t, original.CompletedAt.Unix(), loaded.CompletedAt.Unix())

		assert.Equal(t, original.Task.Description, loaded.Task.Description)
		assert.Equal(t, original.Task.Type, loaded.Task.Type)

		assert.Equal(t, original.Context.ProjectPath, loaded.Context.ProjectPath)
		assert.Equal(t, len(original.Context.Files), len(loaded.Context.Files))
		assert.Equal(t, original.Context.Files[0], loaded.Context.Files[0])

		assert.NotNil(t, loaded.Result)
		assert.Equal(t, original.Result.Success, loaded.Result.Success)
		assert.Equal(t, original.Result.Output, loaded.Result.Output)
		assert.Equal(t, original.Result.TokensIn, loaded.Result.TokensIn)
		assert.Equal(t, original.Result.TokensOut, loaded.Result.TokensOut)

		assert.Equal(t, "value1", loaded.Metadata["meta1"])
		assert.Equal(t, "value2", loaded.Metadata["meta2"])

		assert.NotNil(t, loaded.Strategy)

		assert.True(t, loaded.ValidateChecksum())
	})
}
