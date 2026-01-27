# Tasks: Reduce MCP Tools to Core Set (Phase 4)

## Task Breakdown

### 1. Delete agent tools (10 files)
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Delete `internal/mcp/tools/agent_*.go` files:
   - agent_list.go
   - agent_delegate.go
   - agent_execute.go
   - agent_select.go
   - agent_background.go
   - agent_status.go
   - agent_cancel.go
   - agent_complexity.go
   - agent_deps.go
   - agent_resolve.go
2. Verify no other files import these

**Validation:**
- [ ] 10 files deleted
- [ ] No imports remain

**Files Modified:**
- `internal/mcp/tools/agent_*.go` (deleted)

---

### 2. Delete AST tools (6 files)
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Delete `internal/mcp/tools/ast_*.go` files:
   - ast_parse.go
   - ast_query.go
   - ast_symbols.go
   - ast_references.go
   - ast_validate.go
   - ast_transform.go
2. Verify no other files import these

**Validation:**
- [ ] 6 files deleted
- [ ] No imports remain

**Files Modified:**
- `internal/mcp/tools/ast_*.go` (deleted)

---

### 3. Delete execution tools (7 files)
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Delete `internal/mcp/tools/execution_*.go` files:
   - execution_start.go
   - execution_status.go
   - execution_cancel.go
   - execution_list.go
   - execution_checkpoint.go
   - execution_resume.go
   - execution_cleanup.go
2. Verify no other files import these

**Validation:**
- [ ] 7 files deleted
- [ ] No imports remain

**Files Modified:**
- `internal/mcp/tools/execution_*.go` (deleted)

---

### 4. Delete generation tools (3 files)
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Delete `internal/mcp/tools/generate_*.go` files:
   - generate_agent.go
   - generate_skill.go
   - generate_command.go
2. Verify no other files import these

**Validation:**
- [ ] 3 files deleted
- [ ] No imports remain

**Files Modified:**
- `internal/mcp/tools/generate_*.go` (deleted)

---

### 5. Delete metrics tools (4 files)
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Delete `internal/mcp/tools/metrics_*.go` files:
   - metrics_list.go
   - metrics_get.go
   - metrics_export.go
   - metrics_reset.go
2. Verify no other files import these

**Validation:**
- [ ] 4 files deleted
- [ ] No imports remain

**Files Modified:**
- `internal/mcp/tools/metrics_*.go` (deleted)

---

### 6. Delete plugin tools (4 files)
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Delete `internal/mcp/tools/plugin_*.go` files:
   - plugin_list.go
   - plugin_load.go
   - plugin_unload.go
   - plugin_config.go
2. Verify no other files import these

**Validation:**
- [ ] 4 files deleted
- [ ] No imports remain

**Files Modified:**
- `internal/mcp/tools/plugin_*.go` (deleted)

---

### 7. Delete runtime tools (2 files)
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Delete `internal/mcp/tools/runtime_*.go` files:
   - runtime_list.go
   - runtime_switch.go
2. Verify no other files import these

**Validation:**
- [ ] 2 files deleted
- [ ] No imports remain

**Files Modified:**
- `internal/mcp/tools/runtime_*.go` (deleted)

---

### 8. Delete worker tools (8 files)
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Delete `internal/mcp/tools/worker_*.go` files:
   - worker_list.go
   - worker_spawn.go
   - worker_cancel.go
   - worker_status.go
   - worker_output.go
   - worker_health.go
   - worker_pool.go
   - worker_config.go
2. Verify no other files import these

**Validation:**
- [ ] 8 files deleted
- [ ] No imports remain

**Files Modified:**
- `internal/mcp/tools/worker_*.go` (deleted)

---

### 9. Delete meta tools (4 files)
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Delete `internal/mcp/tools/meta_*.go` files:
   - meta.go
   - discovery.go
   - search.go
   - inspect.go
2. Verify no other files import these

**Validation:**
- [ ] 4 files deleted
- [ ] No imports remain

**Files Modified:**
- `internal/mcp/tools/meta_*.go` (deleted)

---

### 10. Delete provider tools (2 files)
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Delete `internal/mcp/tools/provider_*.go` files:
   - provider_list.go
   - provider_config.go
