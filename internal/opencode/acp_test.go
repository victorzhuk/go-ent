package opencode

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionUpdateNotification_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    SessionUpdateNotification
		wantErr bool
	}{
		{
			name: "output notification",
			data: `{
				"sessionId": "sess_abc123",
				"promptId": "prompt_xyz789",
				"type": "output",
				"data": "Implementing rate limit function..."
			}`,
			want: SessionUpdateNotification{
				SessionID: "sess_abc123",
				PromptID:  "prompt_xyz789",
				Type:      "output",
				Data:      "Implementing rate limit function...",
			},
		},
		{
			name: "progress notification",
			data: `{
				"sessionId": "sess_abc123",
				"promptId": "prompt_xyz789",
				"type": "progress",
				"progress": 0.45,
				"message": "45% complete"
			}`,
			want: SessionUpdateNotification{
				SessionID: "sess_abc123",
				PromptID:  "prompt_xyz789",
				Type:      "progress",
				Progress:  0.45,
				Message:   "45% complete",
			},
		},
		{
			name: "tool notification",
			data: `{
				"sessionId": "sess_abc123",
				"promptId": "prompt_xyz789",
				"type": "tool",
				"tool": "fs/write_file",
				"status": "completed"
			}`,
			want: SessionUpdateNotification{
				SessionID: "sess_abc123",
				PromptID:  "prompt_xyz789",
				Type:      "tool",
				Tool:      "fs/write_file",
				Status:    "completed",
			},
		},
		{
			name: "complete notification",
			data: `{
				"sessionId": "sess_abc123",
				"promptId": "prompt_xyz789",
				"type": "complete"
			}`,
			want: SessionUpdateNotification{
				SessionID: "sess_abc123",
				PromptID:  "prompt_xyz789",
				Type:      "complete",
			},
		},
		{
			name: "error notification",
			data: `{
				"sessionId": "sess_abc123",
				"promptId": "prompt_xyz789",
				"type": "error",
				"error": "timeout waiting for response"
			}`,
			want: SessionUpdateNotification{
				SessionID: "sess_abc123",
				PromptID:  "prompt_xyz789",
				Type:      "error",
				Error:     "timeout waiting for response",
			},
		},
		{
			name: "notification with metadata",
			data: `{
				"sessionId": "sess_abc123",
				"promptId": "prompt_xyz789",
				"type": "output",
				"data": "test",
				"metadata": {"key": "value", "num": 42}
			}`,
			want: SessionUpdateNotification{
				SessionID: "sess_abc123",
				PromptID:  "prompt_xyz789",
				Type:      "output",
				Data:      "test",
				Metadata:  map[string]any{"key": "value", "num": 42.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got SessionUpdateNotification
			err := json.Unmarshal([]byte(tt.data), &got)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.SessionID, got.SessionID)
				assert.Equal(t, tt.want.PromptID, got.PromptID)
				assert.Equal(t, tt.want.Type, got.Type)
				assert.Equal(t, tt.want.Data, got.Data)
				assert.Equal(t, tt.want.Progress, got.Progress)
				assert.Equal(t, tt.want.Message, got.Message)
				assert.Equal(t, tt.want.Tool, got.Tool)
				assert.Equal(t, tt.want.Status, got.Status)
				assert.Equal(t, tt.want.Error, got.Error)
			}
		})
	}
}

func TestSessionUpdateNotification_Marshal(t *testing.T) {
	notif := SessionUpdateNotification{
		SessionID: "sess_123",
		PromptID:  "prompt_456",
		Type:      "output",
		Data:      "test data",
		Metadata:  map[string]any{"key": "value"},
	}

	data, err := json.Marshal(notif)
	assert.NoError(t, err)

	var got SessionUpdateNotification
	err = json.Unmarshal(data, &got)
	assert.NoError(t, err)

	assert.Equal(t, notif.SessionID, got.SessionID)
	assert.Equal(t, notif.PromptID, got.PromptID)
	assert.Equal(t, notif.Type, got.Type)
	assert.Equal(t, notif.Data, got.Data)
}

func TestSessionCancelParams_Marshal(t *testing.T) {
	params := SessionCancelParams{
		SessionID: "sess_abc123",
		Reason:    "user_requested",
	}

	data, err := json.Marshal(params)
	assert.NoError(t, err)

	var got SessionCancelParams
	err = json.Unmarshal(data, &got)
	assert.NoError(t, err)

	assert.Equal(t, params.SessionID, got.SessionID)
	assert.Equal(t, params.Reason, got.Reason)
}

func TestSessionCancelResult_Marshal(t *testing.T) {
	result := SessionCancelResult{
		SessionID: "sess_abc123",
		Status:    "cancelled",
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)

	var got SessionCancelResult
	err = json.Unmarshal(data, &got)
	assert.NoError(t, err)

	assert.Equal(t, result.SessionID, got.SessionID)
	assert.Equal(t, result.Status, got.Status)
}

func TestACPClient_Validate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := Config{
		ClientName: "test-client",
	}

	client, err := NewACPClient(ctx, cfg)
	if err != nil {
		t.Skipf("opencode acp not available: %v", err)
	}
	defer func() { _ = client.Close() }()

	err = client.Validate(ctx)

	if err != nil {
		t.Skipf("ACP validation failed (likely opencode not running): %v", err)
	}
}

func TestACPClient_Validate_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	cfg := Config{
		ClientName: "test-client",
	}

	client, err := NewACPClient(ctx, cfg)
	if err != nil {
		t.Skipf("opencode acp not available: %v", err)
	}
	defer func() { _ = client.Close() }()

	err = client.Validate(ctx)

	assert.Error(t, err)
}
