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

type mockLongRunningRunner struct {
	*CLIRunner
	interruptChan chan struct{}
	executed      bool
	interrupted   bool
	mu            sync.Mutex
}

func newMockLongRunningRunner(logger interface{}) *mockLongRunningRunner {
	return &mockLongRunningRunner{
		CLIRunner:     &CLIRunner{},
		interruptChan: make(chan struct{}),
	}
}

func (r *mockLongRunningRunner) Execute(ctx context.Context, req *Request) (*Result, error) {
	r.mu.Lock()
	r.executed = true
	r.mu.Unlock()

	select {
	case <-time.After(5 * time.Second):
		return &Result{
			Success:  true,
			Output:   "Long running task completed",
			Duration: 5 * time.Second,
		}, nil
	case <-r.interruptChan:
		r.mu.Lock()
		r.interrupted = true
		r.mu.Unlock()
		return nil, fmt.Errorf("execution interrupted")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (r *mockLongRunningRunner) Interrupt(ctx context.Context) error {
	close(r.interruptChan)
	return nil
}

func (r *mockLongRunningRunner) WasExecuted() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.executed
}

func (r *mockLongRunningRunner) WasInterrupted() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.interrupted
}

func TestIntegration_InterruptLongRunningExecution(t *testing.T) {
	originalExecDir := execDirPath
	defer func() { execDirPath = originalExecDir }()

	testDir := filepath.Join(os.TempDir(), "go-ent-test-interrupt-long-"+strings.ReplaceAll(t.Name(), "/", "-"))
	defer os.RemoveAll(testDir)

	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	execDirPath = testDir

	selector := agent.NewSelector(agent.Config{}, nil)
	engine := New(Config{
		IsMCPMode:              true,
		EnableAutoCheckpoint:   true,
		EnableSummarization:    true,
		SummarizationThreshold: DefaultSummarizationThreshold(),
		Logger:                 nil,
	}, selector)

	mockRunner := newMockLongRunningRunner(nil)
	engine.RegisterRunner(mockRunner)

	task := NewTask("Long running task that should be interrupted").
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
	require.NoError(t, state.Interrupt())

	require.NoError(t, SaveState(state))

	ctx := context.Background()

	resumeResult, err := engine.ResumeExecution(ctx, state.ID)
	assert.NoError(t, err, "resume should succeed")
	assert.NotNil(t, resumeResult, "resume result should not be nil")

	loadedState, err := LoadState(state.ID)
	require.NoError(t, err)

	assert.True(t, loadedState.IsCompleted(), "state should be completed after resume")
	assert.False(t, loadedState.IsInterrupted(), "state should not be interrupted")
	assert.Equal(t, ExecutionStatusCompleted, loadedState.Status, "status should be completed")
	assert.False(t, loadedState.CompletedAt.IsZero(), "completed_at should be set")
	assert.NotNil(t, loadedState.Result, "result should be set")
}

func TestIntegration_MultipleInterruptResumeCycles(t *testing.T) {
	originalExecDir := execDirPath
	defer func() { execDirPath = originalExecDir }()

	testDir := filepath.Join(os.TempDir(), "go-ent-test-multiple-cycles-"+strings.ReplaceAll(t.Name(), "/", "-"))
	defer os.RemoveAll(testDir)

	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	execDirPath = testDir

	selector := agent.NewSelector(agent.Config{}, nil)
	engine := New(Config{
		IsMCPMode:              true,
		EnableAutoCheckpoint:   true,
		EnableSummarization:    true,
		SummarizationThreshold: DefaultSummarizationThreshold(),
		Logger:                 nil,
	}, selector)

	cycleCount := 3

	for i := 0; i < cycleCount; i++ {
		t.Run(fmt.Sprintf("cycle_%d_full_workflow", i), func(t *testing.T) {
			task := NewTask(fmt.Sprintf("Task for cycle %d", i)).
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

			interruptedState, err := LoadState(state.ID)
			require.NoError(t, err)
			assert.True(t, interruptedState.IsInterrupted(), "state should be interrupted in cycle %d", i)

			_, err = engine.ResumeExecution(ctx, state.ID)
			assert.NoError(t, err, "resume should succeed in cycle %d", i)

			finalState, err := LoadState(state.ID)
			require.NoError(t, err)
			assert.True(t, finalState.IsCompleted(), "state should be completed after cycle %d", i)
		})
	}
}

