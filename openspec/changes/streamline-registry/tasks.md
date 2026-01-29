## Tasks

### 1. Remove registry.yaml

- [ ] **1.1** Remove registry.yaml generation from sync
  - Files: `internal/spec/store.go`
  - Dependencies: none
  - Effort: 1h

- [ ] **1.2** Remove registry.yaml loading
  - Files: `internal/spec/store.go`, `internal/spec/registry.go`
  - Dependencies: 1.1
  - Effort: 1h

- [ ] **1.3** Delete registry.yaml file
  - Files: `openspec/registry.yaml`
  - Dependencies: 1.2
  - Effort: 0.5h

### 2. Simplify Registry Types

- [ ] **2.1** Remove Registry struct YAML fields
  - Files: `internal/spec/registry.go`
  - Dependencies: 1.3
  - Effort: 1h

- [ ] **2.2** Simplify RegistryTask to essential fields
  - Files: `internal/spec/registry.go`
  - Dependencies: 2.1
  - Effort: 1h

- [ ] **2.3** Remove unused registry types
  - Files: `internal/spec/registry.go`
  - Dependencies: 2.2
  - Effort: 1h

### 3. Update Store

- [ ] **3.1** Remove YAML persistence from Store
  - Files: `internal/spec/store.go`
  - Dependencies: 2.3
  - Effort: 1h

- [ ] **3.2** Simplify Store to BoltDB only
  - Files: `internal/spec/store.go`
  - Dependencies: 3.1
  - Effort: 1h

- [ ] **3.3** Update Store tests
  - Files: `internal/spec/store_config_test.go`, `internal/spec/boltdb_test.go`
  - Dependencies: 3.2
  - Effort: 1h

### 4. Simplify Sync Logic

- [ ] **4.1** Change sync to one-directional (tasks.md → db)
  - Files: `internal/spec/tracker.go`
  - Dependencies: 3.2
  - Effort: 2h

- [ ] **4.2** Remove bidirectional sync logic
  - Files: `internal/spec/tracker.go`, `internal/spec/sync.go` (if exists)
  - Dependencies: 4.1
  - Effort: 1h

- [ ] **4.3** Update sync tests
  - Files: `internal/spec/tracker_test.go`
  - Dependencies: 4.2
  - Effort: 1h

### 5. Update State Management

- [ ] **5.1** Simplify StateStore to runtime only
  - Files: `internal/spec/state.go`
  - Dependencies: 4.2
  - Effort: 1h

- [ ] **5.2** Remove state.md generation
  - Files: `internal/spec/state.go`
  - Dependencies: 5.1
  - Effort: 1h

- [ ] **5.3** Update state tests
  - Files: `internal/spec/state_test.go`
  - Dependencies: 5.2
  - Effort: 1h

### 6. Update MCP Tools

- [ ] **6.1** Update spec_sync tool for one-directional sync
  - Files: `internal/mcp/tools/spec.go`
  - Dependencies: 4.2
  - Effort: 1h

- [ ] **6.2** Remove registry_* tools (handled in consolidate-mcp-tools)
  - Files: none (already removed)
  - Dependencies: 6.1
  - Effort: 0h

### 7. Documentation

- [ ] **7.1** Update OPENSPEC_WORKFLOW.md with simplified registry
  - Files: `docs/OPENSPEC_WORKFLOW.md`
  - Dependencies: 6.1
  - Effort: 1h

- [ ] **7.2** Update ARCHITECTURE.md
  - Files: `docs/ARCHITECTURE.md`
  - Dependencies: 7.1
  - Effort: 1h

---

**Total Tasks**: 18
**Estimated Effort**: ~16 hours
