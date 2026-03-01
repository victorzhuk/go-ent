package repository

import (
	"fmt"
	"sync"

	skilldomain "github.com/victorzhuk/go-ent/internal/skill/domain"
)

// InMemoryRepository provides thread-safe in-memory storage for skills.
type InMemoryRepository struct {
	mu     sync.RWMutex
	skills map[string]*skilldomain.Info
	names  map[string]string // name -> id mapping
}

// NewInMemoryRepository creates a new in-memory repository.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		skills: make(map[string]*skilldomain.Info),
		names:  make(map[string]string),
	}
}

// Save stores a new skill in the repository.
func (r *InMemoryRepository) Save(skill *skilldomain.Info) error {
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
		return fmt.Errorf("%w: %s", skilldomain.ErrDuplicate, skill.Name)
	}

	if _, exists := r.names[skill.Name]; exists {
		return fmt.Errorf("%w: %s", skilldomain.ErrDuplicate, skill.Name)
	}

	r.skills[skill.ID] = skill
	r.names[skill.Name] = skill.ID
	return nil
}

// FindByID retrieves a skill by its ID.
func (r *InMemoryRepository) FindByID(id string) (*skilldomain.Info, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill, exists := r.skills[id]
	if !exists {
		return nil, fmt.Errorf("%w: %s", skilldomain.ErrNotFound, id)
	}
	return skill, nil
}

// FindByName retrieves a skill by its name.
func (r *InMemoryRepository) FindByName(name string) (*skilldomain.Info, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.names[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", skilldomain.ErrNotFound, name)
	}
	return r.skills[id], nil
}

// ListAll returns all skills in the repository.
func (r *InMemoryRepository) ListAll() ([]*skilldomain.Info, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]*skilldomain.Info, 0, len(r.skills))
	for _, skill := range r.skills {
		skills = append(skills, skill)
	}
	return skills, nil
}

// Delete removes a skill by ID.
func (r *InMemoryRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	skill, exists := r.skills[id]
	if !exists {
		return fmt.Errorf("%w: %s", skilldomain.ErrNotFound, id)
	}

	delete(r.skills, id)
	delete(r.names, skill.Name)
	return nil
}

// Update updates an existing skill.
func (r *InMemoryRepository) Update(skill *skilldomain.Info) error {
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
		return fmt.Errorf("%w: %s", skilldomain.ErrNotFound, skill.ID)
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
func (r *InMemoryRepository) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.names[name]
	return exists
}
