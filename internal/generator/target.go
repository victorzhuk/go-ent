package generator

// Target represents a code tool target (Claude Code, OpenCode, etc.)
type Target interface {
	// Name returns the target name (claude, opencode)
	Name() string

	// Generate transforms agent source + prompts into target-specific format
	Generate(agent *AgentSource, prompts *PromptContent) ([]byte, error)

	// OutputPath returns the output file path relative to project root
	OutputPath(agentName string) string
}
