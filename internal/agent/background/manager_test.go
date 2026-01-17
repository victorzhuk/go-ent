package background

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/domain"
)

func TestNewAgent(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		role    string
		model   string
		task    string
		wantErr bool
	}{
		{
			name:  "valid agent",
			id:    "test-id",
			role:  "developer",
			model: "haiku",
			task:  "test task",
		},
		{
			name:    "missing id",
			role:    "developer",
			model:   "haiku",
			task:    "test task",
			wantErr: true,
		},
		{
			name:    "missing role",
			id:      "test-id",
			model:   "haiku",
			task:    "test task",
			wantErr: true,
		},
		{
			name:    "missing model",
			id:      "test-id",
			role:    "developer",
			task:    "test task",
			wantErr: true,
		},
		{
			name:    "missing task",
			id:      "test-id",
			role:    "developer",
			model:   "haiku",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			agent, err := NewAgent(tt.id, tt.role, tt.model, tt.task)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, agent)
			} else {
				require.NoError(t, err)
				require.NotNil(t, agent)
				assert.Equal(t, tt.id, agent.ID)
				assert.Equal(t, tt.role, agent.Role)
				assert.Equal(t, tt.model, agent.Model)
				assert.Equal(t, tt.task, agent.Task)
				assert.Equal(t, StatusPending, agent.Status)
				assert.False(t, agent.CreatedAt.IsZero())
			}
		})
	}
}

func TestAgent_Start(t *testing.T) {
	t.Parallel()

	agent, err := NewAgent("test-id", "developer", "haiku", "test task")
	require.NoError(t, err)

	agent.Start()

	assert.Equal(t, StatusRunning, agent.Status)
	assert.False(t, agent.StartedAt.IsZero())
}

func TestAgent_Complete(t *testing.T) {
	t.Parallel()

	agent, err := NewAgent("test-id", "developer", "haiku", "test task")
	require.NoError(t, err)

	agent.Start()
	output := "task completed successfully"
	agent.Complete(output)

	assert.Equal(t, StatusCompleted, agent.Status)
	assert.Equal(t, output, agent.Output)
	assert.False(t, agent.CompletedAt.IsZero())
}

func TestAgent_Fail(t *testing.T) {
	t.Parallel()

	agent, err := NewAgent("test-id", "developer", "haiku", "test task")
	require.NoError(t, err)

	agent.Start()
	err = assert.AnError
	agent.Fail(err)

	assert.Equal(t, StatusFailed, agent.Status)
	assert.Equal(t, err, agent.Error)
	assert.False(t, agent.CompletedAt.IsZero())
}

func TestAgent_Kill(t *testing.T) {
	t.Parallel()

	agent, err := NewAgent("test-id", "developer", "haiku", "test task")
	require.NoError(t, err)

	agent.Start()
	agent.Kill()

	assert.Equal(t, StatusKilled, agent.Status)
	assert.False(t, agent.CompletedAt.IsZero())
}

func TestAgent_Duration(t *testing.T) {
	t.Parallel()

	agent, err := NewAgent("test-id", "developer", "haiku", "test task")
	require.NoError(t, err)

	assert.Zero(t, agent.Duration())

	agent.Start()
	time.Sleep(10 * time.Millisecond)

	duration := agent.Duration()
	assert.Greater(t, duration.Milliseconds(), int64(0))

	agent.Complete("done")
	assert.Greater(t, duration.Milliseconds(), int64(0))
}

func TestStatus_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status Status
		want   bool
	}{
		{StatusPending, true},
		{StatusRunning, true},
		{StatusCompleted, true},
		{StatusFailed, true},
		{StatusKilled, true},
		{Status("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.status.Valid())
		})
	}
}

type mockSelector struct{}

func (m *mockSelector) Select(ctx context.Context, task interface{}) (*SelectionResult, error) {
	return &SelectionResult{
		Role:   "developer",
		Model:  "haiku",
		Skills: []string{"go-code"},
		Reason: "test reason",
	}, nil
}

