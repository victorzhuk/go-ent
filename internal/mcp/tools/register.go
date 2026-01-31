package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/skill"
)

func Register(s *mcp.Server, skillRegistry *skill.Registry) {
	// Skill tools
	registerSkillList(s, skillRegistry)
	registerSkillInfo(s, skillRegistry)
	registerSkillValidate(s, skillRegistry)

	// Generate tool
	registerGenerate(s)
}
