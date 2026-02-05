# Execution Engine v2 - API Reference

This document provides detailed API reference for the execution engine v2 components.

## Table of Contents

1. [Engine API](#engine-api)
2. [ExecutionState API](#executionstate-api)
3. [Context Management API](#context-management-api)
4. [Sandbox API](#sandbox-api)
5. [Storage API](#storage-api)
6. [MCP Tools API](#mcp-tools-api)
7. [Types and Constants](#types-and-constants)
8. [Error Handling](#error-handling)

---

## Engine API

### New

Creates a new execution engine with the given configuration.

```go
func New(cfg Config, selector *agent.Selector) *Engine
```

**Parameters:**
- `cfg` - Engine configuration (see Config type)
- `selector` - Agent selector for runtime/agent/model selection

**Returns:**
- `*Engine` - New engine instance

**Example:**
```go
cfg := execution.Config{
    Logger:              logger,
    PreferredRuntime:      domain.RuntimeClaudeCode,
    EnableSummarization:  true,
}
selector := agent.NewSelector(agent.Config{}, nil)
engine := execution.New(cfg, selector)
```

### Config

Engine configuration structure.

```go
type Config struct {
    PreferredRuntime      domain.Runtime
    IsMCPMode           bool
    Logger              *log/slog.Logger
    EnableSummarization bool
    SummarizationThreshold execution.SummarizationThreshold
    SummarizationModel   string
    EnableAutoCheckpoint bool
    MaxCheckpoints       int
    CheckpointAgeLimit   time.Duration
}
```

**Fields:**
- `PreferredRuntime` - Default runtime to use (auto-select if empty)
- `IsMCPMode` - If true, budget behaves differently for MCP
- `Logger` - Logger for execution events (default: slog.Default())
- `EnableSummarization` - Enable automatic context summarization
- `SummarizationThreshold` - Thresholds for triggering summarization
- `SummarizationModel` - LLM model for summarization (default: "claude-3.5-sonnet")
- `EnableAutoCheckpoint` - Enable automatic checkpointing after task completion
- `MaxCheckpoints` - Maximum number of checkpoints to keep (default: 10)
- `CheckpointAgeLimit` - Maximum age for checkpoints (default: 24h)

### Execute

Runs a task with automatic runner and strategy selection.

```go
func (e *Engine) Execute(ctx context.Context, task *Task) (*Result, error)
```

**Parameters:**
- `ctx` - Context for cancellation/timeout
- `task` - Task to execute

**Returns:**
- `*Result` - Execution result
- `error` - Error if execution failed

**Behavior:**
1. Creates execution state with unique ID
2. Saves initial checkpoint (if enabled)
3. Selects appropriate strategy
4. Executes task
5. Saves final checkpoint
6. Cleans up old checkpoints

**Example:**
```go
task := &execution.Task{
    Description: "Implement feature X",
    Type:        domain.TaskTypeDev,
    Context:     execution.NewTaskContext("/path/to/project"),
}
result, err := engine.Execute(ctx, task)
if err != nil {
    log.Fatal("Execution failed:", err)
}
log.Printf("Success: %v, Output: %s", result.Success, result.Output)
```

### ExecuteWithRunner

Runs a task using a specific runner.

```go
func (e *Engine) ExecuteWithRunner(ctx context.Context, runtime domain.Runtime, task *Task) (*Result, error)
```

**Parameters:**
- `ctx` - Context for cancellation/timeout
- `runtime` - Specific runtime to use
- `task` - Task to execute

**Returns:**
- `*Result` - Execution result
- `error` - Error if execution failed

**Example:**
```go
result, err := engine.ExecuteWithRunner(
    ctx,
    domain.RuntimeClaudeCode,
    task,
)
```

### ResumeExecution

Resumes execution from a saved state.

```go
func (e *Engine) ResumeExecution(ctx context.Context, executionID string) (*Result, error)
```

**Parameters:**
- `ctx` - Context for cancellation/timeout
- `executionID` - ID of execution to resume

**Returns:**
- `*Result` - Execution result
- `error` - Error if resume failed

**Behavior:**
1. Loads execution state
2. Validates version compatibility
3. Validates state integrity
4. Validates environment compatibility
5. Resumes execution from checkpoint

**Errors:**
- `ErrExecutionNotFound` - Execution state not found
- `ErrVersionIncompatible` - State version incompatible
- `ErrStateInvalid` - State validation failed
- `ErrRuntimeUnavailable` - Required runtime not available

**Example:**
```go
result, err := engine.ResumeExecution(ctx, "abc123-def456-ghi789")
if err != nil {
    log.Fatal("Resume failed:", err)
}
```

### Interrupt

Interrupts a running execution.

```go
func (e *Engine) Interrupt(ctx context.Context, executionID string) error
```

**Parameters:**
- `ctx` - Context for cancellation/timeout
- `executionID` - ID of execution to interrupt

**Returns:**
- `error` - Error if interrupt failed

**Behavior:**
1. Loads execution state
2. Validates state is running
3. Sends interrupt signal to runner
4. Marks state as interrupted
5. Saves checkpoint

**Errors:**
- `ErrExecutionNotFound` - Execution state not found
- `ErrNotRunning` - Execution is not running
- `ErrInterruptFailed` - Runner failed to interrupt

**Example:**
```go
err := engine.Interrupt(ctx, "abc123-def456-ghi789")
if err != nil {
    log.Fatal("Interrupt failed:", err)
}
```

### TriggerSummarization

Manually triggers context summarization.

```go
func (e *Engine) TriggerSummarization(ctx context.Context, task *Task, executionID, model string, state interface{}, force bool) bool
```

**Parameters:**
- `ctx` - Context for cancellation/timeout
- `task` - Task with context to summarize
- `executionID` - Execution ID for logging
- `model` - LLM model to use
- `state` - Current execution state
- `force` - Force summarization regardless of thresholds

**Returns:**
- `bool` - True if summarization was performed

**Example:**
```go
summarized := engine.TriggerSummarization(
    ctx,
    task,
    executionID,
    "claude-3.5-sonnet",
    nil,
    true, // force
)
if summarized {
    log.Printf("Context summarized")
}
```

### CreateManualCheckpoint

Creates a manual checkpoint at any point.

```go
func (e *Engine) CreateManualCheckpoint(ctx context.Context, task *Task, status string) (*ExecutionState, error)
```

**Parameters:**
- `ctx` - Context for cancellation/timeout
- `task` - Task to checkpoint
- `status` - Status for checkpoint (e.g., "running", "pending")

**Returns:**
- `*ExecutionState` - Created execution state
- `error` - Error if checkpoint creation failed

**Example:**
```go
state, err := engine.CreateManualCheckpoint(ctx, task, "running")
if err != nil {
    log.Fatal("Checkpoint failed:", err)
}
log.Printf("Checkpoint ID: %s", state.ID)
```

### GetCheckpoint

Retrieves an execution state by ID.

```go
func (e *Engine) GetCheckpoint(executionID string) (*ExecutionState, error)
```

**Parameters:**
- `executionID` - ID of checkpoint to retrieve

**Returns:**
- `*ExecutionState` - Execution state
- `error` - Error if not found

**Example:**
```go
state, err := engine.GetCheckpoint("abc123-def456-ghi789")
if err != nil {
    log.Fatal("Failed to get checkpoint:", err)
}
log.Printf("Status: %s", state.Status)
```

### ListCheckpoints

Lists all available checkpoint IDs.

```go
func (e *Engine) ListCheckpoints() ([]string, error)
```

**Returns:**
- `[]string` - List of execution IDs
- `error` - Error if list failed

**Example:**
```go
ids, err := engine.ListCheckpoints()
if err != nil {
    log.Fatal("Failed to list checkpoints:", err)
}
for _, id := range ids {
    fmt.Println(id)
}
```

### CleanupOldCheckpoints

Removes old checkpoint files based on retention policy.

```go
func (e *Engine) CleanupOldCheckpoints() error
```

**Returns:**
- `error` - Error if cleanup failed

**Behavior:**
- Removes checkpoints older than `CheckpointAgeLimit`
- Keeps at most `MaxCheckpoints` most recent

**Example:**
```go
err := engine.CleanupOldCheckpoints()
if err != nil {
    log.Printf("Cleanup warning: %v", err)
}
```

### ValidateExecutionState

Validates a saved execution state without loading it.

```go
func (e *Engine) ValidateExecutionState(executionID string) (*StateValidationResult, error)
```

**Parameters:**
- `executionID` - ID of state to validate

**Returns:**
- `*StateValidationResult` - Validation result
- `error` - Error if validation failed

**Example:**
```go
result, err := engine.ValidateExecutionState(executionID)
if err != nil {
    log.Fatal("Validation failed:", err)
}
if !result.CanResume {
    log.Printf("Cannot resume: %s", result.Message)
}
```

### DeleteCorruptedState

Deletes a corrupted state file.

```go
func (e *Engine) DeleteCorruptedState(executionID string) error
```

**Parameters:**
- `executionID` - ID of corrupted state to delete

**Returns:**
- `error` - Error if deletion failed

**Example:**
```go
err := engine.DeleteCorruptedState("abc123-def456-ghi789")
if err != nil {
    log.Fatal("Delete failed:", err)
}
```

### Status

Returns current engine status.

```go
func (e *Engine) Status(ctx context.Context) StatusInfo
```

**Returns:**
- `StatusInfo` - Engine status information

**Example:**
```go
status := engine.Status(ctx)
fmt.Printf("Available runtimes: %v", status.AvailableRuntimes)
fmt.Printf("Daily spending: $%.2f", status.Budget.DailySpending)
```

### GetBudgetTracker

Returns the budget tracker.

```go
func (e *Engine) GetBudgetTracker() *BudgetTracker
```

**Returns:**
- `*BudgetTracker` - Budget tracker instance

---

## ExecutionState API

### NewExecutionState

Creates a new execution state from a task.

```go
func NewExecutionState(task *Task) *ExecutionState
```

**Parameters:**
- `task` - Task to create state for

**Returns:**
- `*ExecutionState` - New execution state

**Example:**
```go
state := execution.NewExecutionState(task)
log.Printf("Execution ID: %s", state.ID)
```

### Start

Marks execution as started.

```go
func (s *ExecutionState) Start() error
```

**Returns:**
- `error` - Error if state cannot be started (e.g., already started)

**Errors:**
- `ErrInvalidStateTransition` - State not in pending status

### Complete

Marks execution as completed.

```go
func (s *ExecutionState) Complete(result *Result) error
```

**Parameters:**
- `result` - Execution result

**Returns:**
- `error` - Error if state cannot be completed

**Errors:**
- `ErrInvalidStateTransition` - State not in running status

### Fail

Marks execution as failed.

```go
func (s *ExecutionState) Fail(err error) error
```

**Parameters:**
- `err` - Error that caused failure

**Returns:**
- `error` - Error if state cannot be failed

### Interrupt

Marks execution as interrupted.

```go
func (s *ExecutionState) Interrupt() error
```

**Returns:**
- `error` - Error if state cannot be interrupted

### Resume

Marks interrupted/failed execution as resuming.

```go
func (s *ExecutionState) Resume() error
```

**Returns:**
- `error` - Error if state cannot be resumed

### IsRunning

Returns true if execution is currently running.

```go
func (s *ExecutionState) IsRunning() bool
```

**Returns:**
- `bool` - True if status is "running"

### IsCompleted

Returns true if execution completed successfully.

```go
func (s *ExecutionState) IsCompleted() bool
```

**Returns:**
- `bool` - True if status is "completed"

### IsFailed

Returns true if execution failed.

```go
func (s *ExecutionState) IsFailed() bool
```

**Returns:**
- `bool` - True if status is "failed"

### IsInterrupted

Returns true if execution was interrupted.

```go
func (s *ExecutionState) IsInterrupted() bool
```

**Returns:**
- `bool` - True if status is "interrupted"

### CanResume

Returns true if execution can be resumed.

```go
func (s *ExecutionState) CanResume() bool
```

**Returns:**
- `bool` - True if status is "pending", "interrupted", or "failed"

### Duration

Returns execution duration.

```go
func (s *ExecutionState) Duration() time.Duration
```

**Returns:**
- `time.Duration` - Time from start to completion or current time if running

### ValidateChecksum

Validates state checksum.

```go
func (s *ExecutionState) ValidateChecksum() bool
```

**Returns:**
- `bool` - True if checksum is valid

### Validate

Validates execution state integrity.

```go
func (s *ExecutionState) Validate() error
```

**Returns:**
- `error` - Error if validation fails

### CheckVersionCompatibility

Checks version compatibility.

```go
func (s *ExecutionState) CheckVersionCompatibility() VersionCompatibility
```

**Returns:**
- `VersionCompatibility` - Compatibility information

### ValidateForResume

Validates state for resume.

```go
func (s *ExecutionState) ValidateForResume() []ValidationError
```

**Returns:**
- `[]ValidationError` - List of validation errors

### ValidateEnvironment

Validates execution environment.

```go
func (s *ExecutionState) ValidateEnvironment() []ValidationError
```

**Returns:**
- `[]ValidationError` - List of validation errors

### SetMetadata

Sets metadata key-value pair.

```go
func (s *ExecutionState) SetMetadata(key, value string)
```

### GetMetadata

Gets metadata value by key.

```go
func (s *ExecutionState) GetMetadata(key string) (string, bool)
```

**Returns:**
- `string` - Metadata value
- `bool` - True if key exists

### ToJSON

Serializes state to JSON.

```go
func (s *ExecutionState) ToJSON() ([]byte, error)
```

**Returns:**
- `[]byte` - JSON data
- `error` - Error if serialization fails

### FromJSON

Deserializes state from JSON.

```go
func (s *ExecutionState) FromJSON(data []byte) error
```

**Parameters:**
- `data` - JSON data

**Returns:**
- `error` - Error if deserialization fails

### Clone

Creates a deep copy of the state.

```go
func (s *ExecutionState) Clone() *ExecutionState
```

**Returns:**
- `*ExecutionState` - Cloned state

---

## Context Management API

### NewTaskContext

Creates a new task context.

```go
func NewTaskContext(projectPath string) *TaskContext
```

**Parameters:**
- `projectPath` - Root path of project

**Returns:**
- `*TaskContext` - New task context

### WithChange

Sets change ID.

```go
func (tc *TaskContext) WithChange(changeID string) *TaskContext
```

### WithTask

Sets task ID.

```go
func (tc *TaskContext) WithTask(taskID string) *TaskContext
```

### WithWorkflow

Sets workflow ID.

```go
func (tc *TaskContext) WithWorkflow(workflowID string) *TaskContext
```

### WithFiles

Sets relevant files.

```go
func (tc *TaskContext) WithFiles(files []string) *TaskContext
```

### AddFile

Adds a file to context.

```go
func (tc *TaskContext) AddFile(file string) *TaskContext
```

### HasFiles

Returns true if context has files.

```go
func (tc *TaskContext) HasFiles() bool
```

### RelativePath

Converts absolute path to relative path.

```go
func (tc *TaskContext) RelativePath(absPath string) (string, error)
```

### AbsolutePath

Converts relative path to absolute path.

```go
func (tc *TaskContext) AbsolutePath(relPath string) string
```

### LoadSummarizationThreshold

Loads summarization thresholds from config file.

```go
func LoadSummarizationThreshold(projectPath string) (SummarizationThreshold, error)
```

**Parameters:**
- `projectPath` - Path to project

**Returns:**
- `SummarizationThreshold` - Thresholds from file or defaults
- `error` - Error if file exists but is invalid

### SaveSummarizationThreshold

Saves summarization thresholds to config file.

```go
func SaveSummarizationThreshold(projectPath string, threshold SummarizationThreshold, model string) error
```

**Parameters:**
- `projectPath` - Path to project
- `threshold` - Thresholds to save
- `model` - Model to use for summarization

**Returns:**
- `error` - Error if save fails

### DefaultSummarizationThreshold

Returns default summarization thresholds.

```go
func DefaultSummarizationThreshold() SummarizationThreshold
```

**Returns:**
- `SummarizationThreshold` - Default thresholds:
  - FileCount: 50
  - ContextLength: 50000
  - TokenCount: 10000

---

## Sandbox API

### NewSandbox

Creates a new sandbox with given limits.

```go
func NewSandbox(limits ResourceLimits) *Sandbox
```

**Parameters:**
- `limits` - Resource limits

**Returns:**
- `*Sandbox` - New sandbox instance

### DefaultResourceLimits

Returns safe default limits.

```go
func DefaultResourceLimits() ResourceLimits
```

**Returns:**
- `ResourceLimits` - Default limits:
  - MaxMemoryMB: 128
  - MaxCPUTime: 30s
  - MaxExecTime: 60s

### CheckMemoryLimit

Checks if current memory usage is within limits.

```go
func (s *Sandbox) CheckMemoryLimit() error
```

**Returns:**
- `error` - ErrMemoryLimitExceeded if over limit

### CheckCPULimit

Checks if CPU time is within limits.

```go
func (s *Sandbox) CheckCPULimit(elapsed time.Duration) error
```

**Parameters:**
- `elapsed` - Elapsed CPU time

**Returns:**
- `error` - ErrCPULimitExceeded if over limit

### CheckExecLimit

Checks if wall-clock time is within limits.

```go
func (s *Sandbox) CheckExecLimit(elapsed time.Duration) error
```

**Parameters:**
- `elapsed` - Elapsed wall-clock time

**Returns:**
- `error` - ErrExecLimitExceeded if over limit

### WithFileAccess

Adds allowed file paths.

```go
func (s *Sandbox) WithFileAccess(paths ...string) *Sandbox
```

### WithAPIAccess

Adds allowed API calls.

```go
func (s *Sandbox) WithAPIAccess(apis ...string) *Sandbox
```

### CheckFileAccess

Checks if file access is allowed.

```go
func (s *Sandbox) CheckFileAccess(path string) error
```

**Returns:**
- `error` - Error if access denied

### CheckAPIAccess

Checks if API call is allowed.

```go
func (s *Sandbox) CheckAPIAccess(api string) error
```

**Returns:**
- `error` - Error if access denied

### GetLimits

Returns sandbox resource limits.

```go
func (s *Sandbox) GetLimits() ResourceLimits
```

---

## Storage API

### SaveState

Saves execution state to disk.

```go
func SaveState(state *ExecutionState) error
```

**Parameters:**
- `state` - State to save

**Returns:**
- `error` - Error if save fails

**Behavior:**
- Saves to `~/.go-ent/executions/{executionID}.json`
- Uses atomic write (temp file + rename)
- Validates state before save

### LoadState

Loads execution state from disk.

```go
func LoadState(executionID string) (*ExecutionState, error)
```

**Parameters:**
- `executionID` - ID of state to load

**Returns:**
- `*ExecutionState` - Loaded state
- `error` - Error if load fails

**Behavior:**
- Loads from `~/.go-ent/executions/{executionID}.json`
- Validates JSON format
- Validates checksum
- Validates state integrity

### LoadStateWithValidation

Loads state with additional validation.

```go
func LoadStateWithValidation(executionID string) (*ExecutionState, []ValidationIssue, error)
```

**Returns:**
- `*ExecutionState` - Loaded state
- `[]ValidationIssue` - List of validation issues
- `error` - Error if load fails

### ValidateStateFile

Validates a state file without loading it.

```go
func ValidateStateFile(executionID string) (*StateValidationResult, error)
```

**Returns:**
- `*StateValidationResult` - Validation result
- `error` - Error if validation cannot be performed

### DeleteState

Deletes a state file.

```go
func DeleteState(executionID string) error
```

**Parameters:**
- `executionID` - ID of state to delete

**Returns:**
- `error` - Error if deletion fails

### ListExecutions

Lists all execution IDs.

```go
func ListExecutions() ([]string, error)
```

**Returns:**
- `[]string` - List of execution IDs
- `error` - Error if list fails

### EnsureExecutionsDir

Ensures executions directory exists.

```go
func EnsureExecutionsDir() error
```

**Returns:**
- `error` - Error if directory creation fails

---

## MCP Tools API

### engine_interrupt

Interrupts a running execution by ID.

**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "execution_id": {
      "type": "string",
      "description": "ID of execution to interrupt"
    }
  },
  "required": ["execution_id"]
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

**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "execution_id": {
      "type": "string",
      "description": "ID of execution to resume"
    }
  },
  "required": ["execution_id"]
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

**Input Schema:**
```json
{
  "type": "object",
  "properties": {}
}
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

**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "execution_id": {
      "type": "string",
      "description": "ID of execution to find"
    }
  },
  "required": ["execution_id"]
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

## Types and Constants

### Task

```go
type Task struct {
    Description string
    Type        domain.TaskType
    Context     *TaskContext
    ForceAgent  domain.AgentRole
    ForceModel  string
    ForceRuntime domain.Runtime
    ForceStrategy domain.ExecutionStrategy
    Budget     *BudgetLimit
    Skills     []string
    Metadata    map[string]interface{}
}
```

### Result

```go
type Result struct {
    Success     bool
    Output      string
    Error       string
    TokensIn    int
    TokensOut   int
    Cost        float64
    Duration    time.Duration
    Adjustments []string
    Metadata    map[string]interface{}
}
```

### Execution Status Constants

```go
const (
    ExecutionStatusPending     = "pending"
    ExecutionStatusRunning     = "running"
    ExecutionStatusInterrupted = "interrupted"
    ExecutionStatusCompleted   = "completed"
    ExecutionStatusFailed      = "failed"
)
```

### LLM Model Constants

```go
const (
    LLMModelClaude35   = "claude-3.5-sonnet"
    LLMModelClaude4    = "claude-4-5-20251101"
    LLMModelGPT4       = "gpt-4"
    LLMModelGPT4Turbo  = "gpt-4-turbo"
    LLMModelGPT35Turbo = "gpt-3.5-turbo"
)
```

### State Version

```go
const StateVersion = "1.0.0"
```

---

## Error Handling

### Standard Errors

```go
var (
    ErrExecutionNotFound   = errors.New("execution not found")
    ErrNotRunning        = errors.New("execution is not running")
    ErrInvalidState      = errors.New("invalid execution state")
    ErrVersionMismatch   = errors.New("version mismatch")
    ErrResourceExceeded  = errors.New("resource limit exceeded")
)
```

### Resource Exceeded Errors

```go
type ResourceExceededError struct {
    Resource string
    Limit    interface{}
}

func (e *ResourceExceededError) Error() string
func (e *ResourceExceededError) Unwrap() error
```

### Corrupted State Errors

```go
type CorruptedStateError struct {
    ExecutionID string
    Filename    string
    Reason      string
    CanRecover  bool
}

func (e *CorruptedStateError) Error() string
func (e *CorruptedStateError) Unwrap() error
```

### Validation Errors

```go
type ValidationError struct {
    Field   string
    Message string
    Level   ValidationLevel
}
```

### Validation Levels

```go
const (
    ValidationLevelInfo    = "info"
    ValidationLevelWarning = "warning"
    ValidationLevelError   = "error"
)
```

### Error Checking Patterns

```go
// Check for specific error type
var resourceErr *execution.ResourceExceededError
if errors.As(err, &resourceErr) {
    log.Printf("Resource %s exceeded: %v", resourceErr.Resource, resourceErr.Limit)
}

// Check for corrupted state
var corruptedErr *execution.CorruptedStateError
if errors.As(err, &corruptedErr) {
    if corruptedErr.CanRecover {
        // Attempt recovery
    } else {
        // Delete corrupted state
    }
}

// Wrap errors with context
return fmt.Errorf("execute task: %w", err)
```

---

## See Also

- [Execution Engine v2 Documentation](EXECUTION_ENGINE_V2.md)
- [Execution Engine Examples](EXECUTION_ENGINE_EXAMPLES.md)
- [Troubleshooting Guide](EXECUTION_ENGINE_TROUBLESHOOTING.md)
