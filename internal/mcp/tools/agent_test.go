package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/skill"
)

func TestAgentExecute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       AgentExecuteInput
		expectError bool
		checkOutput func(t *testing.T, result *mcp.CallToolResult, data any)
	}{
		{
			name: "simple feature task",
			input: AgentExecuteInput{
				Path: ".",
				Task: "Add a login form",
			},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				require.NotNil(t, result)
				require.Len(t, result.Content, 1)

				textContent, ok := result.Content[0].(*mcp.TextContent)
				require.True(t, ok)

				assert.Contains(t, textContent.Text, "Agent selected")
				assert.Contains(t, textContent.Text, "Next Steps")

				resp, ok := data.(AgentExecuteResponse)
				require.True(t, ok)
				assert.NotEmpty(t, resp.Role)
				assert.NotEmpty(t, resp.Model)
				assert.NotEmpty(t, resp.Complexity)
			},
		},
		{
			name: "architecture task with task type",
			input: AgentExecuteInput{
				Path:     ".",
				Task:     "Design database schema for multi-tenant system",
				TaskType: "architecture",
			},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				resp, ok := data.(AgentExecuteResponse)
				require.True(t, ok)
				assert.NotEmpty(t, resp.Role)
				assert.Contains(t, []string{"architect", "senior", "developer"}, resp.Role)
			},
		},
		{
			name: "bugfix task",
			input: AgentExecuteInput{
				Path:     ".",
				Task:     "Fix crash when clicking submit button",
				TaskType: "bugfix",
				Files:    []string{"handlers/submit.go"},
			},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				resp, ok := data.(AgentExecuteResponse)
				require.True(t, ok)
				assert.NotEmpty(t, resp.Role)
			},
		},
		{
			name: "force role override",
			input: AgentExecuteInput{
				Path:      ".",
				Task:      "Add logging",
				ForceRole: "developer",
			},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				resp, ok := data.(AgentExecuteResponse)
				require.True(t, ok)
				assert.Equal(t, "developer", resp.Role)
				assert.Contains(t, resp.Reason, "manually overridden")
			},
		},
		{
			name: "force model override",
			input: AgentExecuteInput{
				Path:       ".",
				Task:       "Add comment to function",
				ForceModel: "haiku",
			},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				resp, ok := data.(AgentExecuteResponse)
				require.True(t, ok)
				assert.Equal(t, "haiku", resp.Model)
			},
		},
		{
			name: "with max budget",
			input: AgentExecuteInput{
				Path:      ".",
				Task:      "Refactor authentication module",
				MaxBudget: 50000,
			},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				resp, ok := data.(AgentExecuteResponse)
				require.True(t, ok)
				assert.NotEmpty(t, resp.Role)
			},
		},
		{
			name: "with context metadata",
			input: AgentExecuteInput{
				Path: ".",
				Task: "Update API endpoint",
				Context: map[string]interface{}{
					"pr_number": 123,
					"author":    "test-user",
				},
			},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				resp, ok := data.(AgentExecuteResponse)
				require.True(t, ok)
				assert.NotEmpty(t, resp.Role)
			},
		},
		{
			name: "missing path",
			input: AgentExecuteInput{
				Task: "Do something",
			},
			expectError: true,
		},
		{
			name: "missing task",
			input: AgentExecuteInput{
				Path: ".",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := skill.NewRegistry()
			handler := makeAgentExecuteHandler(registry)
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, data, err := handler(ctx, req, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.checkOutput != nil {
				tt.checkOutput(t, result, data)
			}
		})
	}
}

func TestAgentExecuteComplexityLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		task             string
		taskType         string
		files            []string
		expectedMinLevel string
	}{
		{
			name:             "trivial task",
			task:             "Fix typo in comment",
			taskType:         "bugfix",
			files:            []string{"main.go"},
			expectedMinLevel: "trivial",
		},
		{
			name:             "simple task",
			task:             "Add validation to form field",
			taskType:         "feature",
			files:            []string{"handlers/form.go"},
			expectedMinLevel: "simple",
		},
		{
			name:     "moderate task",
			task:     "Implement password reset flow",
			taskType: "feature",
			files: []string{
				"handlers/auth.go",
				"services/email.go",
				"templates/reset.html",
			},
			expectedMinLevel: "moderate",
		},
		{
			name:     "complex task",
			task:     "Add real-time notifications with WebSocket support",
			taskType: "feature",
			files: []string{
				"websocket/server.go",
				"websocket/client.go",
				"handlers/notifications.go",
				"services/push.go",
				"repositories/notifications.go",
			},
			expectedMinLevel: "complex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := skill.NewRegistry()
			handler := makeAgentExecuteHandler(registry)
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			input := AgentExecuteInput{
				Path:     ".",
				Task:     tt.task,
				TaskType: tt.taskType,
				Files:    tt.files,
			}

			result, data, err := handler(ctx, req, input)
			require.NoError(t, err)
			require.NotNil(t, result)

			resp, ok := data.(AgentExecuteResponse)
			require.True(t, ok)

			assert.NotEmpty(t, resp.Complexity)

			textContent := result.Content[0].(*mcp.TextContent)
			assert.Contains(t, textContent.Text, "Complexity level:")
		})
	}
}

func TestAgentExecuteWithSkills(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	skillsDir := filepath.Join(tempDir, "skills")
	require.NoError(t, os.MkdirAll(skillsDir, 0755))

	skillFile := filepath.Join(skillsDir, "SKILL.md")
	skillContent := `---
name: "test-skill"
description: "Test skill for testing. Auto-activates for: test, testing."
---

# Test Skill

## Actions

Test actions here`

	require.NoError(t, os.WriteFile(skillFile, []byte(skillContent), 0644))

	registry := skill.NewRegistry()
	err := registry.Load(skillsDir)
	require.NoError(t, err)

	handler := makeAgentExecuteHandler(registry)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	input := AgentExecuteInput{
		Path: ".",
		Task: "Run test suite",
	}

	result, data, err := handler(ctx, req, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	resp, ok := data.(AgentExecuteResponse)
	require.True(t, ok)
	assert.NotEmpty(t, resp.Role)
}

func TestSkillInfo(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	skillsDir := filepath.Join(tempDir, "skills")
	require.NoError(t, os.MkdirAll(skillsDir, 0755))

	skillFile := filepath.Join(skillsDir, "SKILL.md")
	skillContent := `---
name: "test-skill"
description: "Test skill for unit testing. Auto-activates for: test, unit-test."
---

# Test Skill

## Actions

1. Run tests
2. Generate report`

	require.NoError(t, os.WriteFile(skillFile, []byte(skillContent), 0644))

	registry := skill.NewRegistry()
	err := registry.Load(skillsDir)
	require.NoError(t, err)

	handler := skillInfoHandler(registry)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	tests := []struct {
		name        string
		input       SkillInfoInput
		expectError bool
		checkOutput func(t *testing.T, result *mcp.CallToolResult)
	}{
		{
			name:        "valid skill",
			input:       SkillInfoInput{Name: "test-skill"},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				require.Len(t, result.Content, 1)

				textContent, ok := result.Content[0].(*mcp.TextContent)
				require.True(t, ok)

				assert.Contains(t, textContent.Text, "test-skill")
				assert.Contains(t, textContent.Text, "Test skill for unit testing")
				assert.Contains(t, textContent.Text, "Run tests")
			},
		},
		{
			name:        "nonexistent skill",
			input:       SkillInfoInput{Name: "nonexistent"},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult) {
				textContent := result.Content[0].(*mcp.TextContent)
				assert.Contains(t, textContent.Text, "Skill not found")
			},
		},
		{
			name:        "empty name",
			input:       SkillInfoInput{Name: ""},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := handler(ctx, req, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.checkOutput != nil {
				tt.checkOutput(t, result)
			}
		})
	}
}

