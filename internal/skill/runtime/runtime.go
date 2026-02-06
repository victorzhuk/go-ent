package runtime

import (
	"context"
	"fmt"
	"sync"

	"github.com/victorzhuk/go-ent/internal/domain"
)

// Runtime defines the interface for skill runtime operations.
type Runtime interface {
	// RegisterSkill adds a runtime skill to runtime.
	RegisterSkill(skill domain.Skill) error

	// UnregisterSkill removes a runtime skill from runtime.
	UnregisterSkill(name string) error

	// GetSkill retrieves a runtime skill by name.
	GetSkill(name string) (domain.Skill, error)

	// Execute runs a skill with given context and request.
	Execute(ctx context.Context, name string, req domain.SkillRequest) (*domain.SkillResult, error)

	// ListSkills returns all registered runtime skills.
	ListSkills() []domain.Skill

	// Exists checks if a skill with the given name is registered.
	Exists(name string) bool
}

// skillRuntime implements Runtime with thread-safe operations.
type skillRuntime struct {
	mu     sync.RWMutex
	skills map[string]domain.Skill
}

// NewRuntime creates a new Runtime.
func NewRuntime() Runtime {
	return &skillRuntime{
		skills: make(map[string]domain.Skill),
	}
}

// RegisterSkill adds a runtime skill to runtime.
func (r *skillRuntime) RegisterSkill(skill domain.Skill) error {
	if skill == nil {
		return fmt.Errorf("skill cannot be nil")
	}

	name := skill.Name()
	if name == "" {
		return fmt.Errorf("skill name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.skills[name]; exists {
		return fmt.Errorf("%w: %s already registered", ErrDuplicateSkill, name)
	}

	r.skills[name] = skill
	return nil
}

// UnregisterSkill removes a runtime skill from runtime.
func (r *skillRuntime) UnregisterSkill(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.skills[name]; !exists {
		return fmt.Errorf("%w: %s", ErrSkillNotFound, name)
	}

	delete(r.skills, name)
	return nil
}

// GetSkill retrieves a runtime skill by name.
func (r *skillRuntime) GetSkill(name string) (domain.Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill, exists := r.skills[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrSkillNotFound, name)
	}
	return skill, nil
}

// Execute runs a skill with the given context and request.
func (r *skillRuntime) Execute(ctx context.Context, name string, req domain.SkillRequest) (*domain.SkillResult, error) {
	skill, err := r.GetSkill(name)
	if err != nil {
		return nil, fmt.Errorf("get skill: %w", err)
	}

	result, err := skill.Execute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("execute skill: %w", err)
	}

	return &result, nil
}

// ListSkills returns all registered runtime skills.
func (r *skillRuntime) ListSkills() []domain.Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]domain.Skill, 0, len(r.skills))
	for _, skill := range r.skills {
		skills = append(skills, skill)
	}
	return skills
}

// Exists checks if a skill with the given name is registered.
func (r *skillRuntime) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.skills[name]
	return exists
}

var (
	// ErrDuplicateSkill is returned when attempting to register a duplicate runtime skill.
	ErrDuplicateSkill = fmt.Errorf("duplicate runtime skill")

	// ErrSkillNotFound is returned when a runtime skill cannot be found.
	ErrSkillNotFound = fmt.Errorf("runtime skill not found")
)
