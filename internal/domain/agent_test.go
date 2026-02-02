package domain

//nolint:gosec // test file with necessary file operations

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
