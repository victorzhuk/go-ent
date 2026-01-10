package execution

import (
	"context"
	"fmt"

	"github.com/victorzhuk/go-ent/internal/domain"
)

// MultiStrategy executes tasks with multiple agents in conversation.
type MultiStrategy struct{}

// NewMultiStrategy creates a new multi-agent execution strategy.
func NewMultiStrategy() *MultiStrategy {
	return &MultiStrategy{}
}

// Name returns the strategy identifier.
func (m *MultiStrategy) Name() domain.ExecutionStrategy {
	return domain.ExecutionStrategyMulti
}

// Execute runs the task using multiple agents with handoffs.
func (m *MultiStrategy) Execute(ctx context.Context, engine *Engine, task *Task) (*Result, error) {
	// Multi-agent strategy not yet implemented
	// For now, fall back to single strategy
	return nil, fmt.Errorf("multi-agent strategy not yet implemented")
}

// CanHandle checks if this strategy can handle the task.
func (m *MultiStrategy) CanHandle(task *Task) bool {
	// Multi-agent strategy for complex tasks
	// For now, return false to use single strategy
	return false
}
