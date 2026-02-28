package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/agent"
	"gopkg.in/yaml.v3"
)

func TestAgentListHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		agents        []agent.AgentMeta
		input         AgentListInput
		wantAgents    int
		wantTextInMsg string
		dontWantInMsg string
	}{
		{
			name:          "empty list",
			agents:        nil,
			input:         AgentListInput{},
			wantAgents:    0,
			wantTextInMsg: "No agents found",
		},
		{
			name: "single agent",
			agents: []agent.AgentMeta{
				{Name: "coder", Description: "Code agent", Model: "claude-sonnet"},
			},
			input:         AgentListInput{},
			wantAgents:    1,
			wantTextInMsg: "coder",
		},
		{
			name: "multiple agents",
			agents: []agent.AgentMeta{
				{Name: "coder", Description: "Code agent", Model: "claude-sonnet"},
				{Name: "planner", Description: "Planning agent", Model: "claude-haiku"},
				{Name: "reviewer", Description: "Review agent", Model: "claude-opus"},
			},
			input:         AgentListInput{},
			wantAgents:    3,
			wantTextInMsg: "coder",
		},
		{
			name: "filter by role",
			agents: []agent.AgentMeta{
				{Name: "coder", Description: "Code agent", Model: "claude-sonnet", Role: "execution"},
				{Name: "planner", Description: "Planning agent", Model: "claude-haiku", Role: "planning"},
			},
			input:         AgentListInput{Role: "planning"},
			wantAgents:    1,
			wantTextInMsg: "planner",
			dontWantInMsg: "coder",
		},
		{
			name: "filter by complexity",
			agents: []agent.AgentMeta{
				{Name: "fast-agent", Description: "Fast agent", Model: "claude-haiku", Complexity: "fast"},
				{Name: "heavy-agent", Description: "Heavy agent", Model: "claude-opus", Complexity: "heavy"},
			},
			input:         AgentListInput{Complexity: "fast"},
			wantAgents:    1,
			wantTextInMsg: "fast-agent",
			dontWantInMsg: "heavy-agent",
		},
		{
			name: "filter by role and complexity",
			agents: []agent.AgentMeta{
				{Name: "agent1", Description: "Agent 1", Model: "claude-haiku", Role: "planning", Complexity: "fast"},
				{Name: "agent2", Description: "Agent 2", Model: "claude-sonnet", Role: "planning", Complexity: "standard"},
				{Name: "agent3", Description: "Agent 3", Model: "claude-opus", Role: "execution", Complexity: "heavy"},
			},
			input:         AgentListInput{Role: "planning", Complexity: "fast"},
			wantAgents:    1,
			wantTextInMsg: "agent1",
			dontWantInMsg: "agent2",
		},
		{
			name: "filter no match",
			agents: []agent.AgentMeta{
				{Name: "coder", Description: "Code agent", Model: "claude-sonnet", Role: "execution"},
			},
			input:         AgentListInput{Role: "planning"},
			wantAgents:    0,
			wantTextInMsg: "No agents found matching filters",
		},
		{
			name: "agent with skills displayed",
			agents: []agent.AgentMeta{
				{Name: "coder", Description: "Code agent", Model: "claude-sonnet", Skills: []string{"go-code", "go-test"}},
			},
			input:         AgentListInput{},
			wantAgents:    1,
			wantTextInMsg: "**Skills**",
		},
		{
			name: "agent with dependencies displayed",
			agents: []agent.AgentMeta{
				{Name: "coder", Description: "Code agent", Model: "claude-sonnet", Dependencies: []string{"planner"}},
			},
			input:         AgentListInput{},
			wantAgents:    1,
			wantTextInMsg: "**Dependencies**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := populateAgentRegistry(t, tt.agents)

			handler := agentListHandler(registry)
			result, resp, err := handler(t.Context(), nil, tt.input)

			assert.NoError(t, err)
			require.NotNil(t, result)
			require.Len(t, result.Content, 1)

			textContent, ok := result.Content[0].(*mcp.TextContent)
			require.True(t, ok)

			if tt.wantAgents > 0 {
				assert.Contains(t, textContent.Text, tt.wantTextInMsg)

				if tt.dontWantInMsg != "" {
					assert.NotContains(t, textContent.Text, tt.dontWantInMsg)
				}

				listResp, ok := resp.(AgentListResponse)
				require.True(t, ok)
				assert.Equal(t, tt.wantAgents, listResp.Total)
				assert.Len(t, listResp.Agents, tt.wantAgents)
			} else {
				assert.Contains(t, textContent.Text, tt.wantTextInMsg)

				listResp, ok := resp.(AgentListResponse)
				require.True(t, ok)
				assert.Equal(t, 0, listResp.Total)
				assert.Empty(t, listResp.Agents)
			}
		})
	}
}

