---
name: ent:bug-flow
description: Debug and fix bugs with reproduction and root cause analysis
---

# Flow: Bug Fixing

{{include "context/generic.md"}}

Systematic debugging: reproduce -> analyze -> fix -> validate.

## Delegation Strategy

| Phase | Approach | If Subagent: Model + Type |
|-------|----------|---------------------------|
| Reproduce bug | Inline | -- |
| Root cause analysis | Inline or subagent | sonnet/opus, Explore |
| Fix strategy | Inline (debug-core skill) | -- |
| Implement fix | Inline (go-code skill) | sonnet, general-purpose |
| Validate fix | Inline (go-test skill) | -- |
| Acceptance | Inline or subagent | opus, Explore |

---

## Workflow

### Phase 1: Reproduce Bug

**Approach**: Inline

**Goal**: Create minimal, reliable reproduction

**Steps**:
1. Gather: error messages, stack traces, steps to reproduce, input data
2. Write failing test that reproduces the issue
3. Verify test fails consistently
4. Ensure test has clear assertion (expected vs actual)

**Output**: Failing test in relevant test file

### Phase 2: Root Cause Analysis

**Approach**: Inline (or subagent for deep investigation)

**Goal**: Understand the underlying issue

For complex root cause analysis, spawn a read-only subagent:
```
Task(model: "opus", subagent_type: "Explore",
  prompt: "Investigate root cause of {bug description}. Trace execution from {entry point} to failure at {location}. Use find_symbol and find_referencing_symbols to understand call chain.")
```

**Process**:
1. Analyze stack trace -> find failure point
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

**Use code navigation tools**:
- Find symbols — locate functions/types
- Find references — understand call chain
- Search patterns — find similar code

### Phase 3: Determine Fix Strategy

**Approach**: Inline (debug-core skill auto-activates)

**Goal**: Design the fix

**For simple bugs**:
- Single file changes
- Clear root cause
- Obvious solution

**For complex bugs** (spawn subagent if needed):
```
Task(model: "opus", subagent_type: "general-purpose",
  prompt: "Debug {concurrency/performance} issue at {location}. Root cause: {description}. Design and implement fix with proper synchronization/optimization.")
```

### Phase 4: Implement Fix

**Approach**: Inline

**Goal**: Apply the designed fix

**Steps**:
1. Implement minimal fix addressing root cause (not symptoms)
2. Add defensive checks if needed
3. Update related code if necessary
4. Run: build and test

### Phase 5: Validate Fix

**Approach**: Inline

**Goal**: Ensure fix works and no regressions

**Validation checklist**:
- [ ] Previously failing test now passes
- [ ] No regression in existing tests
- [ ] Edge cases covered
- [ ] Race detector passes
- [ ] Build succeeds
- [ ] Linter passes

**Code review** (for non-trivial fixes, spawn read-only subagent):
```
Task(model: "opus", subagent_type: "Explore",
  prompt: "Review the bug fix in {files}. Check: fix addresses root cause, no new bugs introduced, error handling is proper, tests cover the fix.")
```

### Phase 6: Acceptance

**Approach**: Inline or subagent

**Steps**:
1. Verify all tests pass
2. Check for regressions
3. Verify fix matches expected behavior
4. Sign off

**Outcome**:
- **ACCEPTED** -> Mark bug complete
- **NEEDS_WORK** -> Return to implementation

### Phase 7: Complete

Update tracking system:
- Mark bug as completed
- Add root cause analysis
- Document fix details

---

## Output Format

```
BUG FIX: {description}

Reproduction:
   Test: {file}:{line}
   Status: FAILING (as expected)

Root Cause:
   Location: {file}:{line}
   Cause: {explanation}
   Impact: {scope}

Fix Applied:
   Files modified: {count}
   Changes: {description}
   Approach: {strategy}

Validation:
   Test: PASS
   All tests: PASS ({passed}/{total})
   Race detector: PASS
   Build: PASS

COMPLETE

Bug fixed and validated.
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

## Best Practices

1. **Always write failing test first** — proves bug exists, validates fix, prevents regression
2. **Find root cause, don't patch symptoms** — understand why, fix underlying issue
3. **Keep fix minimal** — don't refactor while fixing, fix one bug at a time
4. **Validate thoroughly** — full test suite, race detector, edge cases

---

## Guardrails

- ALWAYS write failing test before fixing
- NEVER guess at root cause — investigate
- ALWAYS run full test suite after fix
- NEVER skip race detector for concurrency bugs
- ALWAYS document what caused the bug
