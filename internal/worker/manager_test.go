package worker

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/execution"
)

func TestWorkerManager_New(t *testing.T) {
	t.Run("creates manager with defaults", func(t *testing.T) {
		mgr := NewWorkerManager()

		assert.NotNil(t, mgr)
		assert.NotNil(t, mgr.workers)
		assert.NotNil(t, mgr.pool)
	})
}

func TestWorkerManager_Spawn(t *testing.T) {
	t.Run("spawns worker with ACP method", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManager()
		task := execution.NewTask("test task")

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   MethodACP,
			Task:     task,
			Timeout:  5 * time.Minute,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, workerID)

		worker := mgr.Get(workerID)
		require.NotNil(t, worker)
		assert.Equal(t, workerID, worker.ID)
		assert.Equal(t, "anthropic", worker.Provider)
		assert.Equal(t, "claude-3-opus", worker.Model)
		assert.Equal(t, MethodACP, worker.Method)
		assert.Equal(t, StatusIdle, worker.Status)
		assert.NotNil(t, worker.Task)
	})

	t.Run("spawns worker with CLI method", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManager()
		task := execution.NewTask("test task")

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "openai",
			Model:    "gpt-4",
			Method:   MethodCLI,
			Task:     task,
			Timeout:  5 * time.Minute,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, workerID)

		worker := mgr.Get(workerID)
		require.NotNil(t, worker)
		assert.Equal(t, MethodCLI, worker.Method)
	})

	t.Run("spawns worker with API method", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManager()
		task := execution.NewTask("test task")

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   MethodAPI,
			Task:     task,
			Timeout:  5 * time.Minute,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, workerID)

		worker := mgr.Get(workerID)
		require.NotNil(t, worker)
		assert.Equal(t, MethodAPI, worker.Method)
	})

	t.Run("rejects spawn with existing worker ID", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManager()
		task := execution.NewTask("test task")

		workerID := "existing-worker-id"
		mgr.workers[workerID] = &Worker{ID: workerID}

		_, err := mgr.Spawn(ctx, SpawnRequest{
			WorkerID: workerID,
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   MethodACP,
			Task:     task,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("rejects spawn with invalid method", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManager()
		task := execution.NewTask("test task")

		_, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   CommunicationMethod("invalid"),
			Task:     task,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})
}

func TestWorkerManager_Get(t *testing.T) {
	t.Run("returns nil for non-existent worker", func(t *testing.T) {
		mgr := NewWorkerManager()

		worker := mgr.Get("non-existent")
		assert.Nil(t, worker)
	})

	t.Run("returns existing worker", func(t *testing.T) {
		mgr := NewWorkerManager()
		workerID := "test-worker"
		expectedWorker := &Worker{
			ID:       workerID,
			Provider: "anthropic",
			Status:   StatusIdle,
		}
		mgr.workers[workerID] = expectedWorker

		worker := mgr.Get(workerID)
		assert.Equal(t, expectedWorker, worker)
	})
}

func TestWorkerManager_Cancel(t *testing.T) {
	t.Run("cancels idle worker", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManager()
		workerID := "test-worker"

		mgr.workers[workerID] = &Worker{
			ID:        workerID,
			Status:    StatusIdle,
			StartedAt: time.Now(),
		}

		err := mgr.Cancel(ctx, workerID)
		require.NoError(t, err)

		worker := mgr.Get(workerID)
		assert.Equal(t, StatusCancelled, worker.Status)
	})

	t.Run("cancels running worker", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManager()
		workerID := "test-worker"

		mgr.workers[workerID] = &Worker{
			ID:        workerID,
			Status:    StatusRunning,
			StartedAt: time.Now(),
		}

		err := mgr.Cancel(ctx, workerID)
		require.NoError(t, err)

		worker := mgr.Get(workerID)
		assert.Equal(t, StatusCancelled, worker.Status)
	})

	t.Run("returns error for non-existent worker", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManager()

		err := mgr.Cancel(ctx, "non-existent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestWorkerManager_List(t *testing.T) {
	t.Run("returns empty list when no workers", func(t *testing.T) {
		mgr := NewWorkerManager()

		workers := mgr.List()
		assert.Empty(t, workers)
	})

	t.Run("returns all workers", func(t *testing.T) {
		mgr := NewWorkerManager()
		mgr.workers["worker-1"] = &Worker{ID: "worker-1", Status: StatusIdle}
		mgr.workers["worker-2"] = &Worker{ID: "worker-2", Status: StatusRunning}
		mgr.workers["worker-3"] = &Worker{ID: "worker-3", Status: StatusCompleted}

		workers := mgr.List()
		assert.Len(t, workers, 3)
	})

	t.Run("filters by status", func(t *testing.T) {
		mgr := NewWorkerManager()
		mgr.workers["worker-1"] = &Worker{ID: "worker-1", Status: StatusIdle}
		mgr.workers["worker-2"] = &Worker{ID: "worker-2", Status: StatusRunning}
		mgr.workers["worker-3"] = &Worker{ID: "worker-3", Status: StatusCompleted}

		running := mgr.List(StatusRunning)
		assert.Len(t, running, 1)
		assert.Equal(t, "worker-2", running[0].ID)
	})
}

func TestWorkerManager_Cleanup(t *testing.T) {
	t.Run("removes completed workers", func(t *testing.T) {
		mgr := NewWorkerManager()
		now := time.Now()

		mgr.workers["worker-1"] = &Worker{ID: "worker-1", Status: StatusIdle}
		mgr.workers["worker-2"] = &Worker{ID: "worker-2", Status: StatusCompleted, StartedAt: now.Add(-2 * time.Hour)}
		mgr.workers["worker-3"] = &Worker{ID: "worker-3", Status: StatusFailed, StartedAt: now.Add(-2 * time.Hour)}
		mgr.workers["worker-4"] = &Worker{ID: "worker-4", Status: StatusCancelled, StartedAt: now.Add(-2 * time.Hour)}

		count := mgr.Cleanup()
		assert.Equal(t, 3, count)

		workers := mgr.List()
		assert.Len(t, workers, 1)
		assert.Equal(t, "worker-1", workers[0].ID)
	})

	t.Run("respects maxAge parameter", func(t *testing.T) {
		mgr := NewWorkerManager()
		now := time.Now()

		mgr.workers["old-worker"] = &Worker{
			ID:        "old-worker",
			Status:    StatusCompleted,
			StartedAt: now.Add(-2 * time.Hour),
		}
		mgr.workers["new-worker"] = &Worker{
			ID:        "new-worker",
			Status:    StatusCompleted,
			StartedAt: now.Add(-30 * time.Minute),
		}

		count := mgr.Cleanup(1 * time.Hour)
		assert.Equal(t, 1, count)

		workers := mgr.List()
		assert.Len(t, workers, 1)
		assert.Equal(t, "new-worker", workers[0].ID)
	})
}

func TestWorkerManager_SetWorkerStatus(t *testing.T) {
	t.Run("updates worker status", func(t *testing.T) {
		mgr := NewWorkerManager()
		workerID := "test-worker"

		mgr.workers[workerID] = &Worker{
			ID:     workerID,
			Status: StatusIdle,
		}

		mgr.SetWorkerStatus(workerID, StatusRunning)

		worker := mgr.Get(workerID)
		assert.Equal(t, StatusRunning, worker.Status)
	})

	t.Run("panics for non-existent worker", func(t *testing.T) {
		mgr := NewWorkerManager()

		assert.Panics(t, func() {
			mgr.SetWorkerStatus("non-existent", StatusRunning)
		})
	})
}

func TestWorkerManager_GetStatus(t *testing.T) {
	t.Run("returns worker status", func(t *testing.T) {
		mgr := NewWorkerManager()
		workerID := "test-worker"

		mgr.workers[workerID] = &Worker{
			ID:     workerID,
			Status: StatusRunning,
		}

		status, err := mgr.GetStatus(workerID)
		require.NoError(t, err)
		assert.Equal(t, StatusRunning, status)
	})

	t.Run("returns error for non-existent worker", func(t *testing.T) {
		mgr := NewWorkerManager()

		status, err := mgr.GetStatus("non-existent")
		require.Error(t, err)
		assert.Equal(t, WorkerStatus(""), status)
	})
}

func TestWorkerStatus_String(t *testing.T) {
	tests := []struct {
		status   WorkerStatus
		expected string
	}{
		{StatusIdle, "idle"},
		{StatusRunning, "running"},
		{StatusCompleted, "completed"},
		{StatusFailed, "failed"},
		{StatusCancelled, "cancelled"},
		{WorkerStatus("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestCommunicationMethod_String(t *testing.T) {
	tests := []struct {
		method   CommunicationMethod
		expected string
	}{
		{MethodACP, "acp"},
		{MethodCLI, "cli"},
		{MethodAPI, "api"},
		{CommunicationMethod("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.method.String())
		})
	}
}

func TestWorkerStatus_Valid(t *testing.T) {
	tests := []struct {
		status   WorkerStatus
		expected bool
	}{
		{StatusIdle, true},
		{StatusRunning, true},
		{StatusCompleted, true},
		{StatusFailed, true},
		{StatusCancelled, true},
		{WorkerStatus("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.Valid())
		})
	}
}

func TestCommunicationMethod_Valid(t *testing.T) {
	tests := []struct {
		method   CommunicationMethod
		expected bool
	}{
		{MethodACP, true},
		{MethodCLI, true},
		{MethodAPI, true},
		{CommunicationMethod("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.method.String(), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.method.Valid())
		})
	}
}
