package background

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Manager manages background agent lifecycle.
type Manager struct {
	mu       sync.RWMutex
	agents   map[string]*Agent
	selector Selector
	cfg      Config
}

// Selector defines the interface for selecting agent configuration.
type Selector interface {
	Select(ctx context.Context, task interface{}) (*SelectionResult, error)
}

// SelectionResult represents a selected agent configuration.
type SelectionResult struct {
	Role   string
	Model  string
	Skills []string
	Reason string
}

// Config holds background agent configuration.
type Config struct {
	// MaxConcurrent is the maximum number of concurrent agents.
	MaxConcurrent int

	// DefaultRole is the default agent role to use.
	DefaultRole string

	// DefaultModel is the default model to use.
	DefaultModel string

	// Timeout is the maximum execution duration per agent.
	Timeout int
}

// DefaultConfig returns sensible default configuration.
func DefaultConfig() Config {
	return Config{
		MaxConcurrent: 5,
		DefaultRole:   "developer",
		DefaultModel:  "haiku",
		Timeout:       300,
	}
}

// NewManager creates a new background agent manager.
func NewManager(selector Selector, cfg Config) *Manager {
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 5
	}
	if cfg.DefaultRole == "" {
		cfg.DefaultRole = "developer"
	}
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = "haiku"
	}

	return &Manager{
		agents:   make(map[string]*Agent),
		selector: selector,
		cfg:      cfg,
	}
}

// Spawn creates and starts a new background agent.
func (m *Manager) Spawn(ctx context.Context, task string) (*Agent, error) {
	if task == "" {
		return nil, fmt.Errorf("task required")
	}

	id := uuid.New().String()

	role := m.cfg.DefaultModel
	model := m.cfg.DefaultModel

	if m.selector != nil {
		result, err := m.selector.Select(ctx, task)
		if err == nil && result != nil {
			role = result.Role
			model = result.Model
		}
	}

	agent, err := NewAgent(id, role, model, task)
	if err != nil {
		return nil, fmt.Errorf("new agent: %w", err)
	}

	m.mu.Lock()
	m.agents[id] = agent
	m.mu.Unlock()

	go m.runAgent(ctx, agent)

	return agent, nil
}

// runAgent executes the agent and updates its status.
func (m *Manager) runAgent(ctx context.Context, agent *Agent) {
	agent.Start()

	defer func() {
		if r := recover(); r != nil {
			agent.Fail(fmt.Errorf("panic: %v", r))
		}
	}()

	agent.Complete("task executed")
}

// Get retrieves an agent by ID.
func (m *Manager) Get(id string) (*Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, exists := m.agents[id]
	if !exists {
		return nil, ErrAgentNotFound
	}

	agent.mu.RLock()
	defer agent.mu.RUnlock()

	return agent, nil
}

// List returns all agents, optionally filtered by status.
func (m *Manager) List(status Status) []*Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var agents []*Agent
	for _, agent := range m.agents {
		agent.mu.RLock()
		match := status == "" || agent.Status == status
		agent.mu.RUnlock()
		if match {
			agents = append(agents, agent)
		}
	}
	return agents
}

// Kill terminates a running agent.
func (m *Manager) Kill(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, exists := m.agents[id]
	if !exists {
		return ErrAgentNotFound
	}

	agent.mu.RLock()
	agentStatus := agent.Status
	agent.mu.RUnlock()

	if agentStatus != StatusRunning {
		return fmt.Errorf("agent not running: %s", agentStatus)
	}

	agent.Kill()
	return nil
}

// Cleanup removes completed or failed agents.
func (m *Manager) Cleanup(ctx context.Context) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	var count int
	for id, agent := range m.agents {
		agent.mu.RLock()
		agentStatus := agent.Status
		agent.mu.RUnlock()

		if agentStatus == StatusCompleted || agentStatus == StatusFailed || agentStatus == StatusKilled {
			delete(m.agents, id)
			count++
		}
	}
	return count
}

// Count returns the total number of agents.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.agents)
}

// CountByStatus returns the number of agents with the given status.
func (m *Manager) CountByStatus(status Status) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var count int
	for _, agent := range m.agents {
		agent.mu.RLock()
		agentStatus := agent.Status
		agent.mu.RUnlock()

		if agentStatus == status {
			count++
		}
	}
	return count
}

// Shutdown gracefully stops all running agents and cleans up.
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, agent := range m.agents {
		agent.mu.RLock()
		agentStatus := agent.Status
		agent.mu.RUnlock()

		if agentStatus == StatusRunning {
			agent.Kill()
		}
	}

	m.agents = make(map[string]*Agent)
	return nil
}
