# Tasks: Add Background Agent Execution

## Dependencies
- Requires: add-agent-system (agent infrastructure)
- Related: add-acp-agent-mode (for heavy implementation via OpenCode workers)

## Scope

This proposal covers **lightweight internal agents** that run via direct API calls.
For **heavy implementation work** (>5 min, multi-file), use OpenCode workers (add-acp-agent-mode).

## 1. Core Infrastructure

- [x] 1.1 Create `internal/agent/background/manager.go` - Background agent lifecycle manager ✓
- [x] 1.2 Create `internal/agent/background/registry.go` - Track spawned agents with status ✓ 2026-01-15
- [x] 1.3 Create `internal/agent/background/output.go` - Buffer and stream agent outputs ✓ 2026-01-17
- [x] 1.4 Add background agent configuration to config system ✓ 2026-01-17

## 2. MCP Tools (Internal Agents)

Note: These tools are for lightweight internal agents, NOT OpenCode workers.
For OpenCode workers, see `worker_*` tools in add-acp-agent-mode.

- [x] 2.1 Implement `go_ent_agent_spawn` - Spawn internal agent with task (direct API) ✓ 2026-01-17
- [x] 2.2 Implement `go_ent_agent_status` - Check agent progress and output ✓ 2026-01-17
- [x] 2.3 Implement `go_ent_agent_output` - Retrieve agent output (with optional regex filter) ✓ 2026-01-17
- [x] 2.4 Implement `go_ent_agent_kill` - Terminate running internal agent ✓ 2026-01-17
- [x] 2.5 Implement `go_ent_agent_list` - List all internal agents with status ✓ 2026-01-17

## 3. Model Tiering (Direct API)

Route internal agents to appropriate models via direct provider API:

- [x] 3.1 Add model tier configuration ✓ 2026-01-17
  - exploration/analysis → Haiku (fast, cheap)
  - complex reasoning → Sonnet (balanced)
  - critical decisions → Opus (high quality)
- [x] 3.2 Implement model routing logic based on task type ✓ 2026-01-17
- [x] 3.3 Add override capability for explicit model selection ✓ 2026-01-17

## 4. Integration

- [x] 4.1 Integrate with existing agent selector for role assignment ✓ 2026-01-17
- [x] 4.2 Add cleanup hooks for session termination ✓ 2026-01-17
- [x] 4.3 Add resource limits per background agent ✓ 2026-01-17

## 5. Testing

- [x] 5.1 Unit tests for background manager ✓
- [x] 5.2 Integration tests for MCP tools ✓ 2026-01-17
- [x] 5.3 Test parallel agent spawning and coordination ✓
