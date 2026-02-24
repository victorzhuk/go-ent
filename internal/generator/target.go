package generator

type Target interface {
	Name() string
	Runtime() string
	Generate(agent *AgentSource, prompts *PromptContent) ([]byte, error)
	OutputPath(agentName string) string
	GenerateSkill(skill *SkillSource) ([]byte, error)
	SkillOutputPath(category, name string) string
}