func TestNewManager(t *testing.T) {
	t.Parallel()

	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	require.NotNil(t, mgr)
	assert.Equal(t, 0, mgr.Count())
}

func TestManager_Spawn(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	agent, err := mgr.Spawn(ctx, "test task", SpawnOpts{})
	require.NoError(t, err)
	require.NotNil(t, agent)
	assert.NotEmpty(t, agent.ID)
	assert.Equal(t, "test task", agent.Task)
	assert.Equal(t, cfg.DefaultModel, agent.Model)
	assert.Equal(t, 1, mgr.Count())

	time.Sleep(100 * time.Millisecond)

	updated, err := mgr.Get(agent.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, updated.Status)
	assert.NotEmpty(t, updated.Output)
}

func TestManager_SpawnEmptyTask(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	agent, err := mgr.Spawn(ctx, "", SpawnOpts{})
	assert.Error(t, err)
	assert.Nil(t, agent)
	assert.Equal(t, 0, mgr.Count())
}

func TestManager_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	agent, err := mgr.Spawn(ctx, "test task", SpawnOpts{})
	require.NoError(t, err)

	retrieved, err := mgr.Get(agent.ID)
	require.NoError(t, err)
	assert.Equal(t, agent.ID, retrieved.ID)
	assert.Equal(t, agent.Task, retrieved.Task)

	_, err = mgr.Get("non-existent")
	assert.Error(t, err)
	assert.Equal(t, ErrAgentNotFound, err)
}

func TestManager_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	agents := mgr.List("")
	assert.Empty(t, agents)

	_, err := mgr.Spawn(ctx, "task 1", SpawnOpts{})
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	_, err = mgr.Spawn(ctx, "task 2", SpawnOpts{})
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	allAgents := mgr.List("")
	assert.Len(t, allAgents, 2)

	completedAgents := mgr.List(StatusCompleted)
	assert.GreaterOrEqual(t, len(completedAgents), 1)

	pendingAgents := mgr.List(StatusPending)
	assert.Empty(t, pendingAgents)
}

func TestManager_Kill(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	agent, err := mgr.Spawn(ctx, "test task", SpawnOpts{})
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	updated, err := mgr.Get(agent.ID)
	require.NoError(t, err)
	if updated.Status == StatusRunning {
		err = mgr.Kill(ctx, agent.ID)
		require.NoError(t, err)

		killed, err := mgr.Get(agent.ID)
		require.NoError(t, err)
		assert.Equal(t, StatusKilled, killed.Status)
	}
}

func TestManager_KillNonExistent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	err := mgr.Kill(ctx, "non-existent")
	assert.Error(t, err)
	assert.Equal(t, ErrAgentNotFound, err)
}

func TestManager_Cleanup(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	_, err := mgr.Spawn(ctx, "task 1", SpawnOpts{})
	require.NoError(t, err)

	_, err = mgr.Spawn(ctx, "task 2", SpawnOpts{})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	count := mgr.Cleanup(ctx)
	assert.GreaterOrEqual(t, count, 2)
}

func TestManager_CountByStatus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	assert.Zero(t, mgr.CountByStatus(StatusCompleted))

	_, err := mgr.Spawn(ctx, "test task", SpawnOpts{})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	completedCount := mgr.CountByStatus(StatusCompleted)
	assert.GreaterOrEqual(t, completedCount, 1)

	totalCount := mgr.Count()
	assert.Equal(t, totalCount, completedCount)
}

func TestManager_Shutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	_, err := mgr.Spawn(ctx, "task 1", SpawnOpts{})
	require.NoError(t, err)

	_, err = mgr.Spawn(ctx, "task 2", SpawnOpts{})
	require.NoError(t, err)

	err = mgr.Shutdown(ctx)
	require.NoError(t, err)

	assert.Zero(t, mgr.Count())
}

