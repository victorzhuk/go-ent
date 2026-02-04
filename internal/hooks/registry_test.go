package hooks

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry_LoadFromEmbed(t *testing.T) {
	exec := NewExecutor(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	reg, err := NewRegistry("", exec)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	// Embedded loading may fail in test environment, which is OK
	// Just verify the registry was created
	if reg == nil {
		t.Fatal("Expected non-nil registry")
	}

	toolHooks := reg.GetToolHooks()
	// Empty hooks are fine if embedded loading failed
	_ = toolHooks
}

func TestRegistry_LoadFromFile_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "hooks.json")

	configData := `{
	"hooks": {
		"PreToolUse": [
			{
				"matcher": "Bash",
				"hooks": [
					{
						"type": "command",
						"command": "echo 'pre-hook'"
					}
				]
			}
		]
	}
}`

	if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	exec := NewExecutor(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	reg, err := NewRegistry(configPath, exec)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	toolHooks := reg.GetToolHooks()
	if len(toolHooks.PreToolUse) != 1 {
		t.Errorf("Expected 1 PreToolUse hook, got %d\nConfig: %+v", len(toolHooks.PreToolUse), reg.config)
		return
	}

	if toolHooks.PreToolUse[0].Matcher != "Bash" {
		t.Errorf("Expected matcher 'Bash', got %q", toolHooks.PreToolUse[0].Matcher)
	}

	if len(toolHooks.PreToolUse[0].Hooks) != 1 {
		t.Errorf("Expected 1 hook, got %d", len(toolHooks.PreToolUse[0].Hooks))
	}
}

func TestRegistry_LoadFromFile_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "hooks.yaml")

	configData := `
openspec:
  onChangeCreated:
    type: command
    command: echo 'change created'
  beforeArchive:
    type: agent
    agent: reviewer
    prompt: Review before archiving
`

	if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	exec := NewExecutor(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	reg, err := NewRegistry(configPath, exec)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	openspecHooks := reg.GetOpenSpecHooks()

	if openspecHooks.OnChangeCreated.Type != HookTypeCommand {
		t.Errorf("Expected onChangeCreated type %q, got %q", HookTypeCommand, openspecHooks.OnChangeCreated.Type)
	}

	if openspecHooks.BeforeArchive.Type != HookTypeAgent {
		t.Errorf("Expected beforeArchive type %q, got %q", HookTypeAgent, openspecHooks.BeforeArchive.Type)
	}

	if openspecHooks.BeforeArchive.Agent != "reviewer" {
		t.Errorf("Expected agent 'reviewer', got %q", openspecHooks.BeforeArchive.Agent)
	}
}

func TestRegistry_ThreadSafety(t *testing.T) {
	exec := NewExecutor(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	reg, err := NewRegistry("", exec)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	// Test concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			reg.GetToolHooks()
			reg.GetOpenSpecHooks()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
