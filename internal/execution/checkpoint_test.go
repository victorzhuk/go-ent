package execution

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/domain"
)

func TestEngine_AutoCheckpoint(t *testing.T) {
	t.Run("saves checkpoint on task completion", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger:               nil,
			EnableAutoCheckpoint: true,
			MaxCheckpoints:       10,
			CheckpointAgeLimit:   24 * time.Hour,
		}
		engine := New(cfg, selector)

		task := NewTask("test task").WithType("feature")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
			ChangeID:    "change-1",
			TaskID:      "task-1",
		}
		task.Context = ctx

		// Create execution state directly to test checkpointing
		state := NewExecutionState(task)
		state.Start()
		state.Complete(&Result{Success: true})

		// Save checkpoint
		err := engine.createCheckpoint(state)
		assert.NoError(t, err)

		// Verify checkpoint was created
		executions, err := ListExecutions()
		assert.NoError(t, err)
		assert.Len(t, executions, 1)

		// Verify checkpoint contains expected data
		loadedState, err := LoadState(executions[0])
		assert.NoError(t, err)
		assert.Equal(t, ExecutionStatusCompleted, loadedState.Status)
		assert.NotNil(t, loadedState.Task)
	})

	t.Run("saves checkpoint on task failure", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger:               nil,
			EnableAutoCheckpoint: true,
			MaxCheckpoints:       10,
			CheckpointAgeLimit:   24 * time.Hour,
		}

		engine := New(cfg, selector)

		// Create a task that failed
		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
			ChangeID:    "change-1",
			TaskID:      "task-1",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Start()
		state.Fail(assert.AnError)

		// Save checkpoint
		err := engine.createCheckpoint(state)
		assert.NoError(t, err)

		// Verify checkpoint was created
		executions, err := ListExecutions()
		assert.NoError(t, err)
		assert.Len(t, executions, 1)

		// Verify checkpoint contains error
		loadedState, err := LoadState(executions[0])
		assert.NoError(t, err)
		assert.Equal(t, ExecutionStatusFailed, loadedState.Status)
		assert.NotNil(t, loadedState.Result)
		assert.False(t, loadedState.Result.Success)
	})

	t.Run("does not save checkpoint when disabled", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger:               nil,
			EnableAutoCheckpoint: false,
		}

		engine := New(cfg, selector)

		// Verify auto checkpoint is disabled
		assert.False(t, engine.autoCheckpointEnabled)
	})
}

func TestEngine_ManualCheckpoint(t *testing.T) {
	t.Run("creates manual checkpoint successfully", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger: nil,
		}

		engine := New(cfg, selector)

		task := NewTask("test task").WithType("feature")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
			ChangeID:    "change-1",
			TaskID:      "task-1",
		}
		task.Context = ctx

		state, err := engine.CreateManualCheckpoint(nil, task, ExecutionStatusRunning)

		assert.NoError(t, err)
		assert.NotNil(t, state)
		assert.Equal(t, ExecutionStatusRunning, state.Status)
		assert.NotEmpty(t, state.ID)

		// Verify checkpoint file exists
		filename := filepath.Join(testDir, state.ID+".json")
		_, err = os.Stat(filename)
		assert.NoError(t, err)
	})

	t.Run("creates manual checkpoint with custom status", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger: nil,
		}

		engine := New(cfg, selector)

		task := NewTask("test task").WithAgent(domain.AgentRoleDeveloper)
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state, err := engine.CreateManualCheckpoint(nil, task, ExecutionStatusInterrupted)

		assert.NoError(t, err)
		assert.Equal(t, ExecutionStatusInterrupted, state.Status)
		assert.Equal(t, domain.AgentRoleDeveloper, state.Agent)
	})

	t.Run("creates manual checkpoint without status", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger: nil,
		}

		engine := New(cfg, selector)

		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state, err := engine.CreateManualCheckpoint(nil, task, "")

		assert.NoError(t, err)
		assert.Equal(t, ExecutionStatusPending, state.Status)
	})
}

