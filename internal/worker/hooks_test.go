package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/execution"
)

func TestHookType_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		hook HookType
		want bool
	}{
		{"valid pre_spawn", HookPreSpawn, true},
		{"valid pre_prompt", HookPrePrompt, true},
		{"valid post_complete", HookPostComplete, true},
		{"valid post_fail", HookPostFail, true},
		{"invalid", HookType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.hook.Valid(); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHookChain_Add(t *testing.T) {
	t.Parallel()

	hc := NewHookChain()

	hook := func(ctx context.Context, hookCtx *HookContext) error {
		return nil
	}

	hc.Add(HookPreSpawn, hook, nil)
	hc.Add(HookPrePrompt, hook, nil)

	if len(hc.PreSpawn) != 1 {
		t.Errorf("PreSpawn count = %d, want 1", len(hc.PreSpawn))
	}

	if len(hc.PrePrompt) != 1 {
		t.Errorf("PrePrompt count = %d, want 1", len(hc.PrePrompt))
	}
}

func TestHookChain_Execute(t *testing.T) {
	t.Parallel()

	t.Run("execute hooks", func(t *testing.T) {
		t.Parallel()

		hc := NewHookChain()

		called := []string{}
		hc.Add(HookPreSpawn, func(ctx context.Context, hookCtx *HookContext) error {
			called = append(called, "hook1")
			return nil
		}, nil)

		hc.Add(HookPreSpawn, func(ctx context.Context, hookCtx *HookContext) error {
			called = append(called, "hook2")
			return nil
		}, nil)

		ctx := context.Background()
		hookCtx := &HookContext{
			HookType: HookPreSpawn,
		}

		err := hc.Execute(ctx, HookPreSpawn, hookCtx)
		if err != nil {
			t.Errorf("Execute() error = %v", err)
		}

		if len(called) != 2 {
			t.Errorf("called hooks = %v, want 2", called)
		}
	})

	t.Run("invalid hook type", func(t *testing.T) {
		t.Parallel()

		hc := NewHookChain()
		ctx := context.Background()
		hookCtx := &HookContext{
			HookType: HookType("invalid"),
		}

		err := hc.Execute(ctx, HookType("invalid"), hookCtx)
		if err == nil {
			t.Error("Execute() expected error for invalid hook type")
		}
	})

	t.Run("hook returns error", func(t *testing.T) {
		t.Parallel()

		hc := NewHookChain()

		hc.Add(HookPreSpawn, func(ctx context.Context, hookCtx *HookContext) error {
			return errors.New("hook error")
		}, nil)

		ctx := context.Background()
		hookCtx := &HookContext{
			HookType: HookPreSpawn,
		}

		err := hc.Execute(ctx, HookPreSpawn, hookCtx)
		if err == nil {
			t.Error("Execute() expected error from hook")
		}
	})
}

func TestHookFilter_Matches(t *testing.T) {
	t.Parallel()

	t.Run("no filter", func(t *testing.T) {
		t.Parallel()

		task := execution.NewTask("test task")
		worker := &Worker{
			Provider: "glm",
			Model:    "glm-4",
		}

		hook := Hook{
			Func: func(ctx context.Context, hookCtx *HookContext) error {
				return nil
			},
			Filter: nil,
		}

		hookCtx := &HookContext{
			Task:   task,
			Worker: worker,
		}

		if !hook.matchesFilter(hookCtx) {
			t.Error("expected hook to match when no filter")
		}
	})

	t.Run("filter by provider", func(t *testing.T) {
		t.Parallel()

		task := execution.NewTask("test task")
		worker := &Worker{
			Provider: "glm",
		}

		filter := &HookFilter{
			Provider: "glm",
		}

		hook := Hook{
			Func:   func(ctx context.Context, hookCtx *HookContext) error { return nil },
			Filter: filter,
		}

		hookCtx := &HookContext{
			Task:   task,
			Worker: worker,
		}

		if !hook.matchesFilter(hookCtx) {
			t.Error("expected hook to match provider filter")
		}

		worker.Provider = "kimi"
		if hook.matchesFilter(hookCtx) {
			t.Error("expected hook not to mismatch provider filter")
		}
	})

	t.Run("filter by task type", func(t *testing.T) {
		t.Parallel()

		task := execution.NewTask("test task").WithType("feature")

		filter := &HookFilter{
			TaskType: "feature",
		}

		hook := Hook{
			Func:   func(ctx context.Context, hookCtx *HookContext) error { return nil },
			Filter: filter,
		}

		hookCtx := &HookContext{
			Task: task,
		}

		if !hook.matchesFilter(hookCtx) {
			t.Error("expected hook to match task type filter")
		}

		task.Type = "bugfix"
		if hook.matchesFilter(hookCtx) {
			t.Error("expected hook not to mismatch task type filter")
		}
	})
}

func TestHookChain_FilteredExecution(t *testing.T) {
	t.Parallel()

	hc := NewHookChain()

	called := []string{}

	hc.Add(HookPreSpawn, func(ctx context.Context, hookCtx *HookContext) error {
		called = append(called, "glm_hook")
		return nil
	}, &HookFilter{Provider: "glm"})

	hc.Add(HookPreSpawn, func(ctx context.Context, hookCtx *HookContext) error {
		called = append(called, "kimi_hook")
		return nil
	}, &HookFilter{Provider: "kimi"})

	task := execution.NewTask("test task")
	worker := &Worker{
		Provider: "glm",
	}

	hookCtx := &HookContext{
		HookType: HookPreSpawn,
		Task:     task,
		Worker:   worker,
	}

	ctx := context.Background()
	err := hc.Execute(ctx, HookPreSpawn, hookCtx)

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if len(called) != 1 || called[0] != "glm_hook" {
		t.Errorf("called hooks = %v, want [glm_hook]", called)
	}
}

func TestWorkerManager_RegisterHook(t *testing.T) {
	t.Parallel()

	m := NewWorkerManagerWithoutTracking()

	hookCalled := false
	hook := func(ctx context.Context, hookCtx *HookContext) error {
		hookCalled = true
		return nil
	}

	m.RegisterHook(HookPreSpawn, hook, nil)

	hc := m.hooks
	if len(hc.PreSpawn) != 1 {
		t.Errorf("PreSpawn count = %d, want 1", len(hc.PreSpawn))
	}

	ctx := context.Background()
	hookCtx := &HookContext{
		HookType: HookPreSpawn,
	}

	err := m.ExecuteHooks(ctx, HookPreSpawn, hookCtx)
	if err != nil {
		t.Errorf("ExecuteHooks() error = %v", err)
	}

	if !hookCalled {
		t.Error("hook was not called")
	}
}

func TestWorkerManager_HookIntegration(t *testing.T) {
	t.Parallel()

	t.Run("pre_spawn hook", func(t *testing.T) {
		t.Parallel()

		m := NewWorkerManagerWithoutTracking()

		hookCalled := false
		m.RegisterHook(HookPreSpawn, func(ctx context.Context, hookCtx *HookContext) error {
			hookCalled = true
			if hookCtx.Task == nil {
				t.Error("expected Task in hook context")
			}
			if hookCtx.Worker == nil {
				t.Error("expected Worker in hook context")
			}
			return nil
		}, nil)

		task := execution.NewTask("test task")
		req := SpawnRequest{
			Task:     task,
			Method:   "acp",
			Provider: "glm",
			Model:    "glm-4",
		}

		ctx := context.Background()
		_, err := m.Spawn(ctx, req)

		if err != nil {
			t.Errorf("Spawn() error = %v", err)
		}

		if !hookCalled {
			t.Error("pre_spawn hook was not called")
		}
	})

	t.Run("pre_prompt hook", func(t *testing.T) {
		t.Parallel()

		m := NewWorkerManagerWithoutTracking()

		task := execution.NewTask("test task")
		worker := &Worker{
			ID:       "test-worker",
			Provider: "glm",
			Model:    "glm-4",
			Method:   config.MethodACP,
			Status:   StatusIdle,
			Task:     task,
		}

		m.mu.Lock()
		m.workers["test-worker"] = worker
		m.mu.Unlock()

		hookCalled := false
		m.RegisterHook(HookPrePrompt, func(ctx context.Context, hookCtx *HookContext) error {
			hookCalled = true
			if hookCtx.Task == nil {
				t.Error("expected Task in hook context")
			}
			if hookCtx.Worker == nil {
				t.Error("expected Worker in hook context")
			}
			return nil
		}, nil)

		req := PromptRequest{
			WorkerID: "test-worker",
			Prompt:   "test prompt",
		}

		ctx := context.Background()
		_, err := m.SendPrompt(ctx, req)

		if err == nil {
			t.Error("expected error (no ACP client)")
		}

		if !hookCalled {
			t.Error("pre_prompt hook was not called")
		}
	})

	t.Run("post_fail hook not called for validation errors", func(t *testing.T) {
		t.Parallel()

		m := NewWorkerManagerWithoutTracking()

		task := execution.NewTask("test task")
		worker := &Worker{
			ID:       "test-worker",
			Provider: "glm",
			Model:    "glm-4",
			Method:   config.MethodACP,
			Status:   StatusIdle,
			Task:     task,
		}

		m.mu.Lock()
		m.workers["test-worker"] = worker
		m.mu.Unlock()

		postFailCalled := false
		m.RegisterHook(HookPostFail, func(ctx context.Context, hookCtx *HookContext) error {
			postFailCalled = true
			return nil
		}, nil)

		req := PromptRequest{
			WorkerID: "test-worker",
			Prompt:   "test prompt",
		}

		ctx := context.Background()
		_, err := m.SendPrompt(ctx, req)

		if err == nil {
			t.Error("expected error (no ACP client)")
		}

		if postFailCalled {
			t.Error("post_fail hook should not be called for validation errors")
		}
	})
}

func TestExternalCommandHook(t *testing.T) {
	t.Skip("requires shell, skipped in CI")

	t.Parallel()

	ctx := context.Background()
	hookCtx := &HookContext{
		HookType: HookPreSpawn,
	}

	hook := ExternalCommandHook("echo", "test")
	err := hook(ctx, hookCtx)

	if err != nil {
		t.Errorf("ExternalCommandHook() error = %v", err)
	}
}

func TestMCPToolHook(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	hookCtx := &HookContext{
		HookType: HookPreSpawn,
	}

	hook := MCPToolHook("test_tool", nil)
	err := hook(ctx, hookCtx)

	if err != nil {
		t.Errorf("MCPToolHook() error = %v", err)
	}
}

func TestHookContext(t *testing.T) {
	t.Parallel()

	task := execution.NewTask("test task")
	worker := &Worker{
		ID:       "test-worker",
		Provider: "glm",
		Model:    "glm-4",
		Status:   StatusRunning,
		Task:     task,
	}

	result := &WorkerResult{
		WorkerID: "test-worker",
		Success:  true,
		Output:   "test output",
		Cost:     0.05,
	}

	hookCtx := &HookContext{
		HookType: HookPostComplete,
		Task:     task,
		Worker:   worker,
		Result:   result,
		Error:    nil,
	}

	if hookCtx.HookType != HookPostComplete {
		t.Errorf("HookType = %v, want %v", hookCtx.HookType, HookPostComplete)
	}

	if hookCtx.Task != task {
		t.Error("Task mismatch")
	}

	if hookCtx.Worker != worker {
		t.Error("Worker mismatch")
	}

	if hookCtx.Result != result {
		t.Error("Result mismatch")
	}

	if hookCtx.Error != nil {
		t.Error("Expected nil Error")
	}
}
