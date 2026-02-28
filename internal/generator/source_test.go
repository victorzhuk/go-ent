package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertMetaToSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		meta     *AgentMetaSource
		validate func(t *testing.T, agent *AgentSource)
	}{
		{
			name: "converts basic meta",
			meta: &AgentMetaSource{
				Name:        "test-agent",
				Description: "Test agent",
				Model:       "main",
			},
			validate: func(t *testing.T, agent *AgentSource) {
				assert.Equal(t, "test-agent", agent.Name)
				assert.Equal(t, "Test agent", agent.Description)
				assert.Equal(t, "sonnet", agent.Model.Claude)
				assert.Equal(t, "main", agent.Model.OpenCode)
			},
		},
		{
			name: "converts fast model",
			meta: &AgentMetaSource{
				Name:        "fast-agent",
				Description: "Fast agent",
				Model:       "fast",
			},
			validate: func(t *testing.T, agent *AgentSource) {
				assert.Equal(t, "haiku", agent.Model.Claude)
				assert.Equal(t, "fast", agent.Model.OpenCode)
			},
		},
		{
			name: "converts heavy model",
			meta: &AgentMetaSource{
				Name:        "heavy-agent",
				Description: "Heavy agent",
				Model:       "heavy",
			},
			validate: func(t *testing.T, agent *AgentSource) {
				assert.Equal(t, "opus", agent.Model.Claude)
				assert.Equal(t, "heavy", agent.Model.OpenCode)
			},
		},
		{
			name: "converts with whenToUse",
			meta: &AgentMetaSource{
				Name:        "agent-with-when",
				Description: "Base description",
				WhenToUse:   "Use when debugging.",
				Model:       "main",
			},
			validate: func(t *testing.T, agent *AgentSource) {
				assert.Contains(t, agent.Description, "Base description")
				assert.Contains(t, agent.Description, "Use when debugging.")
			},
		},
		{
			name: "converts with skills and presets",
			meta: &AgentMetaSource{
				Name:        "full-agent",
				Description: "Full agent",
				Model:       "main",
				Skills:      []string{"go-code", "go-test"},
				ToolPresets: []string{"editing"},
			},
			validate: func(t *testing.T, agent *AgentSource) {
				assert.Equal(t, []string{"go-code", "go-test"}, agent.Skills)
				assert.Contains(t, agent.Tools.Claude.Allowed, "Read")
				assert.Contains(t, agent.Tools.Claude.Allowed, "Write")
			},
		},
		{
			name: "converts with disallowed presets",
			meta: &AgentMetaSource{
				Name:                  "restricted-agent",
				Description:           "Restricted agent",
				Model:                 "main",
				DisallowedToolPresets: []string{"serena-editing"},
			},
			validate: func(t *testing.T, agent *AgentSource) {
				assert.NotEmpty(t, agent.Tools.Claude.Disallowed)
				assert.Contains(t, agent.Tools.Claude.Disallowed, "mcp__plugin_serena_serena__replace_symbol_body")
			},
		},
		{
			name: "converts with optional fields",
			meta: &AgentMetaSource{
				Name: "optional-agent",
				ComplexityHints: map[string]string{
					"planning": "complex",
				},
				ModelMapping: map[string]string{
					"debug": "opus",
				},
				Color: "blue",
			},
			validate: func(t *testing.T, agent *AgentSource) {
				assert.Equal(t, "blue", agent.Color)
				assert.Equal(t, map[string]string{"planning": "complex"}, agent.ComplexityHints)
				assert.Equal(t, map[string]string{"debug": "opus"}, agent.ModelMapping)
			},
		},
		{
			name: "handles unknown model category",
			meta: &AgentMetaSource{
				Name:  "unknown-model",
				Model: "custom-model",
			},
			validate: func(t *testing.T, agent *AgentSource) {
				assert.Equal(t, "custom-model", agent.Model.Claude)
				assert.Equal(t, "custom-model", agent.Model.OpenCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			agent := ConvertMetaToSource(tt.meta)
			require.NotNil(t, agent)

			if tt.validate != nil {
				tt.validate(t, agent)
			}
		})
	}
}

func TestConvertToolPresetsToConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		presets           []string
		disallowedPresets []string
		validate          func(t *testing.T, cfg ToolsConfig)
	}{
		{
			name:    "empty presets",
			presets: []string{},
			validate: func(t *testing.T, cfg ToolsConfig) {
				assert.Empty(t, cfg.Claude.Allowed)
				assert.Empty(t, cfg.Claude.Disallowed)
			},
		},
		{
			name:    "editing preset",
			presets: []string{"editing"},
			validate: func(t *testing.T, cfg ToolsConfig) {
				assert.Contains(t, cfg.Claude.Allowed, "Read")
				assert.Contains(t, cfg.Claude.Allowed, "Write")
				assert.Contains(t, cfg.Claude.Allowed, "Edit")
				assert.Contains(t, cfg.Claude.Allowed, "Bash")
				assert.True(t, cfg.OpenCode["read"])
				assert.True(t, cfg.OpenCode["write"])
				assert.True(t, cfg.OpenCode["edit"])
			},
		},
		{
			name:    "readonly preset",
			presets: []string{"readonly"},
			validate: func(t *testing.T, cfg ToolsConfig) {
				assert.Contains(t, cfg.Claude.Allowed, "Read")
				assert.Contains(t, cfg.Claude.Allowed, "Bash")
				assert.NotContains(t, cfg.Claude.Allowed, "Write")
				assert.NotContains(t, cfg.Claude.Allowed, "Edit")
				assert.True(t, cfg.OpenCode["read"])
				assert.True(t, cfg.OpenCode["bash"])
			},
		},
		{
			name:              "serena-editing disallowed",
			disallowedPresets: []string{"serena-editing"},
			validate: func(t *testing.T, cfg ToolsConfig) {
				assert.Contains(t, cfg.Claude.Disallowed, "mcp__plugin_serena_serena__replace_symbol_body")
				assert.Contains(t, cfg.Claude.Disallowed, "mcp__plugin_serena_serena__insert_after_symbol")
			},
		},
		{
			name:              "combined presets",
			presets:           []string{"editing"},
			disallowedPresets: []string{"serena-editing"},
			validate: func(t *testing.T, cfg ToolsConfig) {
				assert.Contains(t, cfg.Claude.Allowed, "Read")
				assert.Contains(t, cfg.Claude.Disallowed, "mcp__plugin_serena_serena__replace_symbol_body")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := convertToolPresetsToConfig(tt.presets, tt.disallowedPresets)
			tt.validate(t, cfg)
		})
	}
}

func TestContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		slice    []string
		item     string
		expected bool
	}{
		{slice: []string{"a", "b", "c"}, item: "b", expected: true},
		{slice: []string{"a", "b", "c"}, item: "d", expected: false},
		{slice: []string{}, item: "a", expected: false},
		{slice: []string{"a"}, item: "a", expected: true},
		{slice: []string{"A", "B"}, item: "a", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.item, func(t *testing.T) {
			t.Parallel()

			result := contains(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestListAgents(t *testing.T) {
	t.Parallel()

	t.Run("lists yaml files as agents", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "agent1.yaml"), []byte("name: agent1"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "agent2.yaml"), []byte("name: agent2"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("readme"), 0o600))

		agents, err := ListAgents(tmpDir)
		require.NoError(t, err)

		assert.Len(t, agents, 2)
		assert.Contains(t, agents, "agent1")
		assert.Contains(t, agents, "agent2")
	})

	t.Run("returns empty for non-existent directory", func(t *testing.T) {
		t.Parallel()

		agents, err := ListAgents("/nonexistent/path")
		assert.NoError(t, err)
		assert.Empty(t, agents)
	})

	t.Run("skips directories", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "agent.yaml"), []byte("name: agent"), 0o600))
		require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "subdir.yaml"), 0o750))

		agents, err := ListAgents(tmpDir)
		require.NoError(t, err)

		assert.Len(t, agents, 1)
		assert.Contains(t, agents, "agent")
	})

	t.Run("returns empty for empty directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		agents, err := ListAgents(tmpDir)
		require.NoError(t, err)
		assert.Empty(t, agents)
	})
}

func TestLoadSkillSource(t *testing.T) {
	t.Parallel()

	t.Run("loads valid skill", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		skillDir := filepath.Join(tmpDir, "coding", "test-skill")
		require.NoError(t, os.MkdirAll(skillDir, 0o750))

		skillContent := `---
name: test-skill
description: Test skill
triggers:
  keywords:
    - test
    - go
---
This is the skill content.
`

		require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0o600))

		skill, err := LoadSkillSource(tmpDir, "coding", "test-skill")
		require.NoError(t, err)

		assert.Equal(t, "test-skill", skill.Name)
		assert.Equal(t, "Test skill", skill.Description)
		assert.Equal(t, []string{"test", "go"}, skill.Triggers.Keywords)
		assert.Contains(t, skill.Content, "This is the skill content.")
	})

	t.Run("fails for missing skill", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		_, err := LoadSkillSource(tmpDir, "coding", "nonexistent")
		assert.Error(t, err)
	})

	t.Run("fails for invalid frontmatter", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		skillDir := filepath.Join(tmpDir, "coding", "bad-skill")
		require.NoError(t, os.MkdirAll(skillDir, 0o750))

		skillContent := `This has no frontmatter at all.`

		require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0o600))

		_, err := LoadSkillSource(tmpDir, "coding", "bad-skill")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing frontmatter")
	})
}

