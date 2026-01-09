# Proposal: Add ACP Proxy Mode for OpenCode Worker Orchestration

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
        │ opencode acp          │ opencode -p "..."     │ provider APIs
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
| **CLI** | `opencode -p "..." -f json` | Quick one-shot tasks | Simple, fast | No streaming |
| **API** | Direct HTTP to provider | Simple queries | Fastest | No OpenCode tools |

### 3. OpenCode Worker Configuration
Each OpenCode worker can be pre-configured with different providers:

```json
// ~/.opencode-glm.json (for bulk implementation)
{
  "provider": "openai-compatible",
  "model": "glm-4",
  "baseUrl": "https://api.z.ai/v1",
  "apiKey": "${ZAI_API_KEY}"
}

// ~/.opencode-kimi.json (for large context)
{
  "provider": "openai-compatible",
  "model": "moonshot-v1-128k",
  "baseUrl": "https://api.moonshot.cn/v1",
  "apiKey": "${MOONSHOT_API_KEY}"
}

// ~/.opencode-deepseek.json (for code-heavy)
{
  "provider": "openai-compatible",
  "model": "deepseek-coder",
  "apiKey": "${DEEPSEEK_API_KEY}"
}
```

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

func (m *WorkerManager) SpawnACP(provider string, task *Task) (*OpenCodeWorker, error) {
    configPath := m.configs[provider]
    cmd := exec.Command("opencode", "acp", "--config", configPath)
    // ... setup stdin/stdout pipes for JSON-RPC
}

func (m *WorkerManager) SpawnCLI(provider string, prompt string) (string, error) {
    configPath := m.configs[provider]
    cmd := exec.Command("opencode", "-p", prompt, "-f", "json", "--config", configPath)
    // ... capture output
}
```

### 5. MCP Tools for Claude Code

| Tool | Purpose |
|------|---------|
| `worker_spawn` | Spawn OpenCode worker with provider selection |
| `worker_prompt` | Send task to ACP worker |
| `worker_status` | Check worker status |
| `worker_output` | Get streaming output |
| `worker_cancel` | Cancel worker |
| `worker_list` | List active workers |
| `provider_list` | List configured providers |
| `provider_recommend` | Get optimal provider for task |
| **`mcp_find`** | **Search available MCP servers dynamically** |
| **`mcp_add`** | **Add MCP server to current session** |
| **`mcp_remove`** | **Remove MCP server from session** |
| **`mcp_active`** | **List currently active MCP servers** |

### 6. Task Routing Rules

```yaml
# .goent/routing.yaml
routing_rules:
  # Simple tasks → Direct API or CLI (fast)
  - match: { type: "lint", complexity: "trivial" }
    method: cli
    provider: haiku

  # Bulk implementation → OpenCode ACP with GLM
  - match: { type: "implement", files: ">3" }
    method: acp
    provider: glm

  # Large file analysis → OpenCode ACP with Kimi (128K)
  - match: { context_tokens: ">50000" }
    method: acp
    provider: kimi

  # Complex refactoring → OpenCode ACP with DeepSeek
  - match: { type: "refactor", complexity: "high" }
    method: acp
    provider: deepseek

  # Default fallback
  - match: { default: true }
    method: cli
    provider: glm
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
    method: acp  # opencode acp --config ~/.opencode-glm.json
    opencode_config: ~/.opencode-glm.json
    best_for: ["bulk", "implementation", "mass-edits"]
    cost: "$0.01/1M tokens"

  kimi:
    method: acp
    opencode_config: ~/.opencode-kimi.json
    best_for: ["large-context", "file-analysis"]
    context_limit: 128000
    cost: "$0.02/1M tokens"

  deepseek:
    method: acp
    opencode_config: ~/.opencode-deepseek.json
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
   → opencode acp --config ~/.opencode-glm.json (Worker 1)
   → opencode acp --config ~/.opencode-glm.json (Worker 2)
   → opencode acp --config ~/.opencode-kimi.json (Worker 3)

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

## Phase 3 Enhancement: Dynamic MCP Discovery and Docker Gateway

### Dynamic MCP Discovery

Allows agents to discover and activate MCP servers at runtime without hardcoding configurations.

**Problem**: Currently all MCP servers must be pre-configured in Claude Code settings. This creates brittleness and requires manual configuration updates.

