package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentRole_String(t *testing.T) {
	tests := []struct {
		name string
		role AgentRole
		want string
	}{
		{"product", AgentRoleProduct, "product"},
		{"architect", AgentRoleArchitect, "architect"},
		{"senior", AgentRoleSenior, "senior"},
		{"developer", AgentRoleDeveloper, "developer"},
		{"reviewer", AgentRoleReviewer, "reviewer"},
		{"ops", AgentRoleOps, "ops"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.role.String())
		})
	}
}

func TestAgentRole_Valid(t *testing.T) {
	tests := []struct {
		name string
		role AgentRole
		want bool
	}{
		{"valid product", AgentRoleProduct, true},
		{"valid architect", AgentRoleArchitect, true},
		{"valid senior", AgentRoleSenior, true},
		{"valid developer", AgentRoleDeveloper, true},
		{"valid reviewer", AgentRoleReviewer, true},
		{"valid ops", AgentRoleOps, true},
		{"invalid empty", AgentRole(""), false},
		{"invalid unknown", AgentRole("unknown"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.role.Valid())
		})
	}
}

func TestAgentConfig_Valid(t *testing.T) {
	tests := []struct {
		name string
		cfg  *AgentConfig
		want bool
	}{
		{
			name: "valid config",
			cfg: &AgentConfig{
				Role:        AgentRoleDeveloper,
				Model:       "sonnet",
				Skills:      []string{"code", "test"},
				Tools:       []string{"bash", "edit"},
				BudgetLimit: 100000,
				Priority:    5,
			},
			want: true,
		},
		{
			name: "valid config zero budget",
			cfg: &AgentConfig{
				Role:        AgentRoleArchitect,
				Model:       "opus",
				BudgetLimit: 0,
			},
			want: true,
		},
		{
			name: "invalid role",
			cfg: &AgentConfig{
				Role:  AgentRole("invalid"),
				Model: "sonnet",
			},
			want: false,
		},
		{
			name: "empty model",
			cfg: &AgentConfig{
				Role:  AgentRoleDeveloper,
				Model: "",
			},
			want: false,
		},
		{
			name: "negative budget",
			cfg: &AgentConfig{
				Role:        AgentRoleDeveloper,
				Model:       "sonnet",
				BudgetLimit: -1,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.cfg.Valid())
		})
	}
}

func TestAgentConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *AgentConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &AgentConfig{
				Role:  AgentRoleDeveloper,
				Model: "sonnet",
			},
			wantErr: false,
		},
		{
			name: "invalid role",
			cfg: &AgentConfig{
				Role:  AgentRole("invalid"),
				Model: "sonnet",
			},
			wantErr: true,
		},
		{
			name: "empty model",
			cfg: &AgentConfig{
				Role:  AgentRoleDeveloper,
				Model: "",
			},
			wantErr: true,
		},
		{
			name: "negative budget",
			cfg: &AgentConfig{
				Role:        AgentRoleDeveloper,
				Model:       "sonnet",
				BudgetLimit: -1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrInvalidAgentConfig)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAgentCapability_Has(t *testing.T) {
	tests := []struct {
		name  string
		caps  AgentCapability
		check AgentCapability
		want  bool
	}{
		{
			name:  "has code generation",
			caps:  CapabilityCodeGeneration,
			check: CapabilityCodeGeneration,
			want:  true,
		},
		{
			name:  "has code review in combined",
			caps:  CapabilityCodeGeneration | CapabilityCodeReview,
			check: CapabilityCodeReview,
			want:  true,
		},
		{
			name:  "does not have testing",
			caps:  CapabilityCodeGeneration,
			check: CapabilityTesting,
			want:  false,
		},
		{
			name:  "empty has nothing",
			caps:  0,
			check: CapabilityCodeGeneration,
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.caps.Has(tt.check))
		})
	}
}

func TestAgentCapability_Add(t *testing.T) {
	tests := []struct {
		name string
		caps AgentCapability
		add  AgentCapability
		want AgentCapability
	}{
		{
			name: "add to empty",
			caps: 0,
			add:  CapabilityCodeGeneration,
			want: CapabilityCodeGeneration,
		},
		{
			name: "add to existing",
			caps: CapabilityCodeGeneration,
			add:  CapabilityCodeReview,
			want: CapabilityCodeGeneration | CapabilityCodeReview,
		},
		{
			name: "add already present",
			caps: CapabilityCodeGeneration,
			add:  CapabilityCodeGeneration,
			want: CapabilityCodeGeneration,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.caps.Add(tt.add)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestAgentCapability_Remove(t *testing.T) {
	tests := []struct {
		name   string
		caps   AgentCapability
		remove AgentCapability
		want   AgentCapability
	}{
		{
			name:   "remove from empty",
			caps:   0,
			remove: CapabilityCodeGeneration,
			want:   0,
		},
		{
			name:   "remove existing",
			caps:   CapabilityCodeGeneration | CapabilityCodeReview,
			remove: CapabilityCodeReview,
			want:   CapabilityCodeGeneration,
		},
		{
			name:   "remove not present",
			caps:   CapabilityCodeGeneration,
			remove: CapabilityCodeReview,
			want:   CapabilityCodeGeneration,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.caps.Remove(tt.remove)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestAgentCapability_String(t *testing.T) {
	tests := []struct {
		name string
		caps AgentCapability
		want string
	}{
		{
			name: "empty",
			caps: 0,
			want: "none",
		},
		{
			name: "single capability",
			caps: CapabilityCodeGeneration,
			want: "[code-generation]",
		},
		{
			name: "multiple capabilities",
			caps: CapabilityCodeGeneration | CapabilityCodeReview,
			want: "[code-generation, code-review]",
		},
		{
			name: "all capabilities",
			caps: CapabilityCodeGeneration | CapabilityCodeReview | CapabilityArchitecture |
				CapabilityTesting | CapabilityDebugging | CapabilityDocumentation |
				CapabilityRefactoring | CapabilityDeployment,
			want: "[code-generation, code-review, architecture, testing, debugging, documentation, refactoring, deployment]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.caps.String())
		})
	}
}