func TestListSkills(t *testing.T) {
	t.Parallel()

	t.Run("lists skills by category", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		skill1Dir := filepath.Join(tmpDir, "coding", "go-code")
		require.NoError(t, os.MkdirAll(skill1Dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("---\nname: go-code\n---\ncontent"), 0o600))

		skill2Dir := filepath.Join(tmpDir, "coding", "go-test")
		require.NoError(t, os.MkdirAll(skill2Dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("---\nname: go-test\n---\ncontent"), 0o600))

		skill3Dir := filepath.Join(tmpDir, "infra", "docker")
		require.NoError(t, os.MkdirAll(skill3Dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(skill3Dir, "SKILL.md"), []byte("---\nname: docker\n---\ncontent"), 0o600))

		skills, err := ListSkills(tmpDir)
		require.NoError(t, err)

		assert.Contains(t, skills, "coding")
		assert.Contains(t, skills, "infra")
		assert.Contains(t, skills["coding"], "go-code")
		assert.Contains(t, skills["coding"], "go-test")
		assert.Contains(t, skills["infra"], "docker")
	})

	t.Run("skips directories without SKILL.md", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		skillDir := filepath.Join(tmpDir, "coding", "valid-skill")
		require.NoError(t, os.MkdirAll(skillDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: valid\n---\ncontent"), 0o600))

		emptyDir := filepath.Join(tmpDir, "coding", "empty-skill")
		require.NoError(t, os.MkdirAll(emptyDir, 0o750))

		skills, err := ListSkills(tmpDir)
		require.NoError(t, err)

		assert.Contains(t, skills["coding"], "valid-skill")
		assert.NotContains(t, skills["coding"], "empty-skill")
	})

	t.Run("returns error for non-existent directory", func(t *testing.T) {
		t.Parallel()

		_, err := ListSkills("/nonexistent/path")
		assert.Error(t, err)
	})

	t.Run("handles empty directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		skills, err := ListSkills(tmpDir)
		require.NoError(t, err)
		assert.Empty(t, skills)
	})
}

func TestMergeSkillDirs(t *testing.T) {
	t.Parallel()

	t.Run("merges skill directories", func(t *testing.T) {
		t.Parallel()

		primaryDir := t.TempDir()
		extraDir := t.TempDir()

		skill1Dir := filepath.Join(primaryDir, "coding", "skill1")
		require.NoError(t, os.MkdirAll(skill1Dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("---\nname: skill1\n---\n"), 0o600))

		skill2Dir := filepath.Join(extraDir, "coding", "skill2")
		require.NoError(t, os.MkdirAll(skill2Dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("---\nname: skill2\n---\n"), 0o600))

		skills, err := MergeSkillDirs(primaryDir, extraDir)
		require.NoError(t, err)

		assert.Contains(t, skills["coding"], "skill1")
		assert.Contains(t, skills["coding"], "skill2")
	})

	t.Run("deduplicates skills", func(t *testing.T) {
		t.Parallel()

		primaryDir := t.TempDir()
		extraDir := t.TempDir()

		skill1Dir := filepath.Join(primaryDir, "coding", "shared-skill")
		require.NoError(t, os.MkdirAll(skill1Dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("---\nname: shared\n---\n"), 0o600))

		skill2Dir := filepath.Join(extraDir, "coding", "shared-skill")
		require.NoError(t, os.MkdirAll(skill2Dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("---\nname: shared\n---\n"), 0o600))

		skills, err := MergeSkillDirs(primaryDir, extraDir)
		require.NoError(t, err)

		count := 0
		for _, name := range skills["coding"] {
			if name == "shared-skill" {
				count++
			}
		}
		assert.Equal(t, 1, count, "skill should appear only once")
	})

	t.Run("handles non-existent directories", func(t *testing.T) {
		t.Parallel()

		skills, err := MergeSkillDirs("/nonexistent/primary", "/nonexistent/extra")
		require.NoError(t, err)
		assert.Empty(t, skills)
	})
}

