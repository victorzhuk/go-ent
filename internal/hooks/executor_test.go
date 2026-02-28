package hooks

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestExecutor_MatchTool(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	tests := []struct {
		pattern  string
		toolName string
		want     bool
	}{
		{"", "any_tool", true}, // Empty pattern matches all
		{"Bash", "Bash", true},
		{"Bash", "Edit", false},
		{"Edit|Write", "Edit", true},
		{"Edit|Write", "Write", true},
		{"Edit|Write", "Read", false},
		{".*Edit.*", "MultiEdit", true},
		{"^Bash$", "Bash", true},
		{"^Bash$", "BashOutput", false},
	}

	for _, tt := range tests {
		got := exec.MatchTool(tt.pattern, tt.toolName)
		if got != tt.want {
			t.Errorf("MatchTool(%q, %q) = %v, want %v", tt.pattern, tt.toolName, got, tt.want)
		}
	}
}

func TestExecutor_RunOpenSpecHook_Command(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	ctx := context.Background()

	hook := Hook{
		Type:    HookTypeCommand,
		Command: "echo 'test'",
	}

	err := exec.RunOpenSpecHook(ctx, hook, "test_event", map[string]string{
		"TEST_VAR": "test_value",
	})
	if err != nil {
		t.Errorf("RunOpenSpecHook() error = %v, want nil", err)
	}
}

func TestExecutor_RunOpenSpecHook_Agent(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	ctx := context.Background()

	hook := Hook{
		Type:   HookTypeAgent,
		Agent:  "reviewer",
		Prompt: "Review the code",
	}

	// Agent hooks should not error, just log suggestion
	err := exec.RunOpenSpecHook(ctx, hook, "test_event", map[string]string{
		"CHANGE_ID": "test-change",
	})
	if err != nil {
		t.Errorf("RunOpenSpecHook() error = %v, want nil", err)
	}
}

func TestExecutor_RunOpenSpecHook_EmptyHook(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	ctx := context.Background()

	hook := Hook{} // Empty hook

	// Empty hook should be no-op
	err := exec.RunOpenSpecHook(ctx, hook, "test_event", nil)
	if err != nil {
		t.Errorf("RunOpenSpecHook() error = %v, want nil", err)
	}
}

func TestExecutor_ExecuteCommand_Timeout(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	ctx := context.Background()

	// Command that would hang indefinitely
	err := exec.ExecuteCommand(ctx, "sleep 20", nil)

	// Should timeout within 10 seconds
	if err == nil {
		t.Error("ExecuteCommand() expected timeout error, got nil")
	}
}

func TestExecutor_ExecuteCommand_WithEnv(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	ctx := context.Background()

	// Command that checks environment variable
	err := exec.ExecuteCommand(ctx, "test \"$HOOK_VAR\" = \"hook_value\"", map[string]string{
		"HOOK_VAR": "hook_value",
	})
	if err != nil {
		t.Errorf("ExecuteCommand() error = %v, want nil", err)
	}
}
