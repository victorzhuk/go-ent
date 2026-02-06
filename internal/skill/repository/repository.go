package repository

import (
	"fmt"
	"sync"

	skilldomain "github.com/victorzhuk/go-ent/internal/skill/domain"
)

// Repository defines the interface for skill persistence operations.
type Repository interface {
	// Save stores a new skill in the repository.
	Save(skill *skilldomain.Skill) error

	// FindByID retrieves a skill by its ID.
	FindByID(id string) (*skilldomain.Skill, error)

	// FindByName retrieves a skill by its name.
	FindByName(name string) (*skilldomain.Skill, error)

	// ListAll returns all skills in the repository.
	ListAll() ([]*skilldomain.Skill, error)

	// Delete removes a skill by ID.
	Delete(id string) error

	// Update updates an existing skill.
	Update(skill *skilldomain.Skill) error

	// Exists checks if a skill with the given name exists.
	Exists(name string) bool
}

// inMemoryRepository implements Repository using in-memory storage with thread-safe operations.
type inMemoryRepository struct {
	mu     sync.RWMutex
	skills map[string]*skilldomain.Skill
	names  map[string]string // name -> id mapping
}

// NewInMemoryRepository creates a new in-memory repository.
func NewInMemoryRepository() Repository {
	return &inMemoryRepository{
		skills: make(map[string]*skilldomain.Skill),
		names:  make(map[string]string),
	}
}

// Save stores a new skill in the repository.
func (r *inMemoryRepository) Save(skill *skilldomain.Skill) error {
	if skill == nil {
		return fmt.Errorf("skill cannot be nil")
	}
	if skill.ID == "" {
		return fmt.Errorf("skill ID cannot be empty")
	}
	if skill.Name == "" {
		return fmt.Errorf("skill name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.skills[skill.ID]; exists {
		return fmt.Errorf("%w: %s", skilldomain.ErrDuplicateSkill, skill.Name)
	}

	if _, exists := r.names[skill.Name]; exists {
		return fmt.Errorf("%w: %s", skilldomain.ErrDuplicateSkill, skill.Name)
	}

	r.skills[skill.ID] = skill
	r.names[skill.Name] = skill.ID
	return nil
}

// FindByID retrieves a skill by its ID.
func (r *inMemoryRepository) FindByID(id string) (*skilldomain.Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill, exists := r.skills[id]
	if !exists {
		return nil, fmt.Errorf("%w: %s", skilldomain.ErrSkillNotFound, id)
	}
	return skill, nil
}

// FindByName retrieves a skill by its name.
func (r *inMemoryRepository) FindByName(name string) (*skilldomain.Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.names[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", skilldomain.ErrSkillNotFound, name)
	}
	return r.skills[id], nil
}

// ListAll returns all skills in the repository.
func (r *inMemoryRepository) ListAll() ([]*skilldomain.Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]*skilldomain.Skill, 0, len(r.skills))
	for _, skill := range r.skills {
		skills = append(skills, skill)
	}
	return skills, nil
}

// Delete removes a skill by ID.
func (r *inMemoryRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	skill, exists := r.skills[id]
	if !exists {
		return fmt.Errorf("%w: %s", skilldomain.ErrSkillNotFound, id)
	}

	delete(r.skills, id)
	delete(r.names, skill.Name)
	return nil
}

// Update updates an existing skill.
func (r *inMemoryRepository) Update(skill *skilldomain.Skill) error {
	if skill == nil {
		return fmt.Errorf("skill cannot be nil")
	}
	if skill.ID == "" {
		return fmt.Errorf("skill ID cannot be empty")
	}
	if skill.Name == "" {
		return fmt.Errorf("skill name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.skills[skill.ID]; !exists {
		return fmt.Errorf("%w: %s", skilldomain.ErrSkillNotFound, skill.ID)
	}

	oldName := ""
	for name, id := range r.names {
		if id == skill.ID {
			oldName = name
			break
		}
	}

	if oldName != "" && oldName != skill.Name {
		delete(r.names, oldName)
	}

	r.skills[skill.ID] = skill
	r.names[skill.Name] = skill.ID
	return nil
}

// Exists checks if a skill with the given name exists.
func (r *inMemoryRepository) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.names[name]
	return exists
}
