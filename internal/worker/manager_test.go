package worker

//nolint:gosec // test file with necessary file operations

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

func TestWorkerManager_New(t *testing.T) {
	t.Run("creates manager with defaults", func(t *testing.T) {
		mgr := NewWorkerManagerWithoutTracking()

		assert.NotNil(t, mgr)
		assert.NotNil(t, mgr.workers)
		assert.NotNil(t, mgr.pool)
	})
}

func TestWorkerManager_Spawn(t *testing.T) {
	t.Run("spawns worker with ACP method", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()
		task := execution.NewTask("test task")

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
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
		assert.Equal(t, config.MethodACP, worker.Method)
		assert.Equal(t, StatusIdle, worker.Status)
		assert.NotNil(t, worker.Task)
	})

	t.Run("spawns worker with CLI method", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()
		task := execution.NewTask("test task")

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "openai",
			Model:    "gpt-4",
			Method:   config.MethodCLI,
			Task:     task,
			Timeout:  5 * time.Minute,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, workerID)

		worker := mgr.Get(workerID)
		require.NotNil(t, worker)
		assert.Equal(t, config.MethodCLI, worker.Method)
	})

	t.Run("spawns worker with API method", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()
		task := execution.NewTask("test task")

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodAPI,
			Task:     task,
			Timeout:  5 * time.Minute,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, workerID)

		worker := mgr.Get(workerID)
		require.NotNil(t, worker)
		assert.Equal(t, config.MethodAPI, worker.Method)
	})

	t.Run("rejects spawn with existing worker ID", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()
		task := execution.NewTask("test task")

		workerID := "existing-worker-id"
		mgr.workers[workerID] = &Worker{ID: workerID}

		_, err := mgr.Spawn(ctx, SpawnRequest{
			WorkerID: workerID,
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Task:     task,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("rejects spawn with invalid method", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()
		task := execution.NewTask("test task")

		_, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.CommunicationMethod("invalid"),
			Task:     task,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})
}

func TestWorkerManager_Get(t *testing.T) {
	t.Run("returns nil for non-existent worker", func(t *testing.T) {
		mgr := NewWorkerManagerWithoutTracking()

		worker := mgr.Get("non-existent")
		assert.Nil(t, worker)
	})

	t.Run("returns existing worker", func(t *testing.T) {
		mgr := NewWorkerManagerWithoutTracking()
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
		mgr := NewWorkerManagerWithoutTracking()
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
		mgr := NewWorkerManagerWithoutTracking()
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
		mgr := NewWorkerManagerWithoutTracking()

		err := mgr.Cancel(ctx, "non-existent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestWorkerManager_List(t *testing.T) {
	t.Run("returns empty list when no workers", func(t *testing.T) {
		mgr := NewWorkerManagerWithoutTracking()

		workers := mgr.List()
		assert.Empty(t, workers)
	})

	t.Run("returns all workers", func(t *testing.T) {
		mgr := NewWorkerManagerWithoutTracking()
		mgr.workers["worker-1"] = &Worker{ID: "worker-1", Status: StatusIdle}
		mgr.workers["worker-2"] = &Worker{ID: "worker-2", Status: StatusRunning}
		mgr.workers["worker-3"] = &Worker{ID: "worker-3", Status: StatusCompleted}

		workers := mgr.List()
		assert.Len(t, workers, 3)
	})

	t.Run("filters by status", func(t *testing.T) {
		mgr := NewWorkerManagerWithoutTracking()
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
		mgr := NewWorkerManagerWithoutTracking()
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
		mgr := NewWorkerManagerWithoutTracking()
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
		mgr := NewWorkerManagerWithoutTracking()
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
		mgr := NewWorkerManagerWithoutTracking()

		assert.Panics(t, func() {
			mgr.SetWorkerStatus("non-existent", StatusRunning)
		})
	})
}

func TestWorkerManager_GetStatus(t *testing.T) {
	t.Run("returns worker status", func(t *testing.T) {
		mgr := NewWorkerManagerWithoutTracking()
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
		mgr := NewWorkerManagerWithoutTracking()

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
		method   config.CommunicationMethod
		expected string
	}{
		{config.MethodACP, "acp"},
		{config.MethodCLI, "cli"},
		{config.MethodAPI, "api"},
		{config.CommunicationMethod("unknown"), "unknown"},
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
		method   config.CommunicationMethod
		expected bool
	}{
		{config.MethodACP, true},
		{config.MethodCLI, true},
		{config.MethodAPI, true},
		{config.CommunicationMethod("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.method.String(), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.method.Valid())
		})
	}
}

func TestWorkerManager_Spawn_ConfigPath(t *testing.T) {
	t.Run("spawns worker with config path", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()
		task := execution.NewTask("test task")

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider:           "anthropic",
			Model:              "claude-3-opus",
			Method:             config.MethodACP,
			Task:               task,
			OpenCodeConfigPath: "/tmp/config.json",
		})

		require.NoError(t, err)
		assert.NotEmpty(t, workerID)

		worker := mgr.Get(workerID)
		require.NotNil(t, worker)
		assert.Equal(t, "/tmp/config.json", worker.configPath)
	})

	t.Run("spawns worker without config path", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()
		task := execution.NewTask("test task")

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Task:     task,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, workerID)

		worker := mgr.Get(workerID)
		require.NotNil(t, worker)
		assert.Empty(t, worker.configPath)
	})
}

