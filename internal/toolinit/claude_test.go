package toolinit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeAdapter_Name(t *testing.T) {
	adapter := NewClaudeAdapter()
	assert.Equal(t, "claude", adapter.Name())
}

func TestClaudeAdapter_TargetDir(t *testing.T) {
	adapter := NewClaudeAdapter()
	assert.Equal(t, ".claude", adapter.TargetDir())
}

func TestClaudeAdapter_TransformAgent(t *testing.T) {
	adapter := NewClaudeAdapter()

	tests := []struct {
		name      string
		meta      *AgentMeta
		wantName  string
		wantDesc  string
		wantModel string
	}{
		{
			name: "architect agent with opus model",
			meta: &AgentMeta{
				Name:        "architect",
				Description: "System architect",
				Model:       "opus",
				Color:       "#4169E1",
				Skills:      []string{"go-arch", "go-api"},
				Body:        "# Architect\n\nSystem design specialist.",
			},
			wantName:  "architect",
			wantDesc:  "System architect",
			wantModel: "claude-opus-4-5-20250514",
		},
		{
			name: "planner agent with sonnet model",
			meta: &AgentMeta{
				Name:        "planner",
				Description: "Task planner",
				Model:       "sonnet",
				Body:        "# Planner\n\nBreaks down tasks.",
			},
			wantName:  "planner",
			wantDesc:  "Task planner",
			wantModel: "claude-sonnet-4-5-20250929",
		},
		{
			name: "planner-smoke agent with haiku model",
			meta: &AgentMeta{
				Name:        "planner-smoke",
				Description: "Quick triage",
				Model:       "haiku",
				Body:        "# Planner Smoke\n\nFast assessment.",
			},
			wantName:  "planner-smoke",
			wantDesc:  "Quick triage",
			wantModel: "claude-haiku-4-5-20250429",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.TransformAgent(tt.meta)
			require.NoError(t, err)
			require.NotEmpty(t, result)

			// Verify frontmatter contains expected fields
			assert.Contains(t, result, "name: "+tt.wantName)
			assert.Contains(t, result, "description: "+tt.wantDesc)
			if tt.wantModel != "" {
				assert.Contains(t, result, "model: "+tt.wantModel)
			}
			// Verify body is included
			assert.Contains(t, result, tt.meta.Body)
		})
	}
}

func TestClaudeAdapter_TransformCommand(t *testing.T) {
	adapter := NewClaudeAdapter()

	meta := &CommandMeta{
		Name:        "plan",
		Description: "Create change proposal",
		FilePath:    "plugins/go-ent/commands/plan.md",
		Body:        "# Planning Workflow\n\nComplete planning process.",
	}

	result, err := adapter.TransformCommand(meta)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Verify frontmatter
	assert.Contains(t, result, "name: plan")
	assert.Contains(t, result, "description: Create change proposal")
	// Verify body
	assert.Contains(t, result, "# Planning Workflow")
}

func TestClaudeAdapter_TransformSkill(t *testing.T) {
	adapter := NewClaudeAdapter()

	meta := &SkillMeta{
		Name:        "go-arch",
		Description: "Go architecture patterns",
		Body:        "# Go Architecture\n\nClean Architecture for Go.",
	}

	result, err := adapter.TransformSkill(meta)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Verify frontmatter
	assert.Contains(t, result, "name: go-arch")
	assert.Contains(t, result, "description: Go architecture patterns")
	assert.Contains(t, result, "version: 1.0.0")
	// Verify body
	assert.Contains(t, result, "# Go Architecture")
}
