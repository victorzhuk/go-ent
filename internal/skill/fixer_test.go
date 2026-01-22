package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixer_FixFrontmatter(t *testing.T) {
	f := NewFixer()

	tests := []struct {
		name        string
		input       string
		wantFixed   bool
		wantChanges int
	}{
		{
			name: "unsorted keys",
			input: `---
version: "1.0.0"
name: go-code
description: Go coding patterns
---

Content`,
			wantFixed:   true,
			wantChanges: 1,
		},
		{
			name:        "no frontmatter",
			input:       `Just some content without frontmatter`,
			wantFixed:   false,
			wantChanges: 0,
		},
		{
			name: "invalid yaml",
			input: `---
name: go-code
description: [unclosed array
---

Content`,
			wantFixed:   false,
			wantChanges: 0,
		},
		{
			name: "with tags and tools",
			input: `---
allowedTools:
  - bash
  - write
name: go-code
tags:
  - go
  - code
---

Content`,
			wantFixed:   false,
			wantChanges: 0,
		},
		{
			name: "triggers list",
			input: `---
name: go-code
triggers:
  - patterns:
      - "test.*"
    weight: 0.8
---

Content`,
			wantFixed:   true,
			wantChanges: 1,
		},
		{
			name: "unordered keys",
			input: `---
version: "1.0.0"
tags:
  - go
  - code
name: go-code
description: Go coding patterns
---

Content`,
			wantFixed:   true,
			wantChanges: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, changes, err := f.FixFrontmatter(tt.input)

			if tt.name == "invalid yaml" {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantChanges, len(changes))

			if tt.wantFixed {
				assert.NotEqual(t, tt.input, normalized)
			} else {
				assert.Equal(t, tt.input, normalized)
			}
		})
	}
}

func TestFixer_FrontmatterOrdering(t *testing.T) {
	f := NewFixer()

	input := `---
version: "1.0.0"
tags:
  - go
  - code
name: go-code
description: Go coding patterns
author: John Doe
---

Content`

	normalized, _, err := f.FixFrontmatter(input)
	require.NoError(t, err)

	lines := strings.Split(normalized, "\n")
	keyOrder := []string{}

	inFrontmatter := false
	for _, line := range lines {
		if line == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}
		if strings.Contains(line, ":") && !strings.HasPrefix(line, " ") {
			key := strings.SplitN(line, ":", 2)[0]
			keyOrder = append(keyOrder, key)
		}
	}

	expectedOrder := []string{"author", "description", "name", "tags", "version"}
	assert.Equal(t, expectedOrder, keyOrder)
}

func TestFixer_PreservesContent(t *testing.T) {
	f := NewFixer()

	bodyContent := `# Title

Some content

## Subtitle

More content
`

	input := `---
name: go-code
version: "1.0.0"
---

` + bodyContent

	normalized, _, err := f.FixFrontmatter(input)
	require.NoError(t, err)

	parts := strings.SplitN(normalized, "\n---\n", 3)
	require.Len(t, parts, 2)
	assert.Contains(t, parts[1], bodyContent)
}

