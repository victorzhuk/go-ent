# Tasks: CLI Cleanup and ACP Integration (Phase 5)

## Task Breakdown

### 1. Update cli/run.go to use ACP client
**Priority:** High
**Estimated Complexity:** High
**Dependencies:** None

**Steps:**
1. Read current `internal/cli/run.go` implementation
2. Replace execution.Engine imports with acp.Executor
3. Update run command logic:
   - Initialize ACP executor
   - Execute task via ACP
   - Stream output from ACP
   - Handle errors and status
4. Remove old execution engine code
5. Write integration test

**Validation:**
- [ ] Uses ACP client
- [ ] Task execution works
- [ ] Output streaming works
- [ ] Integration test passes

**Files Modified:**
- `internal/cli/run.go`
- `internal/cli/run_test.go`

---

### 2. Update cli/task.go to use ACP client
**Priority:** High
**Estimated Complexity:** High
**Dependencies:** None

**Steps:**
1. Read current `internal/cli/task.go` implementation
2. Replace worker.Manager imports with acp.Executor
3. Update task command logic:
   - Use acp.ExecuteTask for single tasks
   - Use acp.ExecuteTasks for parallel execution
   - Handle task status via ACP
   - Remove worker pool code
4. Update task list/status commands if needed
5. Write integration test

**Validation:**
- [ ] Uses ACP client
- [ ] Single task execution works
- [ ] Parallel execution works
- [ ] Integration test passes

**Files Modified:**
- `internal/cli/task.go`
- `internal/cli/task_test.go`

---

### 3. Delete cli/agent.go
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Delete `internal/cli/agent.go` file
2. Search for imports: `grep -r "cli.*agent" .`
3. Remove command registration from main.go
4. Verify build succeeds

**Validation:**
- [ ] File deleted
- [ ] Command registration removed
- [ ] Build succeeds

**Files Modified:**
- `internal/cli/agent.go` (deleted)
- `cmd/go-ent/main.go`

---

### 4. Delete cli/model.go
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Delete `internal/cli/model.go` file
2. Search for imports: `grep -r "cli.*model" .`
3. Remove command registration from main.go
4. Verify build succeeds

**Validation:**
- [ ] File deleted
- [ ] Command registration removed
- [ ] Build succeeds

**Files Modified:**
- `internal/cli/model.go` (deleted)
- `cmd/go-ent/main.go`

---

### 5. Delete cli/migrate.go
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Delete `internal/cli/migrate.go` file (if exists)
2. Search for imports: `grep -r "cli.*migrate" .`
3. Remove command registration from main.go
4. Verify build succeeds

**Validation:**
- [ ] File deleted
- [ ] Command registration removed
- [ ] Build succeeds

**Files Modified:**
- `internal/cli/migrate.go` (deleted)
- `cmd/go-ent/main.go`

---

### 6. Update cli/spec.go to use simplified spec package
**Priority:** High
**Estimated Complexity:** Medium
**Dependencies:** None

**Steps:**
1. Read current `internal/cli/spec.go` implementation
2. Update imports:
   - Change `internal/openspec` to `internal/spec`
   - Update any references to merged types
3. Verify all spec commands work:
   - spec init
   - spec list
   - spec show
   - spec create/update/delete
   - spec registry
   - spec workflow
   - spec validate
   - spec archive
4. Remove any dead code
5. Write tests if missing

**Validation:**
- [ ] Imports updated
- [ ] All commands work
- [ ] Build succeeds
- [ ] Tests pass

**Files Modified:**
- `internal/cli/spec.go`
- `internal/cli/spec_test.go`

---

### 7. Update cli/skill.go to use simplified skill package
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Read current `internal/cli/skill.go` implementation
2. Verify imports use `internal/skill`
3. Verify all skill commands work:
   - skill list
   - skill info
   - skill validate
   - skill quality
4. Remove any dead code
5. Write tests if missing

**Validation:**
- [ ] All commands work
- [ ] Build succeeds
- [ ] Tests pass

**Files Modified:**
- `internal/cli/skill.go`
- `internal/cli/skill_test.go`

---

### 8. Update cmd/go-ent/main.go initialization
**Priority:** High
**Estimated Complexity:** Medium
**Dependencies:** Tasks 1-7

