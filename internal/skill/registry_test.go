package skill

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func extractSkillNames(results []MatchResult) []string {
	names := make([]string, 0, len(results))
	for _, r := range results {
		if r.Skill != nil {
			names = append(names, r.Skill.Name)
		}
	}
	return names
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
	require.NoError(t, os.MkdirAll(filepath.Dir(skill1), 0750))
	require.NoError(t, os.WriteFile(skill1, []byte(`---
name: skill1
description: "Test skill 1. Auto-activates for: test, example."
---

# Skill 1
Content here.
`), 0600))

	skill2 := filepath.Join(tmpDir, "skill2", "SKILL.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(skill2), 0750))
	require.NoError(t, os.WriteFile(skill2, []byte(`---
name: skill2
description: "Test skill 2. Auto-activates for: demo, sample."
---

# Skill 2
Content here.
`), 0600))

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
	require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
	require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: test-skill
description: "Test skill. Auto-activates for: architecture, design, planning."
---

# Test Skill
`), 0600))

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
	require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
	require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: test-skill
description: "Test skill"
---
`), 0600))

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
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: skill`+string(rune('0'+i))+`
description: "Test"
---
`), 0600))
	}

	r := NewRegistry()
	require.NoError(t, r.Load(tmpDir))

	all := r.All()
	assert.Len(t, all, 3)
}

func TestRegistry_ValidateSkill(t *testing.T) {
	t.Run("valid skill", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "valid-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: valid-skill
description: "Valid skill. Auto-activates for: testing."
version: 1.0.0
tags: [test]
---

<role>
Test role
Line 2 of role
</role>

<instructions>
Test instructions
</instructions>

<examples>
<example>
<input>Test input 1</input>
<output>Test output 1</output>
</example>
<example>
<input>Test input 2</input>
<output>Test output 2</output>
</example>
</examples>
`), 0600))

		r := NewRegistry()
		err := r.Load(tmpDir)
		require.NoError(t, err)

		result, err := r.ValidateSkill("valid-skill")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Valid)
		assert.Greater(t, result.Score.Total, 0.0)
	})

	t.Run("invalid skill", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "invalid-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: invalid-skill
description: "Invalid skill"
---

<role>
`), 0600))

		r := NewRegistry()
		err := r.Load(tmpDir)
		require.NoError(t, err)

		result, err := r.ValidateSkill("invalid-skill")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Valid)
	})

	t.Run("skill not found", func(t *testing.T) {
		r := NewRegistry()
		_, err := r.ValidateSkill("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestRegistry_ValidateAll(t *testing.T) {
	t.Run("multiple valid skills", func(t *testing.T) {
		tmpDir := t.TempDir()

		for i := 1; i <= 2; i++ {
			skillPath := filepath.Join(tmpDir, "skill"+string(rune('0'+i)), "SKILL.md")
			require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
			require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: skill`+string(rune('0'+i))+`
description: "Test skill. Auto-activates for: testing."
version: 1.0.0
tags: [test]
---

<role>
Test role
Line 2 of role
</role>

<instructions>
Test instructions
</instructions>

<examples>
<example>
<input>Test input 1</input>
<output>Test output 1</output>
</example>
<example>
<input>Test input 2</input>
<output>Test output 2</output>
</example>
</examples>
`), 0600))
		}

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		result, err := r.ValidateAll()
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Valid)
		assert.Greater(t, result.Score.Total, 0.0)
	})

	t.Run("mixed valid and invalid skills", func(t *testing.T) {
		tmpDir := t.TempDir()

		validPath := filepath.Join(tmpDir, "valid-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(validPath), 0750))
		require.NoError(t, os.WriteFile(validPath, []byte(`---
name: valid-skill
description: "Valid skill. Auto-activates for: testing."
version: 1.0.0
tags: [test]
---

<role>
Test role
Line 2 of role
</role>

<instructions>
Test instructions
</instructions>

<examples>
<example>
<input>Test input 1</input>
<output>Test output 1</output>
</example>
<example>
<input>Test input 2</input>
<output>Test output 2</output>
</example>
</examples>
`), 0600))

		invalidPath := filepath.Join(tmpDir, "invalid-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(invalidPath), 0750))
		require.NoError(t, os.WriteFile(invalidPath, []byte(`---
name: invalid-skill
description: "Invalid skill"
---

<role>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		result, err := r.ValidateAll()
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.Greater(t, len(result.Issues), 0)
	})

	t.Run("no skills loaded", func(t *testing.T) {
		r := NewRegistry()

		result, err := r.ValidateAll()
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Issues)
		assert.Nil(t, result.Score)
	})
}

func TestRegistry_GetQualityReport(t *testing.T) {
	tmpDir := t.TempDir()

	skill1Path := filepath.Join(tmpDir, "skill1", "SKILL.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(skill1Path), 0750))
	require.NoError(t, os.WriteFile(skill1Path, []byte(`---
name: skill1
description: "Test skill 1. Auto-activates for: testing."
version: 1.0.0
tags: [test]
---

<role>
Test role
Line 2 of role
</role>

<instructions>
Test instructions
</instructions>

<examples>
<example>
<input>Test input 1</input>
<output>Test output 1</output>
</example>
<example>
<input>Test input 2</input>
<output>Test output 2</output>
</example>
</examples>
`), 0600))

	skill2Path := filepath.Join(tmpDir, "skill2", "SKILL.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(skill2Path), 0750))
	require.NoError(t, os.WriteFile(skill2Path, []byte(`---
name: skill2
description: "Test skill 2. Auto-activates for: testing."
version: 1.0.0
tags: [test]
---

<role>
Test role
Line 2 of role
</role>

<instructions>
Test instructions
</instructions>

<examples>
<example>
<input>Test input 1</input>
<output>Test output 1</output>
</example>
<example>
<input>Test input 2</input>
<output>Test output 2</output>
</example>
</examples>

<edge_cases>
Test edge cases
</edge_cases>
`), 0600))

	r := NewRegistry()
	require.NoError(t, r.Load(tmpDir))

	report := r.GetQualityReport()
	assert.Len(t, report, 2)

	score1, ok := report["skill1"]
	assert.True(t, ok)
	assert.Greater(t, score1, 0.0)

	score2, ok := report["skill2"]
	assert.True(t, ok)
	assert.Greater(t, score2, 0.0)
	assert.Greater(t, score2, score1)
}

func TestRegistry_Load_ComputesQualityScores(t *testing.T) {
	tmpDir := t.TempDir()

	skillPath := filepath.Join(tmpDir, "test-skill", "SKILL.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
	require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: test-skill
description: "Test skill. Auto-activates for: testing."
version: 1.0.0
tags: [test]
---

<role>
Test role
Line 2 of role
</role>

<instructions>
Test instructions
</instructions>

<examples>
<example>
<input>Test input 1</input>
<output>Test output 1</output>
</example>
<example>
<input>Test input 2</input>
<output>Test output 2</output>
</example>
</examples>
`), 0600))

	r := NewRegistry()
	require.NoError(t, r.Load(tmpDir))

	all := r.All()
	assert.Len(t, all, 1)

	meta, err := r.Get("test-skill")
	require.NoError(t, err)
	assert.Greater(t, meta.QualityScore.Total, 0.0)
}

func TestRegistry_FindMatchingSkills_WithContext_Boosting(t *testing.T) {
	t.Run("file type boosting", func(t *testing.T) {
		tmpDir := t.TempDir()

		skill1Path := filepath.Join(tmpDir, "go-code", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skill1Path), 0750))
		require.NoError(t, os.WriteFile(skill1Path, []byte(`---
name: go-code
description: "Go code patterns"
triggers:
  - patterns:
      - "go"
    file_patterns:
      - "*.go"
    weight: 0.7
---

<role>
Go code skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		skill2Path := filepath.Join(tmpDir, "py-code", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skill2Path), 0750))
		require.NoError(t, os.WriteFile(skill2Path, []byte(`---
name: py-code
description: "Python code patterns"
triggers:
  - patterns:
      - "python"
    file_patterns:
      - "*.py"
    weight: 0.7
---

<role>
Python code skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:     "code",
			FileTypes: []string{".go"},
		}

		matched := r.FindMatchingSkills("code", ctx)
		assert.Len(t, matched, 2)
		assert.Equal(t, "go-code", matched[0].Skill.Name, "go-code should be ranked higher due to file type boost")
	})

	t.Run("task type boosting", func(t *testing.T) {
		tmpDir := t.TempDir()

		testSkillPath := filepath.Join(tmpDir, "go-test", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(testSkillPath), 0750))
		require.NoError(t, os.WriteFile(testSkillPath, []byte(`---
name: go-test
description: "Testing patterns with testify"
triggers:
  - patterns:
      - "test"
    weight: 0.7
---

<role>
Go test skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		debugSkillPath := filepath.Join(tmpDir, "go-debug", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(debugSkillPath), 0750))
		require.NoError(t, os.WriteFile(debugSkillPath, []byte(`---
name: go-debug
description: "Debugging methodology"
triggers:
  - patterns:
      - "debug"
    weight: 0.7
---

<role>
Go debug skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:    "write tests",
			TaskType: "test",
		}

		matched := r.FindMatchingSkills("go", ctx)
		assert.Len(t, matched, 2)
		assert.Equal(t, "go-test", matched[0].Skill.Name, "go-test should be ranked higher due to task type boost")
	})

	t.Run("task type from query extraction", func(t *testing.T) {
		tmpDir := t.TempDir()

		testSkillPath := filepath.Join(tmpDir, "go-test", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(testSkillPath), 0750))
		require.NoError(t, os.WriteFile(testSkillPath, []byte(`---
name: go-test
description: "Testing patterns"
---

# Go Test
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query: "implement tests for new feature",
		}

		matched := r.FindMatchingSkills("go", ctx)
		names := extractSkillNames(matched)
		assert.Contains(t, names, "go-test", "go-test should match based on 'tests' keyword in query")
	})

	t.Run("affinity boosting", func(t *testing.T) {
		tmpDir := t.TempDir()

		skill1Path := filepath.Join(tmpDir, "skill1", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skill1Path), 0750))
		require.NoError(t, os.WriteFile(skill1Path, []byte(`---
name: skill1
description: "First skill"
---

<role>
Skill 1
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		skill2Path := filepath.Join(tmpDir, "skill2", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skill2Path), 0750))
		require.NoError(t, os.WriteFile(skill2Path, []byte(`---
name: skill2
description: "Second skill"
---

<role>
Skill 2
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:        "skill",
			ActiveSkills: []string{"skill2"},
		}

		matched := r.FindMatchingSkills("skill", ctx)
		assert.Len(t, matched, 2)
		assert.Equal(t, "skill2", matched[0].Skill.Name, "skill2 should be ranked higher due to affinity boost")
	})

	t.Run("combined boosts", func(t *testing.T) {
		tmpDir := t.TempDir()

		skill1Path := filepath.Join(tmpDir, "go-test", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skill1Path), 0750))
		require.NoError(t, os.WriteFile(skill1Path, []byte(`---
name: go-test
description: "Testing patterns"
triggers:
  - patterns:
      - "test"
    file_patterns:
      - "*.go"
    weight: 0.7
---

<role>
Go test skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		skill2Path := filepath.Join(tmpDir, "py-test", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skill2Path), 0750))
		require.NoError(t, os.WriteFile(skill2Path, []byte(`---
name: py-test
description: "Testing patterns"
triggers:
  - patterns:
      - "test"
    file_patterns:
      - "*.py"
    weight: 0.7
---

<role>
Python test skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:        "test",
			TaskType:     "test",
			FileTypes:    []string{".go"},
			ActiveSkills: []string{"go-test"},
		}

		matched := r.FindMatchingSkills("test", ctx)
		assert.Len(t, matched, 2)
		assert.Equal(t, "go-test", matched[0].Skill.Name, "go-test should be ranked highest with multiple boosts")
	})

	t.Run("no context - backward compatible", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "test-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: test-skill
description: "Test skill"
---

# Test Skill
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		matched := r.FindMatchingSkills("test")
		names := extractSkillNames(matched)
		assert.Contains(t, names, "test-skill")
	})
}

func TestRegistry_matchesFilePattern(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		name     string
		pattern  string
		fileType string
		expected bool
	}{
		{"exact match", ".go", ".go", true},
		{"wildcard match", "*.go", ".go", true},
		{"wildcard match with extension", "*.go", "main.go", true},
		{"wildcard mismatch", "*.go", ".py", false},
		{"exact mismatch", ".go", ".py", false},
		{"wildcard with file containing extension", "*.go", "file.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.matchesFilePattern(tt.pattern, tt.fileType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegistry_extractTaskType(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{"implement keyword", "implement new feature", "implement"},
		{"review keyword", "review the code", "review"},
		{"debug keyword", "debug the issue", "debug"},
		{"test keyword", "write tests", "test"},
		{"refactor keyword", "refactor old code", "refactor"},
		{"no keyword", "create something", ""},
		{"multiple keywords", "implement and test feature", "implement"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.extractTaskType(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegistry_applyContextBoosts(t *testing.T) {
	t.Run("runtime skill gets affinity boost", func(t *testing.T) {
		r := NewRegistry()
		skill := &mockSkill{
			name:        "test-skill",
			description: "Test",
			canHandle:   func(ctx domain.SkillContext) bool { return true },
		}
		require.NoError(t, r.Register(skill))

		ctx := &MatchContext{
			Query:        "test",
			ActiveSkills: []string{"test-skill"},
		}

		boost := r.applyContextBoosts("test-skill", ctx)
		assert.Equal(t, 0.1, boost)
	})

	t.Run("no boosts when context is empty", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "other-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: other-skill
description: "Generic skill"
---

# Other Skill
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query: "generic",
		}

		boost := r.applyContextBoosts("other-skill", ctx)
		assert.Equal(t, 0.0, boost)
	})

	t.Run("file type boost exact amount", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "go-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: go-skill
description: "Go patterns"
triggers:
  - patterns:
      - "go"
    file_patterns:
      - "*.go"
    weight: 0.7
---

<role>
Go code skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:     "go",
			FileTypes: []string{".go"},
		}

		boost := r.applyContextBoosts("go-skill", ctx)
		assert.Equal(t, 0.2, boost)
	})

	t.Run("task type boost exact amount", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "test-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: test-skill
description: "Testing patterns"
triggers:
  - patterns:
      - "test"
    weight: 0.7
---

# Test Skill
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:    "code",
			TaskType: "test",
		}

		boost := r.applyContextBoosts("test-skill", ctx)
		assert.Equal(t, 0.15, boost)
	})

	t.Run("combined boosts sum correctly", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "go-test", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: go-test
description: "Go testing patterns"
triggers:
  - patterns:
      - "test"
    file_patterns:
      - "*.go"
    weight: 0.7
---

<role>
Go test skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:        "test",
			TaskType:     "test",
			FileTypes:    []string{".go"},
			ActiveSkills: []string{"go-test"},
		}

		boost := r.applyContextBoosts("go-test", ctx)
		assert.InDelta(t, 0.45, boost, 0.001)
	})
}

func TestRegistry_FindMatchingSkills_WildcardPatterns(t *testing.T) {
	t.Run("wildcard matches multiple file types", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "code-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: code-skill
description: "Code patterns"
triggers:
  - patterns:
      - "code"
    file_patterns:
      - "**/*.go"
    weight: 0.7
  - patterns:
      - "code"
    file_patterns:
      - "**/*.md"
    weight: 0.7
---

<role>
Code skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:     "code",
			FileTypes: []string{".go", ".md"},
		}

		matched := r.FindMatchingSkills("code", ctx)
		assert.Len(t, matched, 1)
		assert.Equal(t, "code-skill", matched[0].Skill.Name)
	})

	t.Run("wildcard pattern with directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "docs-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: docs-skill
description: "Documentation patterns"
triggers:
  - patterns:
      - "doc"
    file_patterns:
      - "**/*.md"
    weight: 0.7
---

<role>
Docs skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:     "doc",
			FileTypes: []string{"docs/guide.md"},
		}

		matched := r.FindMatchingSkills("doc", ctx)
		names := extractSkillNames(matched)
		assert.Contains(t, names, "docs-skill")
	})
}

func TestRegistry_FindMatchingSkills_BoostedScoreOrdering(t *testing.T) {
	t.Run("results sorted by boosted score", func(t *testing.T) {
		tmpDir := t.TempDir()

		skill1Path := filepath.Join(tmpDir, "go-code", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skill1Path), 0750))
		require.NoError(t, os.WriteFile(skill1Path, []byte(`---
name: go-code
description: "Go code patterns"
triggers:
  - patterns:
      - "code"
    file_patterns:
      - "*.go"
    weight: 0.7
---

<role>
Go code skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		skill2Path := filepath.Join(tmpDir, "py-code", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skill2Path), 0750))
		require.NoError(t, os.WriteFile(skill2Path, []byte(`---
name: py-code
description: "Python code patterns"
triggers:
  - patterns:
      - "code"
    file_patterns:
      - "*.py"
    weight: 0.7
---

<role>
Python code skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		skill3Path := filepath.Join(tmpDir, "generic-code", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skill3Path), 0750))
		require.NoError(t, os.WriteFile(skill3Path, []byte(`---
name: generic-code
description: "Generic code patterns"
triggers:
  - patterns:
      - "code"
    weight: 0.7
---

<role>
Generic code skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:     "code",
			FileTypes: []string{".go"},
		}

		matched := r.FindMatchingSkills("code", ctx)
		assert.Len(t, matched, 3)
		assert.Equal(t, "go-code", matched[0].Skill.Name, "go-code should be first with file boost")
		assert.NotEqual(t, "go-code", matched[1].Skill.Name)
		assert.NotEqual(t, "go-code", matched[2].Skill.Name)
	})

	t.Run("multiple boosts rank correctly", func(t *testing.T) {
		tmpDir := t.TempDir()

		skill1Path := filepath.Join(tmpDir, "go-test", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skill1Path), 0750))
		require.NoError(t, os.WriteFile(skill1Path, []byte(`---
name: go-test
description: "Go testing"
triggers:
  - patterns:
      - "test"
    file_patterns:
      - "*.go"
    weight: 0.7
---

<role>
Go test skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		skill2Path := filepath.Join(tmpDir, "generic-test", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skill2Path), 0750))
		require.NoError(t, os.WriteFile(skill2Path, []byte(`---
name: generic-test
description: "Generic testing"
triggers:
  - patterns:
      - "test"
    weight: 0.7
---

<role>
Generic test skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:        "test",
			TaskType:     "test",
			FileTypes:    []string{".go"},
			ActiveSkills: []string{"go-test"},
		}

		matched := r.FindMatchingSkills("test", ctx)
		assert.Len(t, matched, 2)
		assert.Equal(t, "go-test", matched[0].Skill.Name, "go-test should be first with all boosts")
		assert.Equal(t, "generic-test", matched[1].Skill.Name)
	})
}

func TestRegistry_FindMatchingSkills_BackwardCompatibility(t *testing.T) {
	t.Run("query only - no context", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "test-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: test-skill
description: "Test skill"
---

<role>
Test skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		matched := r.FindMatchingSkills("test")
		names := extractSkillNames(matched)
		assert.Contains(t, names, "test-skill")
	})

	t.Run("nil context", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "test-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: test-skill
description: "Test skill"
---

<role>
Test skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		matched := r.FindMatchingSkills("test", nil)
		names := extractSkillNames(matched)
		assert.Contains(t, names, "test-skill")
	})

	t.Run("empty MatchContext", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "test-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: test-skill
description: "Test skill"
---

<role>
Test skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{}
		matched := r.FindMatchingSkills("test", ctx)
		names := extractSkillNames(matched)
		assert.Contains(t, names, "test-skill")
	})

	t.Run("query-only and nil context produce same results", func(t *testing.T) {
		tmpDir := t.TempDir()

		skill1Path := filepath.Join(tmpDir, "skill1", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skill1Path), 0750))
		require.NoError(t, os.WriteFile(skill1Path, []byte(`---
name: skill1
description: "Skill 1"
---

<role>
Skill 1
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		skill2Path := filepath.Join(tmpDir, "skill2", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skill2Path), 0750))
		require.NoError(t, os.WriteFile(skill2Path, []byte(`---
name: skill2
description: "Skill 2"
---

<role>
Skill 2
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		matchedQueryOnly := r.FindMatchingSkills("skill")
		matchedWithNilContext := r.FindMatchingSkills("skill", nil)

		namesQueryOnly := extractSkillNames(matchedQueryOnly)
		namesWithNilContext := extractSkillNames(matchedWithNilContext)
		assert.ElementsMatch(t, namesQueryOnly, namesWithNilContext)
	})

	t.Run("no errors with empty context fields", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "test-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: test-skill
description: "Test skill"
triggers:
  - patterns:
      - "test"
    file_patterns:
      - "*.go"
    weight: 0.7
---

<role>
Test skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:        "",
			FileTypes:    []string{},
			TaskType:     "",
			ActiveSkills: []string{},
		}

		matched := r.FindMatchingSkills("test", ctx)
		names := extractSkillNames(matched)
		assert.Contains(t, names, "test-skill")
	})

	t.Run("empty file types list handled gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "test-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: test-skill
description: "Test skill"
triggers:
  - patterns:
      - "test"
    file_patterns:
      - "*.go"
    weight: 0.7
---

<role>
Test skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:     "test",
			FileTypes: []string{},
		}

		matched := r.FindMatchingSkills("test", ctx)
		names := extractSkillNames(matched)
		assert.Contains(t, names, "test-skill")
	})

	t.Run("empty task type handled gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "test-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: test-skill
description: "Test skill"
triggers:
  - patterns:
      - "test"
    weight: 0.7
---

<role>
Test skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:    "test",
			TaskType: "",
		}

		matched := r.FindMatchingSkills("test", ctx)
		names := extractSkillNames(matched)
		assert.Contains(t, names, "test-skill")
	})

	t.Run("empty active skills handled gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()

		skillPath := filepath.Join(tmpDir, "test-skill", "SKILL.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0750))
		require.NoError(t, os.WriteFile(skillPath, []byte(`---
name: test-skill
description: "Test skill"
---

<role>
Test skill
</role>

<instructions>
Test instructions
</instructions>
`), 0600))

		r := NewRegistry()
		require.NoError(t, r.Load(tmpDir))

		ctx := &MatchContext{
			Query:        "test",
			ActiveSkills: []string{},
		}

		matched := r.FindMatchingSkills("test", ctx)
		names := extractSkillNames(matched)
		assert.Contains(t, names, "test-skill")
	})
}

