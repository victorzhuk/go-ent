package background

import "fmt"

// RegistryStats holds statistics about registered agents.
type RegistryStats struct {
	Pending   int
	Running   int
	Completed int
	Failed    int
	Killed    int
	Total     int
}

// Registry provides read-only access to managed background agents.
type Registry struct {
	manager *Manager
}

// NewRegistry creates a new registry backed by the given manager.
func NewRegistry(manager *Manager) *Registry {
	return &Registry{
		manager: manager,
	}
}

// Get retrieves an agent by ID.
func (r *Registry) Get(id string) (*Agent, error) {
	if r.manager == nil {
		return nil, fmt.Errorf("manager not initialized")
	}
	return r.manager.Get(id)
}

// List returns agents filtered by status, or all if status is empty.
func (r *Registry) List(status Status) []*Agent {
	if r.manager == nil {
		return nil
	}
	return r.manager.List(status)
}

// ListAll returns all agents regardless of status.
func (r *Registry) ListAll() []*Agent {
	if r.manager == nil {
		return nil
	}
	return r.manager.List("")
}

// Count returns the total number of agents.
func (r *Registry) Count() int {
	if r.manager == nil {
		return 0
	}
	return r.manager.Count()
}

// CountByStatus returns the number of agents with the given status.
func (r *Registry) CountByStatus(status Status) int {
	if r.manager == nil {
		return 0
	}
	return r.manager.CountByStatus(status)
}

// GetStats returns statistics about registered agents.
func (r *Registry) GetStats() RegistryStats {
	if r.manager == nil {
		return RegistryStats{}
	}

	return RegistryStats{
		Pending:   r.manager.CountByStatus(StatusPending),
		Running:   r.manager.CountByStatus(StatusRunning),
		Completed: r.manager.CountByStatus(StatusCompleted),
		Failed:    r.manager.CountByStatus(StatusFailed),
		Killed:    r.manager.CountByStatus(StatusKilled),
		Total:     r.manager.Count(),
	}
}
