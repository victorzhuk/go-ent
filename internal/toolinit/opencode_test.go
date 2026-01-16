package toolinit

//nolint:gosec // test file with necessary file operations

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenCodeAdapter_Name(t *testing.T) {
	adapter := NewOpenCodeAdapter()
	assert.Equal(t, "opencode", adapter.Name())
}

func TestOpenCodeAdapter_TargetDir(t *testing.T) {
	adapter := NewOpenCodeAdapter()
	assert.Equal(t, ".opencode", adapter.TargetDir())
}

func TestOpenCodeAdapter_TransformAgent(t *testing.T) {
	adapter := NewOpenCodeAdapter()

	tests := []struct {
		name      string
		meta      *AgentMeta
		wantDesc  string
		wantModel string
		wantMode  string
		wantTemp  float64
	}{
		{
			name: "task-smoke agent with GLM model",
			meta: &AgentMeta{
				Name:        "task-smoke",
				Description: "Execute simple tasks efficiently",
				Model:       "glm-4-flash",
				Skills:      []string{"go-code", "go-test"},
				Tools:       []string{"read", "write", "edit"},
				Body:        "# Task Smoke\n\nFast implementation.",
			},
			wantDesc:  "Execute simple tasks efficiently",
			wantModel: "zai-coding-plan/glm-4.7",
			wantMode:  "subagent",
			wantTemp:  0.0,
		},
		{
			name: "task-heavy agent with Kimi model",
			meta: &AgentMeta{
				Name:        "task-heavy",
				Description: "Execute complex tasks",
				Model:       "kimi-k2",
				Skills:      []string{"go-code", "go-arch"},
				Body:        "# Task Heavy\n\nComplex implementation.",
			},
			wantDesc:  "Execute complex tasks",
			wantModel: "kimi-for-coding/kimi-k2-thinking",
			wantMode:  "subagent",
			wantTemp:  0.0,
		},
		{
			name: "coder agent",
			meta: &AgentMeta{
				Name:        "coder",
				Description: "Implementation specialist",
				Model:       "glm-4-flash",
				Body:        "# Coder\n\nWrites code.",
			},
			wantDesc:  "Implementation specialist",
			wantModel: "zai-coding-plan/glm-4.7",
			wantMode:  "subagent",
			wantTemp:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.TransformAgent(tt.meta)
			require.NoError(t, err)
			require.NotEmpty(t, result)

			// Verify frontmatter contains expected fields
			assert.Contains(t, result, "description: "+tt.wantDesc)
			assert.Contains(t, result, "mode: "+tt.wantMode)
			if tt.wantModel != "" {
				assert.Contains(t, result, "model: "+tt.wantModel)
			}
			assert.Contains(t, result, "temperature: 0")
			// Verify body is included
			assert.Contains(t, result, tt.meta.Body)
		})
	}
}

func TestOpenCodeAdapter_TransformCommand(t *testing.T) {
	adapter := NewOpenCodeAdapter()

	meta := &CommandMeta{
		Name:        "task",
		Description: "Execute OpenSpec tasks with TDD and validation",
		FilePath:    "plugins/go-ent/commands/task.md",
		Body:        "# Task Execution\n\nExecute tasks from registry.",
	}

	result, err := adapter.TransformCommand(meta)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Verify frontmatter
	assert.Contains(t, result, "description: Execute OpenSpec tasks with TDD and validation")
	// Verify body
	assert.Contains(t, result, "# Task Execution")
}

func TestOpenCodeAdapter_TransformSkill(t *testing.T) {
	adapter := NewOpenCodeAdapter()

	meta := &SkillMeta{
		Name:        "go-code",
		Description: "Go implementation patterns",
		Body:        "# Go Code\n\nGo coding best practices.",
	}

	result, err := adapter.TransformSkill(meta)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Verify frontmatter
	assert.Contains(t, result, "name: go-code")
	assert.Contains(t, result, "description: Go implementation patterns")
	// Verify body
	assert.Contains(t, result, "# Go Code")
}

func TestOpenCodeAdapter_AgentPermissions(t *testing.T) {
	adapter := NewOpenCodeAdapter()

	meta := &AgentMeta{
		Name:        "coder",
		Description: "Implementation specialist",
		Model:       "glm-4-flash",
		Skills:      []string{"go-code", "go-test"},
		Tools:       []string{"read", "write"},
		Body:        "# Coder",
	}

	result, err := adapter.TransformAgent(meta)
	require.NoError(t, err)

	// Verify permission section exists for skills
	assert.Contains(t, result, "permission:")
	assert.Contains(t, result, "skill:")
	assert.Contains(t, result, "go-code: allow")
	assert.Contains(t, result, "go-test: allow")

	// Verify tools section
	assert.Contains(t, result, "tools:")
	assert.Contains(t, result, "read: true")
	// write should be disabled by default
	assert.Contains(t, result, "write: false")
}

func TestOpenCodeAdapter_DirectoryNames(t *testing.T) {
	// This test documents the SINGULAR directory naming for OpenCode
	adapter := NewOpenCodeAdapter()

	// Key difference from Claude: SINGULAR vs PLURAL
	assert.Equal(t, ".opencode", adapter.TargetDir())

	// OpenCode uses:
	// - command/ (not commands/)
	// - agent/ (not agents/)
	// - skill/ (not skills/)
}
