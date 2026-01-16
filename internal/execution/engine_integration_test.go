package execution

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/domain"
)

func TestEngine_ExecuteWithSingleStrategy(t *testing.T) {
	ctx := context.Background()
	selector := agent.NewSelector(agent.Config{}, nil)
	engine := New(Config{
		PreferredRuntime: domain.RuntimeCLI,
		IsMCPMode:        false,
	}, selector)

	task := NewTask("List files in current directory").
		WithType("bugfix").
		WithAgent(domain.AgentRoleDeveloper).
		WithModel("haiku").
		WithRuntime(domain.RuntimeCLI).
		WithStrategy(domain.ExecutionStrategySingle)

	result, err := engine.Execute(ctx, task)
	require.NoError(t, err)
	assert.NotNil(t, result)
	// Result may not be successful in test environment, but should not error
}

func TestEngine_ExecuteWithMultiStrategy(t *testing.T) {
	ctx := context.Background()
	selector := agent.NewSelector(agent.Config{}, nil)
	engine := New(Config{
		PreferredRuntime: domain.RuntimeCLI,
		IsMCPMode:        false,
	}, selector)

	task := NewTask("Design and implement a simple function").
		WithType("feature").
		WithModel("haiku").
		WithRuntime(domain.RuntimeCLI).
		WithStrategy(domain.ExecutionStrategyMulti)

	result, err := engine.Execute(ctx, task)
	require.NoError(t, err)
	assert.NotNil(t, result)
	// Multi strategy executes architect -> developer
	if result.Metadata != nil {
		chain, ok := result.Metadata["agent_chain"]
		if ok {
			assert.NotEmpty(t, chain)
		}
	}
}

func TestEngine_Status(t *testing.T) {
	ctx := context.Background()
	selector := agent.NewSelector(agent.Config{}, nil)
	engine := New(Config{
		PreferredRuntime: domain.RuntimeClaudeCode,
		IsMCPMode:        true,
	}, selector)

	status := engine.Status(ctx)

	assert.NotEmpty(t, status.AvailableRuntimes)
	assert.NotEmpty(t, status.AvailableStrategies)
	assert.Equal(t, string(domain.RuntimeClaudeCode), status.PreferredRuntime)
}

func TestEngine_BudgetTracking(t *testing.T) {
	selector := agent.NewSelector(agent.Config{}, nil)
	engine := New(Config{
		IsMCPMode: true,
	}, selector)

	// Record some spending
	tracker := engine.GetBudgetTracker()
	tracker.Record("test-task", 1000, 500, 0.05)

	// Check spending
	tokens, cost := tracker.GetDailySpending()
	assert.Equal(t, 1500, tokens)
	assert.Equal(t, 0.05, cost)
}

func TestEngine_FallbackRuntimes(t *testing.T) {
	selector := agent.NewSelector(agent.Config{}, nil)
	engine := New(Config{
		PreferredRuntime: domain.RuntimeClaudeCode,
	}, selector)

	// Try to get a runner (may fallback)
	runner, err := engine.getRunner(domain.RuntimeClaudeCode)
	require.NoError(t, err)
	assert.NotNil(t, runner)
}

func TestEngine_StrategySelection(t *testing.T) {
	selector := agent.NewSelector(agent.Config{}, nil)
	engine := New(Config{}, selector)

	tests := []struct {
		name             string
		task             *Task
		expectedStrategy domain.ExecutionStrategy
	}{
		{
			name: "Single strategy for simple task",
			task: NewTask("Fix typo").
				WithStrategy(domain.ExecutionStrategySingle),
			expectedStrategy: domain.ExecutionStrategySingle,
		},
		{
			name: "Multi strategy for complex task",
			task: NewTask("Implement new feature").
				WithStrategy(domain.ExecutionStrategyMulti),
			expectedStrategy: domain.ExecutionStrategyMulti,
		},
		{
			name: "Parallel strategy with metadata",
			task: NewTask("Run multiple tasks").
				WithMetadata("parallel_tasks", []ParallelTask{
					{ID: "t1", Description: "Task 1"},
				}).
				WithStrategy(domain.ExecutionStrategyParallel),
			expectedStrategy: domain.ExecutionStrategyParallel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := engine.selectStrategy(tt.task)
			assert.Equal(t, tt.expectedStrategy, strategy.Name())
		})
	}
}
