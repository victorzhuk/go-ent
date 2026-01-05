# Proposal: Add Execution Engine

## Overview

Implement the execution engine that runs agents on different runtimes: Claude Code (MCP), OpenCode (native), and CLI (standalone). Supports Single, Multi-agent, and Parallel execution strategies.

## Rationale

### Problem
No runtime abstraction - can't execute tasks via OpenCode or CLI, only MCP.

### Solution
- **Runner interface**: Abstract runtime execution (Claude Code, OpenCode, CLI)
- **Execution strategies**: Single, Multi (conversation), Parallel (with dependency graph)
- **Budget tracking**: Monitor and enforce spending limits
- **Result aggregation**: Collect outputs from parallel executions

## Key Components

1. `internal/execution/engine.go` - Main orchestration engine
2. `internal/execution/claude.go` - Claude Code MCP runner
3. `internal/execution/opencode.go` - **OpenCode native runner (required for v3.0)**
4. `internal/execution/cli.go` - CLI standalone runner
5. `internal/execution/strategy.go` - Execution strategy implementations

## Dependencies

- Requires: P0-P3 (all foundation)
- Blocks: P5 (agent-mcp-tools), P6 (cli-commands)

## Success Criteria

- [ ] Claude Code runner executes via MCP
- [ ] OpenCode runner executes via native API
- [ ] CLI runner executes standalone
- [ ] Parallel execution with dependency graph works
- [ ] Budget tracking prevents overspending
