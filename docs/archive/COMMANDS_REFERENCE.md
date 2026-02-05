# Slash Commands Reference

Complete reference for go-ent slash commands (workflows).

---

## Table of Contents

- [Overview](#overview)
- [Workflow Commands](#workflow-commands)
- [Registry Commands](#registry-commands)
- [Utility Commands](#utility-commands)
- [Command Patterns](#command-patterns)

---

## Overview

go-ent provides **16 slash commands** for common development workflows. Commands are organized by category:

- **Workflow Commands** (3): Plan, execute, debug
- **Registry Commands** (4): Task management
- **Utility Commands** (4): Status, archive, sync
- **Management Commands** (5): Init, scaffold, lint, etc.

### Command Format

```
/ent:<command> [arguments]
```

**Examples**:
```
/ent:plan "Add user authentication"
/ent:task
/ent:bug "Login returns 500 error"
```

### OpenSpec Aliases

For **OpenSpec compatibility**, the following command aliases are available:

| go-ent Command | OpenSpec Alias | Description |
|----------------|----------------|-------------|
| `/ent:plan` | `/opsx:new` | Create new change proposal |
| `/ent:task` | `/opsx:apply` | Execute tasks from registry |
| `/ent:archive` | `/opsx:archive` | Archive completed change |

Both namespaces work identically - use whichever fits your workflow. The `/ent:*` namespace is primary for go-ent, while `/opsx:*` enables compatibility with OpenSpec-based tools and workflows.

**Examples using aliases**:
```
/opsx:new "Add two-factor authentication"
/opsx:apply
/opsx:archive add-2fa
```

---

## Workflow Commands

### `/ent:plan <description>`

**Full planning workflow**: clarify → research → design → decompose

**Purpose**: Create complete OpenSpec change proposal with research and task breakdown.

**Arguments**:
- `<description>`: Feature description or existing change ID

**Examples**:
```
/ent:plan "Add two-factor authentication with OTP"
/ent:plan "Implement Redis caching for user queries"
/ent:plan add-user-auth  # Continue existing change
```

**Agent Chain**:
1. `@ent:planner-fast` - Quick feasibility check
2. `@ent:architect` - System design (if architectural)
3. `@ent:planner` - Detailed planning, spec writing
4. `@ent:decomposer` - Task breakdown with dependencies

**Workflow Phases**:

1. **Assessment**: Quick feasibility check, complexity routing
2. **Clarification**: Resolve all unknowns before design
3. **Research**: Evaluate approaches, choose solution
4. **Design**: Architecture and implementation design
5. **Specification**: Write delta specs
6. **Decomposition**: Break into concrete tasks

**Output**:
- `openspec/changes/<change-id>/proposal.md`
- `openspec/changes/<change-id>/design.md` (if needed)
- `openspec/changes/<change-id>/specs/` (delta specs)
- `openspec/changes/<change-id>/tasks.md`

**Use When**:
- New features requiring design
- Breaking changes
- Architecture changes
- Multi-component changes
- Complex refactoring

**Skip For**:
- Simple bug fixes
- Typos, formatting
- Documentation only
- Configuration tweaks

---

### `/ent:task [task-id]`

**Execute OpenSpec tasks** with TDD and validation.

**Purpose**: Implement tasks from the registry with test-driven development.

**Arguments**:
- `[task-id]`: Specific task ID (optional, auto-selects if omitted)

**Examples**:
```
/ent:task                    # Auto-select next unblocked task
/ent:task add-auth-2.1      # Execute specific task
```

**Agent Chain**:
1. `@ent:task-fast` - Quick assessment, complexity routing
2. `@ent:task-heavy` - Complex task analysis (conditional)
3. `@ent:coder` - Code implementation
4. `@ent:reviewer` - Code review (conditional)
5. `@ent:tester` - Test validation
6. `@ent:acceptor` - Acceptance validation

**Workflow Phases**:

1. **Assessment**: Load task, assess complexity
2. **Deep Analysis**: Clarify complex requirements (if needed)
3. **Implementation**: TDD cycle (RED → GREEN → REFACTOR)
4. **Review**: Code quality review (if non-trivial)
5. **Testing**: Run tests, verify coverage
6. **Acceptance**: Validate against spec
7. **Complete**: Update registry and tasks.md

**Task Selection Logic**:
- If task-id provided: Execute that task
- If omitted: Auto-select from registry
  1. Filter by status (pending only)
  2. Check dependencies satisfied
  3. Sort by priority
  4. Return highest priority

**Use When**:
- Implementing planned changes
- Following TDD workflow
- Incremental development

---

### `/ent:bug <description>`

**Debug-focused workflow**: reproduce → research → fix → test

**Purpose**: Investigate and fix bugs with systematic approach.

**Arguments**:
- `<description>`: Bug description

**Examples**:
```
/ent:bug "Auth middleware returns 500 instead of 401"
/ent:bug "User creation fails with duplicate email"
```

**Agent Chain**:
1. `@ent:debugger-fast` - Quick assessment
2. `@ent:debugger` - Standard debugging
3. `@ent:debugger-heavy` - Complex debugging (if needed)
4. `@ent:reproducer` - Minimal reproduction
5. `@ent:coder` - Fix implementation
6. `@ent:tester` - Regression test

**Workflow Phases**:

1. **Reproduce**: Create minimal reproduction
2. **Research**: Investigate root cause
3. **Fix**: Implement solution
4. **Test**: Verify fix and add regression test

**Output**:
- `openspec/changes/<bug-id>/proposal.md` (minimal)
- `openspec/changes/<bug-id>/tasks.md`
- Regression test
- Bug fix commit

**Use When**:
- Fixing bugs
- Need to understand failure first
- Creating regression tests

---

## Registry Commands

### `/ent:registry list [filters]`

**List all tasks** in the registry.

**Purpose**: View tasks across all changes with optional filtering.

**Arguments**:
- `[filters]`: Optional filters (status, priority, change_id)

**Examples**:
```
/ent:registry list
/ent:registry list status=pending
/ent:registry list priority=high
/ent:registry list change_id=add-auth
```

**Output**:
```
Tasks:
1. [pending] add-auth/1: Implement JWT validation (priority: high)
2. [in_progress] add-auth/2: Add middleware (priority: medium)
3. [completed] add-auth/3: Write tests (priority: medium)
4. [pending] cache/1: Add Redis client (priority: low, blocked by: add-auth/1)
```

**Filters**:
- `status`: `pending`, `in_progress`, `completed`
- `priority`: `low`, `medium`, `high`
- `change_id`: Specific change ID

---

### `/ent:registry next [count]`

**Get next recommended task(s)**.

**Purpose**: Find next available task(s) to work on.

**Arguments**:
- `[count]`: Number of tasks to return (default: 1)

**Examples**:
```
/ent:registry next         # Next 1 task
/ent:registry next 3       # Next 3 tasks
```

**Selection Algorithm**:
1. Filter by status (pending only)
2. Check dependencies satisfied
3. Sort by priority (high → medium → low)
4. Return top N

**Output**:
```
Next task:
ID: add-auth/1
Title: Implement JWT validation
Priority: high
Dependencies: none
Estimated: 2h
Files: internal/auth/jwt.go
```

---

### `/ent:registry update <task-id> <field=value>`

**Update task** status or metadata.

**Purpose**: Manually update task information.

**Arguments**:
- `<task-id>`: Task identifier
- `<field=value>`: Field to update

**Examples**:
```
/ent:registry update add-auth/1 status=completed
/ent:registry update add-auth/2 priority=high
/ent:registry update cache/1 assignee=@user
```

**Updateable Fields**:
- `status`: `pending`, `in_progress`, `completed`
- `priority`: `low`, `medium`, `high`
- `assignee`: Agent or user ID
- `notes`: Additional notes

---

### `/ent:registry deps <task-id> <operation>`

**Manage task dependencies**.

**Purpose**: Add, remove, or list task dependencies.

**Arguments**:
- `<task-id>`: Task identifier
- `<operation>`: `add`, `remove`, `list`

**Examples**:
```
/ent:registry deps add-auth/2 list
/ent:registry deps add-auth/2 add add-auth/1
/ent:registry deps add-auth/2 remove add-auth/1
```

**Output** (for `list`):
```
Dependencies for add-auth/2:
- add-auth/1 (completed) ✓
- cache/1 (in_progress) ⏳
```

---

## Utility Commands

### `/ent:status`

**View workflow state** and progress.

**Purpose**: Get overview of all changes and tasks.

**Arguments**: None

**Output**:
```
═══════════════════════════════════════════
PROJECT STATUS
═══════════════════════════════════════════

Active Changes: 2

1. add-auth (in_progress)
   Tasks: 3/5 complete (60%)
   - [✓] 1: Implement JWT validation
   - [✓] 2: Add middleware
   - [✓] 3: Write tests
   - [ ] 4: Update documentation
   - [ ] 5: Add integration tests

2. cache (pending)
   Tasks: 0/4 complete (0%)
   Blocked by: add-auth/1

Next Available: add-auth/4 (priority: medium)

Budget: $12.34 / $100.00 daily (12%)
═══════════════════════════════════════════
```

---

### `/ent:archive <change-id>`

**Archive completed change**.

**Purpose**: Merge delta specs and move change to archive.

**Arguments**:
- `<change-id>`: Change identifier

**Examples**:
```
/ent:archive add-auth
/ent:archive add-user-caching
```

**Process**:
1. Validates all tasks complete
2. Merges delta specs into main specs
   - ADDED sections → append
   - MODIFIED sections → replace
   - REMOVED sections → delete
3. Moves change to `changes/archive/YYYY-MM-DD-<change-id>/`
4. Updates change index
5. Creates archive commit

**Validation Checks**:
- All tasks marked completed
- All acceptance criteria met
- No TODO/TBD markers
- Specs validated

---

### `/ent:registry sync`

**Sync registry** with tasks.md files.

**Purpose**: Update registry from all `tasks.md` files in changes.

**Arguments**: None

**Process**:
1. Scan all `openspec/changes/*/tasks.md`
2. Parse task definitions
3. Update registry
4. Preserve status and metadata
5. Detect new/removed tasks

**Output**:
```
Synced tasks from 3 changes:
- add-auth: 5 tasks (3 new, 2 updated)
- cache: 4 tasks (4 new)
- optimize: 2 tasks (1 removed, 1 updated)

Total: 11 tasks in registry
```

---

### `/ent:skill-sync`

**Sync skills** from plugins to Claude skills directory.

**Purpose**: Update Claude Code's skill registry with plugin skills.

**Arguments**: None

**Process**:
1. Find all skills in `plugins/*/skills/`
2. Copy to `.claude/skills/`
3. Update skill index
4. Preserve quality scores

**Output**:
```
Synced 15 skills:
✓ go-code (quality: 92)
✓ go-api (quality: 95)
✓ go-arch (quality: 94)
...

Updated .claude/skills/index.yaml
```

---

## Management Commands

### `/ent:init --tool=<runtime>`

**Initialize tool configuration** with agents.

**Purpose**: Set up go-ent for Claude Code or OpenCode.

**Arguments**:
- `--tool`: `claude`, `opencode`, or `all` (required)
- `--agents`: Comma-separated agent names
- `--include-deps`: Auto-resolve dependencies
- `--dry-run`: Preview changes
- `--force`: Overwrite existing config

**Examples**:
```
/ent:init --tool=claude
/ent:init --tool=opencode --agents=architect,coder,tester
/ent:init --tool=all --include-deps
/ent:init --tool=claude --dry-run
```

**Creates** (for Claude Code):
- `.claude/agents/` - Agent definitions
- `.claude/skills/` - Skill definitions
- `.go-ent/config.yaml` - Project configuration

---

### `/ent:scaffold <type> <name>`

**Generate components** via agents.

**Purpose**: Scaffold code using agent delegation.

**Arguments**:
- `<type>`: Component type
- `<name>`: Component name

**Types**:
- `entity`: Domain entity
- `repository`: Repository with interface
- `usecase`: Use case with tests
- `handler`: HTTP handler
- `service`: Full service (all layers)

**Examples**:
```
/ent:scaffold entity User
/ent:scaffold repository User
/ent:scaffold usecase CreateUser
/ent:scaffold handler UserHandler
/ent:scaffold service Order
```

**Delegates to**: `@ent:coder`

---

### `/ent:lint`

**Run linters**.

**Purpose**: Execute Go linters on codebase.

**Arguments**: None

**Runs**:
- `golangci-lint`
- `gofumpt`
- Custom validators

**Output**:
```
Running linters...

✓ golangci-lint: PASS
✓ gofumpt: PASS
✓ go vet: PASS

No issues found.
```

---

### `/ent:gen`

**Generate code** from OpenAPI/Proto.

**Purpose**: Run code generators.

**Arguments**: None

**Generates**:
- OpenAPI → Go client/server (ogen)
- Proto → Go code (protoc-gen-go)
- SQL → Go types (sqlc)

**Output**:
```
Generating code...

✓ OpenAPI → internal/api/generated/
✓ Proto → internal/grpc/generated/
✓ SQL → internal/repository/generated/

3 generators completed successfully.
```

---

### `/ent:tdd <feature>`

**Red-Green-Refactor** TDD cycle.

**Purpose**: Guided test-driven development.

**Arguments**:
- `<feature>`: Feature description

**Examples**:
```
/ent:tdd "User authentication"
/ent:tdd "Rate limiting middleware"
```

**Cycle**:
1. **RED**: Write failing test
2. **GREEN**: Minimal implementation
3. **REFACTOR**: Clean up code

**Delegates to**: `@ent:tester` → `@ent:coder`

---

## Command Patterns

### Workflow Pattern

Commands follow a consistent workflow pattern:

```
1. Parse arguments
2. Load context (specs, tasks, code)
3. Select agent chain
4. Execute phases
5. Validate output
6. Update state (registry, tasks.md)
7. Report results
```

### Agent Delegation

Commands delegate to specialized agents:

```
Command → Fast Agent (triage)
    ↓
  Standard Agent (execution)
    ↓
  Heavy Agent (complex cases)
```

**Example**:
```
/ent:plan → @ent:planner-fast
              ↓
            @ent:architect (if architectural)
              ↓
            @ent:planner
              ↓
            @ent:decomposer
```

### Error Handling

Commands use consistent error format:

```
❌ ERROR: <error message>

Cause: <root cause>
Solution: <suggested fix>

See: <documentation link>
```

**Example**:
```
❌ ERROR: Task has unmet dependencies

Cause: Task add-auth/2 depends on add-auth/1 (pending)
Solution: Complete add-auth/1 first or use /ent:task add-auth/1

See: docs/OPENSPEC_WORKFLOW.md#task-dependencies
```

---

## See Also

- [OpenSpec Workflow](./OPENSPEC_WORKFLOW.md) - Workflow details
- [MCP API Reference](./MCP_API.md) - Underlying MCP tools
- [Agents and Skills](./AGENTS_AND_SKILLS.md) - Agent system
- [CLI Reference](./CLI_REFERENCE.md) - CLI commands

---

**Version:** v0.3.0
**Last updated:** 2026-01-28
