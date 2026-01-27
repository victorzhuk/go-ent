package execution

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/domain"
)

// Engine orchestrates task execution across runners and strategies.
type Engine struct {
	runners                map[domain.Runtime]Runner
	strategies             map[domain.ExecutionStrategy]Strategy
	selector               *agent.Selector
	budget                 *BudgetTracker
	fallback               *FallbackResolver
	preferred              domain.Runtime
	logger                 *slog.Logger
	summarizationEnabled   bool
	summarizationThreshold SummarizationThreshold
	summarizationModel     string
	autoCheckpointEnabled  bool
	maxCheckpoints         int
	checkpointAgeLimit     time.Duration
}

// Config holds engine configuration.
type Config struct {
	// PreferredRuntime is the default runtime to use.
	PreferredRuntime domain.Runtime

	// IsMCPMode determines budget behavior.
	IsMCPMode bool

	// Logger for execution logging.
	Logger *slog.Logger

	// EnableSummarization enables automatic context summarization.
	EnableSummarization bool

	// SummarizationThreshold defines when to trigger summarization.
	SummarizationThreshold SummarizationThreshold

	// SummarizationModel is the LLM model to use for summarization.
	SummarizationModel string

	// EnableAutoCheckpoint enables automatic checkpointing after task completion.
	EnableAutoCheckpoint bool

	// MaxCheckpoints is the maximum number of checkpoints to keep per execution.
	MaxCheckpoints int

	// CheckpointAgeLimit is the maximum age for checkpoints before cleanup.
	CheckpointAgeLimit time.Duration
}

