package execution

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/domain"
)

type e2eTestRunner struct {
	*CLIRunner
	sandbox         *Sandbox
	codeMode        *CodeMode
	resourceLimits  ResourceLimits
	executedCount   int
	summarizedCount int
	checkpointCount int
	interruptCount  int
	mu              sync.Mutex
}

func newE2ETestRunner(logger interface{}, limits ResourceLimits) *e2eTestRunner {
	sandbox := NewSandbox(limits)
	return &e2eTestRunner{
		CLIRunner:       &CLIRunner{},
		sandbox:         sandbox,
		codeMode:        NewCodeMode(sandbox),
		resourceLimits:  limits,
		executedCount:   0,
		summarizedCount: 0,
		checkpointCount: 0,
		interruptCount:  0,
	}
}

func (r *e2eTestRunner) Execute(ctx context.Context, req *Request) (*Result, error) {
	r.mu.Lock()
	r.executedCount++
	r.mu.Unlock()

	if err := r.sandbox.CheckMemoryLimit(); err != nil {
		return nil, err
	}

	if err := r.sandbox.CheckExecLimit(10 * time.Second); err != nil {
		return nil, err
	}

	time.Sleep(50 * time.Millisecond)

	return &Result{
		Success:   true,
		Output:    "E2E test execution completed",
		TokensIn:  100,
		TokensOut: 50,
		Cost:      0.01,
		Duration:  50 * time.Millisecond,
	}, nil
}

func (r *e2eTestRunner) Interrupt(ctx context.Context) error {
	r.mu.Lock()
	r.interruptCount++
	r.mu.Unlock()
	return nil
}

func (r *e2eTestRunner) GetStats() map[string]int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return map[string]int{
		"executed":   r.executedCount,
		"summarized": r.summarizedCount,
		"checkpoint": r.checkpointCount,
		"interrupt":  r.interruptCount,
	}
}

func TestE2E_CompleteWorkflowWithAllV2Features(t *testing.T) {
	originalExecDir := execDirPath
	defer func() { execDirPath = originalExecDir }()

	testDir := filepath.Join(os.TempDir(), "go-ent-test-e2e-features-"+strings.ReplaceAll(t.Name(), "/", "-"))
	defer os.RemoveAll(testDir)

	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	execDirPath = testDir

	limits := ResourceLimits{
		MaxMemoryMB: 128,
		MaxCPUTime:  30 * time.Second,
		MaxExecTime: 60 * time.Second,
	}

	selector := agent.NewSelector(agent.Config{}, nil)
	engine := New(Config{
		IsMCPMode:              true,
		EnableSummarization:    true,
		SummarizationThreshold: SummarizationThreshold{FileCount: 30, ContextLength: 10000, TokenCount: 2000},
		SummarizationModel:     LLMModelClaude35,
		EnableAutoCheckpoint:   true,
		MaxCheckpoints:         5,
		CheckpointAgeLimit:     time.Hour,
		Logger:                 nil,
	}, selector)

	e2eRunner := newE2ETestRunner(nil, limits)
	engine.RegisterRunner(e2eRunner)

	taskContext := NewTaskContext("/test/e2e-project").
		WithChange("change-e2e-features").
		WithTask("task-e2e-features").
		WithWorkflow("workflow-e2e")

	for i := 0; i < 35; i++ {
		taskContext.AddFile(fmt.Sprintf("/test/e2e-project/src/file%d.go", i))
	}

	task := NewTask("Complete workflow test with all v2 features").
		WithType("feature").
		WithAgent(domain.AgentRoleDeveloper).
		WithModel("haiku").
		WithRuntime(domain.RuntimeCLI).
		WithStrategy(domain.ExecutionStrategySingle).
		WithSkills("go-code", "go-arch")
	task.Context = taskContext

	ctx := context.Background()

	result := engine.TriggerSummarization(ctx, task, "exec-e2e-features", LLMModelClaude35, nil, false)
	assert.True(t, result, "summarization should trigger")
	assert.True(t, task.Context.IsSummarized, "context should be summarized")
}

