# Execution Engine v2 - Examples

This document provides practical, runnable examples demonstrating the execution engine v2 features.

## Table of Contents

1. [Basic Execution with Sandbox](#basic-execution-with-sandbox)
2. [Code-Mode Execution](#code-mode-execution)
3. [Automatic Context Summarization](#automatic-context-summation)
4. [Manual Checkpointing](#manual-checkpointing)
5. [Interrupting and Resuming Execution](#interrupting-and-resuming-execution)
6. [Complete Workflow with All Features](#complete-workflow-with-all-features)

---

## Basic Execution with Sandbox

### Example 1: Execute Task with Default Sandbox Limits

```go
package main

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/victorzhuk/go-ent/internal/agent"
    "github.com/victorzhuk/go-ent/internal/domain"
    "github.com/victorzhuk/go-ent/internal/execution"
)

func main() {
    ctx := context.Background()

    // Create logger
    logger := slog.Default()

    // Create agent selector
    selector := agent.NewSelector(agent.Config{}, nil)

    // Create engine with default config
    cfg := execution.Config{
        Logger:              logger,
        PreferredRuntime:      domain.RuntimeClaudeCode,
    }

    engine := execution.New(cfg, selector)

    // Create task
    task := &execution.Task{
        Description: "Analyze project structure",
        Type:        domain.TaskTypeDev,
        Context:     execution.NewTaskContext("/path/to/project"),
    }

    // Execute task
    result, err := engine.Execute(ctx, task)
    if err != nil {
        log.Fatal("Execution failed:", err)
    }

    fmt.Printf("Success: %v\n", result.Success)
    fmt.Printf("Output: %s\n", result.Output)
    fmt.Printf("Duration: %v\n", result.Duration)
}
```

### Example 2: Execute Task with Custom Sandbox Limits

```go
func main() {
    ctx := context.Background()
    logger := slog.Default()
    selector := agent.NewSelector(agent.Config{}, nil)

    // Create engine with sandbox limits
    cfg := execution.Config{
        Logger:              logger,
        PreferredRuntime:      domain.RuntimeClaudeCode,
    }

    engine := execution.New(cfg, selector)

    // Create task with resource limits
    task := &execution.Task{
        Description: "Process large dataset",
        Type:        domain.TaskTypeDev,
        Context:     execution.NewTaskContext("/path/to/project"),
        ForceRuntime: domain.RuntimeClaudeCode,
        Budget: &execution.BudgetLimit{
            MaxTokens: 100000,
            MaxCost:   0.50,
        },
    }

    result, err := engine.Execute(ctx, task)
    if err != nil {
        log.Fatal("Execution failed:", err)
    }

    fmt.Printf("Tokens: %d\n", result.TotalTokens())
    fmt.Printf("Cost: $%.4f\n", result.Cost)
}
```

### Example 3: Handle Resource Exceeded Errors

```go
func executeWithResourceLimits(ctx context.Context, engine *execution.Engine, task *execution.Task) (*execution.Result, error) {
    result, err := engine.Execute(ctx, task)
    if err != nil {
        // Check if resource limit exceeded
        var resourceErr *execution.ResourceExceededError
        if errors.As(err, &resourceErr) {
            log.Printf("Resource limit exceeded: %s (limit: %v)",
                resourceErr.Resource,
                resourceErr.Limit,
            )

            // Retry with higher limits
            if resourceErr.Resource == "memory" {
                task.Budget = &execution.BudgetLimit{
                    MaxTokens: 200000,
                }
                return engine.Execute(ctx, task)
            }
        }
        return nil, err
    }

    if !result.Success {
        log.Printf("Execution failed: %s", result.Error)
        return nil, fmt.Errorf("task failed: %s", result.Error)
    }

    return result, nil
}
```

---

## Code-Mode Execution

### Example 1: Execute JavaScript Code

```go
package main

import (
    "context"
    "log/slog"

    "github.com/victorzhuk/go-ent/internal/agent"
    "github.com/victorzhuk/go-ent/internal/domain"
    "github.com/victorzhuk/go-ent/internal/execution"
)

func main() {
    ctx := context.Background()
    logger := slog.Default()
    selector := agent.NewSelector(agent.Config{}, nil)

    cfg := execution.Config{
        Logger:          logger,
        PreferredRuntime: domain.RuntimeClaudeCode,
    }

    engine := execution.New(cfg, selector)

    // Create task with JavaScript code
    task := &execution.Task{
        Description: "Transform data using JavaScript",
        Type:        domain.TaskTypeDev,
        Context:     execution.NewTaskContext("/path/to/project"),
        ForceRuntime: domain.RuntimeClaudeCode,
        Metadata: map[string]interface{}{
            "code": `
                const data = read_file("input.json");
                const parsed = JSON.parse(data);

                const transformed = parsed.map(item => ({
                    id: item.id,
                    name: item.name.toUpperCase(),
                    value: item.value * 2
                }));

                write_file("output.json", JSON.stringify(transformed, null, 2));
                log("Transformed " + transformed.length + " items");
            `,
        },
    }

    result, err := engine.Execute(ctx, task)
    if err != nil {
        log.Fatal("Execution failed:", err)
    }

    log.Printf("Output: %s", result.Output)
}
```

### Example 2: Code-Mode with Resource Limits

```go
func main() {
    ctx := context.Background()
    logger := slog.Default()
    selector := agent.NewSelector(agent.Config{}, nil)

    cfg := execution.Config{
        Logger:          logger,
        PreferredRuntime: domain.RuntimeClaudeCode,
    }

    engine := execution.New(cfg, selector)

    // Create task with code and resource limits
    task := &execution.Task{
        Description: "Execute script with limits",
        Type:        domain.TaskTypeDev,
        Context:     execution.NewTaskContext("/path/to/project"),
        ForceRuntime: domain.RuntimeClaudeCode,
        Metadata: map[string]interface{}{
            "code": script,
            "limits": execution.ResourceLimits{
                MaxMemoryMB: 64,
                MaxCPUTime:  30 * time.Second,
                MaxExecTime: 60 * time.Second,
            },
        },
    }

    result, err := engine.Execute(ctx, task)
    if err != nil {
        log.Fatal("Execution failed:", err)
    }
}
```

### Example 3: Error Handling in Code-Mode

```go
func executeCodeWithRetry(ctx context.Context, engine *execution.Engine, code string) (*execution.Result, error) {
    task := &execution.Task{
        Description: "Execute JavaScript",
        Type:        domain.TaskTypeDev,
        Context:     execution.NewTaskContext("/path/to/project"),
        ForceRuntime: domain.RuntimeClaudeCode,
        Metadata: map[string]interface{}{
            "code": code,
        },
    }

    result, err := engine.Execute(ctx, task)
    if err != nil {
        return nil, err
    }

    if !result.Success {
        log.Printf("Code execution failed: %s", result.Error)

        // Check for specific errors
        if strings.Contains(result.Error, "resource limit") {
            log.Printf("Hit resource limit, retrying with more resources")
            task.Budget = &execution.BudgetLimit{
                MaxTokens: 50000,
            }
            return engine.Execute(ctx, task)
        }

        if strings.Contains(result.Error, "syntax") {
            return nil, fmt.Errorf("syntax error in code: %s", result.Error)
        }
    }

    return result, nil
}
```

---

## Automatic Context Summarization

### Example 1: Enable Automatic Summarization

```go
package main

import (
    "context"
    "log/slog"
    "time"

    "github.com/victorzhuk/go-ent/internal/agent"
    "github.com/victorzhuk/go-ent/internal/domain"
    "github.com/victorzhuk/go-ent/internal/execution"
)

func main() {
    ctx := context.Background()
    logger := slog.Default()
    selector := agent.NewSelector(agent.Config{}, nil)

    // Configure summarization
    cfg := execution.Config{
        Logger:              logger,
        PreferredRuntime:      domain.RuntimeClaudeCode,
        EnableSummarization:  true,
        SummarizationThreshold: execution.SummarizationThreshold{
            FileCount:     50,
            ContextLength: 50000,
            TokenCount:    10000,
        },
        SummarizationModel: "claude-3.5-sonnet",
    }

    engine := execution.New(cfg, selector)

    // Create task with many files
    task := &execution.Task{
        Description: "Analyze large project",
        Type:        domain.TaskTypeDev,
        Context:     execution.NewTaskContext("/path/to/large-project"),
    }

    // Add many files to context
    for i := 0; i < 60; i++ {
        task.Context.AddFile(fmt.Sprintf("src/file%d.go", i))
    }

    result, err := engine.Execute(ctx, task)
    if err != nil {
        log.Fatal("Execution failed:", err)
    }

    // Check if context was summarized
    if task.Context.IsSummarized {
        log.Printf("Context was summarized %d times",
            task.Context.SummarizationCount)
    }
}
```

### Example 2: Custom Summarization Thresholds

```go
func main() {
    ctx := context.Background()
    logger := slog.Default()
    selector := agent.NewSelector(agent.Config{}, nil)

    // Load custom thresholds from config file
    projectPath := "/path/to/project"
    threshold, err := execution.LoadSummarizationThreshold(projectPath)
    if err != nil {
        log.Printf("Using default thresholds: %v", err)
        threshold = execution.DefaultSummarizationThreshold()
    }

    cfg := execution.Config{
        Logger:                 logger,
        EnableSummarization:    true,
        SummarizationThreshold:  threshold,
        SummarizationModel:      "claude-3.5-sonnet",
    }

    engine := execution.New(cfg, selector)

    task := &execution.Task{
        Description: "Execute with custom thresholds",
        Type:        domain.TaskTypeDev,
        Context:     execution.NewTaskContext(projectPath),
    }

    result, err := engine.Execute(ctx, task)
    if err != nil {
        log.Fatal("Execution failed:", err)
    }
}
```

### Example 3: Manual Summarization Trigger

```go
func main() {
    ctx := context.Background()
    logger := slog.Default()
    selector := agent.NewSelector(agent.Config{}, nil)

    cfg := execution.Config{
        Logger:              logger,
        EnableSummarization: true,
    }

    engine := execution.New(cfg, selector)

    task := &execution.Task{
        Description: "Task with manual summarization",
        Type:        domain.TaskTypeDev,
        Context:     execution.NewTaskContext("/path/to/project"),
        Metadata: map[string]interface{}{
            "execution_id": "manual-summary-123",
        },
    }

    // Force summarization before execution
    executionID := "manual-summary-123"
    summarized := engine.TriggerSummarization(
        ctx,
        task,
        executionID,
        "claude-3.5-sonnet",
        nil,
        true, // force summarization
    )

    if summarized {
        log.Printf("Context manually summarized")
    }

    result, err := engine.Execute(ctx, task)
    if err != nil {
        log.Fatal("Execution failed:", err)
    }
}
```

### Example 4: Monitor Context Size During Execution

```go
func monitorContextSize(task *execution.Task, threshold execution.SummarizationThreshold) {
    // Check current context size
    fileCount := len(task.Context.Files)

    totalLength := 0
    for _, file := range task.Context.Files {
        totalLength += len(file)
    }

    estimatedTokens := totalLength / 4

    log.Printf("Context size: %d files, %d chars, ~%d tokens",
        fileCount, totalLength, estimatedTokens)

    // Check against thresholds
    if fileCount > threshold.FileCount {
        log.Printf("Warning: File count exceeds threshold (%d > %d)",
            fileCount, threshold.FileCount)
    }

    if totalLength > threshold.ContextLength {
        log.Printf("Warning: Context length exceeds threshold (%d > %d)",
            totalLength, threshold.ContextLength)
    }

    if estimatedTokens > threshold.TokenCount {
        log.Printf("Warning: Token count exceeds threshold (%d > %d)",
            estimatedTokens, threshold.TokenCount)
    }
}
```

---

## Manual Checkpointing

### Example 1: Create Manual Checkpoint

```go
package main

import (
    "context"
    "log/slog"

    "github.com/victorzhuk/go-ent/internal/agent"
    "github.com/victorzhuk/go-ent/internal/execution"
)

func main() {
    ctx := context.Background()
    logger := slog.Default()
    selector := agent.NewSelector(agent.Config{}, nil)

    cfg := execution.Config{
        Logger: logger,
    }

    engine := execution.New(cfg, selector)

    task := &execution.Task{
        Description: "Long-running task",
        Context:     execution.NewTaskContext("/path/to/project"),
    }

    // Create manual checkpoint
    state, err := engine.CreateManualCheckpoint(ctx, task, "running")
    if err != nil {
        log.Fatal("Failed to create checkpoint:", err)
    }

    log.Printf("Checkpoint created: %s", state.ID)
    log.Printf("Status: %s", state.Status)

    // Save execution ID for later resume
    executionID := state.ID
    log.Printf("Execution ID: %s (save this for resume)", executionID)
}
```

### Example 2: Load and Inspect Execution State

```go
func inspectExecutionState(executionID string) {
    state, err := execution.LoadState(executionID)
    if err != nil {
        log.Fatal("Failed to load state:", err)
    }

    fmt.Printf("Execution ID: %s\n", state.ID)
    fmt.Printf("Status: %s\n", state.Status)
    fmt.Printf("Task: %s\n", state.Task.Description)
    fmt.Printf("Agent: %s\n", state.Agent)
    fmt.Printf("Model: %s\n", state.Model)
    fmt.Printf("Runtime: %s\n", state.Runtime)
    fmt.Printf("Strategy: %s\n", state.Strategy)
    fmt.Printf("Created: %s\n", state.CreatedAt)
    fmt.Printf("Updated: %s\n", state.UpdatedAt)
    fmt.Printf("Duration: %v\n", state.Duration())

    if state.Result != nil {
        fmt.Printf("Success: %v\n", state.Result.Success)
        fmt.Printf("Tokens: %d\n", state.Result.TotalTokens())
        fmt.Printf("Cost: $%.4f\n", state.Result.Cost)
    }

    // Validate checksum
    if state.ValidateChecksum() {
        fmt.Println("Checksum: VALID")
    } else {
        fmt.Println("Checksum: INVALID")
    }
}
```

### Example 3: List All Executions

```go
func listAllExecutions() {
    executionIDs, err := execution.ListExecutions()
    if err != nil {
        log.Fatal("Failed to list executions:", err)
    }

    fmt.Printf("Found %d executions:\n", len(executionIDs))

    for _, id := range executionIDs {
        state, err := execution.LoadState(id)
        if err != nil {
            log.Printf("Failed to load %s: %v", id, err)
            continue
        }

        fmt.Printf("- %s: %s (%s)\n",
            id,
            truncate(state.Task.Description, 50),
            state.Status,
        )
    }
}
```

### Example 4: Cleanup Old Checkpoints

```go
func cleanupOldCheckpoints(engine *execution.Engine) {
    err := engine.CleanupOldCheckpoints()
    if err != nil {
        log.Printf("Cleanup failed: %v", err)
    } else {
        log.Printf("Cleanup completed successfully")
    }
}
```

### Example 5: Custom Checkpoint Management

```go
type CheckpointManager struct {
    engine *execution.Engine
}

func NewCheckpointManager(engine *execution.Engine) *CheckpointManager {
    return &CheckpointManager{engine: engine}
}

func (cm *CheckpointManager) CreateCheckpointWithMetadata(
    ctx context.Context,
    task *execution.Task,
    status string,
    metadata map[string]string,
) (*execution.ExecutionState, error) {
    state, err := cm.engine.CreateManualCheckpoint(ctx, task, status)
    if err != nil {
        return nil, err
    }

    // Add custom metadata
    for key, value := range metadata {
        state.SetMetadata(key, value)
    }

    // Save updated state
    err = execution.SaveState(state)
    if err != nil {
        return nil, err
    }

    return state, nil
}

func (cm *CheckpointManager) FindCheckpointsByMetadata(key, value string) ([]*execution.ExecutionState, error) {
    executionIDs, err := execution.ListExecutions()
    if err != nil {
        return nil, err
    }

    var results []*execution.ExecutionState
    for _, id := range executionIDs {
        state, err := execution.LoadState(id)
        if err != nil {
            continue
        }

        if metadataValue, ok := state.GetMetadata(key); ok && metadataValue == value {
            results = append(results, state)
        }
    }

    return results, nil
}
```

---

## Interrupting and Resuming Execution

### Example 1: Interrupt Running Execution

```go
package main

import (
    "context"
    "log/slog"
    "time"

    "github.com/victorzhuk/go-ent/internal/agent"
    "github.com/victorzhuk/go-ent/internal/execution"
)

func main() {
    ctx := context.Background()
    logger := slog.Default()
    selector := agent.NewSelector(agent.Config{}, nil)

    cfg := execution.Config{
        Logger:              logger,
        EnableAutoCheckpoint: true,
    }

    engine := execution.New(cfg, selector)

    task := &execution.Task{
        Description: "Long-running task",
        Context:     execution.NewTaskContext("/path/to/project"),
    }

    // Start execution in goroutine
    done := make(chan *execution.Result, 1)
    errChan := make(chan error, 1)

    go func() {
        result, err := engine.Execute(ctx, task)
        if err != nil {
            errChan <- err
            return
        }
        done <- result
    }()

    // Let it run for a while
    time.Sleep(10 * time.Second)

    // Interrupt execution
    state, err := engine.CreateManualCheckpoint(ctx, task, "running")
    if err != nil {
        log.Fatal("Failed to create checkpoint:", err)
    }

    executionID := state.ID

    log.Printf("Interrupting execution: %s", executionID)
    err = engine.Interrupt(ctx, executionID)
    if err != nil {
        log.Fatal("Failed to interrupt:", err)
    }

    log.Printf("Execution interrupted successfully")

    // Wait for goroutine to finish
    select {
    case result := <-done:
        log.Printf("Result: %v", result.Success)
    case err := <-errChan:
        log.Printf("Error: %v", err)
    }
}
```

### Example 2: Resume Interrupted Execution

```go
func resumeInterruptedExecution(ctx context.Context, engine *execution.Engine, executionID string) {
    log.Printf("Resuming execution: %s", executionID)

    result, err := engine.ResumeExecution(ctx, executionID)
    if err != nil {
        log.Fatal("Failed to resume:", err)
    }

    log.Printf("Resumed execution completed")
    log.Printf("Success: %v", result.Success)
    log.Printf("Output: %s", result.Output)
    log.Printf("Duration: %v", result.Duration)
    log.Printf("Tokens: %d", result.TotalTokens())
    log.Printf("Cost: $%.4f", result.Cost)
}
```

### Example 3: Multiple Interrupt/Resume Cycles

```go
func main() {
    ctx := context.Background()
    logger := slog.Default()
    selector := agent.NewSelector(agent.Config{}, nil)

    cfg := execution.Config{
        Logger:              logger,
        EnableAutoCheckpoint: true,
    }

    engine := execution.New(cfg, selector)

    task := &execution.Task{
        Description: "Task with multiple cycles",
        Context:     execution.NewTaskContext("/path/to/project"),
    }

    state, err := engine.CreateManualCheckpoint(ctx, task, "pending")
    if err != nil {
        log.Fatal("Failed to create checkpoint:", err)
    }

    executionID := state.ID

    // Execute with 3 interrupt/resume cycles
    for cycle := 1; cycle <= 3; cycle++ {
        log.Printf("=== Cycle %d ===", cycle)

        // Start execution
        result, err := engine.ResumeExecution(ctx, executionID)
        if err != nil {
            log.Printf("Resume error: %v", err)
            continue
        }

        log.Printf("Cycle %d output: %s", cycle, truncate(result.Output, 100))

        // Let it run briefly
        time.Sleep(2 * time.Second)

        // Interrupt
        err = engine.Interrupt(ctx, executionID)
        if err != nil {
            log.Printf("Interrupt error: %v", err)
            continue
        }

        log.Printf("Cycle %d interrupted", cycle)
    }

    // Final execution
    log.Printf("=== Final Execution ===")
    result, err := engine.ResumeExecution(ctx, executionID)
    if err != nil {
        log.Fatal("Final execution failed:", err)
    }

    log.Printf("Final result: %v", result.Success)
}
```

### Example 4: Validate State Before Resume

```go
func safeResume(ctx context.Context, engine *execution.Engine, executionID string) (*execution.Result, error) {
    // Validate state before resuming
    validation, err := engine.ValidateExecutionState(executionID)
    if err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    if !validation.Valid {
        return nil, fmt.Errorf("state invalid: %s", validation.Message)
    }

    if !validation.CanResume {
        return nil, fmt.Errorf("state cannot be resumed: %s", validation.Message)
    }

    log.Printf("State validated: %s", validation.Message)
    log.Printf("Checksum valid: %v", validation.ChecksumValid)
    log.Printf("Version compatible: %v", validation.VersionCompatible)

    // Resume execution
    return engine.ResumeExecution(ctx, executionID)
}
```

### Example 5: Handle Corrupted State

```go
func loadWithCorruptionCheck(executionID string) (*execution.ExecutionState, error) {
    state, err := execution.LoadState(executionID)
    if err != nil {
        var corruptedErr *execution.CorruptedStateError
        if errors.As(err, &corruptedErr) {
            log.Printf("Corrupted state: %s", corruptedErr.Error())

            if corruptedErr.CanRecover {
                log.Printf("Attempting recovery...")
                return attemptRecovery(executionID)
            } else {
                log.Printf("Cannot recover from corruption")

                // Prompt user for action
                fmt.Println("State is corrupted and cannot be recovered.")
                fmt.Printf("Delete corrupted state %s? (y/n): ", executionID)

                var response string
                fmt.Scanln(&response)

                if response == "y" {
                    selector := agent.NewSelector(agent.Config{}, nil)
                    cfg := execution.Config{}
                    engine := execution.New(cfg, selector)

                    err := engine.DeleteCorruptedState(executionID)
                    if err != nil {
                        return nil, fmt.Errorf("delete failed: %w", err)
                    }
                    log.Printf("Deleted corrupted state: %s", executionID)
                }

                return nil, err
            }
        }
        return nil, err
    }

    return state, nil
}

func attemptRecovery(executionID string) (*execution.ExecutionState, error) {
    // Implement recovery logic based on corruption type
    // This might involve:
    // - Partial data extraction
    // - Checksum recalculation
    // - Manual intervention

    return nil, fmt.Errorf("recovery not implemented")
}
```

---

## Complete Workflow with All Features

### Example: End-to-End Workflow

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/victorzhuk/go-ent/internal/agent"
    "github.com/victorzhuk/go-ent/internal/domain"
    "github.com/victorzhuk/go-ent/internal/execution"
)

func main() {
    ctx := context.Background()
    logger := slog.Default()
    selector := agent.NewSelector(agent.Config{}, nil)

    // Full configuration with all v2 features
    cfg := execution.Config{
        Logger:               logger,
        PreferredRuntime:      domain.RuntimeClaudeCode,
        EnableSummarization:  true,
        SummarizationThreshold: execution.SummarizationThreshold{
            FileCount:     50,
            ContextLength: 50000,
            TokenCount:    10000,
        },
        SummarizationModel:   "claude-3.5-sonnet",
        EnableAutoCheckpoint: true,
        MaxCheckpoints:       10,
        CheckpointAgeLimit:   24 * time.Hour,
    }

    engine := execution.New(cfg, selector)

    // Create task
    task := &execution.Task{
        Description: "Analyze and refactor large codebase",
        Type:        domain.TaskTypeDev,
        Context:     execution.NewTaskContext("/path/to/project"),
        ForceAgent:  domain.AgentRoleArchitect,
        ForceModel:  "claude-3.5-sonnet",
        Budget: &execution.BudgetLimit{
            MaxTokens: 500000,
            MaxCost:   5.00,
        },
    }

    // Create initial checkpoint
    state, err := engine.CreateManualCheckpoint(ctx, task, "pending")
    if err != nil {
        log.Fatal("Failed to create initial checkpoint:", err)
    }

    executionID := state.ID
    fmt.Printf("Execution ID: %s\n", executionID)

    // Setup signal handling for graceful interrupt
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Execute with interrupt support
    done := make(chan *execution.Result, 1)
    errChan := make(chan error, 1)

    go func() {
        result, err := engine.Execute(ctx, task)
        if err != nil {
            errChan <- err
            return
        }
        done <- result
    }()

    // Wait for completion or interrupt signal
    select {
    case result := <-done:
        fmt.Println("\n=== Execution Completed ===")
        fmt.Printf("Success: %v\n", result.Success)
        fmt.Printf("Output: %s\n", result.Output)
        fmt.Printf("Duration: %v\n", result.Duration)
        fmt.Printf("Tokens: %d\n", result.TotalTokens())
        fmt.Printf("Cost: $%.4f\n", result.Cost)

        // Check if context was summarized
        if task.Context.IsSummarized {
            fmt.Printf("Context summarized: %d times\n",
                task.Context.SummarizationCount)
        }

    case err := <-errChan:
        fmt.Printf("\n=== Execution Error ===")
        fmt.Printf("Error: %v\n", err)

        // Check for resource limit errors
        var resourceErr *execution.ResourceExceededError
        if errors.As(err, &resourceErr) {
            fmt.Printf("Resource limit exceeded: %s (limit: %v)\n",
                resourceErr.Resource,
                resourceErr.Limit)
        }

    case sig := <-sigChan:
        fmt.Printf("\n=== Received Signal: %v ===\n", sig)
        fmt.Println("Gracefully interrupting execution...")

        err := engine.Interrupt(ctx, executionID)
        if err != nil {
            log.Fatal("Failed to interrupt:", err)
        }

        fmt.Println("Execution interrupted successfully")
        fmt.Printf("Execution ID: %s (use engine_resume to continue)\n", executionID)

        // Show interrupt info
        state, err := execution.LoadState(executionID)
        if err == nil {
            fmt.Printf("Status: %s\n", state.Status)
            fmt.Printf("Updated: %s\n", state.UpdatedAt)
            fmt.Printf("Duration: %v\n", state.Duration())
        }
    }

    // Cleanup
    fmt.Println("\n=== Cleaning Up ===")
    err = engine.CleanupOldCheckpoints()
    if err != nil {
        log.Printf("Cleanup warning: %v", err)
    } else {
        fmt.Println("Cleanup completed")
    }
}
```

### Example: Resume After Process Restart

```go
func resumeAfterRestart() {
    ctx := context.Background()
    logger := slog.Default()
    selector := agent.NewSelector(agent.Config{}, nil)

    cfg := execution.Config{
        Logger:               logger,
        EnableSummarization:   true,
        EnableAutoCheckpoint:  true,
    }

    engine := execution.New(cfg, selector)

    // Get execution ID from user or config
    executionID := "abc123-def456-ghi789"

    fmt.Printf("Resuming execution: %s\n", executionID)

    // Validate before resume
    validation, err := engine.ValidateExecutionState(executionID)
    if err != nil {
        log.Fatal("Validation failed:", err)
    }

    fmt.Printf("Validation: %s\n", validation.Message)
    fmt.Printf("Can resume: %v\n", validation.CanResume)

    if !validation.CanResume {
        log.Fatal("Execution cannot be resumed")
    }

    // Load state to inspect
    state, err := execution.LoadState(executionID)
    if err != nil {
        log.Fatal("Failed to load state:", err)
    }

    fmt.Printf("Task: %s\n", state.Task.Description)
    fmt.Printf("Status: %s\n", state.Status)
    fmt.Printf("Agent: %s\n", state.Agent)
    fmt.Printf("Model: %s\n", state.Model)

    if state.Result != nil {
        fmt.Printf("Previous output length: %d\n", len(state.Result.Output))
    }

    // Resume execution
    fmt.Println("\nResuming execution...")
    result, err := engine.ResumeExecution(ctx, executionID)
    if err != nil {
        log.Fatal("Resume failed:", err)
    }

    fmt.Println("\n=== Execution Resumed Successfully ===")
    fmt.Printf("Success: %v\n", result.Success)
    fmt.Printf("Output: %s\n", result.Output)
    fmt.Printf("Duration: %v\n", result.Duration)
}
```

---

## Utility Functions

### Truncate String

```go
func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen] + "..."
}
```

### Format Duration

```go
func formatDuration(d time.Duration) string {
    if d < time.Minute {
        return fmt.Sprintf("%.2fs", d.Seconds())
    } else if d < time.Hour {
        return fmt.Sprintf("%.2fm", d.Minutes())
    } else {
        return fmt.Sprintf("%.2fh", d.Hours())
    }
}
```

### Format Cost

```go
func formatCost(cost float64) string {
    return fmt.Sprintf("$%.4f", cost)
}
```

---

## Running Examples

All examples can be compiled and run directly:

```bash
# Compile an example
go build -o example1 examples/basic_execution.go

# Run the example
./example1

# Or run directly
go run examples/basic_execution.go
```

---

## See Also

- [Execution Engine v2 Documentation](EXECUTION_ENGINE_V2.md)
- [API Reference](EXECUTION_ENGINE_API.md)
- [Troubleshooting Guide](EXECUTION_ENGINE_TROUBLESHOOTING.md)