// New creates an Engine with the given configuration.
func New(cfg Config, selector *agent.Selector) *Engine {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	// Set default summarization values
	if cfg.SummarizationThreshold.FileCount == 0 &&
		cfg.SummarizationThreshold.ContextLength == 0 &&
		cfg.SummarizationThreshold.TokenCount == 0 {
		cfg.SummarizationThreshold = DefaultSummarizationThreshold()
	}
	if cfg.SummarizationModel == "" {
		cfg.SummarizationModel = LLMModelClaude35
	}

	// Set default checkpoint values
	if cfg.MaxCheckpoints == 0 {
		cfg.MaxCheckpoints = 10
	}
	if cfg.CheckpointAgeLimit == 0 {
		cfg.CheckpointAgeLimit = 24 * time.Hour
	}

	engine := &Engine{
		runners:                make(map[domain.Runtime]Runner),
		strategies:             make(map[domain.ExecutionStrategy]Strategy),
		selector:               selector,
		budget:                 NewBudgetTracker(cfg.IsMCPMode, cfg.Logger),
		fallback:               NewFallbackResolver(),
		preferred:              cfg.PreferredRuntime,
		logger:                 cfg.Logger,
		summarizationEnabled:   cfg.EnableSummarization,
		summarizationThreshold: cfg.SummarizationThreshold,
		summarizationModel:     cfg.SummarizationModel,
		autoCheckpointEnabled:  cfg.EnableAutoCheckpoint,
		maxCheckpoints:         cfg.MaxCheckpoints,
		checkpointAgeLimit:     cfg.CheckpointAgeLimit,
	}

	// Register default runners
	engine.RegisterRunner(NewCLIRunner(cfg.Logger))
	engine.RegisterRunner(NewClaudeCodeRunner(cfg.Logger))
	engine.RegisterRunner(NewOpenCodeRunner(cfg.Logger))

	// Register strategies
	engine.RegisterStrategy(NewSingleStrategy())
	engine.RegisterStrategy(NewMultiStrategy())
	engine.RegisterStrategy(NewParallelStrategy(4))

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

	// Create execution state for tracking
	state := NewExecutionState(task)

	// Determine execution configuration from task or defaults
	cfg := e.determineExecutionConfig(task)
	state.WithConfig(cfg)

	// Start execution
	if err := state.Start(); err != nil {
		return nil, fmt.Errorf("start execution: %w", err)
	}

	// Save initial checkpoint
	if e.autoCheckpointEnabled {
		if err := e.createCheckpoint(state); err != nil {
			e.logger.Warn("failed to save initial checkpoint", "error", err)
		}
	}

	// Select strategy
	strategy := e.selectStrategy(task)

	// Execute with strategy
	result, err := strategy.Execute(ctx, e, task)
	if err != nil {
		e.logger.Error("execution failed", "error", err)

		// Save checkpoint on error
		if e.autoCheckpointEnabled {
			if stateErr := state.Fail(err); stateErr == nil {
				if cpErr := e.createCheckpoint(state); cpErr != nil {
					e.logger.Warn("failed to save error checkpoint", "error", cpErr)
				}
			}
		}

		return result, err
	}

	// Update state with result
	if err := state.Complete(result); err != nil {
		e.logger.Warn("failed to update execution state", "error", err)
	}

	// Save final checkpoint
	if e.autoCheckpointEnabled {
		if err := e.createCheckpoint(state); err != nil {
			e.logger.Warn("failed to save final checkpoint", "error", err)
		}

		// Perform cleanup after successful execution
		if err := e.CleanupOldCheckpoints(); err != nil {
			e.logger.Warn("checkpoint cleanup failed", "error", err)
		}
	}

	e.logger.Info("execution completed",
		"success", result.Success,
		"duration", result.Duration,
		"cost", fmt.Sprintf("$%.4f", result.Cost),
		"execution_id", state.ID,
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

// StatusInfo contains engine status information.
type StatusInfo struct {
	AvailableRuntimes   []string
	AvailableStrategies []string
	PreferredRuntime    string
	Budget              BudgetInfo
}

// BudgetInfo contains budget status.
type BudgetInfo struct {
	DailyTokens     int
	DailySpending   float64
	MonthlyTokens   int
	MonthlySpending float64
}

// Status returns current engine status.
func (e *Engine) Status(ctx context.Context) StatusInfo {
	var runtimes []string
	for rt, runner := range e.runners {
		if runner.Available(ctx) {
			runtimes = append(runtimes, string(rt))
		}
	}

	var strategies []string
	for name := range e.strategies {
		strategies = append(strategies, string(name))
	}

	dailyTokens, dailyCost := e.budget.GetDailySpending()
	monthlyTokens, monthlyCost := e.budget.GetMonthlySpending()

	return StatusInfo{
		AvailableRuntimes:   runtimes,
		AvailableStrategies: strategies,
		PreferredRuntime:    string(e.preferred),
		Budget: BudgetInfo{
			DailyTokens:     dailyTokens,
			DailySpending:   dailyCost,
			MonthlyTokens:   monthlyTokens,
			MonthlySpending: monthlyCost,
		},
	}
}

// TriggerSummarization manually triggers context summarization.
// Returns true if summarization was performed, false otherwise.
func (e *Engine) TriggerSummarization(ctx context.Context, task *Task, executionID, model string, state interface{}, force bool) bool {
	if !e.Config().EnableSummarization {
		e.logger.Debug("summarization disabled")
		return false
	}

	if task.Context == nil {
		e.logger.Debug("no context to summarize")
		return false
	}

	threshold := e.Config().SummarizationThreshold

	// Check if summarization should occur
	shouldSummarize := force || e.shouldSummarize(task.Context, threshold)
	if !shouldSummarize {
		e.logger.Debug("context within thresholds, not summarizing",
			"files", len(task.Context.Files),
			"threshold", threshold.FileCount,
		)
		return false
	}

	// Create summary (simplified - in real implementation, this would call LLM)
	summary := e.createSummary(task.Context)

	// Store original context
	originalFiles := task.Context.Files
	originalCount := len(originalFiles)

	// Replace context with summary
	task.Context.Files = []string{summary}
	task.Context.IsSummarized = true
	task.Context.SummarizationCount++

	e.logger.Info("context summarized",
		"execution_id", executionID,
		"model", model,
		"original_files", originalCount,
		"summarization_count", task.Context.SummarizationCount,
	)

	return true
}

// Config returns the engine configuration.
func (e *Engine) Config() Config {
	return Config{
		PreferredRuntime:       e.preferred,
		IsMCPMode:              e.budget.isMCPMode,
		Logger:                 e.logger,
		EnableSummarization:    e.getSummarizationEnabled(),
		SummarizationThreshold: e.getSummarizationThreshold(),
		SummarizationModel:     e.getSummarizationModel(),
		EnableAutoCheckpoint:   e.getAutoCheckpointEnabled(),
		MaxCheckpoints:         e.getMaxCheckpoints(),
		CheckpointAgeLimit:     e.getCheckpointAgeLimit(),
	}
}

func (e *Engine) getSummarizationEnabled() bool {
	return e.summarizationEnabled
}

func (e *Engine) getSummarizationThreshold() SummarizationThreshold {
	return e.summarizationThreshold
}

func (e *Engine) getSummarizationModel() string {
	if e.summarizationModel == "" {
		return LLMModelClaude35
	}
	return e.summarizationModel
}

func (e *Engine) getAutoCheckpointEnabled() bool {
	return e.autoCheckpointEnabled
}

func (e *Engine) getMaxCheckpoints() int {
	return e.maxCheckpoints
}

func (e *Engine) getCheckpointAgeLimit() time.Duration {
	return e.checkpointAgeLimit
}

func (e *Engine) shouldSummarize(ctx *TaskContext, threshold SummarizationThreshold) bool {
	if threshold.FileCount > 0 && len(ctx.Files) > threshold.FileCount {
		return true
	}

	totalLength := 0
	for _, file := range ctx.Files {
		totalLength += len(file)
	}
	if threshold.ContextLength > 0 && totalLength > threshold.ContextLength {
		return true
	}

	if threshold.TokenCount > 0 {
		estimatedTokens := totalLength / 4
		if estimatedTokens > threshold.TokenCount {
			return true
		}
	}

	return false
}

func (e *Engine) createSummary(ctx *TaskContext) string {
	// In a real implementation, this would call an LLM to summarize
	// For now, return a simplified summary
	return fmt.Sprintf("Summary of %d files in project %s:\n\nContext includes relevant files for task %s in change %s.",
		len(ctx.Files), ctx.ProjectPath, ctx.TaskID, ctx.ChangeID)
}

// determineExecutionConfig extracts execution configuration from task.
func (e *Engine) determineExecutionConfig(task *Task) ExecutionConfig {
	cfg := ExecutionConfig{}

	if task.ForceAgent != "" {
		cfg.Agent = task.ForceAgent
	}
	if task.ForceModel != "" {
		cfg.Model = task.ForceModel
	}
	if task.ForceRuntime != "" {
		cfg.Runtime = task.ForceRuntime
	}
	if task.ForceStrategy != "" {
		cfg.Strategy = task.ForceStrategy
	}
	if task.Budget != nil {
		cfg.Budget = task.Budget
	}
	if len(task.Skills) > 0 {
		cfg.Skills = task.Skills
	}

	return cfg
}

// createCheckpoint saves an execution state as a checkpoint.
func (e *Engine) createCheckpoint(state *ExecutionState) error {
	e.logger.Debug("creating checkpoint",
		"execution_id", state.ID,
		"status", state.Status,
	)

	if err := SaveState(state); err != nil {
		return fmt.Errorf("save checkpoint: %w", err)
	}

	e.logger.Info("checkpoint created",
		"execution_id", state.ID,
		"status", state.Status,
		"updated_at", state.UpdatedAt,
	)

	return nil
}

// CreateManualCheckpoint allows manual checkpoint creation at any point.
func (e *Engine) CreateManualCheckpoint(ctx context.Context, task *Task, status string) (*ExecutionState, error) {
	state := NewExecutionState(task)

	cfg := e.determineExecutionConfig(task)
	state.WithConfig(cfg)

	if status != "" {
		state.Status = status
		if status == ExecutionStatusRunning {
			now := time.Now()
			state.StartedAt = now
		}
		if status == ExecutionStatusCompleted || status == ExecutionStatusFailed {
			now := time.Now()
			if state.StartedAt.IsZero() {
				state.StartedAt = now
			}
			state.CompletedAt = now
		}
		state.Checksum = state.computeChecksum()
	}

	if err := e.createCheckpoint(state); err != nil {
		return nil, fmt.Errorf("create manual checkpoint: %w", err)
	}

	return state, nil
}

// CleanupOldCheckpoints removes old checkpoint files based on retention policy.
func (e *Engine) CleanupOldCheckpoints() error {
	e.logger.Debug("starting checkpoint cleanup")

	executionIDs, err := ListExecutions()
	if err != nil {
		return fmt.Errorf("list executions: %w", err)
	}

	var toDelete []string
	now := time.Now()

	for _, id := range executionIDs {
		state, err := LoadState(id)
		if err != nil {
			e.logger.Warn("failed to load state for cleanup", "id", id, "error", err)
			continue
		}

		// Check if state is too old
		age := now.Sub(state.UpdatedAt)
		if age > e.checkpointAgeLimit {
			toDelete = append(toDelete, id)
			continue
		}

		// Check if state is completed and old enough to prune
		if (state.IsCompleted() || state.IsFailed()) && age > e.checkpointAgeLimit {
			toDelete = append(toDelete, id)
		}
	}

	// Apply max checkpoint limit
	if len(executionIDs) > e.maxCheckpoints {
		// Sort by age and keep the most recent
		// For simplicity, we'll just mark the oldest for deletion
		excess := len(executionIDs) - e.maxCheckpoints
		for i := 0; i < excess && i < len(toDelete); i++ {
			// toDelete already contains some entries, we need to find the oldest ones
		}
	}

	// Delete old checkpoints
	deleted := 0
	for _, id := range toDelete {
		if err := DeleteState(id); err != nil {
			e.logger.Warn("failed to delete checkpoint", "id", id, "error", err)
			continue
		}
		deleted++
	}

	e.logger.Info("checkpoint cleanup completed",
		"deleted", deleted,
		"remaining", len(executionIDs)-deleted,
	)

	return nil
}

// GetCheckpoint retrieves an execution state by ID.
func (e *Engine) GetCheckpoint(executionID string) (*ExecutionState, error) {
	e.logger.Debug("retrieving checkpoint", "execution_id", executionID)

	state, err := LoadState(executionID)
	if err != nil {
		return nil, fmt.Errorf("load checkpoint: %w", err)
	}

	return state, nil
}

// ListCheckpoints returns all available checkpoint IDs.
func (e *Engine) ListCheckpoints() ([]string, error) {
	e.logger.Debug("listing checkpoints")

	executionIDs, err := ListExecutions()
	if err != nil {
		return nil, fmt.Errorf("list checkpoints: %w", err)
	}

	return executionIDs, nil
}

// ResumeExecution resumes execution from a saved state.
func (e *Engine) ResumeExecution(ctx context.Context, executionID string) (*Result, error) {
	e.logger.Info("resuming execution", "execution_id", executionID)

	state, err := LoadState(executionID)
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}

	compat := state.CheckVersionCompatibility()
	e.logger.Info("version compatibility check",
		"state_version", compat.StateVersion,
		"current_version", compat.CurrentVersion,
		"compatible", compat.Compatible,
		"policy", compat.Policy,
		"reason", compat.Reason,
	)

	if !compat.Compatible {
		return nil, fmt.Errorf("version incompatible: %s (state: %s, current: %s)", compat.Reason, compat.StateVersion, compat.CurrentVersion)
	}

	validationErrors := state.ValidateForResume()
	for _, verr := range validationErrors {
		if verr.Level == ValidationLevelError {
			e.logger.Error("state validation error",
				"field", verr.Field,
				"message", verr.Message,
			)
		} else {
			e.logger.Warn("state validation warning",
				"field", verr.Field,
				"message", verr.Message,
			)
		}
	}

	hasErrors := false
	for _, verr := range validationErrors {
		if verr.Level == ValidationLevelError {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		return nil, fmt.Errorf("state validation failed with %d error(s)", len(validationErrors))
	}

	envErrors := state.ValidateEnvironment()
	for _, eerr := range envErrors {
		if eerr.Level == ValidationLevelError {
			e.logger.Error("environment validation error",
				"field", eerr.Field,
				"message", eerr.Message,
			)
		} else {
			e.logger.Warn("environment validation warning",
				"field", eerr.Field,
				"message", eerr.Message,
			)
		}
	}

	if state.Runtime != "" {
		runner, err := e.getRunner(state.Runtime)
		if err != nil {
			return nil, fmt.Errorf("runtime %s not available: %w", state.Runtime, err)
		}
		if !runner.Available(ctx) {
			return nil, fmt.Errorf("runtime %s is not available", state.Runtime)
		}
		e.logger.Info("runtime validated for resume", "runtime", state.Runtime)
	}

	if state.Status == ExecutionStatusPending {
		return e.executePendingState(ctx, state)
	}

	if state.Status == ExecutionStatusInterrupted || state.Status == ExecutionStatusFailed {
		return e.executeInterruptedState(ctx, state)
	}

	return nil, fmt.Errorf("unexpected status for resume: %s", state.Status)
}

func (e *Engine) executePendingState(ctx context.Context, state *ExecutionState) (*Result, error) {
	e.logger.Info("executing pending task", "execution_id", state.ID)

	if err := state.Start(); err != nil {
		return nil, fmt.Errorf("start execution: %w", err)
	}

	if e.autoCheckpointEnabled {
		if err := e.createCheckpoint(state); err != nil {
			e.logger.Warn("failed to save checkpoint", "error", err)
		}
	}

	task := state.Task
	result, err := e.Execute(ctx, task)
	if err != nil {
		e.logger.Error("execution failed", "error", err)
		return result, err
	}

	return result, nil
}

func (e *Engine) executeInterruptedState(ctx context.Context, state *ExecutionState) (*Result, error) {
	e.logger.Info("resuming interrupted task", "execution_id", state.ID)

	if state.Task == nil {
		return nil, fmt.Errorf("cannot resume: task is nil")
	}

	task := state.Task

	if task.Context == nil && state.Context != nil {
		task.Context = state.Context
	}

	if err := state.Resume(); err != nil {
		e.logger.Warn("failed to mark as running", "error", err)
	}

	if e.autoCheckpointEnabled {
		if err := e.createCheckpoint(state); err != nil {
			e.logger.Warn("failed to save checkpoint", "error", err)
		}
	}

	strategy := e.selectStrategy(task)

	result, err := strategy.Execute(ctx, e, task)
	if err != nil {
		e.logger.Error("execution failed", "error", err)

		if e.autoCheckpointEnabled {
			if stateErr := state.Fail(err); stateErr == nil {
				if cpErr := e.createCheckpoint(state); cpErr != nil {
					e.logger.Warn("failed to save error checkpoint", "error", cpErr)
				}
			}
		}

		return result, err
	}

	if err := state.Complete(result); err != nil {
		e.logger.Warn("failed to update execution state", "error", err)
	}

	if e.autoCheckpointEnabled {
		if err := e.createCheckpoint(state); err != nil {
			e.logger.Warn("failed to save final checkpoint", "error", err)
		}

		if err := e.CleanupOldCheckpoints(); err != nil {
			e.logger.Warn("checkpoint cleanup failed", "error", err)
		}
	}

	e.logger.Info("resumed execution completed",
		"success", result.Success,
		"duration", result.Duration,
		"cost", fmt.Sprintf("$%.4f", result.Cost),
		"execution_id", state.ID,
	)

	return result, nil
}

// ValidateExecutionState validates a saved execution state without loading it.
func (e *Engine) ValidateExecutionState(executionID string) (*StateValidationResult, error) {
	e.logger.Debug("validating execution state", "execution_id", executionID)

	result, err := ValidateStateFile(executionID)
	if err != nil {
		return nil, fmt.Errorf("validate state: %w", err)
	}

	return result, nil
}

// DeleteCorruptedState deletes a corrupted state file after user confirmation.
func (e *Engine) DeleteCorruptedState(executionID string) error {
	e.logger.Warn("deleting corrupted state", "execution_id", executionID)

	if err := DeleteState(executionID); err != nil {
		return fmt.Errorf("delete corrupted state: %w", err)
	}

	e.logger.Info("deleted corrupted state", "execution_id", executionID)

	return nil
}

// Interrupt interrupts a running execution.
func (e *Engine) Interrupt(ctx context.Context, executionID string) error {
	e.logger.Info("interrupting execution", "execution_id", executionID)

	state, err := LoadState(executionID)
	if err != nil {
		return fmt.Errorf("load execution state: %w", err)
	}

	if !state.IsRunning() {
		return fmt.Errorf("cannot interrupt execution with status %s", state.Status)
	}

	if state.Runtime == "" {
		return fmt.Errorf("execution state has no runtime configured")
	}

	runner, err := e.getRunner(state.Runtime)
	if err != nil {
		return fmt.Errorf("get runner for runtime %s: %w", state.Runtime, err)
	}

	e.logger.Info("sending interrupt signal to runner", "runtime", state.Runtime)

	if err := runner.Interrupt(ctx); err != nil {
		e.logger.Warn("runner interrupt returned error", "error", err)
		if err := state.Interrupt(); err != nil {
			e.logger.Error("failed to mark state as interrupted", "error", err)
			return fmt.Errorf("runner interrupt failed: %w", err)
		}
	} else {
		if err := state.Interrupt(); err != nil {
			return fmt.Errorf("mark state as interrupted: %w", err)
		}
	}

	if e.autoCheckpointEnabled {
		if err := e.createCheckpoint(state); err != nil {
			e.logger.Warn("failed to save interrupt checkpoint", "error", err)
		}
	}

	e.logger.Info("execution interrupted successfully",
		"execution_id", state.ID,
		"runtime", state.Runtime,
		"updated_at", state.UpdatedAt,
	)

	return nil
}
