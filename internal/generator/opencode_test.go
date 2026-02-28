package generator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewOpenCodeTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		outputDir string
	}{
		{name: "creates with output dir", outputDir: "/.opencode/agents"},
		{name: "creates with empty output dir", outputDir: ""},
		{name: "creates with relative path", outputDir: ".opencode/agents"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			target := NewOpenCodeTarget(tt.outputDir)
			require.NotNil(t, target)
			assert.Equal(t, tt.outputDir, target.OutputDir)
		})
	}
}

func TestOpenCodeTarget_Name(t *testing.T) {
	t.Parallel()

	target := NewOpenCodeTarget("/output")
	assert.Equal(t, "opencode", target.Name())
}

func TestOpenCodeTarget_Runtime(t *testing.T) {
	t.Parallel()

	target := NewOpenCodeTarget("/output")
	assert.Equal(t, "opencode", target.Runtime())
}

func TestOpenCodeTarget_OutputPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		outputDir  string
		agentName  string
		wantSuffix string
	}{
		{
			name:       "generates path for agent",
			outputDir:  "/.opencode/agents",
			agentName:  "my-agent",
			wantSuffix: "/.opencode/agents/my-agent.md",
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

			target := NewOpenCodeTarget(tt.outputDir)
			path := target.OutputPath(tt.agentName)
			assert.Equal(t, tt.wantSuffix, path)
		})
	}
}

func TestOpenCodeTarget_SkillOutputPath(t *testing.T) {
	t.Parallel()

	target := NewOpenCodeTarget("/.opencode/agents")
	path := target.SkillOutputPath("go", "go-code")

	assert.Contains(t, path, "skills")
	assert.Contains(t, path, "go")
	assert.Contains(t, path, "go-code")
	assert.Contains(t, path, "SKILL.md")
}

