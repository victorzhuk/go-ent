# Tasks: Remove WHAT Comments

## Task Breakdown

### 1. Remove WHAT comments from execution package
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Audit all files in `internal/execution/` (excluding `*_test.go`)
2. Identify WHAT comments (explain what code does)
3. Remove WHAT comments while preserving:
   - WHY comments (counterintuitive behavior, legacy constraints)
   - Package documentation
   - Exported function/type documentation
4. Verify no accidental code removal

**Pattern to remove:**
```go
// Create|Initialize|Get|Set|Check|Validate|Parse|Build|Start|Stop|Update|Delete|Remove|Add|Load|Save|Read|Write|Execute|Run|Call|Return|Handle [something]
```

**Keep:**
```go
// Required by [reason]
// Counterintuitive: [explanation]
// Package execution provides...
// FunctionName does X and returns Y (exported functions)
```

**Validation:**
- [ ] `go build ./internal/execution` succeeds
- [ ] `go test ./internal/execution` passes
- [ ] `golint ./internal/execution` has no new warnings
- [ ] Manual review confirms no WHY comments removed

**Files Modified:**
- `internal/execution/engine.go`
- `internal/execution/runner.go`
- `internal/execution/opencode.go`
- `internal/execution/codemode.go`
- `internal/execution/cli.go`
- `internal/execution/budget.go`
- `internal/execution/sandbox.go`
- `internal/execution/parallel.go`
- `internal/execution/multi.go`
- `internal/execution/single.go`
- `internal/execution/strategy.go`
- `internal/execution/context.go`
- `internal/execution/fallback.go`
- `internal/execution/summarization.go`

---

### 2. Remove WHAT comments from config package
**Priority:** Medium
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Audit files in `internal/config/` (focus on main implementation files)
2. Remove WHAT comments using same criteria as Task 1
3. Preserve package docs and exported API docs

**Validation:**
- [ ] `go build ./internal/config` succeeds
- [ ] `go test ./internal/config` passes
- [ ] Manual review confirms no WHY comments removed

**Files Modified:**
- `internal/config/config.go`
- `internal/config/defaults.go`
- `internal/config/loader.go`
- `internal/config/providers.go`

---

### 3. Remove WHAT comments from mcp/tools package
**Priority:** Medium
**Estimated Complexity:** Medium
**Dependencies:** None

**Steps:**
1. Audit files in `internal/mcp/tools/` (excluding `*_test.go` initially)
2. Remove WHAT comments using same criteria
3. This package has more files, may need multiple passes

**Validation:**
- [ ] `go build ./internal/mcp/tools` succeeds
- [ ] `go test ./internal/mcp/tools` passes
- [ ] Manual review confirms no WHY comments removed

**Files Modified:**
- `internal/mcp/tools/register.go`
- `internal/mcp/tools/archetypes.go`
- `internal/mcp/tools/execution.go`
- `internal/mcp/tools/generate.go`
- `internal/mcp/tools/generate_component.go`
- `internal/mcp/tools/generate_from_spec.go`
- `internal/mcp/tools/archive.go`
- And other tool implementation files

---

### 4. Verify and document cleanup
**Priority:** Low
**Estimated Complexity:** Low
**Dependencies:** Tasks 1, 2, 3

**Steps:**
1. Run full build: `go build ./...`
2. Run full test suite: `go test ./...`
3. Run linter: `golangci-lint run` (if configured)
4. Scan for any remaining WHAT comments:
   ```bash
   rg -n "^[[:space:]]*// (Create|Initialize|Get|Set|Check|Validate|Parse|Build|Start|Stop|Update|Delete|Remove|Add|Load|Save|Read|Write|Execute|Run|Call|Return|Handle)" internal/
   ```
5. Review output to ensure only legitimate comments remain

**Validation:**
- [ ] All builds succeed
- [ ] All tests pass
- [ ] No WHAT comments in scoped files
- [ ] All WHY comments preserved

---

### 5. Optional: Add linter rule (future work)
**Priority:** Low
**Estimated Complexity:** Medium
**Dependencies:** Tasks 1-4 completed

**Steps:**
1. Research golangci-lint comment rules
2. Add custom linter configuration to prevent WHAT comments
3. Document the rule in project guidelines

**Note:** This is a follow-up task to prevent future violations. Not required for this proposal.

---

## Task Order

1. **Parallel:** Tasks 1, 2, 3 can be done independently
2. **Sequential:** Task 4 must follow all others
3. **Future:** Task 5 is optional enhancement

## Strategy

### Automated Removal (Safe)
Comments that are obvious WHAT statements:
- `// Create X` before `X := NewX()`
- `// Initialize Y` before `Y.Init()`
- `// Check if Z` before `if Z {`
- `// Get A from B` before `a := b.A`

### Manual Review Required
- Comments with multiple sentences
- Comments near complex logic
- Comments that might explain WHY disguised as WHAT
- Comments with words like "because", "since", "required"

### Never Remove
- Package documentation (`// Package X provides...`)
- Exported function docs (needed for godoc)
- WHY comments (counterintuitive, legacy, workarounds)
- TODO/FIXME comments (unless referencing internal IDs)
- Security comments (`// #nosec`, vulnerability explanations)

## Estimated Total Time

- Task 1: 1.5 hours (many files)
- Task 2: 30 minutes (fewer files)
- Task 3: 1 hour (many files)
- Task 4: 20 minutes (verification)
- **Total:** ~3 hours

## Testing Strategy

This is a pure comment removal task, so testing is minimal:
1. **Compilation test:** Code must still compile
2. **Test suite:** All tests must still pass
3. **Manual review:** No accidental code/doc removal
4. **Spot check:** Sample 5-10 files for quality
