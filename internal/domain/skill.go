package domain

import "context"

// Skill defines a reusable capability that agents can execute.
// Implementations can be built-in, custom, or provided by plugins.
type Skill interface {
	// Name returns the unique identifier for this skill.
	Name() string

	// Description returns a human-readable description of what this skill does.
	Description() string

	// CanHandle determines if this skill can handle the given context.
	// Returns true if the skill is applicable for the current situation.
	CanHandle(ctx SkillContext) bool

	// Execute performs the skill's operation.
	// Returns the result of execution or an error if the operation fails.
	Execute(ctx context.Context, req SkillRequest) (SkillResult, error)
}

// SkillMetadata holds information about a skill.
type SkillMetadata struct {
	// Name is the unique skill identifier.
	Name string

	// Description explains what the skill does.
	Description string

	// Version is the skill version (e.g., "1.0.0").
	Version string

	// Author is the skill creator.
	Author string

	// Tags are keywords for skill discovery.
	Tags []string
}

// SkillContext provides context for skill execution decisions.
type SkillContext struct {
	// Action is the type of action being performed.
	Action SpecAction

	// Phase is the current development phase.
	Phase ActionPhase

	// Runtime is the execution environment.
	Runtime Runtime

	// Agent is the role of the executing agent.
	Agent AgentRole

	// Metadata holds additional context information.
	Metadata map[string]interface{}
}

// SkillRequest represents a request to execute a skill.
type SkillRequest struct {
	// Input contains the input data for the skill.
	Input string

	// Parameters holds skill-specific parameters.
	Parameters map[string]interface{}

	// Context provides execution context.
	Context SkillContext
}

// SkillResult represents the outcome of skill execution.
type SkillResult struct {
	// Success indicates if the skill executed successfully.
	Success bool

	// Output contains the skill's output data.
	Output string

	// Error contains the error message if execution failed.
	Error string

	// Metadata holds additional result information.
	Metadata map[string]interface{}
}
