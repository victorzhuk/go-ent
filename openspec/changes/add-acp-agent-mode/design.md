# Design: ACP Proxy Architecture

## Context

go-ent needs to act as an **ACP proxy/bridge** between Claude Code (master orchestrator) and OpenCode workers (execution layer).

**Key Roles:**
- **Claude Code (Opus 4.5)** = Brain - research, planning, review
- **go-ent** = Orchestration layer - spawns/manages workers, routes tasks
- **OpenCode** = Hands - actual task execution with various AI backends

go-ent does NOT execute tasks itself. It manages OpenCode workers.

## Goals

- Enable Claude Code to delegate tasks to OpenCode workers
- Support multiple AI providers via OpenCode configurations
- Provide three communication methods: ACP, CLI, API
- Implement intelligent task routing based on characteristics
- Aggregate results from parallel workers

## Non-Goals

- go-ent as a worker (it's a proxy only)
- Reimplementing OpenCode's tools (leverage existing)
- Supporting non-OpenCode workers in v1

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│              CLAUDE CODE (Master Orchestrator)                       │
│                    Opus 4.5 - High Reasoning                         │
└───────────────────────────────┬─────────────────────────────────────┘
                                │ MCP Protocol
                                │ (go-ent plugin)
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    go-ent (ACP PROXY)                                │
│                                                                      │
│   ┌───────────────────────────────────────────────────────────────┐ │
│   │                      MCP Server                                │ │
│   │   Tools: worker_spawn, worker_prompt, worker_status,          │ │
│   │          worker_output, worker_cancel, worker_list,           │ │
│   │          provider_list, provider_recommend                    │ │
│   └───────────────────────────────────────────────────────────────┘ │
│                                │                                     │
│   ┌────────────────┐  ┌────────┴────────┐  ┌────────────────┐       │
│   │ Task Router    │  │ Worker Manager  │  │ Result         │       │
│   │                │  │                 │  │ Aggregator     │       │
│   │ - Rules engine │  │ - Spawn workers │  │                │       │
│   │ - Provider     │  │ - Track status  │  │ - Collect      │       │
│   │   selection    │  │ - Collect output│  │ - Merge        │       │
│   │ - Method       │  │ - Handle errors │  │ - Detect       │       │
│   │   selection    │  │                 │  │   conflicts    │       │
│   └────────────────┘  └────────┬────────┘  └────────────────┘       │
│                                │                                     │
└────────────────────────────────┼─────────────────────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
         │ ACP (stdio)           │ CLI (exec)            │ API (http)
         │ Long-running          │ Quick one-shot        │ Simple queries
         ▼                       ▼                       ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│ opencode acp    │     │ opencode -p     │     │ Direct API      │
│ --config glm    │     │ --config kimi   │     │ Anthropic/      │
│                 │     │                 │     │ OpenAI-compat   │
│ Provider:       │     │ Provider:       │     │                 │
│ GLM 4.7         │     │ Kimi K2         │     │ Provider:       │
│                 │     │ (128K context)  │     │ Haiku           │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

## Communication Methods

### 1. ACP (stdio) - For Long-Running Tasks

```go
type ACPClient struct {
    process  *exec.Cmd
    stdin    io.WriteCloser
    stdout   io.ReadCloser
    decoder  *json.Decoder
    encoder  *json.Encoder
}

func (c *ACPClient) Connect(configPath string) error {
    c.process = exec.Command("opencode", "acp", "--config", configPath)
    c.stdin, _ = c.process.StdinPipe()
    c.stdout, _ = c.process.StdoutPipe()
    c.process.Start()

    c.decoder = json.NewDecoder(c.stdout)
    c.encoder = json.NewEncoder(c.stdin)

    // Initialize handshake
    return c.initialize()
}

func (c *ACPClient) SendPrompt(prompt string) (<-chan Response, error) {
    req := Request{
        JSONRPC: "2.0",
        ID:      uuid.New().String(),
        Method:  "session/prompt",
        Params:  map[string]string{"message": prompt},
    }
    c.encoder.Encode(req)

    // Return channel for streaming responses
    results := make(chan Response)
    go c.streamResponses(results)
    return results, nil
}
```

### 2. CLI (exec) - For Quick One-Shot Tasks

```go
func ExecuteCLI(configPath, prompt string) (string, error) {
    cmd := exec.Command("opencode",
        "-p", prompt,
        "-f", "json",
        "-q",  // quiet mode
        "--config", configPath,
    )

    output, err := cmd.Output()
    if err != nil {
        return "", err
    }

    var result CLIResult
    json.Unmarshal(output, &result)
    return result.Response, nil
}
```

### 3. Direct API - For Simple Queries

```go
type DirectAPI struct {
    client  *http.Client
    baseURL string
    apiKey  string
}

func (d *DirectAPI) Query(prompt string) (string, error) {
    // Direct call to Anthropic/OpenAI-compatible API
    // No OpenCode process involved
}
```

## Task Routing

```go
type Router struct {
    rules []RoutingRule
}

type RoutingRule struct {
    Match    MatchCondition
    Method   string  // "acp", "cli", "api"
    Provider string  // "glm", "kimi", "deepseek", "haiku"
}

func (r *Router) Route(task *Task) (method, provider string) {
    for _, rule := range r.rules {
        if rule.Match.Matches(task) {
            return rule.Method, rule.Provider
        }
    }
    return "cli", "glm"  // default
}
```

**Routing Logic:**

| Task Characteristic | Method | Provider | Rationale |
|---------------------|--------|----------|-----------|
| Simple (lint, format) | cli/api | haiku/glm | Fast, cheap |
| Bulk implementation | acp | glm | Streaming, multiple files |
| Large context (>50K) | acp | kimi | 128K context window |
| Code-heavy refactor | acp | deepseek | Code-optimized model |
| Complex reasoning | STAYS IN CLAUDE CODE | opus | High reasoning |

## Worker Manager

```go
type WorkerManager struct {
    workers    map[string]*Worker
    configs    map[string]ProviderConfig
    pool       *WorkerPool
    aggregator *ResultAggregator
}

type Worker struct {
    ID         string
    Provider   string
    Method     CommunicationMethod
    Status     WorkerStatus
    Task       *Task
    ACPClient  *ACPClient      // for ACP method
    Process    *exec.Cmd       // for CLI method
    StartedAt  time.Time
    Output     strings.Builder
}

func (m *WorkerManager) Spawn(provider string, task *Task) (*Worker, error) {
    config := m.configs[provider]
    method := m.router.SelectMethod(task)

    worker := &Worker{
        ID:       uuid.New().String(),
        Provider: provider,
        Method:   method,
        Task:     task,
    }

    switch method {
    case MethodACP:
        worker.ACPClient = NewACPClient()
        worker.ACPClient.Connect(config.OpenCodeConfig)
    case MethodCLI:
        // Will execute on demand
    case MethodAPI:
        // Will use direct API client
    }

    m.workers[worker.ID] = worker
    return worker, nil
}
```

## Provider Configuration

```yaml
# .goent/providers.yaml
providers:
  glm:
    method: acp
    opencode_config: ~/.opencode-glm.json
    cost_per_1m_tokens: 0.01
    best_for:
      - bulk_implementation
      - mass_file_edits

  kimi:
    method: acp
    opencode_config: ~/.opencode-kimi.json
    context_limit: 128000
    cost_per_1m_tokens: 0.02
    best_for:
      - large_context
      - file_analysis

  deepseek:
    method: acp
    opencode_config: ~/.opencode-deepseek.json
    cost_per_1m_tokens: 0.01
    best_for:
      - refactoring
      - code_heavy

  haiku:
    method: api  # Direct API, no OpenCode
    provider: anthropic
    model: claude-3-haiku
    cost_per_1m_tokens: 0.25
    best_for:
      - simple_tasks
      - quick_fixes
```

## OpenCode Configuration Files

```json
// ~/.opencode-glm.json
{
  "provider": "openai-compatible",
  "model": "glm-4",
  "baseUrl": "https://api.z.ai/v1",
  "apiKey": "${ZAI_API_KEY}"
}

// ~/.opencode-kimi.json
{
  "provider": "openai-compatible",
  "model": "moonshot-v1-128k",
  "baseUrl": "https://api.moonshot.cn/v1",
  "apiKey": "${MOONSHOT_API_KEY}"
}

// ~/.opencode-deepseek.json
{
  "provider": "deepseek",
  "model": "deepseek-coder",
  "apiKey": "${DEEPSEEK_API_KEY}"
}
```

## Decisions

### D1: go-ent as Proxy, Not Worker

**Decision**: go-ent manages OpenCode workers, does not execute tasks itself.

**Rationale**:
- Leverage OpenCode's existing tools (LSP, MCP, agents)
- Avoid reimplementing execution logic
- Clean separation of concerns
- Provider flexibility via OpenCode config

### D2: Three Communication Methods

**Decision**: Support ACP, CLI, and direct API.

**Rationale**:
- ACP: Best for long-running tasks with streaming
- CLI: Simple, fast for one-shot tasks
- API: Fastest for trivial queries without OpenCode overhead

### D3: Multiple OpenCode Configs

**Decision**: Each provider has its own OpenCode config file.

**Rationale**:
- OpenCode config determines AI provider
- Easy to add new providers without code changes
- User can customize per-provider settings

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| OpenCode not installed | Check on startup, provide install instructions |
| Config file errors | Validate configs, provide diagnostics |
| Process leaks | Timeout + cleanup on shutdown |
| Provider rate limits | Track limits, route to alternatives |

## Open Questions

1. Should we support OpenCode's built-in agents (Build, Plan)?
   - **Leaning**: Yes, pass through to leverage their capabilities

2. How to handle OpenCode's MCP servers?
   - **Leaning**: Inherit from user's OpenCode config

3. Should we cache OpenCode processes?
   - **Leaning**: Yes, for ACP connections to avoid startup overhead
