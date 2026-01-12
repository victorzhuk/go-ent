---
description: Debug and fix bugs with reproduction and root cause analysis
---

# Bug Fixing Workflow

Systematic debugging: reproduce → analyze → fix → validate.

## Input

`$ARGUMENTS`: bug description, issue ID, or error message

Examples:
- `/ent:bug "nil pointer in user service"`
- `/ent:bug #123` - Fix tracked issue
- `/ent:bug "panic in HTTP handler"`

## Agent Chain

| Agent                 | Purpose                                 | Tier     |
|-----------------------|-----------------------------------------|----------|
| @ent:reproducer       | Create minimal failing test             | fast     |
| @ent:researcher       | Code analysis, root cause investigation | fast     |
| @ent:debugger-fast    | Simple bugs, single component           | fast     |
| @ent:debugger         | Standard debugging, multiple components | standard |
| @ent:debugger-heavy   | Complex bugs (concurrency, performance) | heavy    |
| @ent:coder            | Implement fix                           | fast     |
| @ent:reviewer         | Code review                             | standard |
| @ent:tester           | Validate fix                            | fast     |
| @ent:acceptor         | Verify no regression                    | fast     |

**Escalation**: debugger-fast → debugger → debugger-heavy (when multi-component, concurrency, performance, or architecture changes needed)

---

## Workflow

### 1. Reproduce Bug

**Goal**: Create minimal, reliable reproduction

**Steps**:
1. Gather: error messages, stack traces, steps to reproduce, input data
2. Write failing test that reproduces the issue
3. Verify test fails consistently
4. Ensure test has clear assertion (expected vs actual)

**Output**: Failing test in relevant `_test.go` file

### 2. Root Cause Analysis

**Goal**: Understand the underlying issue

**Process**:
1. Analyze stack trace → find failure point
2. Understand data flow through relevant code
3. Form hypothesis about cause
4. Validate hypothesis

**Common root causes**:
| Pattern | Typical Cause |
|---------|---------------|
| Nil pointer | Missing initialization, unvalidated input |
| Race condition | Unprotected shared state |
| Index out of bounds | Off-by-one error, empty slice |
| Panic | Unhandled error, type assertion |
| Wrong result | Logic error, incorrect algorithm |

**Use Serena**:
- `find_symbol` - locate functions/types
- `find_referencing_symbols` - understand call chain
- `search_for_pattern` - find similar patterns

### 3. Implement Fix

**Goal**: Apply minimal fix addressing root cause

**For simple bugs** (fast agent):
- Single file changes
- Clear root cause
- Obvious solution

**For complex bugs** (standard/heavy agents):
- Multi-component issues
- Concurrency bugs
- Performance issues
- Architecture problems

**Implementation**:
1. Apply minimal fix addressing root cause (not symptoms)
2. Add defensive checks if needed
3. Update related code if necessary
4. Run: `go build && go test -race`

### 4. Validate Fix

**Goal**: Ensure fix works and no regressions

**Validation checklist**:
- [ ] Previously failing test now passes
- [ ] No regression in existing tests
- [ ] Edge cases covered
- [ ] Race detector passes (`-race` flag)
- [ ] Build succeeds
- [ ] Linter passes

**Code review**:
- Fix addresses root cause (not just symptoms)
- No new bugs introduced
- Error handling is proper
- Code is clear and maintainable
- Tests cover the fix

### 5. Complete

Update registry:
```
registry_update:
  task_id: "{change-id}/bug-{num}"
  status: "completed"
  notes: "Root cause: {cause}. Fix: {description}."
```

Update tasks.md:
```markdown
- [x] **bug-1** {description} ✓ {date}
  - Root cause: {explanation}
  - Fix: {solution}
```

---

## Output Format

```
═══════════════════════════════════════════
BUG FIX: {description}
═══════════════════════════════════════════

🔍 Reproduction:
   Test: {file}:{line}
   Status: ❌ FAILING (as expected)

🧠 Root Cause:
   Location: {file}:{line}
   Cause: {explanation}
   Impact: {scope}

🔨 Fix Applied:
   Files modified: {count}
   Changes: {description}
   Approach: {strategy}

🧪 Validation:
   Test: ✅ PASS
   All tests: ✅ PASS ({passed}/{total})
   Race detector: ✅ PASS
   Build: ✅ PASS

<promise>COMPLETE</promise>

Bug fixed and validated.
═══════════════════════════════════════════
```

---

## Bug Categories & Strategies

### Logic Errors
**Symptoms**: Wrong result, off-by-one, incorrect operator
**Fix**: Write test for edge case, fix logic

### Nil Pointer Dereference
**Symptoms**: Uninitialized variable, missing nil check
**Fix**: Add validation, initialize properly

### Concurrency Issues
**Symptoms**: Race condition, deadlock, data race
**Fix**: Add mutex, use channels correctly, fix synchronization

### Resource Leaks
**Symptoms**: Goroutine/file/connection/memory leak
**Fix**: Add cleanup, use defer, cancel contexts

### Integration Issues
**Symptoms**: API violation, constraint violation, timeout
**Fix**: Fix integration point, add retry, update dependencies

---

## Debugging Tools

**Built-in**:
- `go test -v` - Verbose output
- `go test -race` - Race detector
- `go test -cover` - Coverage analysis
- `GODEBUG=gctrace=1` - GC tracing

**External**:
- `dlv` (delve) - Go debugger
- `pprof` - Profiling
- `strace` - System call tracing

---

## Best Practices

1. **Always write failing test first**
   - Proves bug exists
   - Validates fix works
   - Prevents regression

2. **Find root cause, don't patch symptoms**
   - Understand why bug happened
   - Fix underlying issue
   - Prevent similar bugs

3. **Keep fix minimal**
   - Don't refactor while fixing
   - Fix one bug at a time
   - Separate concerns

4. **Validate thoroughly**
   - Run full test suite
   - Check for regressions
   - Use race detector
   - Test edge cases

---

## Integration with Registry

If bug tracked in registry:
1. Mark as `in_progress` at start
2. Update with root cause analysis
3. Mark `completed` with fix details
4. Record in task notes

---

## Guardrails

- ALWAYS write failing test before fixing
- NEVER guess at root cause - investigate
- ALWAYS run full test suite after fix
- NEVER skip race detector for concurrency bugs
- ALWAYS document what caused the bug
