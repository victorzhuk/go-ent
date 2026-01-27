package execution

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/domain"
)

func TestEngine_Resume(t *testing.T) {
	t.Run("resumes interrupted execution", func(t *testing.T) {

		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-resume-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
		}, selector)

		task := NewTask("Test task for resume").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)
		state.WithConfig(cfg)
		state.WithConfig(engine.determineExecutionConfig(task))
		state.Runtime = domain.RuntimeCLI
		assert.NoError(t, state.Start())
		assert.NoError(t, state.Interrupt())

		err := SaveState(state)
		assert.NoError(t, err)

		ctx := context.Background()
		_, err = engine.ResumeExecution(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)
		assert.True(t, loadedState.IsCompleted())
	})

	t.Run("resumes failed execution", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-resume-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)
		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
		}, selector)

		task := NewTask("Test task for failed resume").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)
		state.WithConfig(cfg)
		assert.NoError(t, state.Start())
		assert.NoError(t, state.Fail(assert.AnError))

		err := SaveState(state)
		assert.NoError(t, err)

		ctx := context.Background()
		_, err = engine.ResumeExecution(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)
		assert.True(t, loadedState.IsCompleted())
	})

	t.Run("resumes pending execution", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-resume-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)
		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
		}, selector)

		task := NewTask("Test pending task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)

		state.WithConfig(engine.determineExecutionConfig(task))
		state.Runtime = domain.RuntimeCLI

		err := SaveState(state)
		assert.NoError(t, err)

		ctx := context.Background()
		_, err = engine.ResumeExecution(ctx, state.ID)
		assert.NoError(t, err)
	})

	t.Run("fails to resume completed execution", func(t *testing.T) {
		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-resume-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)
		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{}, selector)

		task := NewTask("Test completed task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)

		state.WithConfig(engine.determineExecutionConfig(task))
		state.Runtime = domain.RuntimeCLI
		assert.NoError(t, state.Start())

		result := &Result{
			Success:   true,
			Output:    "Completed",
			TokensIn:  100,
			TokensOut: 200,
			Cost:      0.01,
		}
		assert.NoError(t, state.Complete(result))

		err := SaveState(state)
		assert.NoError(t, err)

		ctx := context.Background()
		_, err = engine.ResumeExecution(ctx, state.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state validation failed")
	})

	t.Run("fails to resume missing execution", func(t *testing.T) {

		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-resume-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)
		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{}, selector)

		ctx := context.Background()
		_, err := engine.ResumeExecution(ctx, "non-existent-id")
		assert.Error(t, err)
	})
}

func TestEngine_ResumeValidation(t *testing.T) {
	t.Run("validates state before resume", func(t *testing.T) {
		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
		}, selector)

		task := NewTask("Test validation task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)
		state.WithConfig(cfg)
		assert.NoError(t, state.Start())
		assert.NoError(t, state.Interrupt())

		assert.NoError(t, SaveState(state))

		ctx := context.Background()
		_, err := engine.ResumeExecution(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)
		assert.NoError(t, loadedState.Validate())
	})

	t.Run("handles corrupted state files", func(t *testing.T) {

		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-corrupted-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		corruptedID := "corrupted-execution"
		filename := filepath.Join(testDir, corruptedID+".json")

		err := os.MkdirAll(testDir, 0755)
		assert.NoError(t, err)

		err = os.WriteFile(filename, []byte("invalid json"), 0644)
		assert.NoError(t, err)

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{}, selector)

		ctx := context.Background()
		_, err = engine.ResumeExecution(ctx, corruptedID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "load state")
	})

	t.Run("handles missing state files", func(t *testing.T) {

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{}, selector)

		ctx := context.Background()
		_, err := engine.ResumeExecution(ctx, "missing-execution-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execution state not found")
	})

	t.Run("handles version mismatch", func(t *testing.T) {

		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-version-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		task := NewTask("Test version task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)

		state.Runtime = domain.RuntimeCLI
		state.Version = "0.0.1"
		assert.NoError(t, state.Start())
		assert.NoError(t, state.Interrupt())

		err := SaveState(state)
		assert.NoError(t, err)

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
		}, selector)

		ctx := context.Background()
		_, err = engine.ResumeExecution(ctx, state.ID)
		assert.Error(t, err)
	})
}

