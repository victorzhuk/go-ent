# Tasks: Add BoltDB Registry with state.md System

## 1. BoltDB Infrastructure ✅
- [x] Add go.etcd.io/bbolt dependency
- [x] Create internal/spec/boltdb.go with BoltStore
- [x] Implement 5 buckets (tasks, changes, deps, blocking, meta)
- [x] Implement O(1) GetTask, UpdateTask, ListTasks
- [x] Implement AddDependency with cycle detection
- [x] Implement GetBlockers with reverse index

## 2. State Generator ✅
- [x] Create internal/spec/state.go
- [x] Implement ParseDependencies with HTML comment regex
- [x] Implement GenerateChangeState (per-change state)
- [x] Implement GenerateRootState (aggregated view)
- [x] Implement WriteChangeStateMd
- [x] Implement WriteRootStateMd
- [x] Implement ParseTasksWithDependencies
- [x] Implement SyncFromTasksMd

## 3. MCP Tools ✅
- [x] Create internal/mcp/tools/state.go
- [x] Implement state_sync handler
- [x] Implement state_show handler
- [x] Register tools in register.go
- [x] Test tool signatures compile

## 4. Configuration ✅
- [x] Add openspec/registry.db to .gitignore
- [x] Add Store.RootPath() method

## 5. Registry Store Integration ✅
- [x] Update internal/spec/registry_store.go to use BoltStore <!-- depends: 1.6, 2.8 -->
- [x] Replace YAML Load/Save with BoltDB operations
- [x] Update RebuildFromSource to use ParseTasksWithDependencies
- [x] Preserve existing API compatibility
- [x] Update recalculateBlockedBy to use BoltDB reverse index

## 6. Migration Script
- [ ] Create cmd/migrate-registry/main.go <!-- depends: 5.1 -->
- [ ] Read existing registry.yaml
- [ ] Populate BoltDB from YAML data
- [ ] Preserve all metadata (deps, notes, assignees)
- [ ] Add --dry-run flag for preview
- [ ] Test migration with actual data

## 7. MCP Tool Updates
- [ ] Update registry_list to read from BoltDB <!-- depends: 5.1 -->
- [ ] Update registry_next to use BoltDB NextTasks
- [ ] Update registry_update to use BoltDB UpdateTask
- [ ] Update registry_deps to use BoltDB Add/RemoveDependency
- [ ] Update registry_sync to also generate state.md
- [ ] Add fallback to state.md if registry.db missing

## 8. Documentation
- [ ] Update openspec/AGENTS.md with new workflow <!-- depends: 5.5, 7.5 -->
- [ ] Document HTML comment dependency syntax
- [ ] Document state.md format
- [ ] Document BoltDB cache strategy
- [ ] Add examples of dependency usage
- [ ] Update workflow diagrams

## 9. Testing
- [x] Unit tests for BoltStore operations <!-- depends: 1.6 -->
- [x] Unit tests for StateStore parsing
- [ ] Integration test: tasks.md → BoltDB → state.md
- [ ] Test dependency parsing edge cases
- [ ] Test migration script with sample data
- [ ] Test MCP tools end-to-end

## 10. /task Command (Deferred)
- [ ] Create plugins/go-ent/commands/task.md
- [ ] Implement task execution logic
- [ ] Add ACP delegation support
- [ ] Integrate with state.md for task context
- [ ] Update tasks.md checkbox on completion
- [ ] Regenerate state.md after task done
