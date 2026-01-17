package config

import "github.com/victorzhuk/go-ent/internal/domain"

// MetricsConfig configures metrics collection and privacy settings.
type MetricsConfig struct {
	// Enabled enables metrics collection (default: true)
	Enabled bool `yaml:"enabled" env:"GOENT_METRICS_ENABLED"`
}

// Validate validates the metrics configuration.
func (m *MetricsConfig) Validate() error {
	return nil
}

// Config represents the complete go-ent configuration.
// Supports hierarchical loading from project-level (.go-ent/config.yaml)
// with environment variable overrides.
type Config struct {
	// Version is the config file format version (e.g., "1.0").
	Version string `yaml:"version"`

	// Agents configures agent roles, models, and skills.
	Agents AgentsConfig `yaml:"agents"`

	// Runtime configures execution environment preferences.
	Runtime RuntimeConfig `yaml:"runtime"`

	// Budget configures spending limits and tracking.
	Budget BudgetConfig `yaml:"budget"`

	// Models maps friendly names to actual model IDs.
	Models ModelsConfig `yaml:"models"`

	// Skills configures enabled skills and custom skill directories.
	Skills SkillsConfig `yaml:"skills"`

	// Background configures background agent execution.
	Background BackgroundConfig `yaml:"background,omitempty"`

	// Metrics configures metrics collection and privacy settings.
	Metrics MetricsConfig `yaml:"metrics,omitempty"`
}

// RuntimeConfig configures execution environment preferences.
type RuntimeConfig struct {
	// Preferred is the preferred runtime environment.
	Preferred domain.Runtime `yaml:"preferred"`

	// Fallback is the ordered list of fallback runtimes.
	Fallback []domain.Runtime `yaml:"fallback,omitempty"`

	// Options contains runtime-specific configuration.
	Options RuntimeOptions `yaml:"options,omitempty"`
}

// RuntimeOptions contains optional runtime-specific settings.
type RuntimeOptions struct {
	// ClaudeCodePath is the path to the Claude Code binary (if custom).
	ClaudeCodePath string `yaml:"claude_code_path,omitempty"`
}

// BudgetConfig configures spending limits and tracking.
type BudgetConfig struct {
	// Daily is the daily spending limit in USD.
	Daily float64 `yaml:"daily"`

	// Monthly is the monthly spending limit in USD.
	Monthly float64 `yaml:"monthly"`

	// PerTask is the per-task spending limit in USD.
	PerTask float64 `yaml:"per_task"`

	// Tracking enables budget tracking (default: true).
	Tracking bool `yaml:"tracking"`
}

// ModelsConfig maps friendly model names to actual model IDs.
// Example: {"opus": "claude-opus-4-5-20251101", "sonnet": "claude-sonnet-4-5-20251101"}
type ModelsConfig map[string]string

// SkillsConfig configures enabled skills and custom directories.
type SkillsConfig struct {
	// Enabled is a list of enabled skill IDs.
	Enabled []string `yaml:"enabled,omitempty"`

	// CustomDir is the path to custom skill definitions.
	CustomDir string `yaml:"custom_dir,omitempty"`
}

// Validate validates the entire configuration.
func (c *Config) Validate() error {
	if c.Version == "" {
		return ErrInvalidConfig
	}

	if err := c.Agents.Validate(); err != nil {
		return err
	}

	if err := c.Runtime.Validate(); err != nil {
		return err
	}

	if err := c.Budget.Validate(); err != nil {
		return err
	}

	if err := c.Models.Validate(); err != nil {
		return err
	}

	if err := c.Background.Validate(); err != nil {
		return err
	}

	if err := c.Metrics.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate validates the runtime configuration.
func (r *RuntimeConfig) Validate() error {
	if !r.Preferred.Valid() {
		return ErrInvalidRuntimeConfig
	}

	for _, rt := range r.Fallback {
		if !rt.Valid() {
			return ErrInvalidRuntimeConfig
		}
	}

	return nil
}

// Validate validates the budget configuration.
func (b *BudgetConfig) Validate() error {
	if b.Daily < 0 {
		return ErrInvalidBudgetConfig
	}
	if b.Monthly < 0 {
		return ErrInvalidBudgetConfig
	}
	if b.PerTask < 0 {
		return ErrInvalidBudgetConfig
	}
	return nil
}

// Validate validates the models configuration.
func (m ModelsConfig) Validate() error {
	if len(m) == 0 {
		return ErrInvalidModelConfig
	}
	for name, id := range m {
		if name == "" || id == "" {
			return ErrInvalidModelConfig
		}
	}
	return nil
}