func TestManager_ConcurrentSpawns(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := Config{
		MaxConcurrent: 10,
		DefaultRole:   "developer",
		DefaultModel:  "haiku",
		Timeout:       300,
		ResourceLimits: config.ResourceLimits{
			MaxMemoryMB:   512,
			MaxGoroutines: 100,
			MaxCPUPercent: 80,
		},
	}
	mgr := NewManager(selector, cfg)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := mgr.Spawn(ctx, fmt.Sprintf("task %d", i), SpawnOpts{})
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()

	time.Sleep(200 * time.Millisecond)

	agents := mgr.List("")
	assert.Len(t, agents, 10)

	allCompleted := true
	for _, agent := range agents {
		if agent.Status != StatusCompleted {
			allCompleted = false
			break
		}
	}
	assert.True(t, allCompleted)
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	assert.Equal(t, 5, cfg.MaxConcurrent)
	assert.Equal(t, "haiku", cfg.DefaultModel)
	assert.Equal(t, 300, cfg.Timeout)
	assert.Equal(t, 512, cfg.ResourceLimits.MaxMemoryMB)
	assert.Equal(t, 100, cfg.ResourceLimits.MaxGoroutines)
	assert.Equal(t, 80, cfg.ResourceLimits.MaxCPUPercent)
}

func TestNewManager_ZeroConfig(t *testing.T) {
	t.Parallel()

	selector := &mockSelector{}
	cfg := Config{
		MaxConcurrent: 0,
		DefaultRole:   "",
		DefaultModel:  "",
		Timeout:       0,
	}

	mgr := NewManager(selector, cfg)
	assert.NotNil(t, mgr)
	assert.Equal(t, 0, mgr.Count())
}

func TestClassifyTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		task     string
		expected TaskType
	}{
		{
			name:     "critical task",
			task:     "Critical security review for production deployment",
			expected: TaskTypeCritical,
		},
		{
			name:     "decision task",
			task:     "Approve breaking API changes",
			expected: TaskTypeCritical,
		},
		{
			name:     "delete task",
			task:     "Remove all temporary files and clean up directory",
			expected: TaskTypeCritical,
		},
		{
			name:     "complexity task",
			task:     "Implement new authentication system with JWT tokens",
			expected: TaskTypeComplexity,
		},
		{
			name:     "refactor task",
			task:     "Refactor database layer for better performance",
			expected: TaskTypeComplexity,
		},
		{
			name:     "design task",
			task:     "Design microservices architecture for the billing system",
			expected: TaskTypeComplexity,
		},
		{
			name:     "exploration task",
			task:     "Explore the codebase and find all TODO comments",
			expected: TaskTypeExploration,
		},
		{
			name:     "analyze task",
			task:     "Analyze the current performance bottlenecks",
			expected: TaskTypeExploration,
		},
		{
			name:     "search task",
			task:     "Search for all unused imports in the project",
			expected: TaskTypeExploration,
		},
		{
			name:     "default exploration",
			task:     "Some random task description",
			expected: TaskTypeExploration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := classifyTask(tt.task)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSelectModelByTier(t *testing.T) {
	t.Parallel()

	modelTier := config.ModelTierConfig{
		Exploration: "haiku",
		Complexity:  "sonnet",
		Critical:    "opus",
	}

	tests := []struct {
		name      string
		taskType  TaskType
		modelTier config.ModelTierConfig
		expected  string
	}{
		{
			name:      "exploration selects haiku",
			taskType:  TaskTypeExploration,
			modelTier: modelTier,
			expected:  "haiku",
		},
		{
			name:      "complexity selects sonnet",
			taskType:  TaskTypeComplexity,
			modelTier: modelTier,
			expected:  "sonnet",
		},
		{
			name:      "critical selects opus",
			taskType:  TaskTypeCritical,
			modelTier: modelTier,
			expected:  "opus",
		},
		{
			name:      "unknown type returns empty",
			taskType:  TaskType("unknown"),
			modelTier: modelTier,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := selectModelByTier(tt.taskType, tt.modelTier)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_Spawn_WithModelTier(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	models := config.ModelsConfig{
		"haiku":  "claude-haiku-4-5",
		"sonnet": "claude-sonnet-4-5",
		"opus":   "claude-opus-4-5",
	}

	modelTier := config.ModelTierConfig{
		Exploration: "haiku",
		Complexity:  "sonnet",
		Critical:    "opus",
	}

	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "default-model",
		Timeout:       300,
		ModelTier:     modelTier,
		Models:        models,
	}

	mgr := NewManager(nil, cfg)

	tests := []struct {
		name          string
		task          string
		expectedModel string
	}{
		{
			name:          "exploration task gets haiku",
			task:          "Explore the codebase and find all TODO comments",
			expectedModel: "claude-haiku-4-5",
		},
		{
			name:          "complexity task gets sonnet",
			task:          "Implement new authentication system with JWT tokens",
			expectedModel: "claude-sonnet-4-5",
		},
		{
			name:          "critical task gets opus",
			task:          "Critical security review for production deployment",
			expectedModel: "claude-opus-4-5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := mgr.Spawn(ctx, tt.task, SpawnOpts{})
			require.NoError(t, err)
			require.NotNil(t, agent)
			assert.Equal(t, tt.expectedModel, agent.Model)
		})
	}
}

func TestManager_Spawn_WithModelTier_AndSelector(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	selector := &mockSelector{}

	models := config.ModelsConfig{
		"haiku":  "claude-haiku-4-5",
		"sonnet": "claude-sonnet-4-5",
		"opus":   "claude-opus-4-5",
	}

	modelTier := config.ModelTierConfig{
		Exploration: "haiku",
		Complexity:  "sonnet",
		Critical:    "opus",
	}

	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "default-model",
		Timeout:       300,
		ModelTier:     modelTier,
		Models:        models,
	}

	mgr := NewManager(selector, cfg)

	agent, err := mgr.Spawn(ctx, "Critical task", SpawnOpts{})
	require.NoError(t, err)
	require.NotNil(t, agent)

	assert.Equal(t, "haiku", agent.Model)
}

func TestManager_Spawn_WithModelTier_MissingModel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	models := config.ModelsConfig{
		"haiku": "claude-haiku-4-5",
		"opus":  "claude-opus-4-5",
	}

	modelTier := config.ModelTierConfig{
		Exploration: "haiku",
		Complexity:  "sonnet",
		Critical:    "opus",
	}

	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "default-model",
		Timeout:       300,
		ModelTier:     modelTier,
		Models:        models,
	}

	mgr := NewManager(nil, cfg)

	agent, err := mgr.Spawn(ctx, "Implement new feature", SpawnOpts{})
	require.NoError(t, err)
	require.NotNil(t, agent)

	assert.Equal(t, "default-model", agent.Model)
}

func TestManager_Spawn_WithModelTier_EmptyModels(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	modelTier := config.ModelTierConfig{
		Exploration: "haiku",
		Complexity:  "sonnet",
		Critical:    "opus",
	}

	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "default-model",
		Timeout:       300,
		ModelTier:     modelTier,
		Models:        nil,
	}

	mgr := NewManager(nil, cfg)

	agent, err := mgr.Spawn(ctx, "Explore codebase", SpawnOpts{})
	require.NoError(t, err)
	require.NotNil(t, agent)

	assert.Equal(t, "default-model", agent.Model)
}

func TestManager_Spawn_WithRoleOverride(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	opts := SpawnOpts{
		Role: "architect",
	}

	agent, err := mgr.Spawn(ctx, "test task", opts)
	require.NoError(t, err)
	require.NotNil(t, agent)

	assert.Equal(t, "architect", agent.Role)
	assert.Equal(t, "test task", agent.Task)
}