func TestEngine_CleanupOldCheckpoints(t *testing.T) {
	t.Run("removes old checkpoints", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger:               nil,
			EnableAutoCheckpoint: true,
			MaxCheckpoints:       10,
			CheckpointAgeLimit:   1 * time.Hour,
		}

		engine := New(cfg, selector)

		// Create old checkpoints
		for i := 0; i < 3; i++ {
			task := NewTask("old task")
			ctx := &TaskContext{
				ProjectPath: "/test/path",
				ChangeID:    "change-old",
				TaskID:      "task-old",
			}
			task.Context = ctx

			state := NewExecutionState(task)
			state.Start()

			// Manually set timestamp to make it old and recompute checksum
			state.UpdatedAt = time.Now().Add(-2 * time.Hour)
			state.Checksum = state.computeChecksum()

			SaveState(state)
		}

		// Create new checkpoints
		for i := 0; i < 2; i++ {
			task := NewTask("new task")
			ctx := &TaskContext{
				ProjectPath: "/test/path",
				ChangeID:    "change-new",
				TaskID:      "task-new",
			}
			task.Context = ctx

			state := NewExecutionState(task)
			state.Start()

			SaveState(state)
		}

		// Run cleanup
		err := engine.CleanupOldCheckpoints()

		assert.NoError(t, err)

		// Verify old checkpoints were removed
		executions, err := ListExecutions()
		assert.NoError(t, err)
		assert.Len(t, executions, 2)
	})

	t.Run("removes completed and failed old checkpoints", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger:               nil,
			EnableAutoCheckpoint: true,
			MaxCheckpoints:       10,
			CheckpointAgeLimit:   1 * time.Hour,
		}

		engine := New(cfg, selector)

		// Create an old completed checkpoint
		task := NewTask("completed task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Start()
		state.Complete(&Result{Success: true})
		state.UpdatedAt = time.Now().Add(-2 * time.Hour)
		state.Checksum = state.computeChecksum()

		SaveState(state)

		// Run cleanup
		err := engine.CleanupOldCheckpoints()

		assert.NoError(t, err)

		// Verify checkpoint was removed
		executions, err := ListExecutions()
		assert.NoError(t, err)
		assert.Len(t, executions, 0)
	})

	t.Run("handles empty checkpoint directory", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger: nil,
		}

		engine := New(cfg, selector)

		err := engine.CleanupOldCheckpoints()

		assert.NoError(t, err)
	})
}

func TestEngine_GetCheckpoint(t *testing.T) {
	t.Run("retrieves existing checkpoint", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger: nil,
		}

		engine := New(cfg, selector)

		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		createdState, err := engine.CreateManualCheckpoint(nil, task, ExecutionStatusRunning)
		assert.NoError(t, err)

		retrievedState, err := engine.GetCheckpoint(createdState.ID)

		assert.NoError(t, err)
		assert.Equal(t, createdState.ID, retrievedState.ID)
		assert.Equal(t, ExecutionStatusRunning, retrievedState.Status)
	})

	t.Run("returns error for non-existent checkpoint", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger: nil,
		}

		engine := New(cfg, selector)

		_, err := engine.GetCheckpoint("nonexistent-id")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestEngine_ListCheckpoints(t *testing.T) {
	t.Run("lists all checkpoints", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger: nil,
		}

		engine := New(cfg, selector)

		// Create multiple checkpoints
		for i := 0; i < 3; i++ {
			task := NewTask("test task")
			ctx := &TaskContext{
				ProjectPath: "/test/path",
			}
			task.Context = ctx

			engine.CreateManualCheckpoint(nil, task, ExecutionStatusPending)
		}

		checkpoints, err := engine.ListCheckpoints()

		assert.NoError(t, err)
		assert.Len(t, checkpoints, 3)
	})

	t.Run("returns empty list when no checkpoints", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger: nil,
		}

		engine := New(cfg, selector)

		checkpoints, err := engine.ListCheckpoints()

		assert.NoError(t, err)
		assert.Empty(t, checkpoints)
	})
}

