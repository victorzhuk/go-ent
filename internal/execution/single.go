package execution

import (
	"context"
	"fmt"

	"github.com/victorzhuk/go-ent/internal/domain"
)

// SingleStrategy executes tasks with a single agent sequentially.
type SingleStrategy struct{}

// NewSingleStrategy creates a new single-agent execution strategy.
func NewSingleStrategy() *SingleStrategy {
	return &SingleStrategy{}
}

// Name returns the strategy identifier.
func (s *SingleStrategy) Name() domain.ExecutionStrategy {
	return domain.ExecutionStrategySingle
}

// Execute runs the task using a single agent.
func (s *SingleStrategy) Execute(ctx context.Context, engine *Engine, task *Task) (*Result, error) {
	// Select agent/model/skills if not forced
	agent := task.ForceAgent
	model := task.ForceModel
	skills := task.Skills

	if agent == "" || model == "" {
		// Use engine's selector for auto-selection
		selected, err := engine.selectAgent(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("agent selection: %w", err)
		}
		if agent == "" {
			agent = selected.Agent
		}
		if model == "" {
			model = selected.Model
		}
		if len(skills) == 0 {
			skills = selected.Skills
		}
	}

	// Build execution request
	req := &Request{
		Task:     task.Description,
		Agent:    agent,
		Model:    model,
		Skills:   skills,
		Strategy: domain.ExecutionStrategySingle,
		Budget:   task.Budget,
		Context:  task.Context,
		Metadata: task.Metadata,
	}

	// Select runtime
	runtime := task.ForceRuntime
	if runtime == "" {
		runtime = engine.selectRuntime(ctx)
	}

	// Get runner
	runner, err := engine.getRunner(runtime)
	if err != nil {
		return nil, fmt.Errorf("get runner: %w", err)
	}

	// Check budget before execution
	if task.Budget != nil {
		estimate := NewCostEstimate(model, 2000, 1000) // Rough estimate
		if err := engine.budget.Check(ctx, estimate, task.Budget); err != nil {
			return nil, fmt.Errorf("budget check: %w", err)
		}
	}

	// Execute
	result, err := runner.Execute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("execution: %w", err)
	}

	// Record spending
	if result.Success {
		cost := CalculateCost(model, result.TokensIn, result.TokensOut)
		result.Cost = cost

		taskID := ""
		if task.Context != nil {
			taskID = task.Context.String()
		}
		engine.budget.Record(taskID, result.TokensIn, result.TokensOut, cost)
	}

	return result, nil
}

// CanHandle checks if this strategy can handle the task.
func (s *SingleStrategy) CanHandle(task *Task) bool {
	// Single strategy can handle any task
	return true
}
