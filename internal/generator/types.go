package generator

// AgentSource represents the unified source format for an agent
type AgentSource struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Model       ModelConfig    `yaml:"model"`
	Skills      []string       `yaml:"skills"`
	Tools       ToolsConfig    `yaml:"tools"`
	Prompts     PromptsConfig  `yaml:"prompts"`
	OpenCode    OpenCodeConfig `yaml:"opencode,omitempty"`
	Color       string         `yaml:"color,omitempty"`
}

// ModelConfig maps model aliases to tool-specific model IDs
type ModelConfig struct {
	Claude   string `yaml:"claude"`   // haiku, sonnet, opus, inherit
	OpenCode string `yaml:"opencode"` // anthropic/claude-*
}

// ToolsConfig defines tool access per target
type ToolsConfig struct {
	Claude   ClaudeTools   `yaml:"claude"`
	OpenCode OpenCodeTools `yaml:"opencode"`
}

// ClaudeTools defines allowed/disallowed tools for Claude Code
type ClaudeTools struct {
	Allowed    []string `yaml:"allowed"`
	Disallowed []string `yaml:"disallowed"`
}

// OpenCodeTools defines tool access as boolean flags
type OpenCodeTools map[string]bool

// PromptsConfig defines shared and main prompt references
type PromptsConfig struct {
	Shared []string `yaml:"shared"` // References to src/prompts/*.md
	Main   string   `yaml:"main"`   // Agent-specific prompt name
}

// PromptContent holds loaded prompt content
type PromptContent struct {
	Shared map[string]string // prompt name -> content
	Main   string            // agent-specific content
}

// OpenCodeConfig holds OpenCode-specific agent configuration
type OpenCodeConfig struct {
	Mode   string `yaml:"mode,omitempty"`   // primary, subagent, all
	Hidden bool   `yaml:"hidden,omitempty"` // hide from UI
}

// SkillSource represents the unified source format for a skill
// Skills are single markdown files with YAML frontmatter
type SkillSource struct {
	// Standard fields (kept for all tools)
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Triggers    Triggers `yaml:"triggers"`

	// Claude-specific fields (stripped for OpenCode)
	Version       string            `yaml:"version,omitempty"`
	Author        string            `yaml:"author,omitempty"`
	License       string            `yaml:"license,omitempty"`
	Compatibility map[string]string `yaml:"compatibility,omitempty"`
	Tags          []string          `yaml:"tags,omitempty"`
	QualityScore  int               `yaml:"quality_score,omitempty"`
	Category      string            `yaml:"category,omitempty"`

	// Content (after frontmatter)
	Content string `yaml:"-"` // Not in YAML, stored separately
}

// Triggers defines skill activation triggers
type Triggers struct {
	Keywords    []string `yaml:"keywords,omitempty"`
	FilePattern string   `yaml:"file_pattern,omitempty"`
	Weight      float64  `yaml:"weight,omitempty"`
}
