package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	agent, exists := r.agents[name]
	if !exists {
		return AgentMeta{}, fmt.Errorf("agent not found: %s", name)
	}
	return agent, nil
}

// All returns all loaded agents.
func (r *Registry) All() []AgentMeta {
	agents := make([]AgentMeta, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	return agents
}

// Has checks if an agent exists.
func (r *Registry) Has(name string) bool {
	_, exists := r.agents[name]
	return exists
}

// Count returns the number of loaded agents.
func (r *Registry) Count() int {
	return len(r.agents)
}
