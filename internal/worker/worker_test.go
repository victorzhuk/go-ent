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

func TestWorker_Start(t *testing.T) {
	tests := []struct {
		name       string
		worker     *Worker
		configPath string
		wantErr    string
	}{
		{
			name: "starts ACP worker",
			worker: &Worker{
				ID:       "worker-1",
				Provider: "anthropic",
				Model:    "claude-3-opus",
				Method:   config.MethodACP,
				Status:   StatusIdle,
			},
			wantErr: "", // May fail if opencode not installed
		},
		{
			name: "starts CLI worker",
			worker: &Worker{
				ID:       "worker-2",
				Provider: "openai",
				Model:    "gpt-4",
				Method:   config.MethodCLI,
				Status:   StatusIdle,
			},
			wantErr: "",
		},
		{
			name: "rejects API worker",
			worker: &Worker{
				ID:       "worker-3",
				Provider: "anthropic",
				Model:    "claude-3-opus",
				Method:   config.MethodAPI,
				Status:   StatusIdle,
			},
			wantErr: "API method not implemented",
		},
		{
			name: "rejects running worker",
			worker: &Worker{
				ID:       "worker-4",
				Provider: "anthropic",
				Model:    "claude-3-opus",
				Method:   config.MethodACP,
				Status:   StatusRunning,
			},
			wantErr: "cannot start, current status: running",
		},
		{
			name: "rejects completed worker",
			worker: &Worker{
				ID:       "worker-5",
				Provider: "anthropic",
				Model:    "claude-3-opus",
				Method:   config.MethodACP,
				Status:   StatusCompleted,
			},
			wantErr: "cannot start, current status: completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.worker.Method == config.MethodACP || tt.worker.Method == config.MethodCLI {
				t.Skip("requires opencode binary - skipping integration test")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := tt.worker.Start(ctx, tt.configPath)

			if tt.wantErr == "" {
				if err != nil {
					t.Skip("opencode not available:", err)
				}
				assert.Equal(t, StatusRunning, tt.worker.Status)

				if tt.worker.cmd != nil {
					_ = tt.worker.Stop()
				} else {
					tt.worker.Status = StatusIdle // cleanup for CLI
				}
			} else {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestWorker_Stop(t *testing.T) {
	t.Run("stops running worker", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-1",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Status:   StatusRunning,
		}

		_, cancel := context.WithCancel(context.Background())
		cancel()
		worker.cancel = cancel

		err := worker.Stop()
		assert.NoError(t, err)
		assert.Equal(t, StatusCancelled, worker.Status)
	})

	t.Run("no error stopping idle worker", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-2",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Status:   StatusIdle,
		}

		err := worker.Stop()
		assert.NoError(t, err)
		assert.Equal(t, StatusIdle, worker.Status)
	})

	t.Run("no error stopping completed worker", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-3",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Status:   StatusCompleted,
		}

		err := worker.Stop()
		assert.NoError(t, err)
		assert.Equal(t, StatusCompleted, worker.Status)
	})

	t.Run("no error stopping failed worker", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-4",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Status:   StatusFailed,
		}

		err := worker.Stop()
		assert.NoError(t, err)
		assert.Equal(t, StatusFailed, worker.Status)
	})

	t.Run("no error stopping cancelled worker", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-5",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Status:   StatusCancelled,
		}

		err := worker.Stop()
		assert.NoError(t, err)
		assert.Equal(t, StatusCancelled, worker.Status)
	})
}

func TestWorker_SendPrompt(t *testing.T) {
	t.Run("sends prompt to CLI worker", func(t *testing.T) {
		t.Skip("requires opencode binary")

		worker := &Worker{
			ID:       "worker-1",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodCLI,
			Status:   StatusRunning,
		}

		ctx := context.Background()
		output, err := worker.SendPrompt(ctx, "hello world", 5*time.Minute)

		if err != nil {
			t.Skip("opencode not available:", err)
		}

		assert.NoError(t, err)
		assert.NotEmpty(t, output)
		assert.Equal(t, StatusCompleted, worker.Status)
		assert.Equal(t, output, worker.Output)
	})

	t.Run("rejects prompt to idle worker", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-2",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodCLI,
			Status:   StatusIdle,
		}

		ctx := context.Background()
		_, err := worker.SendPrompt(ctx, "hello world", 5*time.Minute)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot send prompt, current status: idle")
	})

	t.Run("rejects prompt to ACP worker", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-3",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Status:   StatusRunning,
		}

		ctx := context.Background()
		_, err := worker.SendPrompt(ctx, "hello world", 5*time.Minute)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "ACP prompt requires ACP client (task 2.1)")
	})

	t.Run("rejects prompt to API worker", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-4",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodAPI,
			Status:   StatusRunning,
		}

		ctx := context.Background()
		_, err := worker.SendPrompt(ctx, "hello world", 5*time.Minute)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "API method not implemented")
	})
}

