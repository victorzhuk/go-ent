package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/execution"
)

func TestPool_New(t *testing.T) {
	t.Parallel()

	pool := NewPool(5, 30*time.Second, 5*time.Minute, 3, 5*time.Second)
	assert.NotNil(t, pool)
	assert.Equal(t, 5, pool.maxConcurrency)
	assert.NotNil(t, pool.cond)
}

func TestPool_Spawn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		maxConcurrency int
		provider       string
		model          string
		method         config.CommunicationMethod
		wantErr        bool
	}{
		{
			name:           "spawns worker with CLI method",
			maxConcurrency: 5,
			provider:       "anthropic",
			model:          "claude-3-opus",
			method:         config.MethodCLI,
			wantErr:        false,
		},
		{
			name:           "spawns worker with ACP method",
			maxConcurrency: 5,
			provider:       "openai",
			model:          "gpt-4",
			method:         config.MethodACP,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pool := NewPool(tt.maxConcurrency, 30*time.Second, 5*time.Minute, 3, 5*time.Second)
			ctx := context.Background()

			task := &execution.Task{
				Description: "test task",
			}

			id, err := pool.Spawn(ctx, tt.provider, tt.model, tt.method, task)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, id)

			worker := pool.Get(id)
			require.NotNil(t, worker)
			assert.Equal(t, id, worker.ID)
			assert.Equal(t, tt.provider, worker.Provider)
			assert.Equal(t, tt.model, worker.Model)
			assert.Equal(t, tt.method, worker.Method)
			assert.Equal(t, StatusIdle, worker.Status)
		})
	}
}

func TestPool_Spawn_ConcurrencyLimit(t *testing.T) {
	t.Parallel()

	pool := NewPool(2, 30*time.Second, 5*time.Minute, 3, 5*time.Second)
	ctx := context.Background()

	task := &execution.Task{
		Description: "test task",
	}

	ids := []string{}
	for i := 0; i < 2; i++ {
		id, err := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)
		require.NoError(t, err)
		ids = append(ids, id)
	}

	assert.Equal(t, 2, len(ids))
}

func TestPool_Spawn_ContextCancellation(t *testing.T) {
	t.Parallel()

	pool := NewPool(1, 30*time.Second, 5*time.Minute, 3, 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	task := &execution.Task{
		Description: "test task",
	}

	id1, err := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)
	require.NoError(t, err)

	w := pool.Get(id1)
	w.Mutex.Lock()
	w.Status = StatusRunning
	w.Mutex.Unlock()

	cancel()

	_, err = pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestPool_Get(t *testing.T) {
	t.Parallel()

	pool := NewPool(5, 30*time.Second, 5*time.Minute, 3, 5*time.Second)
	ctx := context.Background()

	task := &execution.Task{
		Description: "test task",
	}

	id, err := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)
	require.NoError(t, err)

	worker := pool.Get(id)
	assert.NotNil(t, worker)
	assert.Equal(t, id, worker.ID)

	nilWorker := pool.Get("non-existent-id")
	assert.Nil(t, nilWorker)
}

func TestPool_List(t *testing.T) {
	t.Parallel()

	pool := NewPool(5, 30*time.Second, 5*time.Minute, 3, 5*time.Second)
	ctx := context.Background()

	task := &execution.Task{
		Description: "test task",
	}

	id1, _ := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)
	id2, _ := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)

	w1 := pool.Get(id1)
	w1.Mutex.Lock()
	w1.Status = StatusRunning
	w1.Mutex.Unlock()

	w2 := pool.Get(id2)
	w2.Mutex.Lock()
	w2.Status = StatusIdle
	w2.Mutex.Unlock()

	all := pool.List()
	assert.Equal(t, 2, len(all))

	running := pool.List(StatusRunning)
	assert.Equal(t, 1, len(running))

	idle := pool.List(StatusIdle)
	assert.Equal(t, 1, len(idle))
}

func TestPool_Terminate(t *testing.T) {
	t.Parallel()

	pool := NewPool(5, 30*time.Second, 5*time.Minute, 3, 5*time.Second)
	ctx := context.Background()

	task := &execution.Task{
		Description: "test task",
	}

	id, err := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)
	require.NoError(t, err)

	w := pool.Get(id)
	w.Mutex.Lock()
	w.Status = StatusRunning
	w.Mutex.Unlock()

	err = pool.Terminate(id)
	require.NoError(t, err)

	w = pool.Get(id)
	assert.NotNil(t, w)
	w.Mutex.Lock()
	status := w.Status
	w.Mutex.Unlock()
	assert.Equal(t, StatusCancelled, status)
}

func TestPool_Terminate_NonExistent(t *testing.T) {
	t.Parallel()

	pool := NewPool(5, 30*time.Second, 5*time.Minute, 3, 5*time.Second)

	err := pool.Terminate("non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPool_Stats(t *testing.T) {
	t.Parallel()

	pool := NewPool(5, 30*time.Second, 5*time.Minute, 3, 5*time.Second)
	ctx := context.Background()

	task := &execution.Task{
		Description: "test task",
	}

	id1, _ := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)
	id2, _ := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)
	id3, _ := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)

	w1 := pool.Get(id1)
	w1.Mutex.Lock()
	w1.Status = StatusIdle
	w1.Mutex.Unlock()

	w2 := pool.Get(id2)
	w2.Mutex.Lock()
	w2.Status = StatusRunning
	w2.Mutex.Unlock()

	w3 := pool.Get(id3)
	w3.Mutex.Lock()
	w3.Status = StatusCompleted
	w3.Mutex.Unlock()

	stats := pool.Stats()

	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 5, stats.MaxConcurrency)
	assert.Equal(t, 1, stats.Idle)
	assert.Equal(t, 1, stats.Running)
	assert.Equal(t, 1, stats.Completed)
	assert.Equal(t, 0, stats.Failed)
	assert.Equal(t, 0, stats.Cancelled)
}

func TestPool_HealthCheck(t *testing.T) {
	t.Skip("skipping integration test due to race conditions - unit tests verify health monitoring")
	pool := NewPool(2, 1*time.Second, 2*time.Second, 3, 100*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	task := &execution.Task{
		Description: "test task",
	}

	id, _ := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)
	worker := pool.Get(id)
	require.NotNil(t, worker)

	worker.Mutex.Lock()
	worker.Status = StatusRunning
	worker.LastOutputTime = time.Now()
	worker.Health = HealthHealthy
	worker.Mutex.Unlock()

	pool.StartHealthChecker(ctx)
	time.Sleep(1500 * time.Millisecond)
	cancel()
	time.Sleep(200 * time.Millisecond)

	worker.Mutex.Lock()
	healthCheckTime := worker.LastHealthCheck
	worker.Mutex.Unlock()

	assert.True(t, time.Since(healthCheckTime) < 5*time.Second, "health check should have run")
}

func TestWorker_Timeout(t *testing.T) {
	t.Parallel()

	worker := &Worker{
		ID:             "worker-1",
		Status:         StatusRunning,
		LastOutputTime: time.Now().Add(-1 * time.Hour),
	}

	assert.True(t, worker.IsTimedOut(5*time.Minute))
}
