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

func TestParser_ParseSkillFile_V2WithExplicitTriggers(t *testing.T) {
	p := NewParser()

	content := `---
name: go-code
description: 'Testing patterns with testify, testcontainers'
version: '1.0.0'
author: Test Author
triggers:
  - patterns:
      - "write.*test"
    keywords:
      - testing
      - tdd
    weight: 0.8
  - patterns:
      - "test.*framework"
    weight: 0.7
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
	assert.Equal(t, "Testing patterns with testify, testcontainers", result.Description)
	assert.Equal(t, "v2", result.StructureVersion)
	assert.Equal(t, path, result.FilePath)
	assert.Equal(t, "1.0.0", result.Version)
	assert.Equal(t, "Test Author", result.Author)
	assert.Len(t, result.ExplicitTriggers, 2)
	assert.Equal(t, []string{"write.*test"}, result.ExplicitTriggers[0].Patterns)
	assert.Equal(t, []string{"testing", "tdd"}, result.ExplicitTriggers[0].Keywords)
	assert.Equal(t, 0.8, result.ExplicitTriggers[0].Weight)
	assert.Equal(t, []string{"test.*framework"}, result.ExplicitTriggers[1].Patterns)
	assert.Equal(t, 0.7, result.ExplicitTriggers[1].Weight)
	assert.Contains(t, result.Triggers, "write.*test")
	assert.Contains(t, result.Triggers, "testing")
	assert.Contains(t, result.Triggers, "tdd")
	assert.Contains(t, result.Triggers, "test.*framework")
}

func TestParser_ParseSkillFile_V2FallbackTriggers(t *testing.T) {
	p := NewParser()

	content := `---
name: go-code
description: 'Testing patterns. Auto-activates for: writing tests, TDD.'
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
	assert.Equal(t, "Testing patterns. Auto-activates for: writing tests, TDD.", result.Description)
	assert.Equal(t, "v2", result.StructureVersion)
	assert.Equal(t, []string{"writing tests", "tdd"}, result.Triggers)
	assert.Len(t, result.ExplicitTriggers, 2)
	assert.Equal(t, []string{"writing tests"}, result.ExplicitTriggers[0].Keywords)
	assert.Equal(t, 0.5, result.ExplicitTriggers[0].Weight)
	assert.Equal(t, []string{"tdd"}, result.ExplicitTriggers[1].Keywords)
	assert.Equal(t, 0.5, result.ExplicitTriggers[1].Weight)
}

func TestParser_triggersToStrings(t *testing.T) {
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

func TestParser_stringsToTriggers(t *testing.T) {
	p := NewParser()

	strings := []string{"testing", "tdd", "write code"}
	result := p.stringsToTriggers(strings, 0.5)

	assert.Len(t, result, 3)
	assert.Equal(t, []string{"testing"}, result[0].Keywords)
	assert.Equal(t, 0.5, result[0].Weight)
	assert.Equal(t, []string{"tdd"}, result[1].Keywords)
	assert.Equal(t, 0.5, result[1].Weight)
	assert.Equal(t, []string{"write code"}, result[2].Keywords)
	assert.Equal(t, 0.5, result[2].Weight)
}

func TestParser_ParseSkillFile_ExplicitTriggerWeightValidation(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
	}{
		{
			name: "default weight when not specified",
			content: `---
name: test-skill
description: Test description
triggers:
  - pattern: "test"
  - pattern: "another"
---
<role>test</role>`,
			wantErr: false,
		},
		{
			name: "valid weight at lower bound 0.0",
			content: `---
name: test-skill
description: Test description
triggers:
  - pattern: "test"
    weight: 0.0
---
<role>test</role>`,
			wantErr: false,
		},
		{
			name: "valid weight at upper bound 1.0",
			content: `---
name: test-skill
description: Test description
triggers:
  - pattern: "test"
    weight: 1.0
---
<role>test</role>`,
			wantErr: false,
		},
		{
			name: "valid weight in middle range",
			content: `---
name: test-skill
description: Test description
triggers:
  - pattern: "test"
    weight: 0.5
---
<role>test</role>`,
			wantErr: false,
		},
		{
			name: "error on negative weight",
			content: `---
name: test-skill
description: Test description
triggers:
  - pattern: "test"
    weight: -0.1
---
<role>test</role>`,
			wantErr:     true,
			errContains: "weight must be between 0.0 and 1.0",
		},
		{
			name: "error on weight greater than 1.0",
			content: `---
name: test-skill
description: Test description
triggers:
  - pattern: "test"
    weight: 1.1
---
<role>test</role>`,
			wantErr:     true,
			errContains: "weight must be between 0.0 and 1.0",
		},
		{
			name: "error on large negative weight",
			content: `---
name: test-skill
description: Test description
triggers:
  - pattern: "test"
    weight: -10.0
---
<role>test</role>`,
			wantErr:     true,
			errContains: "weight must be between 0.0 and 1.0",
		},
		{
			name: "error on large weight",
			content: `---
name: test-skill
description: Test description
triggers:
  - pattern: "test"
    weight: 100.0
---
<role>test</role>`,
			wantErr:     true,
			errContains: "weight must be between 0.0 and 1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "skill.md")
			err := os.WriteFile(path, []byte(tt.content), 0o644)
			require.NoError(t, err)

			result, err := p.ParseSkillFile(path)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if len(result.ExplicitTriggers) > 0 {
					if result.ExplicitTriggers[0].Weight == 0 {
						assert.Equal(t, 0.7, result.ExplicitTriggers[0].Weight, "default weight should be 0.7")
					}
				}
			}
		})
	}
}