func TestWorker_GetStatus(t *testing.T) {
	t.Run("returns idle status", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-1",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Status:   StatusIdle,
		}

		status := worker.GetStatus()
		assert.Equal(t, StatusIdle, status)
	})

	t.Run("returns running status", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-2",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Status:   StatusRunning,
		}

		status := worker.GetStatus()
		assert.Equal(t, StatusRunning, status)
	})

	t.Run("returns completed status", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-3",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Status:   StatusCompleted,
		}

		status := worker.GetStatus()
		assert.Equal(t, StatusCompleted, status)
	})

	t.Run("returns failed status", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-4",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Status:   StatusFailed,
		}

		status := worker.GetStatus()
		assert.Equal(t, StatusFailed, status)
	})

	t.Run("returns cancelled status", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-5",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodACP,
			Status:   StatusCancelled,
		}

		status := worker.GetStatus()
		assert.Equal(t, StatusCancelled, status)
	})
}

func TestWorker_sendCLI(t *testing.T) {
	t.Run("CLI worker with model flag", func(t *testing.T) {
		t.Skip("requires opencode binary")

		worker := &Worker{
			ID:         "worker-1",
			Provider:   "anthropic",
			Model:      "claude-3-opus",
			Method:     config.MethodCLI,
			Status:     StatusRunning,
			configPath: "/tmp/config.json",
		}

		ctx := context.Background()
		output, err := worker.sendCLI(ctx, "test prompt")

		if err != nil {
			t.Skip("opencode not available:", err)
		}

		assert.NoError(t, err)
		assert.NotEmpty(t, output)
		assert.Equal(t, StatusCompleted, worker.Status)
	})

	t.Run("CLI worker without model flag", func(t *testing.T) {
		t.Skip("requires opencode binary")

		worker := &Worker{
			ID:       "worker-2",
			Provider: "anthropic",
			Method:   config.MethodCLI,
			Status:   StatusRunning,
		}

		ctx := context.Background()
		_, err := worker.sendCLI(ctx, "test prompt")

		if err != nil {
			t.Skip("opencode not available:", err)
		}

		assert.NoError(t, err)
		assert.Equal(t, StatusCompleted, worker.Status)
	})
}

func TestWorkerWithTask(t *testing.T) {
	t.Run("worker with execution task", func(t *testing.T) {
		task := execution.NewTask("implement feature")
		task = task.WithType("feature")

		worker := &Worker{
			ID:       "worker-1",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   config.MethodCLI,
			Status:   StatusIdle,
			Task:     task,
		}

		assert.Equal(t, "implement feature", worker.Task.Description)
		assert.Equal(t, "feature", worker.Task.Type)
	})
}

func TestWorker_CheckHealth(t *testing.T) {
	t.Parallel()

	t.Run("unhealthy worker - no process", func(t *testing.T) {
		worker := &Worker{
			ID:             "worker-1",
			Provider:       "anthropic",
			Model:          "claude-3-opus",
			Method:         config.MethodACP,
			Status:         StatusRunning,
			LastOutputTime: time.Now(),
		}

		ctx := context.Background()
		health := worker.CheckHealth(ctx, 5*time.Minute)

		assert.Equal(t, HealthUnhealthy, health)
	})

	t.Run("timed out worker", func(t *testing.T) {
		worker := &Worker{
			ID:             "worker-2",
			Provider:       "anthropic",
			Model:          "claude-3-opus",
			Method:         config.MethodACP,
			Status:         StatusRunning,
			LastOutputTime: time.Now().Add(-1 * time.Hour),
		}

		ctx := context.Background()
		health := worker.CheckHealth(ctx, 5*time.Minute)

		assert.Equal(t, HealthTimeout, health)
		assert.Equal(t, HealthTimeout, worker.Health)
		assert.False(t, worker.UnhealthySince.IsZero())
	})

	t.Run("unhealthy worker - no process", func(t *testing.T) {
		worker := &Worker{
			ID:             "worker-3",
			Provider:       "anthropic",
			Model:          "claude-3-opus",
			Method:         config.MethodACP,
			Status:         StatusRunning,
			LastOutputTime: time.Now(),
		}

		ctx := context.Background()
		health := worker.CheckHealth(ctx, 5*time.Minute)

		assert.Equal(t, HealthUnhealthy, health)
	})

	t.Run("idle worker returns unknown", func(t *testing.T) {
		worker := &Worker{
			ID:             "worker-4",
			Provider:       "anthropic",
			Model:          "claude-3-opus",
			Method:         config.MethodACP,
			Status:         StatusIdle,
			LastOutputTime: time.Now(),
		}

		ctx := context.Background()
		health := worker.CheckHealth(ctx, 5*time.Minute)

		assert.Equal(t, HealthUnknown, health)
	})
}

func TestWorker_IsHealthy(t *testing.T) {
	t.Parallel()

	t.Run("running healthy worker", func(t *testing.T) {
		worker := &Worker{
			ID:     "worker-1",
			Status: StatusRunning,
			Health: HealthHealthy,
		}

		assert.True(t, worker.IsHealthy())
	})

	t.Run("running unhealthy worker", func(t *testing.T) {
		worker := &Worker{
			ID:     "worker-2",
			Status: StatusRunning,
			Health: HealthUnhealthy,
		}

		assert.False(t, worker.IsHealthy())
	})

	t.Run("idle worker", func(t *testing.T) {
		worker := &Worker{
			ID:     "worker-3",
			Status: StatusIdle,
			Health: HealthHealthy,
		}

		assert.False(t, worker.IsHealthy())
	})
}

