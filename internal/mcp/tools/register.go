package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/hooks"
	"github.com/victorzhuk/go-ent/internal/openspec"
	"github.com/victorzhuk/go-ent/internal/skill"
	"github.com/victorzhuk/go-ent/internal/spec"
)

func Register(s *mcp.Server, toolRegistry *ToolRegistry, skillRegistry *skill.Registry, agentRegistry *agent.Registry, cwd string, store *spec.BoltStore, hookRegistry *hooks.Registry) {
	// Skill tools
	registerSkillList(s, toolRegistry, skillRegistry)
	registerSkillInfo(s, toolRegistry, skillRegistry)
	registerSkillValidate(s, toolRegistry, skillRegistry)
	registerSkillMatch(s, toolRegistry, skillRegistry)

	// Generate tool
	registerGenerate(s, toolRegistry)

	// OpenSpec tools (mutation only - use registry tools for queries)
	openspecClient := openspec.New(cwd)
	registerOpenSpecNewChange(s, toolRegistry, openspecClient, hookRegistry)
	registerOpenSpecArchive(s, toolRegistry, openspecClient, hookRegistry)
	registerOpenSpecValidate(s, toolRegistry, openspecClient)
	registerOpenSpecInstructions(s, toolRegistry, openspecClient)

	// Registry tools
	if store != nil {
		registerRegistryListChanges(s, toolRegistry, store)
		registerRegistryListTasks(s, toolRegistry, store)
		registerRegistryGetChange(s, toolRegistry, store)
		registerRegistryStatus(s, toolRegistry, store)
		registerRegistryNextTask(s, toolRegistry, store)
		registerRegistryDeps(s, toolRegistry, store)
		registerRegistryMarkDone(s, toolRegistry, cwd, hookRegistry)
		registerRegistryStartTask(s, toolRegistry, store, hookRegistry)
		registerRegistrySync(s, toolRegistry, store)
	}

	// Agent tools
	registerAgentList(s, toolRegistry, agentRegistry)
	registerAgentInfo(s, toolRegistry, agentRegistry)
	registerAgentGenerate(s, toolRegistry, "pkg/agents/meta")

	// Config tools
	registerConfigShow(s, toolRegistry)
	registerConfigSet(s, toolRegistry)

	// Workspace tools
	registerWorkspaceSpecs(s, toolRegistry)
	registerWorkspaceProjects(s, toolRegistry)

	// Discovery tools
	registerToolList(s, toolRegistry)
}