func TestLoadAgentMetaSource(t *testing.T) {
	t.Parallel()

	t.Run("loads valid agent meta", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		agentsMetaDir := filepath.Join(tmpDir, "agents", "meta")

		agentYAML := `name: test-agent
description: Test agent description
model: main
skills:
  - go-code
  - go-test
toolPresets:
  - editing
prompts:
  shared:
    - coding
  main: test-agent
`
		require.NoError(t, os.MkdirAll(agentsMetaDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(agentsMetaDir, "test-agent.yaml"), []byte(agentYAML), 0o600))

		// Prompts are at agents/prompts (one level up from meta)
		promptsDir := filepath.Join(tmpDir, "agents", "prompts")
		sharedDir := filepath.Join(promptsDir, "shared")
		agentsPromptsDir := filepath.Join(promptsDir, "agents")
		require.NoError(t, os.MkdirAll(sharedDir, 0o750))
		require.NoError(t, os.MkdirAll(agentsPromptsDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(sharedDir, "_coding.md"), []byte("Coding guidelines."), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(agentsPromptsDir, "test-agent.md"), []byte("Main agent prompt."), 0o600))

		agent, prompts, err := LoadAgentMetaSource(agentsMetaDir, "test-agent")
		require.NoError(t, err)

		assert.Equal(t, "test-agent", agent.Name)
		assert.Equal(t, "Test agent description", agent.Description)
		assert.Equal(t, "main", agent.Model)
		assert.Equal(t, []string{"go-code", "go-test"}, agent.Skills)
		assert.Contains(t, prompts.Shared, "coding")
		assert.Contains(t, prompts.Main, "Main agent prompt.")
	})

	t.Run("returns error for missing agent file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		_, _, err := LoadAgentMetaSource(tmpDir, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "read agent meta")
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "bad.yaml"), []byte("invalid: yaml: content:"), 0o600))

		_, _, err := LoadAgentMetaSource(tmpDir, "bad")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})
}

func TestLoadPrompts(t *testing.T) {
	t.Parallel()

	t.Run("loads shared and main prompts", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		sharedDir := filepath.Join(tmpDir, "prompts", "shared")
		agentsDir := filepath.Join(tmpDir, "prompts", "agents")
		require.NoError(t, os.MkdirAll(sharedDir, 0o750))
		require.NoError(t, os.MkdirAll(agentsDir, 0o750))

		require.NoError(t, os.WriteFile(filepath.Join(sharedDir, "_coding.md"), []byte("Coding guidelines."), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(sharedDir, "_testing.md"), []byte("Testing guidelines."), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(agentsDir, "my-agent.md"), []byte("Agent specific prompt."), 0o600))

		prompts, err := LoadPrompts(tmpDir, PromptsConfig{
			Shared: []string{"coding", "testing"},
			Main:   "my-agent",
		})
		require.NoError(t, err)

		assert.Equal(t, "Coding guidelines.", prompts.Shared["coding"])
		assert.Equal(t, "Testing guidelines.", prompts.Shared["testing"])
		assert.Equal(t, "Agent specific prompt.", prompts.Main)
	})

	t.Run("loads shared prompts without underscore prefix", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		sharedDir := filepath.Join(tmpDir, "prompts", "shared")
		require.NoError(t, os.MkdirAll(sharedDir, 0o750))

		require.NoError(t, os.WriteFile(filepath.Join(sharedDir, "_design.md"), []byte("Design guidelines."), 0o600))

		prompts, err := LoadPrompts(tmpDir, PromptsConfig{
			Shared: []string{"design"},
		})
		require.NoError(t, err)

		assert.Equal(t, "Design guidelines.", prompts.Shared["design"])
	})

	t.Run("returns error for missing shared prompt", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		_, err := LoadPrompts(tmpDir, PromptsConfig{
			Shared: []string{"nonexistent"},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "read shared prompt")
	})

	t.Run("returns error for missing main prompt", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		_, err := LoadPrompts(tmpDir, PromptsConfig{
			Main: "nonexistent",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "read main prompt")
	})

	t.Run("handles empty main prompt", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		sharedDir := filepath.Join(tmpDir, "prompts", "shared")
		require.NoError(t, os.MkdirAll(sharedDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(sharedDir, "_coding.md"), []byte("Coding."), 0o600))

		prompts, err := LoadPrompts(tmpDir, PromptsConfig{
			Shared: []string{"coding"},
			Main:   "",
		})
		require.NoError(t, err)

		assert.Equal(t, "Coding.", prompts.Shared["coding"])
		assert.Empty(t, prompts.Main)
	})

	t.Run("handles main prompt with path", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		agentsDir := filepath.Join(tmpDir, "prompts", "agents", "nested")
		require.NoError(t, os.MkdirAll(agentsDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(agentsDir, "deep.md"), []byte("Deep prompt."), 0o600))

		prompts, err := LoadPrompts(tmpDir, PromptsConfig{
			Main: "agents/nested/deep",
		})
		require.NoError(t, err)

		assert.Equal(t, "Deep prompt.", prompts.Main)
	})
}