func TestIntegration_ResumeAfterProcessRestart(t *testing.T) {
	originalExecDir := execDirPath
	defer func() { execDirPath = originalExecDir }()

	testDir := filepath.Join(os.TempDir(), "go-ent-test-process-restart-"+strings.ReplaceAll(t.Name(), "/", "-"))
	defer os.RemoveAll(testDir)

	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	execDirPath = testDir

	selector1 := agent.NewSelector(agent.Config{}, nil)
	engine1 := New(Config{
		IsMCPMode:              true,
		EnableAutoCheckpoint:   true,
		EnableSummarization:    true,
		SummarizationThreshold: DefaultSummarizationThreshold(),
		Logger:                 nil,
	}, selector1)

	taskContext := &TaskContext{
		ProjectPath: "/test/project",
		Files:       []string{"file1.go", "file2.go"},
		ChangeID:    "change-123",
		TaskID:      "task-456",
	}

	task := NewTask("Task that survives process restart").
		WithType("test").
		WithAgent(domain.AgentRoleDeveloper).
		WithModel("haiku").
		WithRuntime(domain.RuntimeCLI).
		WithStrategy(domain.ExecutionStrategySingle)
	task.Context = taskContext

	state := NewExecutionState(task)
	cfg := engine1.determineExecutionConfig(task)
	state.WithConfig(cfg)
	state.Runtime = domain.RuntimeCLI
	require.NoError(t, state.Start())
	require.NoError(t, state.Interrupt())

	state.SetMetadata("cycle", "1")
	state.SetMetadata("process_id", "1")
	state.SetMetadata("custom_key", "custom_value")

	require.NoError(t, SaveState(state))

	executionID := state.ID

	_ = engine1
	_ = selector1

	time.Sleep(50 * time.Millisecond)

	execDirPath = testDir

	selector2 := agent.NewSelector(agent.Config{}, nil)
	engine2 := New(Config{
		IsMCPMode:              true,
		EnableAutoCheckpoint:   true,
		EnableSummarization:    true,
		SummarizationThreshold: DefaultSummarizationThreshold(),
		Logger:                 nil,
	}, selector2)

	ctx := context.Background()

	result, err := engine2.ResumeExecution(ctx, executionID)
	assert.NoError(t, err, "resume should succeed after process restart")
	assert.NotNil(t, result)

	loadedState, err := LoadState(executionID)
	require.NoError(t, err)

	assert.True(t, loadedState.IsCompleted(), "state should be completed")
	assert.Equal(t, ExecutionStatusCompleted, loadedState.Status)

	assert.NotNil(t, loadedState.Context)
	assert.Equal(t, "/test/project", loadedState.Context.ProjectPath)
	assert.Equal(t, "change-123", loadedState.Context.ChangeID)
	assert.Equal(t, "task-456", loadedState.Context.TaskID)
	assert.Equal(t, 2, len(loadedState.Context.Files))

	cycle, ok := loadedState.GetMetadata("cycle")
	assert.True(t, ok, "cycle metadata should be preserved")
	assert.Equal(t, "1", cycle)

	processID, ok := loadedState.GetMetadata("process_id")
	assert.True(t, ok, "process_id metadata should be preserved")
	assert.Equal(t, "1", processID)
}

