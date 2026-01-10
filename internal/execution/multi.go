package execution

import (
	"context"
	"fmt"
	"strings"

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
	// Build agent chain based on task complexity
	chain := m.buildAgentChain(task)
	if len(chain) == 0 {
		return nil, fmt.Errorf("no agents selected for multi-agent execution")
	}

	// Execute agents in sequence
	var results []*Result
	var aggregatedOutput strings.Builder
	var totalTokensIn, totalTokensOut int
	var totalCost float64
	var adjustments []string

	// Start with task context
	currentContext := task.Context

	for i, agent := range chain {
		// Build request for this agent
		model := task.ForceModel
		if model == "" {
			// Select model for this agent
			selected, err := engine.selectAgent(ctx, task)
			if err != nil {
				return nil, fmt.Errorf("agent selection for %s: %w", agent, err)
			}
			model = selected.Model
		}

		req := &Request{
			Task:     task.Description,
			Agent:    agent,
			Model:    model,
			Skills:   task.Skills,
			Strategy: domain.ExecutionStrategyMulti,
			Budget:   task.Budget,
			Context:  currentContext,
			Metadata: task.Metadata,
		}

		// Add previous agent's output to context
		if i > 0 {
			if req.Metadata == nil {
				req.Metadata = make(map[string]interface{})
			}
			req.Metadata["previous_agent"] = chain[i-1].String()
			req.Metadata["previous_output"] = results[i-1].Output
		}

		// Select runtime
		runtime := task.ForceRuntime
		if runtime == "" {
			runtime = engine.selectRuntime(ctx)
		}

		// Get runner
		runner, err := engine.getRunner(runtime)
		if err != nil {
			return nil, fmt.Errorf("get runner for %s: %w", agent, err)
		}

		// Check budget before execution
		if task.Budget != nil {
			estimate := NewCostEstimate(model, 2000, 1000)
			if err := engine.budget.Check(ctx, estimate, task.Budget); err != nil {
				return nil, fmt.Errorf("budget check for %s: %w", agent, err)
			}
		}

		// Execute this agent
		result, err := runner.Execute(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("execution by %s: %w", agent, err)
		}

		// Collect results
		results = append(results, result)
		aggregatedOutput.WriteString(fmt.Sprintf("=== %s ===\n%s\n\n", agent, result.Output))
		totalTokensIn += result.TokensIn
		totalTokensOut += result.TokensOut
		adjustments = append(adjustments, result.Adjustments...)

		// Calculate and record cost
		if result.Success {
			cost := CalculateCost(model, result.TokensIn, result.TokensOut)
			totalCost += cost

			taskID := ""
			if task.Context != nil {
				taskID = fmt.Sprintf("%s-%s", task.Context.ChangeID, task.Context.TaskID)
			}
			engine.budget.Record(taskID, result.TokensIn, result.TokensOut, cost)
		}

		// If this agent failed, stop the chain
		if !result.Success {
			return &Result{
				Success:     false,
				Output:      aggregatedOutput.String(),
				Error:       fmt.Sprintf("agent %s failed: %s", agent, result.Error),
				TokensIn:    totalTokensIn,
				TokensOut:   totalTokensOut,
				Cost:        totalCost,
				Adjustments: adjustments,
				Metadata: map[string]interface{}{
					"agent_chain": chain,
					"failed_at":   agent,
				},
			}, nil
		}
	}

	// All agents succeeded
	return &Result{
		Success:     true,
		Output:      aggregatedOutput.String(),
		TokensIn:    totalTokensIn,
		TokensOut:   totalTokensOut,
		Cost:        totalCost,
		Adjustments: adjustments,
		Metadata: map[string]interface{}{
			"agent_chain": chain,
			"agents_used": len(chain),
		},
	}, nil
}

// buildAgentChain determines the agent sequence based on task properties.
func (m *MultiStrategy) buildAgentChain(task *Task) []domain.AgentRole {
	// Default multi-agent chain: Architect -> Developer
	chain := []domain.AgentRole{
		domain.AgentRoleArchitect,
		domain.AgentRoleDeveloper,
	}

	// For architectural tasks, add Reviewer
	if task.Type == "architecture" || task.Type == "refactor" {
		chain = append(chain, domain.AgentRoleReviewer)
	}

	return chain
}

// CanHandle checks if this strategy can handle the task.
func (m *MultiStrategy) CanHandle(task *Task) bool {
	// Multi-agent strategy for moderate to complex tasks
	// Can handle any task that benefits from design + implementation split
	return true
}
