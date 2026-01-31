package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/openspec"
	"github.com/victorzhuk/go-ent/internal/skill"
)

func Register(s *mcp.Server, skillRegistry *skill.Registry, agentRegistry *agent.Registry, cwd string) {
	// Skill tools
	registerSkillList(s, skillRegistry)
	registerSkillInfo(s, skillRegistry)
	registerSkillValidate(s, skillRegistry)
	registerSkillMatch(s, skillRegistry)

	// Generate tool
	registerGenerate(s)

	// OpenSpec tools
	openspecClient := openspec.New(cwd)
	registerOpenSpecList(s, openspecClient)
	registerOpenSpecShow(s, openspecClient)
	registerOpenSpecNewChange(s, openspecClient)
	registerOpenSpecArchive(s, openspecClient)
	registerOpenSpecValidate(s, openspecClient)
	registerOpenSpecStatus(s, openspecClient)
	registerOpenSpecInstructions(s, openspecClient)

	// Agent tools
	registerAgentList(s, agentRegistry)
	registerAgentInfo(s, agentRegistry)
	registerAgentGenerate(s, "pkg/agents/meta")

	// Config tools
	registerConfigShow(s)
	registerConfigSet(s)

	// Discovery tools
	registerToolList(s)
}
