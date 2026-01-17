package tools

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent/background"
)

func TestAgentBgList(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*background.Manager)
		input       AgentBgListInput
		wantError   bool
		errorPrefix string
	}{
		{
			name:  "empty list",
			setup: func(m *background.Manager) {},
			input: AgentBgListInput{
				Status: "",
			},
			wantError: false,
		},
		{
			name: "list with agents",
			setup: func(m *background.Manager) {
				m.Spawn(context.Background(), "task 1", background.SpawnOpts{})
				m.Spawn(context.Background(), "task 2", background.SpawnOpts{})
			},
			input:     AgentBgListInput{},
			wantError: false,
		},
		{
			name: "filter by status pending",
			setup: func(m *background.Manager) {
				m.Spawn(context.Background(), "task pending", background.SpawnOpts{})
			},
			input: AgentBgListInput{
				Status: string(background.StatusPending),
			},
			wantError: false,
		},
		{
			name: "filter by status running",
			setup: func(m *background.Manager) {
				m.Spawn(context.Background(), "task running", background.SpawnOpts{})
			},
			input: AgentBgListInput{
				Status: string(background.StatusRunning),
			},
			wantError: false,
		},
		{
			name: "filter by status completed",
			setup: func(m *background.Manager) {
				agent, err := m.Spawn(context.Background(), "task completed", background.SpawnOpts{})
				if err == nil && agent != nil {
					agent.Complete("done")
				}
			},
			input: AgentBgListInput{
				Status: string(background.StatusCompleted),
			},
			wantError: false,
		},
		{
			name: "filter by status failed",
			setup: func(m *background.Manager) {
				agent, err := m.Spawn(context.Background(), "task failed", background.SpawnOpts{})
				if err == nil && agent != nil {
					agent.Fail(errors.New("test error"))
				}
			},
			input: AgentBgListInput{
				Status: string(background.StatusFailed),
			},
			wantError: false,
		},
		{
			name: "filter by status killed",
			setup: func(m *background.Manager) {
				agent, err := m.Spawn(context.Background(), "task killed", background.SpawnOpts{})
				if err == nil && agent != nil {
					agent.Kill()
				}
			},
			input: AgentBgListInput{
				Status: string(background.StatusKilled),
			},
			wantError: false,
		},
		{
			name:  "invalid status",
			setup: func(m *background.Manager) {},
			input: AgentBgListInput{
				Status: "invalid_status",
			},
			wantError:   true,
			errorPrefix: "invalid status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := background.NewManager(nil, background.DefaultConfig())
			handler := makeAgentBgListHandler(manager)
			tt.setup(manager)

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

func TestAgentBgListWithMultipleAgents(t *testing.T) {
	manager := background.NewManager(nil, background.DefaultConfig())
	handler := makeAgentBgListHandler(manager)

	_, _ = manager.Spawn(context.Background(), "task 1", background.SpawnOpts{})
	_, _ = manager.Spawn(context.Background(), "task 2", background.SpawnOpts{})
	_, _ = manager.Spawn(context.Background(), "task 3", background.SpawnOpts{})

	// Wait for agents to complete
	time.Sleep(50 * time.Millisecond)

	tests := []struct {
		name          string
		input         AgentBgListInput
		expectedCount int
	}{
		{
			name: "list all agents",
			input: AgentBgListInput{
				Status: "",
			},
			expectedCount: 3,
		},
		{
			name: "filter completed",
			input: AgentBgListInput{
				Status: string(background.StatusCompleted),
			},
			expectedCount: 3,
		},
		{
			name: "filter pending",
			input: AgentBgListInput{
				Status: string(background.StatusPending),
			},
			expectedCount: 0,
		},
		{
			name: "filter running",
			input: AgentBgListInput{
				Status: string(background.StatusRunning),
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, responseAny, err := handler(ctx, nil, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Content) == 0 {
				t.Error("expected content in result")
			}

			response, ok := responseAny.(AgentBgListResponse)
			if !ok {
				t.Fatal("expected AgentBgListResponse")
			}

			if response.TotalCount != tt.expectedCount {
				t.Errorf("expected count %d, got %d", tt.expectedCount, response.TotalCount)
			}

			if len(response.Agents) != tt.expectedCount {
				t.Errorf("expected %d agents, got %d", tt.expectedCount, len(response.Agents))
			}

			if tt.input.Status != "" {
				if response.Counts[tt.input.Status] != tt.expectedCount {
					t.Errorf("expected %d agents with status %s, got %d", tt.expectedCount, tt.input.Status, response.Counts[tt.input.Status])
				}
			}
		})
	}
}

func TestAgentBgListResponseFields(t *testing.T) {
	manager := background.NewManager(nil, background.DefaultConfig())
	handler := makeAgentBgListHandler(manager)

	agent, err := manager.Spawn(context.Background(), "test task", background.SpawnOpts{})
	if err != nil {
		t.Fatalf("failed to spawn agent: %v", err)
	}

	input := AgentBgListInput{}

	ctx := context.Background()

	var result *mcp.CallToolResult
	var responseAny any

	for i := 0; i < 10; i++ {
		// Give agent time to complete
		time.Sleep(20 * time.Millisecond)

		result, responseAny, err = handler(ctx, nil, input)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Content) == 0 {
			t.Error("expected content in result")
		}

		response, ok := responseAny.(AgentBgListResponse)
		if !ok {
			t.Fatal("expected AgentBgListResponse")
		}

		if len(response.Agents) > 0 && response.Agents[0].Status == string(background.StatusCompleted) {
			firstAgent := response.Agents[0]
			if firstAgent.ID != agent.ID {
				t.Errorf("expected agent ID %s, got %s", agent.ID, firstAgent.ID)
			}

			if firstAgent.Role != agent.Role {
				t.Errorf("expected role %s, got %s", agent.Role, firstAgent.Role)
			}

			if firstAgent.Model != agent.Model {
				t.Errorf("expected model %s, got %s", agent.Model, firstAgent.Model)
			}

			if firstAgent.Task != agent.Task {
				t.Errorf("expected task %s, got %s", agent.Task, firstAgent.Task)
			}

			if firstAgent.Status != string(agent.Status) {
				t.Errorf("expected status %s, got %s", agent.Status, firstAgent.Status)
			}

			if firstAgent.Status != string(background.StatusCompleted) {
				t.Errorf("expected status %s, got %s", background.StatusCompleted, firstAgent.Status)
			}

			return
		}
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Content) == 0 {
		t.Error("expected content in result")
	}

	response, ok := responseAny.(AgentBgListResponse)
	if !ok {
		t.Fatal("expected AgentBgListResponse")
	}

	if response.TotalCount == 0 {
		t.Error("expected non-zero total count")
	}

	if len(response.Agents) == 0 {
		t.Error("expected at least one agent")
	}

	firstAgent := response.Agents[0]
	if firstAgent.ID != agent.ID {
		t.Errorf("expected agent ID %s, got %s", agent.ID, firstAgent.ID)
	}

	if firstAgent.Role != agent.Role {
		t.Errorf("expected role %s, got %s", agent.Role, firstAgent.Role)
	}

	if firstAgent.Model != agent.Model {
		t.Errorf("expected model %s, got %s", agent.Model, firstAgent.Model)
	}

	if firstAgent.Task != agent.Task {
		t.Errorf("expected task %s, got %s", agent.Task, firstAgent.Task)
	}

	if firstAgent.Status != string(agent.Status) {
		t.Errorf("expected status %s, got %s", agent.Status, firstAgent.Status)
	}

	if firstAgent.Status != string(background.StatusCompleted) {
		t.Errorf("expected status %s, got %s", background.StatusCompleted, firstAgent.Status)
	}
}
