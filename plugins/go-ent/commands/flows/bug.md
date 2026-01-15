---
description: Debug and fix bugs with reproduction and root cause analysis
---

# Flow: Bug Fixing

{{include "domains/generic.md"}}

Systematic debugging: reproduce → analyze → fix → validate.

## Agent Chain

| Agent               | Phase                          | Tier     |
|---------------------|--------------------------------|----------|
| @ent:reproducer     | Create minimal failing test    | fast     |
| @ent:researcher     | Code analysis, investigation   | fast     |
| @ent:debugger-fast  | Simple bugs, single component  | fast     |
| @ent:debugger       | Standard debugging             | standard |
| @ent:debugger-heavy | Complex bugs                   | heavy    |
| @ent:coder          | Implement fix                  | fast     |
| @ent:reviewer       | Code review                    | standard |
| @ent:tester         | Validate fix                   | fast     |
| @ent:acceptor       | Verify no regression           | fast     |

**Escalation**: reproducer → researcher → debugger-fast/debugger/debugger-heavy → coder → reviewer → tester → acceptor

---

## Workflow

### Phase 1: Reproduce Bug

**Agent**: @ent:reproducer

**Goal**: Create minimal, reliable reproduction

**Steps**:
1. Gather: error messages, stack traces, steps to reproduce, input data
2. Write failing test that reproduces the issue
3. Verify test fails consistently
4. Ensure test has clear assertion (expected vs actual)

**Output**: Failing test in relevant test file

### Phase 2: Root Cause Analysis

**Agent**: @ent:researcher

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

**Use code navigation tools**:
- Find symbols - locate functions/types
- Find references - understand call chain
- Search patterns - find similar code

### Phase 3: Determine Fix Strategy

**Agent**: @ent:debugger-fast / @ent:debugger / @ent:debugger-heavy

**Goal**: Design and implement the fix

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
4. Run validation

### Phase 4: Implement Fix

**Agent**: @ent:coder

**Goal**: Apply the designed fix

**Steps**:
1. Implement minimal fix addressing root cause
2. Add defensive checks if needed
3. Update related code if necessary
4. Run: build and test

### Phase 5: Validate Fix

**Agent**: @ent:tester

**Goal**: Ensure fix works and no regressions

**Validation checklist**:
- [ ] Previously failing test now passes
- [ ] No regression in existing tests
- [ ] Edge cases covered
- [ ] Race detector passes
- [ ] Build succeeds
- [ ] Linter passes

**Code review**:
- Fix addresses root cause (not just symptoms)
- No new bugs introduced
- Error handling is proper
- Code is clear and maintainable
- Tests cover the fix

### Phase 6: Acceptance

**Agent**: @ent:acceptor

**Goal**: Final validation

**Steps**:
1. Verify all tests pass
2. Check for regressions
3. Verify fix matches expected behavior
4. Sign off

**Outcome**:
- **ACCEPTED** → Mark bug complete
- **NEEDS_WORK** → Return to @ent:coder

### Phase 7: Complete

Update tracking system:
- Mark bug as completed
- Add root cause analysis
- Document fix details

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

## Guardrails

- ALWAYS write failing test before fixing
- NEVER guess at root cause - investigate
- ALWAYS run full test suite after fix
- NEVER skip race detector for concurrency bugs
- ALWAYS document what caused the bug