func TestManager_Spawn_WithModelOverride(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}

	models := config.ModelsConfig{
		"haiku":  "claude-haiku-4-5",
		"sonnet": "claude-sonnet-4-5",
		"opus":   "claude-opus-4-5",
	}

	modelTier := config.ModelTierConfig{
		Exploration: "haiku",
		Complexity:  "sonnet",
		Critical:    "opus",
	}

	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "claude-haiku-4-5",
		Timeout:       300,
		ModelTier:     modelTier,
		Models:        models,
	}

	mgr := NewManager(selector, cfg)

	tests := []struct {
		name          string
		task          string
		opts          SpawnOpts
		expectedModel string
	}{
		{
			name:          "override to opus for exploration task",
			task:          "Explore codebase",
			opts:          SpawnOpts{Model: "claude-opus-4-5"},
			expectedModel: "claude-opus-4-5",
		},
		{
			name:          "override to haiku for complexity task",
			task:          "Implement new feature",
			opts:          SpawnOpts{Model: "claude-haiku-4-5"},
			expectedModel: "claude-haiku-4-5",
		},
		{
			name:          "override to custom model",
			task:          "Any task",
			opts:          SpawnOpts{Model: "custom-model-id"},
			expectedModel: "custom-model-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := mgr.Spawn(ctx, tt.task, tt.opts)
			require.NoError(t, err)
			require.NotNil(t, agent)
			assert.Equal(t, tt.expectedModel, agent.Model)
		})
	}
}

func TestManager_Spawn_WithBothRoleAndModelOverride(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	opts := SpawnOpts{
		Role:  "senior",
		Model: "claude-sonnet-4-5",
	}

	agent, err := mgr.Spawn(ctx, "test task", opts)
	require.NoError(t, err)
	require.NotNil(t, agent)

	assert.Equal(t, "senior", agent.Role)
	assert.Equal(t, "claude-sonnet-4-5", agent.Model)
}

func TestManager_Spawn_WithTimeoutOverride(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	opts := SpawnOpts{
		Role:    "developer",
		Model:   "haiku",
		Timeout: 600,
	}

	agent, err := mgr.Spawn(ctx, "test task", opts)
	require.NoError(t, err)
	require.NotNil(t, agent)

	assert.Equal(t, "developer", agent.Role)
	assert.Equal(t, "haiku", agent.Model)
}

func TestManager_Spawn_OverrideTakesPrecedenceOverSelector(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	opts := SpawnOpts{
		Role:  "architect",
		Model: "claude-opus-4-5",
	}

	agent, err := mgr.Spawn(ctx, "test task", opts)
	require.NoError(t, err)
	require.NotNil(t, agent)

	assert.Equal(t, "architect", agent.Role)
	assert.Equal(t, "claude-opus-4-5", agent.Model)
}

func TestManager_Spawn_OverrideTakesPrecedenceOverModelTier(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	models := config.ModelsConfig{
		"haiku":  "claude-haiku-4-5",
		"sonnet": "claude-sonnet-4-5",
		"opus":   "claude-opus-4-5",
	}

	modelTier := config.ModelTierConfig{
		Exploration: "haiku",
		Complexity:  "sonnet",
		Critical:    "opus",
	}

	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "claude-haiku-4-5",
		Timeout:       300,
		ModelTier:     modelTier,
		Models:        models,
	}

	mgr := NewManager(nil, cfg)

	opts := SpawnOpts{
		Model: "claude-opus-4-5",
	}

	agent, err := mgr.Spawn(ctx, "Explore codebase", opts)
	require.NoError(t, err)
	require.NotNil(t, agent)

	assert.Equal(t, "claude-opus-4-5", agent.Model)
}

func TestManager_Spawn_EmptyOptsUsesDefaultRouting(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	models := config.ModelsConfig{
		"haiku":  "claude-haiku-4-5",
		"sonnet": "claude-sonnet-4-5",
		"opus":   "claude-opus-4-5",
	}

	modelTier := config.ModelTierConfig{
		Exploration: "haiku",
		Complexity:  "sonnet",
		Critical:    "opus",
	}

	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "default-model",
		Timeout:       300,
		ModelTier:     modelTier,
		Models:        models,
	}

	mgr := NewManager(nil, cfg)

	opts := SpawnOpts{}

	agent, err := mgr.Spawn(ctx, "Implement new feature", opts)
	require.NoError(t, err)
	require.NotNil(t, agent)

	assert.Equal(t, "claude-sonnet-4-5", agent.Model)
}