func TestWorkerManager_Spawn_AutoWorkerID(t *testing.T) {
	t.Run("generates UUID v7 worker ID", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()
		task := execution.NewTask("test task")

		workerID1, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Task:     task,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, workerID1)

		workerID2, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Task:     task,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, workerID2)
		assert.NotEqual(t, workerID1, workerID2)
	})
}

func TestWorkerManager_SendPrompt(t *testing.T) {
	t.Run("rejects prompt for non-existent worker", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()

		_, err := mgr.SendPrompt(ctx, PromptRequest{
			WorkerID: "non-existent",
			Prompt:   "test prompt",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("rejects prompt for invalid status worker", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()
		task := execution.NewTask("test task")

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Task:     task,
		})
		require.NoError(t, err)

		worker := mgr.Get(workerID)
		worker.Status = StatusCompleted

		_, err = mgr.SendPrompt(ctx, PromptRequest{
			WorkerID: workerID,
			Prompt:   "test prompt",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot accept prompts")
	})

	t.Run("rejects prompt for CLI worker", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()
		task := execution.NewTask("test task")

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodCLI,
			Task:     task,
		})
		require.NoError(t, err)

		_, err = mgr.SendPrompt(ctx, PromptRequest{
			WorkerID: workerID,
			Prompt:   "test prompt",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not support prompting")
	})

	t.Run("rejects prompt for API worker", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()
		task := execution.NewTask("test task")

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodAPI,
			Task:     task,
		})
		require.NoError(t, err)

		_, err = mgr.SendPrompt(ctx, PromptRequest{
			WorkerID: workerID,
			Prompt:   "test prompt",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not support prompting")
	})
}

func TestWorkerManager_GetOutput(t *testing.T) {
	t.Run("gets output from worker", func(t *testing.T) {
		mgr := NewWorkerManagerWithoutTracking()
		workerID := "test-worker"

		mgr.workers[workerID] = &Worker{
			ID:     workerID,
			Status: StatusRunning,
			Output: "Line 1\nLine 2\nLine 3\n",
			Mutex:  sync.Mutex{},
		}

		resp, err := mgr.GetOutput(WorkerOutputRequest{
			WorkerID: workerID,
		})

		require.NoError(t, err)
		assert.Equal(t, workerID, resp.WorkerID)
		assert.Equal(t, "Line 1\nLine 2\nLine 3", resp.Output)
		assert.Equal(t, 3, resp.LineCount)
		assert.False(t, resp.Truncated)
	})

	t.Run("returns error for non-existent worker", func(t *testing.T) {
		mgr := NewWorkerManagerWithoutTracking()

		_, err := mgr.GetOutput(WorkerOutputRequest{
			WorkerID: "non-existent",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("filters output since time", func(t *testing.T) {
		mgr := NewWorkerManagerWithoutTracking()
		workerID := "test-worker"
		now := time.Now()

		mgr.workers[workerID] = &Worker{
			ID:             workerID,
			Status:         StatusRunning,
			Output:         "Old output\nNew output\n",
			Mutex:          sync.Mutex{},
			LastOutputTime: now.Add(-1 * time.Hour),
		}

		resp, err := mgr.GetOutput(WorkerOutputRequest{
			WorkerID: workerID,
			Since:    now.Add(-30 * time.Minute),
		})

		require.NoError(t, err)
		assert.Empty(t, resp.Output)
	})

	t.Run("filters output by regex", func(t *testing.T) {
		mgr := NewWorkerManagerWithoutTracking()
		workerID := "test-worker"

		mgr.workers[workerID] = &Worker{
			ID:     workerID,
			Status: StatusRunning,
			Output: "error: something failed\ninfo: processing\nwarning: timeout\n",
			Mutex:  sync.Mutex{},
		}

		resp, err := mgr.GetOutput(WorkerOutputRequest{
			WorkerID: workerID,
			Filter:   "error|warning",
		})

		require.NoError(t, err)
		assert.Contains(t, resp.Output, "error: something failed")
		assert.Contains(t, resp.Output, "warning: timeout")
		assert.NotContains(t, resp.Output, "info: processing")
	})

	t.Run("limits output lines", func(t *testing.T) {
		mgr := NewWorkerManagerWithoutTracking()
		workerID := "test-worker"

		output := ""
		for i := 1; i <= 10; i++ {
			output += fmt.Sprintf("Line %d\n", i)
		}

		mgr.workers[workerID] = &Worker{
			ID:     workerID,
			Status: StatusRunning,
			Output: output,
			Mutex:  sync.Mutex{},
		}

		resp, err := mgr.GetOutput(WorkerOutputRequest{
			WorkerID: workerID,
			Limit:    5,
		})

		require.NoError(t, err)
		assert.Equal(t, 10, resp.LineCount)
		assert.True(t, resp.Truncated)
		assert.Contains(t, resp.Output, "Line 1")
		assert.Contains(t, resp.Output, "Line 5")
		assert.NotContains(t, resp.Output, "Line 6")
	})

	t.Run("empty output", func(t *testing.T) {
		mgr := NewWorkerManagerWithoutTracking()
		workerID := "test-worker"

		mgr.workers[workerID] = &Worker{
			ID:     workerID,
			Status: StatusIdle,
			Output: "",
			Mutex:  sync.Mutex{},
		}

		resp, err := mgr.GetOutput(WorkerOutputRequest{
			WorkerID: workerID,
		})

		require.NoError(t, err)
		assert.Equal(t, "", resp.Output)
		assert.Equal(t, 0, resp.LineCount)
	})
}

func TestWorkerManager_Concurrency(t *testing.T) {
	t.Run("concurrent spawns", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()

		var wg sync.WaitGroup
		workerIDs := make(chan string, 100)
		errors := make(chan error, 100)

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				task := execution.NewTask(fmt.Sprintf("task %d", n))

				id, err := mgr.Spawn(ctx, SpawnRequest{
					Provider: fmt.Sprintf("provider-%d", n),
					Model:    fmt.Sprintf("model-%d", n),
					Method:   config.MethodACP,
					Task:     task,
				})

				if err != nil {
					errors <- err
					return
				}
				workerIDs <- id
			}(i)
		}

		wg.Wait()
		close(workerIDs)
		close(errors)

		for err := range errors {
			t.Fatalf("spawn failed: %v", err)
		}

		count := 0
		for range workerIDs {
			count++
		}

		assert.Equal(t, 100, count)
		assert.Equal(t, 100, len(mgr.List()))
	})

	t.Run("concurrent reads and writes", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Task:     execution.NewTask("test task"),
		})
		require.NoError(t, err)

		var wg sync.WaitGroup
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				switch i % 4 {
				case 0:
					mgr.Get(workerID)
				case 1:
					mgr.List()
				case 2:
					_, _ = mgr.GetStatus(workerID)
				case 3:
					mgr.List(StatusRunning)
				}
			}()
		}

		wg.Wait()
		close(errors)

		for err := range errors {
			t.Errorf("concurrent operation failed: %v", err)
		}
	})
}

