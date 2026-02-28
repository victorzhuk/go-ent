package tools

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/agent"
)

func TestAgentInfoHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		agents        []agent.AgentMeta
		input         AgentInfoInput
		wantErr       bool
		wantTextInMsg string
		dontWantInMsg string
	}{
		{
			name:    "empty name returns error",
			agents:  nil,
			input:   AgentInfoInput{Name: ""},
			wantErr: true,
		},
		{
			name:    "agent not found",
			agents:  nil,
			input:   AgentInfoInput{Name: "nonexistent"},
			wantErr: true,
		},
		{
			name: "valid agent",
			agents: []agent.AgentMeta{
				{Name: "coder", Description: "Code agent", Model: "claude-sonnet"},
			},
			input:         AgentInfoInput{Name: "coder"},
			wantErr:       false,
			wantTextInMsg: "# Agent: coder",
		},
		{
			name: "agent with full metadata",
			agents: []agent.AgentMeta{
				{
					Name:        "coder",
					Description: "Code agent",
					Model:       "claude-sonnet",
					Role:        "execution",
					Complexity:  "standard",
					Color:       "blue",
					Skills:      []string{"go-code", "go-test"},
				},
			},
			input:         AgentInfoInput{Name: "coder"},
			wantErr:       false,
			wantTextInMsg: "**Role**: execution",
		},
		{
			name: "agent with skills displayed",
			agents: []agent.AgentMeta{
				{Name: "coder", Description: "Code agent", Model: "claude-sonnet", Skills: []string{"go-code"}},
			},
			input:         AgentInfoInput{Name: "coder"},
			wantErr:       false,
			wantTextInMsg: "## Skills",
		},
		{
			name: "agent with tool presets displayed",
			agents: []agent.AgentMeta{
				{Name: "coder", Description: "Code agent", Model: "claude-sonnet", ToolPresets: []string{"fs", "git"}},
			},
			input:         AgentInfoInput{Name: "coder"},
			wantErr:       false,
			wantTextInMsg: "## Tool Presets",
		},
		{
			name: "agent with disallowed tool presets displayed",
			agents: []agent.AgentMeta{
				{Name: "coder", Description: "Code agent", Model: "claude-sonnet", DisallowedToolPresets: []string{"dangerous"}},
			},
			input:         AgentInfoInput{Name: "coder"},
			wantErr:       false,
			wantTextInMsg: "## Disallowed Tool Presets",
		},
		{
			name: "agent with dependencies displayed",
			agents: []agent.AgentMeta{
				{Name: "coder", Description: "Code agent", Model: "claude-sonnet", Dependencies: []string{"planner"}},
			},
			input:         AgentInfoInput{Name: "coder"},
			wantErr:       false,
			wantTextInMsg: "## Dependencies",
		},
		{
			name: "agent without optional fields",
			agents: []agent.AgentMeta{
				{Name: "minimal", Description: "Minimal agent", Model: "claude-haiku"},
			},
			input:         AgentInfoInput{Name: "minimal"},
			wantErr:       false,
			wantTextInMsg: "# Agent: minimal",
			dontWantInMsg: "**Role**:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := populateAgentRegistry(t, tt.agents)

			handler := agentInfoHandler(registry)
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

func TestAgentInfoHandler_FullMetadata(t *testing.T) {
	t.Parallel()

	registry := populateAgentRegistry(t, []agent.AgentMeta{
		{
			Name:                  "full-agent",
			Description:           "Full featured agent",
			Model:                 "claude-opus",
			Role:                  "validation",
			Complexity:            "heavy",
			Color:                 "red",
			Skills:                []string{"go-review", "go-test"},
			ToolPresets:           []string{"fs", "git", "bash"},
			DisallowedToolPresets: []string{"dangerous"},
			Dependencies:          []string{"coder", "planner"},
		},
	})

	handler := agentInfoHandler(registry)
	result, _, err := handler(t.Context(), nil, AgentInfoInput{Name: "full-agent"})

	assert.NoError(t, err)
	require.NotNil(t, result)

	textContent := result.Content[0].(*mcp.TextContent)
	text := textContent.Text

	assert.Contains(t, text, "# Agent: full-agent")
	assert.Contains(t, text, "**Description**: Full featured agent")
	assert.Contains(t, text, "**Model**: claude-opus")
	assert.Contains(t, text, "**Role**: validation")
	assert.Contains(t, text, "**Complexity**: heavy")
	assert.Contains(t, text, "**Color**: red")
	assert.Contains(t, text, "## Skills")
	assert.Contains(t, text, "- go-review")
	assert.Contains(t, text, "- go-test")
	assert.Contains(t, text, "## Tool Presets")
	assert.Contains(t, text, "- fs")
	assert.Contains(t, text, "- git")
	assert.Contains(t, text, "## Disallowed Tool Presets")
	assert.Contains(t, text, "- dangerous")
	assert.Contains(t, text, "## Dependencies")
	assert.Contains(t, text, "- coder")
	assert.Contains(t, text, "- planner")
}

func TestAgentInfoHandler_ReturnsMeta(t *testing.T) {
	t.Parallel()

	registry := populateAgentRegistry(t, []agent.AgentMeta{
		{Name: "coder", Description: "Code agent", Model: "claude-sonnet", Role: "execution"},
	})

	handler := agentInfoHandler(registry)
	result, resp, err := handler(t.Context(), nil, AgentInfoInput{Name: "coder"})

	assert.NoError(t, err)
	assert.NotNil(t, result)

	meta, ok := resp.(agent.AgentMeta)
	require.True(t, ok)
	assert.Equal(t, "coder", meta.Name)
	assert.Equal(t, "Code agent", meta.Description)
	assert.Equal(t, "claude-sonnet", meta.Model)
	assert.Equal(t, "execution", meta.Role)
}

func TestAgentInfoHandler_ErrorMessage(t *testing.T) {
	t.Parallel()

	registry := agent.NewRegistry()

	handler := agentInfoHandler(registry)
	result, _, err := handler(t.Context(), nil, AgentInfoInput{Name: "nonexistent"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent not found")
	assert.Nil(t, result)
}