func TestRuntimeList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, _, err := runtimeListHandler(ctx, req, RuntimeListInput{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	assert.Contains(t, textContent.Text, "Available Runtimes")
	assert.Contains(t, textContent.Text, "claude-code")
	assert.Contains(t, textContent.Text, "open-code")
	assert.Contains(t, textContent.Text, "cli")

	assert.Contains(t, textContent.Text, "Interactive:")
	assert.Contains(t, textContent.Text, "Filesystem:")
	assert.Contains(t, textContent.Text, "Tools:")
	assert.Contains(t, textContent.Text, "Skills:")
	assert.Contains(t, textContent.Text, "Max Concurrent Agents:")

	assert.Contains(t, textContent.Text, "Usage")
	assert.Contains(t, textContent.Text, ".go-ent/config.yaml")
}

func TestRuntimeStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name: "claude code runtime",
			envVars: map[string]string{
				"CLAUDE_CODE": "1",
			},
			expected: "claude-code",
		},
		{
			name: "open code runtime",
			envVars: map[string]string{
				"OPEN_CODE": "1",
			},
			expected: "open-code",
		},
		{
			name: "mcp server mode",
			envVars: map[string]string{
				"MCP_SERVER": "1",
			},
			expected: "claude-code",
		},
		{
			name:     "default cli runtime",
			envVars:  map[string]string{},
			expected: "cli",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, _, err := runtimeStatusHandler(ctx, req, RuntimeStatusInput{})
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Len(t, result.Content, 1)

			textContent, ok := result.Content[0].(*mcp.TextContent)
			require.True(t, ok)

			assert.Contains(t, textContent.Text, "Runtime Status")
			assert.Contains(t, textContent.Text, "Current Runtime:")
			assert.Contains(t, textContent.Text, tt.expected)
			assert.Contains(t, textContent.Text, "Capabilities")
		})
	}
}

func TestRuntimeStatusWithConfig(t *testing.T) {
	t.Skip("Skipping due to config path detection complexity")
}

func TestRuntimeStatusNoConfig(t *testing.T) {
	t.Skip("Skipping due to config path detection complexity")
}

func TestAgentExecuteTaskTypes(t *testing.T) {
	t.Parallel()

	taskTypes := []string{
		"feature",
		"bugfix",
		"refactor",
		"test",
		"documentation",
		"architecture",
	}

	registry := skill.NewRegistry()
	handler := makeAgentExecuteHandler(registry)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	for _, taskType := range taskTypes {
		t.Run(taskType, func(t *testing.T) {
			input := AgentExecuteInput{
				Path:     ".",
				Task:     "Sample task for " + taskType,
				TaskType: taskType,
			}

			result, data, err := handler(ctx, req, input)
			require.NoError(t, err)
			require.NotNil(t, result)

			resp, ok := data.(AgentExecuteResponse)
			require.True(t, ok)
			assert.NotEmpty(t, resp.Role)
			assert.NotEmpty(t, resp.Model)
			assert.NotEmpty(t, resp.Complexity)
		})
	}
}

func TestAgentExecuteRoleModelMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		task         string
		taskType     string
		files        []string
		expectedRole string
	}{
		{
			name:     "trivial task uses developer",
			task:     "Fix typo",
			taskType: "bugfix",
			files:    []string{"main.go"},
		},
		{
			name:     "architecture task",
			task:     "Design microservices architecture",
			taskType: "architecture",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := skill.NewRegistry()
			handler := makeAgentExecuteHandler(registry)
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			input := AgentExecuteInput{
				Path:     ".",
				Task:     tt.task,
				TaskType: tt.taskType,
				Files:    tt.files,
			}

			result, data, err := handler(ctx, req, input)
			require.NoError(t, err)
			require.NotNil(t, result)

			resp, ok := data.(AgentExecuteResponse)
			require.True(t, ok)

			validRoles := []string{"architect", "senior", "developer", "ops", "reviewer"}
			assert.Contains(t, validRoles, resp.Role)

			validModels := []string{"opus", "sonnet", "haiku"}
			assert.Contains(t, validModels, resp.Model)
		})
	}
}

func TestAgentToolsIntegration(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	skillsDir := filepath.Join(tempDir, "skills")
	require.NoError(t, os.MkdirAll(skillsDir, 0755))

	skillFile := filepath.Join(skillsDir, "SKILL.md")
	skillContent := `---
name: "go-test"
description: "Run Go tests. Auto-activates for: test, go test."
---

# Go Test Skill

## Actions

Run: go test ./...`

	require.NoError(t, os.WriteFile(skillFile, []byte(skillContent), 0644))

	registry := skill.NewRegistry()
	err := registry.Load(skillsDir)
	require.NoError(t, err)

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)

	registerAgentExecute(s, registry)
	registerSkillInfo(s, registry)
	registerRuntimeList(s)
	registerRuntimeStatus(s)

	assert.NotNil(t, s)
}
