package execution

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/domain"
)

// Engine orchestrates task execution across runners and strategies.
type Engine struct {
	runners    map[domain.Runtime]Runner
	strategies map[domain.ExecutionStrategy]Strategy
	selector   *agent.Selector
	budget     *BudgetTracker
	fallback   *FallbackResolver
	preferred  domain.Runtime
	logger     *slog.Logger
}

// Config holds engine configuration.
type Config struct {
	// PreferredRuntime is the default runtime to use.
	PreferredRuntime domain.Runtime

	// IsMCPMode determines budget behavior.
	IsMCPMode bool

	// Logger for execution logging.
	Logger *slog.Logger
}

// New creates an Engine with the given configuration.
func New(cfg Config, selector *agent.Selector) *Engine {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	engine := &Engine{
		runners:    make(map[domain.Runtime]Runner),
		strategies: make(map[domain.ExecutionStrategy]Strategy),
		selector:   selector,
		budget:     NewBudgetTracker(cfg.IsMCPMode, cfg.Logger),
		fallback:   NewFallbackResolver(),
		preferred:  cfg.PreferredRuntime,
		logger:     cfg.Logger,
	}

	// Register default runners
	engine.RegisterRunner(NewCLIRunner(cfg.Logger))
	engine.RegisterRunner(NewClaudeCodeRunner(cfg.Logger))
	engine.RegisterRunner(NewOpenCodeRunner(cfg.Logger))

	// Register strategies
	engine.RegisterStrategy(NewSingleStrategy())
	engine.RegisterStrategy(NewMultiStrategy())

	return engine
}

// RegisterRunner adds a runner implementation.
func (e *Engine) RegisterRunner(r Runner) {
	e.runners[r.Runtime()] = r
	e.logger.Debug("registered runner", "runtime", r.Runtime())
}

// RegisterStrategy adds a strategy implementation.
func (e *Engine) RegisterStrategy(s Strategy) {
	e.strategies[s.Name()] = s
	e.logger.Debug("registered strategy", "name", s.Name())
}

// Execute runs a task with automatic runner and strategy selection.
func (e *Engine) Execute(ctx context.Context, task *Task) (*Result, error) {
	e.logger.Info("executing task", "description", truncate(task.Description, 100))

	// Select strategy
	strategy := e.selectStrategy(task)

	// Execute with strategy
	result, err := strategy.Execute(ctx, e, task)
	if err != nil {
		e.logger.Error("execution failed", "error", err)
		return result, err
	}

	e.logger.Info("execution completed",
		"success", result.Success,
		"duration", result.Duration,
		"cost", fmt.Sprintf("$%.4f", result.Cost),
	)

	return result, nil
}

// ExecuteWithRunner runs using a specific runner.
func (e *Engine) ExecuteWithRunner(ctx context.Context, runtime domain.Runtime, task *Task) (*Result, error) {
	// Validate runner exists
	_, err := e.getRunner(runtime)
	if err != nil {
		return nil, err
	}

	// Force the runtime
	task.ForceRuntime = runtime

	// Use single strategy for direct runner execution
	strategy := NewSingleStrategy()
	return strategy.Execute(ctx, e, task)
}

// GetBudgetTracker returns the budget tracker.
func (e *Engine) GetBudgetTracker() *BudgetTracker {
	return e.budget
}

// selectStrategy selects the appropriate execution strategy.
func (e *Engine) selectStrategy(task *Task) Strategy {
	// If strategy is forced, use it
	if task.ForceStrategy != "" {
		if s, exists := e.strategies[task.ForceStrategy]; exists {
			return s
		}
	}

	// Try each strategy's CanHandle method
	for _, strategy := range e.strategies {
		if strategy.CanHandle(task) {
			return strategy
		}
	}

	// Default to single strategy
	return e.strategies[domain.ExecutionStrategySingle]
}

// selectRuntime selects the runtime to use.
func (e *Engine) selectRuntime(ctx context.Context) domain.Runtime {
	// If preferred runtime is set and available, use it
	if e.preferred != "" {
		if runner, exists := e.runners[e.preferred]; exists && runner.Available(ctx) {
			return e.preferred
		}
	}

	// Try runtimes in order: claude-code, open-code, cli
	for _, rt := range []domain.Runtime{
		domain.RuntimeClaudeCode,
		domain.RuntimeOpenCode,
		domain.RuntimeCLI,
	} {
		if runner, exists := e.runners[rt]; exists && runner.Available(ctx) {
			return rt
		}
	}

	// Fallback to CLI
	return domain.RuntimeCLI
}

// getRunner returns the runner for the given runtime with fallback support.
func (e *Engine) getRunner(runtime domain.Runtime) (Runner, error) {
	// Try primary runtime
	runner, exists := e.runners[runtime]
	if exists && runner.Available(context.Background()) {
		return runner, nil
	}

	e.logger.Warn("primary runtime unavailable, trying fallback",
		"runtime", runtime,
	)

	// Try fallbacks (same-family only)
	fallbacks := e.fallback.GetFallbacks(runtime)
	for _, fbRuntime := range fallbacks {
		runner, exists = e.runners[fbRuntime]
		if exists && runner.Available(context.Background()) {
			e.logger.Info("using fallback runtime",
				"original", runtime,
				"fallback", fbRuntime,
			)
			return runner, nil
		}
	}

	return nil, fmt.Errorf("no available runner for runtime %s", runtime)
}

// selectAgent uses the agent selector to choose agent/model/skills.
func (e *Engine) selectAgent(ctx context.Context, task *Task) (*SelectionResult, error) {
	// Convert execution task to agent task
	agentTask := agent.Task{
		Description: task.Description,
		Type:        agent.TaskType(task.Type),
	}

	// Use selector
	selection, err := e.selector.Select(ctx, agentTask)
	if err != nil {
		return nil, err
	}

	return &SelectionResult{
		Agent:  selection.Role,
		Model:  selection.Model,
		Skills: selection.Skills,
	}, nil
}

// SelectionResult holds the selected agent configuration.
type SelectionResult struct {
	Agent  domain.AgentRole
	Model  string
	Skills []string
}