func TestWorkerManager_EdgeCases(t *testing.T) {
	t.Run("empty provider name", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Task:     execution.NewTask("test task"),
		})

		require.NoError(t, err)
		assert.NotEmpty(t, workerID)

		worker := mgr.Get(workerID)
		assert.Empty(t, worker.Provider)
	})

	t.Run("empty model name", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "",
			Method:   config.MethodACP,
			Task:     execution.NewTask("test task"),
		})

		require.NoError(t, err)
		assert.NotEmpty(t, workerID)

		worker := mgr.Get(workerID)
		assert.Empty(t, worker.Model)
	})

	t.Run("nil task", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Task:     nil,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, workerID)

		worker := mgr.Get(workerID)
		assert.Nil(t, worker.Task)
	})

	t.Run("list with multiple status filters", func(t *testing.T) {
		mgr := NewWorkerManagerWithoutTracking()
		mgr.workers["worker-1"] = &Worker{ID: "worker-1", Status: StatusIdle}
		mgr.workers["worker-2"] = &Worker{ID: "worker-2", Status: StatusRunning}
		mgr.workers["worker-3"] = &Worker{ID: "worker-3", Status: StatusCompleted}

		workers := mgr.List()
		assert.Len(t, workers, 3)

		running := mgr.List(StatusRunning)
		assert.Len(t, running, 1)

		idle := mgr.List(StatusIdle)
		assert.Len(t, idle, 1)
	})
}

func TestWorkerManager_Lifecycle(t *testing.T) {
	t.Run("worker status transitions", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()
		task := execution.NewTask("test task")

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Task:     task,
		})
		require.NoError(t, err)

		worker := mgr.Get(workerID)
		assert.Equal(t, StatusIdle, worker.Status)

		mgr.SetWorkerStatus(workerID, StatusRunning)
		assert.Equal(t, StatusRunning, worker.Status)

		mgr.SetWorkerStatus(workerID, StatusCompleted)
		assert.Equal(t, StatusCompleted, worker.Status)
	})

	t.Run("worker not tracked after cleanup", func(t *testing.T) {
		ctx := context.Background()
		mgr := NewWorkerManagerWithoutTracking()
		task := execution.NewTask("test task")

		workerID, err := mgr.Spawn(ctx, SpawnRequest{
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Task:     task,
		})
		require.NoError(t, err)

		worker := mgr.Get(workerID)
		require.NotNil(t, worker)
		worker.Status = StatusCompleted
		worker.StartedAt = time.Now().Add(-2 * time.Hour)

		count := mgr.Cleanup(1 * time.Hour)
		assert.Equal(t, 1, count)

		worker = mgr.Get(workerID)
		assert.Nil(t, worker)
	})
}
