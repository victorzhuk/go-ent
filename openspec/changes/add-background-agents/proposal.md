# Proposal: Add Background Agent Execution

## Why

Current agent execution is synchronous - the orchestrator blocks until spawned agents complete. For large codebase analysis, security audits, or parallel refactoring, this creates significant bottlenecks and prevents true multi-agent orchestration.

Inspired by:
- [Oh-My-OpenCode](https://github.com/code-yeongyu/oh-my-opencode) - Background task processing with cheaper models
- [Claude-Flow](https://github.com/ruvnet/claude-flow) - Swarm coordination with parallel agents
- [Claude Code Feature Request #9905](https://github.com/anthropics/claude-code/issues/9905) - Background Agent Execution

## What Changes

- **Background Agent Spawning**: MCP tool to spawn agents that run asynchronously
- **Agent Status Monitoring**: Tool to check progress of running background agents
- **Agent Termination**: Tool to kill background agents when needed
- **Agent Registry**: Track all spawned agents with their status and outputs
- **Model Tiering**: Route background research to cheaper/faster models (Haiku)

## Impact

- Affected specs: agent-system (new capability)
- Affected code: internal/agent/, cmd/mcp/
- Dependencies: Requires add-agent-system proposal completion

## Key Benefits

1. **Parallel Research**: Spawn multiple exploration agents simultaneously
2. **Context Efficiency**: Background agents don't consume orchestrator's context
3. **Cost Optimization**: Route background work to cheaper models
4. **Responsive UX**: Long-running analysis doesn't freeze the orchestrator
