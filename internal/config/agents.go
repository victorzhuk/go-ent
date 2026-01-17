package config

import "github.com/victorzhuk/go-ent/internal/domain"

// AgentsConfig configures agent roles and their settings.
type AgentsConfig struct {
	// Default is the default agent role when not specified.
	Default domain.AgentRole `yaml:"default"`

	// Roles maps role names to their configuration.
	Roles map[string]AgentRoleConfig `yaml:"roles"`

	// Delegation configures automatic delegation behavior.
	Delegation DelegationConfig `yaml:"delegation,omitempty"`

	// ModelTier configures model selection by task complexity.
	ModelTier ModelTierConfig `yaml:"model_tier,omitempty"`
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

// BackgroundConfig configures background agent execution.
type BackgroundConfig struct {
	// MaxConcurrent is the maximum number of concurrent background agents.
	MaxConcurrent int `yaml:"max_concurrent"`

	// DefaultRole is the default agent role for background tasks.
	DefaultRole string `yaml:"default_role"`

	// DefaultModel is the default model for background tasks.
	DefaultModel string `yaml:"default_model"`

	// Timeout is the maximum execution duration per background agent in seconds.
	Timeout int `yaml:"timeout"`

	// ResourceLimits sets per-agent resource constraints.
	ResourceLimits ResourceLimits `yaml:"resource_limits"`
}

// ResourceLimits configures resource limits per background agent.
type ResourceLimits struct {
	// MaxMemoryMB is the maximum memory limit in MB per agent (0 = unlimited).
	MaxMemoryMB int `yaml:"max_memory_mb"`

	// MaxGoroutines is the maximum goroutines allowed per agent (0 = unlimited).
	MaxGoroutines int `yaml:"max_goroutines"`

	// MaxCPUPercent is the maximum CPU usage percentage per agent (0 = unlimited).
	MaxCPUPercent int `yaml:"max_cpu_percent"`
}

// ModelTierConfig configures model selection based on task complexity.
type ModelTierConfig struct {
	// Exploration is the model key for simple exploration/analysis tasks.
	Exploration string `yaml:"exploration"`

	// Complexity is the model key for complex reasoning tasks.
	Complexity string `yaml:"complexity"`

	// Critical is the model key for critical decision-making tasks.
	Critical string `yaml:"critical"`
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

	if err := a.ModelTier.Validate(); err != nil {
		return err
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

// Validate validates the background configuration.
// Skips validation if all fields are zero (not configured).
func (b *BackgroundConfig) Validate() error {
	if b.MaxConcurrent == 0 && b.Timeout == 0 && b.DefaultRole == "" && b.DefaultModel == "" {
		return nil
	}

	if b.MaxConcurrent <= 0 {
		return ErrInvalidAgentConfig
	}
	if b.Timeout <= 0 {
		return ErrInvalidAgentConfig
	}
	if b.DefaultRole == "" {
		return ErrInvalidAgentConfig
	}
	role := domain.AgentRole(b.DefaultRole)
	if !role.Valid() {
		return ErrInvalidAgentConfig
	}
	if b.DefaultModel == "" {
		return ErrInvalidAgentConfig
	}
	return b.ResourceLimits.Validate()
}

// Validate validates resource limits configuration.
func (r *ResourceLimits) Validate() error {
	if r.MaxMemoryMB < 0 {
		return ErrInvalidAgentConfig
	}
	if r.MaxGoroutines < 0 {
		return ErrInvalidAgentConfig
	}
	if r.MaxCPUPercent < 0 || r.MaxCPUPercent > 100 {
		return ErrInvalidAgentConfig
	}
	return nil
}

// Validate validates the model tier configuration.
// Skips validation if all fields are empty (not configured).
func (m *ModelTierConfig) Validate() error {
	if m.Exploration == "" && m.Complexity == "" && m.Critical == "" {
		return nil
	}
	if m.Exploration == "" {
		return ErrInvalidAgentConfig
	}
	if m.Complexity == "" {
		return ErrInvalidAgentConfig
	}
	if m.Critical == "" {
		return ErrInvalidAgentConfig
	}
	return nil
}
