package tools

import (
	"context"
	"testing"

	"github.com/victorzhuk/go-ent/internal/agent/background"
)

func TestAgentSpawn(t *testing.T) {
	manager := background.NewManager(nil, background.DefaultConfig())
	handler := makeAgentSpawnHandler(manager)

	tests := []struct {
		name        string
		input       AgentSpawnInput
		wantError   bool
		errorPrefix string
	}{
		{
			name: "valid task",
			input: AgentSpawnInput{
				Task: "test task",
			},
			wantError: false,
		},
		{
			name: "empty task",
			input: AgentSpawnInput{
				Task: "",
			},
			wantError:   true,
			errorPrefix: "task is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			if manager.Count() == 0 {
				t.Error("expected agent to be spawned")
			}
		})
	}
}

func TestAgentSpawnWithOptionalFields(t *testing.T) {
	manager := background.NewManager(nil, background.DefaultConfig())
	handler := makeAgentSpawnHandler(manager)

	input := AgentSpawnInput{
		Task:    "test task",
		Role:    "architect",
		Model:   "opus",
		Timeout: 600,
	}

	ctx := context.Background()
	result, _, err := handler(ctx, nil, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Content) == 0 {
		t.Error("expected content in result")
	}

	if manager.Count() == 0 {
		t.Error("expected agent to be spawned")
	}
}
