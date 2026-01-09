package agent

import (
	"context"
	"fmt"

	"github.com/victorzhuk/go-ent/internal/domain"
)

// Selector chooses the optimal agent role and skills for a given task.
type Selector struct {
	analyzer   *Complexity
	registry   SkillRegistry
	maxBudget  int
	strictMode bool
}

// Config holds selector configuration.
type Config struct {
	MaxBudget  int
	StrictMode bool
}

// NewSelector creates a new agent selector.
func NewSelector(cfg Config, registry SkillRegistry) *Selector {
	return &Selector{
		analyzer:   NewComplexity(),
		registry:   registry,
		maxBudget:  cfg.MaxBudget,
		strictMode: cfg.StrictMode,
	}
}

// SelectionResult holds the selected agent configuration.
type SelectionResult struct {
	Role   domain.AgentRole
	Model  string
	Skills []string
	Reason string
}

// Select analyzes a task and returns the optimal agent configuration.
func (s *Selector) Select(ctx context.Context, task Task) (*SelectionResult, error) {
	if err := task.Validate(); err != nil {
		return nil, fmt.Errorf("invalid task: %w", err)
	}

	complexity := s.analyzer.Analyze(task)
	role := s.selectRole(complexity)
	model := s.selectModel(role, complexity)
	skills := s.matchSkills(ctx, task, role)

	return &SelectionResult{
		Role:   role,
		Model:  model,
		Skills: skills,
		Reason: fmt.Sprintf("complexity=%s, task_type=%s", complexity.Level, task.Type),
	}, nil
}

func (s *Selector) selectRole(complexity TaskComplexity) domain.AgentRole {
	switch complexity.Level {
	case ComplexityArchitectural:
		return domain.AgentRoleArchitect
	case ComplexityComplex:
		return domain.AgentRoleSenior
	case ComplexityModerate:
		return domain.AgentRoleDeveloper
	case ComplexitySimple:
		return domain.AgentRoleDeveloper
	case ComplexityTrivial:
		return domain.AgentRoleDeveloper
	default:
		return domain.AgentRoleDeveloper
	}
}

func (s *Selector) selectModel(role domain.AgentRole, complexity TaskComplexity) string {
	switch role {
	case domain.AgentRoleArchitect:
		return "opus"
	case domain.AgentRoleSenior:
		if complexity.Level >= ComplexityComplex {
			return "opus"
		}
		return "sonnet"
	case domain.AgentRoleReviewer:
		return "opus"
	case domain.AgentRoleDeveloper:
		if complexity.Level >= ComplexityModerate {
			return "sonnet"
		}
		return "haiku"
	case domain.AgentRoleOps:
		return "sonnet"
	default:
		return "sonnet"
	}
}

func (s *Selector) matchSkills(ctx context.Context, task Task, role domain.AgentRole) []string {
	skillCtx := domain.SkillContext{
		Action:   task.Action,
		Phase:    task.Phase,
		Agent:    role,
		Metadata: task.Metadata,
	}

	return s.registry.MatchForContext(skillCtx)
}

// Task represents a development task to be analyzed.
type Task struct {
	Description string
	Type        TaskType
	Action      domain.SpecAction
	Phase       domain.ActionPhase
	Files       []string
	Metadata    map[string]interface{}
}

// TaskType categorizes the kind of task.
type TaskType string

const (
	TaskTypeFeature       TaskType = "feature"
	TaskTypeBugFix        TaskType = "bugfix"
	TaskTypeRefactor      TaskType = "refactor"
	TaskTypeTest          TaskType = "test"
	TaskTypeDocumentation TaskType = "documentation"
	TaskTypeArchitecture  TaskType = "architecture"
)

// Validate checks if the task is valid.
func (t *Task) Validate() error {
	if t.Description == "" {
		return fmt.Errorf("task description required")
	}
	if t.Type == "" {
		return fmt.Errorf("task type required")
	}
	return nil
}

// SkillRegistry defines the interface for skill matching.
type SkillRegistry interface {
	MatchForContext(ctx domain.SkillContext) []string
}
