# Proposal: Create ACP Client for OpenCode (Phase 3)

## Metadata
- **Change ID:** `simplify-03-acp-client`
- **Status:** Proposed
- **Type:** Feature
- **Priority:** High
- **Affects Specs:** None
- **Part of:** Go-Ent Simplification Series (3/5)
- **Depends On:** `simplify-02-merge-packages`

## Problem

After deleting the complex execution/worker/agent systems in Phase 1, we need a simple way to execute tasks via OpenCode.

OpenCode exposes an ACP (Agent Communication Protocol) server that allows external tools to:
- Execute tasks
- Monitor task status
- Cancel running tasks
- Stream task output
- Execute tasks in parallel

Current approach was over-engineered with 3 runners, complex worker pools, health checks, and hooks. We need a minimal client that just talks to OpenCode's ACP server.

## Proposed Solution

Create a new `internal/acp/` package with 3 files:

### 1. client.go (~150 LOC)
```go
type Client struct {
    baseURL string
    httpClient *http.Client
}

func NewClient(baseURL string) *Client
func (c *Client) Connect(ctx context.Context) error
func (c *Client) Execute(ctx context.Context, task Task) (*Result, error)
func (c *Client) ExecuteParallel(ctx context.Context, tasks []Task) ([]Result, error)
func (c *Client) Status(ctx context.Context, taskID string) (*Status, error)
func (c *Client) Cancel(ctx context.Context, taskID string) error
func (c *Client) StreamOutput(ctx context.Context, taskID string) (<-chan string, error)
```

### 2. protocol.go (~100 LOC)
```go
type Task struct {
    ID          string
    Description string
    Skills      []string
    Context     map[string]interface{}
}

type Result struct {
    TaskID   string
    Status   string
    Output   string
    Error    string
    Duration time.Duration
}

type Status struct {
    TaskID     string
    State      string
    Progress   int
    Output     string
    StartedAt  time.Time
    FinishedAt *time.Time
}
```

### 3. executor.go (~150 LOC)
```go
type Executor struct {
    client *Client
}

func NewExecutor(baseURL string) (*Executor, error)
func (e *Executor) ExecuteTask(ctx context.Context, task *domain.Task) error
func (e *Executor) ExecuteTasks(ctx context.Context, tasks []*domain.Task) error
func (e *Executor) GetTaskStatus(ctx context.Context, taskID string) (string, error)
func (e *Executor) CancelTask(ctx context.Context, taskID string) error
```

**Total: ~400 LOC (vs 6,000+ LOC in old system)**

## ACP Protocol Details

OpenCode ACP server typically runs on `localhost:8080` (or configured port).

**Protocol:** JSON-RPC 2.0 over HTTP/WebSocket

**Key Endpoints:**
- `POST /execute` - Execute a task
- `GET /status/:id` - Get task status
- `POST /cancel/:id` - Cancel task
- `WS /stream/:id` - Stream output (WebSocket)

**Example Execute Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "execute",
  "params": {
    "id": "task-001",
    "description": "Implement feature X",
    "skills": ["go-code"],
    "context": {}
  },
  "id": 1
}
```

**Example Status Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "task_id": "task-001",
    "state": "running",
    "progress": 50,
    "output": "..."
  },
  "id": 2
}
```

## Impact

- **Breaking Changes:** None (new package)
- **API Changes:** None (internal only)
- **Migration Required:** No
- **Testing Required:** Yes
  - Unit tests for protocol marshaling
  - Integration tests with OpenCode
  - Parallel execution tests
  - Cancellation tests

## Risks

- **Low Risk:** New isolated package
- **Dependencies:** Requires OpenCode ACP server running
- **Fallback:** If ACP doesn't work, can fall back to CLI subprocess

## Dependencies

- **Previous Proposal:** `simplify-02-merge-packages` (must complete first)
- **Next Proposal:** `simplify-04-mcp-tools` (depends on this)

## Success Criteria

- [ ] acp/ package created with 3 files
- [ ] Can connect to OpenCode ACP server
- [ ] Can execute single task
- [ ] Can execute multiple tasks in parallel
- [ ] Can monitor task status
- [ ] Can cancel running tasks
- [ ] Can stream task output
- [ ] Integration tests pass
- [ ] `make build` succeeds
- [ ] `make test` passes
