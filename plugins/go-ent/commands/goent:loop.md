---
description: Start autonomous work loop with self-correction
argument-hint: <task-description> [--max-iterations=10]
allowed-tools: Read, Write, Edit, Bash, Glob, Grep, mcp__plugin_serena_serena, mcp__goent__goent_loop_start, mcp__goent__goent_loop_get, mcp__goent__goent_loop_set, mcp__goent__goent_registry_next, mcp__goent__goent_registry_update
---

# Autonomous Loop

Start a self-correcting autonomous loop for task execution with automatic error recovery.

## Input

Parse `$ARGUMENTS`:
- Task description (required)
- `--max-iterations=N` (optional, default: 10)

Examples:
- `implement user authentication`
- `fix all failing tests --max-iterations=5`
- `add email validation to User entity`

## Loop Workflow

### 1. Initialize Loop

```
Use goent_loop_start:
  path="."
  task="$TASK_DESCRIPTION"
  max_iterations=$MAX_ITER
```

### 2. Execution Loop

Repeat until success, max iterations, or cancellation:

#### Iteration Steps:

1. **Get next task** (if no specific task given):
   ```
   Use goent_registry_next to get recommended task
   OR use the specific task from $ARGUMENTS
   ```

2. **Attempt execution**:
   - Load context using Serena
   - Implement solution
   - Run validation: `go build && go test -race`

3. **Check result**:
   - **SUCCESS**:
     - Mark loop completed: `goent_loop_set status=completed`
     - Output: `<promise>COMPLETE</promise>`
     - Exit loop

   - **FAILURE**:
     - Analyze error message
     - Record error: `goent_loop_set last_error="..."`
     - Determine adjustment strategy
     - Record adjustment: `goent_loop_set adjustment="..."`
     - Increment iteration: `goent_loop_set iteration=$((i+1))`
     - Continue to next iteration

4. **Between iterations**:
   - Wait 2-3 seconds (avoid tight loop)
   - Check if loop was cancelled: `goent_loop_get`
   - If status=cancelled, exit with message

### 3. Self-Correction Strategies

When encountering errors, try these adjustments:

| Error Type | Adjustment Strategy |
|------------|-------------------|
| Compile error | Fix syntax, imports, types |
| Test failure | Fix logic, update test expectations |
| Missing dependency | Add import, install package |
| Type mismatch | Adjust types, add conversions |
| Nil pointer | Add nil checks, initialize pointers |
| Race condition | Add mutex, use channels |

**Record each adjustment** so future iterations don't repeat failed approaches.

### 4. Loop Termination

Exit loop when ANY of these occur:
- ✅ Task completed successfully
- ⚠️ Max iterations reached
- 🛑 User cancelled loop
- ❌ Unrecoverable error (after 3 same-error iterations)

## State Persistence

Loop state saved at: `openspec/.loop-state.yaml`

```yaml
task: "implement user auth"
iteration: 3
max_iterations: 10
last_error: "test failure in auth_test.go:42"
adjustments:
  - "added mock for UserRepo"
  - "fixed import path"
  - "updated test assertion"
status: running
```

## Output Format

### During Loop
```
═══════════════════════════════════════════
AUTONOMOUS LOOP: {task}
═══════════════════════════════════════════

Iteration 1/10
  ├─ Action: Implementing feature
  ├─ Build: ✓ SUCCESS
  └─ Tests: ❌ FAIL (auth_test.go:42)

Error: expected "admin" got "user"
Adjustment: Fix role assignment logic

Iteration 2/10
  ├─ Action: Applied adjustment
  ├─ Build: ✓ SUCCESS
  └─ Tests: ✓ SUCCESS

<promise>COMPLETE</promise>
═══════════════════════════════════════════
```

### On Completion
```
✅ Task completed successfully
   Iterations: 2/10
   Adjustments made: 1
   Duration: 45s
```

### On Max Iterations
```
⚠️ Max iterations reached (10/10)
   Last error: {error}
   Adjustments attempted: {list}

Recommendation: Manual intervention required
Consider: {suggested next steps}
```

## Cancellation

User can stop loop with `/goent:loop-cancel`

Check for cancellation between iterations by reading loop state.

## Best Practices

- Start with simple tasks to test loop behavior
- Set reasonable max iterations (5-15)
- Monitor first iteration carefully
- Cancel if loop enters infinite adjustment cycle
- Use for repetitive fixes (linting, test failures, build errors)

## Integration with Registry

If task comes from registry:
1. Mark task as in_progress at loop start
2. Update task status if loop completes
3. Record adjustments in task notes

## Example Session

```bash
User: /goent:loop "fix all linting errors" --max-iterations=5

Agent: Starting autonomous loop...
       Task: fix all linting errors
       Max iterations: 5

Iteration 1: Run golangci-lint... 12 errors found
  Adjustment: Fix naming conventions in 3 files

Iteration 2: Run golangci-lint... 4 errors found
  Adjustment: Add error checks in handler.go

Iteration 3: Run golangci-lint... 0 errors found

<promise>COMPLETE</promise>

All linting errors fixed in 3 iterations.
```

## Guardrails

- NEVER modify critical files (go.mod, .git/) in loop
- ALWAYS run tests after changes
- STOP if same error occurs 3 times
- NEVER push to remote in autonomous mode
- ALWAYS document adjustments clearly
