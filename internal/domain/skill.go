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

type SkillMetadata struct {
	Name         string
	Description  string
	Version      string
	Author       string
	Tags         []string
	AllowedTools []string
}

type SkillContext struct {
	Action   SpecAction
	Phase    ActionPhase
	Agent    AgentRole
	Metadata map[string]any
}

type SkillRequest struct {
	Input      string
	Parameters map[string]any
	Context    SkillContext
}

type SkillResult struct {
	Success  bool
	Output   string
	Error    string
	Metadata map[string]any
}
