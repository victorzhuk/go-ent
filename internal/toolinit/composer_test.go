package toolinit

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptComposer_Compose(t *testing.T) {
	// This test requires embedded filesystem from plugin package
	// Skipping for now as we need to access pluginFS
	t.Skip("requires pluginFS from main package")
}

func TestParseAgentMetaYAML(t *testing.T) {
	yamlContent := `
name: test-agent
description: Test agent description
model: fast
color: "#123456"
skills:
  - go-code
  - go-test
tools:
  - read
  - write
dependencies:
  - planner
`

	meta, err := ParseAgentMetaYAML(yamlContent, "test.yaml")
	require.NoError(t, err)
	require.Equal(t, "test-agent", meta.Name)
	require.Equal(t, "Test agent description", meta.Description)
	require.Equal(t, "fast", meta.Model)
	require.Equal(t, "#123456", meta.Color)
	require.Equal(t, []string{"go-code", "go-test"}, meta.Skills)
	require.Equal(t, []string{"read", "write"}, meta.Tools)
	require.Equal(t, []string{"planner"}, meta.Dependencies)
}

func TestProcessIncludes(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		fs         embed.FS
		wantResult string
	}{
		{
			name:       "no includes",
			content:    "# No includes here",
			wantResult: "# No includes here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processIncludes(tt.content, tt.fs)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
