package tools

import (
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/marketplace"
	"github.com/victorzhuk/go-ent/internal/plugin"
	"github.com/victorzhuk/go-ent/internal/skill"
)

func Register(s *mcp.Server, skillRegistry *skill.Registry, pluginManager *plugin.Manager, marketplaceSearcher *marketplace.Searcher) {
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
	registerEngineExecute(s, skillRegistry)
	registerEngineStatus(s)
	registerEngineBudget(s)
	registerEngineInterrupt(s)
	registerASTParse(s)
	registerASTQuery(s)
	registerASTRename(s)

	// Register plugin tools
	if pluginManager != nil {
		registerPluginList(s, pluginManager)
		registerPluginInstall(s, pluginManager)
		registerPluginInfo(s, pluginManager)
	}

	if marketplaceSearcher != nil {
		registerPluginSearch(s, marketplaceSearcher)
	}

	// Register state tools
	registerStateSync(s)
	registerStateShow(s)

	// Register meta tools (tool discovery system)
	registerMetaTools(s, toolRegistry)

	// Build search index
	if err := toolRegistry.BuildIndex(); err != nil {
		slog.Warn("failed to build tool search index", "error", err)
	} else {
		slog.Info("tool discovery initialized", "total_tools", toolRegistry.Count())
	}
}
