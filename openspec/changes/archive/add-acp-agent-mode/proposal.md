# Proposal: Add ACP Proxy Mode for OpenCode Worker Orchestration

**Status**: complete (2026-01-20)
**Started**: 2026-01-20
**Completed**: 2026-01-20

## Why

go-ent currently operates as an MCP server for Claude Code. To enable true multi-agent orchestration where Claude Code (Opus) acts as master and **OpenCode instances** run as parallel workers with different AI backends, go-ent needs to act as an **ACP proxy/bridge**.

**Key Architecture Insight**:
- **Claude Code (Opus 4.5)** = Master orchestrator for research, planning, review
- **go-ent** = ACP proxy that spawns and manages OpenCode workers
- **OpenCode** = Actual workers configured with different AI providers (GLM 4.7, Kimi K2, DeepSeek)

go-ent is NOT a worker itself - it's the orchestration layer between Claude Code and OpenCode.

**Benefits of this architecture**:
- 2-5x faster execution through parallel OpenCode workers
- 80-95% cost reduction using cheap providers (GLM 4.7, Kimi K2) for bulk work
- Provider diversity (avoid rate limits, leverage model strengths)
- Isolated context windows (workers don't bloat orchestrator)
- Leverage OpenCode's existing tooling (LSP, MCP, agents)

Inspired by:
- [OpenCode ACP Support](https://opencode.ai/docs/acp/) - `opencode acp` command for editor integration
- [Agent Client Protocol](https://github.com/agentclientprotocol/agent-client-protocol) - JSON-RPC 2.0 standard
- [Claude Agent SDK](https://www.anthropic.com/engineering/building-agents-with-the-claude-agent-sdk) - Subagent orchestration patterns

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│              CLAUDE CODE (Master Orchestrator)                       │
│                    Opus 4.5 + Claude Agent SDK                       │
├─────────────────────────────────────────────────────────────────────┤
│  Research (Opus)  │  Planning (Opus)   │  Review (Opus)              │
│  - Explore        │  - Task breakdown  │  - Quality gate             │
│  - Analyze        │  - Dependency graph│  - Standards check          │
│  - Pattern find   │  - Delegate work   │  - Approval                 │
└───────────────────┴───────────┬────────┴────────────────────────────┘
                                │
                      MCP Protocol (stdio)
              go-ent plugin: agents, skills, commands
              worker_spawn, worker_status, worker_output
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    go-ent (ACP PROXY / BRIDGE)                       │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐   │
│   │ MCP Server      │   │ Worker Manager  │   │ Task Router     │   │
│   │ - Receives cmds │   │ - Spawn workers │   │ - Select provider│   │
│   │ - Exposes tools │   │ - Track status  │   │ - Apply rules   │   │
│   │ - Return results│   │ - Collect output│   │ - Cost optimize │   │
│   └─────────────────┘   └─────────────────┘   └─────────────────┘   │
│                                                                      │
│   Communication Methods:                                             │
│   ├── ACP (stdio): opencode acp → JSON-RPC over stdin/stdout        │
│   ├── CLI (exec):  opencode -p "prompt" -f json → batch mode        │
│   └── API (http):  Direct provider API calls for simple queries     │
│                                                                      │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
        │ ACP (stdio)           │ CLI (subprocess)      │ API (HTTP)
        │ opencode acp          │ opencode run --model  │ provider APIs
        │ Long-running tasks    │ Quick one-shot        │ Simple queries
        ▼                       ▼                       ▼
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│ OpenCode      │       │ OpenCode      │       │ Direct API    │
│ Worker 1      │       │ Worker 2      │       │ Call          │
│               │       │               │       │               │
│ ~/.opencode/  │       │ ~/.opencode/  │       │ Provider:     │
│ config:       │       │ config:       │       │ Anthropic     │
│ GLM 4.7       │       │ Kimi K2       │       │ Haiku         │
│ (Z.AI)        │       │ (Moonshot)    │       │               │
│               │       │               │       │               │
│ Best for:     │       │ Best for:     │       │ Best for:     │
│ Bulk impl     │       │ Large files   │       │ Quick fixes   │
│ Mass edits    │       │ 128K context  │       │ Simple tasks  │
└───────────────┘       └───────────────┘       └───────────────┘
    Task: T001              Task: T002              Task: T003
```

## What Changes

### 1. go-ent as ACP Proxy (Not Worker)
- go-ent does NOT execute tasks itself
- go-ent SPAWNS OpenCode workers and manages their lifecycle
- go-ent ROUTES tasks to appropriate OpenCode configs/providers
- go-ent AGGREGATES results back to Claude Code

### 2. Three Communication Methods with OpenCode

| Method | Command | Use Case | Pros | Cons |
|--------|---------|----------|------|------|
| **ACP** | `opencode acp` | Long-running, streaming | Bidirectional, progress | Process overhead |
| **CLI** | `opencode run --model <provider/model> --prompt "..."` | Quick one-shot tasks | Simple, fast | No streaming |
| **API** | Direct HTTP to provider | Simple queries | Fastest | No OpenCode tools |

### 3. OpenCode Worker Configuration
OpenCode uses a single configuration file with multiple providers:

```json
// ~/.config/opencode/opencode.json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "moonshot": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Moonshot AI",
      "options": {
        "baseURL": "https://api.moonshot.cn/v1"
      },
      "models": {
        "glm-4": {
          "name": "GLM 4.7",
          "provider": "z.ai"
        },
        "kimi-k2": {
          "name": "Kimi K2 (128K context)"
        }
      }
    },
    "deepseek": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "DeepSeek",
      "options": {
        "baseURL": "https://api.deepseek.com/v1"
      },
      "models": {
        "deepseek-coder": {
          "name": "DeepSeek Coder"
        }
      }
    },
    "anthropic": {
      "options": {
        "baseURL": "https://api.anthropic.com/v1"
      }
    }
  },
  "defaultModel": "moonshot/glm-4"
}
```

**Provider/Model Selection:**
- In CLI mode: Use `--model provider/model` flag
- In ACP mode: Model is bound to session at creation
- Cannot switch providers mid-session in ACP mode

### 4. go-ent Worker Manager
```go
type WorkerManager struct {
    workers  map[string]*OpenCodeWorker
    configs  map[string]string  // provider -> config path
}

