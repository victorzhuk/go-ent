package worker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victorzhuk/go-ent/internal/opencode"
)

func TestWorkerHandleUpdate(t *testing.T) {
	tests := []struct {
		name     string
		update   opencode.SessionUpdateNotification
		wantType string
		wantData string
	}{
		{
			name: "output update",
			update: opencode.SessionUpdateNotification{
				SessionID: "sess_123",
				PromptID:  "prompt_456",
				Type:      "output",
				Data:      "test output",
			},
			wantData: "test output",
		},
		{
			name: "progress update",
			update: opencode.SessionUpdateNotification{
				SessionID: "sess_123",
				PromptID:  "prompt_456",
				Type:      "progress",
				Progress:  0.75,
				Message:   "75% complete",
			},
			wantType: "progress",
		},
		{
			name: "tool update",
			update: opencode.SessionUpdateNotification{
				SessionID: "sess_123",
				PromptID:  "prompt_456",
				Type:      "tool",
				Tool:      "fs/write_file",
				Status:    "completed",
			},
			wantType: "tool",
		},
		{
			name: "complete update",
			update: opencode.SessionUpdateNotification{
				SessionID: "sess_123",
				PromptID:  "prompt_456",
				Type:      "complete",
			},
			wantType: "complete",
		},
		{
			name: "error update",
			update: opencode.SessionUpdateNotification{
				SessionID: "sess_123",
				PromptID:  "prompt_456",
				Type:      "error",
				Error:     "something went wrong",
			},
			wantType: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Worker{
				ID:     "worker_123",
				Status: StatusIdle,
				Output: "Initial output\n",
			}

			w.handleUpdate(tt.update)

			switch tt.wantType {
			case "complete":
				assert.Equal(t, StatusCompleted, w.Status)
			case "error":
				assert.Equal(t, StatusFailed, w.Status)
			default:
				assert.Equal(t, StatusIdle, w.Status)
			}
		})
	}
}

func TestWorkerHandleUpdateComplete(t *testing.T) {
	w := &Worker{
		ID:     "worker_123",
		Status: StatusRunning,
		Output: "Working...\n",
	}

	update := opencode.SessionUpdateNotification{
		SessionID: "sess_123",
		PromptID:  "prompt_456",
		Type:      "complete",
	}

	w.handleUpdate(update)

	assert.Equal(t, StatusCompleted, w.Status)
	assert.Contains(t, w.Output, "[Complete]")
}

func TestWorkerHandleUpdateError(t *testing.T) {
	w := &Worker{
		ID:     "worker_123",
		Status: StatusRunning,
		Output: "Working...\n",
		Health: HealthHealthy,
	}

	update := opencode.SessionUpdateNotification{
		SessionID: "sess_123",
		PromptID:  "prompt_456",
		Type:      "error",
		Error:     "API error",
	}

	w.handleUpdate(update)

	assert.Equal(t, StatusFailed, w.Status)
	assert.Contains(t, w.Output, "[Error]")
	assert.Contains(t, w.Output, "API error")
	assert.Equal(t, HealthUnhealthy, w.Health)
}

func TestWorkerHandleUpdateCancelled(t *testing.T) {
	w := &Worker{
		ID:     "worker_123",
		Status: StatusRunning,
		Output: "Working...\n",
	}

	update := opencode.SessionUpdateNotification{
		SessionID: "sess_123",
		PromptID:  "prompt_456",
		Type:      "cancelled",
	}

	w.handleUpdate(update)

	assert.Equal(t, StatusCancelled, w.Status)
	assert.Contains(t, w.Output, "[Cancelled]")
}

func TestWorkerHandleUpdateWithOutput(t *testing.T) {
	w := &Worker{
		ID:     "worker_123",
		Status: StatusRunning,
		Output: "",
	}

	update := opencode.SessionUpdateNotification{
		SessionID: "sess_123",
		PromptID:  "prompt_456",
		Type:      "output",
		Data:      "Line 1\n",
	}

	w.handleUpdate(update)

	assert.Contains(t, w.Output, "Line 1")
	assert.NotZero(t, w.LastOutputTime)
	assert.Equal(t, HealthHealthy, w.Health)
}

func TestWorkerHandleMultipleUpdates(t *testing.T) {
	w := &Worker{
		ID:     "worker_123",
		Status: StatusRunning,
		Output: "",
	}

	updates := []opencode.SessionUpdateNotification{
		{
			Type: "output",
			Data: "Step 1\n",
		},
		{
			Type:     "progress",
			Progress: 0.5,
			Message:  "50% done",
		},
		{
			Type:   "tool",
			Tool:   "fs/write",
			Status: "completed",
		},
		{
			Type: "complete",
		},
	}

	for _, update := range updates {
		w.handleUpdate(update)
	}

	assert.Contains(t, w.Output, "Step 1")
	assert.Contains(t, w.Output, "50.0%")
	assert.Contains(t, w.Output, "fs/write")
	assert.Contains(t, w.Output, "[Complete]")
	assert.Equal(t, StatusCompleted, w.Status)
}
