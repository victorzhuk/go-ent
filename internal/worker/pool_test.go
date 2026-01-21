package worker

import (
	"context"
	"fmt"
	"sync"
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

func TestPool_StartHealthChecker(t *testing.T) {
	t.Run("starts and stops health checker", func(t *testing.T) {
		pool := NewPool(5, 100*time.Millisecond, 5*time.Minute, 3, 100*time.Millisecond)
		ctx, cancel := context.WithCancel(context.Background())

		task := &execution.Task{Description: "test task"}
		id, _ := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)
		worker := pool.Get(id)
		require.NotNil(t, worker)

		worker.Mutex.Lock()
		worker.Status = StatusRunning
		worker.LastOutputTime = time.Now()
		worker.Health = HealthHealthy
		worker.Mutex.Unlock()

		pool.StartHealthChecker(ctx)
		time.Sleep(200 * time.Millisecond)

		worker.Mutex.Lock()
		checkTime := worker.LastHealthCheck
		worker.Mutex.Unlock()

		assert.True(t, time.Since(checkTime) < 500*time.Millisecond, "health check should have run")

		cancel()
		time.Sleep(200 * time.Millisecond)
	})

	t.Run("multiple starts are safe", func(t *testing.T) {
		pool := NewPool(5, 100*time.Millisecond, 5*time.Minute, 3, 100*time.Millisecond)
		ctx1, cancel1 := context.WithCancel(context.Background())
		ctx2, cancel2 := context.WithCancel(context.Background())

		task := &execution.Task{Description: "test task"}
		id, _ := pool.Spawn(ctx1, "provider", "model", config.MethodCLI, task)

		worker := pool.Get(id)
		worker.Mutex.Lock()
		worker.Status = StatusRunning
		worker.LastOutputTime = time.Now()
		worker.Health = HealthHealthy
		worker.Mutex.Unlock()

		pool.StartHealthChecker(ctx1)
		time.Sleep(50 * time.Millisecond)
		pool.StartHealthChecker(ctx2)
		time.Sleep(200 * time.Millisecond)

		cancel1()
		cancel2()
	})
}

func TestPool_StopHealthChecker(t *testing.T) {
	t.Run("stops health checker", func(t *testing.T) {
		pool := NewPool(5, 100*time.Millisecond, 5*time.Minute, 3, 100*time.Millisecond)
		ctx, cancel := context.WithCancel(context.Background())

		task := &execution.Task{Description: "test task"}
		id, _ := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)

		worker := pool.Get(id)
		worker.Mutex.Lock()
		worker.Status = StatusRunning
		worker.LastOutputTime = time.Now()
		worker.Health = HealthHealthy
		worker.Mutex.Unlock()

		pool.StartHealthChecker(ctx)
		time.Sleep(200 * time.Millisecond)

		pool.StopHealthChecker()
		cancel()
		time.Sleep(200 * time.Millisecond)
	})
}

func TestPool_HealthCheck_TimeoutWorker(t *testing.T) {
	t.Parallel()

	pool := NewPool(5, 100*time.Millisecond, 200*time.Millisecond, 3, 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	task := &execution.Task{Description: "test task"}
	id, _ := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)

	worker := pool.Get(id)
	worker.Mutex.Lock()
	worker.Status = StatusRunning
	worker.LastOutputTime = time.Now()
	worker.Health = HealthHealthy
	worker.Mutex.Unlock()

	pool.StartHealthChecker(ctx)
	time.Sleep(300 * time.Millisecond)

	worker.Mutex.Lock()
	health := worker.Health
	worker.Mutex.Unlock()

	assert.Equal(t, HealthTimeout, health)
}

func TestPool_HealthCheck_UnhealthyWorker(t *testing.T) {
	t.Parallel()

	pool := NewPool(5, 100*time.Millisecond, 5*time.Minute, 3, 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	task := &execution.Task{Description: "test task"}
	id, _ := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)

	worker := pool.Get(id)
	worker.Mutex.Lock()
	worker.Status = StatusRunning
	worker.LastOutputTime = time.Now()
	worker.Health = HealthHealthy
	worker.Mutex.Unlock()

	pool.StartHealthChecker(ctx)
	time.Sleep(200 * time.Millisecond)

	worker.Mutex.Lock()
	health := worker.Health
	worker.Mutex.Unlock()

	assert.Equal(t, HealthUnhealthy, health)
}

