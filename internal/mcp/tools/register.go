package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/skill"
)

func Register(s *mcp.Server, skillRegistry *skill.Registry) {
	// OpenSpec tools
	registerInit(s)
	registerList(s)
	registerShow(s)
	registerCRUD(s)
	registerRegistry(s)
	registerWorkflow(s)
	registerValidate(s)
	registerArchive(s)

	// Skill tools
	registerSkillList(s, skillRegistry)
	registerSkillInfo(s, skillRegistry)
	registerSkillValidate(s, skillRegistry)

	// State tools
	registerStateSync(s)
}