func Test_matchesPattern_CacheBehavior(t *testing.T) {
	t.Run("caches compiled patterns", func(t *testing.T) {
		pattern := "test.*pattern"

		firstMatch := matchesPattern("test123pattern", pattern)
		secondMatch := matchesPattern("test456pattern", pattern)

		assert.True(t, firstMatch)
		assert.True(t, secondMatch)

		cacheMutex.RLock()
		_, exists := patternCache[strings.ToLower(pattern)]
		cacheMutex.RUnlock()

		assert.True(t, exists, "pattern should be cached after first use")
	})

	t.Run("reuses cached pattern across multiple queries", func(t *testing.T) {
		pattern := "go.*code"
		queries := []string{"go code", "go123code", "go-xyz-code"}

		for _, query := range queries {
			matched := matchesPattern(query, pattern)
			assert.True(t, matched, "query '%s' should match pattern", query)
		}

		cacheMutex.RLock()
		cachedPattern, exists := patternCache[strings.ToLower(pattern)]
		cacheMutex.RUnlock()

		assert.True(t, exists, "pattern should be in cache")
		assert.NotNil(t, cachedPattern, "cached pattern should not be nil")
	})

	t.Run("handles invalid patterns gracefully", func(t *testing.T) {
		invalidPattern := "[invalid(regex"

		matched := matchesPattern("test", invalidPattern)
		assert.False(t, matched, "invalid pattern should not match")

		cacheMutex.RLock()
		_, exists := patternCache[strings.ToLower(invalidPattern)]
		cacheMutex.RUnlock()

		assert.False(t, exists, "invalid pattern should not be cached")
	})

	t.Run("case insensitive pattern caching", func(t *testing.T) {
		pattern1 := "Test.*Pattern"
		pattern2 := "test.*pattern"

		matchesPattern("test123pattern", pattern1)
		matchesPattern("test456pattern", pattern2)

		cacheMutex.RLock()
		count := 0
		for key := range patternCache {
			if strings.Contains(key, "test.*pattern") {
				count++
			}
		}
		cacheMutex.RUnlock()

		assert.Equal(t, 1, count, "should have only one cached entry for case-insensitive patterns")
	})
}

