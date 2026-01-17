package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				Method:   MethodACP,
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
				Method:   MethodCLI,
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
				Method:   MethodAPI,
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
				Method:   MethodACP,
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
				Method:   MethodACP,
				Status:   StatusCompleted,
			},
			wantErr: "cannot start, current status: completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
			Method:   MethodACP,
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
			Method:   MethodACP,
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
			Method:   MethodACP,
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
			Method:   MethodACP,
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
			Method:   MethodACP,
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
			Method:   MethodCLI,
			Status:   StatusRunning,
		}

		ctx := context.Background()
		output, err := worker.SendPrompt(ctx, "hello world")

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
			Method:   MethodCLI,
			Status:   StatusIdle,
		}

		ctx := context.Background()
		_, err := worker.SendPrompt(ctx, "hello world")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot send prompt, current status: idle")
	})

	t.Run("rejects prompt to ACP worker", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-3",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   MethodACP,
			Status:   StatusRunning,
		}

		ctx := context.Background()
		_, err := worker.SendPrompt(ctx, "hello world")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "ACP prompt requires ACP client (task 2.1)")
	})

	t.Run("rejects prompt to API worker", func(t *testing.T) {
		worker := &Worker{
			ID:       "worker-4",
			Provider: "anthropic",
			Model:    "claude-3-opus",
			Method:   MethodAPI,
			Status:   StatusRunning,
		}

		ctx := context.Background()
		_, err := worker.SendPrompt(ctx, "hello world")

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
			Method:   MethodACP,
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
			Method:   MethodACP,
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
			Method:   MethodACP,
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
			Method:   MethodACP,
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
			Method:   MethodACP,
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
			Method:     MethodCLI,
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
			Method:   MethodCLI,
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
			Method:   MethodCLI,
			Status:   StatusIdle,
			Task:     task,
		}

		assert.Equal(t, "implement feature", worker.Task.Description)
		assert.Equal(t, "feature", worker.Task.Type)
	})
}
