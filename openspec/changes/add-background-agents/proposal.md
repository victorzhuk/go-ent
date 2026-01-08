# Proposal: Add Background Agent Execution

## Why

Current agent execution is synchronous - the orchestrator blocks until spawned agents complete. For large codebase analysis, security audits, or parallel refactoring, this creates significant bottlenecks and prevents true multi-agent orchestration.

Inspired by:
- [Oh-My-OpenCode](https://github.com/code-yeongyu/oh-my-opencode) - Background task processing with cheaper models
- [Claude-Flow](https://github.com/ruvnet/claude-flow) - Swarm coordination with parallel agents
- [Claude Code Feature Request #9905](https://github.com/anthropics/claude-code/issues/9905) - Background Agent Execution

## Architecture Context

This proposal provides **lightweight internal agents** for quick tasks:
- Codebase exploration and analysis
- Pattern finding and research
- Quick fixes and validation

For **heavy implementation work**, use OpenCode workers via add-acp-agent-mode proposal.

```
┌─────────────────────────────────────────────────────────────────┐
│                  Agent Type Comparison                          │
├─────────────────────────────────────────────────────────────────┤
│  Internal Agents (this proposal)  │  OpenCode Workers (ACP)     │
│  go_ent_agent_* tools             │  worker_* tools             │
│  Direct API calls (Haiku/Sonnet)  │  ACP/CLI to OpenCode        │
│  Quick: exploration, analysis     │  Heavy: implementation      │
│  Low overhead, no process         │  Process per worker         │
│  Simple tasks < 5 min             │  Complex tasks > 5 min      │
└─────────────────────────────────────────────────────────────────┘
```

## What Changes

- **Background Agent Spawning**: MCP tool to spawn internal agents that run asynchronously
- **Agent Status Monitoring**: Tool to check progress of running background agents
- **Agent Termination**: Tool to kill background agents when needed
- **Agent Registry**: Track all spawned agents with their status and outputs
- **Model Tiering**: Route internal agents to appropriate models via direct API
  - Exploration/analysis → Haiku (fast, cheap)
  - Complex reasoning → Sonnet (balanced)
  - Critical decisions → Opus (high quality)

## Impact

- Affected specs: agent-system (new capability)
- Affected code: internal/agent/, cmd/mcp/
- Dependencies: Requires add-agent-system proposal completion
- Related: add-acp-agent-mode (for heavy implementation via OpenCode workers)

## Key Benefits

1. **Parallel Research**: Spawn multiple exploration agents simultaneously
2. **Context Efficiency**: Background agents don't consume orchestrator's context
3. **Cost Optimization**: Route background work to cheaper models (Haiku)
4. **Responsive UX**: Long-running analysis doesn't freeze the orchestrator
5. **Lightweight**: No process overhead for simple tasks (vs OpenCode workers)