func TestManager_Spawn_WithAgentSelector(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	registry := &mockSkillRegistry{}
	agentSelector := agent.NewSelector(agent.Config{
		MaxBudget:  1000,
		StrictMode: false,
	}, registry)

	selectorAdapter := NewAgentSelectorAdapter(agentSelector)

	cfg := DefaultConfig()
	mgr := NewManager(selectorAdapter, cfg)

	tests := []struct {
		name         string
		task         string
		expectedRole string
	}{
		{
			name:         "moderately complex task gets developer",
			task:         "Implement new authentication system with JWT tokens",
			expectedRole: "developer",
		},
		{
			name:         "simple task gets developer",
			task:         "Explore the codebase",
			expectedRole: "developer",
		},
		{
			name:         "architecture task gets architect",
			task:         "Design microservices architecture for billing",
			expectedRole: "architect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := mgr.Spawn(ctx, tt.task, SpawnOpts{})
			require.NoError(t, err)
			require.NotNil(t, agent)
			assert.Equal(t, tt.expectedRole, agent.Role)
		})
	}
}

type mockSkillRegistry struct{}

func (m *mockSkillRegistry) MatchForContext(ctx domain.SkillContext) []string {
	return []string{"go-code", "go-arch"}
}

func TestToAgentTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		taskString string
		wantAction domain.SpecAction
		wantPhase  domain.ActionPhase
		wantType   agent.TaskType
	}{
		{
			name:       "design task maps correctly",
			taskString: "Design microservices architecture for billing",
			wantAction: domain.SpecActionDesign,
			wantPhase:  domain.ActionPhasePlanning,
			wantType:   agent.TaskTypeArchitecture,
		},
		{
			name:       "implement task maps correctly",
			taskString: "Implement new feature",
			wantAction: domain.SpecActionImplement,
			wantPhase:  domain.ActionPhaseExecution,
			wantType:   agent.TaskTypeFeature,
		},
		{
			name:       "analyze task maps correctly",
			taskString: "Analyze performance",
			wantAction: domain.SpecActionAnalyze,
			wantPhase:  domain.ActionPhaseDiscovery,
			wantType:   agent.TaskTypeDocumentation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toAgentTask(tt.taskString)
			assert.Equal(t, tt.taskString, result.Description)
			assert.Equal(t, tt.wantAction, result.Action)
			assert.Equal(t, tt.wantPhase, result.Phase)
			assert.Equal(t, tt.wantType, result.Type)
			assert.NotNil(t, result.Files)
			assert.NotNil(t, result.Metadata)
		})
	}
}

func TestAgentSelectorAdapter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	registry := &mockSkillRegistry{}
	agentSelector := agent.NewSelector(agent.Config{
		MaxBudget:  1000,
		StrictMode: false,
	}, registry)

	adapter := NewAgentSelectorAdapter(agentSelector)

	task := agent.Task{
		Description: "Design microservices architecture for billing",
		Type:        agent.TaskTypeArchitecture,
		Action:      domain.SpecActionDesign,
		Phase:       domain.ActionPhasePlanning,
		Files:       []string{},
		Metadata:    make(map[string]interface{}),
	}

	result, err := adapter.Select(ctx, task)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "architect", result.Role)
	assert.Equal(t, "opus", result.Model)
}

func TestManager_CleanupOld(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "haiku",
		Timeout:       300,
	}
	mgr := NewManager(selector, cfg)

	_, err := mgr.Spawn(ctx, "task 1", SpawnOpts{})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	_, err = mgr.Spawn(ctx, "task 2", SpawnOpts{})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 2, mgr.Count())

	cleaned := mgr.CleanupOld(ctx, 10*time.Millisecond)
	assert.GreaterOrEqual(t, cleaned, 2)

	assert.LessOrEqual(t, mgr.Count(), 2)
}

