# Execution Engine v2 - Feature Documentation

## Overview

The Execution Engine v2 provides advanced capabilities for long-running, fault-tolerant task execution with state persistence, context management, and sandbox security.

### Key Features

- **Sandbox Execution**: Isolated code execution with resource limits
- **Code-Mode Execution**: Safe JavaScript execution with controlled API surface
- **Context Summarization**: Automatic LLM-based context compression
- **Checkpointing**: Persistent execution state for recovery
- **Interrupt/Resume**: Graceful control over long-running executions
- **Execution ID Tracking**: Unique identifiers for execution lifecycle management

## Table of Contents

1. [Sandbox Execution](#sandbox-execution)
2. [Code-Mode Execution](#code-mode-execution)
3. [Context Summarization](#context-summation)
4. [Checkpointing](#checkpointing)
5. [Interrupt/Resume](#interruptresume)
6. [Configuration](#configuration)
7. [MCP Tools](#mcp-tools)

---

## Sandbox Execution

### Overview

Sandbox execution isolates untrusted code with configurable resource limits and access controls.

### Resource Limits

```go
type ResourceLimits struct {
    MaxMemoryMB int           // Maximum memory in megabytes
    MaxCPUTime  time.Duration // Maximum CPU time
    MaxExecTime time.Duration // Maximum wall-clock time
}
```

### Default Limits

```go
limits := execution.DefaultResourceLimits()
// Results in:
// - MaxMemoryMB: 128 MB
// - MaxCPUTime:  30 seconds
// - MaxExecTime: 60 seconds
```

### Creating a Sandbox

```go
import "github.com/victorzhuk/go-ent/internal/execution"

limits := execution.ResourceLimits{
    MaxMemoryMB: 256,
    MaxCPUTime:  5 * time.Minute,
    MaxExecTime: 10 * time.Minute,
}

sandbox := execution.NewSandbox(limits)
```

### Configuring Access Controls

```go
// Allow specific file paths
sandbox.WithFileAccess(
    "/path/to/project",
    "/tmp/workspace",
)

// Allow specific API calls
sandbox.WithAPIAccess(
    "read_file",
    "write_file",
    "bash",
)
```

### Checking Resource Limits

```go
// Check memory usage
if err := sandbox.CheckMemoryLimit(); err != nil {
    log.Printf("Memory limit exceeded: %v", err)
}

// Check CPU time
if err := sandbox.CheckCPULimit(elapsedCPU); err != nil {
    log.Printf("CPU limit exceeded: %v", err)
}

// Check execution time
if err := sandbox.CheckExecLimit(elapsedWall); err != nil {
    log.Printf("Execution time exceeded: %v", err)
}
```

### Resource Exceeded Errors

```go
var (
    execution.ErrMemoryLimitExceeded  // Memory limit exceeded
    execution.ErrCPULimitExceeded    // CPU time exceeded
    execution.ErrExecLimitExceeded   // Execution time exceeded
)

// Check specific error type
var resourceErr *execution.ResourceExceededError
if errors.As(err, &resourceErr) {
    log.Printf("Resource %s exceeded limit %v",
        resourceErr.Resource,
        resourceErr.Limit,
    )
}
```

### Use Cases

- **Untrusted Code**: Execute user-generated code safely
- **Plugin Execution**: Run third-party plugins with limits
- **Testing**: Isolate test execution with controlled resources
- **Automation**: Safely execute scripts with time limits

---

## Code-Mode Execution

### Overview

Code-mode enables safe JavaScript execution using the goja VM with controlled function exposure.

### Safe API Surface

Only specific functions are exposed to the code execution environment:

- `read_file(path)` - Read file contents
- `write_file(path, content)` - Write file contents
- `bash(command)` - Execute shell command
- `log(message)` - Log messages

### Creating a Code-Mode Runner

```go
import "github.com/victorzhuk/go-ent/internal/execution"

runner := execution.NewCodeModeRunner(logger)
```

### Executing JavaScript Code

```go
req := &execution.Request{
    Task:     "Execute analysis script",
    Agent:    domain.AgentRoleDev,
    Model:    "claude-3.5-sonnet",
    Context:  ctx,
    Metadata: map[string]interface{}{
        "code": `
            const content = read_file("data.json");
            const data = JSON.parse(content);
            const result = analyze(data);
            write_file("output.json", JSON.stringify(result));
            log("Analysis complete");
        `,
    },
}

result, err := runner.Execute(ctx, req)
if err != nil {
    log.Fatalf("Execution failed: %v", err)
}

log.Printf("Output: %s", result.Output)
```

### Code-Mode with Resource Limits

```go
runner := execution.NewCodeModeRunner(logger)

req := &execution.Request{
    Task:    "Execute with limits",
    Context: ctx,
    Metadata: map[string]interface{}{
        "code":     script,
        "limits": execution.ResourceLimits{
            MaxMemoryMB: 64,
            MaxCPUTime:  30 * time.Second,
        },
    },
}
```

### Error Handling

```go
result, err := runner.Execute(ctx, req)
if result != nil && !result.Success {
    log.Printf("Execution failed: %s", result.Error)
    // Check for specific errors
    if strings.Contains(result.Error, "resource limit") {
        log.Printf("Hit resource limit")
    }
}
```

### Use Cases

- **Data Processing**: Transform data with JavaScript
- **Testing**: Run test scripts in isolation
- **Prototyping**: Quick code execution without full setup
- **Automation**: Script-based workflows

---

## Context Summarization

### Overview

Context summarization automatically compresses large execution contexts using LLM to stay within token limits.

### Summarization Thresholds

```go
type SummarizationThreshold struct {
    FileCount     int // Trigger when file count exceeds
    ContextLength int // Trigger when total chars exceed
    TokenCount    int // Trigger when estimated tokens exceed
}
```

### Default Thresholds

```go
threshold := execution.DefaultSummarizationThreshold()
// Results in:
// - FileCount:     50 files
// - ContextLength:  50,000 characters
// - TokenCount:    10,000 tokens
```

### Configuration File

Create `.go-ent/summarization.yaml`:

```yaml
file_count: 100
context_length: 100000
token_count: 20000
model: "claude-3.5-sonnet"
```

### Loading Thresholds

```go
threshold, err := execution.LoadSummarizationThreshold(projectPath)
if err != nil {
    log.Printf("Using default thresholds: %v", err)
    threshold = execution.DefaultSummarizationThreshold()
}
```

### Saving Thresholds

```go
err := execution.SaveSummarizationThreshold(
    projectPath,
    execution.SummarizationThreshold{
        FileCount:     100,
        ContextLength: 100000,
        TokenCount:    20000,
    },
    "claude-3.5-sonnet",
)
```

### Enabling Summarization

```go
cfg := execution.Config{
    Logger:                 logger,
    EnableSummarization:    true,
    SummarizationThreshold:  threshold,
    SummarizationModel:      "claude-3.5-sonnet",
}

engine := execution.New(cfg, selector)
```

### Manual Summarization Trigger

```go
// Force summarization regardless of thresholds
summarized := engine.TriggerSummarization(
    ctx,
    task,
    executionID,
    "claude-3.5-sonnet",
    nil,
    true, // force summarization
)
```

### Automatic Summarization

The engine automatically triggers summarization when any threshold is exceeded:

```go
// During execution, the engine checks:
if len(ctx.Files) > threshold.FileCount {
    // Trigger summarization
}

if totalChars > threshold.ContextLength {
    // Trigger summarization
}

if estimatedTokens > threshold.TokenCount {
    // Trigger summarization
}
```

### Tracking Summarization

```go
// Check if context has been summarized
if task.Context.IsSummarized {
    log.Printf("Context has been summarized %d times",
        task.Context.SummarizationCount)
}

// Access original context if available
if task.Context.OriginalContext != nil {
    original := task.Context.OriginalContext
}
```

### Use Cases

- **Long Conversations**: Multi-turn interactions spanning many files
- **Large Projects**: Context with hundreds of files
- **Extended Workflows**: Long-running agent processes
- **Token Optimization**: Stay within LLM token limits

---

## Checkpointing

### Overview

Checkpointing persists execution state to disk for recovery and inspection.

### Execution State

```go
type ExecutionState struct {
    ID         string        // Unique execution ID (UUID)
    Task       *Task        // Task being executed
    Context    *TaskContext // Execution context
    Result     *Result      // Execution result (if completed)
    Status     string       // pending, running, interrupted, completed, failed
    Agent      domain.AgentRole
    Model      string
    Runtime    domain.Runtime
    Strategy   domain.ExecutionStrategy
    Checksum   string        // SHA-256 checksum
    Version    string        // State version
    CreatedAt  time.Time
    UpdatedAt  time.Time
    StartedAt  time.Time
    CompletedAt time.Time
    Metadata   map[string]string
}
```

### Enabling Auto-Checkpoint

```go
cfg := execution.Config{
    Logger:              logger,
    EnableAutoCheckpoint: true,
    MaxCheckpoints:      10,
    CheckpointAgeLimit:  24 * time.Hour,
}

engine := execution.New(cfg, selector)
```

### Execution Lifecycle

```
1. Create Task
   ↓
2. Engine.Execute(task)
   ↓
3. Create initial checkpoint (status: pending)
   ↓
4. Start execution (status: running)
   ↓
5. Auto-checkpoint after task completion
   ↓
6. Save final checkpoint (status: completed/failed)
   ↓
7. Cleanup old checkpoints
```

### Manual Checkpoint Creation

```go
// Create checkpoint at any point
state, err := engine.CreateManualCheckpoint(ctx, task, "running")
if err != nil {
    log.Fatalf("Failed to create checkpoint: %v", err)
}

log.Printf("Checkpoint created: %s", state.ID)
```

### Loading Execution State

```go
state, err := execution.LoadState(executionID)
if err != nil {
    log.Fatalf("Failed to load state: %v", err)
}

log.Printf("Status: %s", state.Status)
log.Printf("Task: %s", state.Task.Description)
```

### Listing Executions

```go
executionIDs, err := execution.ListExecutions()
if err != nil {
    log.Fatalf("Failed to list executions: %v", err)
}

for _, id := range executionIDs {
    state, err := execution.LoadState(id)
    if err != nil {
        continue
    }
    log.Printf("%s: %s (%s)", id, state.Task.Description, state.Status)
}
```

### Checkpoint Cleanup

```go
// Cleanup is automatic after successful execution
// Can also be triggered manually:
err := engine.CleanupOldCheckpoints()
if err != nil {
    log.Printf("Cleanup failed: %v", err)
}
```

### State Checksum Validation

```go
// Validate checksum
valid := state.ValidateChecksum()
if !valid {
    log.Printf("State checksum invalid - data may be corrupted")
}
```

### Use Cases

- **Recovery**: Resume interrupted executions
- **Debugging**: Inspect execution state
- **Audit Trail**: Track execution history
- **Fault Tolerance**: Survive process restarts

---

## Interrupt/Resume

### Overview

Interrupt/resume provides control over long-running executions with graceful state management.

### Interrupting Execution

#### Via Engine API

```go
ctx := context.Background()

err := engine.Interrupt(ctx, executionID)
if err != nil {
    log.Fatalf("Failed to interrupt: %v", err)
}

log.Printf("Execution interrupted: %s", executionID)
```

#### Via MCP Tool

```json
{
  "execution_id": "abc123-def456-ghi789"
}
```

Call `engine_interrupt` tool with the execution ID.

### Resuming Execution

#### Via Engine API

```go
ctx := context.Background()

result, err := engine.ResumeExecution(ctx, executionID)
if err != nil {
    log.Fatalf("Failed to resume: %v", err)
}

log.Printf("Resumed execution completed: %s", result.Output)
```

#### Via MCP Tool

```json
{
  "execution_id": "abc123-def456-ghi789"
}
```

Call `engine_resume` tool with the execution ID.

### Execution Status Flow

```
pending → running → [interrupt] → interrupted → [resume] → running
                                          ↓
                                    (save checkpoint)

pending → running → [fail] → failed → [resume] → running
                                        ↓
                                  (save checkpoint)
```

### Validation Before Resume

```go
// Validate state before resuming
result, err := engine.ValidateExecutionState(executionID)
if err != nil {
    log.Fatalf("Validation failed: %v", err)
}

if !result.Valid {
    log.Printf("State validation failed: %s", result.Message)
    return
}

if !result.CanResume {
    log.Printf("State cannot be resumed: %s", result.Message)
    return
}
```

### Handling Corrupted State

```go
state, err := execution.LoadState(executionID)
var corruptedErr *execution.CorruptedStateError
if errors.As(err, &corruptedErr) {
    if corruptedErr.CanRecover {
        log.Printf("State corrupted but recoverable: %v", err)
    } else {
        log.Printf("State corrupted - cannot recover: %v", err)
        err := engine.DeleteCorruptedState(executionID)
        if err != nil {
            log.Printf("Failed to delete corrupted state: %v", err)
        }
    }
}
```

### Multiple Interrupt/Resume Cycles

```go
// Execution can be interrupted and resumed multiple times
for i := 0; i < 3; i++ {
    // Start execution
    go func() {
        _, err := engine.Execute(ctx, task)
        if err != nil {
            log.Printf("Execution error: %v", err)
        }
    }()

    // Let it run for a while
    time.Sleep(10 * time.Second)

    // Interrupt
    err := engine.Interrupt(ctx, executionID)
    if err != nil {
        log.Printf("Interrupt failed: %v", err)
        continue
    }

    // Resume
    result, err := engine.ResumeExecution(ctx, executionID)
    if err != nil {
        log.Printf("Resume failed: %v", err)
        continue
    }

    log.Printf("Resume %d completed", i+1)
}
```

### Use Cases

- **Long-Running Tasks**: Control multi-hour executions
- **Debugging**: Pause execution to inspect state
- **Resource Management**: Stop/start based on system load
- **User Control**: Allow users to pause/resume workflows

---

## Configuration

### Engine Configuration

```go
type Config struct {
    PreferredRuntime       domain.Runtime
    IsMCPMode            bool
    Logger               *log/slog.Logger
    EnableSummarization   bool
    SummarizationThreshold execution.SummarizationThreshold
    SummarizationModel   string
    EnableAutoCheckpoint bool
    MaxCheckpoints       int
    CheckpointAgeLimit   time.Duration
}
```

### Full Configuration Example

```go
cfg := execution.Config{
    PreferredRuntime:      domain.RuntimeClaudeCode,
    IsMCPMode:           true,
    Logger:              logger,
    EnableSummarization:  true,
    SummarizationThreshold: execution.SummarizationThreshold{
        FileCount:     100,
        ContextLength: 100000,
        TokenCount:    20000,
    },
    SummarizationModel:    "claude-3.5-sonnet",
    EnableAutoCheckpoint:  true,
    MaxCheckpoints:       10,
    CheckpointAgeLimit:    24 * time.Hour,
}

selector := agent.NewSelector(agent.Config{}, nil)
engine := execution.New(cfg, selector)
```

### Default Values

If not specified, the following defaults are used:

- `PreferredRuntime`: None (auto-select)
- `IsMCPMode`: `false`
- `Logger`: `slog.Default()`
- `EnableSummarization`: `false`
- `SummarizationThreshold`: See [Default Thresholds](#default-thresholds)
- `SummarizationModel`: `"claude-3.5-sonnet"`
- `EnableAutoCheckpoint`: `false`
- `MaxCheckpoints`: `10`
- `CheckpointAgeLimit`: `24 * time.Hour`

### Environment Variables

The engine can be configured via environment variables:

```bash
export GO_ENT_SUMMARIZATION_ENABLED=true
export GO_ENT_SUMMARIZATION_MODEL="claude-3.5-sonnet"
export GO_ENT_CHECKPOINT_ENABLED=true
export GO_ENT_MAX_CHECKPOINTS=10
```

---

## MCP Tools

### engine_interrupt

Interrupts a running execution by ID.

**Input:**
```json
{
  "execution_id": "string"
}
```

**Response:**
```json
{
  "success": true,
  "message": "interrupted successfully",
  "execution_id": "abc123-def456",
  "status": "interrupted",
  "error": ""
}
```

### engine_resume

Resumes a previously interrupted, failed, or pending execution.

**Input:**
```json
{
  "execution_id": "string"
}
```

**Response:**
```json
{
  "success": true,
  "message": "resumed successfully",
  "execution_id": "abc123-def456",
  "status": "running",
  "output": "...",
  "tokens_in": 1000,
  "tokens_out": 500,
  "cost": 0.005,
  "error": ""
}
```

### engine_list

Lists all available execution states.

**Input:**
```json
{}
```

**Response:**
```json
{
  "success": true,
  "executions": [
    {
      "id": "abc123",
      "status": "running",
      "task": "Build and test",
      "created_at": "2026-01-26T10:00:00Z"
    }
  ]
}
```

### engine_find

Finds execution state by ID.

**Input:**
```json
{
  "execution_id": "string"
}
```

**Response:**
```json
{
  "success": true,
  "state": {
    "id": "abc123",
    "status": "running",
    "task": {...},
    "context": {...}
  }
}
```

---

## Best Practices

### 1. Resource Limits

Always set appropriate resource limits for sandbox execution:

```go
limits := execution.ResourceLimits{
    MaxMemoryMB: 256,  // Adjust based on task
    MaxCPUTime:  5 * time.Minute,
    MaxExecTime:  10 * time.Minute,
}
```

### 2. Context Management

Monitor context size and adjust summarization thresholds:

```go
// Check context size before execution
if len(task.Context.Files) > threshold.FileCount {
    log.Printf("Large context: %d files", len(task.Context.Files))
}
```

### 3. Checkpoint Strategy

Enable auto-checkpoint for long-running tasks:

```go
cfg := execution.Config{
    EnableAutoCheckpoint: true,
    MaxCheckpoints:       10,  // Keep recent 10
    CheckpointAgeLimit:    24 * time.Hour,
}
```

### 4. Error Handling

Always handle resource exceeded errors:

```go
if err != nil {
    var resourceErr *execution.ResourceExceededError
    if errors.As(err, &resourceErr) {
        log.Printf("Resource limit hit: %s", resourceErr.Resource)
        return nil, fmt.Errorf("task requires more resources: %w", err)
    }
    return nil, err
}
```

### 5. Interrupt Safety

Handle interruptions gracefully:

```go
select {
case result := <-resultChan:
    return result, nil
case <-ctx.Done():
    log.Printf("Context cancelled, interrupting execution")
    err := engine.Interrupt(ctx, executionID)
    if err != nil {
        log.Printf("Interrupt failed: %v", err)
    }
    return nil, ctx.Err()
}
```

---

## See Also

- [Execution Engine Examples](EXECUTION_ENGINE_EXAMPLES.md)
- [API Reference](EXECUTION_ENGINE_API.md)
- [Troubleshooting Guide](EXECUTION_ENGINE_TROUBLESHOOTING.md)
- [Test Coverage](../internal/execution/...)
