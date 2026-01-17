package tools

import (
	"context"
	"testing"
	"time"

	"github.com/victorzhuk/go-ent/internal/agent/background"
)

func TestAgentBgOutput(t *testing.T) {
	manager := background.NewManager(nil, background.DefaultConfig())
	handler := makeAgentBgOutputHandler(manager)

	tests := []struct {
		name        string
		setup       func() string
		input       AgentBgOutputInput
		wantError   bool
		errorPrefix string
	}{
		{
			name: "valid completed agent with output",
			setup: func() string {
				agent, _ := manager.Spawn(context.Background(), "test task", background.SpawnOpts{})
				agent.Complete("test output")
				return agent.ID
			},
			input:     AgentBgOutputInput{},
			wantError: false,
		},
		{
			name: "empty agent id",
			setup: func() string {
				return ""
			},
			input:       AgentBgOutputInput{},
			wantError:   true,
			errorPrefix: "agent_id is required",
		},
		{
			name: "agent not found",
			setup: func() string {
				return "nonexistent-id"
			},
			input: AgentBgOutputInput{
				AgentID: "nonexistent-id",
			},
			wantError:   true,
			errorPrefix: "agent not found",
		},
		{
			name: "running agent returns empty output",
			setup: func() string {
				agent, _ := manager.Spawn(context.Background(), "test task", background.SpawnOpts{})
				agent.Start()
				return agent.ID
			},
			input: AgentBgOutputInput{
				FilterPattern: ".*",
			},
			wantError: false,
		},
		{
			name: "failed agent with output",
			setup: func() string {
				agent, _ := manager.Spawn(context.Background(), "test task", background.SpawnOpts{})
				agent.Complete("error occurred during execution")
				agent.Status = background.StatusFailed
				return agent.ID
			},
			input:     AgentBgOutputInput{},
			wantError: false,
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
				if tt.errorPrefix != "" && len(err.Error()) >= len(tt.errorPrefix) && err.Error()[:len(tt.errorPrefix)] != tt.errorPrefix {
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

func TestAgentBgOutputWithFilter(t *testing.T) {
	manager := background.NewManager(nil, background.DefaultConfig())
	handler := makeAgentBgOutputHandler(manager)

	tests := []struct {
		name       string
		output     string
		pattern    string
		wantOutput string
	}{
		{
			name:       "filters by error pattern",
			output:     "INFO: Starting\nERROR: Something went wrong\nDEBUG: Done",
			pattern:    "ERROR:.*",
			wantOutput: "ERROR: Something went wrong",
		},
		{
			name:       "filters multiple matches",
			output:     "line 1\nline 2\nline 3\nline 4",
			pattern:    "line [13]",
			wantOutput: "line 1line 3",
		},
		{
			name:       "empty pattern returns all output",
			output:     "test output",
			pattern:    "",
			wantOutput: "test output",
		},
		{
			name:       "no matches returns empty",
			output:     "INFO: Starting\nDEBUG: Done",
			pattern:    "ERROR:.*",
			wantOutput: "",
		},
		{
			name:       "case-insensitive filter",
			output:     "Error: failed\nERROR: retry\nerror: success",
			pattern:    "(?i)error:.*",
			wantOutput: "Error: failedERROR: retryerror: success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, _ := manager.Spawn(context.Background(), "test task", background.SpawnOpts{})
			agent.Complete(tt.output)

			input := AgentBgOutputInput{
				AgentID:       agent.ID,
				FilterPattern: tt.pattern,
			}

			ctx := context.Background()
			result, responseAny, err := handler(ctx, nil, input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Content) == 0 {
				t.Error("expected content in result")
			}

			response, ok := responseAny.(AgentBgOutputResponse)
			if !ok {
				t.Fatal("expected AgentBgOutputResponse")
			}

			if response.Output != tt.wantOutput {
				t.Errorf("expected output %q, got %q", tt.wantOutput, response.Output)
			}

			if tt.pattern != "" {
				if response.Filter != tt.pattern {
					t.Errorf("expected filter %q, got %q", tt.pattern, response.Filter)
				}
			}
		})
	}
}

func TestAgentBgOutputInvalidRegex(t *testing.T) {
	manager := background.NewManager(nil, background.DefaultConfig())
	handler := makeAgentBgOutputHandler(manager)

	agent, _ := manager.Spawn(context.Background(), "test task", background.SpawnOpts{})
	agent.Complete("some output")

	input := AgentBgOutputInput{
		AgentID:       agent.ID,
		FilterPattern: "[invalid(",
	}

	ctx := context.Background()
	result, _, err := handler(ctx, nil, input)

	if err == nil {
		t.Error("expected error for invalid regex but got none")
	}

	if len(result.Content) == 0 {
		t.Error("expected content in result even with error")
	}
}

func TestGetAgentOutput(t *testing.T) {
	tests := []struct {
		name       string
		snap       background.Snapshot
		pattern    string
		wantOutput string
		wantError  bool
	}{
		{
			name: "completed agent with no filter",
			snap: background.Snapshot{
				Status: background.StatusCompleted,
				Output: "test output",
			},
			pattern:    "",
			wantOutput: "test output",
			wantError:  false,
		},
		{
			name: "completed agent with filter",
			snap: background.Snapshot{
				Status: background.StatusCompleted,
				Output: "error: failed\ninfo: ok\nerror: timeout",
			},
			pattern:    "error:.*",
			wantOutput: "error: failederror: timeout",
			wantError:  false,
		},
		{
			name: "failed agent with output",
			snap: background.Snapshot{
				Status: background.StatusFailed,
				Output: "execution failed",
			},
			pattern:    "",
			wantOutput: "execution failed",
			wantError:  false,
		},
		{
			name: "running agent",
			snap: background.Snapshot{
				Status: background.StatusRunning,
				Output: "",
			},
			pattern:    "",
			wantOutput: "",
			wantError:  false,
		},
		{
			name: "completed agent with empty output",
			snap: background.Snapshot{
				Status: background.StatusCompleted,
				Output: "",
			},
			pattern:    "",
			wantOutput: "",
			wantError:  false,
		},
		{
			name: "completed agent with no matches",
			snap: background.Snapshot{
				Status: background.StatusCompleted,
				Output: "no errors here",
			},
			pattern:    "ERROR:.*",
			wantOutput: "",
			wantError:  false,
		},
		{
			name: "invalid regex pattern",
			snap: background.Snapshot{
				Status: background.StatusCompleted,
				Output: "test",
			},
			pattern:   "[invalid(",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := getAgentOutput(tt.snap, tt.pattern)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if output != tt.wantOutput {
				t.Errorf("expected output %q, got %q", tt.wantOutput, output)
			}
		})
	}
}

func TestAgentBgOutputResponseStructure(t *testing.T) {
	manager := background.NewManager(nil, background.DefaultConfig())
	handler := makeAgentBgOutputHandler(manager)

	agent, _ := manager.Spawn(context.Background(), "test task", background.SpawnOpts{})

	time.Sleep(15 * time.Millisecond)

	agent.Complete("test output line 1\ntest output line 2")

	input := AgentBgOutputInput{
		AgentID:       agent.ID,
		FilterPattern: "line 1",
	}

	ctx := context.Background()
	result, responseAny, err := handler(ctx, nil, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	response, ok := responseAny.(AgentBgOutputResponse)
	if !ok {
		t.Fatal("expected AgentBgOutputResponse")
	}

	snap := agent.GetSnapshot()

	if response.AgentID != snap.ID {
		t.Errorf("expected agent ID %s, got %s", snap.ID, response.AgentID)
	}

	if snap.Status != background.StatusCompleted {
		t.Errorf("expected agent to be completed, got %s", snap.Status)
	}

	if response.Filter != input.FilterPattern {
		t.Errorf("expected filter %s, got %s", input.FilterPattern, response.Filter)
	}

	if response.Output != "line 1" {
		t.Errorf("expected filtered output 'line 1', got %q", response.Output)
	}
}
