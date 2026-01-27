package execution

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victorzhuk/go-ent/internal/domain"
)

func TestEngine_Interrupt(t *testing.T) {
	t.Run("interrupts running execution", func(t *testing.T) {

		originalExecDir := execDirPath
		defer func() { execDirPath = originalExecDir }()

		testDir := filepath.Join(os.TempDir(), "go-ent-test-interrupt-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		execDirPath = testDir

		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
			Logger:                 nil,
		}, nil)

		task := &Task{
			Description: "Test task for interrupt",
			Type:        "test",
		}

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)

		state.WithConfig(cfg)

		state.Runtime = domain.RuntimeCLI
		assert.NoError(t, state.Start())

		err := SaveState(state)
		assert.NoError(t, err)

		ctx := context.Background()
		err = engine.Interrupt(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)
		assert.True(t, loadedState.IsInterrupted())
		assert.Equal(t, ExecutionStatusInterrupted, loadedState.Status)
	})

	t.Run("fails to interrupt non-running execution", func(t *testing.T) {
		engine := New(Config{Logger: nil}, nil)

		task := &Task{
			Description: "Test task",
		}

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)

		state.WithConfig(cfg)

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

		assert.NoError(t, SaveState(state))

		ctx := context.Background()
		err := engine.Interrupt(ctx, state.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot interrupt execution with status")
	})

	t.Run("fails to interrupt missing execution", func(t *testing.T) {

		engine := New(Config{Logger: nil}, nil)

		ctx := context.Background()
		err := engine.Interrupt(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "load execution state")
	})

	t.Run("saves checkpoint on interrupt", func(t *testing.T) {
		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
			Logger:                 nil,
		}, nil)

		task := &Task{
			Description: "Test task for checkpoint",
		}

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)

		state.WithConfig(cfg)

		state.Runtime = domain.RuntimeCLI
		assert.NoError(t, state.Start())

		assert.NoError(t, SaveState(state))

		ctx := context.Background()
		err := engine.Interrupt(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)
		assert.True(t, loadedState.IsInterrupted())
		assert.Equal(t, ExecutionStatusInterrupted, loadedState.Status)
	})
}

func TestEngine_InterruptErrorMessages(t *testing.T) {
	t.Run("missing execution ID", func(t *testing.T) {

		engine := New(Config{Logger: nil}, nil)

		ctx := context.Background()
		err := engine.Interrupt(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "load execution state")
	})
}

func TestEngine_InterruptErrorTypes(t *testing.T) {
	t.Run("returns error for completed execution", func(t *testing.T) {

		engine := New(Config{Logger: nil}, nil)

		task := &Task{
			Description: "Test task",
		}

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)

		state.WithConfig(cfg)

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

		assert.NoError(t, SaveState(state))

		ctx := context.Background()
		err := engine.Interrupt(ctx, state.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot interrupt execution with status")
	})

	t.Run("returns error for failed execution", func(t *testing.T) {

		engine := New(Config{Logger: nil}, nil)

		task := &Task{
			Description: "Test task",
		}

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)

		state.WithConfig(cfg)

		state.Runtime = domain.RuntimeCLI
		assert.NoError(t, state.Start())
		assert.NoError(t, state.Fail(assert.AnError))

		assert.NoError(t, SaveState(state))

		ctx := context.Background()
		err := engine.Interrupt(ctx, state.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot interrupt execution with status")
	})

	t.Run("returns error for interrupted execution", func(t *testing.T) {

		engine := New(Config{Logger: nil}, nil)

		task := &Task{
			Description: "Test task",
		}

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)

		state.WithConfig(cfg)

		state.Runtime = domain.RuntimeCLI
		assert.NoError(t, state.Start())
		assert.NoError(t, state.Interrupt())

		assert.NoError(t, SaveState(state))

		ctx := context.Background()
		err := engine.Interrupt(ctx, state.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot interrupt execution with status")
	})

	t.Run("returns error for execution without runtime", func(t *testing.T) {

		engine := New(Config{Logger: nil}, nil)

		task := &Task{
			Description: "Test task",
		}

		state := NewExecutionState(task)

		assert.NoError(t, state.Start())

		assert.NoError(t, SaveState(state))

		ctx := context.Background()
		err := engine.Interrupt(ctx, state.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no runtime configured")
	})
}

func TestEngine_InterruptContext(t *testing.T) {
	t.Run("respects context cancellation", func(t *testing.T) {

		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
			Logger:                 nil,
		}, nil)

		task := &Task{
			Description: "Test task for context",
		}

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)

		state.WithConfig(cfg)

		state.Runtime = domain.RuntimeCLI
		assert.NoError(t, state.Start())

		assert.NoError(t, SaveState(state))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := engine.Interrupt(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)
		assert.True(t, loadedState.IsInterrupted())
	})

	t.Run("updates timestamps", func(t *testing.T) {

		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
			Logger:                 nil,
		}, nil)

		task := &Task{
			Description: "Test task for timestamps",
		}

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)

		state.WithConfig(cfg)

		state.Runtime = domain.RuntimeCLI
		startTime := time.Now()
		assert.NoError(t, state.Start())

		assert.NoError(t, SaveState(state))

		time.Sleep(10 * time.Millisecond)

		ctx := context.Background()
		err := engine.Interrupt(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)
		assert.True(t, loadedState.UpdatedAt.After(startTime))
		assert.Equal(t, ExecutionStatusInterrupted, loadedState.Status)
	})
}

