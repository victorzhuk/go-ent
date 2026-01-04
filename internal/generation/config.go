// Package generation provides hybrid code generation combining templates and AI.
package generation

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// GenerationConfig represents the configuration for project generation.
type GenerationConfig struct {
	Defaults   Defaults              `yaml:"defaults"`
	Archetypes map[string]*Archetype `yaml:"archetypes,omitempty"`
	Components []ComponentConfig     `yaml:"components,omitempty"`
}

// Defaults contains default generation settings.
type Defaults struct {
	GoVersion string `yaml:"go_version"`
	Archetype string `yaml:"archetype"`
}

// Archetype defines a project archetype with its template set.
type Archetype struct {
	Description string   `yaml:"description,omitempty"`
	Templates   []string `yaml:"templates"`
	Skip        []string `yaml:"skip,omitempty"`
}

// ComponentConfig maps a spec to a generated component.
type ComponentConfig struct {
	Name      string `yaml:"name"`
	Spec      string `yaml:"spec"`
	Archetype string `yaml:"archetype,omitempty"`
	Output    string `yaml:"output"`
}

// LoadConfig loads generation configuration from openspec/generation.yaml.
// Returns defaults if file doesn't exist.
func LoadConfig(projectRoot string) (*GenerationConfig, error) {
	configPath := filepath.Join(projectRoot, "openspec", "generation.yaml")

	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return defaultConfig(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("stat generation.yaml: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read generation.yaml: %w", err)
	}

	var cfg GenerationConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse generation.yaml: %w", err)
	}

	// Apply defaults if not set
	if cfg.Defaults.GoVersion == "" {
		cfg.Defaults.GoVersion = defaultConfig().Defaults.GoVersion
	}
	if cfg.Defaults.Archetype == "" {
		cfg.Defaults.Archetype = defaultConfig().Defaults.Archetype
	}

	return &cfg, nil
}

// defaultConfig returns the default generation configuration.
func defaultConfig() *GenerationConfig {
	return &GenerationConfig{
		Defaults: Defaults{
			GoVersion: "1.25",
			Archetype: "standard",
		},
	}
}