2. Verify no other files import these

**Validation:**
- [ ] 2 files deleted
- [ ] No imports remain

**Files Modified:**
- `internal/mcp/tools/provider_*.go` (deleted)

---

### 11. Update tool registration
**Priority:** High
**Estimated Complexity:** High
**Dependencies:** Tasks 1-10

**Steps:**
1. Open `internal/mcp/tools/register.go`
2. Remove all registration calls for deleted tools:
   - Remove agent tool registrations
   - Remove AST tool registrations
   - Remove execution tool registrations
   - Remove generation tool registrations
   - Remove metrics tool registrations
   - Remove plugin tool registrations
   - Remove runtime tool registrations
   - Remove worker tool registrations
   - Remove meta tool registrations
   - Remove provider tool registrations
3. Keep only:
   - OpenSpec tool registrations (~10)
   - Skill tool registrations (~4)
   - State tool registrations (~1)
4. Simplify registration logic if needed
5. Verify file compiles

**Validation:**
- [ ] Only core tools registered
- [ ] File compiles
- [ ] No references to deleted tools

**Files Modified:**
- `internal/mcp/tools/register.go`

---

### 12. Update MCP server struct and initialization
**Priority:** High
**Estimated Complexity:** Medium
**Dependencies:** Task 11

**Steps:**
1. Open `internal/mcp/server/server.go`
2. Remove struct fields for deleted managers:
   - Remove agent manager
   - Remove execution manager
   - Remove worker manager
   - Remove metrics manager
   - Remove plugin manager
3. Update `New()` constructor - remove deleted manager initialization
4. Update `Start()` method - remove deleted manager startup
5. Update `Stop()` method - remove deleted manager cleanup
6. Simplify Server struct to only have:
   - Config
   - Spec registry
   - Skill registry
7. Verify server compiles

**Validation:**
- [ ] Deleted manager fields removed
- [ ] Initialization simplified
- [ ] Server compiles

**Files Modified:**
- `internal/mcp/server/server.go`

---

### 13. Verify remaining tools work
**Priority:** High
**Estimated Complexity:** Medium
**Dependencies:** Tasks 11-12

**Steps:**
1. Start MCP server: `go run cmd/go-ent/main.go`
2. Test OpenSpec tools:
   - spec_init
   - spec_list
   - spec_show
   - spec_create
   - spec_update
   - spec_delete
   - spec_registry_list
   - spec_workflow_status
   - spec_validate
   - spec_archive
3. Test Skill tools:
   - skill_list
   - skill_info
   - skill_validate
   - skill_quality
4. Test State tools:
   - state_sync
5. Verify all tools execute without errors

**Validation:**
- [ ] All core tools work
- [ ] Server lists only 15 tools
- [ ] No errors in tool execution

**Files Modified:**
- None (verification only)

---

### 14. Final verification
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** All previous tasks

**Steps:**
1. Run `make build` - must succeed
2. Run `make test` - all tests must pass
3. Run `make lint` - no lint errors
4. Count remaining tools: should be ~15
5. Verify tool count reduction: from 70+ to 15
6. Start server and list tools

**Validation:**
- [ ] Build succeeds
- [ ] Tests pass
- [ ] Lint clean
- [ ] Tool count ~15
- [ ] Server starts and lists only core tools

**Files Modified:**
- None (verification only)

---

## Task Order

**Parallel:** Tasks 1-10 can be done in parallel (independent deletions)
**Sequential:** Tasks 11-14 must be done after 1-10

## Estimated Total Time

- Tasks 1-10: 2 hours (delete 50+ files)
- Task 11: 2 hours (update registration)
- Task 12: 1.5 hours (update server)
- Task 13: 1 hour (verify tools)
- Task 14: 30 minutes (final verification)
- **Total:** ~7 hours

## Testing Strategy

1. **Compilation Check:** Verify build after registration update
2. **Tool Verification:** Test each remaining tool manually
3. **Server Start:** Verify server starts without errors
4. **Tool List:** Verify only core tools listed
5. **Integration:** Full make build/test/lint cycle
