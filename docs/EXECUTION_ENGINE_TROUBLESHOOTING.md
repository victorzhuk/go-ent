# Execution Engine v2 - Troubleshooting Guide

This guide helps you diagnose and resolve common issues with the execution engine v2.

## Table of Contents

1. [Execution Not Checkpointing](#execution-not-checkpointing)
2. [Context Not Summarizing](#context-not-summarizing)
3. [Interrupt Not Working](#interrupt-not-working)
4. [Resume Failing](#resume-failing)
5. [Performance Issues](#performance-issues)
6. [State File Corruption](#state-file-corruption)
7. [Debugging Tips](#debugging-tips)
8. [Recovery Procedures](#recovery-procedures)

---

## Execution Not Checkpointing

### Symptoms

- Checkpoint files are not being created in `~/.go-ent/executions/`
- Execution completes but no state files exist
- Manual checkpoint creation fails

### Common Causes

#### 1. Auto-Checkpoint Not Enabled

**Problem:** Checkpointing is disabled by default.

**Solution:** Enable auto-checkpoint in configuration:

```go
cfg := execution.Config{
    EnableAutoCheckpoint: true,
    MaxCheckpoints:       10,
    CheckpointAgeLimit:   24 * time.Hour,
}
```

#### 2. Executions Directory Not Accessible

**Problem:** Cannot create or write to `~/.go-ent/executions/`.

**Solution:** Check directory permissions:

```bash
# Check if directory exists
ls -la ~/.go-ent/executions/

# Create directory with correct permissions
mkdir -p ~/.go-ent/executions/
chmod 755 ~/.go-ent/executions/
```

**Code fix:**

```go
import "os"

err := execution.EnsureExecutionsDir()
if err != nil {
    if os.IsPermission(err) {
        log.Fatal("Permission denied: cannot create executions directory")
    }
    log.Fatal("Failed to create executions directory:", err)
}
```

#### 3. Disk Space Full

**Problem:** Cannot write checkpoint files due to insufficient disk space.

**Solution:** Check available disk space:

```bash
df -h ~/.go-ent/
```

**Cleanup old checkpoints:**

```go
err := engine.CleanupOldCheckpoints()
if err != nil {
    log.Printf("Cleanup failed: %v", err)
}
```

#### 4. State Validation Failing

**Problem:** Checkpoint creation fails during state validation.

**Solution:** Validate state before attempting to save:

```go
state := execution.NewExecutionState(task)

err := state.Validate()
if err != nil {
    log.Printf("State validation failed: %v", err)
    log.Printf("Task: %s", state.Task.Description)
    log.Printf("ID: %s", state.ID)

    // Fix common issues
    if state.Task.Description == "" {
        state.Task.Description = "Default description"
    }
}

// Retry validation
if state.Validate() != nil {
    log.Fatal("Cannot create valid state")
}

// Save state
err = execution.SaveState(state)
if err != nil {
    log.Fatal("Failed to save state:", err)
}
```

### Logs to Check

```bash
# Enable debug logging
export LOG_LEVEL=debug

# Run execution and check logs
go run main.go 2>&1 | grep -i checkpoint

# Look for these messages:
# - "creating checkpoint"
# - "saved execution state"
# - "checkpoint created"
```

### Diagnostic Checklist

- [ ] Auto-checkpoint enabled in configuration?
- [ ] Executions directory exists and is writable?
- [ ] Sufficient disk space available?
- [ ] State validation passing?
- [ ] No permission errors in logs?
- [ ] Checkpoint file created after execution?

---

## Context Not Summarizing

### Symptoms

- Context size keeps growing without summarization
- Token limit errors during execution
- `IsSummarized` flag remains false
- Summarization never triggers automatically

### Common Causes

#### 1. Summarization Not Enabled

**Problem:** Summarization is disabled by default.

**Solution:** Enable summarization in configuration:

```go
cfg := execution.Config{
    EnableSummarization: true,
    SummarizationThreshold: execution.SummarizationThreshold{
        FileCount:     50,
        ContextLength: 50000,
        TokenCount:    10000,
    },
    SummarizationModel: "claude-3.5-sonnet",
}
```

#### 2. Thresholds Too High

**Problem:** Context never exceeds configured thresholds.

**Solution:** Lower thresholds:

```go
cfg := execution.Config{
    SummarizationThreshold: execution.SummarizationThreshold{
        FileCount:     25,  // Lower from 50
        ContextLength: 25000,  // Lower from 50000
        TokenCount:    5000,  // Lower from 10000
    },
}
```

#### 3. Manual Trigger Required

**Problem:** Automatic summarization not working, need to force it.

**Solution:** Trigger summarization manually:

```go
summarized := engine.TriggerSummarization(
    ctx,
    task,
    executionID,
    "claude-3.5-sonnet",
    nil,
    true,  // Force summarization
)

if summarized {
    log.Printf("Context summarized")
}
```

#### 4. LLM Client Not Configured

**Problem:** Summarization requires LLM client but it's not set up.

**Solution:** Configure LLM client for summarization:

```go
import "github.com/victorzhuk/go-ent/internal/llm"

client := llm.NewClient(llm.Config{
    APIKey:  os.Getenv("ANTHROPIC_API_KEY"),
    Model:    "claude-3.5-sonnet",
})

// Pass LLM client to engine (if API supports it)
```

#### 5. Context Already Summarized

**Problem:** Context was already summarized and doesn't need it again.

**Solution:** Check summarization state:

```go
if task.Context.IsSummarized {
    log.Printf("Context already summarized %d times",
        task.Context.SummarizationCount)

    // Access original context if available
    if task.Context.OriginalContext != nil {
        log.Printf("Original context available")
        original := task.Context.OriginalContext
    }
}
```

### Monitoring Context Size

```go
func monitorContextSize(task *execution.Task, threshold execution.SummarizationThreshold) {
    fileCount := len(task.Context.Files)

    totalLength := 0
    for _, file := range task.Context.Files {
        totalLength += len(file)
    }

    estimatedTokens := totalLength / 4

    log.Printf("Context size: %d files, %d chars, ~%d tokens",
        fileCount, totalLength, estimatedTokens)

    // Check if summarization should trigger
    if fileCount > threshold.FileCount {
        log.Printf("⚠️  File count exceeds threshold: %d > %d",
            fileCount, threshold.FileCount)
    }

    if totalLength > threshold.ContextLength {
        log.Printf("⚠️  Context length exceeds threshold: %d > %d",
            totalLength, threshold.ContextLength)
    }

    if estimatedTokens > threshold.TokenCount {
        log.Printf("⚠️  Token count exceeds threshold: %d > %d",
            estimatedTokens, threshold.TokenCount)
    }

    // Trigger manual summarization if needed
    if fileCount > threshold.FileCount ||
       totalLength > threshold.ContextLength ||
       estimatedTokens > threshold.TokenCount {
        summarized := engine.TriggerSummarization(
            ctx, task, executionID, "claude-3.5-sonnet", nil, true)
        if summarized {
            log.Printf("✅ Context summarized")
        }
    }
}
```

### Config File Check

```bash
# Check if config file exists
cat ~/.go-ent/summarization.yaml

# Expected format:
# file_count: 50
# context_length: 50000
# token_count: 10000
# model: "claude-3.5-sonnet"

# If missing, create it
cat > ~/.go-ent/summarization.yaml << EOF
file_count: 50
context_length: 50000
token_count: 10000
model: "claude-3.5-sonnet"
EOF
```

### Logs to Check

```bash
# Check summarization logs
go run main.go 2>&1 | grep -i summarize

# Look for:
# - "context summarized"
# - "context within thresholds"
# - "triggering summarization"
```

---

## Interrupt Not Working

### Symptoms

- `engine.Interrupt()` returns error
- Execution continues running after interrupt
- Status not changed to "interrupted"
- No checkpoint saved after interrupt

### Common Causes

#### 1. Execution Not Running

**Problem:** Cannot interrupt an execution that isn't running.

**Solution:** Check execution status before interrupt:

```go
state, err := execution.LoadState(executionID)
if err != nil {
    log.Fatal("Failed to load state:", err)
}

if !state.IsRunning() {
    log.Printf("Cannot interrupt: execution is %s", state.Status)
    return fmt.Errorf("execution is not running (status: %s)", state.Status)
}

// Now interrupt
err = engine.Interrupt(ctx, executionID)
```

#### 2. Runner Not Supporting Interrupt

**Problem:** Selected runner doesn't implement interrupt properly.

**Solution:** Use a different runtime:

```go
// Force use of a runtime that supports interrupt
task := &execution.Task{
    ForceRuntime: domain.RuntimeClaudeCode,  // Supports interrupt
}

result, err := engine.Execute(ctx, task)
```

#### 3. Execution ID Not Found

**Problem:** Interrupt called with invalid execution ID.

**Solution:** Validate execution ID before interrupt:

```go
// List available executions
executionIDs, err := execution.ListExecutions()
if err != nil {
    log.Fatal("Failed to list executions:", err)
}

log.Printf("Available executions:")
for _, id := range executionIDs {
    state, _ := execution.LoadState(id)
    log.Printf("- %s: %s (%s)", id, state.Task.Description, state.Status)
}

// Prompt user for valid ID
fmt.Print("Enter execution ID to interrupt: ")
var executionID string
fmt.Scanln(&executionID)

// Check if ID exists
if !contains(executionIDs, executionID) {
    log.Fatal("Invalid execution ID")
}

// Proceed with interrupt
err = engine.Interrupt(ctx, executionID)
```

#### 4. Context Cancelled

**Problem:** Context passed to `Interrupt()` is already cancelled.

**Solution:** Use fresh context for interrupt:

```go
// Don't use the execution context
// Use a fresh context for interrupt
interruptCtx := context.Background()

err := engine.Interrupt(interruptCtx, executionID)
if err != nil {
    log.Fatal("Interrupt failed:", err)
}
```

#### 5. Runner Not Responding

**Problem:** Runner is hung and not responding to interrupt signal.

**Solution:** Add timeout to interrupt:

```go
interruptCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

err := engine.Interrupt(interruptCtx, executionID)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Printf("Interrupt timed out, execution may be stuck")

        // Force kill if necessary
        log.Printf("Execution ID: %s - may require manual cleanup", executionID)
    }
    log.Fatal("Interrupt failed:", err)
}
```

### Diagnostic Code

```go
func diagnoseInterrupt(ctx context.Context, engine *execution.Engine, executionID string) {
    // 1. Check if execution exists
    state, err := execution.LoadState(executionID)
    if err != nil {
        if os.IsNotExist(err) {
            log.Printf("❌ Execution not found: %s", executionID)
        } else {
            log.Printf("❌ Failed to load execution: %v", err)
        }
        return
    }

    // 2. Check status
    log.Printf("Execution status: %s", state.Status)
    if !state.IsRunning() {
        log.Printf("⚠️  Cannot interrupt execution with status: %s", state.Status)
        return
    }

    // 3. Check runtime
    log.Printf("Runtime: %s", state.Runtime)
    if state.Runtime == "" {
        log.Printf("⚠️  No runtime configured")
        return
    }

    // 4. Check runner availability
    selector := agent.NewSelector(agent.Config{}, nil)
    cfg := execution.Config{}
    testEngine := execution.New(cfg, selector)

    // Try to get runner
    if runner, err := testEngine.getRunner(state.Runtime); err != nil {
        log.Printf("⚠️  Runner error: %v", err)
    } else if !runner.Available(ctx) {
        log.Printf("⚠️  Runner not available: %s", state.Runtime)
    }

    // 5. Attempt interrupt
    log.Printf("Attempting interrupt...")
    err = engine.Interrupt(ctx, executionID)
    if err != nil {
        log.Printf("❌ Interrupt failed: %v", err)
    } else {
        log.Printf("✅ Interrupt succeeded")

        // Verify status changed
        state, _ = execution.LoadState(executionID)
        log.Printf("New status: %s", state.Status)
    }
}
```

### Logs to Check

```bash
# Check interrupt logs
go run main.go 2>&1 | grep -i interrupt

# Look for:
# - "interrupting execution"
# - "sending interrupt signal"
# - "execution interrupted successfully"
# - "runner interrupt returned error"
```

---

## Resume Failing

### Symptoms

- `engine.ResumeExecution()` returns error
- State validation fails on resume
- Version mismatch errors
- Runtime not available after resume

### Common Causes

#### 1. State Version Incompatible

**Problem:** State version doesn't match current engine version.

**Solution:** Check version compatibility:

```go
state, err := execution.LoadState(executionID)
if err != nil {
    log.Fatal("Failed to load state:", err)
}

compat := state.CheckVersionCompatibility()
log.Printf("Version compatibility:")
log.Printf("  State version: %s", compat.StateVersion)
log.Printf("  Current version: %s", compat.CurrentVersion)
log.Printf("  Compatible: %v", compat.Compatible)
log.Printf("  Policy: %s", compat.Policy)
log.Printf("  Reason: %s", compat.Reason)

if !compat.Compatible {
    log.Printf("❌ Cannot resume: version incompatible")

    if compat.Policy == "backward-compatible" {
        log.Printf("✓ Version is backward compatible, can attempt resume")
    } else if compat.Policy == "incompatible" {
        log.Printf("✗ Version is incompatible, cannot resume")
        return
    }
}
```

#### 2. Checksum Validation Failed

**Problem:** State file checksum doesn't match data (corruption).

**Solution:** Handle corrupted state:

```go
state, err := execution.LoadState(executionID)
if err != nil {
    var corruptedErr *execution.CorruptedStateError
    if errors.As(err, &corruptedErr) {
        log.Printf("❌ Corrupted state: %s", corruptedErr.Error())
        log.Printf("  Filename: %s", corruptedErr.Filename)
        log.Printf("  Reason: %s", corruptedErr.Reason)
        log.Printf("  Can recover: %v", corruptedErr.CanRecover)

        if !corruptedErr.CanRecover {
            // Delete corrupted state
            selector := agent.NewSelector(agent.Config{}, nil)
            cfg := execution.Config{}
            engine := execution.New(cfg, selector)

            fmt.Printf("Delete corrupted state? (y/n): ")
            var response string
            fmt.Scanln(&response)

            if response == "y" {
                err := engine.DeleteCorruptedState(executionID)
                if err != nil {
                    log.Fatal("Delete failed:", err)
                }
                log.Printf("Deleted corrupted state")
            }
        }
        return
    }
    log.Fatal("Load failed:", err)
}

// Proceed with resume
result, err := engine.ResumeExecution(ctx, executionID)
```

#### 3. State Validation Failed

**Problem:** State validation fails on resume.

**Solution:** Check validation errors:

```go
state, err := execution.LoadState(executionID)
if err != nil {
    log.Fatal("Failed to load state:", err)
}

validationErrors := state.ValidateForResume()
if len(validationErrors) > 0 {
    log.Printf("❌ Validation errors found:")

    hasErrors := false
    for _, verr := range validationErrors {
        if verr.Level == execution.ValidationLevelError {
            log.Printf("  [ERROR] %s: %s", verr.Field, verr.Message)
            hasErrors = true
        } else {
            log.Printf("  [WARNING] %s: %s", verr.Field, verr.Message)
        }
    }

    if hasErrors {
        log.Printf("Cannot resume due to validation errors")
        return
    }
}

// Resume with warnings
result, err := engine.ResumeExecution(ctx, executionID)
```

#### 4. Runtime Not Available

**Problem:** Required runtime is not available on resume.

**Solution:** Check runtime availability:

```go
state, err := execution.LoadState(executionID)
if err != nil {
    log.Fatal("Failed to load state:", err)
}

if state.Runtime == "" {
    log.Printf("⚠️  No runtime configured in state")
    return
}

// Check if runtime is available
selector := agent.NewSelector(agent.Config{}, nil)
cfg := execution.Config{}
testEngine := execution.New(cfg, selector)

runner, err := testEngine.getRunner(state.Runtime)
if err != nil {
    log.Printf("⚠️  Runtime %s not registered", state.Runtime)
    return
}

if !runner.Available(ctx) {
    log.Printf("⚠️  Runtime %s is not available", state.Runtime)
    log.Printf("Available runtimes:")

    status := testEngine.Status(ctx)
    for _, rt := range status.AvailableRuntimes {
        log.Printf("  - %s", rt)
    }

    return
}

// Runtime is available, proceed with resume
result, err := engine.ResumeExecution(ctx, executionID)
```

#### 5. Environment Validation Failed

**Problem:** Execution environment changed since state was created.

**Solution:** Check environment validation:

```go
state, err := execution.LoadState(executionID)
if err != nil {
    log.Fatal("Failed to load state:", err)
}

envErrors := state.ValidateEnvironment()
if len(envErrors) > 0 {
    log.Printf("⚠️  Environment validation issues:")

    for _, eerr := range envErrors {
        if eerr.Level == execution.ValidationLevelError {
            log.Printf("  [ERROR] %s: %s", eerr.Field, eerr.Message)
        } else {
            log.Printf("  [WARNING] %s: %s", eerr.Field, eerr.Message)
        }
    }

    // Ask user if they want to proceed
    log.Printf("Some files or paths may have changed")
    fmt.Printf("Proceed with resume anyway? (y/n): ")
    var response string
    fmt.Scanln(&response)

    if response != "y" {
        return
    }
}

// Proceed with resume
result, err := engine.ResumeExecution(ctx, executionID)
```

### Pre-Resume Validation Checklist

```go
func validateForResume(executionID string) error {
    // 1. Load state
    state, err := execution.LoadState(executionID)
    if err != nil {
        return fmt.Errorf("load state: %w", err)
    }

    // 2. Check version compatibility
    compat := state.CheckVersionCompatibility()
    if !compat.Compatible {
        return fmt.Errorf("version incompatible: %s (state: %s, current: %s)",
            compat.Reason, compat.StateVersion, compat.CurrentVersion)
    }

    // 3. Validate state integrity
    validationErrors := state.ValidateForResume()
    for _, verr := range validationErrors {
        if verr.Level == execution.ValidationLevelError {
            return fmt.Errorf("validation error: %s: %s", verr.Field, verr.Message)
        }
    }

    // 4. Validate environment
    envErrors := state.ValidateEnvironment()
    for _, eerr := range envErrors {
        if eerr.Level == execution.ValidationLevelError {
            return fmt.Errorf("environment error: %s: %s", eerr.Field, eerr.Message)
        }
    }

    // 5. Check if resume is possible
    if !state.CanResume() {
        return fmt.Errorf("cannot resume from status: %s", state.Status)
    }

    log.Printf("✅ Resume validation passed")
    return nil
}
```

### Logs to Check

```bash
# Check resume logs
go run main.go 2>&1 | grep -i resume

# Look for:
# - "resuming execution"
# - "version compatibility check"
# - "state validation error"
# - "environment validation error"
```

---

## Performance Issues

### Symptoms

- Slow execution
- High memory usage
- Long checkpoint save times
- Summarization taking too long

### Common Causes

#### 1. Too Many Checkpoints

**Problem:** Accumulating too many checkpoint files slows down operations.

**Solution:** Adjust retention policy:

```go
cfg := execution.Config{
    EnableAutoCheckpoint: true,
    MaxCheckpoints:       5,  // Keep only 5 recent checkpoints
    CheckpointAgeLimit:   12 * time.Hour,  // Delete after 12 hours
}
```

**Manual cleanup:**

```go
// List all checkpoints
executionIDs, err := execution.ListExecutions()
if err != nil {
    log.Fatal("Failed to list:", err)
}

log.Printf("Found %d checkpoints", len(executionIDs))

// Delete old ones
deleted := 0
for _, id := range executionIDs {
    state, err := execution.LoadState(id)
    if err != nil {
        continue
    }

    // Check age
    age := time.Since(state.UpdatedAt)
    if age > 24*time.Hour {
        err := execution.DeleteState(id)
        if err != nil {
            log.Printf("Failed to delete %s: %v", id, err)
        } else {
            deleted++
        }
    }
}

log.Printf("Deleted %d old checkpoints", deleted)
```

#### 2. Large Context Files

**Problem:** Large context files slow down operations.

**Solution:** Enable summarization more aggressively:

```go
cfg := execution.Config{
    EnableSummarization: true,
    SummarizationThreshold: execution.SummarizationThreshold{
        FileCount:     20,  // Trigger earlier
        ContextLength: 20000,  // Trigger earlier
        TokenCount:    5000,  // Trigger earlier
    },
}
```

#### 3. Frequent Checkpointing

**Problem:** Checkpointing too frequently impacts performance.

**Solution:** Reduce checkpoint frequency:

```go
// Only checkpoint on major milestones, not every task
if task.IsMajorMilestone() {
    err := engine.CreateManualCheckpoint(ctx, task, "running")
    if err != nil {
        log.Printf("Checkpoint failed: %v", err)
    }
}
```

#### 4. Summarization Overhead

**Problem:** LLM summarization takes time and costs money.

**Solution:** Cache summaries or use cheaper model:

```go
cfg := execution.Config{
    SummarizationModel: "claude-3-haiku",  // Cheaper model
}

// Or implement caching
type SummaryCache struct {
    cache map[string]string
}

func (c *SummaryCache) Get(key string) (string, bool) {
    summary, ok := c.cache[key]
    return summary, ok
}

func (c *SummaryCache) Set(key, summary string) {
    c.cache[key] = summary
}
```

### Performance Monitoring

```go
func monitorPerformance(engine *execution.Engine) {
    status := engine.Status(ctx)

    log.Printf("Engine Status:")
    log.Printf("  Available runtimes: %d", len(status.AvailableRuntimes))
    log.Printf("  Daily tokens: %d", status.Budget.DailyTokens)
    log.Printf("  Daily cost: $%.2f", status.Budget.DailySpending)
    log.Printf("  Monthly tokens: %d", status.Budget.MonthlyTokens)
    log.Printf("  Monthly cost: $%.2f", status.Budget.MonthlySpending)

    // List checkpoints
    executionIDs, err := engine.ListCheckpoints()
    if err == nil {
        log.Printf("  Checkpoints: %d", len(executionIDs))

        var totalSize int64
        for _, id := range executionIDs {
            state, _ := engine.GetCheckpoint(id)
            if state != nil {
                data, _ := state.ToJSON()
                totalSize += int64(len(data))
            }
        }

        log.Printf("  Total checkpoint size: %.2f MB",
            float64(totalSize)/(1024*1024))
    }
}
```

### Performance Tuning

```go
// Optimize configuration for performance
cfg := execution.Config{
    Logger:               logger,
    PreferredRuntime:      domain.RuntimeClaudeCode,
    EnableSummarization:  true,
    SummarizationThreshold: execution.SummarizationThreshold{
        FileCount:     30,  // Balance between context size and summarization overhead
        ContextLength: 30000,
        TokenCount:    8000,
    },
    SummarizationModel:   "claude-3-haiku",  // Faster summarization
    EnableAutoCheckpoint: true,
    MaxCheckpoints:       5,  // Keep fewer checkpoints
    CheckpointAgeLimit:   6 * time.Hour,  // Shorter retention
}
```

---

## State File Corruption

### Symptoms

- Checksum validation fails
- JSON parse errors
- State loads but validation fails
- CorruptedStateError on load

### Common Causes

#### 1. Incomplete Write

**Problem:** Checkpoint write interrupted (power loss, process kill).

**Solution:** Engine uses atomic writes (temp file + rename), but check for `.tmp` files:

```bash
# Check for incomplete writes
ls -la ~/.go-ent/executions/*.tmp

# Remove temp files
rm ~/.go-ent/executions/*.tmp
```

#### 2. Disk Errors

**Problem:** Disk corruption affecting state files.

**Solution:** Check disk health:

```bash
# Check disk for errors
sudo fsck -f /dev/sda1  # Replace with your device

# Check SMART status
sudo smartctl -a /dev/sda
```

#### 3. Concurrent Access

**Problem:** Multiple processes writing to same state file.

**Solution:** Ensure single instance:

```go
import "os"

lockFile := "/tmp/go-ent-execution.lock"

f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
if err != nil {
    if os.IsExist(err) {
        log.Fatal("Another execution is already running")
    }
    log.Fatal("Failed to create lock:", err)
}
defer os.Remove(lockFile)

// Proceed with execution
```

#### 4. Memory Corruption

**Problem:** In-memory state corrupted before save.

**Solution:** Validate before save:

```go
state := execution.NewExecutionState(task)

// Validate checksum in memory
if !state.ValidateChecksum() {
    log.Fatal("State checksum invalid in memory")
}

// Validate fields
err := state.Validate()
if err != nil {
    log.Fatal("State validation failed:", err)
}

// Save with atomic write
err = execution.SaveState(state)
if err != nil {
    log.Fatal("Save failed:", err)
}

// Verify after save
loaded, err := execution.LoadState(state.ID)
if err != nil {
    log.Fatal("Load verification failed:", err)
}

if !loaded.ValidateChecksum() {
    log.Fatal("Checksum mismatch after save")
}
```

### Recovery Procedures

#### Recoverable Corruption

```go
state, err := execution.LoadState(executionID)
if err != nil {
    var corruptedErr *execution.CorruptedStateError
    if errors.As(err, &corruptedErr) {
        if corruptedErr.CanRecover {
            log.Printf("Attempting recovery...")

            // Try to extract valid data
            recovered := attemptRecovery(corruptedErr.Filename)
            if recovered != nil {
                // Save recovered state
                err := execution.SaveState(recovered)
                if err != nil {
                    log.Fatal("Failed to save recovered state:", err)
                }
                log.Printf("✅ Recovered state saved")
                return
            }
        }
    }
    log.Fatal("Cannot recover:", err)
}
```

#### Unrecoverable Corruption

```go
state, err := execution.LoadState(executionID)
if err != nil {
    var corruptedErr *execution.CorruptedStateError
    if errors.As(err, &corruptedErr) {
        if !corruptedErr.CanRecover {
            log.Printf("❌ State is unrecoverable")
            log.Printf("  Reason: %s", corruptedErr.Reason)
            log.Printf("  Filename: %s", corruptedErr.Filename)

            // Offer to delete
            fmt.Printf("Delete corrupted state? (y/n): ")
            var response string
            fmt.Scanln(&response)

            if response == "y" {
                selector := agent.NewSelector(agent.Config{}, nil)
                cfg := execution.Config{}
                engine := execution.New(cfg, selector)

                err := engine.DeleteCorruptedState(executionID)
                if err != nil {
                    log.Fatal("Delete failed:", err)
                }
                log.Printf("Deleted corrupted state")
            }
            return
        }
    }
    log.Fatal("Load failed:", err)
}
```

### Backup Strategy

```go
func backupExecutionState(executionID string) error {
    state, err := execution.LoadState(executionID)
    if err != nil {
        return err
    }

    // Create backup with timestamp
    timestamp := time.Now().Format("20060102-150405")
    backupID := fmt.Sprintf("%s-backup-%s", executionID, timestamp)

    // Save as backup
    state.ID = backupID
    err = execution.SaveState(state)
    if err != nil {
        return fmt.Errorf("save backup: %w", err)
    }

    log.Printf("Backup created: %s", backupID)
    return nil
}
```

---

## Debugging Tips

### Enable Debug Logging

```go
import "log/slog"

logger := slog.New(
    slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    }),
)

cfg := execution.Config{
    Logger: logger,
}
```

### Add Structured Logging

```go
logger.Debug("starting execution",
    "task", task.Description,
    "execution_id", executionID,
    "runtime", task.Runtime,
    "model", task.Model,
)

logger.Debug("checkpoint created",
    "execution_id", state.ID,
    "status", state.Status,
    "checksum", state.Checksum,
)
```

### Use Checkpoints for Debugging

```go
func debugWithCheckpoints(ctx context.Context, engine *execution.Engine, task *execution.Task) {
    step := 0

    // Create initial checkpoint
    state, err := engine.CreateManualCheckpoint(ctx, task, fmt.Sprintf("step-%d", step))
    if err != nil {
        log.Fatal("Checkpoint failed:", err)
    }
    executionID := state.ID

    // Execute task
    result, err := engine.Execute(ctx, task)
    if err != nil {
        log.Fatal("Execution failed:", err)
    }

    // Create final checkpoint
    state, err = engine.CreateManualCheckpoint(ctx, task, "final")
    if err != nil {
        log.Fatal("Checkpoint failed:", err)
    }

    // Compare states
    initial, _ := engine.GetCheckpoint(executionID)
    log.Printf("Initial: %s, Final: %s",
        initial.Status, state.Status)
}
```

### Profile Execution

```go
import (
    "runtime"
    "runtime/pprof"
)

func profileExecution() {
    // Start CPU profiling
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // Execute task
    result, err := engine.Execute(ctx, task)

    // Start memory profiling
    f2, _ := os.Create("mem.prof")
    runtime.GC()
    pprof.WriteHeapProfile(f2)
    f2.Close()

    // Analyze profiles
    // go tool pprof cpu.prof
    // go tool pprof mem.prof
}
```

---

## Recovery Procedures

### Emergency Recovery

```go
func emergencyRecover() {
    executionIDs, err := execution.ListExecutions()
    if err != nil {
        log.Fatal("Failed to list:", err)
    }

    log.Printf("Found %d executions\n", len(executionIDs))

    for _, id := range executionIDs {
        state, err := execution.LoadState(id)
        if err != nil {
            log.Printf("❌ %s: %v", id, err)
            continue
        }

        log.Printf("✓ %s: %s (%s)", id, state.Task.Description, state.Status)

        // Check for stale running executions
        if state.IsRunning() {
            duration := time.Since(state.UpdatedAt)
            if duration > 1*time.Hour {
                log.Printf("⚠️  Stale execution: %s (running for %v)", id, duration)

                // Offer to interrupt
                fmt.Printf("Interrupt stale execution %s? (y/n): ", id)
                var response string
                fmt.Scanln(&response)

                if response == "y" {
                    ctx := context.Background()
                    selector := agent.NewSelector(agent.Config{}, nil)
                    cfg := execution.Config{}
                    engine := execution.New(cfg, selector)

                    err := engine.Interrupt(ctx, id)
                    if err != nil {
                        log.Printf("Interrupt failed: %v", err)
                    } else {
                        log.Printf("Interrupted %s", id)
                    }
                }
            }
        }
    }
}
```

### State Repair

```go
func repairState(executionID string) error {
    // Load corrupted state
    state, err := execution.LoadState(executionID)
    if err != nil {
        return err
    }

    // Attempt repairs
    repairs := 0

    // 1. Fix missing fields
    if state.Task == nil {
        state.Task = &execution.Task{
            Description: "Recovery task",
        }
        repairs++
    }

    if state.Context == nil {
        state.Context = execution.NewTaskContext(".")
        repairs++
    }

    // 2. Fix invalid timestamps
    if state.CreatedAt.IsZero() {
        state.CreatedAt = time.Now()
        repairs++
    }

    if state.StartedAt.IsZero() && state.Status == execution.ExecutionStatusRunning {
        state.StartedAt = time.Now()
        repairs++
    }

    // 3. Recalculate checksum
    oldChecksum := state.Checksum
    state.Checksum = ""
    state.Checksum = state.computeChecksum()
    if oldChecksum != state.Checksum {
        repairs++
    }

    // 4. Validate
    err = state.Validate()
    if err != nil {
        return fmt.Errorf("validation failed after repairs: %w", err)
    }

    // Save repaired state
    err = execution.SaveState(state)
    if err != nil {
        return fmt.Errorf("save failed: %w", err)
    }

    log.Printf("Repaired state: %d fixes applied", repairs)
    return nil
}
```

### Rollback Procedure

```go
func rollbackToCheckpoint(executionID, targetCheckpointID string) error {
    // Load target checkpoint
    target, err := execution.LoadState(targetCheckpointID)
    if err != nil {
        return fmt.Errorf("load checkpoint: %w", err)
    }

    // Validate target
    err = target.Validate()
    if err != nil {
        return fmt.Errorf("checkpoint invalid: %w", err)
    }

    // Delete corrupted execution
    err = execution.DeleteState(executionID)
    if err != nil {
        return fmt.Errorf("delete corrupted: %w", err)
    }

    // Save checkpoint as new execution
    target.ID = executionID
    err = execution.SaveState(target)
    if err != nil {
        return fmt.Errorf("save rollback: %w", err)
    }

    log.Printf("Rolled back %s to checkpoint %s", executionID, targetCheckpointID)
    return nil
}
```

---

## Getting Help

If you encounter issues not covered in this guide:

1. **Check Logs:**
   ```bash
   # Enable debug logging
   export LOG_LEVEL=debug

   # Run with verbose output
   go run main.go -v
   ```

2. **Validate State:**
   ```go
   validation, err := engine.ValidateExecutionState(executionID)
   if err != nil {
       log.Printf("Validation failed: %v", err)
   }
   log.Printf("Validation result: %+v", validation)
   ```

3. **Test Configuration:**
   ```go
   cfg := execution.Config{
       Logger: logger,
       EnableSummarization:  true,
       EnableAutoCheckpoint: true,
   }

   engine := execution.New(cfg, selector)

   status := engine.Status(ctx)
   log.Printf("Engine status: %+v", status)
   ```

4. **Report Issues:**
   - Include execution ID
   - Include relevant logs
   - Include error messages
   - Describe expected vs actual behavior

---

## See Also

- [Execution Engine v2 Documentation](EXECUTION_ENGINE_V2.md)
- [Execution Engine Examples](EXECUTION_ENGINE_EXAMPLES.md)
- [API Reference](EXECUTION_ENGINE_API.md)
