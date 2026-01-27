# Proposal: Remove WHAT Comments

## Metadata
- **Change ID:** `remove-what-comments`
- **Status:** Proposed
- **Type:** Code Style
- **Priority:** Medium
- **Affects Specs:** None (style cleanup)

## Problem

Code review found numerous violations of the CLAUDE.md "ZERO COMMENTS" rule in recent code:

**Rule Violated:**
> Comments explaining WHAT code does = BAD NAMING. Fix the name instead.
> ONLY acceptable (rare): Comments explaining WHY - counterintuitive behavior, legacy requirements, etc.

**Files with violations:**
- `internal/execution/engine.go` - 40+ WHAT comments
- `internal/execution/runner.go` - 10+ WHAT comments
- `internal/execution/opencode.go` - 15+ WHAT comments
- `internal/execution/codemode.go` - 10+ WHAT comments
- `internal/execution/cli.go` - 8+ WHAT comments
- `internal/execution/budget.go` - 8+ WHAT comments
- `internal/execution/sandbox.go` - 6+ WHAT comments
- `internal/execution/parallel.go` - 10+ WHAT comments
- `internal/execution/multi.go` - 10+ WHAT comments
- `internal/execution/single.go` - 6+ WHAT comments
- `internal/execution/strategy.go` - 2+ WHAT comments
- `internal/execution/context.go` - 2+ WHAT comments
- `internal/execution/fallback.go` - 2+ WHAT comments
- `internal/execution/summarization.go` - 3+ WHAT comments
- `internal/config/*.go` - 10+ WHAT comments
- `internal/mcp/tools/*.go` - 30+ WHAT comments

**Total:** ~180+ WHAT comments to remove

## Examples of Violations

```go
// BAD: WHAT comment
// Create execution state for tracking
state := NewExecutionState(task)

// BAD: WHAT comment
// Build execution request
req := &ExecutionRequest{...}

// BAD: WHAT comment
// Check if timeout occurred
if ctx.Err() == context.DeadlineExceeded {
```

**ACCEPTABLE Comments (WHY):**
```go
// Required by legacy API - remove after v2 migration
resp.Header.Set("X-Legacy-Token", token)

// Counterintuitive: zero means unlimited per vendor docs
if limit == 0 {
```

## Proposed Solution

Remove all WHAT comments from the codebase that simply describe what the code does. Keep only:

1. **WHY comments** - Explain counterintuitive behavior, legacy requirements
2. **Package comments** - `// Package execution provides...`
3. **Exported function docs** - Required by golint for public APIs
4. **TODO/FIXME comments** - Action items (if they don't reference internal tickets)

## Scope

**In Scope:**
- All files in `internal/execution/` package
- All files in `internal/config/` package (only provider-related files)
- All files in `internal/mcp/tools/` package (excluding test files initially)

**Out of Scope:**
- Test files (can have explanatory comments for test cases)
- Generated code
- Third-party code
- Comments that explain WHY (keep these)

## Impact

- **Breaking Changes:** None (comments don't affect runtime behavior)
- **API Changes:** None
- **Migration Required:** No
- **Testing Required:** No (pure comment removal)

## Risks

- **Very Low Risk:** Only removing comments, no code changes
- **Review Needed:** Human review to ensure we don't accidentally remove WHY comments
- **CI/CD:** No impact, builds continue to work

## Implementation Strategy

1. **Automated removal** where safe (obvious WHAT comments)
2. **Manual review** for borderline cases
3. **Preserve** any WHY comments or genuinely useful explanations

## Success Criteria

- [ ] No WHAT comments remain in execution package
- [ ] No WHAT comments remain in config package
- [ ] No WHAT comments remain in mcp/tools package
- [ ] All WHY comments preserved
- [ ] Code compiles and tests pass (no accidental code removal)
- [ ] Documentation comments for public APIs remain intact

## Alternatives Considered

1. **Keep comments** - Violates project standards, creates inconsistency
2. **Improve naming instead** - Some cases, but most code is already well-named
3. **Add linter rule** - Good follow-up, but doesn't fix existing violations