func TestManager_CleanupOld_RecentAgents(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "haiku",
		Timeout:       300,
	}
	mgr := NewManager(selector, cfg)

	_, err := mgr.Spawn(ctx, "task 1", SpawnOpts{})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	cleaned := mgr.CleanupOld(ctx, time.Hour)
	assert.Zero(t, cleaned)

	assert.Equal(t, 1, mgr.Count())
}

func TestManager_OnShutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	hookCalled := false
	mgr.OnShutdown(func(context.Context) error {
		hookCalled = true
		return nil
	})

	err := mgr.Shutdown(ctx)
	require.NoError(t, err)

	assert.True(t, hookCalled)
	assert.Zero(t, mgr.Count())
}

func TestManager_OnShutdown_MultipleHooks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	hook1Called := false
	hook2Called := false

	mgr.OnShutdown(func(context.Context) error {
		hook1Called = true
		return nil
	})

	mgr.OnShutdown(func(context.Context) error {
		hook2Called = true
		return nil
	})

	err := mgr.Shutdown(ctx)
	require.NoError(t, err)

	assert.True(t, hook1Called)
	assert.True(t, hook2Called)
}

func TestManager_OnShutdown_HookError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	expectedErr := fmt.Errorf("hook failed")
	mgr.OnShutdown(func(context.Context) error {
		return expectedErr
	})

	err := mgr.Shutdown(ctx)

	assert.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
}

func TestManager_StartCleanupRoutine(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	selector := &mockSelector{}
	cfg := Config{
		MaxConcurrent:   5,
		DefaultRole:     "developer",
		DefaultModel:    "haiku",
		Timeout:         300,
		CleanupInterval: 100 * time.Millisecond,
		MaxAgentAge:     50 * time.Millisecond,
	}
	mgr := NewManager(selector, cfg)

	_, err := mgr.Spawn(ctx, "task 1", SpawnOpts{})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, mgr.Count())

	mgr.StartCleanupRoutine(ctx)

	time.Sleep(250 * time.Millisecond)

	assert.LessOrEqual(t, mgr.Count(), 1)

	cancel()
	time.Sleep(150 * time.Millisecond)
}

func TestManager_ShutdownWithActiveAgents(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "haiku",
		Timeout:       300,
	}
	mgr := NewManager(selector, cfg)

	_, err := mgr.Spawn(ctx, "task 1", SpawnOpts{})
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	_, err = mgr.Spawn(ctx, "task 2", SpawnOpts{})
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	err = mgr.Shutdown(ctx)
	require.NoError(t, err)

	assert.Zero(t, mgr.Count())
}

func TestManager_IncrementGoroutines(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "haiku",
		Timeout:       300,
		ResourceLimits: config.ResourceLimits{
			MaxGoroutines: 10,
		},
	}
	mgr := NewManager(selector, cfg)

	agent, err := mgr.Spawn(ctx, "test task", SpawnOpts{})
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		err := mgr.IncrementGoroutines(agent.ID)
		assert.NoError(t, err)
	}

	err = mgr.IncrementGoroutines(agent.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "goroutine limit")
}

func TestManager_IncrementGoroutines_NoLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "haiku",
		Timeout:       300,
		ResourceLimits: config.ResourceLimits{
			MaxGoroutines: 0,
		},
	}
	mgr := NewManager(selector, cfg)

	agent, err := mgr.Spawn(ctx, "test task", SpawnOpts{})
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		err := mgr.IncrementGoroutines(agent.ID)
		assert.NoError(t, err)
	}
}

func TestManager_DecrementGoroutines(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "haiku",
		Timeout:       300,
		ResourceLimits: config.ResourceLimits{
			MaxGoroutines: 10,
		},
	}
	mgr := NewManager(selector, cfg)

	agent, err := mgr.Spawn(ctx, "test task", SpawnOpts{})
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		err := mgr.IncrementGoroutines(agent.ID)
		assert.NoError(t, err)
	}

	for i := 0; i < 5; i++ {
		mgr.DecrementGoroutines(agent.ID)
	}

	usage := mgr.GetResourceUsage()
	assert.Equal(t, 0, usage[agent.ID])
}