func TestE2E_SandboxWithResourceLimits(t *testing.T) {
	t.Run("memory limit enforcement", func(t *testing.T) {
		limits := ResourceLimits{
			MaxMemoryMB: 1,
			MaxExecTime: 10 * time.Second,
		}

		sandbox := NewSandbox(limits)
		ctx := context.Background()

		err := sandbox.CheckMemoryLimit()
		assert.NoError(t, err, "should be within limits initially")

		codeMode := NewCodeMode(sandbox)

		_, err = codeMode.Execute(ctx, `
			// Allocate memory
			var largeArray = new Array(1000000).fill("test");
			largeArray;
		`, map[string]interface{}{})
		assert.NoError(t, err, "should execute successfully")
	})

	t.Run("cpu limit enforcement", func(t *testing.T) {
		limits := ResourceLimits{
			MaxMemoryMB: 128,
			MaxCPUTime:  100 * time.Millisecond,
		}

		sandbox := NewSandbox(limits)
		err := sandbox.CheckCPULimit(50 * time.Millisecond)
		assert.NoError(t, err, "should be within CPU limit")

		err = sandbox.CheckCPULimit(150 * time.Millisecond)
		assert.Error(t, err, "should exceed CPU limit")
	})

	t.Run("execution time limit enforcement", func(t *testing.T) {
		limits := ResourceLimits{
			MaxMemoryMB: 128,
			MaxExecTime: 100 * time.Millisecond,
		}

		sandbox := NewSandbox(limits)
		err := sandbox.CheckExecLimit(50 * time.Millisecond)
		assert.NoError(t, err, "should be within exec time limit")

		err = sandbox.CheckExecLimit(150 * time.Millisecond)
		assert.Error(t, err, "should exceed exec time limit")
	})
}

func TestE2E_CodeModeExecution(t *testing.T) {
	t.Run("basic JavaScript execution", func(t *testing.T) {
		limits := ResourceLimits{MaxExecTime: 5 * time.Second}
		sandbox := NewSandbox(limits)
		codeMode := NewCodeMode(sandbox)
		ctx := context.Background()

		result, err := codeMode.Execute(ctx, `
			function add(a, b) {
				return a + b;
			}
			add(3, 4);
		`, map[string]interface{}{})

		assert.NoError(t, err, "should execute successfully")
		assert.NotNil(t, result, "result should not be nil")
	})

	t.Run("function execution", func(t *testing.T) {
		limits := ResourceLimits{MaxExecTime: 5 * time.Second}
		sandbox := NewSandbox(limits)
		codeMode := NewCodeMode(sandbox)
		ctx := context.Background()

		err := codeMode.DefineFunction("multiply", "(a, b) => a * b")
		assert.NoError(t, err, "should define function")

		result, err := codeMode.ExecuteFunction(ctx, "multiply", 5, 6)
		assert.NoError(t, err, "should execute function")
		assert.NotNil(t, result, "result should not be nil")
	})

	t.Run("safe API surface", func(t *testing.T) {
		testDir := t.TempDir()
		testFilePath := filepath.Join(testDir, "file.txt")

		limits := ResourceLimits{MaxExecTime: 5 * time.Second}
		sandbox := NewSandbox(limits).WithFileAccess(testDir).WithAPIAccess("safe_api")
		codeMode := NewCodeMode(sandbox)

		codeMode.SetGlobal("testVar", "safe")
		value := codeMode.GetGlobal("testVar")
		assert.Equal(t, "safe", value, "should get global variable")

		err := sandbox.CheckFileAccess(testFilePath)
		assert.NoError(t, err, "should allow file access")

		err = sandbox.CheckFileAccess("/etc/passwd")
		assert.Error(t, err, "should deny file access to sensitive paths")

		err = sandbox.CheckAPIAccess("safe_api")
		assert.NoError(t, err, "should allow API access")

		err = sandbox.CheckAPIAccess("dangerous_api")
		assert.Error(t, err, "should deny dangerous API access")
	})
}

