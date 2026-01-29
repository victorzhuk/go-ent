package generator

// AgentSource represents the unified source format for an agent
type AgentSource struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Model       ModelConfig   `yaml:"model"`
	Skills      []string      `yaml:"skills"`
	Tools       ToolsConfig   `yaml:"tools"`
	Prompts     PromptsConfig `yaml:"prompts"`
	Color       string        `yaml:"color,omitempty"`
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