**Steps:**
1. Read current `cmd/go-ent/main.go` implementation
2. Remove imports to deleted packages:
   - internal/execution
   - internal/worker
   - internal/agent
   - internal/openspec
   - internal/model
3. Remove initialization code for deleted systems
4. Add ACP client initialization if needed
5. Update command registration:
   - Remove agent command
   - Remove model command
   - Remove migrate command
6. Simplify overall initialization
7. Verify binary compiles and runs

**Validation:**
- [ ] No imports to deleted packages
- [ ] Initialization simplified
- [ ] Binary compiles
- [ ] Binary runs without panic
- [ ] All remaining commands work

**Files Modified:**
- `cmd/go-ent/main.go`

---

### 9. Integration test: /ent:plan → /ent:apply cycle
**Priority:** High
**Estimated Complexity:** High
**Dependencies:** All previous tasks

**Steps:**
1. Create integration test scenario
2. Test full workflow:
   - Create a change proposal via `/ent:plan`
   - Accept the proposal
   - Execute tasks via `/ent:apply`
   - Verify task execution via ACP
   - Check registry state consistency
3. Test with multiple tasks (sequential and parallel)
4. Verify state transitions
5. Check for any errors or inconsistencies

**Validation:**
- [ ] Can create proposal
- [ ] Can accept proposal
- [ ] Can execute tasks via ACP
- [ ] Registry state consistent
- [ ] No errors in workflow

**Files Modified:**
- `internal/cli/integration_test.go` (new)

---

### 10. Test parallel task execution
**Priority:** High
**Estimated Complexity:** Medium
**Dependencies:** Tasks 1-2

**Steps:**
1. Create test with 3+ independent tasks
2. Execute via `/ent:apply` or `task` command
3. Verify tasks execute in parallel via ACP
4. Verify all tasks complete successfully
5. Check for race conditions or state issues

**Validation:**
- [ ] Can execute multiple tasks
- [ ] Tasks run in parallel
- [ ] All tasks complete
- [ ] No race conditions

**Files Modified:**
- `internal/cli/task_test.go`

---

### 11. Final verification
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** All previous tasks

**Steps:**
1. Run `make build` - must succeed
2. Run `make test` - all tests must pass
3. Run `make lint` - no lint errors
4. Manual smoke test:
   - Start binary
   - Run `ent:plan` command
   - Accept proposal
   - Run `ent:apply` command
   - Verify execution
5. Check CLI help output - verify deleted commands removed
6. Verify package count: should be ~10 (from 27)
7. Verify LOC count: should be ~4,000 (from ~15,000)

**Validation:**
- [ ] Build succeeds
- [ ] All tests pass
- [ ] Lint clean
- [ ] Full workflow works
- [ ] Deleted commands removed from help
- [ ] Package count reduced to ~10
- [ ] LOC count reduced to ~4,000

**Files Modified:**
- None (verification only)

---

## Task Order

**Parallel:** Tasks 1-2 and 3-7 can be done in parallel
**Sequential:** Task 8 must be done after 1-7, tasks 9-11 after 8

## Estimated Total Time

- Tasks 1-2: 4 hours (ACP integration)
- Tasks 3-5: 1 hour (delete commands)
- Tasks 6-7: 2 hours (update commands)
- Task 8: 2 hours (main.go cleanup)
- Task 9: 2 hours (integration test)
- Task 10: 1 hour (parallel execution test)
- Task 11: 30 minutes (final verification)
- **Total:** ~12.5 hours

## Testing Strategy

1. **Unit Tests:** For individual command updates
2. **Integration Tests:** For full workflow cycle
3. **Parallel Execution:** Test concurrent task execution
4. **Smoke Tests:** Manual verification of key workflows
5. **Regression:** Ensure no existing functionality breaks

## Success Metrics

After Phase 5 completion:
- **Packages:** 10 (from 27) - 63% reduction
- **LOC:** ~4,000 (from ~15,000) - 73% reduction
- **MCP Tools:** 15 (from 70+) - 79% reduction
- **Complexity:** Dramatically simplified
- **Maintainability:** Significantly improved
