package server

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/marketplace"
	"github.com/victorzhuk/go-ent/internal/mcp/tools"
	"github.com/victorzhuk/go-ent/internal/plugin"
	"github.com/victorzhuk/go-ent/internal/skill"
	"github.com/victorzhuk/go-ent/internal/version"
)

func New() *mcp.Server {
	return NewWithSkillsPath("")
}

func NewWithSkillsPath(skillsPath string) *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "go-ent",
			Version: version.String(),
		},
		nil,
	)

	// Initialize skill registry
	registry := skill.NewRegistry()
	if skillsPath == "" {
		// Default to plugins/go-ent/skills relative to executable
		exe, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exe)
			skillsPath = filepath.Join(exeDir, "..", "plugins", "go-ent", "skills")
		} else {
			// Fallback to relative path
			skillsPath = "plugins/go-ent/skills"
		}
	}

	if err := registry.Load(skillsPath); err != nil {
		slog.Warn("failed to load skills", "path", skillsPath, "error", err)
	} else {
		slog.Info("loaded skills", "count", len(registry.All()), "path", skillsPath)
	}

	pluginsDir := "plugins"
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		pluginsDir = filepath.Join(exeDir, "..", "plugins")
	}

	marketplaceClient := marketplace.NewClient()
	marketplaceSearcher := marketplace.NewSearcher(marketplaceClient)
	registryWrapper := &skillRegistryWrapper{registry: registry}
	pluginManager := plugin.NewManager(pluginsDir, registryWrapper, marketplaceClient)

	if err := pluginManager.Initialize(nil); err != nil {
		slog.Warn("failed to initialize plugin manager", "error", err)
	} else {
		slog.Info("plugin manager initialized", "plugins_dir", pluginsDir)
	}

	tools.Register(s, registry, pluginManager, marketplaceSearcher)

	return s
}

type skillRegistryWrapper struct {
	registry *skill.Registry
}

func (w *skillRegistryWrapper) RegisterSkill(name, path string) error {
	return w.registry.RegisterSkill(name, path)
}

func (w *skillRegistryWrapper) RegisterAgent(name, _ string) error {
	return nil
}

func (w *skillRegistryWrapper) UnregisterSkill(name string) error {
	return w.registry.UnregisterSkill(name)
}

func (w *skillRegistryWrapper) UnregisterAgent(name string) error {
	return nil
}
