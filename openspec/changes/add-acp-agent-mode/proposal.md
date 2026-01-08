# Proposal: Add ACP Agent Mode with Multi-Provider Workers

## Why

go-ent currently only operates as an MCP server for Claude Code. To enable true multi-agent orchestration where Claude Code (Opus) acts as master orchestrator and heterogeneous workers execute tasks in parallel, we need:

1. **go-ent as ACP Agent**: Run as worker process spawned by orchestrator
2. **Multi-Provider Support**: Workers can use Claude, OpenCode (GLM 4.7, Kimi K2), or other AI backends
3. **Heterogeneous Swarm**: Mix different providers based on task requirements

**Key Insight**: Claude Code with Opus 4.5 excels at research, planning, and review (high reasoning). For bulk implementation tasks, spawning multiple workers with **different AI backends** provides:
- 2-5x faster execution through parallelization
- 80-95% cost reduction using cheaper models (GLM 4.7, Kimi K2) for bulk work
- Provider diversity (avoid rate limits, leverage model strengths)
- Isolated context windows (no context bloat in orchestrator)

Inspired by:
- [Agent Client Protocol](https://github.com/agentclientprotocol/agent-client-protocol) - JSON-RPC 2.0 standard for editor-agent communication
- [Claude Agent SDK](https://www.anthropic.com/engineering/building-agents-with-the-claude-agent-sdk) - Subagent orchestration patterns
- [Oh-My-OpenCode](https://github.com/code-yeongyu/oh-my-opencode) - Multi-agent Sisyphus model with specialized teammates
- [Claude-Flow](https://github.com/ruvnet/claude-flow) - Swarm coordination with 64 specialized agents

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│              CLAUDE CODE (Master Orchestrator)                       │
│                    Opus 4.5 + Agent SDK                              │
├─────────────────────────────────────────────────────────────────────┤
│  Research (Opus)  │  Orchestration (Opus)  │  Review (Opus)         │
│  - Explore        │  - Task routing        │  - Quality gate        │
│  - Analyze        │  - Provider selection  │  - Approval            │
│  - Pattern find   │  - Worker spawn        │  - Standards check     │
└───────────────────┴───────────┬────────────┴────────────────────────┘
                                │
              ┌─────────────────┼─────────────────┐
              │                 │                 │
              │  ACP Protocol   │  Claude Task    │  Direct API
              │  (stdio)        │  (subagent)     │  (HTTP)
              │                 │                 │
    ┌─────────┴─────────┐ ┌─────┴─────┐ ┌────────┴────────┐
    ▼                   ▼ ▼           ▼ ▼                 ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│ OpenCode    │ │ OpenCode    │ │ Claude Code │ │ OpenCode    │
│ Worker      │ │ Worker      │ │ Subagent    │ │ Worker      │
│             │ │             │ │             │ │             │
│ Provider:   │ │ Provider:   │ │ Provider:   │ │ Provider:   │
│ GLM 4.7     │ │ Kimi K2     │ │ Haiku       │ │ DeepSeek    │
│ (Z.AI)      │ │ (Moonshot)  │ │ (Anthropic) │ │             │
│             │ │             │ │             │ │             │
│ Best for:   │ │ Best for:   │ │ Best for:   │ │ Best for:   │
│ Fast impl   │ │ Long ctx    │ │ Simple fix  │ │ Code-heavy  │
│ Bulk tasks  │ │ Large files │ │ Quick tasks │ │ Refactoring │
└─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘
    Task: T001      Task: T002      Task: T003      Task: T004
```

## What Changes

### 1. Multi-Mode Operation
- **MCP Server Mode** (existing): For Claude Code direct integration
- **ACP Agent Mode** (new): Run as worker spawned by orchestrator
- **ACP Server Mode** (new): Spawn and manage heterogeneous workers

### 2. Multi-Provider Backend Support
go-ent workers can use different AI providers:

| Provider | Model | Strength | Cost | Best For |
|----------|-------|----------|------|----------|
| Anthropic | Haiku | Fast, cheap | $ | Simple tasks, linting |
| Anthropic | Sonnet | Balanced | $$ | Standard implementation |
| Z.AI | GLM 4.7 | Fast, bulk | ¢ | Mass file edits |
| Moonshot | Kimi K2 | Long context (128K) | ¢ | Large file analysis |
| DeepSeek | DeepSeek-V3 | Code-focused | ¢ | Complex refactoring |
| Alibaba | Qwen3 | Multilingual | ¢ | i18n, docs |

### 3. ACP Protocol Implementation
- JSON-RPC 2.0 over stdio transport
- Initialize/authenticate handshake with provider config
- Session management with streaming
- Permission flow for tool execution
- File operations with context mentions

### 4. Worker Spawning Interface
Three spawning mechanisms:
1. **ACP (stdio)**: For OpenCode workers - spawn subprocess, communicate via JSON-RPC
2. **Claude Task tool**: For Claude subagents - use existing Task tool with run_in_background
3. **Direct API**: For stateless calls to provider APIs

### 5. Provider-Aware Task Routing
```yaml
routing_rules:
  - match: { type: "lint", files: "<10" }
    provider: anthropic
    model: haiku

  - match: { type: "implement", loc: "<100" }
    provider: z-ai
    model: glm-4.7

  - match: { type: "analyze", context_size: ">50000" }
    provider: moonshot
    model: kimi-k2

  - match: { type: "refactor", complexity: "high" }
    provider: anthropic
    model: sonnet
```

## Impact

- Affected specs: acp-protocol (new capability), execution-engine (provider abstraction)
- Affected code: cmd/acp/, internal/acp/, internal/execution/, internal/provider/
- Dependencies: Extends add-execution-engine, requires add-background-agents
- Breaking: None (additive feature)

## Key Benefits

1. **Heterogeneous Swarm**: Mix providers based on task requirements
2. **Cost Optimization**: 80-95% savings using GLM/Kimi for bulk work
3. **Rate Limit Avoidance**: Distribute across providers
4. **Context Optimization**: Use Kimi K2 for large file analysis (128K context)
5. **Provider Failover**: Auto-switch if one provider is down
6. **Context Isolation**: Workers don't pollute orchestrator context
7. **Editor Agnostic**: Works with Zed, Neovim, JetBrains, any ACP client

## Provider Configuration

```yaml
# .goent/providers.yaml
providers:
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
    models:
      - haiku    # $0.25/1M input, $1.25/1M output
      - sonnet   # $3/1M input, $15/1M output
      - opus     # $15/1M input, $75/1M output

  z-ai:
    api_key: ${ZAI_API_KEY}
    base_url: https://api.z.ai/v1
    models:
      - glm-4.7  # ~$0.01/1M tokens

  moonshot:
    api_key: ${MOONSHOT_API_KEY}
    base_url: https://api.moonshot.cn/v1
    models:
      - kimi-k2  # 128K context, ~$0.02/1M tokens

  deepseek:
    api_key: ${DEEPSEEK_API_KEY}
    models:
      - deepseek-v3

defaults:
  orchestrator: anthropic/opus
  research: anthropic/sonnet
  implementation: z-ai/glm-4.7
  long_context: moonshot/kimi-k2
  review: anthropic/opus
```
