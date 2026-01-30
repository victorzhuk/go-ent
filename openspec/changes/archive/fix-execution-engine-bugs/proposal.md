# Proposal: Fix Execution Engine Bugs

## Metadata
- **Change ID:** `fix-execution-engine-bugs`
- **Status:** Proposed
- **Type:** Bug Fix
- **Priority:** High
- **Affects Specs:** None (bug fixes)

## Problem

Code review of Execution Engine v2 implementation found three high-confidence bugs:

1. **Empty loop body in checkpoint cleanup** (`engine.go:589-597`)
   - The max checkpoint limit enforcement has an empty loop body
   - Logic to delete excess checkpoints is missing
   - Result: Max checkpoint limit is never enforced

2. **State tracking duplication** (`engine.go:727-747`)
   - `executePendingState()` calls `Execute()` which creates duplicate state
   - State is created twice for the same execution
   - Result: Memory waste and potential state inconsistencies

3. **Platform-specific syscalls without build constraints** (`opencode.go:164-184`)
   - Uses Unix-only `syscall.Setpgid` and `syscall.Kill` without build tags
   - Code will fail to compile on Windows
   - Result: Cross-platform build is broken

## Proposed Solution

### 1. Fix Checkpoint Cleanup Logic

Implement the missing deletion logic in the max checkpoint enforcement loop:
- Sort execution IDs by age (oldest first)
- Add excess checkpoints beyond the limit to `toDelete`
- Ensure we're deleting the oldest checkpoints first

### 2. Fix State Duplication

Refactor `executePendingState()` to avoid calling `Execute()`:
- Move execution logic inline or create a new internal method
- Prevent double state creation
- Maintain existing checkpoint behavior

### 3. Add Platform Build Constraints

Option A: Add build constraints to existing file
- Add `//go:build !windows` to `opencode.go`
- Create `opencode_windows.go` with Windows-compatible implementation

Option B: Split platform-specific code
- Move process management to separate files
- `opencode_unix.go` with `//go:build !windows`
- `opencode_windows.go` with Windows implementation

Recommend **Option A** for minimal changes.

## Impact

- **Breaking Changes:** None
- **API Changes:** None
- **Migration Required:** No
- **Testing Required:** Yes
  - Unit tests for checkpoint cleanup
  - Integration test for state tracking
  - Cross-platform build verification (`GOOS=windows go build ./...`)

## Risks

- **Low Risk:** These are localized bug fixes
- **Testing:** Need to verify Windows builds after fix
- **Dependencies:** None

## Alternatives Considered

1. **Disable max checkpoint limit** - Not viable, feature is needed
2. **Remove Windows support** - Not acceptable, cross-platform is requirement
3. **Keep state duplication** - Wastes memory, violates design principles

## Success Criteria

- [ ] Checkpoint cleanup enforces max limit correctly
- [ ] No duplicate state creation in execution flow
- [ ] Code compiles on Windows (`GOOS=windows go build ./...`)
- [ ] All existing tests pass
- [ ] New tests added for fixed bugs
