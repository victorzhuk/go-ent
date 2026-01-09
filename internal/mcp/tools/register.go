package tools

import (
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/skill"
)

func Register(s *mcp.Server, skillRegistry *skill.Registry) {
	// Create tool discovery registry
	toolRegistry := NewToolRegistry(s)

	// Register existing tools directly (legacy pattern)
	registerInit(s)
	registerList(s)
	registerShow(s)
	registerCRUD(s)
	registerRegistry(s)
	registerWorkflow(s)
	registerLoop(s)
	registerGenerate(s)
	registerValidate(s)
	registerArchive(s)
	registerListArchetypes(s)
	registerGenerateComponent(s)
	registerGenerateFromSpec(s)
	registerAgentExecute(s, skillRegistry)
	registerAgentStatus(s)
	registerAgentList(s)
	registerAgentDelegate(s)
	registerSkillList(s, skillRegistry)
	registerSkillInfo(s, skillRegistry)
	registerRuntimeList(s)
	registerRuntimeStatus(s)

	// Register meta tools (tool discovery system)
	registerMetaTools(s, toolRegistry)

	// Build search index
	if err := toolRegistry.BuildIndex(); err != nil {
		slog.Warn("failed to build tool search index", "error", err)
	} else {
		slog.Info("tool discovery initialized", "total_tools", toolRegistry.Count())
	}
}
