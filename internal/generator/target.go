package generator

// Target represents a code tool target (Claude Code, OpenCode, etc.)
type Target interface {
	// Name returns the target name (claude, opencode)
	Name() string

	// Generate transforms agent source + prompts into target-specific format
	Generate(agent *AgentSource, prompts *PromptContent) ([]byte, error)

	// OutputPath returns the output file path relative to project root
	OutputPath(agentName string) string

	// GenerateSkill transforms skill source into target-specific format
	// For Claude: keeps all fields
	// For OpenCode: strips Claude-specific fields (version, author, etc.)
	GenerateSkill(skill *SkillSource) ([]byte, error)

	// SkillOutputPath returns the output file path for a skill
	SkillOutputPath(category, name string) string
}
