package skill

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/domain"
)

type mockSkill struct {
	name        string
	description string
	canHandle   func(ctx domain.SkillContext) bool
}

func (m *mockSkill) Name() string        { return m.name }
func (m *mockSkill) Description() string { return m.description }
func (m *mockSkill) CanHandle(ctx domain.SkillContext) bool {
	return m.canHandle(ctx)
}
func (m *mockSkill) Execute(_ context.Context, _ domain.SkillRequest) (domain.SkillResult, error) {
	return domain.SkillResult{}, nil
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r)
	assert.NotNil(t, r.parser)
	assert.Empty(t, r.skills)
	assert.Empty(t, r.runtimeSkills)
}

func TestRegistry_Register(t *testing.T) {
	tests := []struct {
		name    string
		skill   domain.Skill
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid skill",
			skill: &mockSkill{
				name:        "test-skill",
				description: "Test skill",
				canHandle:   func(ctx domain.SkillContext) bool { return true },
			},
			wantErr: false,
		},
		{
			name:    "nil skill",
			skill:   nil,
			wantErr: true,
			errMsg:  "skill cannot be nil",
		},
		{
			name: "empty name",
			skill: &mockSkill{
				name:        "",
				description: "Test",
				canHandle:   func(ctx domain.SkillContext) bool { return true },
			},
			wantErr: true,
			errMsg:  "skill name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			err := r.Register(tt.skill)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				skill, err := r.GetSkill(tt.skill.Name())
				assert.NoError(t, err)
				assert.Equal(t, tt.skill, skill)
			}
		})
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := NewRegistry()
	skill := &mockSkill{
		name:        "test-skill",
		description: "Test",
		canHandle:   func(ctx domain.SkillContext) bool { return true },
	}

	err := r.Register(skill)
	require.NoError(t, err)

	err = r.Register(skill)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegistry_Unregister(t *testing.T) {
	r := NewRegistry()
	skill := &mockSkill{
		name:        "test-skill",
		description: "Test",
		canHandle:   func(ctx domain.SkillContext) bool { return true },
	}

	err := r.Register(skill)
	require.NoError(t, err)

	err = r.Unregister("test-skill")
	assert.NoError(t, err)

	_, err = r.GetSkill("test-skill")
	assert.Error(t, err)
}

