package tools

import (
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent/background"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/marketplace"
	"github.com/victorzhuk/go-ent/internal/plugin"
	"github.com/victorzhuk/go-ent/internal/skill"
	"github.com/victorzhuk/go-ent/internal/worker"
)

func Register(s *mcp.Server, skillRegistry *skill.Registry, pluginManager *plugin.Manager, marketplaceSearcher *marketplace.Searcher, backgroundManager *background.Manager, workerManager *worker.WorkerManager, providerConfig *config.ProvidersConfig) {
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
	if backgroundManager != nil {
		registerAgentSpawn(s, backgroundManager)
		registerAgentBgList(s, backgroundManager)
		registerAgentBgStatus(s, backgroundManager)
		registerAgentBgOutput(s, backgroundManager)
		registerAgentBgKill(s, backgroundManager)
	}
	if workerManager != nil {
		registerWorkerSpawn(s, workerManager, providerConfig)
		registerWorkerPrompt(s, workerManager)
		registerWorkerStatus(s, workerManager)
		registerWorkerOutput(s, workerManager)
		registerWorkerCancel(s, workerManager)
		registerWorkerList(s, workerManager)
	}
	if providerConfig != nil {
		registerProviderList(s, providerConfig)
		registerProviderRecommend(s, providerConfig)
	}
	registerSkillList(s, skillRegistry)
	registerSkillInfo(s, skillRegistry)
	registerSkillValidate(s, skillRegistry)
	registerSkillQuality(s, skillRegistry)
	registerRuntimeList(s)
	registerRuntimeStatus(s)
	registerEngineExecute(s, skillRegistry)
	registerEngineStatus(s)
	registerEngineBudget(s)
	registerEngineInterrupt(s)
	registerASTParse(s)
	registerASTQuery(s)
	registerASTRefs(s)
	registerASTRename(s)
	registerASTExtract(s)
	registerASTGenerate(s)

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

	// Register metrics tools
	registerMetricsShow(s)
	registerMetricsSummary(s)
	registerMetricsExport(s)
	registerMetricsReset(s)

	// Build search index
	if err := toolRegistry.BuildIndex(); err != nil {
		slog.Warn("failed to build tool search index", "error", err)
	} else {
		slog.Info("tool discovery initialized", "total_tools", toolRegistry.Count())
	}
}
