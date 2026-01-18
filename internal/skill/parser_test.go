package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_detectVersion(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "v2 with role tag",
			content:  "<role>test</role>",
			expected: "v2",
		},
		{
			name:     "v2 with instructions tag",
			content:  "<instructions>test</instructions>",
			expected: "v2",
		},
		{
			name:     "v2 with both tags",
			content:  "<role>test</role>\n<instructions>test</instructions>",
			expected: "v2",
		},
		{
			name:     "v1 without tags",
			content:  "Some text without tags",
			expected: "v1",
		},
		{
			name:     "v1 with auto-activates",
			content:  "Auto-activates for: testing",
			expected: "v1",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.detectVersion(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_parseFrontmatterV2(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name        string
		frontmatter string
		wantErr     bool
		errContains string
		expected    *skillMetaV2
	}{
		{
			name: "valid v2 frontmatter",
			frontmatter: `name: go-code
description: Go coding patterns
version: "1.0.0"
author: John Doe
tags:
  - go
  - code
allowedTools:
  - bash
  - write`,
			wantErr: false,
			expected: &skillMetaV2{
				Name:         "go-code",
				Description:  "Go coding patterns",
				Version:      "1.0.0",
				Author:       "John Doe",
				Tags:         []string{"go", "code"},
				AllowedTools: []string{"bash", "write"},
			},
		},
		{
			name: "v2 with optional fields",
			frontmatter: `name: go-code
description: Go coding patterns`,
			wantErr: false,
			expected: &skillMetaV2{
				Name:         "go-code",
				Description:  "Go coding patterns",
				Version:      "",
				Author:       "",
				Tags:         nil,
				AllowedTools: nil,
			},
		},
		{
			name:        "missing name",
			frontmatter: `description: Test skill`,
			wantErr:     true,
			errContains: "missing name",
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
			result, err := p.parseFrontmatterV2(tt.frontmatter)

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

func TestParser_ParseSkillFile_V1(t *testing.T) {
	p := NewParser()

	content := `---
description: 'Testing patterns with testify, testcontainers. Auto-activates for: writing tests, TDD.'
name: go-code
---
Some content here`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "skill.md")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := p.ParseSkillFile(path)

	require.NoError(t, err)
	assert.Equal(t, "go-code", result.Name)
	assert.Equal(t, "Testing patterns with testify, testcontainers. Auto-activates for: writing tests, TDD.", result.Description)
	assert.Equal(t, "v1", result.StructureVersion)
	assert.Equal(t, path, result.FilePath)
	assert.Equal(t, []string{"writing tests", "tdd"}, result.Triggers)
	assert.Equal(t, path, result.FilePath)
	assert.Equal(t, []string{"writing tests", "tdd"}, result.Triggers)
	assert.Empty(t, result.Version)
	assert.Empty(t, result.Author)
	assert.Nil(t, result.Tags)
	assert.Nil(t, result.AllowedTools)
}

func TestParser_ParseSkillFile_V2(t *testing.T) {
	p := NewParser()

	content := `---
name: go-code
description: 'Testing patterns with testify, testcontainers'
version: '1.0.0'
author: Test Author
tags:
  - go
  - test
allowedTools:
  - bash
  - write
---
<role>
You are a Go testing expert.
</role>
<instructions>
Follow Go testing best practices.
</instructions>`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "skill.md")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := p.ParseSkillFile(path)

	require.NoError(t, err)
	assert.Equal(t, "go-code", result.Name)
	assert.Equal(t, "Testing patterns with testify, testcontainers", result.Description)
	assert.Equal(t, "v2", result.StructureVersion)
	assert.Equal(t, path, result.FilePath)
	assert.Equal(t, "1.0.0", result.Version)
	assert.Equal(t, "Test Author", result.Author)
	assert.Equal(t, []string{"go", "test"}, result.Tags)
	assert.Equal(t, []string{"bash", "write"}, result.AllowedTools)
	assert.Equal(t, path, result.FilePath)
	assert.Equal(t, "1.0.0", result.Version)
	assert.Equal(t, "Test Author", result.Author)
	assert.Equal(t, []string{"go", "test"}, result.Tags)
	assert.Equal(t, []string{"bash", "write"}, result.AllowedTools)
}

func TestParser_ParseSkillFile_V2WithTriggers(t *testing.T) {
	p := NewParser()

	content := `---
name: go-code
description: 'Testing patterns with testify, testcontainers. Auto-activates for: writing tests.'
version: '1.0.0'
author: Test Author
---
<role>
Test role
</role>`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "skill.md")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := p.ParseSkillFile(path)

	require.NoError(t, err)
	assert.Equal(t, "go-code", result.Name)
	assert.Equal(t, "Testing patterns with testify, testcontainers. Auto-activates for: writing tests.", result.Description)
	assert.Equal(t, "v2", result.StructureVersion)
	assert.Equal(t, path, result.FilePath)
	assert.Equal(t, "1.0.0", result.Version)
	assert.Equal(t, "Test Author", result.Author)
	assert.Equal(t, []string{"writing tests"}, result.Triggers)
}
