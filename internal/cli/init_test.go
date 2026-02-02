package cli

import (
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAgentPath(t *testing.T) {
	tests := []struct {
		name     string
		tool     string
		prefix   string
		agentID  string
		expected string
	}{
		{
			name:     "claude with prefixed agent name",
			tool:     "claude",
			prefix:   "ent",
			agentID:  "ent:coder",
			expected: ".claude/agents/ent/coder.md",
		},
		{
			name:     "opencode with prefixed agent name",
			tool:     "opencode",
			prefix:   "ent",
			agentID:  "ent:planner",
			expected: ".opencode/agents/ent/planner.md",
		},
		{
			name:     "claude with unprefixed agent name",
			tool:     "claude",
			prefix:   "ent",
			agentID:  "tester",
			expected: ".claude/agents/ent/tester.md",
		},
		{
			name:     "multiple colons in name",
			tool:     "claude",
			prefix:   "myapp",
			agentID:  "plugin:myapp:special",
			expected: ".claude/agents/myapp/special.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAgentPath(tt.tool, tt.prefix, tt.agentID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderAgent_NoDoubleDashes(t *testing.T) {
	// Setup test template (mimics claude.yaml.tmpl)
	tplContent := `---
name: {{.Name}}
description: "{{.Description}}"
model: {{.Model}}
---`

	tpl, err := template.New("test").Parse(tplContent)
	require.NoError(t, err)

	meta := &agentMeta{
		Name:        "test",
		Description: "Test agent",
		Model:       "main",
	}

	prompt := "# Test Prompt Content"

	result, err := renderAgent(meta, prompt, tpl)
	require.NoError(t, err)

	// Split into lines to verify structure
	lines := strings.Split(result, "\n")

	// Find all lines with exactly "---"
	dashLines := []int{}
	for i, line := range lines {
		if line == "---" {
			dashLines = append(dashLines, i)
		}
	}

	// Should have exactly 2 lines with "---" (opening and closing frontmatter)
	assert.Len(t, dashLines, 2, "Should have exactly 2 '---' delimiters")

	// First should be at line 0, second at line 4 (after frontmatter)
	if len(dashLines) == 2 {
		assert.Equal(t, 0, dashLines[0], "First '---' should be at line 0")
		assert.Greater(t, dashLines[1], 0, "Second '---' should be after frontmatter")
	}

	// Verify no double dashes
	for i := 0; i < len(lines)-1; i++ {
		if lines[i] == "---" {
			assert.NotEqual(t, "---", lines[i+1], "Should not have consecutive '---' lines at position %d", i)
		}
	}

	// Verify content is present (shared content is now loaded via skills)
	assert.Contains(t, result, "Test Prompt Content")
}

func TestValidateAgent(t *testing.T) {
	tests := []struct {
		name    string
		agent   agentMeta
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid agent",
			agent: agentMeta{
				Name:        "coder",
				Description: "A valid description with more than 10 chars",
				Model:       "main",
				Color:       "#FF0000",
				Role:        "execution",
				Complexity:  "standard",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			agent: agentMeta{
				Description: "Valid description",
				Model:       "main",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "invalid name with colon",
			agent: agentMeta{
				Name:        "invalid:name",
				Description: "Valid description",
				Model:       "main",
			},
			wantErr: true,
			errMsg:  "name should not contain ':'",
		},
		{
			name: "missing description",
			agent: agentMeta{
				Name:  "test",
				Model: "main",
			},
			wantErr: true,
			errMsg:  "description is required",
		},
		{
			name: "description too short",
			agent: agentMeta{
				Name:        "test",
				Description: "Short",
				Model:       "main",
			},
			wantErr: true,
			errMsg:  "description must be at least 10 characters",
		},
		{
			name: "missing model",
			agent: agentMeta{
				Name:        "test",
				Description: "Valid description",
			},
			wantErr: true,
			errMsg:  "model is required",
		},
		{
			name: "invalid model",
			agent: agentMeta{
				Name:        "test",
				Description: "Valid description",
				Model:       "invalid",
			},
			wantErr: true,
			errMsg:  "model must be one of [fast, main, heavy]",
		},
		{
			name: "invalid color format",
			agent: agentMeta{
				Name:        "test",
				Description: "Valid description",
				Model:       "main",
				Color:       "red",
			},
			wantErr: true,
			errMsg:  "color must be a hex code starting with #",
		},
		{
			name: "invalid role",
			agent: agentMeta{
				Name:        "test",
				Description: "Valid description",
				Model:       "main",
				Role:        "invalid",
			},
			wantErr: true,
			errMsg:  "role must be one of [planning, execution, validation, research]",
		},
		{
			name: "invalid complexity",
			agent: agentMeta{
				Name:        "test",
				Description: "Valid description",
				Model:       "main",
				Complexity:  "invalid",
			},
			wantErr: true,
			errMsg:  "complexity must be one of [simple, standard, heavy]",
		},
		{
			name: "invalid tool preset",
			agent: agentMeta{
				Name:        "test",
				Description: "Valid description",
				Model:       "main",
				ToolPresets: []string{"invalid-preset"},
			},
			wantErr: true,
			errMsg:  "unknown tool preset",
		},
		{
			name: "valid agent with valid tool presets",
			agent: agentMeta{
				Name:        "test",
				Description: "Valid description with presets",
				Model:       "main",
				ToolPresets: []string{"readonly", "editing"},
			},
			wantErr: false,
		},
		{
			name: "invalid disallowed tool preset",
			agent: agentMeta{
				Name:                  "test",
				Description:           "Valid description",
				Model:                 "main",
				DisallowedToolPresets: []string{"invalid-preset"},
			},
			wantErr: true,
			errMsg:  "unknown disallowed tool preset",
		},
		{
			name: "invalid dependency with colon",
			agent: agentMeta{
				Name:         "test",
				Description:  "Valid description",
				Model:        "main",
				Dependencies: []string{"ent:invalid"},
			},
			wantErr: true,
			errMsg:  "dependency should not contain ':'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAgent(&tt.agent, "test.yaml")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