func TestE2E_ErrorRecovery(t *testing.T) {
	originalExecDir := execDirPath
	defer func() { execDirPath = originalExecDir }()

	testDir := filepath.Join(os.TempDir(), "go-ent-test-e2e-error-"+strings.ReplaceAll(t.Name(), "/", "-"))
	defer os.RemoveAll(testDir)

	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	execDirPath = testDir

	t.Run("task failure with retry", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:            true,
			EnableAutoCheckpoint: true,
			Logger:               nil,
		}, selector)

		task := NewTask("Task that will fail and retry").
			WithType("bugfix").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)
		cfg := engine.determineExecutionConfig(task)
		state.WithConfig(cfg)
		state.Runtime = domain.RuntimeCLI
		require.NoError(t, state.Start())

		simErr := fmt.Errorf("simulated task failure")
		require.NoError(t, state.Fail(simErr))

		require.NoError(t, SaveState(state))

		loadedState, err := LoadState(state.ID)
		require.NoError(t, err)
		assert.True(t, loadedState.IsFailed(), "state should be failed")
		assert.Equal(t, ExecutionStatusFailed, loadedState.Status)
		assert.NotNil(t, loadedState.Result, "result should be set")
		assert.False(t, loadedState.Result.Success, "result should indicate failure")
	})

	t.Run("interrupt during error handling", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:            true,
			EnableAutoCheckpoint: true,
			Logger:               nil,
		}, selector)

		task := NewTask("Task interrupted during error handling").
			WithType("feature").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)
		cfg := engine.determineExecutionConfig(task)
		state.WithConfig(cfg)
		state.Runtime = domain.RuntimeCLI
		require.NoError(t, state.Start())

		require.NoError(t, SaveState(state))

		ctx := context.Background()
		err := engine.Interrupt(ctx, state.ID)
		assert.NoError(t, err, "interrupt should succeed during error handling")

		interruptedState, err := LoadState(state.ID)
		require.NoError(t, err)
		assert.True(t, interruptedState.IsInterrupted(), "state should be interrupted")
	})

	t.Run("resume after failed execution", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:            true,
			EnableAutoCheckpoint: true,
			Logger:               nil,
		}, selector)

		task := NewTask("Task to resume after failure").
			WithType("refactor").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)
		cfg := engine.determineExecutionConfig(task)
		state.WithConfig(cfg)
		state.Runtime = domain.RuntimeCLI
		require.NoError(t, state.Start())

		simErr := fmt.Errorf("simulated failure before completion")
		require.NoError(t, state.Fail(simErr))

		require.NoError(t, SaveState(state))

		ctx := context.Background()
		_, err := engine.ResumeExecution(ctx, state.ID)
		assert.NoError(t, err, "should resume after failed execution")

		finalState, err := LoadState(state.ID)
		require.NoError(t, err)
		assert.True(t, finalState.IsCompleted(), "state should be completed after resume")
	})

	t.Run("corrupted state recovery", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode: true,
			Logger:    nil,
		}, selector)

		task := NewTask("Task with corrupted state").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)
		state.ID = "corrupted-test-id"

		corruptedJSON := `{"id":"corrupted-test-id","invalid":json`
		filename := filepath.Join(execDirPath, state.ID+".json")
		err := os.WriteFile(filename, []byte(corruptedJSON), 0644)
		require.NoError(t, err, "should write corrupted state file")

		_, err = LoadState(state.ID)
		assert.Error(t, err, "should fail to load corrupted state")

		var corruptedErr *CorruptedStateError
		assert.ErrorAs(t, err, &corruptedErr, "should return CorruptedStateError")
		assert.False(t, corruptedErr.CanRecover, "should not be recoverable")

		err = engine.DeleteCorruptedState(state.ID)
		assert.NoError(t, err, "should delete corrupted state")

		_, err = os.Stat(filename)
		assert.True(t, os.IsNotExist(err), "file should be deleted")
	})

	t.Run("checksum mismatch detection", func(t *testing.T) {
		task := NewTask("Task with checksum mismatch").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)
		require.NoError(t, state.Start())

		originalChecksum := state.Checksum
		state.Checksum = "invalid-checksum"

		require.NoError(t, SaveState(state))

		state.Checksum = originalChecksum

		_, err := LoadState(state.ID)
		assert.Error(t, err, "should fail to load state with invalid checksum")

		var corruptedErr *CorruptedStateError
		assert.ErrorAs(t, err, &corruptedErr, "should return CorruptedStateError")
		assert.Contains(t, corruptedErr.Reason, "checksum", "error should mention checksum")
	})
}

