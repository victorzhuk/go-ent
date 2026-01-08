# Design: ACP Agent Mode Architecture

## Context

go-ent needs to operate in multiple modes:
1. **MCP Server**: Direct integration with Claude Code (current)
2. **ACP Agent**: Async worker for multi-agent orchestration (new)
3. **CLI Tool**: Standalone automation (existing proposal)

This design focuses on ACP Agent mode, enabling Claude Code (Opus 4.5) to orchestrate multiple go-ent workers for parallel task execution.

## Goals

- Enable go-ent to run as ACP-compatible agent
- Support parallel worker spawning from Claude Code
- Implement cost-effective model tiering (Opus → Sonnet → Haiku)
- Maintain context isolation between workers and orchestrator
- Provide streaming progress and results

## Non-Goals

- Replacing MCP server mode
- Building custom orchestration UI
- Supporting non-Claude AI providers in workers

## Architecture

### Multi-Mode Binary

```go
// cmd/main.go
func main() {
    switch os.Args[1] {
    case "mcp":
        runMCPServer()      // Existing: MCP server for Claude Code
    case "acp":
        runACPAgent()       // New: ACP agent for worker mode
    case "run":
        runCLI()            // Existing proposal: CLI mode
    default:
        runMCPServer()      // Default to MCP for backward compat
    }
}
```

### ACP Protocol Layer

```
┌─────────────────────────────────────────────────────────────────────┐
│                        ACP Protocol Stack                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │ Application Layer                                            │   │
│   │ - Task parsing and execution                                 │   │
│   │ - Model selection                                           │   │
│   │ - Result formatting                                         │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │ Session Layer                                                │   │
│   │ - Session state management                                   │   │
│   │ - Message history                                           │   │
│   │ - Permission flow                                           │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │ Protocol Layer                                               │   │
│   │ - JSON-RPC 2.0 message handling                              │   │
│   │ - Request/Response correlation                               │   │
│   │ - Notification streaming                                     │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │ Transport Layer                                              │   │
│   │ - Newline-delimited JSON                                     │   │
│   │ - stdin/stdout                                              │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Message Types

```go
// JSON-RPC 2.0 Messages
type Request struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      interface{}     `json:"id"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      interface{}     `json:"id"`
    Result  json.RawMessage `json:"result,omitempty"`
    Error   *RPCError       `json:"error,omitempty"`
}

type Notification struct {
    JSONRPC string          `json:"jsonrpc"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params,omitempty"`
}

// ACP-specific types
type InitializeParams struct {
    ProtocolVersion int                    `json:"protocolVersion"`
    Capabilities    ClientCapabilities     `json:"capabilities"`
}

type SessionPromptParams struct {
    SessionID string            `json:"sessionId"`
    Message   string            `json:"message"`
    Context   []ContextItem     `json:"context,omitempty"`
}

type ProgressNotification struct {
    SessionID  string  `json:"sessionId"`
    Step       string  `json:"step"`
    Progress   float64 `json:"progress"`  // 0.0 - 1.0
    Message    string  `json:"message,omitempty"`
}
```

### Orchestration Flow

```
Claude Code (Opus)                    go-ent Worker (Haiku)
       │                                     │
       │  ┌──────────────────────────────────┘
       │  │
       │  │  1. Spawn subprocess
       │  │     go-ent acp --model haiku --allowed-tools Read,Write,Edit
       │  │
       ├──┼──────────────────────────────────►│
       │  │                                   │
       │  │  2. Initialize (acp/initialize)   │
       │  │  ─────────────────────────────────►│
       │  │  ◄─────────────────────────────────│
       │  │                                   │
       │  │  3. Create session                │
       │  │  ─────────────────────────────────►│
       │  │  ◄─────────────────────────────────│
       │  │                                   │
       │  │  4. Send prompt                   │
       │  │  "Implement rate limiting config" │
       │  │  ─────────────────────────────────►│
       │  │                                   │
       │  │  5. Progress notifications        │
       │  │  ◄─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│
       │  │  "Reading existing config..."     │
       │  │  ◄─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│
       │  │  "Writing rate_limit.go..."       │
       │  │  ◄─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│
       │  │                                   │
       │  │  6. Permission request            │
       │  │  (for Write operation)            │
       │  │  ◄─────────────────────────────────│
       │  │  ─────────────────────────────────►│
       │  │  (approve)                        │
       │  │                                   │
       │  │  7. Final result                  │
       │  │  ◄─────────────────────────────────│
       │  │                                   │
       │  │  8. Close session                 │
       │  │  ─────────────────────────────────►│
       │  │                                   │
       └──┴──────────────────────────────────►│ (terminate)
