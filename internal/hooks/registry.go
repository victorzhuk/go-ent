package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Registry manages hook configuration and provides access to hooks.
type Registry struct {
	config   HooksConfig
	executor *Executor
	mu       sync.RWMutex
}

// NewRegistry creates a new hook registry.
// If configPath is empty, embedded defaults will be loaded.
func NewRegistry(configPath string, executor *Executor) (*Registry, error) {
	r := &Registry{
		executor: executor,
	}

	if configPath == "" {
		if err := r.LoadFromEmbed(); err != nil {
			return nil, fmt.Errorf("load embedded hooks: %w", err)
		}
	} else {
		if err := r.LoadFromFile(configPath); err != nil {
			return nil, fmt.Errorf("load hooks from %s: %w", configPath, err)
		}
	}

	return r, nil
}

// LoadFromFile loads hook configuration from a YAML or JSON file.
func (r *Registry) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Try JSON first for .json files, otherwise try YAML then JSON
	ext := ""
	if len(path) > 5 {
		ext = path[len(path)-5:]
	}

	if ext == ".json" {
		if err := json.Unmarshal(data, &r.config); err != nil {
			return fmt.Errorf("parse JSON config: %w", err)
		}
	} else {
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &r.config); err != nil {
			if err := json.Unmarshal(data, &r.config); err != nil {
				return fmt.Errorf("parse config: %w", err)
			}
		}
	}

	return nil
}

// LoadFromEmbed loads hook configuration from embedded pkg/hooks/hooks.json.
func (r *Registry) LoadFromEmbed() error {
	// Load from embedded hooks.json
	data, err := os.ReadFile("pkg/hooks/hooks.json")
	if err != nil {
		// Fallback to empty config if embedded file not found
		r.mu.Lock()
		r.config = HooksConfig{}
		r.mu.Unlock()
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := json.Unmarshal(data, &r.config); err != nil {
		return fmt.Errorf("parse embedded hooks: %w", err)
	}

	return nil
}

// GetToolHooks returns the tool hooks configuration (thread-safe).
func (r *Registry) GetToolHooks() ToolHooks {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config.Tool
}

// GetOpenSpecHooks returns the OpenSpec hooks configuration (thread-safe).
func (r *Registry) GetOpenSpecHooks() OpenSpecHooks {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config.OpenSpec
}

// Executor returns the hook executor.
func (r *Registry) Executor() *Executor {
	return r.executor
}