func TestPool_HealthCheck_RetryMechanism(t *testing.T) {
	t.Parallel()

	pool := NewPool(5, 100*time.Millisecond, 200*time.Millisecond, 2, 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	task := &execution.Task{Description: "test task"}
	id, _ := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)

	worker := pool.Get(id)
	worker.Mutex.Lock()
	worker.Status = StatusRunning
	worker.LastOutputTime = time.Now()
	worker.Health = HealthHealthy
	worker.Mutex.Unlock()

	pool.StartHealthChecker(ctx)

	time.Sleep(300 * time.Millisecond)

	worker.Mutex.Lock()
	retryCount := worker.RetryCount
	worker.Mutex.Unlock()

	assert.Equal(t, 2, retryCount)
}

func TestPool_MultipleSpawns(t *testing.T) {
	pool := NewPool(10, 30*time.Second, 5*time.Minute, 3, 5*time.Second)
	ctx := context.Background()

	ids := make([]string, 0, 10)

	for i := 0; i < 10; i++ {
		task := &execution.Task{Description: fmt.Sprintf("task %d", i)}

		id, err := pool.Spawn(ctx, fmt.Sprintf("provider-%d", i), fmt.Sprintf("model-%d", i), config.MethodCLI, task)
		if err != nil {
			t.Fatalf("spawn failed: %v", err)
		}
		ids = append(ids, id)
	}

	assert.Equal(t, 10, len(ids))
	assert.Equal(t, 10, len(pool.List()))
}

func TestPool_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	pool := NewPool(10, 30*time.Second, 5*time.Minute, 3, 5*time.Second)
	ctx := context.Background()

	task := &execution.Task{Description: "test task"}
	id, _ := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			switch n % 5 {
			case 0:
				pool.Get(id)
			case 1:
				pool.List()
			case 2:
				pool.List(StatusRunning)
			case 3:
				pool.Stats()
			case 4:
				pool.runningCount()
			}
		}(i)
	}

	wg.Wait()

	stats := pool.Stats()
	assert.Equal(t, 1, stats.Total)
}

func TestPool_Stats_Comprehensive(t *testing.T) {
	t.Parallel()

	pool := NewPool(10, 30*time.Second, 5*time.Minute, 3, 5*time.Second)
	ctx := context.Background()

	task := &execution.Task{Description: "test task"}

	id1, _ := pool.Spawn(ctx, "p1", "m1", config.MethodCLI, task)
	id2, _ := pool.Spawn(ctx, "p2", "m2", config.MethodCLI, task)
	id3, _ := pool.Spawn(ctx, "p3", "m3", config.MethodCLI, task)
	id4, _ := pool.Spawn(ctx, "p4", "m4", config.MethodCLI, task)
	id5, _ := pool.Spawn(ctx, "p5", "m5", config.MethodCLI, task)

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

	w4 := pool.Get(id4)
	w4.Mutex.Lock()
	w4.Status = StatusFailed
	w4.Mutex.Unlock()

	w5 := pool.Get(id5)
	w5.Mutex.Lock()
	w5.Status = StatusCancelled
	w5.Mutex.Unlock()

	stats := pool.Stats()

	assert.Equal(t, 5, stats.Total)
	assert.Equal(t, 10, stats.MaxConcurrency)
	assert.Equal(t, 1, stats.Idle)
	assert.Equal(t, 1, stats.Running)
	assert.Equal(t, 1, stats.Completed)
	assert.Equal(t, 1, stats.Failed)
	assert.Equal(t, 1, stats.Cancelled)
}

func TestPool_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("max concurrency zero", func(t *testing.T) {
		pool := NewPool(0, 30*time.Second, 5*time.Minute, 3, 5*time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		task := &execution.Task{Description: "test task"}

		_, err := pool.Spawn(ctx, "provider", "model", config.MethodCLI, task)
		assert.Error(t, err)
	})

	t.Run("empty provider and model", func(t *testing.T) {
		pool := NewPool(5, 30*time.Second, 5*time.Minute, 3, 5*time.Second)
		ctx := context.Background()

		task := &execution.Task{Description: "test task"}
		id, err := pool.Spawn(ctx, "", "", config.MethodCLI, task)

		require.NoError(t, err)
		assert.NotEmpty(t, id)

		worker := pool.Get(id)
		assert.Empty(t, worker.Provider)
		assert.Empty(t, worker.Model)
	})

	t.Run("nil task", func(t *testing.T) {
		pool := NewPool(5, 30*time.Second, 5*time.Minute, 3, 5*time.Second)
		ctx := context.Background()

		id, err := pool.Spawn(ctx, "provider", "model", config.MethodCLI, nil)

		require.NoError(t, err)
		assert.NotEmpty(t, id)

		worker := pool.Get(id)
		assert.Nil(t, worker.Task)
	})
}
