package execution

import (
	"context"

	"github.com/victorzhuk/go-ent/internal/domain"
)

// Strategy defines an execution approach.
type Strategy interface {
	// Name returns the strategy identifier.
	Name() domain.ExecutionStrategy

	// Execute runs the task using this strategy.
	Execute(ctx context.Context, engine *Engine, task *Task) (*Result, error)

	// CanHandle checks if this strategy can handle the task.
	CanHandle(task *Task) bool
}

// Task extends execution request with task metadata.
type Task struct {
	// Description is the task description.
	Description string

	// Type identifies the task type (feature, bugfix, etc.).
	Type string

	// Context provides project and file context.
	Context *TaskContext

	// Agent overrides auto-selected agent.
	ForceAgent domain.AgentRole

	// Model overrides auto-selected model.
	ForceModel string

	// Runtime overrides auto-selected runtime.
	ForceRuntime domain.Runtime

	// Strategy overrides auto-selected strategy.
	ForceStrategy domain.ExecutionStrategy

	// Budget specifies spending limits.
	Budget *BudgetLimit

	// Skills to activate.
	Skills []string

	// Metadata holds additional task data.
	Metadata map[string]interface{}
}

// NewTask creates a new task with the given description.
func NewTask(description string) *Task {
	return &Task{
		Description: description,
		Metadata:    make(map[string]interface{}),
	}
}

// WithContext sets the task context.
func (t *Task) WithContext(ctx *TaskContext) *Task {
	t.Context = ctx
	return t
}

// WithType sets the task type.
func (t *Task) WithType(typ string) *Task {
	t.Type = typ
	return t
}

// WithAgent forces a specific agent role.
func (t *Task) WithAgent(agent domain.AgentRole) *Task {
	t.ForceAgent = agent
	return t
}

// WithModel forces a specific model.
func (t *Task) WithModel(model string) *Task {
	t.ForceModel = model
	return t
}

// WithRuntime forces a specific runtime.
func (t *Task) WithRuntime(runtime domain.Runtime) *Task {
	t.ForceRuntime = runtime
	return t
}

// WithStrategy forces a specific strategy.
func (t *Task) WithStrategy(strategy domain.ExecutionStrategy) *Task {
	t.ForceStrategy = strategy
	return t
}

// WithBudget sets the budget limit.
func (t *Task) WithBudget(budget *BudgetLimit) *Task {
	t.Budget = budget
	return t
}

// WithSkills sets the skills to activate.
func (t *Task) WithSkills(skills ...string) *Task {
	t.Skills = skills
	return t
}

// WithMetadata sets a metadata key-value pair.
func (t *Task) WithMetadata(key string, value interface{}) *Task {
	t.Metadata[key] = value
	return t
}
