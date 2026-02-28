package generator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewClaudeTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		outputDir string
	}{
		{name: "creates with output dir", outputDir: "/.claude/agents"},
		{name: "creates with empty output dir", outputDir: ""},
		{name: "creates with relative path", outputDir: ".claude/agents"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			target := NewClaudeTarget(tt.outputDir)
			require.NotNil(t, target)
			assert.Equal(t, tt.outputDir, target.OutputDir)
		})
	}
}

func TestClaudeTarget_Name(t *testing.T) {
	t.Parallel()

	target := NewClaudeTarget("/output")
	assert.Equal(t, "claude", target.Name())
}

func TestClaudeTarget_Runtime(t *testing.T) {
	t.Parallel()

	target := NewClaudeTarget("/output")
	assert.Equal(t, "claude", target.Runtime())
}

func TestClaudeTarget_OutputPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		outputDir  string
		agentName  string
		wantSuffix string
	}{
		{
			name:       "generates path for agent",
			outputDir:  "/.claude/agents",
			agentName:  "my-agent",
			wantSuffix: "/.claude/agents/my-agent.md",
		},
		{
			name:       "generates path for agent with dashes",
			outputDir:  "/output",
			agentName:  "go-code-agent",
			wantSuffix: "/output/go-code-agent.md",
		},
		{
			name:       "handles empty output dir",
			outputDir:  "",
			agentName:  "test",
			wantSuffix: "test.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			target := NewClaudeTarget(tt.outputDir)
			path := target.OutputPath(tt.agentName)
			assert.Equal(t, tt.wantSuffix, path)
		})
	}
}

func TestClaudeTarget_SkillOutputPath(t *testing.T) {
	t.Parallel()

	target := NewClaudeTarget("/.claude/agents")
	path := target.SkillOutputPath("go", "go-code")

	assert.Contains(t, path, "skills")
	assert.Contains(t, path, "go")
	assert.Contains(t, path, "go-code")
	assert.Contains(t, path, "SKILL.md")
}

