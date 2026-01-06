package config

import "github.com/victorzhuk/go-ent/internal/domain"

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
}

// AgentsConfig configures agent roles and their settings.
type AgentsConfig struct {
	// Default is the default agent role when not specified.
	Default domain.AgentRole `yaml:"default"`

	// Roles maps role names to their configuration.
	Roles map[string]AgentRoleConfig `yaml:"roles"`

	// Delegation configures automatic delegation behavior.
	Delegation DelegationConfig `yaml:"delegation,omitempty"`
}

// AgentRoleConfig configures a specific agent role.
type AgentRoleConfig struct {
	// Model is the model mapping key (e.g., "opus", "sonnet").
	Model string `yaml:"model"`

	// Skills is a list of enabled skill IDs for this role.
	Skills []string `yaml:"skills,omitempty"`

	// BudgetLimit is the per-execution budget limit in USD (0 = unlimited).
	BudgetLimit float64 `yaml:"budget_limit,omitempty"`
}

// DelegationConfig configures automatic agent delegation.
type DelegationConfig struct {
	// Auto enables automatic delegation based on task complexity.
	Auto bool `yaml:"auto"`

	// ApprovalRequired lists roles that require user approval before delegation.
	ApprovalRequired []domain.AgentRole `yaml:"approval_required,omitempty"`
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

	return nil
}

// Validate validates the agents configuration.
func (a *AgentsConfig) Validate() error {
	if !a.Default.Valid() {
		return ErrInvalidAgentConfig
	}

	if len(a.Roles) == 0 {
		return ErrInvalidAgentConfig
	}

	for roleName, roleConfig := range a.Roles {
		role := domain.AgentRole(roleName)
		if !role.Valid() {
			return ErrInvalidAgentConfig
		}
		if err := roleConfig.Validate(); err != nil {
			return err
		}
	}

	for _, role := range a.Delegation.ApprovalRequired {
		if !role.Valid() {
			return ErrInvalidAgentConfig
		}
	}

	return nil
}

// Validate validates an agent role configuration.
func (a *AgentRoleConfig) Validate() error {
	if a.Model == "" {
		return ErrInvalidAgentConfig
	}
	if a.BudgetLimit < 0 {
		return ErrInvalidAgentConfig
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
