package worker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"log/slog"

	"github.com/victorzhuk/go-ent/internal/execution"
)

type HookType string

const (
	HookPreSpawn     HookType = "pre_spawn"
	HookPrePrompt    HookType = "pre_prompt"
	HookPostComplete HookType = "post_complete"
	HookPostFail     HookType = "post_fail"
)

func (h HookType) String() string {
	return string(h)
}

func (h HookType) Valid() bool {
	switch h {
	case HookPreSpawn, HookPrePrompt, HookPostComplete, HookPostFail:
		return true
	default:
		return false
	}
}

type HookContext struct {
	HookType HookType
	Task     *execution.Task
	Worker   *Worker
	Result   *WorkerResult
	Error    error
}

type HookFunc func(ctx context.Context, hookCtx *HookContext) error

type HookFilter struct {
	Provider string
	Model    string
	Method   string
	TaskType string
}

type Hook struct {
	Func   HookFunc
	Filter *HookFilter
}

type HookChain struct {
	PreSpawn     []Hook
	PrePrompt    []Hook
	PostComplete []Hook
	PostFail     []Hook
	mu           sync.RWMutex
}

func NewHookChain() *HookChain {
	return &HookChain{
		PreSpawn:     make([]Hook, 0),
		PrePrompt:    make([]Hook, 0),
		PostComplete: make([]Hook, 0),
		PostFail:     make([]Hook, 0),
	}
}

func (hc *HookChain) Add(hookType HookType, hook HookFunc, filter *HookFilter) {
	if !hookType.Valid() {
		return
	}

	hc.mu.Lock()
	defer hc.mu.Unlock()

	h := Hook{
		Func:   hook,
		Filter: filter,
	}

	switch hookType {
	case HookPreSpawn:
		hc.PreSpawn = append(hc.PreSpawn, h)
	case HookPrePrompt:
		hc.PrePrompt = append(hc.PrePrompt, h)
	case HookPostComplete:
		hc.PostComplete = append(hc.PostComplete, h)
	case HookPostFail:
		hc.PostFail = append(hc.PostFail, h)
	}
}

func (hc *HookChain) Execute(ctx context.Context, hookType HookType, hookCtx *HookContext) error {
	if !hookType.Valid() {
		return fmt.Errorf("invalid hook type: %s", hookType)
	}

	hc.mu.RLock()
	defer hc.mu.RUnlock()

	var hooks []Hook
	switch hookType {
	case HookPreSpawn:
		hooks = hc.PreSpawn
	case HookPrePrompt:
		hooks = hc.PrePrompt
	case HookPostComplete:
		hooks = hc.PostComplete
	case HookPostFail:
		hooks = hc.PostFail
	}

	for _, h := range hooks {
		if !h.matchesFilter(hookCtx) {
			continue
		}

		if err := h.Func(ctx, hookCtx); err != nil {
			return fmt.Errorf("hook failed: %w", err)
		}
	}

	return nil
}

func (h *Hook) matchesFilter(hookCtx *HookContext) bool {
	if h.Filter == nil {
		return true
	}

	f := h.Filter

	if f.Provider != "" && hookCtx.Worker != nil && hookCtx.Worker.Provider != f.Provider {
		return false
	}

	if f.Model != "" && hookCtx.Worker != nil && hookCtx.Worker.Model != f.Model {
		return false
	}

	if f.Method != "" && hookCtx.Worker != nil && hookCtx.Worker.Method.String() != f.Method {
		return false
	}

	if f.TaskType != "" && hookCtx.Task != nil && hookCtx.Task.Type != f.TaskType {
		return false
	}

	return true
}

type WorkerResult struct {
	WorkerID  string
	Success   bool
	Output    string
	Error     error
	TokensIn  int
	TokensOut int
	Cost      float64
	Duration  float64
}

func ExternalCommandHook(cmd string, args ...string) HookFunc {
	return func(ctx context.Context, hookCtx *HookContext) error {
		c := exec.CommandContext(ctx, cmd, args...)
		output, err := c.CombinedOutput()
		if err != nil {
			return fmt.Errorf("external hook command failed: %w: %s", err, string(output))
		}
		slog.Debug("external hook executed",
			"hook_type", hookCtx.HookType,
			"cmd", cmd,
			"output", strings.TrimSpace(string(output)),
		)
		return nil
	}
}

func MCPToolHook(toolName string, toolParams map[string]interface{}) HookFunc {
	return func(ctx context.Context, hookCtx *HookContext) error {
		slog.Debug("MCP tool hook called",
			"hook_type", hookCtx.HookType,
			"tool", toolName,
		)
		return nil
	}
}
