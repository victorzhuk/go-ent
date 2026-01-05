# Proposal: Add CLI Commands

## Overview

Add standalone CLI commands (`goent run`, `goent status`, etc.) for non-MCP usage. Enables goent to work outside of Claude Code environment for automation and CI/CD.

## Rationale

### Problem
GoEnt only works via MCP (Claude Code) - can't use it standalone for scripts, CI/CD, or local development.

### Solution
Add CLI commands that wrap the execution engine and spec management:
```
goent run <task>              # Execute with agent selection
goent status                  # Show execution status
goent agent list/info         # Agent management
goent skill list/info         # Skill management
goent spec init/list/show     # Spec management
goent config show/set/init    # Config management
```

## Key Components

1. `internal/cli/root.go` - Root command and CLI framework
2. `internal/cli/run.go` - Execute tasks
3. `internal/cli/agent.go` - Agent commands
4. `internal/cli/spec.go` - Spec commands
5. `internal/cli/config.go` - Config commands

## Dependencies

- Requires: P0-P4 (execution engine)
- Can develop in parallel with P5 (agent-mcp-tools)

## Success Criteria

- [ ] `goent run <task>` executes with agent selection
- [ ] `goent spec list` works like MCP tool
- [ ] `goent config show` displays current config
- [ ] All commands have `--help` text
