package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ClaudeFrontmatter represents Claude Code agent frontmatter
type ClaudeFrontmatter struct {
	Name            string            `yaml:"name"`
	Description     string            `yaml:"description"`
	Model           string            `yaml:"model,omitempty"`
	Skills          []string          `yaml:"skills,omitempty"`
	DisallowedTools []string          `yaml:"disallowedTools,omitempty"`
	Color           string            `yaml:"color,omitempty"`
	ComplexityHints map[string]string `yaml:"complexityHints,omitempty"`
	ModelMapping    map[string]string `yaml:"modelMapping,omitempty"`
}

// ClaudeTarget generates Claude Code agent files
type ClaudeTarget struct {
	OutputDir string // .claude/agents/
}

func NewClaudeTarget(outputDir string) *ClaudeTarget {
	return &ClaudeTarget{OutputDir: outputDir}
}

func (t *ClaudeTarget) Name() string {
	return "claude"
}

func (t *ClaudeTarget) Runtime() string {
	return "claude"
}

func (t *ClaudeTarget) OutputPath(agentName string) string {
	return filepath.Join(t.OutputDir, agentName+".md")
}

func (t *ClaudeTarget) Generate(agent *AgentSource, prompts *PromptContent) ([]byte, error) {
	// Build frontmatter
	fm := ClaudeFrontmatter{
		Name:            agent.Name,
		Description:     agent.Description,
		Model:           agent.Model.Claude,
		Skills:          agent.Skills,
		Color:           agent.Color,
		ComplexityHints: agent.ComplexityHints,
		ModelMapping:    agent.ModelMapping,
	}

	// Add disallowed tools if any
	if len(agent.Tools.Claude.Disallowed) > 0 {
		fm.DisallowedTools = agent.Tools.Claude.Disallowed
	}

	// Marshal frontmatter
	fmData, err := yaml.Marshal(&fm)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}

	// Inline prompts
	content := InlinePrompts(prompts, agent)

	// Build final markdown with frontmatter
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmData)
	sb.WriteString("---\n\n")
	sb.WriteString(content)

	return []byte(sb.String()), nil
}

func (t *ClaudeTarget) SkillOutputPath(category, name string) string {
	return filepath.Join(t.OutputDir, "..", "skills", category, name, "SKILL.md")
}

func (t *ClaudeTarget) GenerateSkill(skill *SkillSource) ([]byte, error) {
	// For Claude: keep all fields (no stripping)
	fmData, err := yaml.Marshal(skill)
	if err != nil {
		return nil, fmt.Errorf("marshal skill frontmatter: %w", err)
	}

	// Build final markdown with frontmatter
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmData)
	sb.WriteString("---\n\n")
	sb.WriteString(skill.Content)

	return []byte(sb.String()), nil
}
