# Domain: OpenSpec

OpenSpec-specific rules for change management and task tracking.

## File Structure

```
openspec/
├── changes/
│   └── {change-id}/
│       ├── proposal.md       # Change proposal
│       ├── design.md         # Architecture/design
│       ├── specs/            # Detailed specifications
│       │   └── *.md
│       └── tasks.md          # Executable task list
└── registry.json             # Change registry
```

## Change IDs

Change IDs follow the pattern: `{YYYYMM}-{short-name}`

Examples:
- `202501-feature-auth`
- `202502-fix-bug-123`
- `202503-refactor-api`

## Registry Operations

### List Registry
```bash
/go-ent:registry list
```

### Show Active Change
```bash
/go-ent:status
```

### Archive Change
```bash
/go-ent:archive {change-id}
```

## Task Tracking in tasks.md

Task format:
```markdown
- [ ] **1.1** Add user entity [feature/auth] (P1)
  Depends on: 1.0
  Effort: 2h
  Files: internal/domain/user.go
```

Task states:
- `[ ]` - Not started
- `[x]` - Complete (with ✓ for sign-off)
- `[-]` - In progress (optional)

Task IDs: `{major}.{minor}` (e.g., `1.1`, `1.2`)

## Spec Validation

### Validate All Specs
```bash
make validate-specs
```

### Validate Single Spec
```bash
go run ./cmd/spec-validator validate openspec/changes/{id}/specs/{spec}.md
```

### Update Registry
```bash
make update-registry
```

## Proposal Template

```markdown
# Change: {title}

## Change ID
{YYYYMM}-{short-name}

## Type
{feature|enhancement|refactor|fix|docs}

## Description
{what and why}

## Breaking Changes
{yes|no}

## Dependencies
{list of other changes}
```

## Design Template

```markdown
# Design: {title}

## Overview
{high-level approach}

## Architecture
{components and their relationships}

## Data Flow
{how data moves through the system}

## API Changes
{endpoints affected}

## Database Changes
{migrations required}

## Alternatives Considered
{options and why they were rejected}
```

## Task Sign-Off Pattern

Mark task complete:
```markdown
- [x] **1.1** Add user entity [feature/auth] ✓
```

Include notes:
```markdown
- [x] **1.2** Write tests for user entity [feature/auth] ✓
  Tests: user_test.go (92% coverage)
```

## Task Dependencies

Dependencies form a directed acyclic graph (DAG):

```markdown
- [ ] **1.0** Create database schema [feature/auth]
- [ ] **1.1** Add user entity [feature/auth]
  Depends on: 1.0
- [ ] **1.2** Write tests for user entity [feature/auth]
  Depends on: 1.1
```

Valid dependency rules:
- Tasks can only depend on lower-numbered tasks
- No circular dependencies
- Parallel tasks cannot depend on each other

## Progress Calculation

Progress = `(completed tasks / total tasks) * 100%`

Example:
- Total tasks: 10
- Completed tasks: 7
- Progress: 70%

## Context Loading

When working on a task, always load:
1. `proposal.md` - Context and scope
2. `design.md` - Architecture decisions
3. `specs/*.md` - Detailed requirements
4. `tasks.md` - Task list and dependencies

## OpenSpec Commands Reference

| Command                      | Purpose                         |
|------------------------------|---------------------------------|
| `/go-ent:plan <description>` | Create new change proposal      |
| `/go-ent:apply`              | Execute next task from registry |
| `/go-ent:status`             | Show active change progress     |
| `/go-ent:registry list`      | List all changes                |
| `/go-ent:archive {id}`       | Archive completed change        |

## Spec Validation Rules

Specs must have:
- Clear acceptance criteria (GIVEN/WHEN/THEN or similar)
- Concrete scenarios with expected outcomes
- Cross-references to related specs
- Traceability to tasks that implement them

Example spec:
```markdown
## Requirement: User Authentication

### Scenario: User logs in with valid credentials
**Given** a user exists with email "user@example.com" and password "pass"
**When** the user logs in with those credentials
**Then** authentication succeeds and returns a token

Implemented by: tasks 1.2, 1.3
```

## Task Completion Checklist

Before marking a task complete:
- [ ] All acceptance criteria met
- [ ] Tests pass (including race detector)
- [ ] Coverage >= 80% for new code
- [ ] Code follows project conventions
- [ ] Linter passes
- [ ] Documentation updated (if applicable)
- [ ] Dependent tasks can now proceed
