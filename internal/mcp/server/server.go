package server

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/mcp/tools"
	"github.com/victorzhuk/go-ent/internal/skill"
	"github.com/victorzhuk/go-ent/internal/spec"
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
	skillRegistry := skill.NewRegistry()
	if skillsPath == "" {
		// Default to pkg/skills relative to executable
		exe, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exe)
			skillsPath = filepath.Join(exeDir, "..", "plugins", "go-ent", "skills")
		} else {
			// Fallback to relative path
			skillsPath = "pkg/skills"
		}
	}

	if err := skillRegistry.Load(skillsPath); err != nil {
		slog.Warn("failed to load skills", "path", skillsPath, "error", err)
	} else {
		slog.Info("loaded skills", "count", len(skillRegistry.All()), "path", skillsPath)
	}

	// Initialize agent registry
	agentRegistry := agent.NewRegistry()
	agentsPath := "pkg/agents/meta"
	if err := agentRegistry.Load(agentsPath); err != nil {
		slog.Warn("failed to load agents", "path", agentsPath, "error", err)
	} else {
		slog.Info("loaded agents", "count", len(agentRegistry.All()), "path", agentsPath)
	}

	// Get current working directory for OpenSpec client
	cwd, err := os.Getwd()
	if err != nil {
		slog.Warn("failed to get working directory, using current dir", "error", err)
		cwd = "."
	}

	// Initialize BoltStore for registry
	store, err := spec.NewBoltStore(cwd)
	if err != nil {
		slog.Warn("failed to initialize BoltStore", "error", err)
		store = nil
	} else {
		slog.Info("initialized BoltStore for OpenSpec registry")
	}

	// Initialize file watcher for BoltStore
	var watcher *spec.Watcher
	if store != nil {
		watcher, err = spec.NewWatcher(store, 0)
		if err != nil {
			slog.Warn("failed to create watcher", "error", err)
			watcher = nil
		} else {
			if err := watcher.Start(cwd); err != nil {
				slog.Warn("failed to start watcher", "error", err)
				watcher = nil
			} else {
				slog.Info("started file watcher for OpenSpec registry")
			}
		}
	}

	// Trigger initial sync
	if store != nil {
		if err := store.RebuildFromMarkdown(cwd); err != nil {
			slog.Warn("failed to perform initial sync", "error", err)
		} else {
			slog.Info("completed initial sync from OpenSpec markdown")
		}
	}

	// Create tool registry
	toolRegistry := tools.NewToolRegistry()

	// Register MCP tools
	tools.Register(s, toolRegistry, skillRegistry, agentRegistry, cwd, store)

	return s
}
