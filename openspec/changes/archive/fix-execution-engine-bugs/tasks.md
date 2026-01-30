# Tasks: Fix Execution Engine Bugs

## Task Breakdown

### 1. Fix checkpoint cleanup empty loop body
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Read `internal/execution/engine.go:589-597` to understand current logic
2. Implement sorting of execution IDs by age
3. Add excess checkpoints to `toDelete` slice
4. Write unit test verifying max checkpoint enforcement

**Validation:**
- [ ] Unit test passes for checkpoint limit enforcement
- [ ] Manual test: create 10 checkpoints with limit of 5, verify 5 oldest deleted

**Files Modified:**
- `internal/execution/engine.go`
- `internal/execution/engine_test.go` (new test)

---

### 2. Fix state tracking duplication
**Priority:** High
**Estimated Complexity:** Medium
**Dependencies:** None

**Steps:**
1. Analyze `executePendingState()` and `Execute()` to understand state creation
2. Refactor to prevent double state creation
3. Options:
   - Extract shared execution logic to new private method `executeWithState()`
   - Have `Execute()` check if state already exists before creating
4. Update checkpoint saving logic if needed
5. Write integration test verifying single state creation

**Validation:**
- [ ] Integration test confirms no duplicate states
- [ ] All existing execution tests pass
- [ ] Memory profiling shows reduced state allocation

**Files Modified:**
- `internal/execution/engine.go`
- `internal/execution/e2e_test.go` or `internal/execution/engine_test.go`

---

### 3. Add platform build constraints for syscalls
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Add `//go:build !windows` to `internal/execution/opencode.go`
2. Create `internal/execution/opencode_windows.go` with:
   - Same package and imports
   - `//go:build windows` constraint
   - Windows-compatible process management (no `Setpgid`, use `CreationFlags`)
3. Test compilation on both Unix and Windows

**Validation:**
- [ ] `go build ./...` succeeds on Linux/macOS
- [ ] `GOOS=windows go build ./...` succeeds
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` succeeds
- [ ] Existing opencode tests pass on Unix

**Files Modified:**
- `internal/execution/opencode.go` (add build constraint)
- `internal/execution/opencode_windows.go` (new file)

---

### 4. Verify all fixes
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** Tasks 1, 2, 3

**Steps:**
1. Run full test suite: `go test ./...`
2. Run with race detector: `go test -race ./internal/execution/...`
3. Build for all platforms:
   - `GOOS=linux go build ./...`
   - `GOOS=darwin go build ./...`
   - `GOOS=windows go build ./...`
4. Manual smoke test of execution engine
5. Update TASK_6.3.4_SUMMARY.md with fix details

**Validation:**
- [ ] All tests pass
- [ ] No race conditions detected
- [ ] Cross-platform builds succeed
- [ ] Documentation updated

**Files Modified:**
- `TASK_6.3.4_SUMMARY.md` (if exists)
- `docs/EXECUTION_ENGINE_V2.md` (add troubleshooting notes)

---

## Task Order

1. **Parallel:** Tasks 1, 2, 3 can be done independently
2. **Sequential:** Task 4 must follow all others

## Estimated Total Time

- Task 1: 30 minutes
- Task 2: 1 hour
- Task 3: 45 minutes
- Task 4: 30 minutes
- **Total:** ~3 hours

## Testing Strategy

1. **Unit Tests:** For checkpoint cleanup logic
2. **Integration Tests:** For state duplication fix
3. **Platform Tests:** Cross-compilation verification
4. **Regression Tests:** Ensure no existing functionality breaks
