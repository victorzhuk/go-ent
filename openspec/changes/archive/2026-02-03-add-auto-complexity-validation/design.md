# Design: Complexity Validation Fix

## Context

The go-ent project uses agent metadata files to define AI agents with various properties including `complexity`. The `debugger` agent uses `complexity: auto` to support dynamic model selection based on runtime complexity assessment, with `complexityHints` and `modelMapping` to guide the selection.

However, the validation function in `internal/cli/init.go` only accepts "simple", "standard", and "heavy" as valid complexity values, causing the debugger agent to fail validation.

**Constraints:**
- Validation must be backward compatible (existing agents must continue to validate)
- The "auto" complexity is a valid business concept for agents that dynamically adjust their approach
- The `debugger` agent is part of the core agent set and must validate successfully
- Test coverage must verify all valid complexity values

## Goals / Non-Goals

**Goals:**
- Add "auto" to the list of valid complexity values
- Update the validation error message to reflect all valid values
- Ensure the debugger agent validates successfully
- Maintain backward compatibility for existing agents

**Non-Goals:**
- Changing the semantics of "auto" complexity
- Implementing runtime complexity selection (that's handled elsewhere)
- Refactoring the validation architecture

## Decisions

### Decision 1: Add "auto" to validComplexity Map

**What:** Update the `validComplexity` map in the `validateAgent()` function to include `"auto": true`.

**Why:**
- Simple, minimal change that solves the immediate problem
- Maintains the existing validation pattern
- The "auto" complexity is already in use (debugger.yaml)
- Zero risk to existing functionality (backward compatible)

**Alternatives considered:**
1. **Remove complexity validation**: Too risky - would catch typos and invalid values
2. **Make complexity optional**: Would allow typos to slip through
3. **Create separate validation for "auto"**: Over-engineering for a simple enum addition

**Trade-offs:**
- ✅ Simple, one-line fix
- ✅ Maintains validation for typos and invalid values
- ✅ Backward compatible
- ⚠️ None identified

### Decision 2: Update Error Message

**What:** Update the validation error message from "complexity must be one of [simple, standard, heavy]" to "complexity must be one of [auto, simple, standard, heavy]".

**Why:**
- Error messages should reflect actual validation rules
- Developers need accurate guidance for valid values
- Maintains consistency with other enum validations in the same function

**Trade-offs:**
- ✅ Accurate error messages
- ✅ Consistent with existing patterns
- ⚠️ None identified

## Implementation Details

### Location

File: `internal/cli/init.go`
Function: `validateAgent()`
Lines: 447-452

### Current Code

```go
if meta.Complexity != "" {
    validComplexity := map[string]bool{"simple": true, "standard": true, "heavy": true}
    if !validComplexity[meta.Complexity] {
        return fmt.Errorf("%s: complexity must be one of [simple, standard, heavy] (got: %s)", filename, meta.Complexity)
    }
}
```

### Updated Code

```go
if meta.Complexity != "" {
    validComplexity := map[string]bool{"auto": true, "simple": true, "standard": true, "heavy": true}
    if !validComplexity[meta.Complexity] {
        return fmt.Errorf("%s: complexity must be one of [auto, simple, standard, heavy] (got: %s)", filename, meta.Complexity)
    }
}
```

### Changes Summary

1. Add `"auto": true` to the `validComplexity` map
2. Update error message to include "auto" in the list of valid values

## Verification

After implementation, verify:

1. **Validation passes for debugger agent**:
   ```bash
   ent validate
   ```
   Should succeed without errors

2. **Validation accepts all valid complexity values**:
   - Verify agents with "auto", "simple", "standard", "heavy" all pass

3. **Validation rejects invalid complexity values**:
   - Create test agent with invalid complexity (e.g., "medium")
   - Verify error message lists all valid values including "auto"

4. **Tests pass**:
   ```bash
   go test ./internal/cli/...
   ```

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| None identified | This is a simple, backward-compatible addition |

## Open Questions

None - design is complete and ready for implementation.