func TestWorker_IsTimedOut(t *testing.T) {
	t.Parallel()

	t.Run("worker with recent output", func(t *testing.T) {
		worker := &Worker{
			ID:             "worker-1",
			Status:         StatusRunning,
			LastOutputTime: time.Now(),
		}

		assert.False(t, worker.IsTimedOut(5*time.Minute))
	})

	t.Run("worker with old output", func(t *testing.T) {
		worker := &Worker{
			ID:             "worker-2",
			Status:         StatusRunning,
			LastOutputTime: time.Now().Add(-1 * time.Hour),
		}

		assert.True(t, worker.IsTimedOut(5*time.Minute))
	})

	t.Run("idle worker not timed out", func(t *testing.T) {
		worker := &Worker{
			ID:             "worker-3",
			Status:         StatusIdle,
			LastOutputTime: time.Now().Add(-1 * time.Hour),
		}

		assert.False(t, worker.IsTimedOut(5*time.Minute))
	})
}

func TestWorker_RecordOutput(t *testing.T) {
	t.Parallel()

	worker := &Worker{
		ID:             "worker-1",
		Status:         StatusRunning,
		LastOutputTime: time.Now().Add(-1 * time.Hour),
		Health:         HealthUnhealthy,
		UnhealthySince: time.Now().Add(-30 * time.Minute),
	}

	oldTime := worker.LastOutputTime
	time.Sleep(10 * time.Millisecond)

	worker.RecordOutput()

	assert.True(t, worker.LastOutputTime.After(oldTime))
	assert.Equal(t, HealthHealthy, worker.Health)
	assert.True(t, worker.UnhealthySince.IsZero())
}

func TestWorker_UpdateHealth(t *testing.T) {
	t.Parallel()

	t.Run("health status change", func(t *testing.T) {
		worker := &Worker{
			ID:               "worker-1",
			Health:           HealthHealthy,
			HealthCheckCount: 0,
		}

		worker.UpdateHealth(HealthUnhealthy, "process exited")

		assert.Equal(t, HealthUnhealthy, worker.Health)
		assert.Equal(t, 1, worker.HealthCheckCount)
		assert.False(t, worker.UnhealthySince.IsZero())
	})

	t.Run("health status change with reason", func(t *testing.T) {
		worker := &Worker{
			ID:               "worker-2",
			Health:           HealthHealthy,
			HealthCheckCount: 0,
		}

		worker.UpdateHealth(HealthTimeout, "no output for 5 minutes")

		assert.Equal(t, HealthTimeout, worker.Health)
		assert.Equal(t, 1, worker.HealthCheckCount)
	})

	t.Run("health recovery", func(t *testing.T) {
		worker := &Worker{
			ID:             "worker-3",
			Health:         HealthUnhealthy,
			UnhealthySince: time.Now().Add(-1 * time.Hour),
		}

		worker.UpdateHealth(HealthHealthy, "output received")

		assert.Equal(t, HealthHealthy, worker.Health)
		assert.True(t, worker.UnhealthySince.IsZero())
	})
}

func TestWorker_RetryCount(t *testing.T) {
	t.Parallel()

	t.Run("increment retry count", func(t *testing.T) {
		worker := &Worker{
			ID:         "worker-1",
			RetryCount: 0,
		}

		count := worker.IncrementRetryCount()
		assert.Equal(t, 1, count)
		assert.Equal(t, 1, worker.RetryCount)

		count = worker.IncrementRetryCount()
		assert.Equal(t, 2, count)
	})

	t.Run("reset retry count", func(t *testing.T) {
		worker := &Worker{
			ID:         "worker-2",
			RetryCount: 5,
		}

		worker.ResetRetryCount()
		assert.Equal(t, 0, worker.RetryCount)
	})

	t.Run("should retry", func(t *testing.T) {
		worker := &Worker{
			ID:         "worker-3",
			RetryCount: 1,
		}

		assert.True(t, worker.ShouldRetry(3))
		assert.False(t, worker.ShouldRetry(1))
	})
}

func TestHealthStatus_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   HealthStatus
		expected string
	}{
		{"healthy", HealthHealthy, "healthy"},
		{"unhealthy", HealthUnhealthy, "unhealthy"},
		{"unknown", HealthUnknown, "unknown"},
		{"timeout", HealthTimeout, "timeout"},
		{"empty", HealthStatus(""), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestHealthStatus_Valid(t *testing.T) {
	t.Parallel()

	assert.True(t, HealthHealthy.Valid())
	assert.True(t, HealthUnhealthy.Valid())
	assert.True(t, HealthUnknown.Valid())
	assert.True(t, HealthTimeout.Valid())
	assert.False(t, HealthStatus("invalid").Valid())
}