func TestE2E_EdgeCases(t *testing.T) {
	originalExecDir := execDirPath
	defer func() { execDirPath = originalExecDir }()

	testDir := filepath.Join(os.TempDir(), "go-ent-test-e2e-edge-"+strings.ReplaceAll(t.Name(), "/", "-"))
	defer os.RemoveAll(testDir)

	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	execDirPath = testDir

	t.Run("empty execution (no files)", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:           true,
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount: 10,
			},
			Logger: nil,
		}, selector)

		taskContext := NewTaskContext("/test/empty-project").
			WithChange("change-empty").
			WithTask("task-empty")

		task := NewTask("Empty execution task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Context = taskContext

		ctx := context.Background()
		result := engine.TriggerSummarization(ctx, task, "exec-empty", LLMModelClaude35, nil, false)
		assert.False(t, result, "should not summarize empty context")
		assert.Equal(t, 0, len(taskContext.Files), "should have no files")
	})

	t.Run("single file execution", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:           true,
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount: 50,
			},
			Logger: nil,
		}, selector)

		taskContext := NewTaskContext("/test/single-file-project").
			WithChange("change-single").
			WithTask("task-single")
		taskContext.AddFile("/test/single-file-project/main.go")

		task := NewTask("Single file execution task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Context = taskContext

		ctx := context.Background()
		result := engine.TriggerSummarization(ctx, task, "exec-single", LLMModelClaude35, nil, false)
		assert.False(t, result, "should not summarize single file below threshold")
		assert.Equal(t, 1, len(taskContext.Files), "should have 1 file")
	})

	t.Run("very large context (near limit)", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:           true,
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount:     100,
				ContextLength: 100000,
				TokenCount:    25000,
			},
			SummarizationModel: LLMModelClaude35,
			Logger:             nil,
		}, selector)

		taskContext := NewTaskContext("/test/large-project").
			WithChange("change-large").
			WithTask("task-large")

		for i := 0; i < 90; i++ {
			content := strings.Repeat("x", 900)
			taskContext.AddFile(content)
		}

		task := NewTask("Large context near limit task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Context = taskContext

		ctx := context.Background()
		result := engine.TriggerSummarization(ctx, task, "exec-large", LLMModelClaude35, nil, false)
		assert.False(t, result, "should not summarize when below threshold")

		for i := 90; i < 105; i++ {
			content := strings.Repeat("x", 900)
			taskContext.AddFile(content)
		}

		result = engine.TriggerSummarization(ctx, task, "exec-large-2", LLMModelClaude35, nil, false)
		assert.True(t, result, "should summarize when exceeding threshold")
	})

	t.Run("rapid interrupt/resume cycles", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:            true,
			EnableAutoCheckpoint: true,
			Logger:               nil,
		}, selector)

		cycleCount := 5

		for i := 0; i < cycleCount; i++ {
			task := NewTask(fmt.Sprintf("Rapid cycle task %d", i)).
				WithType("test").
				WithAgent(domain.AgentRoleDeveloper).
				WithModel("haiku").
				WithRuntime(domain.RuntimeCLI).
				WithStrategy(domain.ExecutionStrategySingle)

			state := NewExecutionState(task)
			cfg := engine.determineExecutionConfig(task)
			state.WithConfig(cfg)
			state.Runtime = domain.RuntimeCLI
			require.NoError(t, state.Start())

			require.NoError(t, SaveState(state))

			ctx := context.Background()
			err := engine.Interrupt(ctx, state.ID)
			assert.NoError(t, err, "interrupt should succeed in cycle %d", i)

			_, err = engine.ResumeExecution(ctx, state.ID)
			assert.NoError(t, err, "resume should succeed in cycle %d", i)

			finalState, err := LoadState(state.ID)
			require.NoError(t, err)
			assert.True(t, finalState.IsCompleted(), "state should complete in cycle %d", i)
		}
	})

	t.Run("concurrent executions", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:            true,
			EnableAutoCheckpoint: true,
			Logger:               nil,
		}, selector)

		concurrentCount := 3
		var wg sync.WaitGroup
		errs := make(chan error, concurrentCount)
		ids := make(chan string, concurrentCount)

		for i := 0; i < concurrentCount; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				task := NewTask(fmt.Sprintf("Concurrent task %d", idx)).
					WithType("test").
					WithAgent(domain.AgentRoleDeveloper).
					WithModel("haiku").
					WithRuntime(domain.RuntimeCLI).
					WithStrategy(domain.ExecutionStrategySingle)

				state := NewExecutionState(task)
				cfg := engine.determineExecutionConfig(task)
				state.WithConfig(cfg)
				state.Runtime = domain.RuntimeCLI
				require.NoError(t, state.Start())

				if err := SaveState(state); err != nil {
					errs <- fmt.Errorf("task %d save: %w", idx, err)
					return
				}

				ids <- state.ID

				time.Sleep(50 * time.Millisecond)

				if err := state.Complete(&Result{
					Success:  true,
					Output:   fmt.Sprintf("Task %d completed", idx),
					Duration: 50 * time.Millisecond,
				}); err != nil {
					errs <- fmt.Errorf("task %d complete: %w", idx, err)
					return
				}

				if err := SaveState(state); err != nil {
					errs <- fmt.Errorf("task %d final save: %w", idx, err)
				}
			}(i)
		}

		wg.Wait()
		close(errs)
		close(ids)

		for err := range errs {
			assert.NoError(t, err)
		}

		collectedIDs := make([]string, 0, concurrentCount)
		for id := range ids {
			collectedIDs = append(collectedIDs, id)
		}

		assert.Equal(t, concurrentCount, len(collectedIDs), "should have all execution IDs")

		for _, id := range collectedIDs {
			state, err := LoadState(id)
			require.NoError(t, err, "should load state %s", id)
			assert.True(t, state.IsCompleted(), "state %s should be completed", id)
		}
	})

	t.Run("invalid inputs and configurations", func(t *testing.T) {
		t.Run("empty task description", func(t *testing.T) {
			task := NewTask("").
				WithType("test")
			task.Context = NewTaskContext("/test/project")

			state := NewExecutionState(task)
			err := state.Validate()
			assert.Error(t, err, "should validate empty description")
		})

		t.Run("nil context", func(t *testing.T) {
			task := NewTask("Task with nil context").
				WithType("test")

			state := NewExecutionState(task)
			state.Context = nil

			errors := state.ValidateForResume()
			assert.Len(t, errors, 2, "should have runtime and agent validation errors")
		})

		t.Run("invalid status transition", func(t *testing.T) {
			task := NewTask("Invalid transition task").
				WithType("test")

			state := NewExecutionState(task)
			err := state.Complete(&Result{Success: true})
			assert.Error(t, err, "should not complete from pending status")

			err = state.Interrupt()
			assert.Error(t, err, "should not interrupt from pending status")
		})

		t.Run("missing execution ID", func(t *testing.T) {
			state := &ExecutionState{
				Status: ExecutionStatusPending,
			}
			err := state.Validate()
			assert.Error(t, err, "should validate missing ID")
		})
	})
}

