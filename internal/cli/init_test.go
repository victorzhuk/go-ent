package cli

import (
	"os"
	"path/filepath"
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
			errMsg:  "role must be one of [planning, execution, validation, research, orchestration]",
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
			errMsg:  "complexity must be one of [auto, simple, standard, heavy]",
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
		{
			name: "valid primary mode",
			agent: agentMeta{
				Name:        "driver",
				Description: "Valid description with primary mode",
				Model:       "main",
				Mode:        "primary",
			},
			wantErr: false,
		},
		{
			name: "valid subagent mode",
			agent: agentMeta{
				Name:        "helper",
				Description: "Valid description with subagent mode",
				Model:       "main",
				Mode:        "subagent",
			},
			wantErr: false,
		},
		{
			name: "valid hidden mode",
			agent: agentMeta{
				Name:        "internal",
				Description: "Valid description with hidden mode",
				Model:       "main",
				Mode:        "hidden",
			},
			wantErr: false,
		},
		{
			name: "invalid mode",
			agent: agentMeta{
				Name:        "test",
				Description: "Valid description",
				Model:       "main",
				Mode:        "invalid",
			},
			wantErr: true,
			errMsg:  "mode must be one of [primary, subagent, hidden]",
		},
		{
			name: "valid orchestration role",
			agent: agentMeta{
				Name:        "driver",
				Description: "Driver agent for orchestration",
				Model:       "main",
				Role:        "orchestration",
			},
			wantErr: false,
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

func TestCleanDirs(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, tmpDir string)
		tool    string
		prefix  string
		dryRun  bool
		wantErr bool
		verify  func(t *testing.T, tmpDir string, dryRun bool)
	}{
		{
			name: "removes existing directories",
			setup: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".claude/agents/ent"), 0o755))
				require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".claude/commands/ent"), 0o755))
				require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".claude/skills/ent"), 0o755))
			},
			tool:    "claude",
			prefix:  "ent",
			dryRun:  false,
			wantErr: false,
			verify: func(t *testing.T, tmpDir string, dryRun bool) {
				if !dryRun {
					// Directories should not exist after cleaning
					_, err := os.Stat(filepath.Join(tmpDir, ".claude/agents/ent"))
					assert.True(t, os.IsNotExist(err), "agents directory should be removed")
					_, err = os.Stat(filepath.Join(tmpDir, ".claude/commands/ent"))
					assert.True(t, os.IsNotExist(err), "commands directory should be removed")
					_, err = os.Stat(filepath.Join(tmpDir, ".claude/skills/ent"))
					assert.True(t, os.IsNotExist(err), "skills directory should be removed")
				}
			},
		},
		{
			name: "dry run does not remove directories",
			setup: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".claude/agents/ent"), 0o755))
			},
			tool:    "claude",
			prefix:  "ent",
			dryRun:  true,
			wantErr: false,
			verify: func(t *testing.T, tmpDir string, dryRun bool) {
				if dryRun {
					// Directory should still exist in dry run mode
					_, err := os.Stat(filepath.Join(tmpDir, ".claude/agents/ent"))
					assert.NoError(t, err, "agents directory should still exist in dry run")
				}
			},
		},
		{
			name: "handles non-existent directories",
			setup: func(t *testing.T, tmpDir string) {
				// No setup needed
			},
			tool:    "claude",
			prefix:  "ent",
			dryRun:  false,
			wantErr: false,
			verify:  func(t *testing.T, tmpDir string, dryRun bool) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tt.setup(t, tmpDir)

			// Change to temp directory
			oldCwd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, os.Chdir(oldCwd))
			}()
			require.NoError(t, os.Chdir(tmpDir))

			err = cleanDirs(tt.tool, tt.prefix, tt.dryRun)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.verify(t, tmpDir, tt.dryRun)
		})
	}
}

func TestEffectiveMode(t *testing.T) {
	tests := []struct {
		name     string
		agent    agentMeta
		expected string
	}{
		{
			name:     "mode takes precedence",
			agent:    agentMeta{Mode: "primary"},
			expected: "primary",
		},
		{
			name:     "default is subagent",
			agent:    agentMeta{},
			expected: "subagent",
		},
		{
			name:     "explicit subagent mode",
			agent:    agentMeta{Mode: "subagent"},
			expected: "subagent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.agent.EffectiveMode()
			assert.Equal(t, tt.expected, result)
		})
	}
}
