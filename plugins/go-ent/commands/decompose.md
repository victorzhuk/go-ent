---
description: Break proposal into dependency-aware, trackable tasks
argument-hint: <change-id>
---

# Task Decomposition

Transform OpenSpec change proposal into dependency-aware tasks with:
- Sequential task IDs (T001, T002...)
- Parallelization markers [P]
- Dependency graph
- File path associations

## Input

Change ID: $ARGUMENTS (from `openspec list`)

## Path Resolution

Change directory: `openspec/changes/$ARGUMENTS/`

For the steps below, `$CHANGE_ROOT` refers to `openspec/changes/$ARGUMENTS/`.

## Steps

1. Validate change exists: `openspec show $ARGUMENTS`
2. Resolve change directory path (see Path Resolution above)
3. Read change artifacts:
   - `$CHANGE_ROOT/proposal.md`
   - `$CHANGE_ROOT/design.md` (if exists)
   - Spec deltas in `$CHANGE_ROOT/specs/`
3. Extract requirements from spec deltas
4. Analyze existing `tasks.md` for current structure
5. Generate enhanced task structure:
   - Assign sequential task IDs starting from T001
   - Identify parallelizable tasks (different files, no dependencies)
   - Mark with [P] for parallel execution
   - Extract file paths from design.md or infer from requirements
   - Build dependency graph (which tasks block others)
6. Write enhanced `tasks.md` with format:

```markdown
# Implementation Tasks: <change-id>

## Dependencies
- T001 → T002, T003
- T002, T003 → T004 [P]

## Phase 1: Foundation

### T001: Create domain entities
- **Story**: specs/auth/spec.md#requirement-name
- **Files**: internal/domain/entity/foo.go
- **Depends**: None
- **Parallel**: No
- [ ] 1.1 Define entity structure
- [ ] 1.2 Add validation rules

### T002: Database migration [P]
- **Story**: specs/auth/spec.md#requirement-name
- **Files**: migrations/001_add_table.sql
- **Depends**: T001
- **Parallel**: With T003
- [ ] 2.1 Create migration file
```

## Validation

- Ensure all requirements have corresponding tasks
- Verify no circular dependencies in task graph
- Confirm [P] markers only on truly independent tasks
- Check file paths are realistic (existing or clearly new)

## Output

Enhanced `$CHANGE_ROOT/tasks.md` with task IDs, dependencies, and parallelization markers.

Output file: `openspec/changes/$ARGUMENTS/tasks.md`