```

### Worker Pool Management

```go
type WorkerPool struct {
    maxWorkers  int
    workers     map[string]*Worker
    pending     chan *Task
    results     chan *Result
    mu          sync.RWMutex
}

type Worker struct {
    ID          string
    Process     *os.Process
    Stdin       io.WriteCloser
    Stdout      io.ReadCloser
    Status      WorkerStatus
    CurrentTask *Task
    StartedAt   time.Time
}

func (p *WorkerPool) Spawn(task *Task, model string) (*Worker, error) {
    // 1. Check pool capacity
    if len(p.workers) >= p.maxWorkers {
        return nil, ErrPoolFull
    }

    // 2. Start go-ent subprocess in ACP mode
    cmd := exec.Command("go-ent", "acp",
        "--model", model,
        "--allowed-tools", strings.Join(task.AllowedTools, ","),
        "--timeout", task.Timeout.String(),
    )

    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()
    cmd.Start()

    // 3. Initialize ACP handshake
    worker := &Worker{
        ID:      uuid.New().String(),
        Process: cmd.Process,
        Stdin:   stdin,
        Stdout:  stdout,
        Status:  WorkerStatusReady,
    }

    // 4. Register worker
    p.mu.Lock()
    p.workers[worker.ID] = worker
    p.mu.Unlock()

    return worker, nil
}
```

### Model Tiering Decision Tree

```go
func SelectModel(task *Task) string {
    // 1. Explicit override takes precedence
    if task.ModelOverride != "" {
        return task.ModelOverride
    }

    // 2. Analyze task complexity
    complexity := AnalyzeComplexity(task)

    switch complexity {
    case ComplexityTrivial:
        // Single file, simple change
        return "haiku"

    case ComplexitySimple:
        // 1-3 files, straightforward logic
        return "haiku"

    case ComplexityModerate:
        // Multiple files, some coordination
        return "sonnet"

    case ComplexityComplex:
        // Cross-cutting concerns, architecture
        return "sonnet"

    case ComplexityArchitectural:
        // Should not be in worker, escalate to orchestrator
        return "opus"
    }

    return "haiku" // Default to cheapest
}

func AnalyzeComplexity(task *Task) Complexity {
    score := 0

    // File count factor
    if len(task.AffectedFiles) > 5 {
        score += 2
    } else if len(task.AffectedFiles) > 2 {
        score += 1
    }

    // LOC estimate factor
    if task.EstimatedLOC > 200 {
        score += 2
    } else if task.EstimatedLOC > 50 {
        score += 1
    }

    // Dependency factor
    if len(task.Dependencies) > 3 {
        score += 1
    }

    // Pattern factor (testing, refactoring are simpler)
    if task.Type == TaskTypeTest || task.Type == TaskTypeLint {
        score -= 1
    }

    switch {
    case score <= 0:
        return ComplexityTrivial
    case score <= 2:
        return ComplexitySimple
    case score <= 4:
        return ComplexityModerate
    default:
        return ComplexityComplex
    }
}
```

## Decisions

### D1: Stdio over WebSocket

**Decision**: Use stdio transport instead of WebSocket.

**Rationale**:
- ACP spec mandates stdio for subprocess communication
- Simpler deployment (no network config)
- Natural process lifecycle management
- Aligns with MCP and LSP patterns

### D2: Single Binary, Multiple Modes

**Decision**: Keep single `go-ent` binary with mode selection via subcommand.

**Rationale**:
- Simpler distribution
- Shared code between modes
- Consistent versioning
- Easier testing

### D3: Permission Flow for Sensitive Operations

**Decision**: Require permission requests for Write/Bash, not for Read/Grep/Glob.

**Rationale**:
- Read operations are safe
- Write/Bash can modify state
- Matches Claude Code's permission model
- Prevents accidental damage

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Worker process leaks | Implement timeout + parent process monitoring |
| Stdio buffer overflow | Use newline-delimited JSON, chunked writes |
| Model cost explosion | Per-worker cost limits, pool-level budget |
| Context divergence | Workers share OpenSpec registry state |

## Open Questions

1. Should workers share memory/context database, or is it per-worker?
   - **Leaning**: Shared filesystem, per-worker SQLite for session memory

2. How to handle long-running workers (>5 min)?
   - **Leaning**: Heartbeat mechanism, configurable timeout

3. Should we support remote workers (over network)?
   - **Leaning**: Not in v1, but design should not preclude it
