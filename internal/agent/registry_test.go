package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()

	assert.NotNil(t, reg)
	assert.NotNil(t, reg.agents)
	assert.Empty(t, reg.agents)
}

func TestLoad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T, dir string)
		wantErr     bool
		errContains string
	}{
		{
			name: "valid yaml file",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				content := `name: go-code
description: Go coding patterns
model: claude-sonnet
skills:
  - go-core
  - go-test
`
				err := os.WriteFile(filepath.Join(dir, "go-code.yaml"), []byte(content), 0o644)
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "multiple agent files",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				agent1 := `name: go-code
description: Go coding
model: claude-sonnet
`
				agent2 := `name: go-review
description: Go review
model: claude-opus
`
				err := os.WriteFile(filepath.Join(dir, "go-code.yaml"), []byte(agent1), 0o644)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(dir, "go-review.yaml"), []byte(agent2), 0o644)
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "empty directory",
			setup: func(t *testing.T, dir string) {
				t.Helper()
			},
			wantErr: false,
		},
		{
			name: "skips non-yaml files",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				content := `name: should-not-load
description: This should not be loaded
model: test
`
				err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte(content), 0o644)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{}`), 0o644)
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "skips subdirectories",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755)
				require.NoError(t, err)
				content := `name: should-not-load
description: This should not be loaded
model: test
`
				err = os.WriteFile(filepath.Join(dir, "subdir", "agent.yaml"), []byte(content), 0o644)
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "invalid yaml content",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				content := `name: [invalid
description: broken yaml
`
				err := os.WriteFile(filepath.Join(dir, "invalid.yaml"), []byte(content), 0o644)
				require.NoError(t, err)
			},
			wantErr:     true,
			errContains: "parse agent",
		},
		{
			name: "agent with all fields",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				content := `name: full-agent
description: Agent with all fields
whenToUse: When you need everything
model: claude-opus
skills:
  - skill1
  - skill2
toolPresets:
  - preset1
disallowedToolPresets:
  - forbidden1
role: senior
complexity: high
complexityHints:
  low: use-sonnet
  high: use-opus
modelMapping:
  default: claude-sonnet
dependencies:
  - dep1
color: blue
`
				err := os.WriteFile(filepath.Join(dir, "full-agent.yaml"), []byte(content), 0o644)
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "agent with hooks",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				content := `name: hooked-agent
description: Agent with hooks
model: claude-sonnet
hooks:
  PreToolUse:
    - matcher: "bash"
      hooks:
        - type: command
          command: "echo 'before bash'"
  PostToolUse:
    - matcher: ".*"
      hooks:
        - type: agent
          agent: reviewer
          prompt: "Review this"
`
				err := os.WriteFile(filepath.Join(dir, "hooked-agent.yaml"), []byte(content), 0o644)
				require.NoError(t, err)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			tt.setup(t, dir)

			reg := NewRegistry()
			err := reg.Load(dir)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestLoad_NonExistentDirectory(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	err := reg.Load("/non/existent/directory")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "read agents dir")
}

func TestGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		agents     map[string]AgentMeta
		queryName  string
		wantFound  bool
		wantName   string
		wantDesc   string
		wantModel  string
		wantSkills []string
	}{
		{
			name: "existing agent",
			agents: map[string]AgentMeta{
				"go-code": {
					Name:        "go-code",
					Description: "Go coding patterns",
					Model:       "claude-sonnet",
					Skills:      []string{"go-core", "go-test"},
				},
			},
			queryName:  "go-code",
			wantFound:  true,
			wantName:   "go-code",
			wantDesc:   "Go coding patterns",
			wantModel:  "claude-sonnet",
			wantSkills: []string{"go-core", "go-test"},
		},
		{
			name: "non-existent agent",
			agents: map[string]AgentMeta{
				"go-code": {
					Name:        "go-code",
					Description: "Go coding patterns",
					Model:       "claude-sonnet",
				},
			},
			queryName: "unknown",
			wantFound: false,
		},
		{
			name:      "empty registry",
			agents:    map[string]AgentMeta{},
			queryName: "any",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := &Registry{agents: tt.agents}
			meta, found := reg.Get(tt.queryName)

			assert.Equal(t, tt.wantFound, found)

			if tt.wantFound {
				assert.Equal(t, tt.wantName, meta.Name)
				assert.Equal(t, tt.wantDesc, meta.Description)
				assert.Equal(t, tt.wantModel, meta.Model)
				assert.Equal(t, tt.wantSkills, meta.Skills)
			}
		})
	}
}

func TestAll(t *testing.T) {
	t.Parallel()

	t.Run("multiple agents", func(t *testing.T) {
		t.Parallel()

		reg := &Registry{
			agents: map[string]AgentMeta{
				"go-code":   {Name: "go-code", Description: "Go coding"},
				"go-review": {Name: "go-review", Description: "Go review"},
				"go-test":   {Name: "go-test", Description: "Go testing"},
			},
		}

		all := reg.All()

		assert.Len(t, all, 3)

		names := make(map[string]bool)
		for _, meta := range all {
			names[meta.Name] = true
		}
		assert.True(t, names["go-code"])
		assert.True(t, names["go-review"])
		assert.True(t, names["go-test"])
	})

	t.Run("empty registry", func(t *testing.T) {
		t.Parallel()

		reg := NewRegistry()
		all := reg.All()

		assert.Empty(t, all)
		assert.NotNil(t, all)
	})

	t.Run("single agent", func(t *testing.T) {
		t.Parallel()

		reg := &Registry{
			agents: map[string]AgentMeta{
				"only-one": {Name: "only-one", Description: "Single agent"},
			},
		}

		all := reg.All()

		assert.Len(t, all, 1)
		assert.Equal(t, "only-one", all[0].Name)
	})
}

func TestNames(t *testing.T) {
	t.Parallel()

	t.Run("multiple agents", func(t *testing.T) {
		t.Parallel()

		reg := &Registry{
			agents: map[string]AgentMeta{
				"alpha": {Name: "alpha"},
				"beta":  {Name: "beta"},
				"gamma": {Name: "gamma"},
			},
		}

		names := reg.Names()

		assert.Len(t, names, 3)
		assert.Contains(t, names, "alpha")
		assert.Contains(t, names, "beta")
		assert.Contains(t, names, "gamma")
	})

	t.Run("empty registry", func(t *testing.T) {
		t.Parallel()

		reg := NewRegistry()
		names := reg.Names()

		assert.Empty(t, names)
		assert.NotNil(t, names)
	})
}

func TestRegistry_Integration(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	agent1 := `name: integrator
description: Integration test agent
model: claude-sonnet
skills:
  - skill1
  - skill2
role: developer
complexity: medium
`
	err := os.WriteFile(filepath.Join(dir, "integrator.yaml"), []byte(agent1), 0o644)
	require.NoError(t, err)

	agent2 := `name: reviewer
description: Code review agent
model: claude-opus
skills:
  - review-core
role: reviewer
`
	err = os.WriteFile(filepath.Join(dir, "reviewer.yaml"), []byte(agent2), 0o644)
	require.NoError(t, err)

	reg := NewRegistry()
	err = reg.Load(dir)
	require.NoError(t, err)

	assert.Len(t, reg.All(), 2)

	names := reg.Names()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "integrator")
	assert.Contains(t, names, "reviewer")

	integrator, found := reg.Get("integrator")
	assert.True(t, found)
	assert.Equal(t, "integrator", integrator.Name)
	assert.Equal(t, "Integration test agent", integrator.Description)
	assert.Equal(t, "claude-sonnet", integrator.Model)
	assert.Equal(t, []string{"skill1", "skill2"}, integrator.Skills)
	assert.Equal(t, "developer", integrator.Role)
	assert.Equal(t, "medium", integrator.Complexity)

	reviewer, found := reg.Get("reviewer")
	assert.True(t, found)
	assert.Equal(t, "reviewer", reviewer.Name)
	assert.Equal(t, "claude-opus", reviewer.Model)

	_, found = reg.Get("non-existent")
	assert.False(t, found)
}
