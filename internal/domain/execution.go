package domain

import "time"

// ExecutionStrategy defines how agents execute tasks.
type ExecutionStrategy string

// Execution strategy constants define the available execution modes.
const (
	// ExecutionStrategySingle represents single-agent sequential execution.
	// One agent handles the entire task from start to finish.
	ExecutionStrategySingle ExecutionStrategy = "single"

	// ExecutionStrategyMulti represents multi-agent conversation execution.
	// Multiple agents collaborate through handoffs and discussions.
	ExecutionStrategyMulti ExecutionStrategy = "multi"

	// ExecutionStrategyParallel represents parallel execution.
	// Independent agents work simultaneously on different subtasks.
	ExecutionStrategyParallel ExecutionStrategy = "parallel"
)

// String returns the string representation of the execution strategy.
func (s ExecutionStrategy) String() string {
	return string(s)
}

// Valid returns true if the execution strategy is valid.
func (s ExecutionStrategy) Valid() bool {
	switch s {
	case ExecutionStrategySingle, ExecutionStrategyMulti, ExecutionStrategyParallel:
		return true
	default:
		return false
	}
}

// ExecutionContext holds the context for task execution.
type ExecutionContext struct {
	// Runtime is the execution environment.
	Runtime Runtime

	// Agent is the role of the primary agent.
	Agent AgentRole

	// Strategy is the execution approach.
	Strategy ExecutionStrategy

	// ChangeID identifies the change being worked on.
	ChangeID string

	// TaskID identifies the specific task within the change.
	TaskID string

	// Budget is the token budget for this execution.
	Budget int

	// Metadata holds additional execution metadata.
	Metadata map[string]string
}

// ExecutionResult captures the outcome of task execution.
type ExecutionResult struct {
	// Success indicates if the execution completed successfully.
	Success bool

	// Output contains the execution output or result.
	Output string

	// Error contains the error message if execution failed.
	Error string

	// Tokens is the number of tokens consumed.
	Tokens int

	// Cost is the estimated cost in credits or currency.
	Cost float64

	// Duration is how long the execution took.
	Duration time.Duration

	// Metadata holds additional result metadata.
	Metadata map[string]interface{}
}
