package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_detectVersion(t *testing.T) {
	t.Parallel()
	p := NewParser()

	tests := []struct {
		name        string
		content     string
		frontmatter string
		expected    string
	}{
		{
			name: "v4 with flat triggers and all sections",
			content: `## Role
Expert
## Instructions
Do this
## Examples
Example 1`,
			frontmatter: "triggers:\n  - test\n  - code",
			expected:    "v4",
		},
		{
			name:        "unknown format - missing sections",
			content:     "Some content without proper sections",
			frontmatter: "",
			expected:    "unknown",
		},
		{
			name:        "unknown format - no triggers",
			content:     "## Role\nTest\n## Instructions\nTest",
			frontmatter: "",
			expected:    "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := p.detectVersion(tt.content, tt.frontmatter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_parseFrontmatterV4(t *testing.T) {
	t.Parallel()
	p := NewParser()

	tests := []struct {
		name        string
		frontmatter string
		wantErr     bool
		errContains string
		expected    *skillMetaV4
	}{
		{
			name: "valid v4 frontmatter",
			frontmatter: `name: go-code
description: Go coding patterns
triggers:
  - go code
  - golang
  - go patterns`,
			wantErr: false,
			expected: &skillMetaV4{
				Name:        "go-code",
				Description: "Go coding patterns",
				Triggers:    []string{"go code", "golang", "go patterns"},
			},
		},
		{
			name: "minimal v4 frontmatter",
			frontmatter: `name: test-skill
description: Test description
triggers:
  - test`,
			wantErr: false,
			expected: &skillMetaV4{
				Name:        "test-skill",
				Description: "Test description",
				Triggers:    []string{"test"},
			},
		},
		{
			name: "missing name",
			frontmatter: `description: Test skill
triggers:
  - test`,
			wantErr:     true,
			errContains: "missing name",
		},
		{
			name: "missing description",
			frontmatter: `name: test-skill
triggers:
  - test`,
			wantErr:     true,
			errContains: "missing description",
		},
		{
			name:        "invalid yaml",
			frontmatter: `name: [invalid`,
			wantErr:     true,
			errContains: "parse yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := p.parseFrontmatterV4(tt.frontmatter)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParser_ParseSkillFile_V4(t *testing.T) {
	t.Parallel()
	p := NewParser()

	content := `---
name: go-code
description: Go coding patterns and best practices
triggers:
  - go code
  - golang
  - go patterns
---

## Role

Expert Go developer specializing in idiomatic code.

## Instructions

Follow Go best practices:
1. Use gofmt
2. Write tests
3. Handle errors

## Examples

<example>
<input>Write a simple function</input>
<output>func Example() {}</output>
</example>

## References

- [Go conventions](references/conventions.md)
- [Error patterns](references/errors.md)
`

	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, "skills", "go", "go-code")
	err := os.MkdirAll(skillsDir, 0o755)
	require.NoError(t, err)

	path := filepath.Join(skillsDir, "SKILL.md")
	err = os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := p.ParseSkillFile(path)

	require.NoError(t, err)
	assert.Equal(t, "go-code", result.Name)
	assert.Equal(t, "Go coding patterns and best practices", result.Description)
	assert.Equal(t, "v4", result.StructureVersion)
	assert.Equal(t, path, result.FilePath)
	assert.Equal(t, "go", result.Category)
	assert.Equal(t, []string{"go code", "golang", "go patterns"}, result.Triggers)
	assert.Len(t, result.ExplicitTriggers, 3)
	assert.Equal(t, 0.7, result.ExplicitTriggers[0].Weight)
	assert.Contains(t, result.Role, "Expert Go developer")
	assert.Contains(t, result.Instructions, "Follow Go best practices")
	assert.Contains(t, result.Examples, "<example>")
	assert.Contains(t, result.References, "references/conventions.md")
	assert.Contains(t, result.References, "references/errors.md")
}

func TestParser_ParseSkillFile_V4_UnsupportedFormat(t *testing.T) {
	t.Parallel()
	p := NewParser()

	// Old v2 format with XML tags
	content := `---
name: old-skill
description: Old format skill
---

<role>
Old role format
</role>

<instructions>
Old instructions format
</instructions>
`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "skill.md")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := p.ParseSkillFile(path)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported skill format")
	assert.Contains(t, err.Error(), "only v4 is supported")
}

func TestParser_extractMarkdownSection(t *testing.T) {
	t.Parallel()
	p := NewParser()

	content := `# Title

## Role

This is the role section.
It has multiple lines.

## Instructions

This is instructions.

## Examples

Examples go here.

## References

- [Link](path.md)
`

	tests := []struct {
		name     string
		section  string
		expected string
	}{
		{
			name:     "extract role",
			section:  "Role",
			expected: "This is the role section.\nIt has multiple lines.",
		},
		{
			name:     "extract instructions",
			section:  "Instructions",
			expected: "This is instructions.",
		},
		{
			name:     "extract examples",
			section:  "Examples",
			expected: "Examples go here.",
		},
		{
			name:     "non-existent section",
			section:  "NonExistent",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := p.extractMarkdownSection(content, tt.section)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_extractReferencesSection(t *testing.T) {
	t.Parallel()
	p := NewParser()

	content := `## References

- [Conventions](references/conventions.md)
- [Error patterns](references/errors.md)
- [Best practices](references/best-practices.md)

## Other Section
`

	result := p.extractReferencesSection(content)

	assert.Len(t, result, 3)
	assert.Contains(t, result, "references/conventions.md")
	assert.Contains(t, result, "references/errors.md")
	assert.Contains(t, result, "references/best-practices.md")
}

func TestParser_detectCategory(t *testing.T) {
	t.Parallel()
	p := NewParser()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "unix path with go category",
			path:     "/home/user/project/skills/go/go-code/SKILL.md",
			expected: "go",
		},
		{
			name:     "unix path with core category",
			path:     "/home/user/project/skills/core/api-design/SKILL.md",
			expected: "core",
		},
		{
			name:     "windows path with ent category",
			path:     "C:\\Users\\user\\project\\skills\\ent\\ent-tooling\\SKILL.md",
			expected: "ent",
		},
		{
			name:     "no skills in path",
			path:     "/home/user/project/random/SKILL.md",
			expected: "",
		},
		{
			name:     "relative path",
			path:     "skills/go/go-code/SKILL.md",
			expected: "go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := p.detectCategory(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_stringsToTriggers(t *testing.T) {
	t.Parallel()
	p := NewParser()

	strings := []string{"testing", "tdd", "go code"}
	result := p.stringsToTriggers(strings, 0.7)

	assert.Len(t, result, 3)
	assert.Equal(t, []string{"testing"}, result[0].Keywords)
	assert.Equal(t, 0.7, result[0].Weight)
	assert.Equal(t, []string{"tdd"}, result[1].Keywords)
	assert.Equal(t, 0.7, result[1].Weight)
	assert.Equal(t, []string{"go code"}, result[2].Keywords)
	assert.Equal(t, 0.7, result[2].Weight)
}

func TestParser_triggersToStrings(t *testing.T) {
	t.Parallel()
	p := NewParser()

	triggers := []Trigger{
		{
			Patterns: []string{"write.*test"},
			Keywords: []string{"testing", "tdd"},
		},
		{
			FilePatterns: []string{"**/*_test.go"},
			Weight:       0.7,
		},
	}

	result := p.triggersToStrings(triggers)

	assert.Contains(t, result, "write.*test")
	assert.Contains(t, result, "testing")
	assert.Contains(t, result, "tdd")
	assert.Contains(t, result, "**/*_test.go")
}