func TestParser_ParseSkillFile_ExplicitTriggersEdgeCases(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name    string
		content string
		verify  func(t *testing.T, meta *SkillMeta)
	}{
		{
			name: "trigger with only pattern",
			content: `---
name: test-skill
description: Test description
triggers:
  - patterns:
      - "write.*test"
---
<role>test</role>`,
			verify: func(t *testing.T, meta *SkillMeta) {
				assert.Len(t, meta.ExplicitTriggers, 1)
				assert.Equal(t, []string{"write.*test"}, meta.ExplicitTriggers[0].Patterns)
				assert.Equal(t, 0.7, meta.ExplicitTriggers[0].Weight)
				assert.Empty(t, meta.ExplicitTriggers[0].Keywords)
				assert.Empty(t, meta.ExplicitTriggers[0].FilePatterns)
			},
		},
		{
			name: "trigger with only keywords",
			content: `---
name: test-skill
description: Test description
triggers:
  - keywords:
      - testing
      - tdd
---
<role>test</role>`,
			verify: func(t *testing.T, meta *SkillMeta) {
				assert.Len(t, meta.ExplicitTriggers, 1)
				assert.Equal(t, []string{"testing", "tdd"}, meta.ExplicitTriggers[0].Keywords)
				assert.Equal(t, 0.7, meta.ExplicitTriggers[0].Weight)
				assert.Empty(t, meta.ExplicitTriggers[0].Patterns)
			},
		},
		{
			name: "trigger with only file_pattern",
			content: `---
name: test-skill
description: Test description
triggers:
  - file_patterns:
      - "**/*_test.go"
---
<role>test</role>`,
			verify: func(t *testing.T, meta *SkillMeta) {
				assert.Len(t, meta.ExplicitTriggers, 1)
				assert.Equal(t, []string{"**/*_test.go"}, meta.ExplicitTriggers[0].FilePatterns)
				assert.Equal(t, 0.7, meta.ExplicitTriggers[0].Weight)
				assert.Empty(t, meta.ExplicitTriggers[0].Patterns)
			},
		},
		{
			name: "multiple triggers with mixed fields",
			content: `---
name: test-skill
description: Test description
triggers:
  - patterns:
      - "write.*test"
    keywords:
      - testing
    weight: 0.9
  - file_patterns:
      - "**/*.go"
    weight: 0.6
  - keywords:
      - go code
---
<role>test</role>`,
			verify: func(t *testing.T, meta *SkillMeta) {
				assert.Len(t, meta.ExplicitTriggers, 3)
				assert.Equal(t, []string{"write.*test"}, meta.ExplicitTriggers[0].Patterns)
				assert.Equal(t, []string{"testing"}, meta.ExplicitTriggers[0].Keywords)
				assert.Equal(t, 0.9, meta.ExplicitTriggers[0].Weight)
				assert.Equal(t, []string{"**/*.go"}, meta.ExplicitTriggers[1].FilePatterns)
				assert.Equal(t, 0.6, meta.ExplicitTriggers[1].Weight)
				assert.Equal(t, []string{"go code"}, meta.ExplicitTriggers[2].Keywords)
				assert.Equal(t, 0.7, meta.ExplicitTriggers[2].Weight)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "skill.md")
			err := os.WriteFile(path, []byte(tt.content), 0o644)
			require.NoError(t, err)

			result, err := p.ParseSkillFile(path)
			require.NoError(t, err)
			tt.verify(t, result)
		})
	}
}