**Solution**: Implement dynamic MCP discovery inspired by Docker's dynamic MCP pattern.

**New MCP Tools**:

#### `mcp_find(query: string) -> []MCPServer`
Search for available MCP servers based on capabilities.

```go
// internal/mcp/discovery.go
type MCPServer struct {
    Name         string
    Description  string
    Capabilities []string
    Transport    string  // "stdio" | "sse" | "http"
    Command      string  // Launch command
    Installed    bool
}

func (d *Discovery) Find(query string) []MCPServer {
    // Search Docker Hub MCP Registry
    // Search local .mcp-server-registry/
    // Search system installed servers
}
```

**Example**:
```
mcp_find("database migration") ->
  - mcp-server-postgres (installed: true)
  - mcp-server-mongodb (installed: false)
  - mcp-server-sqlite (installed: true)
```

#### `mcp_add(name: string, config: object) -> success`
Dynamically add an MCP server to the current session.

```go
func (m *Manager) Add(name string, cfg MCPConfig) error {
    // Validate server exists and is safe
    // Launch server process
    // Register tools with go-ent MCP bridge
    // Send tools/list_changed notification
}
```

**Example**:
```
mcp_add("mcp-server-postgres", {
  "connection": "postgresql://localhost/mydb"
})
```

#### `mcp_remove(name: string) -> success`
Remove an MCP server from the current session.

#### `mcp_active() -> []string`
List currently active MCP servers.

### Docker MCP Gateway Integration

Connects to Docker's MCP Gateway to access cloud-hosted MCP servers without local installation.

**Implementation**:
```go
// internal/mcp/gateway.go
type GatewayClient struct {
    apiURL    string
    apiKey    string
    transport *http.Client
}

// Query Docker MCP Registry
func (g *GatewayClient) Search(query string) []RemoteMCP

// Proxy requests to remote MCP server
func (g *GatewayClient) Proxy(server string, tool string, params any) (any, error)
```

**Configuration**:
```yaml
# .go-ent/mcp-gateway.yaml
gateway:
  enabled: true
  api_url: https://mcp-gateway.docker.com/v1
  api_key: ${DOCKER_MCP_KEY}

  # Auto-discovery rules
  discovery:
    - pattern: "docker*"
      source: gateway
    - pattern: "*"
      source: local
```

### Dynamic Tool Selection

Automatically activates MCP servers based on task context.

**Rules Engine**:
```yaml
# .go-ent/mcp-routing.yaml
routing:
  # Database tasks → activate postgres MCP
  - match: { keywords: ["database", "migration", "schema"] }
    mcp: mcp-server-postgres
    auto_activate: true

  # Browser tasks → activate playwright MCP
  - match: { keywords: ["browser", "web", "scrape"] }
    mcp: mcp-server-playwright
    auto_activate: true

  # Cloud tasks → check Docker Gateway first
  - match: { keywords: ["cloud", "deploy", "container"] }
    prefer: gateway
```

**Implementation**:
```go
// internal/mcp/selector.go
type MCPSelector struct {
    rules   []RoutingRule
    gateway *GatewayClient
}

func (s *MCPSelector) SelectForTask(task Task) []string {
    // Match task against rules
    // Return recommended MCP servers
    // Optionally auto-activate
}
```

### Benefits

1. **No Manual Configuration**: Agents discover and activate MCPs as needed
2. **Gateway Access**: Use cloud MCPs without local installation
3. **Context-Aware Loading**: Only load MCPs relevant to current task
4. **Session Isolation**: Different tasks can use different MCP combinations
5. **Reduced Setup Friction**: New users don't need complex MCP configuration

### Example Workflow

```
User: "Analyze database schema and generate migration"

1. Agent receives task, analyzes keywords
2. Agent calls mcp_find("database migration")
3. System returns: mcp-server-postgres (local), mcp-server-prisma (gateway)
4. Agent calls mcp_add("mcp-server-postgres", {...})
5. go-ent launches postgres MCP, registers its tools
6. Agent now has database tools available
7. Task completes, mcp_remove() cleans up
```

### Security Considerations

- **MCP Approval**: User must approve first-time MCP activations
- **Capability Limits**: MCPs can only access resources within their declared capabilities
- **Gateway Authentication**: Require API key for Docker MCP Gateway
- **Resource Quotas**: Limit number of concurrent MCPs per session