func TestIntegration_InterruptAcrossRuntimes(t *testing.T) {
	originalExecDir := execDirPath
	defer func() { execDirPath = originalExecDir }()

	testDir := filepath.Join(os.TempDir(), "go-ent-test-runtimes-"+strings.ReplaceAll(t.Name(), "/", "-"))
	defer os.RemoveAll(testDir)

	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	execDirPath = testDir

	runtimes := []domain.Runtime{
		domain.RuntimeCLI,
		domain.RuntimeClaudeCode,
		domain.RuntimeOpenCode,
	}

	for _, rt := range runtimes {
		t.Run(string(rt), func(t *testing.T) {
			selector := agent.NewSelector(agent.Config{}, nil)
			engine := New(Config{
				IsMCPMode:              true,
				EnableAutoCheckpoint:   true,
				EnableSummarization:    true,
				SummarizationThreshold: DefaultSummarizationThreshold(),
				Logger:                 nil,
			}, selector)

			task := NewTask(fmt.Sprintf("Task for %s runtime", rt)).
				WithType("test").
				WithAgent(domain.AgentRoleDeveloper).
				WithModel("haiku").
				WithRuntime(rt).
				WithStrategy(domain.ExecutionStrategySingle)

			state := NewExecutionState(task)
			cfg := engine.determineExecutionConfig(task)
			state.WithConfig(cfg)
			state.Runtime = rt
			require.NoError(t, state.Start())

			require.NoError(t, SaveState(state))

			ctx := context.Background()
			err := engine.Interrupt(ctx, state.ID)
			assert.NoError(t, err)

			loadedState, err := LoadState(state.ID)
			require.NoError(t, err)

			assert.True(t, loadedState.IsInterrupted())
			assert.Equal(t, rt, loadedState.Runtime)
		})
	}
}

func TestIntegration_StateConsistencyAcrossInterrupts(t *testing.T) {
	originalExecDir := execDirPath
	defer func() { execDirPath = originalExecDir }()

	testDir := filepath.Join(os.TempDir(), "go-ent-test-consistency-"+strings.ReplaceAll(t.Name(), "/", "-"))
	defer os.RemoveAll(testDir)

	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	execDirPath = testDir

	selector := agent.NewSelector(agent.Config{}, nil)
	engine := New(Config{
		IsMCPMode:              true,
		EnableAutoCheckpoint:   true,
		EnableSummarization:    true,
		SummarizationThreshold: DefaultSummarizationThreshold(),
		Logger:                 nil,
	}, selector)

	task := NewTask("Task for state consistency test").
		WithType("test").
		WithAgent(domain.AgentRoleDeveloper).
		WithModel("haiku").
		WithRuntime(domain.RuntimeCLI).
		WithStrategy(domain.ExecutionStrategySingle)

	task.Metadata = map[string]interface{}{
		"test_key_1": "test_value_1",
		"test_key_2": 42,
		"test_key_3": true,
	}

	state := NewExecutionState(task)
	cfg := engine.determineExecutionConfig(task)
	state.WithConfig(cfg)
	state.Runtime = domain.RuntimeCLI
	initialID := state.ID
	require.NoError(t, state.Start())

	require.NoError(t, SaveState(state))

	ctx := context.Background()
	err = engine.Interrupt(ctx, state.ID)
	assert.NoError(t, err)

	stateAfterInterrupt, err := LoadState(state.ID)
	require.NoError(t, err)

	assert.Equal(t, initialID, stateAfterInterrupt.ID)
	assert.Equal(t, state.Task.Description, stateAfterInterrupt.Task.Description)
	assert.Equal(t, state.Agent, stateAfterInterrupt.Agent)
	assert.Equal(t, state.Model, stateAfterInterrupt.Model)
	assert.Equal(t, state.Runtime, stateAfterInterrupt.Runtime)
	assert.Equal(t, state.Strategy, stateAfterInterrupt.Strategy)

	_, err = engine.ResumeExecution(ctx, state.ID)
	assert.NoError(t, err)

	finalState, err := LoadState(state.ID)
	require.NoError(t, err)

	assert.Equal(t, initialID, finalState.ID)
	assert.Equal(t, ExecutionStatusCompleted, finalState.Status)
}

