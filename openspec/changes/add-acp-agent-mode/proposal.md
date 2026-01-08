# Proposal: Add ACP Agent Mode

## Why

go-ent currently only operates as an MCP server for Claude Code. To enable true multi-agent orchestration where Claude Code (Opus) acts as master and go-ent instances run as parallel async workers, we need go-ent to also operate as an ACP (Agent Client Protocol) agent.

**Key Insight**: Claude Code with Opus 4.5 excels at research, planning, and review (high reasoning). For bulk implementation tasks, spawning multiple go-ent workers via ACP with cheaper models (Haiku/Sonnet) provides:
- 2-3x faster execution through parallelization
- 70-90% cost reduction for implementation tasks
- Isolated context windows (no context bloat in orchestrator)

Inspired by:
- [Agent Client Protocol](https://github.com/agentclientprotocol/agent-client-protocol) - JSON-RPC 2.0 standard for editor-agent communication
- [Claude Agent SDK](https://www.anthropic.com/engineering/building-agents-with-the-claude-agent-sdk) - Subagent orchestration patterns
- [Oh-My-OpenCode](https://github.com/code-yeongyu/oh-my-opencode) - Multi-agent Sisyphus model
- [Claude-Flow](https://github.com/ruvnet/claude-flow) - Swarm coordination

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│              CLAUDE CODE (Master Orchestrator)                       │
│                    Opus 4.5 + Agent SDK                              │
├─────────────────────────────────────────────────────────────────────┤
│  Research (Opus)  │  Orchestration (Opus)  │  Review (Opus)         │
│  - Explore        │  - Task routing        │  - Quality gate        │
│  - Analyze        │  - Worker spawn        │  - Approval            │
└───────────────────┴───────────┬────────────┴────────────────────────┘
                                │
                      ACP (JSON-RPC 2.0 / stdio)
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│ go-ent Worker │       │ go-ent Worker │       │ go-ent Worker │
│ (ACP Agent)   │       │ (ACP Agent)   │       │ (ACP Agent)   │
│ Haiku/Sonnet  │       │ Haiku/Sonnet  │       │ Haiku/Sonnet  │
│ Task: T001    │       │ Task: T002    │       │ Task: T003    │
└───────────────┘       └───────────────┘       └───────────────┘
```

## What Changes

### 1. Dual-Mode Operation
- **MCP Server Mode** (existing): For Claude Code direct integration
- **ACP Agent Mode** (new): For async worker execution

### 2. ACP Protocol Implementation
- JSON-RPC 2.0 over stdio transport
- Initialize/authenticate handshake
- Session management with streaming
- Permission flow for tool execution
- File operations with context mentions

### 3. Worker Process Management
- Spawn go-ent as subprocess via ACP
- Pass task context and constraints
- Stream progress updates
- Collect results and errors
- Graceful termination

### 4. Model Tiering Integration
- Default to Haiku for simple tasks
- Auto-escalate to Sonnet for complex tasks
- Explicit model override capability
- Cost tracking per worker

## Impact

- Affected specs: acp-protocol (new capability)
- Affected code: cmd/acp/, internal/acp/, internal/execution/
- Dependencies: Extends add-execution-engine, requires add-background-agents
- Breaking: None (additive feature)

## Key Benefits

1. **True Parallelization**: Spawn 3-10 workers simultaneously
2. **Cost Optimization**: 70-90% savings using Haiku for bulk work
3. **Context Isolation**: Workers don't pollute orchestrator context
4. **Editor Agnostic**: Works with Zed, Neovim, JetBrains, any ACP client
5. **Composable**: Same binary serves MCP, ACP, and CLI modes
