## Tasks

### Phase 1: BoltDB Layer (4h)

- [x] **1.1** Add fsnotify dependency
  - Files: `go.mod`
  - Dependencies: none
  - Effort: 0.5h

- [x] **1.2** Create BoltDB store with buckets
  - Files: `internal/spec/boltdb.go`
  - Buckets: changes, tasks, deps, blockers, runtime, meta
  - Dependencies: 1.1
  - Effort: 1.5h

- [x] **1.3** Implement CRUD operations
  - Files: `internal/spec/boltdb.go`
  - Methods: GetTask, PutTask, GetChange, PutChange, GetDeps, PutDeps
  - Dependencies: 1.2
  - Effort: 1h

- [x] **1.4** Implement full rebuild from markdown
  - Files: `internal/spec/boltdb.go`
  - Method: RebuildFromMarkdown(rootPath)
  - Dependencies: 1.3
  - Effort: 1h

### Phase 2: File Watcher (3h)

- [x] **2.1** Create file watcher with debounce
  - Files: `internal/spec/watcher.go`
  - Debounce: 100ms
  - Dependencies: 1.4
  - Effort: 1.5h

- [x] **2.2** Implement incremental sync
  - Files: `internal/spec/watcher.go`
  - Parse changed file, update BoltDB, recompute deps
  - Dependencies: 2.1
  - Effort: 1h

- [x] **2.3** Add error handling and recovery
  - Files: `internal/spec/watcher.go`
  - Parse errors, corruption detection, fallback
  - Dependencies: 2.2
  - Effort: 0.5h

### Phase 3: MCP Tools (3h)

- [x] **3.1** Create query tools
  - Files: `internal/mcp/tools/registry_list.go`, `registry_get.go`
  - Tools: registry_list_changes, registry_get_change, registry_list_tasks
  - Dependencies: 2.3
  - Effort: 1h

- [x] **3.2** Create next task and deps tools
  - Files: `internal/mcp/tools/registry_next.go`, `registry_deps.go`
  - Tools: registry_next_task, registry_deps
  - Dependencies: 3.1
  - Effort: 1h

- [x] **3.3** Create action tools
  - Files: `internal/mcp/tools/registry_actions.go`
  - Tools: registry_mark_done, registry_start_task, registry_sync
  - Dependencies: 3.2
  - Effort: 1h

### Phase 4: Integration (2h)

- [x] **4.1** Register MCP tools ✓ (2026-02-01)
  - Files: `internal/mcp/tools/register.go`
  - Register all 10 registry tools
  - Dependencies: 3.3
  - Effort: 0.5h

- [x] **4.2** Initialize BoltDB and watcher in server ✓ (2026-02-01)
  - Files: `internal/mcp/server/server.go`
  - Open BoltDB, start watcher, trigger initial sync
  - Dependencies: 4.1
  - Effort: 0.5h

- [x] **4.3** Update .gitignore and add tests ✓ (2026-02-01)
  - Files: `.gitignore`, `internal/spec/boltdb_test.go`
  - Add .state.db to gitignore, unit tests
  - Dependencies: 4.2
  - Effort: 1h

---

**Total Tasks**: 12
**Estimated Effort**: ~12 hours