func TestOpenCodeTarget_Generate(t *testing.T) {
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
				Model:       ModelConfig{OpenCode: "main"},
				Skills:      []string{"go-code"},
				Tools: ToolsConfig{
					Claude:   ClaudeTools{},
					OpenCode: OpenCodeTools{"read": true, "write": true},
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
				assert.Contains(t, outputStr, "model: main")
				assert.Contains(t, outputStr, "skills:")
				assert.Contains(t, outputStr, "You are a test agent.")
			},
		},
		{
			name: "generates agent with mode and hidden",
			agent: &AgentSource{
				Name:        "subagent",
				Description: "Subagent",
				Model:       ModelConfig{OpenCode: "fast"},
				OpenCode: OpenCodeConfig{
					Mode:   "subagent",
					Hidden: true,
				},
				Tools: ToolsConfig{
					Claude:   ClaudeTools{},
					OpenCode: OpenCodeTools{},
				},
			},
			prompts: &PromptContent{
				Main:   "Subagent prompt.",
				Shared: map[string]string{},
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				outputStr := string(output)
				assert.Contains(t, outputStr, "mode: subagent")
				assert.Contains(t, outputStr, "hidden: true")
			},
		},
		{
			name: "generates agent with tools",
			agent: &AgentSource{
				Name:        "tools-agent",
				Description: "Agent with tools",
				Model:       ModelConfig{OpenCode: "heavy"},
				Tools: ToolsConfig{
					Claude: ClaudeTools{},
					OpenCode: OpenCodeTools{
						"read":  true,
						"write": true,
						"bash":  true,
						"edit":  false,
					},
				},
			},
			prompts: &PromptContent{
				Main:   "Tools prompt.",
				Shared: map[string]string{},
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				outputStr := string(output)
				assert.Contains(t, outputStr, "tools:")
				assert.Contains(t, outputStr, "read: true")
				assert.Contains(t, outputStr, "write: true")
			},
		},
		{
			name: "generates agent with shared prompts",
			agent: &AgentSource{
				Name:        "agent-with-shared",
				Description: "Agent with shared prompts",
				Model:       ModelConfig{OpenCode: "fast"},
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

			target := NewOpenCodeTarget("/output")
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

func TestOpenCodeTarget_GenerateSkill(t *testing.T) {
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
			name: "strips claude-specific fields",
			skill: &SkillSource{
				Name:         "stripped-skill",
				Description:  "Skill to be stripped",
				Version:      "1.0.0",
				Author:       "test-author",
				License:      "MIT",
				Category:     "coding",
				Tags:         []string{"go", "testing"},
				QualityScore: 95,
				Triggers: Triggers{
					Keywords: []string{"test"},
				},
				Content: "Content here.",
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				outputStr := string(output)
				assert.Contains(t, outputStr, "name: stripped-skill")
				assert.Contains(t, outputStr, "description: Skill to be stripped")
				assert.Contains(t, outputStr, "triggers:")
				assert.NotContains(t, outputStr, "version:")
				assert.NotContains(t, outputStr, "author:")
				assert.NotContains(t, outputStr, "license:")
				assert.NotContains(t, outputStr, "category:")
				assert.NotContains(t, outputStr, "tags:")
				assert.NotContains(t, outputStr, "quality_score:")
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

			target := NewOpenCodeTarget("/output")
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

func TestOpenCodeTarget_Generate_ParsesAsYAML(t *testing.T) {
	t.Parallel()

	agent := &AgentSource{
		Name:        "yaml-test",
		Description: "Test YAML parsing",
		Model:       ModelConfig{OpenCode: "main"},
		Skills:      []string{"skill1"},
		Tools: ToolsConfig{
			Claude:   ClaudeTools{},
			OpenCode: OpenCodeTools{"read": true},
		},
	}

	prompts := &PromptContent{
		Main:   "Test prompt",
		Shared: map[string]string{},
	}

	target := NewOpenCodeTarget("/output")
	output, err := target.Generate(agent, prompts)
	require.NoError(t, err)

	outputStr := string(output)
	parts := strings.SplitN(outputStr, "---", 3)
	require.GreaterOrEqual(t, len(parts), 3, "should have frontmatter delimiters")

	frontmatter := strings.TrimSpace(parts[1])
	var fm OpenCodeFrontmatter
	err = yaml.Unmarshal([]byte(frontmatter), &fm)
	require.NoError(t, err, "frontmatter should be valid YAML")

	assert.Equal(t, "yaml-test", fm.Name)
	assert.Equal(t, "Test YAML parsing", fm.Description)
	assert.Equal(t, "main", fm.Model)
	assert.Equal(t, []string{"skill1"}, fm.Skills)
	assert.Equal(t, map[string]bool{"read": true}, fm.Tools)
}

func TestOpenCodeTarget_Skill_ParsesAsYAML(t *testing.T) {
	t.Parallel()

	skill := &SkillSource{
		Name:        "yaml-skill",
		Description: "Test skill YAML",
		Triggers: Triggers{
			Keywords:    []string{"go"},
			FilePattern: "*.go",
		},
		Content: "Skill content",
	}

	target := NewOpenCodeTarget("/output")
	output, err := target.GenerateSkill(skill)
	require.NoError(t, err)

	outputStr := string(output)
	parts := strings.SplitN(outputStr, "---", 3)
	require.GreaterOrEqual(t, len(parts), 3, "should have frontmatter delimiters")

	frontmatter := strings.TrimSpace(parts[1])

	type OpenCodeSkill struct {
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
		Triggers    Triggers `yaml:"triggers"`
	}

	var parsed OpenCodeSkill
	err = yaml.Unmarshal([]byte(frontmatter), &parsed)
	require.NoError(t, err, "frontmatter should be valid YAML")

	assert.Equal(t, "yaml-skill", parsed.Name)
	assert.Equal(t, "Test skill YAML", parsed.Description)
	assert.Equal(t, []string{"go"}, parsed.Triggers.Keywords)
	assert.Equal(t, "*.go", parsed.Triggers.FilePattern)
}