func Test_matchesPattern_ThreadSafety(t *testing.T) {
	t.Run("concurrent reads do not block", func(t *testing.T) {
		pattern := "test.*pattern"

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				matchesPattern("test123pattern", pattern)
			}()
		}

		wg.Wait()

		cacheMutex.RLock()
		_, exists := patternCache[strings.ToLower(pattern)]
		cacheMutex.RUnlock()

		assert.True(t, exists)
	})

	t.Run("concurrent writes handle duplicate compilation", func(t *testing.T) {
		pattern := "concurrent.*test"

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				matchesPattern("concurrent123test", pattern)
			}()
		}

		wg.Wait()

		cacheMutex.RLock()
		count := 0
		for key := range patternCache {
			if strings.Contains(key, "concurrent.*test") {
				count++
			}
		}
		cacheMutex.RUnlock()

		assert.Equal(t, 1, count, "should have only one cached entry despite concurrent writes")
	})
}

func TestRegistry_resolveLoadOrder(t *testing.T) {
	r := NewRegistry()

	t.Run("no dependencies - maintains order", func(t *testing.T) {
		skills := []SkillMeta{
			{Name: "skill-a", DependsOn: nil},
			{Name: "skill-b", DependsOn: nil},
			{Name: "skill-c", DependsOn: nil},
		}

		result, err := r.resolveLoadOrder(skills)
		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, "skill-a", result[0].Name)
		assert.Equal(t, "skill-b", result[1].Name)
		assert.Equal(t, "skill-c", result[2].Name)
	})

	t.Run("empty dependencies - treated as no deps", func(t *testing.T) {
		skills := []SkillMeta{
			{Name: "skill-a", DependsOn: []string{}},
			{Name: "skill-b", DependsOn: []string{}},
		}

		result, err := r.resolveLoadOrder(skills)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "skill-a", result[0].Name)
		assert.Equal(t, "skill-b", result[1].Name)
	})

	t.Run("empty skills list", func(t *testing.T) {
		skills := []SkillMeta{}

		result, err := r.resolveLoadOrder(skills)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("simple linear dependency chain", func(t *testing.T) {
		skills := []SkillMeta{
			{Name: "skill-c", DependsOn: []string{"skill-b"}},
			{Name: "skill-a", DependsOn: nil},
			{Name: "skill-b", DependsOn: []string{"skill-a"}},
		}

		result, err := r.resolveLoadOrder(skills)
		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, "skill-a", result[0].Name)
		assert.Equal(t, "skill-b", result[1].Name)
		assert.Equal(t, "skill-c", result[2].Name)
	})

	t.Run("complex dependencies with multiple dependents", func(t *testing.T) {
		skills := []SkillMeta{
			{Name: "skill-d", DependsOn: []string{"skill-b", "skill-c"}},
			{Name: "skill-a", DependsOn: nil},
			{Name: "skill-c", DependsOn: []string{"skill-a"}},
			{Name: "skill-b", DependsOn: []string{"skill-a"}},
		}

		result, err := r.resolveLoadOrder(skills)
		require.NoError(t, err)
		assert.Len(t, result, 4)
		assert.Equal(t, "skill-a", result[0].Name)
		assert.Contains(t, []string{result[1].Name, result[2].Name}, "skill-b")
		assert.Contains(t, []string{result[1].Name, result[2].Name}, "skill-c")
		assert.Equal(t, "skill-d", result[3].Name)

		assert.ElementsMatch(t, []string{"skill-a", "skill-b", "skill-c"}, []string{
			result[0].Name,
			result[1].Name,
			result[2].Name,
		})
	})

	t.Run("circular dependency - A -> B -> A", func(t *testing.T) {
		skills := []SkillMeta{
			{Name: "skill-a", DependsOn: []string{"skill-b"}},
			{Name: "skill-b", DependsOn: []string{"skill-a"}},
		}

		result, err := r.resolveLoadOrder(skills)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "circular dependency")
	})

	t.Run("circular dependency - A -> B -> C -> A", func(t *testing.T) {
		skills := []SkillMeta{
			{Name: "skill-a", DependsOn: []string{"skill-c"}},
			{Name: "skill-b", DependsOn: []string{"skill-a"}},
			{Name: "skill-c", DependsOn: []string{"skill-b"}},
		}

		result, err := r.resolveLoadOrder(skills)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "circular dependency")
	})

	t.Run("missing dependency", func(t *testing.T) {
		skills := []SkillMeta{
			{Name: "skill-a", DependsOn: nil},
			{Name: "skill-b", DependsOn: []string{"missing-skill"}},
		}

		result, err := r.resolveLoadOrder(skills)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "dependency not found")
		assert.Contains(t, err.Error(), "missing-skill")
	})

	t.Run("multiple skills depend on one skill", func(t *testing.T) {
		skills := []SkillMeta{
			{Name: "skill-base", DependsOn: nil},
			{Name: "skill-feature1", DependsOn: []string{"skill-base"}},
			{Name: "skill-feature2", DependsOn: []string{"skill-base"}},
		}

		result, err := r.resolveLoadOrder(skills)
		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, "skill-base", result[0].Name)
		assert.Contains(t, []string{result[1].Name, result[2].Name}, "skill-feature1")
		assert.Contains(t, []string{result[1].Name, result[2].Name}, "skill-feature2")
	})

	t.Run("diamond dependency pattern", func(t *testing.T) {
		skills := []SkillMeta{
			{Name: "skill-top", DependsOn: nil},
			{Name: "skill-left", DependsOn: []string{"skill-top"}},
			{Name: "skill-right", DependsOn: []string{"skill-top"}},
			{Name: "skill-bottom", DependsOn: []string{"skill-left", "skill-right"}},
		}

		result, err := r.resolveLoadOrder(skills)
		require.NoError(t, err)
		assert.Len(t, result, 4)
		assert.Equal(t, "skill-top", result[0].Name)
		assert.Contains(t, []string{result[1].Name, result[2].Name}, "skill-left")
		assert.Contains(t, []string{result[1].Name, result[2].Name}, "skill-right")
		assert.Equal(t, "skill-bottom", result[3].Name)
	})

	t.Run("stability - independent skills maintain original order", func(t *testing.T) {
		skills := []SkillMeta{
			{Name: "skill-1", DependsOn: nil},
			{Name: "skill-2", DependsOn: nil},
			{Name: "skill-3", DependsOn: nil},
			{Name: "skill-4", DependsOn: nil},
		}

		result, err := r.resolveLoadOrder(skills)
		require.NoError(t, err)
		assert.Len(t, result, 4)
		for i := 0; i < 4; i++ {
			assert.Equal(t, fmt.Sprintf("skill-%d", i+1), result[i].Name)
		}
	})
}