func TestManager_GetResourceUsage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "haiku",
		Timeout:       300,
		ResourceLimits: config.ResourceLimits{
			MaxGoroutines: 10,
		},
	}
	mgr := NewManager(selector, cfg)

	agent1, err := mgr.Spawn(ctx, "test task 1", SpawnOpts{})
	require.NoError(t, err)

	agent2, err := mgr.Spawn(ctx, "test task 2", SpawnOpts{})
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		err := mgr.IncrementGoroutines(agent1.ID)
		assert.NoError(t, err)
	}

	for i := 0; i < 5; i++ {
		err := mgr.IncrementGoroutines(agent2.ID)
		assert.NoError(t, err)
	}

	usage := mgr.GetResourceUsage()
	assert.Equal(t, 3, usage[agent1.ID])
	assert.Equal(t, 5, usage[agent2.ID])
}

func TestManager_ResourceLimits_Default(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	assert.Equal(t, 512, cfg.ResourceLimits.MaxMemoryMB)
	assert.Equal(t, 100, cfg.ResourceLimits.MaxGoroutines)
	assert.Equal(t, 80, cfg.ResourceLimits.MaxCPUPercent)
}

func TestManager_Spawn_MaxConcurrentLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := Config{
		MaxConcurrent: 2,
		DefaultRole:   "developer",
		DefaultModel:  "haiku",
		Timeout:       300,
		ResourceLimits: config.ResourceLimits{
			MaxMemoryMB:   512,
			MaxGoroutines: 100,
			MaxCPUPercent: 80,
		},
	}
	mgr := NewManager(selector, cfg)

	_, err := mgr.Spawn(ctx, "task 1", SpawnOpts{})
	require.NoError(t, err)

	_, err = mgr.Spawn(ctx, "task 2", SpawnOpts{})
	require.NoError(t, err)

	_, err = mgr.Spawn(ctx, "task 3", SpawnOpts{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max concurrent agents")
}

func TestManager_Cleanup_RemovesGoroutineTracking(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "haiku",
		Timeout:       300,
	}
	mgr := NewManager(selector, cfg)

	agent, err := mgr.Spawn(ctx, "test task", SpawnOpts{})
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		err := mgr.IncrementGoroutines(agent.ID)
		assert.NoError(t, err)
	}

	time.Sleep(100 * time.Millisecond)

	mgr.Cleanup(ctx)

	usage := mgr.GetResourceUsage()
	_, exists := usage[agent.ID]
	assert.False(t, exists)
}

func TestConfig_ResourceLimits_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		limits        config.ResourceLimits
		expectError   bool
		errorContains string
	}{
		{
			name: "valid limits",
			limits: config.ResourceLimits{
				MaxMemoryMB:   512,
				MaxGoroutines: 100,
				MaxCPUPercent: 80,
			},
			expectError: false,
		},
		{
			name: "zero limits (unlimited)",
			limits: config.ResourceLimits{
				MaxMemoryMB:   0,
				MaxGoroutines: 0,
				MaxCPUPercent: 0,
			},
			expectError: false,
		},
		{
			name: "negative memory",
			limits: config.ResourceLimits{
				MaxMemoryMB: -1,
			},
			expectError:   true,
			errorContains: "invalid",
		},
		{
			name: "negative goroutines",
			limits: config.ResourceLimits{
				MaxGoroutines: -1,
			},
			expectError:   true,
			errorContains: "invalid",
		},
		{
			name: "negative cpu percent",
			limits: config.ResourceLimits{
				MaxCPUPercent: -1,
			},
			expectError:   true,
			errorContains: "invalid",
		},
		{
			name: "cpu percent over 100",
			limits: config.ResourceLimits{
				MaxCPUPercent: 150,
			},
			expectError:   true,
			errorContains: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.limits.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
