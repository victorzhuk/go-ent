package tools

import (
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/victorzhuk/go-ent/internal/skill"
	skilldomain "github.com/victorzhuk/go-ent/internal/skill/domain"
)

func TestSkillListHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		skills        []skilldomain.Info
		input         SkillListInput
		wantSkills    int
		wantTextInMsg string
		dontWantInMsg string
	}{
		{
			name:          "empty list",
			skills:        nil,
			input:         SkillListInput{},
			wantSkills:    0,
			wantTextInMsg: "No skills found",
			dontWantInMsg: "## 1.",
		},
		{
			name: "single skill",
			skills: []skilldomain.Info{
				{Name: "go-code", Description: "Go coding patterns", Triggers: []string{"go", "golang"}},
			},
			input:         SkillListInput{},
			wantSkills:    1,
			wantTextInMsg: "go-code",
		},
		{
			name: "multiple skills",
			skills: []skilldomain.Info{
				{Name: "go-code", Description: "Go coding patterns", Triggers: []string{"go"}},
				{Name: "go-test", Description: "Go testing patterns", Triggers: []string{"test", "testing"}},
				{Name: "python-core", Description: "Python patterns", Triggers: []string{"python"}},
			},
			input:         SkillListInput{},
			wantSkills:    3,
			wantTextInMsg: "go-code",
		},
		{
			name: "filter by name",
			skills: []skilldomain.Info{
				{Name: "go-code", Description: "Go coding", Triggers: []string{"go"}},
				{Name: "python-core", Description: "Python core", Triggers: []string{"python"}},
			},
			input:         SkillListInput{Filter: "go"},
			wantSkills:    1,
			wantTextInMsg: "go-code",
			dontWantInMsg: "python-core",
		},
		{
			name: "filter no match",
			skills: []skilldomain.Info{
				{Name: "go-code", Description: "Go coding", Triggers: []string{"go"}},
			},
			input:         SkillListInput{Filter: "rust"},
			wantSkills:    0,
			wantTextInMsg: "No skills found matching filter: rust",
		},
		{
			name: "case insensitive filter",
			skills: []skilldomain.Info{
				{Name: "Go-Code", Description: "Go coding", Triggers: []string{"go"}},
			},
			input:         SkillListInput{Filter: "GO"},
			wantSkills:    1,
			wantTextInMsg: "Go-Code",
		},
		{
			name: "skill with triggers displayed",
			skills: []skilldomain.Info{
				{Name: "go-code", Description: "Go coding", Triggers: []string{"go", "golang", "testing"}},
			},
			input:         SkillListInput{},
			wantSkills:    1,
			wantTextInMsg: "**Triggers**",
		},
		{
			name: "skill without triggers",
			skills: []skilldomain.Info{
				{Name: "generic", Description: "Generic skill", Triggers: nil},
			},
			input:         SkillListInput{},
			wantSkills:    1,
			dontWantInMsg: "**Triggers**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := skill.NewRegistry()
			populateSkillRegistry(t, registry, tt.skills)

			handler := skillListHandler(registry)
			result, resp, err := handler(t.Context(), nil, tt.input)

			assert.NoError(t, err)

			if tt.wantSkills > 0 {
				assert.NotNil(t, result)
				assert.Len(t, result.Content, 1)

				textContent, ok := result.Content[0].(*mcp.TextContent)
				assert.True(t, ok)
				assert.Contains(t, textContent.Text, tt.wantTextInMsg)

				if tt.dontWantInMsg != "" {
					assert.NotContains(t, textContent.Text, tt.dontWantInMsg)
				}

				listResp, ok := resp.(SkillListResponse)
				assert.True(t, ok)
				assert.Equal(t, tt.wantSkills, listResp.Total)
				assert.Len(t, listResp.Skills, tt.wantSkills)
			} else {
				textContent, ok := result.Content[0].(*mcp.TextContent)
				assert.True(t, ok)
				assert.Contains(t, textContent.Text, tt.wantTextInMsg)

				listResp, ok := resp.(SkillListResponse)
				assert.True(t, ok)
				assert.Equal(t, 0, listResp.Total)
				assert.Empty(t, listResp.Skills)
			}
		})
	}
}

func TestSkillListHandler_ResponseStructure(t *testing.T) {
	t.Parallel()

	registry := skill.NewRegistry()
	populateSkillRegistry(t, registry, []skilldomain.Info{
		{Name: "go-code", Description: "Go coding patterns", Triggers: []string{"go", "golang"}},
		{Name: "go-test", Description: "Testing patterns", Triggers: []string{"test"}},
	})

	handler := skillListHandler(registry)
	result, resp, err := handler(t.Context(), nil, SkillListInput{})

	assert.NoError(t, err)
	assert.NotNil(t, result)

	listResp, ok := resp.(SkillListResponse)
	assert.True(t, ok)
	assert.Equal(t, 2, listResp.Total)

	assert.Len(t, listResp.Skills, 2)
	assert.Equal(t, "go-code", listResp.Skills[0].Name)
	assert.Equal(t, "Go coding patterns", listResp.Skills[0].Description)
	assert.Equal(t, []string{"go", "golang"}, listResp.Skills[0].Triggers)
}

func TestSkillListHandler_MarkdownFormatting(t *testing.T) {
	t.Parallel()

	registry := skill.NewRegistry()
	populateSkillRegistry(t, registry, []skilldomain.Info{
		{Name: "test-skill", Description: "Test description", Triggers: []string{"test"}},
	})

	handler := skillListHandler(registry)
	result, _, err := handler(t.Context(), nil, SkillListInput{})

	assert.NoError(t, err)

	textContent := result.Content[0].(*mcp.TextContent)
	text := textContent.Text

	assert.Contains(t, text, "# Available Skills")
	assert.Contains(t, text, "## 1. test-skill")
	assert.Contains(t, text, "**Description**: Test description")
	assert.Contains(t, text, "**Usage**")
}

func TestSkillListHandler_FilteredMarkdown(t *testing.T) {
	t.Parallel()

	registry := skill.NewRegistry()
	populateSkillRegistry(t, registry, []skilldomain.Info{
		{Name: "go-code", Description: "Go coding", Triggers: []string{"go"}},
	})

	handler := skillListHandler(registry)
	result, _, err := handler(t.Context(), nil, SkillListInput{Filter: "go"})

	assert.NoError(t, err)

	textContent := result.Content[0].(*mcp.TextContent)
	text := textContent.Text

	assert.Contains(t, text, "*Filtered by: go*")
}

func populateSkillRegistry(tb testing.TB, registry *skill.Registry, skills []skilldomain.Info) {
	tb.Helper()

	for _, s := range skills {
		content := "---\nname: " + s.Name + "\ndescription: " + s.Description + "\ntriggers:\n"
		for _, tr := range s.Triggers {
			content += "  - " + tr + "\n"
		}
		content += "---\n## Role\nTest role\n## Instructions\nTest instructions\n## Examples\nTest examples\n"

		tmpFile := tb.TempDir() + "/SKILL.md"
		if err := writeFile(tmpFile, content); err != nil {
			tb.Fatalf("write temp file: %v", err)
		}

		if err := registry.RegisterSkill(s.Name, tmpFile); err != nil {
			tb.Fatalf("register skill %s: %v", s.Name, err)
		}
	}
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
