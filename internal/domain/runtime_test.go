package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuntime_String(t *testing.T) {
	tests := []struct {
		name    string
		runtime Runtime
		want    string
	}{
		{"claude-code", RuntimeClaudeCode, "claude-code"},
		{"open-code", RuntimeOpenCode, "open-code"},
		{"cli", RuntimeCLI, "cli"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.runtime.String())
		})
	}
}

func TestRuntime_Valid(t *testing.T) {
	tests := []struct {
		name    string
		runtime Runtime
		want    bool
	}{
		{"valid claude-code", RuntimeClaudeCode, true},
		{"valid open-code", RuntimeOpenCode, true},
		{"valid cli", RuntimeCLI, true},
		{"invalid empty", Runtime(""), false},
		{"invalid unknown", Runtime("unknown"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.runtime.Valid())
		})
	}
}

func TestNewRuntimeCapability(t *testing.T) {
	tests := []struct {
		name    string
		runtime Runtime
		check   func(t *testing.T, rc RuntimeCapability)
	}{
		{
			name:    "claude-code capabilities",
			runtime: RuntimeClaudeCode,
			check: func(t *testing.T, rc RuntimeCapability) {
				assert.Equal(t, RuntimeClaudeCode, rc.Runtime)
				assert.True(t, rc.SupportsInteractive)
				assert.True(t, rc.SupportsFileSystem)
				assert.True(t, rc.SupportsTools)
				assert.True(t, rc.SupportsSkills)
				assert.Equal(t, 0, rc.MaxConcurrentAgents)
				assert.NotEmpty(t, rc.Description)
			},
		},
		{
			name:    "open-code capabilities",
			runtime: RuntimeOpenCode,
			check: func(t *testing.T, rc RuntimeCapability) {
				assert.Equal(t, RuntimeOpenCode, rc.Runtime)
				assert.True(t, rc.SupportsInteractive)
				assert.True(t, rc.SupportsFileSystem)
				assert.True(t, rc.SupportsTools)
				assert.True(t, rc.SupportsSkills)
				assert.Equal(t, 0, rc.MaxConcurrentAgents)
				assert.NotEmpty(t, rc.Description)
			},
		},
		{
			name:    "cli capabilities",
			runtime: RuntimeCLI,
			check: func(t *testing.T, rc RuntimeCapability) {
				assert.Equal(t, RuntimeCLI, rc.Runtime)
				assert.False(t, rc.SupportsInteractive)
				assert.True(t, rc.SupportsFileSystem)
				assert.True(t, rc.SupportsTools)
				assert.True(t, rc.SupportsSkills)
				assert.Equal(t, 1, rc.MaxConcurrentAgents)
				assert.NotEmpty(t, rc.Description)
			},
		},
		{
			name:    "unknown runtime",
			runtime: Runtime("unknown"),
			check: func(t *testing.T, rc RuntimeCapability) {
				assert.Equal(t, Runtime("unknown"), rc.Runtime)
				assert.False(t, rc.SupportsInteractive)
				assert.False(t, rc.SupportsFileSystem)
				assert.False(t, rc.SupportsTools)
				assert.False(t, rc.SupportsSkills)
				assert.Equal(t, 0, rc.MaxConcurrentAgents)
				assert.Equal(t, "Unknown runtime", rc.Description)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rc := NewRuntimeCapability(tt.runtime)
			tt.check(t, rc)
		})
	}
}

func TestRuntimeCapability_CanRunAgent(t *testing.T) {
	tests := []struct {
		name                string
		rc                  RuntimeCapability
		requiresInteractive bool
		requiresFileSystem  bool
		want                bool
	}{
		{
			name:                "claude-code supports both",
			rc:                  NewRuntimeCapability(RuntimeClaudeCode),
			requiresInteractive: true,
			requiresFileSystem:  true,
			want:                true,
		},
		{
			name:                "cli no interactive",
			rc:                  NewRuntimeCapability(RuntimeCLI),
			requiresInteractive: true,
			requiresFileSystem:  false,
			want:                false,
		},
		{
			name:                "cli supports filesystem only",
			rc:                  NewRuntimeCapability(RuntimeCLI),
			requiresInteractive: false,
			requiresFileSystem:  true,
			want:                true,
		},
		{
			name:                "cli supports neither requirement",
			rc:                  NewRuntimeCapability(RuntimeCLI),
			requiresInteractive: false,
			requiresFileSystem:  false,
			want:                true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.rc.CanRunAgent(tt.requiresInteractive, tt.requiresFileSystem)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestRuntimeCapability_HasFeature(t *testing.T) {
	rc := NewRuntimeCapability(RuntimeClaudeCode)

	tests := []struct {
		name    string
		feature string
		want    bool
	}{
		{"interactive", "interactive", true},
		{"filesystem", "filesystem", true},
		{"tools", "tools", true},
		{"skills", "skills", true},
		{"unknown", "unknown", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := rc.HasFeature(tt.feature)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestRuntimeCapability_HasFeature_CLI(t *testing.T) {
	rc := NewRuntimeCapability(RuntimeCLI)

	tests := []struct {
		name    string
		feature string
		want    bool
	}{
		{"no interactive", "interactive", false},
		{"filesystem", "filesystem", true},
		{"tools", "tools", true},
		{"skills", "skills", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := rc.HasFeature(tt.feature)
			assert.Equal(t, tt.want, result)
		})
	}
}