type OpenCodeWorker struct {
    ID        string
    Provider  string           // "glm", "kimi", "deepseek", "haiku"
    Method    CommunicationMethod  // ACP, CLI, API
    Process   *os.Process      // for ACP/CLI
    Status    WorkerStatus
    Task      *Task
}

func (m *WorkerManager) SpawnACP(provider, model string, task *Task) (*OpenCodeWorker, error) {
    cmd := exec.Command("opencode", "acp")
    cmd.Env = append(os.Environ(),
        fmt.Sprintf("OPENCODE_CONFIG=%s", m.configPath),
    )
    // Provider/model will be set in session/new request
    // ... setup stdin/stdout pipes for JSON-RPC
}

func (m *WorkerManager) SpawnCLI(provider, model string, prompt string) (string, error) {
    cmd := exec.Command("opencode", "run",
        "--model", fmt.Sprintf("%s/%s", provider, model),
        "--prompt", prompt,
    )
    cmd.Env = append(os.Environ(),
        fmt.Sprintf("OPENCODE_CONFIG=%s", m.configPath),
    )
    // ... capture output
}
```

### 5. MCP Tools for Claude Code

| Tool | Purpose |
|------|---------|
| `worker_spawn` | Spawn OpenCode worker with provider/model selection |
| `worker_prompt` | Send task to ACP worker |
| `worker_status` | Check worker status |
| `worker_output` | Get streaming output |
| `worker_cancel` | Cancel worker |
| `worker_list` | List active workers |
| `provider_list` | List configured providers |
| `provider_recommend` | Get optimal provider/model for task |

### 6. Task Routing Rules

```yaml
# .goent/routing.yaml
routing_rules:
  # Simple tasks → Direct API or CLI (fast)
  - match: { type: "lint", complexity: "trivial" }
    method: cli
    provider: anthropic
    model: claude-3-haiku

  # Bulk implementation → OpenCode ACP with GLM
  - match: { type: "implement", files: ">3" }
    method: acp
    provider: moonshot
    model: glm-4

  # Large file analysis → OpenCode ACP with Kimi (128K)
  - match: { context_tokens: ">50000" }
    method: acp
    provider: moonshot
    model: kimi-k2

  # Complex refactoring → OpenCode ACP with DeepSeek
  - match: { type: "refactor", complexity: "high" }
    method: acp
    provider: deepseek
    model: deepseek-coder

  # Default fallback
  - match: { default: true }
    method: cli
    provider: moonshot
    model: glm-4