func TestIntegration_InterruptWithLargeContext(t *testing.T) {
	originalExecDir := execDirPath
	defer func() { execDirPath = originalExecDir }()

	testDir := filepath.Join(os.TempDir(), "go-ent-test-large-context-"+strings.ReplaceAll(t.Name(), "/", "-"))
	defer os.RemoveAll(testDir)

	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	execDirPath = testDir

	selector := agent.NewSelector(agent.Config{}, nil)
	engine := New(Config{
		IsMCPMode:              true,
		EnableAutoCheckpoint:   true,
		EnableSummarization:    true,
		SummarizationThreshold: DefaultSummarizationThreshold(),
		Logger:                 nil,
	}, selector)

	largeFiles := make([]string, 100)
	for i := 0; i < 100; i++ {
		largeFiles[i] = fmt.Sprintf("/path/to/file%d.go", i)
	}

	taskContext := &TaskContext{
		ProjectPath: "/test/large-project",
		Files:       largeFiles,
		ChangeID:    "change-large",
		TaskID:      "task-large",
	}

	task := NewTask("Task with large context").
		WithType("test").
		WithAgent(domain.AgentRoleDeveloper).
		WithModel("haiku").
		WithRuntime(domain.RuntimeCLI).
		WithStrategy(domain.ExecutionStrategySingle)
	task.Context = taskContext

	state := NewExecutionState(task)
	cfg := engine.determineExecutionConfig(task)
	state.WithConfig(cfg)
	state.Runtime = domain.RuntimeCLI
	require.NoError(t, state.Start())

	require.NoError(t, SaveState(state))

	ctx := context.Background()
	err = engine.Interrupt(ctx, state.ID)
	assert.NoError(t, err)

	loadedState, err := LoadState(state.ID)
	require.NoError(t, err)

	assert.True(t, loadedState.IsInterrupted())
	assert.NotNil(t, loadedState.Context)
	assert.Equal(t, 100, len(loadedState.Context.Files))
	assert.Equal(t, "/test/large-project", loadedState.Context.ProjectPath)
}

func TestIntegration_InterruptCheckpointTimestamps(t *testing.T) {
	originalExecDir := execDirPath
	defer func() { execDirPath = originalExecDir }()

	testDir := filepath.Join(os.TempDir(), "go-ent-test-timestamps-"+strings.ReplaceAll(t.Name(), "/", "-"))
	defer os.RemoveAll(testDir)

	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	execDirPath = testDir

	selector := agent.NewSelector(agent.Config{}, nil)
	engine := New(Config{
		IsMCPMode:              true,
		EnableAutoCheckpoint:   true,
		EnableSummarization:    true,
		SummarizationThreshold: DefaultSummarizationThreshold(),
		Logger:                 nil,
	}, selector)

	task := NewTask("Task for timestamp verification").
		WithType("test").
		WithAgent(domain.AgentRoleDeveloper).
		WithModel("haiku").
		WithRuntime(domain.RuntimeCLI).
		WithStrategy(domain.ExecutionStrategySingle)

	state := NewExecutionState(task)
	cfg := engine.determineExecutionConfig(task)
	state.WithConfig(cfg)
	state.Runtime = domain.RuntimeCLI
	startedAt := time.Now()
	require.NoError(t, state.Start())

	require.NoError(t, SaveState(state))

	initialState, err := LoadState(state.ID)
	require.NoError(t, err)

	assert.True(t, initialState.StartedAt.After(startedAt.Add(-1*time.Second)))
	assert.True(t, initialState.StartedAt.Before(startedAt.Add(1*time.Second)))

	time.Sleep(100 * time.Millisecond)

	ctx := context.Background()
	err = engine.Interrupt(ctx, state.ID)
	assert.NoError(t, err)

	interruptedState, err := LoadState(state.ID)
	require.NoError(t, err)

	assert.True(t, interruptedState.UpdatedAt.After(initialState.UpdatedAt))

	_, err = engine.ResumeExecution(ctx, state.ID)
	assert.NoError(t, err)

	finalState, err := LoadState(state.ID)
	require.NoError(t, err)

	assert.True(t, finalState.UpdatedAt.After(interruptedState.UpdatedAt))
	assert.False(t, finalState.CompletedAt.IsZero())
	assert.True(t, finalState.CompletedAt.After(interruptedState.UpdatedAt))
}