func TestEngine_DetermineExecutionConfig(t *testing.T) {
	t.Run("extracts config from task", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger: nil,
		}

		engine := New(cfg, selector)

		task := NewTask("test task").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("claude-3.5-sonnet").
			WithRuntime(domain.RuntimeClaudeCode).
			WithStrategy(domain.ExecutionStrategySingle).
			WithSkills("go-code", "go-test")

		budget := &BudgetLimit{
			MaxCost:   10.0,
			MaxTokens: 10000,
		}
		task.Budget = budget

		execCfg := engine.determineExecutionConfig(task)

		assert.Equal(t, domain.AgentRoleDeveloper, execCfg.Agent)
		assert.Equal(t, "claude-3.5-sonnet", execCfg.Model)
		assert.Equal(t, domain.RuntimeClaudeCode, execCfg.Runtime)
		assert.Equal(t, domain.ExecutionStrategySingle, execCfg.Strategy)
		assert.Equal(t, budget, execCfg.Budget)
		assert.Equal(t, []string{"go-code", "go-test"}, execCfg.Skills)
	})

	t.Run("handles empty task", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger: nil,
		}

		engine := New(cfg, selector)

		task := NewTask("test task")

		execCfg := engine.determineExecutionConfig(task)

		assert.Equal(t, domain.AgentRole(""), execCfg.Agent)
		assert.Equal(t, "", execCfg.Model)
		assert.Equal(t, domain.Runtime(""), execCfg.Runtime)
		assert.Empty(t, execCfg.Skills)
	})
}

func TestEngine_ResumeExecution(t *testing.T) {
	t.Run("checks can resume for interrupted state", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("interrupted task").WithType("bugfix")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
			ChangeID:    "change-1",
			TaskID:      "task-1",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Start()
		state.Interrupt()

		assert.True(t, state.CanResume())
	})

	t.Run("checks can resume for failed state", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("failed task").WithType("bugfix")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
			ChangeID:    "change-1",
			TaskID:      "task-1",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Start()
		state.Fail(assert.AnError)

		assert.True(t, state.CanResume())
	})

	t.Run("rejects resume from completed state", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("completed task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Start()
		state.Complete(&Result{Success: true})

		assert.False(t, state.CanResume())
	})

	t.Run("rejects resume from running state", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("running task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Start()

		assert.False(t, state.CanResume())
	})
}

func TestEngine_ValidateExecutionState(t *testing.T) {
	t.Run("validates correct state", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger: nil,
		}
		engine := New(cfg, selector)

		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		SaveState(state)

		result, err := engine.ValidateExecutionState(state.ID)
		assert.NoError(t, err)
		assert.True(t, result.Valid)
		assert.True(t, result.CanResume)
		assert.True(t, result.ChecksumValid)
		assert.True(t, result.VersionCompatible)
	})

	t.Run("handles missing state file", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger: nil,
		}
		engine := New(cfg, selector)

		result, err := engine.ValidateExecutionState("nonexistent-id")
		assert.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Message, "not found")
	})
}

func TestStorage_CorruptedState(t *testing.T) {
	t.Run("handles empty file", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		err := EnsureExecutionsDir()
		assert.NoError(t, err)

		filename := filepath.Join(testDir, "empty-state.json")
		err = os.WriteFile(filename, []byte{}, 0644)
		assert.NoError(t, err)

		_, err = LoadState("empty-state")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		err := EnsureExecutionsDir()
		assert.NoError(t, err)

		filename := filepath.Join(testDir, "invalid-json.json")
		err = os.WriteFile(filename, []byte("{invalid json}"), 0644)
		assert.NoError(t, err)

		_, err = LoadState("invalid-json")
		assert.Error(t, err)

		var corruptedErr *CorruptedStateError
		assert.ErrorAs(t, err, &corruptedErr)
		assert.False(t, corruptedErr.CanRecover)
		assert.Contains(t, corruptedErr.Reason, "invalid JSON")
	})

	t.Run("handles corrupted checksum", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		data, _ := state.ToJSON()

		var stateMap map[string]interface{}
		_ = json.Unmarshal(data, &stateMap)
		stateMap["checksum"] = "invalid-checksum"
		corruptedData, _ := json.Marshal(stateMap)

		err := EnsureExecutionsDir()
		assert.NoError(t, err)

		filename := filepath.Join(testDir, "corrupted-checksum.json")
		err = os.WriteFile(filename, corruptedData, 0644)
		assert.NoError(t, err)

		_, err = LoadState("corrupted-checksum")
		assert.Error(t, err)

		var corruptedErr *CorruptedStateError
		assert.ErrorAs(t, err, &corruptedErr)
		assert.False(t, corruptedErr.CanRecover)
		assert.Contains(t, corruptedErr.Reason, "checksum")
	})

	t.Run("handles missing required fields", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		err := EnsureExecutionsDir()
		assert.NoError(t, err)

		filename := filepath.Join(testDir, "missing-fields.json")
		err = os.WriteFile(filename, []byte(`{"id": "test", "status": "invalid"}`), 0644)
		assert.NoError(t, err)

		_, err = LoadState("missing-fields")
		assert.Error(t, err)

		var corruptedErr *CorruptedStateError
		assert.ErrorAs(t, err, &corruptedErr)
		assert.Contains(t, corruptedErr.Reason, "validation failed")
	})
}