```

## Impact

- Affected specs: acp-proxy (new capability)
- Affected code: internal/proxy/, internal/worker/, cmd/mcp/
- Dependencies: Requires OpenCode installed on system
- Breaking: None (additive feature)

## Key Benefits

1. **Leverage OpenCode**: Use OpenCode's tools, LSP, MCP without reimplementing
2. **Provider Flexibility**: Switch providers via OpenCode config, not go-ent code
3. **Cost Optimization**: 80-95% savings using GLM/Kimi for bulk work
4. **Context Optimization**: Use Kimi K2 for large files (128K context)
5. **Parallel Execution**: Spawn multiple OpenCode workers simultaneously
6. **Clean Separation**: Claude Code (brain) → go-ent (orchestrator) → OpenCode (hands)

## Provider Configuration

```yaml
# .goent/providers.yaml
providers:
  # Anthropic (via Claude Code subagent or direct API)
  haiku:
    method: api  # Direct API call, fastest for simple tasks
    provider: anthropic
    model: claude-3-haiku

  sonnet:
    method: api
    provider: anthropic
    model: claude-3-5-sonnet

  # OpenCode workers with different backends
  glm:
    method: acp
    provider: moonshot  # Provider name in opencode.json
    model: glm-4        # Model name in opencode.json
    best_for: ["bulk", "implementation", "mass-edits"]
    cost: "$0.01/1M tokens"

  kimi:
    method: acp
    provider: moonshot
    model: kimi-k2
    best_for: ["large-context", "file-analysis"]
    context_limit: 128000
    cost: "$0.02/1M tokens"

  deepseek:
    method: acp
    provider: deepseek
    model: deepseek-coder
    best_for: ["refactoring", "code-heavy"]
    cost: "$0.01/1M tokens"

defaults:
  research: opus        # Stays in Claude Code
  planning: opus        # Stays in Claude Code
  review: opus          # Stays in Claude Code
  implementation: glm   # OpenCode worker
  large_context: kimi   # OpenCode worker
  simple_tasks: haiku   # Direct API
```

## Example Workflow

```
User: "Add rate limiting to all API endpoints"

1. CLAUDE CODE (Opus) receives request
   → Research: Explores codebase, finds 15 endpoints
   → Plan: Creates 15 tasks in tasks.md with dependencies

2. CLAUDE CODE delegates to go-ent via MCP:
   worker_spawn(provider="glm", tasks=["T001", "T002", "T003"])
   worker_spawn(provider="glm", tasks=["T004", "T005", "T006"])
   worker_spawn(provider="kimi", tasks=["T007"])  # Large config file

3. go-ent (ACP Proxy) spawns OpenCode workers:
   → OPENCODE_CONFIG=~/.config/opencode/opencode.json opencode acp (Worker 1: GLM)
   → OPENCODE_CONFIG=~/.config/opencode/opencode.json opencode acp (Worker 2: GLM)
   → OPENCODE_CONFIG=~/.config/opencode/opencode.json opencode acp (Worker 3: Kimi)

4. OpenCode workers execute tasks with their configured models
   → Worker 1 (GLM): Implements T001-T003
   → Worker 2 (GLM): Implements T004-T006
   → Worker 3 (Kimi): Analyzes large config, implements T007

5. go-ent collects results, returns to Claude Code

6. CLAUDE CODE (Opus) reviews:
   → Quality check all implementations
   → Request fixes if needed → delegate again
   → Approve and archive
```

## Dependencies and Blockers

### Blocked By

This proposal depends on:
1. **add-execution-engine** - Runtime abstraction for agent execution
2. **add-background-agents** - Async agent spawning infrastructure

These dependencies must be completed before implementation can begin.

### Phase 3 Separated

Dynamic MCP Discovery features have been moved to a separate proposal:
- **add-dynamic-mcp-discovery** - Dynamic MCP server discovery and activation
- Tools: `mcp_find`, `mcp_add`, `mcp_remove`, `mcp_active`
- Docker MCP Gateway integration
- MCP routing rules engine

This separation keeps add-acp-agent-mode focused on core worker orchestration.

## Implementation Notes

### Configuration Corrections (Based on Research)

**OpenCode Configuration:**
- OpenCode uses a single `opencode.json` config file (not per-provider files)
- Provider selection via `defaultModel` in config or `--model` CLI flag
- No `--config` flag exists; use `OPENCODE_CONFIG` environment variable

**Corrected Worker Spawn:**
```go
func (m *WorkerManager) SpawnACP(provider, model string) (*Worker, error) {
    cmd := exec.Command("opencode", "acp")
    cmd.Env = append(os.Environ(),
        fmt.Sprintf("OPENCODE_CONFIG=%s", m.configPath),
    )
    // Provider/model bound at session creation
    return m.startWorker(cmd, provider, model)
}
```

### ACP Protocol Corrections

**Initialization Sequence:**
1. Start `opencode acp` subprocess
2. Send `initialize` request (NOT `acp/initialize`)
3. Send `authenticate` if required
4. Send `session/new` with provider/model selection
5. Send `session/prompt` with tasks

**Method Names:**
- ✅ `initialize` (not `acp/initialize`)
- ✅ `session/new` (required before prompts)
- ✅ `session/prompt` (correct)
- ✅ `session/cancel` (correct)

See `ACP_RESEARCH_FINDINGS.md` for complete protocol verification details.