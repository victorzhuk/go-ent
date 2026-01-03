---
description: Cancel running autonomous loop
allowed-tools: mcp__goent__goent_loop_cancel, mcp__goent__goent_loop_get
---

# Cancel Autonomous Loop

Stop the currently running autonomous loop immediately.

## Actions

1. **Check if loop is running**:
   ```
   Use goent_loop_get with path="."
   ```

2. **If loop exists**:
   ```
   Use goent_loop_cancel with path="."
   ```

3. **Output results**:
   - Show loop summary
   - Display iterations completed
   - List adjustments made
   - Show final state

## Output Format

### Loop Running
```
⏹️  Cancelling autonomous loop...

Loop Summary:
  Task: {task}
  Iterations: {iteration}/{max_iterations}
  Adjustments: {count}
  Status: CANCELLED

Last iteration:
  Error: {last_error}
  Adjustment: {last_adjustment}

═══════════════════════════════════════════
Loop state saved to: openspec/.loop-state.yaml
Resume later with: /goent:loop "{task}"
═══════════════════════════════════════════
```

### No Loop Found
```
ℹ️  No active loop found.

Use /goent:loop to start a new autonomous loop.
```

## Use Cases

- Loop is stuck in infinite adjustment cycle
- Wrong task was specified
- Need to manually intervene
- User wants to change approach
- Testing loop behavior during development

## After Cancellation

Loop state is preserved in `openspec/.loop-state.yaml` with status=cancelled.

To resume or restart:
- Review adjustments made: `cat openspec/.loop-state.yaml`
- Start new loop with refined task description
- Manually fix identified issues
- Delete state file to fully reset: `rm openspec/.loop-state.yaml`

## Safety

- Cancellation is immediate (no cleanup delays)
- Code changes made during loop are preserved
- Loop state saved for analysis
- No data loss occurs