func TestFixer_HasFixableFrontmatter(t *testing.T) {
	f := NewFixer()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "unsorted keys are fixable",
			input: "---\nversion: \"1.0.0\"\ntags:\n  - go\n  - code\nname: test\n---\nContent",
			want:  true,
		},
		{
			name:  "already sorted is not fixable",
			input: "---\nname: test\nversion: 1.0.0\n---\nContent",
			want:  false,
		},
		{
			name:  "no frontmatter is not fixable",
			input: "Just content",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.HasFixableFrontmatter(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFixer_FixXMLSections(t *testing.T) {
	f := NewFixer()

	tests := []struct {
		name        string
		input       string
		wantFixed   bool
		wantChanges int
	}{
		{
			name: "role section with extra whitespace",
			input: `---
name: test
version: 1.0.0
---

<role>

  Expert Go developer

  Focus on clean code

</role>`,
			wantFixed:   true,
			wantChanges: 1,
		},
		{
			name: "instructions section with inconsistent formatting",
			input: `---
name: test
version: 1.0.0
---

<instructions>


Step 1
  Step 2
Step 3

</instructions>`,
			wantFixed:   true,
			wantChanges: 1,
		},
		{
			name: "no xml sections",
			input: `---
name: test
version: 1.0.0
---

Just markdown content`,
			wantFixed:   false,
			wantChanges: 0,
		},
		{
			name: "multiple sections",
			input: `---
name: test
version: 1.0.0
---

<role>

  Expert

</role>

<instructions>


  Step 1
  Step 2


</instructions>`,
			wantFixed:   true,
			wantChanges: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, changes, err := f.FixXMLSections(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantChanges, len(changes))

			if tt.wantFixed {
				assert.NotEqual(t, tt.input, normalized)

				for _, change := range changes {
					assert.Equal(t, "xml-section-format", change.Rule)
					assert.Contains(t, change.Message, "normalized")
				}
			} else {
				assert.Equal(t, tt.input, normalized)
			}
		})
	}
}

func TestFixer_NormalizeSectionContent(t *testing.T) {
	f := NewFixer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty content",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   \n  \n   ",
			expected: "",
		},
		{
			name:     "normal content",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "content with leading/trailing spaces",
			input:    "  Line 1  \n   Line 2   ",
			expected: "Line 1\nLine 2",
		},
		{
			name:     "content with blank lines",
			input:    "Line 1\n\nLine 2\n\n\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.normalizeSectionContent(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFixer_HasFixableXML(t *testing.T) {
	f := NewFixer()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "section with extra whitespace is fixable",
			input: "---\nname: test\n---\n\n<role>\n\n  Content\n\n</role>",
			want:  true,
		},
		{
			name:  "already normalized is not fixable",
			input: "---\nname: test\n---\n\n<role>\nContent\n</role>",
			want:  false,
		},
		{
			name:  "no xml sections is not fixable",
			input: "---\nname: test\n---\n\nJust content",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.HasFixableXML(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFixer_FixFileWithXML(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFixer()

	testFile := filepath.Join(tmpDir, "SKILL.md")
	content := `---
name: test
version: 1.0.0
---

<role>

  Expert Go developer

</role>

<instructions>

  Step 1
  Step 2


</instructions>`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	result, err := f.FixFile(testFile, content)
	require.NoError(t, err)
	assert.True(t, result.Fixed)
	assert.True(t, len(result.Changes) >= 2)

	for _, change := range result.Changes {
		assert.NotEmpty(t, change.Rule)
		assert.NotEmpty(t, change.Message)
	}

	fixedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)

	assert.NotContains(t, string(fixedContent), "  \n\n")
	assert.Contains(t, string(fixedContent), "<role>")
	assert.Contains(t, string(fixedContent), "</role>")
	assert.Contains(t, string(fixedContent), "<instructions>")
	assert.Contains(t, string(fixedContent), "</instructions>")
}

func TestFixer_PreservesFrontmatterAndMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFixer()

	testFile := filepath.Join(tmpDir, "SKILL.md")
	content := `---
name: test-skill
version: 1.0.0
description: Test description
---

# Introduction

This is markdown content.

<role>

  Expert Go developer

</role>

## More Content

More markdown text.

<instructions>

  Step 1
  Step 2


</instructions>

### Conclusion

Final markdown.
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	result, err := f.FixFile(testFile, content)
	require.NoError(t, err)
	assert.True(t, result.Fixed)

	fixedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)

	fixedStr := string(fixedContent)

	assert.Contains(t, fixedStr, "---")
	assert.Contains(t, fixedStr, "name: test-skill")
	assert.Contains(t, fixedStr, "# Introduction")
	assert.Contains(t, fixedStr, "This is markdown content")
	assert.Contains(t, fixedStr, "## More Content")
	assert.Contains(t, fixedStr, "More markdown text")
	assert.Contains(t, fixedStr, "### Conclusion")
	assert.Contains(t, fixedStr, "Final markdown")

	assert.NotContains(t, fixedStr, "  \n\n")
}

func TestFixer_FixValidationIssues_MissingName(t *testing.T) {
	f := NewFixer()

	input := `---
description: Test description
version: 1.0.0
---

Content`

	fixed, changes, err := f.FixValidationIssues(input)
	require.NoError(t, err)
	assert.Len(t, changes, 1)
	assert.Contains(t, fixed, "name: unnamed-skill")
	assert.Contains(t, changes[0].Message, "name")
}

func TestFixer_FixValidationIssues_MissingDescription(t *testing.T) {
	f := NewFixer()

	input := `---
name: test-skill
version: 1.0.0
---

Content`

	fixed, changes, err := f.FixValidationIssues(input)
	require.NoError(t, err)
	assert.Len(t, changes, 1)
	assert.Contains(t, fixed, "description: Auto-generated description")
	assert.Contains(t, changes[0].Message, "description")
}

func TestFixer_FixValidationIssues_InvalidName(t *testing.T) {
	f := NewFixer()

	input := `---
name: My Test Skill
description: Test
version: 1.0.0
---

Content`

	fixed, changes, err := f.FixValidationIssues(input)
	require.NoError(t, err)
	assert.Len(t, changes, 1)
	assert.Contains(t, fixed, "name: my-test-skill")
	assert.Contains(t, changes[0].Message, "normalized")
}

func TestFixer_FixValidationIssues_InvalidVersion(t *testing.T) {
	f := NewFixer()

	tests := []struct {
		name          string
		input         string
		expected      string
		expectChanged bool
	}{
		{
			name:          "single number",
			input:         "2",
			expected:      "2.0.0",
			expectChanged: true,
		},
		{
			name:          "two numbers",
			input:         "2.5",
			expected:      "2.5.0",
			expectChanged: true,
		},
		{
			name:          "already valid",
			input:         "1.2.3",
			expected:      "1.2.3",
			expectChanged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := fmt.Sprintf(`---
name: test
description: Test
version: %s
---`, tt.input)

			fixed, changes, err := f.FixValidationIssues(input)
			require.NoError(t, err)
			assert.Contains(t, fixed, fmt.Sprintf("version: %s", tt.expected))
			if tt.expectChanged {
				assert.Len(t, changes, 1)
				assert.Contains(t, changes[0].Message, "version")
			}
		})
	}
}

func TestFixer_FixValidationIssues_MissingClosingTag(t *testing.T) {
	f := NewFixer()

	input := `---
name: test
description: Test
version: 1.0.0
---

<role>
Expert developer`

	fixed, changes, err := f.FixValidationIssues(input)
	require.NoError(t, err)
	assert.Len(t, changes, 1)
	assert.Contains(t, fixed, "</role>")
	assert.Contains(t, changes[0].Message, "closing tag")
}

func TestFixer_FixValidationIssues_ConstraintListFormat(t *testing.T) {
	f := NewFixer()

	input := `---
name: test
description: Test
version: 1.0.0
---

<constraints>
Focus on Go code
Use standard library
No external deps
</constraints>`

	fixed, changes, err := f.FixValidationIssues(input)
	require.NoError(t, err)
	assert.Len(t, changes, 1)
	assert.Contains(t, fixed, "- Focus on Go code")
	assert.Contains(t, fixed, "- Use standard library")
	assert.Contains(t, fixed, "- No external deps")
	assert.Contains(t, changes[0].Message, "3")
}

func TestFixer_FixValidationIssues_NoChanges(t *testing.T) {
	f := NewFixer()

	input := `---
name: test-skill
description: Test description
version: 1.0.0
---

<role>
Expert developer
</role>`

	fixed, changes, err := f.FixValidationIssues(input)
	require.NoError(t, err)
	assert.Len(t, changes, 0)
	assert.Equal(t, input, fixed)
}

func TestFixer_HasFixableValidationIssues(t *testing.T) {
	f := NewFixer()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name: "missing name",
			input: `---
description: test
---
Content`,
			want: true,
		},
		{
			name: "valid content",
			input: `---
name: test
description: test
version: 1.0.0
---
Content`,
			want: false,
		},
		{
			name: "invalid name format",
			input: `---
name: My Skill
description: test
---
Content`,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.HasFixableValidationIssues(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFixer_FixValidationFile(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFixer()

	testFile := filepath.Join(tmpDir, "SKILL.md")
	content := `---
description: Test description
version: 1.0.0
---

<role>
Expert developer
</role>`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	result, err := f.FixValidationFile(testFile, content)
	require.NoError(t, err)
	assert.True(t, result.Fixed)
	assert.Len(t, result.Changes, 1)

	fixedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Contains(t, string(fixedContent), "name: unnamed-skill")
}

func TestNormalizeSkillName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid name",
			input:    "go-code",
			expected: "go-code",
		},
		{
			name:     "uppercase to lowercase",
			input:    "Go-Code",
			expected: "go-code",
		},
		{
			name:     "spaces to hyphens",
			input:    "go code expert",
			expected: "go-code-expert",
		},
		{
			name:     "underscores to hyphens",
			input:    "go_code_expert",
			expected: "go-code-expert",
		},
		{
			name:     "multiple hyphens collapsed",
			input:    "go--code--expert",
			expected: "go-code-expert",
		},
		{
			name:     "special chars removed",
			input:    "go@code#expert!",
			expected: "go-code-expert",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "unnamed-skill",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeSkillName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid semver",
			input:    "1.0.0",
			expected: "1.0.0",
		},
		{
			name:     "v prefix valid",
			input:    "v1.0.0",
			expected: "v1.0.0",
		},
		{
			name:     "single number",
			input:    "2",
			expected: "2.0.0",
		},
		{
			name:     "two numbers",
			input:    "2.5",
			expected: "2.5.0",
		},
		{
			name:     "invalid format",
			input:    "invalid",
			expected: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeVersion(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
