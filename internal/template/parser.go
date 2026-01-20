package template

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Prompt represents a user input prompt for template placeholder values.
type Prompt struct {
	Key      string // Placeholder name (e.g., "SKILL_NAME")
	Prompt   string // Text shown to user
	Default  string // Default value
	Required bool   // Whether value must be provided
}

// TemplateConfig represents the configuration from a template's config.yaml file.
type TemplateConfig struct {
	Name        string   // Template identifier
	Category    string   // Template category (go, typescript, etc.)
	Description string   // Brief description
	Author      string   // Template author
	Version     string   // Semantic version
	Prompts     []Prompt // User prompts for placeholder values
}

// Validate checks that required config fields are present.
func (c *TemplateConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("required field 'name' is empty")
	}
	if c.Category == "" {
		return fmt.Errorf("required field 'category' is empty")
	}
	return nil
}

// ParseConfig reads and parses a template config.yaml file.
// Returns a TemplateConfig struct or an error if parsing fails.
func ParseConfig(path string) (*TemplateConfig, error) {
	data, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg TemplateConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}
