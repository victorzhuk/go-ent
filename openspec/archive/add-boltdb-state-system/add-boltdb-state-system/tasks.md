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

## 6. Migration Script ✅
- [x] Create cmd/migrate-registry/main.go <!-- depends: 5.1 -->
- [x] Read existing registry.yaml
- [x] Populate BoltDB from YAML data
- [x] Preserve all metadata (deps, notes, assignees)
- [x] Add --dry-run flag for preview
- [x] Test migration with actual data

## 7. MCP Tool Updates ✅
- [x] Update registry_list to read from BoltDB <!-- depends: 5.1 -->
- [x] Update registry_next to use BoltDB NextTasks
- [x] Update registry_update to use BoltDB UpdateTask
- [x] Update registry_deps to use BoltDB Add/RemoveDependency
- [x] Update registry_sync to also generate state.md
- [x] Add fallback to state.md if registry.db missing

## 8. Documentation
- [x] Update openspec/AGENTS.md with new workflow ✅ <!-- depends: 5.5, 7.5 -->
- [x] Document HTML comment dependency syntax ✅
- [x] Document state.md format ✅
- [x] Document BoltDB cache strategy ✅
- [x] Add examples of dependency usage ✅
- [x] Update workflow diagrams ✅

## 9. Testing ✅
- [x] Unit tests for BoltStore operations <!-- depends: 1.6 -->
- [x] Unit tests for StateStore parsing
- [x] Integration test: tasks.md → BoltDB → state.md
- [x] Test dependency parsing edge cases
- [x] Test migration script with sample data
- [x] Test MCP tools end-to-end

 ## 10. /task Command
- [x] Create plugins/go-ent/commands/task.md ✅ 2026-01-17
- [x] Implement task execution logic ✅ 2026-01-17
- [x] Add ACP delegation support ✅ 2026-01-17
- [x] Integrate with state.md for task context ✅ 2026-01-17
- [x] Update tasks.md checkbox on completion ✅ 2026-01-17
- [x] Regenerate state.md after task done ✅ 2026-01-17