func TestRegistry_DelegationHints(t *testing.T) {
	t.Run("skill with delegation hints and match", func(t *testing.T) {
		r := NewRegistry()

		skill1 := SkillMeta{
			Name: "base-skill",
			DelegatesTo: map[string]string{
				"specialized-skill": "For complex cases",
			},
		}
		r.skills = []SkillMeta{skill1}

		result := scoreSkill(&skill1, "base", nil)
		assert.Greater(t, result.Score, 0.0, "Expected match")
		assert.Len(t, result.Delegations, 1, "Expected 1 delegation")
		assert.Equal(t, "specialized-skill", result.Delegations[0].ToSkill)
		assert.Equal(t, "For complex cases", result.Delegations[0].Reason)
	})

	t.Run("skill with delegation hints but no match", func(t *testing.T) {
		r := NewRegistry()

		skill := SkillMeta{
			Name: "test",
			DelegatesTo: map[string]string{
				"other": "delegate reason",
			},
		}
		r.skills = []SkillMeta{skill}

		result := scoreSkill(&skill, "unrelated query", nil)
		assert.Equal(t, 0.0, result.Score, "Expected no match")
		assert.Len(t, result.Delegations, 0, "Expected 0 delegations when no match")
	})

	t.Run("skill without delegation hints", func(t *testing.T) {
		r := NewRegistry()

		skill := SkillMeta{
			Name: "test",
		}
		r.skills = []SkillMeta{skill}

		result := scoreSkill(&skill, "test", nil)
		assert.Greater(t, result.Score, 0.0, "Expected match by name")
		assert.Len(t, result.Delegations, 0, "Expected 0 delegations for skill without hints")
	})

	t.Run("multiple delegation hints in one skill", func(t *testing.T) {
		r := NewRegistry()

		skill := SkillMeta{
			Name: "generic-skill",
			DelegatesTo: map[string]string{
				"go-code":      "For Go-specific implementations",
				"python-code":  "For Python-specific implementations",
				"architecture": "For system design tasks",
			},
		}
		r.skills = []SkillMeta{skill}

		result := scoreSkill(&skill, "generic", nil)
		assert.Greater(t, result.Score, 0.0, "Expected match")
		assert.Len(t, result.Delegations, 3, "Expected all 3 delegations")

		delegationSkills := make([]string, len(result.Delegations))
		for i, del := range result.Delegations {
			delegationSkills[i] = del.ToSkill
		}
		assert.Contains(t, delegationSkills, "go-code")
		assert.Contains(t, delegationSkills, "python-code")
		assert.Contains(t, delegationSkills, "architecture")
	})

	t.Run("delegation hints with trigger match", func(t *testing.T) {
		r := NewRegistry()

		skill := SkillMeta{
			Name: "base-skill",
			ExplicitTriggers: []Trigger{
				{
					Patterns: []string{"base.*"},
					Weight:   0.7,
				},
			},
			DelegatesTo: map[string]string{
				"advanced-skill": "For advanced use cases",
			},
		}
		r.skills = []SkillMeta{skill}

		result := scoreSkill(&skill, "base implementation", nil)
		assert.GreaterOrEqual(t, result.Score, 0.7, "Expected match with trigger weight")
		assert.Len(t, result.Delegations, 1, "Expected delegation with trigger match")
	})

	t.Run("delegation hints with description match", func(t *testing.T) {
		r := NewRegistry()

		skill := SkillMeta{
			Name:        "test-skill",
			Description: "Test skill. Auto-activates for: testing.",
			DelegatesTo: map[string]string{
				"other-skill": "Delegation reason",
			},
		}
		r.skills = []SkillMeta{skill}

		ctx := &MatchContext{
			Query: "testing",
		}

		matched := r.FindMatchingSkills("test-skill", ctx)
		require.Greater(t, len(matched), 0, "Expected at least one match")
		assert.Len(t, matched[0].Delegations, 1, "Expected delegation from description match")
		assert.Equal(t, "other-skill", matched[0].Delegations[0].ToSkill)
	})
}