func TestEngine_InterruptAcrossRuntimes(t *testing.T) {
	t.Run("interrupts CLI runtime execution", func(t *testing.T) {

		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
			Logger:                 nil,
		}, nil)

		task := &Task{
			Description: "Test CLI task",
		}

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)

		state.WithConfig(cfg)

		state.Runtime = domain.RuntimeCLI
		assert.NoError(t, state.Start())

		assert.NoError(t, SaveState(state))

		ctx := context.Background()
		err := engine.Interrupt(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)
		assert.True(t, loadedState.IsInterrupted())
	})

	t.Run("interrupts claude-code runtime execution", func(t *testing.T) {

		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
			Logger:                 nil,
		}, nil)

		task := &Task{
			Description: "Test claude-code task",
		}

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)

		state.WithConfig(cfg)

		state.Runtime = domain.RuntimeClaudeCode
		assert.NoError(t, state.Start())

		assert.NoError(t, SaveState(state))

		ctx := context.Background()
		err := engine.Interrupt(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)
		assert.True(t, loadedState.IsInterrupted())
	})

	t.Run("interrupts open-code runtime execution", func(t *testing.T) {

		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
			Logger:                 nil,
		}, nil)

		task := &Task{
			Description: "Test open-code task",
		}

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)

		state.WithConfig(cfg)

		state.Runtime = domain.RuntimeOpenCode
		assert.NoError(t, state.Start())

		assert.NoError(t, SaveState(state))

		ctx := context.Background()
		err := engine.Interrupt(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)
		assert.True(t, loadedState.IsInterrupted())
	})
}

func TestEngine_InterruptPreservesMetadata(t *testing.T) {
	t.Run("preserves existing metadata on interrupt", func(t *testing.T) {

		engine := New(Config{
			IsMCPMode:              true,
			EnableAutoCheckpoint:   true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
			Logger:                 nil,
		}, nil)

		task := &Task{
			Description: "Test task with metadata",
			Metadata: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		}

		state := NewExecutionState(task)

		cfg := engine.determineExecutionConfig(task)

		state.WithConfig(cfg)

		state.Runtime = domain.RuntimeCLI
		state.SetMetadata("custom", "data")
		assert.NoError(t, state.Start())

		assert.NoError(t, SaveState(state))

		ctx := context.Background()
		err := engine.Interrupt(ctx, state.ID)
		assert.NoError(t, err)

		loadedState, err := LoadState(state.ID)
		assert.NoError(t, err)
		assert.True(t, loadedState.IsInterrupted())

		customData, ok := loadedState.GetMetadata("custom")
		assert.True(t, ok)
		assert.Equal(t, "data", customData)
	})
}
