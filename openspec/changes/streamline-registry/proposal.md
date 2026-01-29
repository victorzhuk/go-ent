# Streamline Registry

## Summary

Simplify the task registry system by establishing `tasks.md` files as the single source of truth and reducing BoltDB to a runtime cache. Remove the redundant `registry.yaml` and simplify sync logic.

## Problem

The current registry system has three sources of truth:

1. **`openspec/changes/*/tasks.md`** - Markdown task lists (intended source of truth)
2. **`openspec/registry.yaml`** - YAML registry synced from tasks.md (redundant)
3. **`openspec/registry.db`** - BoltDB runtime state (assignments, sessions)

This creates:
- Sync complexity (tasks.md ↔ registry.yaml ↔ registry.db)
- Potential inconsistencies
- Confusion about which to read/write
- Double maintenance

## Solution

### New Architecture

```
┌─────────────────────────────────────────┐
│  tasks.md (Single Source of Truth)     │
│  - Task definitions                     │
│  - Task status (checked/unchecked)      │
│  - Dependencies                         │
└─────────────────┬───────────────────────┘
                  │ parse on startup
                  ▼
┌─────────────────────────────────────────┐
│  registry.db (Runtime Cache)           │
│  - Task assignments                     │
│  - Session tracking                     │
│  - Temporary state                      │
│  - Rebuilt from tasks.md on sync        │
└─────────────────────────────────────────┘
```

### Key Changes

| Aspect | Current | New |
|--------|---------|-----|
| Source of truth | tasks.md + registry.yaml | tasks.md only |
| Sync direction | Bidirectional | One-way: tasks.md → db |
| Registry.yaml | Maintained | Removed |
| Status storage | Both files | tasks.md checkboxes |
| Assignment | registry.yaml | registry.db only |

### tasks.md as Source of Truth

Task status is determined by checkbox state:

```markdown
## Tasks

### 1. Foundation
- [x] **1.1** Create domain entities ✓ (2026-01-15)
- [ ] **1.2** Implement repository
- [ ] **1.3** Add tests
```

- `[ ]` = pending
- `[x]` = completed
- In-progress tracked in registry.db only

### Simplified Sync

```go
// On startup or explicit sync
func Sync() error {
    // 1. Parse all tasks.md files
    tasks := parseTasksFromMarkdown()
    
    // 2. Rebuild registry.db (preserving assignments)
    db.rebuild(tasks, preserveAssignments=true)
    
    // 3. Done - no bidirectional sync
}
```

## Affected Systems

- `internal/spec/registry.go` - Simplify registry types
- `internal/spec/store.go` - Remove registry.yaml handling
- `internal/spec/workflow.go` - Simplify workflow state
- `internal/spec/state.go` - Remove redundant state tracking
- `internal/mcp/tools/registry.go` - Update to use tasks.md
- `openspec/registry.yaml` - Delete
- `openspec/registry.db` - Keep as runtime cache

## Breaking Changes

- [x] Remove registry.yaml
- [x] Change sync to one-directional
- [x] Simplify registry types
- [x] Update MCP tools

## Migration Path

1. Ensure all status is in tasks.md checkboxes
2. Remove registry.yaml generation
3. Update sync logic to one-directional
4. Update tools to read from tasks.md
5. Delete registry.yaml file

## Alternatives Considered

1. **Keep registry.yaml**: Rejected - redundant with tasks.md
2. **Use only registry.db**: Rejected - tasks.md is human-readable source of truth
3. **tasks.md as source + db as cache (chosen)**: Clean separation of concerns

## Success Criteria

- [ ] registry.yaml removed
- [ ] Sync simplified to one-directional
- [ ] tasks.md is single source of truth
- [ ] registry.db is runtime cache only
- [ ] MCP tools updated
- [ ] Tests updated and passing
- [ ] Documentation updated

## Effort Estimate

**~10 hours** across 8 tasks