func TestRegistry_UnregisterNotFound(t *testing.T) {
	r := NewRegistry()
	err := r.Unregister("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_GetSkill(t *testing.T) {
	r := NewRegistry()
	skill := &mockSkill{
		name:        "test-skill",
		description: "Test",
		canHandle:   func(ctx domain.SkillContext) bool { return true },
	}

	err := r.Register(skill)
	require.NoError(t, err)

	retrieved, err := r.GetSkill("test-skill")
	assert.NoError(t, err)
	assert.Equal(t, skill, retrieved)

	_, err = r.GetSkill("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_Load(t *testing.T) {
	tmpDir := t.TempDir()

	skill1 := filepath.Join(tmpDir, "skill1", "SKILL.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(skill1), 0755))
	require.NoError(t, os.WriteFile(skill1, []byte(`---
name: skill1
description: "Test skill 1. Auto-activates for: test, example."
---

# Skill 1
Content here.
`), 0644))

	skill2 := filepath.Join(tmpDir, "skill2", "SKILL.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(skill2), 0755))
	require.NoError(t, os.WriteFile(skill2, []byte(`---
name: skill2
description: "Test skill 2. Auto-activates for: demo, sample."
---

# Skill 2
Content here.
`), 0644))

	r := NewRegistry()
	err := r.Load(tmpDir)
	require.NoError(t, err)

	all := r.All()
	assert.Len(t, all, 2)

	meta, err := r.Get("skill1")
	require.NoError(t, err)
	assert.Equal(t, "skill1", meta.Name)
	assert.Contains(t, meta.Description, "Test skill 1")
	assert.Contains(t, meta.Triggers, "test")
	assert.Contains(t, meta.Triggers, "example")

	meta, err = r.Get("skill2")
	require.NoError(t, err)
	assert.Equal(t, "skill2", meta.Name)
	assert.Contains(t, meta.Triggers, "demo")
	assert.Contains(t, meta.Triggers, "sample")
}

func TestRegistry_LoadNonexistentPath(t *testing.T) {
	r := NewRegistry()
	err := r.Load("/nonexistent/path")
	assert.Error(t, err)
}

func TestRegistry_MatchForContext_RuntimeSkills(t *testing.T) {
	r := NewRegistry()

	skill1 := &mockSkill{
		name:        "skill1",
		description: "Test 1",
		canHandle: func(ctx domain.SkillContext) bool {
			return ctx.Action == domain.SpecActionProposal
		},
	}

	skill2 := &mockSkill{
		name:        "skill2",
		description: "Test 2",
		canHandle: func(ctx domain.SkillContext) bool {
			return ctx.Phase == domain.ActionPhasePlanning
		},
	}

	require.NoError(t, r.Register(skill1))
	require.NoError(t, r.Register(skill2))

	tests := []struct {
		name     string
		ctx      domain.SkillContext
		expected []string
	}{
		{
			name: "matches skill1 by action",
			ctx: domain.SkillContext{
				Action: domain.SpecActionProposal,
			},
			expected: []string{"skill1"},
		},
		{
			name: "matches skill2 by phase",
			ctx: domain.SkillContext{
				Phase: domain.ActionPhasePlanning,
			},
			expected: []string{"skill2"},
		},
		{
			name: "matches both",
			ctx: domain.SkillContext{
				Action: domain.SpecActionProposal,
				Phase:  domain.ActionPhasePlanning,
			},
			expected: []string{"skill1", "skill2"},
		},
		{
			name:     "matches none",
			ctx:      domain.SkillContext{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := r.MatchForContext(tt.ctx)
			assert.ElementsMatch(t, tt.expected, matched)
		})
	}
}

func TestRegistry_MatchForContext_MetadataSkills(t *testing.T) {
	tmpDir := t.TempDir()

	skillPath := filepath.Join(tmpDir, "test-skill", "SKILL.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0755))
	require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: test-skill
description: "Test skill. Auto-activates for: architecture, design, planning."
---

# Test Skill
`), 0644))

	r := NewRegistry()
	require.NoError(t, r.Load(tmpDir))

	tests := []struct {
		name     string
		ctx      domain.SkillContext
		expected []string
	}{
		{
			name: "matches by action",
			ctx: domain.SkillContext{
				Action: domain.SpecActionProposal,
			},
			expected: []string{},
		},
		{
			name: "matches by phase planning",
			ctx: domain.SkillContext{
				Phase: domain.ActionPhasePlanning,
			},
			expected: []string{"test-skill"},
		},
		{
			name: "matches by agent role architect",
			ctx: domain.SkillContext{
				Agent: domain.AgentRoleArchitect,
			},
			expected: []string{"test-skill"},
		},
		{
			name: "matches by metadata",
			ctx: domain.SkillContext{
				Metadata: map[string]interface{}{
					"task_type": "architecture",
				},
			},
			expected: []string{"test-skill"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := r.MatchForContext(tt.ctx)
			assert.ElementsMatch(t, tt.expected, matched)
		})
	}
}

func TestRegistry_buildSearchTerms(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		name     string
		ctx      domain.SkillContext
		expected []string
	}{
		{
			name: "all fields",
			ctx: domain.SkillContext{
				Action: domain.SpecActionProposal,
				Phase:  domain.ActionPhasePlanning,
				Agent:  domain.AgentRoleArchitect,
				Metadata: map[string]interface{}{
					"task_type": "architecture",
					"priority":  "high",
				},
			},
			expected: []string{"proposal", "planning", "architect", "task_type", "architecture", "priority", "high"},
		},
		{
			name: "partial fields",
			ctx: domain.SkillContext{
				Action: domain.SpecActionImplement,
				Phase:  domain.ActionPhaseExecution,
			},
			expected: []string{"implement", "execution"},
		},
		{
			name:     "empty context",
			ctx:      domain.SkillContext{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terms := r.buildSearchTerms(tt.ctx)
			assert.ElementsMatch(t, tt.expected, terms)
		})
	}
}

func TestRegistry_matchesContext(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		name     string
		skill    SkillMeta
		terms    []string
		expected bool
	}{
		{
			name: "exact match",
			skill: SkillMeta{
				Triggers: []string{"design", "architecture"},
			},
			terms:    []string{"design"},
			expected: true,
		},
		{
			name: "partial match - term contains trigger",
			skill: SkillMeta{
				Triggers: []string{"arch"},
			},
			terms:    []string{"architecture"},
			expected: true,
		},
		{
			name: "partial match - trigger contains term",
			skill: SkillMeta{
				Triggers: []string{"architecture"},
			},
			terms:    []string{"arch"},
			expected: true,
		},
		{
			name: "no match",
			skill: SkillMeta{
				Triggers: []string{"design"},
			},
			terms:    []string{"testing"},
			expected: false,
		},
		{
			name: "empty triggers",
			skill: SkillMeta{
				Triggers: []string{},
			},
			terms:    []string{"design"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.matchesContext(tt.skill, tt.terms)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegistry_Get(t *testing.T) {
	tmpDir := t.TempDir()

	skillPath := filepath.Join(tmpDir, "test-skill", "SKILL.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0755))
	require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: test-skill
description: "Test skill"
---
`), 0644))

	r := NewRegistry()
	require.NoError(t, r.Load(tmpDir))

	meta, err := r.Get("test-skill")
	assert.NoError(t, err)
	assert.Equal(t, "test-skill", meta.Name)

	_, err = r.Get("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_All(t *testing.T) {
	tmpDir := t.TempDir()

	for i := 1; i <= 3; i++ {
		skillPath := filepath.Join(tmpDir, "skill"+string(rune('0'+i)), "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0755))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: skill`+string(rune('0'+i))+`
description: "Test"
---
`), 0644))
	}

	r := NewRegistry()
	require.NoError(t, r.Load(tmpDir))

	all := r.All()
	assert.Len(t, all, 3)
}