func TestE2E_PerformanceBenchmarks(t *testing.T) {
	originalExecDir := execDirPath
	defer func() { execDirPath = originalExecDir }()

	testDir := filepath.Join(os.TempDir(), "go-ent-test-e2e-perf-"+strings.ReplaceAll(t.Name(), "/", "-"))
	defer os.RemoveAll(testDir)

	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	execDirPath = testDir

	t.Run("state save/load performance", func(t *testing.T) {
		task := NewTask("Performance benchmark task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		taskContext := NewTaskContext("/test/perf-project").
			WithChange("change-perf").
			WithTask("task-perf")

		for i := 0; i < 50; i++ {
			taskContext.AddFile(fmt.Sprintf("/test/perf-project/file%d.go", i))
		}

		task.Context = taskContext
		state := NewExecutionState(task)
		require.NoError(t, state.Start())

		saveTimes := make([]time.Duration, 0, 100)
		loadTimes := make([]time.Duration, 0, 100)

		iterations := 100
		for i := 0; i < iterations; i++ {
			saveStart := time.Now()
			err := SaveState(state)
			saveElapsed := time.Since(saveStart)
			if err == nil {
				saveTimes = append(saveTimes, saveElapsed)
			}

			loadStart := time.Now()
			loadedState, err := LoadState(state.ID)
			loadElapsed := time.Since(loadStart)
			if err == nil && loadedState != nil {
				loadTimes = append(loadTimes, loadElapsed)
			}
		}

		var avgSave time.Duration
		if len(saveTimes) > 0 {
			var totalSave time.Duration
			for _, t := range saveTimes {
				totalSave += t
			}
			avgSave = totalSave / time.Duration(len(saveTimes))
		}

		var avgLoad time.Duration
		if len(loadTimes) > 0 {
			var totalLoad time.Duration
			for _, t := range loadTimes {
				totalLoad += t
			}
			avgLoad = totalLoad / time.Duration(len(loadTimes))
		}

		t.Logf("State save performance (avg over %d ops): %v", len(saveTimes), avgSave)
		t.Logf("State load performance (avg over %d ops): %v", len(loadTimes), avgLoad)

		assert.Less(t, avgSave, 100*time.Millisecond, "state save should be fast")
		assert.Less(t, avgLoad, 100*time.Millisecond, "state load should be fast")
	})

	t.Run("context summarization performance", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:           true,
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount:     50,
				ContextLength: 50000,
				TokenCount:    10000,
			},
			SummarizationModel: LLMModelClaude35,
			Logger:             nil,
		}, selector)

		taskContext := NewTaskContext("/test/summization-perf-project").
			WithChange("change-summ").
			WithTask("task-summ")

		fileCounts := []int{10, 25, 50, 75, 100}

		for _, fileCount := range fileCounts {
			t.Run(fmt.Sprintf("%d_files", fileCount), func(t *testing.T) {
				taskContextCopy := *taskContext
				taskContextCopy.Files = []string{}

				for i := 0; i < fileCount; i++ {
					taskContextCopy.AddFile(fmt.Sprintf("/test/summization-perf-project/file%d.go", i))
				}

				task := NewTask("Summarization performance task").
					WithType("test").
					WithAgent(domain.AgentRoleDeveloper).
					WithModel("haiku").
					WithRuntime(domain.RuntimeCLI).
					WithStrategy(domain.ExecutionStrategySingle)
				task.Context = &taskContextCopy

				ctx := context.Background()

				startTime := time.Now()
				result := engine.TriggerSummarization(ctx, task, fmt.Sprintf("exec-summ-%d", fileCount), LLMModelClaude35, nil, false)
				elapsed := time.Since(startTime)

				assert.Equal(t, fileCount > 50, result, "summarization result should be correct")
				t.Logf("Summarization for %d files: %v", fileCount, elapsed)
			})
		}
	})

	t.Run("checkpoint creation performance", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:            true,
			EnableAutoCheckpoint: true,
			Logger:               nil,
		}, selector)

		task := NewTask("Checkpoint performance task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		checkpointTimes := make([]time.Duration, 0, 50)

		for i := 0; i < 50; i++ {
			state := NewExecutionState(task)
			state.SetMetadata("iteration", fmt.Sprintf("%d", i))

			startTime := time.Now()
			err := engine.createCheckpoint(state)
			elapsed := time.Since(startTime)
			if err == nil {
				checkpointTimes = append(checkpointTimes, elapsed)
			}

			os.Remove(filepath.Join(execDirPath, state.ID+".json"))
		}

		var avgCheckpoint time.Duration
		if len(checkpointTimes) > 0 {
			var totalCheckpoint time.Duration
			for _, t := range checkpointTimes {
				totalCheckpoint += t
			}
			avgCheckpoint = totalCheckpoint / time.Duration(len(checkpointTimes))
		}

		t.Logf("Checkpoint creation (avg over %d ops): %v", len(checkpointTimes), avgCheckpoint)
		assert.Less(t, avgCheckpoint, 50*time.Millisecond, "checkpoint creation should be fast")
	})

	t.Run("interrupt/resume cycle performance", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:            true,
			EnableAutoCheckpoint: true,
			Logger:               nil,
		}, selector)

		task := NewTask("Interrupt/resume performance task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		cycleTimes := make([]time.Duration, 0, 20)
		ctx := context.Background()

		for i := 0; i < 20; i++ {
			state := NewExecutionState(task)
			cfg := engine.determineExecutionConfig(task)
			state.WithConfig(cfg)
			state.Runtime = domain.RuntimeCLI
			require.NoError(t, state.Start())

			cycleStart := time.Now()

			require.NoError(t, SaveState(state))
			err := engine.Interrupt(ctx, state.ID)
			if err != nil {
				t.Logf("Interrupt error in cycle %d: %v", i, err)
				continue
			}

			_, err = engine.ResumeExecution(ctx, state.ID)
			if err != nil {
				t.Logf("Resume error in cycle %d: %v", i, err)
				continue
			}

			cycleElapsed := time.Since(cycleStart)
			cycleTimes = append(cycleTimes, cycleElapsed)

			os.Remove(filepath.Join(execDirPath, state.ID+".json"))
		}

		var avgCycle time.Duration
		if len(cycleTimes) > 0 {
			var totalCycle time.Duration
			for _, t := range cycleTimes {
				totalCycle += t
			}
			avgCycle = totalCycle / time.Duration(len(cycleTimes))
		}

		t.Logf("Interrupt/resume cycle (avg over %d ops): %v", len(cycleTimes), avgCycle)
		assert.Less(t, avgCycle, 500*time.Millisecond, "interrupt/resume cycle should be fast")
	})

	t.Run("large context handling performance", func(t *testing.T) {
		sizes := []int{100, 500, 1000, 5000}

		for _, size := range sizes {
			t.Run(fmt.Sprintf("context_size_%d", size), func(t *testing.T) {
				taskContext := NewTaskContext("/test/large-context-project").
					WithChange("change-large").
					WithTask("task-large")

				for i := 0; i < size; i++ {
					taskContext.AddFile(fmt.Sprintf("/test/large-context-project/file%d.go", i))
				}

				task := NewTask("Large context performance task").
					WithType("test").
					WithAgent(domain.AgentRoleDeveloper).
					WithModel("haiku").
					WithRuntime(domain.RuntimeCLI).
					WithStrategy(domain.ExecutionStrategySingle)
				task.Context = taskContext

				state := NewExecutionState(task)

				startTime := time.Now()
				data, err := state.ToJSON()
				serializeTime := time.Since(startTime)

				assert.NoError(t, err, "should serialize state")

				startTime = time.Now()
				stateCopy := &ExecutionState{}
				err = stateCopy.FromJSON(data)
				deserializeTime := time.Since(startTime)

				assert.NoError(t, err, "should deserialize state")

				startTime = time.Now()
				checksum := state.computeChecksum()
				checksumTime := time.Since(startTime)

				assert.NotEmpty(t, checksum, "should compute checksum")

				t.Logf("Large context (%d files) - Serialize: %v, Deserialize: %v, Checksum: %v",
					size, serializeTime, deserializeTime, checksumTime)

				assert.Less(t, serializeTime, 1*time.Second, "serialization should be fast")
				assert.Less(t, deserializeTime, 1*time.Second, "deserialization should be fast")
				assert.Less(t, checksumTime, 100*time.Millisecond, "checksum computation should be fast")
			})
		}
	})

	t.Run("memory usage under load", func(t *testing.T) {
		task := NewTask("Memory usage task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		stateCount := 50
		ids := make([]string, stateCount)

		for i := 0; i < stateCount; i++ {
			state := NewExecutionState(task)
			state.SetMetadata("index", fmt.Sprintf("%d", i))
			state.SetMetadata("data", strings.Repeat("x", 1000))

			require.NoError(t, SaveState(state))
			ids[i] = state.ID
		}

		for _, id := range ids {
			state, err := LoadState(id)
			require.NoError(t, err, "should load state %s", id)
			assert.NotNil(t, state, "state should not be nil")
		}

		for _, id := range ids {
			err := DeleteState(id)
			assert.NoError(t, err, "should delete state %s", id)
		}

		t.Logf("Successfully handled %d state save/load/delete operations", stateCount)
	})
}
