package generator

type AgentGenerator interface {
	Name() string
	Runtime() string
	Generate(agent *AgentSource, prompts *PromptContent) ([]byte, error)
	OutputPath(agentName string) string
}

type SkillGenerator interface {
	GenerateSkill(skill *SkillSource) ([]byte, error)
	SkillOutputPath(category, name string) string
}

type Target interface {
	AgentGenerator
	SkillGenerator
}
