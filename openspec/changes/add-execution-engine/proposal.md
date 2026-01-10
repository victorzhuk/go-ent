# Proposal: Add Execution Engine

## Overview

Implement the execution engine that runs agents on different runtimes: Claude Code (MCP), OpenCode (subprocess), and CLI (standalone). Supports Single, Multi-agent, and Parallel execution strategies.

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
3. `internal/execution/opencode.go` - OpenCode subprocess runner (CLI integration)
4. `internal/execution/cli.go` - CLI standalone runner
5. `internal/execution/strategy.go` - Execution strategy implementations
6. `internal/execution/single.go` - Single strategy implementation
7. `internal/execution/multi.go` - Multi-agent strategy implementation
8. `internal/execution/parallel.go` - Parallel strategy with dependency graph
9. `internal/execution/budget.go` - Budget tracking and enforcement
10. `internal/execution/fallback.go` - Runtime fallback resolver
11. `internal/mcp/tools/execution.go` - MCP tools for engine control

## Dependencies

- Requires: P0-P3 (all foundation)
- Blocks: P5 (agent-mcp-tools), P6 (cli-commands)

## Success Criteria

- [x] Claude Code runner executes via MCP
- [x] OpenCode runner executes via subprocess (CLI)
- [x] CLI runner executes standalone
- [x] Single strategy executes tasks sequentially
- [x] Multi strategy executes agent handoff chains (Architect → Developer)
- [x] Parallel execution with dependency graph works
- [x] Budget tracking monitors spending (auto-proceeds with warning in MCP mode)
- [x] MCP tools registered: engine_execute, engine_status, engine_budget, engine_interrupt
- [x] ExecutionHistory tracking in WorkflowState
- [x] Integration tests passing

## Clarified Design Decisions

### Runtime Fallback Strategy
**Same-family fallback**: Runtimes within the same family can substitute for each other:
- **MCP Family**: claude-code ↔ open-code (bidirectional fallback)
- **CLI Family**: cli (isolated, no fallback)

This prevents unsafe cross-runtime failures while maintaining execution reliability.

### Budget Behavior
**In MCP Mode**: Auto-proceed with warning log (cannot prompt user interactively)
**In CLI Mode**: Block execution with prompt for user approval

### OpenCode Integration
OpenCode is a CLI tool (not REST API). Integration uses subprocess pattern:
- Spawns opencode process with `opencode -p "prompt" -f json -q`
- Communicates via stdout
- Parses JSON output for results

## Implementation Status

### Completed (v1)
- ✅ Runner interface and implementations (Claude Code, OpenCode, CLI)
- ✅ Engine with runner/strategy registration
- ✅ Single strategy (sequential execution)
- ✅ Multi strategy (agent handoff: Architect → Developer → Reviewer)
- ✅ Parallel strategy (dependency graph with errgroup)
- ✅ Budget tracking with mode-aware behavior (MCP vs CLI)
- ✅ Fallback resolver (same-family fallback)
- ✅ MCP tools: engine_execute, engine_status, engine_budget, engine_interrupt
- ✅ ExecutionHistory tracking in WorkflowState
- ✅ Integration tests

### Deferred (v2)
- Code-mode JavaScript sandbox for dynamic tool composition
- Tool composition registry for cross-session reuse
- Context summarization for long-running executions
- Resource limits and sandbox enforcement
- Full execution state management for interrupts

## Impact

**Performance**:
- Multi-agent strategy adds coordination overhead but improves quality
- Parallel execution reduces wall-clock time for independent tasks
- Budget tracking adds negligible overhead

**Architecture**:
- Clean separation: runners (how) vs strategies (what)
- Same-family fallback prevents unsafe cross-runtime failures
- Execution history enables cost tracking and workflow analytics