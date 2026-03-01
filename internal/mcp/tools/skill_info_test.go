package tools

import (
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/skill"
	skilldomain "github.com/victorzhuk/go-ent/internal/skill/domain"
)

func TestSkillInfoHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		skills        []skilldomain.Info
		input         SkillInfoInput
		wantErr       bool
		wantTextInMsg string
		dontWantInMsg string
	}{
		{
			name:    "empty name returns error",
			skills:  nil,
			input:   SkillInfoInput{Name: ""},
			wantErr: true,
		},
		{
			name:          "skill not found",
			skills:        nil,
			input:         SkillInfoInput{Name: "nonexistent"},
			wantErr:       false,
			wantTextInMsg: "Skill not found",
		},
		{
			name: "valid skill",
			skills: []skilldomain.Info{
				{Name: "go-code", Description: "Go coding patterns", Triggers: []string{"go"}},
			},
			input:         SkillInfoInput{Name: "go-code"},
			wantErr:       false,
			wantTextInMsg: "# Skill: go-code",
		},
		{
			name: "skill with full metadata",
			skills: []skilldomain.Info{
				{
					Name:        "go-code",
					Description: "Go coding patterns",
					Triggers:    []string{"go", "golang"},
				},
			},
			input:         SkillInfoInput{Name: "go-code"},
			wantErr:       false,
			wantTextInMsg: "## Description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := skill.NewRegistry()
			populateSkillRegistry(t, registry, tt.skills)

			handler := skillInfoHandler(registry)
			result, _, err := handler(t.Context(), nil, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			require.NotNil(t, result)
			require.Len(t, result.Content, 1)

			textContent, ok := result.Content[0].(*mcp.TextContent)
			require.True(t, ok)
			assert.Contains(t, textContent.Text, tt.wantTextInMsg)

			if tt.dontWantInMsg != "" {
				assert.NotContains(t, textContent.Text, tt.dontWantInMsg)
			}
		})
	}
}

func TestSkillInfoHandler_FileContent(t *testing.T) {
	t.Parallel()

	registry := skill.NewRegistry()

	content := `---
name: test-skill
description: A test skill
triggers:
  - test
---
## Role
Test role content

## Instructions
Test instructions here

## Examples
Some examples
`
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/SKILL.md"
	require.NoError(t, os.WriteFile(tmpFile, []byte(content), 0o644))

	require.NoError(t, registry.RegisterSkill("test-skill", tmpFile))

	handler := skillInfoHandler(registry)
	result, _, err := handler(t.Context(), nil, SkillInfoInput{Name: "test-skill"})

	assert.NoError(t, err)
	require.NotNil(t, result)

	textContent := result.Content[0].(*mcp.TextContent)
	text := textContent.Text

	assert.Contains(t, text, "# Skill: test-skill")
	assert.Contains(t, text, "## Description")
	assert.Contains(t, text, "A test skill")
	assert.Contains(t, text, "## Triggers")
	assert.Contains(t, text, "## File Path")
	assert.Contains(t, text, "## Content")
	assert.Contains(t, text, "Test role content")
	assert.Contains(t, text, "Test instructions here")
}

func TestSkillInfoHandler_MarkdownFormat(t *testing.T) {
	t.Parallel()

	registry := skill.NewRegistry()
	populateSkillRegistry(t, registry, []skilldomain.Info{
		{Name: "go-code", Description: "Go patterns", Triggers: []string{"go"}},
	})

	handler := skillInfoHandler(registry)
	result, _, err := handler(t.Context(), nil, SkillInfoInput{Name: "go-code"})

	assert.NoError(t, err)

	textContent := result.Content[0].(*mcp.TextContent)
	text := textContent.Text

	assert.Contains(t, text, "```markdown")
}

func TestSkillInfoHandler_ErrorMessage(t *testing.T) {
	t.Parallel()

	registry := skill.NewRegistry()

	handler := skillInfoHandler(registry)
	result, _, err := handler(t.Context(), nil, SkillInfoInput{Name: ""})

	assert.Error(t, err)
	assert.Equal(t, "name is required", err.Error())
	assert.Nil(t, result)
}
