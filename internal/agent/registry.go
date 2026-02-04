package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/victorzhuk/go-ent/internal/hooks"
	"gopkg.in/yaml.v3"
)

// Registry holds loaded agent metadata.
type Registry struct {
	agents map[string]AgentMeta
}

// AgentMeta represents agent metadata loaded from YAML.
type AgentMeta struct {
	Name                  string            `yaml:"name"`
	Description           string            `yaml:"description"`
	Model                 string            `yaml:"model"`
	Skills                []string          `yaml:"skills,omitempty"`
	ToolPresets           []string          `yaml:"toolPresets,omitempty"`
	DisallowedToolPresets []string          `yaml:"disallowedToolPresets,omitempty"`
	Role                  string            `yaml:"role,omitempty"`
	Complexity            string            `yaml:"complexity,omitempty"`
	ComplexityHints       map[string]string `yaml:"complexityHints,omitempty"`
	ModelMapping          map[string]string `yaml:"modelMapping,omitempty"`
	Dependencies          []string          `yaml:"dependencies,omitempty"`
	Color                 string            `yaml:"color,omitempty"`
	Hooks                 hooks.ToolHooks   `yaml:"hooks,omitempty"`
}

// NewRegistry creates a new agent registry.
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]AgentMeta),
	}
}

// Load loads all agents from a directory (e.g., pkg/agents/meta).
func (r *Registry) Load(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("read agents dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only load .yaml files
		if filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		fullPath := filepath.Join(path, entry.Name())
		// #nosec G304 - fullPath is constructed from validated directory entries
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("read agent %s: %w", entry.Name(), err)
		}

		var meta AgentMeta
		if err := yaml.Unmarshal(data, &meta); err != nil {
			return fmt.Errorf("parse agent %s: %w", entry.Name(), err)
		}

		r.agents[meta.Name] = meta
	}

	return nil
}

// Get retrieves an agent by name.
func (r *Registry) Get(name string) (AgentMeta, bool) {
	meta, ok := r.agents[name]
	return meta, ok
}

// All returns all loaded agents.
func (r *Registry) All() []AgentMeta {
	result := make([]AgentMeta, 0, len(r.agents))
	for _, meta := range r.agents {
		result = append(result, meta)
	}
	return result
}

// Names returns all agent names.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.agents))
	for name := range r.agents {
		names = append(names, name)
	}
	return names
}
