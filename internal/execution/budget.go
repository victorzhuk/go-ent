package execution

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Model pricing per 1M tokens (input/output).
const (
	// Opus pricing: $15 input, $75 output per 1M tokens
	OpusInputPer1M  = 15.0
	OpusOutputPer1M = 75.0

	// Sonnet pricing: $3 input, $15 output per 1M tokens
	SonnetInputPer1M  = 3.0
	SonnetOutputPer1M = 15.0

	// Haiku pricing: $0.25 input, $1.25 output per 1M tokens
	HaikuInputPer1M  = 0.25
	HaikuOutputPer1M = 1.25
)

// ModelPricing holds input and output pricing per 1M tokens.
type ModelPricing struct {
	InputPer1M  float64
	OutputPer1M float64
}

// GetModelPricing returns pricing for the given model.
func GetModelPricing(model string) ModelPricing {
	switch model {
	case "opus", "claude-opus-4", "claude-opus-4-5":
		return ModelPricing{OpusInputPer1M, OpusOutputPer1M}
	case "haiku", "claude-haiku-3", "claude-haiku-3-5":
		return ModelPricing{HaikuInputPer1M, HaikuOutputPer1M}
	default:
		// Default to Sonnet pricing
		return ModelPricing{SonnetInputPer1M, SonnetOutputPer1M}
	}
}

// CalculateCost calculates the cost for the given tokens and model.
func CalculateCost(model string, tokensIn, tokensOut int) float64 {
	pricing := GetModelPricing(model)
	costIn := (float64(tokensIn) / 1_000_000) * pricing.InputPer1M
	costOut := (float64(tokensOut) / 1_000_000) * pricing.OutputPer1M
	return costIn + costOut
}

// CostEstimate represents an estimated cost for an execution.
type CostEstimate struct {
	Model       string
	TokensIn    int
	TokensOut   int
	Cost        float64
	Description string
}

// NewCostEstimate creates a cost estimate for the given tokens and model.
func NewCostEstimate(model string, tokensIn, tokensOut int) CostEstimate {
	return CostEstimate{
		Model:     model,
		TokensIn:  tokensIn,
		TokensOut: tokensOut,
		Cost:      CalculateCost(model, tokensIn, tokensOut),
	}
}

// SpendingWindow tracks spending over a time window.
type SpendingWindow struct {
	mu      sync.RWMutex
	tokens  int
	cost    float64
	started time.Time
}

// NewSpendingWindow creates a new spending window.
func NewSpendingWindow() *SpendingWindow {
	return &SpendingWindow{
		started: time.Now(),
	}
}

// Add adds spending to the window.
func (sw *SpendingWindow) Add(tokens int, cost float64) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.tokens += tokens
	sw.cost += cost
}

// Get returns current spending.
func (sw *SpendingWindow) Get() (tokens int, cost float64) {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.tokens, sw.cost
}

// Reset clears the spending window.
func (sw *SpendingWindow) Reset() {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.tokens = 0
	sw.cost = 0.0
	sw.started = time.Now()
}

// BudgetTracker monitors spending across executions.
type BudgetTracker struct {
	mu        sync.RWMutex
	daily     *SpendingWindow
	monthly   *SpendingWindow
	perTask   map[string]*SpendingWindow
	isMCPMode bool
	logger    *slog.Logger
}

// NewBudgetTracker creates a new budget tracker.
func NewBudgetTracker(isMCPMode bool, logger *slog.Logger) *BudgetTracker {
	if logger == nil {
		logger = slog.Default()
	}
	return &BudgetTracker{
		daily:     NewSpendingWindow(),
		monthly:   NewSpendingWindow(),
		perTask:   make(map[string]*SpendingWindow),
		isMCPMode: isMCPMode,
		logger:    logger,
	}
}

// Check verifies if execution can proceed within budget.
// In MCP mode with AutoProceed: logs warning and returns nil.
// In CLI/strict mode: returns error if budget exceeded.
func (bt *BudgetTracker) Check(ctx context.Context, estimate CostEstimate, limit *BudgetLimit) error {
	if limit == nil {
		return nil
	}

	bt.mu.RLock()
	_, dailyCost := bt.daily.Get()
	bt.mu.RUnlock()

	// Check if estimate would exceed limits
	exceeded := false
	var reason string

	if limit.MaxCost > 0 && dailyCost+estimate.Cost > limit.MaxCost {
		exceeded = true
		reason = fmt.Sprintf("estimated cost $%.4f would exceed daily limit $%.2f (current: $%.4f)",
			estimate.Cost, limit.MaxCost, dailyCost)
	}

	if limit.MaxTokens > 0 {
		dailyTokens, _ := bt.daily.Get()
		estimatedTotal := estimate.TokensIn + estimate.TokensOut
		if dailyTokens+estimatedTotal > limit.MaxTokens {
			exceeded = true
			if reason != "" {
				reason += "; "
			}
			reason += fmt.Sprintf("estimated %d tokens would exceed daily limit %d (current: %d)",
				estimatedTotal, limit.MaxTokens, dailyTokens)
		}
	}

	if !exceeded {
		return nil
	}

	// Budget exceeded
	if bt.isMCPMode || limit.AutoProceed {
		// MCP mode: log warning and proceed
		bt.logger.Warn("budget limit exceeded, auto-proceeding",
			"reason", reason,
			"mode", "mcp",
		)
		return nil
	}

	// CLI mode: return error for user approval
	return fmt.Errorf("budget exceeded: %s", reason)
}

// Record logs actual spending after execution.
func (bt *BudgetTracker) Record(taskID string, tokensIn, tokensOut int, cost float64) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	totalTokens := tokensIn + tokensOut

	bt.daily.Add(totalTokens, cost)
	bt.monthly.Add(totalTokens, cost)

	if taskID != "" {
		if _, exists := bt.perTask[taskID]; !exists {
			bt.perTask[taskID] = NewSpendingWindow()
		}
		bt.perTask[taskID].Add(totalTokens, cost)
	}

	bt.logger.Debug("recorded execution cost",
		"task_id", taskID,
		"tokens_in", tokensIn,
		"tokens_out", tokensOut,
		"cost", fmt.Sprintf("$%.4f", cost),
	)
}

// GetDailySpending returns daily spending totals.
func (bt *BudgetTracker) GetDailySpending() (tokens int, cost float64) {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	return bt.daily.Get()
}

// GetMonthlySpending returns monthly spending totals.
func (bt *BudgetTracker) GetMonthlySpending() (tokens int, cost float64) {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	return bt.monthly.Get()
}

// GetTaskSpending returns spending for a specific task.
func (bt *BudgetTracker) GetTaskSpending(taskID string) (tokens int, cost float64, exists bool) {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	window, exists := bt.perTask[taskID]
	if !exists {
		return 0, 0, false
	}
	tokens, cost = window.Get()
	return tokens, cost, true
}

// ResetDaily resets the daily spending window.
func (bt *BudgetTracker) ResetDaily() {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	bt.daily.Reset()
	bt.logger.Info("daily budget reset")
}

// ResetMonthly resets the monthly spending window.
func (bt *BudgetTracker) ResetMonthly() {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	bt.monthly.Reset()
	bt.logger.Info("monthly budget reset")
}