func TestState_Validation(t *testing.T) {
	t.Run("validates complete state", func(t *testing.T) {
		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Start()

		err := state.Validate()
		assert.NoError(t, err)
	})

	t.Run("rejects state with missing ID", func(t *testing.T) {
		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.ID = ""

		err := state.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID is required")
	})

	t.Run("rejects state with nil task", func(t *testing.T) {
		state := &ExecutionState{
			ID:     "test-id",
			Status: ExecutionStatusPending,
			Task:   nil,
		}

		err := state.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task is required")
	})

	t.Run("rejects state with empty task description", func(t *testing.T) {
		task := NewTask("")
		state := NewExecutionState(task)

		err := state.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task description is required")
	})

	t.Run("rejects state with invalid status", func(t *testing.T) {
		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Status = "invalid-status"

		err := state.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})

	t.Run("rejects running state without started_at", func(t *testing.T) {
		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Status = ExecutionStatusRunning
		state.StartedAt = time.Time{}
		state.Checksum = state.computeChecksum()

		err := state.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "started_at timestamp")
	})

	t.Run("rejects completed state without completed_at", func(t *testing.T) {
		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Status = ExecutionStatusCompleted
		state.StartedAt = time.Now()
		state.CompletedAt = time.Time{}
		state.Checksum = state.computeChecksum()

		err := state.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "completed_at timestamp")
	})

	t.Run("checks checksum validation", func(t *testing.T) {
		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Checksum = "invalid-checksum"

		err := state.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "checksum validation")
	})
}

func TestState_CanResume(t *testing.T) {
	t.Run("returns true for interrupted state", func(t *testing.T) {
		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Start()
		state.Interrupt()

		assert.True(t, state.CanResume())
	})

	t.Run("returns true for failed state", func(t *testing.T) {
		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Start()
		state.Fail(assert.AnError)

		assert.True(t, state.CanResume())
	})

	t.Run("returns true for pending state", func(t *testing.T) {
		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)

		assert.True(t, state.CanResume())
	})

	t.Run("returns false for completed state", func(t *testing.T) {
		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Start()
		state.Complete(&Result{Success: true})

		assert.False(t, state.CanResume())
	})

	t.Run("returns false for running state", func(t *testing.T) {
		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		state.Start()

		assert.False(t, state.CanResume())
	})
}

func TestEngine_DeleteCorruptedState(t *testing.T) {
	t.Run("deletes corrupted state file", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger: nil,
		}
		engine := New(cfg, selector)

		task := NewTask("test task")
		ctx := &TaskContext{
			ProjectPath: "/test/path",
		}
		task.Context = ctx

		state := NewExecutionState(task)
		SaveState(state)

		err := engine.DeleteCorruptedState(state.ID)
		assert.NoError(t, err)

		_, err = LoadState(state.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("handles deleting non-existent state", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		cfg := Config{
			Logger: nil,
		}
		engine := New(cfg, selector)

		err := engine.DeleteCorruptedState("nonexistent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
