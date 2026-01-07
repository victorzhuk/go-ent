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
