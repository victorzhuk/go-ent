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

	agent, err := mgr.Spawn(ctx, "test task")
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

	agent, err := mgr.Spawn(ctx, "")
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

	agent, err := mgr.Spawn(ctx, "test task")
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

	_, err := mgr.Spawn(ctx, "task 1")
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	_, err = mgr.Spawn(ctx, "task 2")
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

	agent, err := mgr.Spawn(ctx, "test task")
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

	_, err := mgr.Spawn(ctx, "task 1")
	require.NoError(t, err)

	_, err = mgr.Spawn(ctx, "task 2")
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

	_, err := mgr.Spawn(ctx, "test task")
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

	_, err := mgr.Spawn(ctx, "task 1")
	require.NoError(t, err)

	_, err = mgr.Spawn(ctx, "task 2")
	require.NoError(t, err)

	err = mgr.Shutdown(ctx)
	require.NoError(t, err)

	assert.Zero(t, mgr.Count())
}

func TestManager_ConcurrentSpawns(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selector := &mockSelector{}
	cfg := DefaultConfig()
	mgr := NewManager(selector, cfg)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := mgr.Spawn(ctx, fmt.Sprintf("task %d", i))
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