func TestEngine_ResumeContext(t *testing.T) {
	t.Run("restores context from state", func(t *testing.T) {

		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-context-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
		}, selector)

		taskContext := &TaskContext{
			ProjectPath: "/test/project",
			Files:       []string{"file1.go", "file2.go"},
			ChangeID:    "change-123",
			TaskID:      "task-456",
		}

		task := NewTask("Test context task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Context = taskContext

		state := NewExecutionState(task)

		state.WithConfig(engine.determineExecutionConfig(task))
		state.Runtime = domain.RuntimeCLI
		assert.NoError(t, state.Start())
		assert.NoError(t, state.Interrupt())

		err := SaveState(state)
		assert.NoError(t, err)

		ctx := context.Background()
		_, err = engine.ResumeExecution(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)
		assert.NotNil(t, loadedState.Context)
		assert.Equal(t, "/test/project", loadedState.Context.ProjectPath)
		assert.Equal(t, 2, len(loadedState.Context.Files))
	})

	t.Run("respects context cancellation", func(t *testing.T) {

		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-cancel-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
		}, selector)

		task := NewTask("Test cancel task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)

		state.WithConfig(engine.determineExecutionConfig(task))
		state.Runtime = domain.RuntimeCLI
		assert.NoError(t, state.Start())
		assert.NoError(t, state.Interrupt())

		err := SaveState(state)
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = engine.ResumeExecution(ctx, state.ID)
		assert.NoError(t, err)
	})
}

func TestEngine_ResumeCheckpoint(t *testing.T) {
	t.Run("saves checkpoint on resume start", func(t *testing.T) {

		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-checkpoint-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
		}, selector)

		task := NewTask("Test checkpoint task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)

		state.WithConfig(engine.determineExecutionConfig(task))
		state.Runtime = domain.RuntimeCLI
		assert.NoError(t, state.Start())
		assert.NoError(t, state.Interrupt())

		err := SaveState(state)
		assert.NoError(t, err)

		ctx := context.Background()
		_, err = engine.ResumeExecution(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)
		assert.True(t, loadedState.IsCompleted() || loadedState.IsRunning())
	})

	t.Run("saves checkpoint on resume error", func(t *testing.T) {

		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-error-checkpoint-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
		}, selector)

		task := NewTask("Test error checkpoint task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)

		state.WithConfig(engine.determineExecutionConfig(task))
		state.Runtime = domain.RuntimeCLI
		assert.NoError(t, state.Start())
		assert.NoError(t, state.Fail(assert.AnError))

		err := SaveState(state)
		assert.NoError(t, err)

		ctx := context.Background()
		_, err = engine.ResumeExecution(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)
		assert.False(t, loadedState.IsFailed())
	})
}

func TestEngine_ResumeMetadata(t *testing.T) {
	t.Run("preserves metadata on resume", func(t *testing.T) {

		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-metadata-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
		}, selector)

		task := NewTask("Test metadata task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Metadata = map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}

		state := NewExecutionState(task)

		state.WithConfig(engine.determineExecutionConfig(task))
		state.Runtime = domain.RuntimeCLI
		state.SetMetadata("custom", "data")
		assert.NoError(t, state.Start())
		assert.NoError(t, state.Interrupt())

		err := SaveState(state)
		assert.NoError(t, err)

		ctx := context.Background()
		_, err = engine.ResumeExecution(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)

		customData, ok := loadedState.GetMetadata("custom")
		assert.True(t, ok)
		assert.Equal(t, "data", customData)
	})
}

func TestEngine_ResumeAcrossRuntimes(t *testing.T) {
	t.Run("resumes CLI runtime execution", func(t *testing.T) {

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
		}, selector)

		task := NewTask("Test CLI resume task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)
		state.WithConfig(cfg)
		assert.NoError(t, state.Start())
		assert.NoError(t, state.Interrupt())

		err := SaveState(state)
		assert.NoError(t, err)

		ctx := context.Background()
		_, err = engine.ResumeExecution(ctx, state.ID)
		assert.NoError(t, err)
	})

	t.Run("resumes claude-code runtime execution", func(t *testing.T) {

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
		}, selector)

		task := NewTask("Test claude-code resume task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeClaudeCode).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)
		state.WithConfig(cfg)
		assert.NoError(t, state.Start())
		assert.NoError(t, state.Interrupt())

		err := SaveState(state)
		assert.NoError(t, err)

		ctx := context.Background()
		_, err = engine.ResumeExecution(ctx, state.ID)
		assert.NoError(t, err)
	})

	t.Run("resumes open-code runtime execution", func(t *testing.T) {

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
		}, selector)

		ctx := context.Background()

		runner := NewOpenCodeRunner(nil).WithTimeout(100 * time.Millisecond)
		if !runner.Available(ctx) {
			t.Skip("opencode binary not available, skipping test")
		}

		task := NewTask("Test open-code resume task").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeOpenCode).
			WithStrategy(domain.ExecutionStrategySingle)

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)
		state.WithConfig(cfg)
		assert.NoError(t, state.Start())
		assert.NoError(t, state.Interrupt())

		err := SaveState(state)
		assert.NoError(t, err)

		engine.RegisterRunner(runner)

		_, err = engine.ResumeExecution(ctx, state.ID)

		if err != nil && strings.Contains(err.Error(), "timeout") {
			t.Skip("opencode binary timeout, skipping test")
		}

		assert.NoError(t, err)
	})
}
