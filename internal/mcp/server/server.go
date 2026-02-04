package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/hooks"
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

	// Initialize hook registry
	hookRegistry, err := hooks.NewRegistry("", hooks.NewExecutor(slog.Default()))
	if err != nil {
		slog.Warn("failed to initialize hook registry", "error", err)
		hookRegistry = nil
	} else {
		slog.Info("initialized hook registry")
	}

	// Add hook middleware if hooks are configured
	if hookRegistry != nil {
		s.AddReceivingMiddleware(createHookMiddleware(hookRegistry))
	}

	// Register MCP tools
	tools.Register(s, toolRegistry, skillRegistry, agentRegistry, cwd, store, hookRegistry)

	return s
}

// createHookMiddleware creates middleware that executes pre/post tool hooks.
func createHookMiddleware(hookRegistry *hooks.Registry) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			// Only intercept tools/call method
			if method != "tools/call" {
				return next(ctx, method, req)
			}

			ctr, ok := req.(*mcp.CallToolRequest)
			if !ok {
				return next(ctx, method, req)
			}

			toolHooks := hookRegistry.GetToolHooks()
			executor := hookRegistry.Executor()

			// Run PreToolUse hooks
			for _, matcher := range toolHooks.PreToolUse {
				if executor.MatchTool(matcher.Matcher, ctr.Params.Name) {
					// Parse tool input for hook context
					var input any
					if ctr.Params.Arguments != nil {
						input = ctr.Params.Arguments
					}

					if err := executor.RunPreToolHooks(ctx, []hooks.HookMatcher{matcher}, ctr.Params.Name, input); err != nil {
						// Pre-hook blocked execution
						return nil, err
					}
				}
			}

			// Execute the tool
			result, err := next(ctx, method, req)

			// Run PostToolUse hooks (don't block on errors)
			for _, matcher := range toolHooks.PostToolUse {
				if executor.MatchTool(matcher.Matcher, ctr.Params.Name) {
					// Parse result for hook context
					var resultData any
					if result != nil {
						if data, jsonErr := json.Marshal(result); jsonErr == nil {
							json.Unmarshal(data, &resultData)
						}
					}

					executor.RunPostToolHooks(ctx, []hooks.HookMatcher{matcher}, ctr.Params.Name, resultData, err)
				}
			}

			return result, err
		}
	}
}
