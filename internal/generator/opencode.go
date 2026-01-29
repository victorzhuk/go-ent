package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// OpenCodeFrontmatter represents OpenCode agent frontmatter
type OpenCodeFrontmatter struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Model       string          `yaml:"model,omitempty"`
	Skills      []string        `yaml:"skills,omitempty"`
	Tools       map[string]bool `yaml:"tools,omitempty"`
}

// OpenCodeTarget generates OpenCode agent files
type OpenCodeTarget struct {
	OutputDir string // .opencode/agents/
}

func NewOpenCodeTarget(outputDir string) *OpenCodeTarget {
	return &OpenCodeTarget{OutputDir: outputDir}
}

func (t *OpenCodeTarget) Name() string {
	return "opencode"
}

func (t *OpenCodeTarget) OutputPath(agentName string) string {
	return filepath.Join(t.OutputDir, agentName+".md")
}

func (t *OpenCodeTarget) Generate(agent *AgentSource, prompts *PromptContent) ([]byte, error) {
	// Build frontmatter
	fm := OpenCodeFrontmatter{
		Name:        agent.Name,
		Description: agent.Description,
		Model:       agent.Model.OpenCode,
		Skills:      agent.Skills,
		Tools:       agent.Tools.OpenCode,
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