func TestClaudeTarget_Generate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		agent    *AgentSource
		prompts  *PromptContent
		wantErr  bool
		validate func(t *testing.T, output []byte)
	}{
		{
			name: "generates basic agent",
			agent: &AgentSource{
				Name:        "test-agent",
				Description: "A test agent",
				Model:       ModelConfig{Claude: "sonnet"},
				Skills:      []string{"go-code"},
				Tools: ToolsConfig{
					Claude:   ClaudeTools{Allowed: []string{"Read", "Write"}},
					OpenCode: OpenCodeTools{},
				},
			},
			prompts: &PromptContent{
				Main:   "You are a test agent.",
				Shared: map[string]string{},
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				outputStr := string(output)
				assert.Contains(t, outputStr, "---")
				assert.Contains(t, outputStr, "name: test-agent")
				assert.Contains(t, outputStr, "description: A test agent")
				assert.Contains(t, outputStr, "model: sonnet")
				assert.Contains(t, outputStr, "skills:")
				assert.Contains(t, outputStr, "You are a test agent.")
			},
		},
		{
			name: "generates agent with disallowed tools",
			agent: &AgentSource{
				Name:        "restricted-agent",
				Description: "Restricted agent",
				Model:       ModelConfig{Claude: "opus"},
				Tools: ToolsConfig{
					Claude: ClaudeTools{
						Disallowed: []string{"Bash", "Write"},
					},
					OpenCode: OpenCodeTools{},
				},
			},
			prompts: &PromptContent{
				Main:   "Be careful.",
				Shared: map[string]string{},
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				outputStr := string(output)
				assert.Contains(t, outputStr, "disallowedTools:")
				assert.Contains(t, outputStr, "Bash")
				assert.Contains(t, outputStr, "Write")
			},
		},
		{
			name: "generates agent with all optional fields",
			agent: &AgentSource{
				Name:        "full-agent",
				Description: "Full featured agent",
				Model:       ModelConfig{Claude: "sonnet"},
				Skills:      []string{"skill1", "skill2"},
				Color:       "blue",
				ComplexityHints: map[string]string{
					"planning": "complex",
				},
				ModelMapping: map[string]string{
					"debug": "opus",
				},
				Tools: ToolsConfig{
					Claude:   ClaudeTools{},
					OpenCode: OpenCodeTools{},
				},
			},
			prompts: &PromptContent{
				Main:   "Full agent prompt.",
				Shared: map[string]string{},
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				outputStr := string(output)
				assert.Contains(t, outputStr, "color: blue")
				assert.Contains(t, outputStr, "complexityHints:")
				assert.Contains(t, outputStr, "modelMapping:")
			},
		},
		{
			name: "generates agent with shared prompts",
			agent: &AgentSource{
				Name:        "agent-with-shared",
				Description: "Agent with shared prompts",
				Model:       ModelConfig{Claude: "haiku"},
				Prompts:     PromptsConfig{Shared: []string{"coding", "testing"}},
				Tools: ToolsConfig{
					Claude:   ClaudeTools{},
					OpenCode: OpenCodeTools{},
				},
			},
			prompts: &PromptContent{
				Main: "Main prompt.",
				Shared: map[string]string{
					"coding":  "Coding guidelines here.",
					"testing": "Testing guidelines here.",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				outputStr := string(output)
				assert.Contains(t, outputStr, "Main prompt.")
				assert.Contains(t, outputStr, "## Coding")
				assert.Contains(t, outputStr, "Coding guidelines here.")
				assert.Contains(t, outputStr, "## Testing")
				assert.Contains(t, outputStr, "Testing guidelines here.")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			target := NewClaudeTarget("/output")
			output, err := target.Generate(tt.agent, tt.prompts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output)

			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestClaudeTarget_GenerateSkill(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		skill    *SkillSource
		wantErr  bool
		validate func(t *testing.T, output []byte)
	}{
		{
			name: "generates basic skill",
			skill: &SkillSource{
				Name:        "go-code",
				Description: "Go coding skill",
				Triggers: Triggers{
					Keywords:    []string{"go", "golang"},
					FilePattern: "*.go",
					Weight:      1.0,
				},
				Content: "This is the skill content.",
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				outputStr := string(output)
				assert.Contains(t, outputStr, "---")
				assert.Contains(t, outputStr, "name: go-code")
				assert.Contains(t, outputStr, "description: Go coding skill")
				assert.Contains(t, outputStr, "triggers:")
				assert.Contains(t, outputStr, "This is the skill content.")
			},
		},
		{
			name: "generates skill with all fields",
			skill: &SkillSource{
				Name:         "advanced-skill",
				Description:  "Advanced skill",
				Version:      "1.0.0",
				Author:       "test-author",
				License:      "MIT",
				Category:     "coding",
				Tags:         []string{"go", "testing"},
				QualityScore: 95,
				Triggers: Triggers{
					Keywords: []string{"test"},
				},
				Content: "Advanced content here.",
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				outputStr := string(output)
				assert.Contains(t, outputStr, "version: 1.0.0")
				assert.Contains(t, outputStr, "author: test-author")
				assert.Contains(t, outputStr, "license: MIT")
				assert.Contains(t, outputStr, "category: coding")
				assert.Contains(t, outputStr, "tags:")
				assert.Contains(t, outputStr, "quality_score: 95")
			},
		},
		{
			name: "generates skill with empty content",
			skill: &SkillSource{
				Name:        "empty-skill",
				Description: "Empty skill",
				Content:     "",
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				assert.Contains(t, string(output), "name: empty-skill")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			target := NewClaudeTarget("/output")
			output, err := target.GenerateSkill(tt.skill)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output)

			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestClaudeTarget_Generate_ParsesAsYAML(t *testing.T) {
	t.Parallel()

	agent := &AgentSource{
		Name:        "yaml-test",
		Description: "Test YAML parsing",
		Model:       ModelConfig{Claude: "sonnet"},
		Skills:      []string{"skill1"},
		Tools: ToolsConfig{
			Claude:   ClaudeTools{},
			OpenCode: OpenCodeTools{},
		},
	}

	prompts := &PromptContent{
		Main:   "Test prompt",
		Shared: map[string]string{},
	}

	target := NewClaudeTarget("/output")
	output, err := target.Generate(agent, prompts)
	require.NoError(t, err)

	outputStr := string(output)
	parts := strings.SplitN(outputStr, "---", 3)
	require.GreaterOrEqual(t, len(parts), 3, "should have frontmatter delimiters")

	frontmatter := strings.TrimSpace(parts[1])
	var fm ClaudeFrontmatter
	err = yaml.Unmarshal([]byte(frontmatter), &fm)
	require.NoError(t, err, "frontmatter should be valid YAML")

	assert.Equal(t, "yaml-test", fm.Name)
	assert.Equal(t, "Test YAML parsing", fm.Description)
	assert.Equal(t, "sonnet", fm.Model)
	assert.Equal(t, []string{"skill1"}, fm.Skills)
}