func TestAgentListHandler_ResponseStructure(t *testing.T) {
	t.Parallel()

	registry := populateAgentRegistry(t, []agent.AgentMeta{
		{Name: "coder", Description: "Code agent", Model: "claude-sonnet", Role: "execution", Complexity: "standard", Skills: []string{"go-code"}},
		{Name: "planner", Description: "Planning agent", Model: "claude-haiku", Role: "planning"},
	})

	handler := agentListHandler(registry)
	result, resp, err := handler(t.Context(), nil, AgentListInput{})

	assert.NoError(t, err)
	assert.NotNil(t, result)

	listResp, ok := resp.(AgentListResponse)
	require.True(t, ok)
	assert.Equal(t, 2, listResp.Total)
	assert.Len(t, listResp.Agents, 2)

	assert.Equal(t, "coder", listResp.Agents[0].Name)
	assert.Equal(t, "Code agent", listResp.Agents[0].Description)
	assert.Equal(t, "execution", listResp.Agents[0].Role)
	assert.Equal(t, "standard", listResp.Agents[0].Complexity)
	assert.Equal(t, "claude-sonnet", listResp.Agents[0].Model)
	assert.Equal(t, []string{"go-code"}, listResp.Agents[0].Skills)
}

func TestAgentListHandler_MarkdownFormatting(t *testing.T) {
	t.Parallel()

	registry := populateAgentRegistry(t, []agent.AgentMeta{
		{Name: "test-agent", Description: "Test description", Model: "claude-sonnet", Role: "execution"},
	})

	handler := agentListHandler(registry)
	result, _, err := handler(t.Context(), nil, AgentListInput{})

	assert.NoError(t, err)

	textContent := result.Content[0].(*mcp.TextContent)
	text := textContent.Text

	assert.Contains(t, text, "# Available Agents")
	assert.Contains(t, text, "## 1. test-agent")
	assert.Contains(t, text, "**Description**: Test description")
	assert.Contains(t, text, "**Model**: claude-sonnet")
	assert.Contains(t, text, "**Role**: execution")
}

func TestAgentListHandler_FilteredMarkdown(t *testing.T) {
	t.Parallel()

	registry := populateAgentRegistry(t, []agent.AgentMeta{
		{Name: "planner", Description: "Planning", Model: "claude-haiku", Role: "planning"},
	})

	handler := agentListHandler(registry)
	result, _, err := handler(t.Context(), nil, AgentListInput{Role: "planning"})

	assert.NoError(t, err)

	textContent := result.Content[0].(*mcp.TextContent)
	text := textContent.Text

	assert.Contains(t, text, "*Filtered*")
}

func populateAgentRegistry(tb testing.TB, agents []agent.AgentMeta) *agent.Registry {
	tb.Helper()

	if len(agents) == 0 {
		return agent.NewRegistry()
	}

	tmpDir := tb.TempDir()

	for _, a := range agents {
		data, err := yaml.Marshal(a)
		require.NoError(tb, err)

		filename := a.Name + ".yaml"
		require.NoError(tb, os.WriteFile(filepath.Join(tmpDir, filename), data, 0o644))
	}

	registry := agent.NewRegistry()
	require.NoError(tb, registry.Load(tmpDir))

	return registry
}
