---
name: ent-openspec
description: "OpenSpec workflow patterns for change proposals, task tracking, and design documentation. Preloaded by workflow agents."
version: "1.0.0"
author: "go-ent"
disable-model-invocation: true
user-invocable: false
---

## Role

Guidance for working with OpenSpec change proposals, task tracking, and design documentation workflows.

## File Structure

```
openspec/changes/{change-id}/
├── proposal.md       # Full proposal with requirements
├── tasks.md          # Task breakdown with checkboxes
└── designs/          # Design documents (optional)
    ├── design.md     # Architecture decisions
    ├── components.md # Component diagrams
    └── data-flow.md  # Data flow documentation
```

## Key Files

**proposal.md:**
- Full change description
- Requirements and acceptance criteria
- Context and rationale
- Related changes/dependencies

**tasks.md:**
- Task breakdown with checkboxes
- Task numbering (e.g., `**1.1**`, `**1.2**`)
- Completion tracking with dates

Example task format:
```markdown
- [ ] **1.1** Create User entity

becomes after completion:

- [x] **1.1** Create User entity ✓ (2025-01-15)
```

## Workflow Steps

### 1. Read Proposal
```
read openspec/changes/{id}/proposal.md
```
Understand what, why, requirements, acceptance criteria

### 2. Read Tasks
```
read openspec/changes/{id}/tasks.md
```
Identify next incomplete task, dependencies, progress

### 3. Find Existing Patterns
Use Serena to understand codebase:
- `find_symbol` - Locate similar features
- `list_dir` - Understand structure
- `get_symbols_overview` - File organization

### 4. Implement Task
Follow conventions from:
- `ent-conventions` - Code style
- `ent-tooling` - Tool patterns
- Domain-specific skills (go-code, go-db, etc.)

### 5. Update Task Status
Mark complete with checkmark and date:
```markdown
- [x] **1.1** Create User entity ✓ (2025-01-15)
```

### 6. Verify Completion
```bash
go build ./...
go test ./... -race
make lint
```

## Task Completion Checklist

- [ ] Code implemented
- [ ] Tests passing
- [ ] Build successful
- [ ] Follows conventions
- [ ] Task marked complete with date
- [ ] No linter errors

## Design Documents

**When to Create:**
- New components requiring architecture decisions
- Complex data flows
- Database schema changes
- API contract changes
- Multi-component integrations

**Design Files:**
- `design.md` - Architecture decisions and rationale
- `components.md` - Component structure with diagrams
- `data-flow.md` - Data flow documentation

## OpenSpec Commands

- `/ent:plan` - Create new change proposals
- `/ent:apply` - Execute next task from registry
- `/ent:status` - View workflow state
- `/ent:registry list` - Show all tasks
- `/ent:archive <id>` - Archive completed change

## Best Practices

1. Always read proposal before starting
2. Follow existing patterns (don't reinvent)
3. Mark tasks complete as you go
4. Test after each task (not just at end)
5. Reference conventions for code style
6. Use Serena to understand codebase
7. Handoff appropriately for specialized work

See full workflow in `plugins/go-ent/agents/prompts/shared/_openspec.md`
