package tools

import (
	"context"
	"testing"
	"time"

	"github.com/victorzhuk/go-ent/internal/agent/background"
)

func TestAgentBgStatus(t *testing.T) {
	manager := background.NewManager(nil, background.DefaultConfig())
	handler := makeAgentBgStatusHandler(manager)

	tests := []struct {
		name        string
		setup       func() string
		input       AgentBgStatusInput
		wantError   bool
		errorPrefix string
	}{
		{
			name: "valid agent",
			setup: func() string {
				agent, _ := manager.Spawn(context.Background(), "test task", background.SpawnOpts{})
				return agent.ID
			},
			input:     AgentBgStatusInput{},
			wantError: false,
		},
		{
			name: "empty agent id",
			setup: func() string {
				return ""
			},
			input:       AgentBgStatusInput{},
			wantError:   true,
			errorPrefix: "agent_id is required",
		},
		{
			name: "agent not found",
			setup: func() string {
				return "nonexistent-id"
			},
			input: AgentBgStatusInput{
				AgentID: "nonexistent-id",
			},
			wantError:   true,
			errorPrefix: "agent not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentID := tt.setup()
			if agentID != "" {
				tt.input.AgentID = agentID
			}

			ctx := context.Background()
			result, _, err := handler(ctx, nil, tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if tt.errorPrefix != "" && err.Error()[:len(tt.errorPrefix)] != tt.errorPrefix {
					t.Errorf("error prefix mismatch: got %s, want %s", err.Error()[:len(tt.errorPrefix)], tt.errorPrefix)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result.Content) == 0 {
				t.Error("expected content in result")
			}
		})
	}
}

func TestAgentBgStatusWithCompletedAgent(t *testing.T) {
	manager := background.NewManager(nil, background.DefaultConfig())
	handler := makeAgentBgStatusHandler(manager)

	agent, err := manager.Spawn(context.Background(), "test task", background.SpawnOpts{})
	if err != nil {
		t.Fatalf("failed to spawn agent: %v", err)
	}

	time.Sleep(15 * time.Millisecond)

	input := AgentBgStatusInput{
		AgentID: agent.ID,
	}

	ctx := context.Background()
	result, responseAny, err := handler(ctx, nil, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Content) == 0 {
		t.Error("expected content in result")
	}

	response, ok := responseAny.(AgentBgStatusResponse)
	if !ok {
		t.Fatal("expected AgentBgStatusResponse")
	}

	snap := agent.GetSnapshot()

	if response.ID != snap.ID {
		t.Errorf("expected agent ID %s, got %s", snap.ID, response.ID)
	}

	if response.Role != snap.Role {
		t.Errorf("expected role %s, got %s", snap.Role, response.Role)
	}

	if response.Model != snap.Model {
		t.Errorf("expected model %s, got %s", snap.Model, response.Model)
	}

	if response.Task != snap.Task {
		t.Errorf("expected task %s, got %s", snap.Task, response.Task)
	}

	if response.Status != string(snap.Status) {
		t.Errorf("expected status %s, got %s", snap.Status, response.Status)
	}
}
