package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// AgentMeta represents parsed agent metadata from agent markdown files.
type AgentMeta struct {
	Name        string
	Description string
	Model       string
	Color       string
	Skills      []string
	Tools       map[string]bool
	Content     string
	FilePath    string
}

// Registry manages agent metadata.
type Registry struct {
	mu     sync.RWMutex
	agents map[string]AgentMeta
	parser *Parser
}

// NewRegistry creates a new agent registry.
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]AgentMeta),
		parser: NewParser(),
	}
}

// Load scans a directory for agent markdown files and loads their metadata.
func (r *Registry) Load(agentsPath string) error {
	entries, err := os.ReadDir(agentsPath)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		path := filepath.Join(agentsPath, name)
		meta, err := r.parser.ParseAgentFile(path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", name, err)
		}

		r.agents[meta.Name] = *meta
	}

	return nil
}

// Get retrieves an agent by name.
func (r *Registry) Get(name string) (AgentMeta, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[name]
	if !exists {
		return AgentMeta{}, fmt.Errorf("agent not found: %s", name)
	}
	return agent, nil
}

// All returns all loaded agents.
func (r *Registry) All() []AgentMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]AgentMeta, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	return agents
}

// Has checks if an agent exists.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.agents[name]
	return exists
}

// Count returns the number of loaded agents.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.agents)
}

// RegisterAgent registers an agent from a file path.
func (r *Registry) RegisterAgent(name, path string) error {
	meta, err := r.parser.ParseAgentFile(path)
	if err != nil {
		return fmt.Errorf("parse agent file: %w", err)
	}

	if meta.Name != name {
		return fmt.Errorf("agent name mismatch: expected %s, got %s", name, meta.Name)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[name]; exists {
		return fmt.Errorf("agent %s already registered", name)
	}

	r.agents[name] = *meta
	return nil
}

// UnregisterAgent removes an agent by name.
func (r *Registry) UnregisterAgent(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[name]; !exists {
		return fmt.Errorf("agent not found: %s", name)
	}

	delete(r.agents, name)
	return nil
}
