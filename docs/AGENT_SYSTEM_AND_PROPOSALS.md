# go-ent Agent System & Proposal Creation Guide

## Quick Reference

| Extension | Location | Invocation | Purpose |
|-----------|----------|------------|---------|
| Skills | `.claude/skills/` | Auto-activated | Domain expertise |
| Agents | `.claude/agents/` | `@ent:name` | Specialized tasks |
| Commands | `.claude/commands/` | `/ent:name` | Workflow triggers |

---

## 1. Agent System Architecture

### Directory Structure
```
plugins/go-ent/agents/
├── meta/              # Agent YAML configs
│   └── bases/         # Base configs for inheritance
├── presets/tools.yaml # Tool preset definitions
├── prompts/
│   └── shared/        # Shared fragments
│       ├── _openspec.md    # OpenSpec workflow
│       ├── _handoffs.md    # Delegation rules
│       ├── _judgment.md    # Decision framework
│       └── _principals.md  # Priority hierarchy
├── schema/agent.schema.json
└── templates/         # Output templates
```

### Agent Metadata Schema

| Field | Type | Values | Description |
|-------|------|--------|-------------|
| `name` | string | `ent:[a-z0-9-]+` | Agent identifier |
| `model` | enum | `fast`, `main`, `heavy` | Model tier |
| `role` | enum | `planning`, `execution`, `validation`, `research` | Role category |
| `skills` | array | skill names | Auto-load skills |
| `toolPresets` | array | preset names | Tool groups |
| `dependencies` | array | agent names | Delegation targets |
| `extends` | string | base name | Inherit from base |

### Tool Presets (from `presets/tools.yaml`)

| Preset | Tools | Purpose |
|--------|-------|---------|
| `read-only` | Read, Glob, Grep | Basic file reading |
| `editing` | Read, Write, Edit, Bash, Glob, Grep | Full modification access |
| `serena-analysis` | find_symbol, find_referencing_symbols, get_symbols_overview, search_for_pattern, list_dir, read_file | Symbolic code analysis |
| `serena-editing` | replace_symbol_body, insert_after_symbol, insert_before_symbol, create_text_file, replace_content | (Disallowed by default) |

**Usage Pattern**: Most agents use `read-only` + `serena-analysis`. Execution agents add `editing`. The `serena-editing` preset is typically in `disallowedToolPresets` for safety.

### Agent Inventory

| Agent | Model | Skills | Delegates To |
|-------|-------|--------|--------------|
| `ent:architect` | heavy | go-arch, go-api, api-design | planner, coder |
| `ent:planner` | main | go-arch, go-code | (from base) |
| `ent:planner-fast` | fast | arch-core | planner, architect |
| `ent:planner-heavy` | heavy | go-arch, go-api, go-code, go-db | decomposer, architect |
| `ent:coder` | main | go-code, go-db | tester, reviewer, debugger |
| `ent:debugger` | main | go-code, go-perf, go-test | tester, reviewer |
| `ent:debugger-fast` | fast | (from base) | tester, coder, reviewer |
| `ent:debugger-heavy` | heavy | go-code, go-perf, debug-core, go-sec | reviewer, acceptor, tester |
| `ent:tester` | fast | go-test | none |
| `ent:reviewer` | heavy | go-review, security-core | none |
| `ent:decomposer` | heavy | arch-core | coder |
| `ent:researcher` | heavy | go-code, debug-core | debugger-fast, architect |
| `ent:acceptor` | heavy | go-test, review-core | coder, architect |
| `ent:task-fast` | fast | arch-core | coder, task-heavy |
| `ent:task-heavy` | heavy | go-arch, go-code, go-api | coder, architect |

### Delegation Chains

```
Feature:  architect → planner → coder → tester → reviewer
Bug Fix:  debugger-fast → debugger → debugger-heavy → tester
Review:   reviewer → reviewer-heavy (architecture/security)
```

### Agent Inheritance (`extends` field)

Agents can extend base configs in `meta/bases/`:

```yaml
# meta/bases/debugger.yaml (base)
name: ent:debugger
model: main
skills: [go-code, debug-core]
toolPresets: [editing, serena-analysis]
disallowedToolPresets: [serena-editing]
dependencies: [ent:tester, ent:reviewer]

# meta/debugger-fast.yaml (derived)
extends: debugger
model: fast  # Override model tier
dependencies:
  - ent:tester
  - ent:coder    # Additional delegation target
  - ent:reviewer
```

**Base agents**: `bases/debugger.yaml`, `bases/planner.yaml`, `bases/task.yaml`

---

## 1.5 Constitutional AI Prompts

Shared prompt fragments in `plugins/go-ent/agents/prompts/shared/`:

| File | Purpose |
|------|---------|
| `_openspec.md` | OpenSpec workflow instructions |
| `_handoffs.md` | Agent delegation rules, escalation tiers |
| `_judgment.md` | Decision framework when rules conflict |
| `_principals.md` | Priority hierarchy (conventions > intent > best practices > safety > simplicity) |
| `_conventions.md` | Go code style and architecture patterns |
| `_tooling.md` | Tool usage patterns (native vs Serena) |

### Principal Hierarchy (Conflict Resolution)

When values conflict, apply in order:
1. **Project conventions** - Established patterns in THIS codebase
2. **User intent** - What the human wants
3. **Best practices** - Industry standards
4. **Safety** - Security, data integrity
5. **Simplicity** - KISS, YAGNI

### Escalation Rules

**Ask when**: Ambiguous intent, high-risk changes, irreversible operations
**Decide when**: Clear requirements, low-risk changes, following patterns

---

## 2. OpenSpec Proposal System

### Directory Structure
```
openspec/
├── specs/           # Permanent specifications
├── changes/         # Active changes
│   └── {change-id}/
│       ├── proposal.md     # Required
│       ├── tasks.md        # Required
│       ├── design.md       # Optional (complex changes)
│       └── specs/          # Spec deltas
└── archive/         # Completed changes (YYYY-MM-DD-id/)
```

### Naming Conventions
- **Directory:** `kebab-case-description` or `YYYY-MM-DD-kebab-case`
- **Change ID:** Same as directory name
- **Files:** `proposal.md`, `tasks.md`, `design.md`

---

## 3. Proposal Template (proposal.md)

```markdown
# Proposal: <Title>

## Status: pending

## Why
<1-3 paragraphs explaining the problem or motivation>

## What Changes
<Bullet list of proposed changes with affected files>

### Key Components
- `internal/path/file.go` - Description
- `internal/path/other.go` - Description

## Success Criteria
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Tests pass: `go test ./... -race`
- [ ] Build succeeds: `go build ./...`

## Impact
<Performance, architecture, or breaking changes>

## Alternatives Considered
1. **Alternative A** - Why rejected
2. **Alternative B** - Why rejected
```

### Real Examples

**Simple (cleanup):**
```markdown
## Why
Remove orphaned packages and deprecated functions to reduce maintenance burden.

## What Changes
- Delete 4 orphaned directories
- Remove deprecated functions
- Total reduction: ~500 lines
```

**Medium (feature restoration):**
```markdown
## Problem
The `init` command was deleted but is essential:
1. External projects cannot scaffold configs
2. Self-hosted development loses workflow
3. Documentation is stale

## Proposed Solution
Restore `init` command with minimal complexity (~300-400 LOC)
```

**Complex (architecture):**
```markdown
## Why
go-ent currently operates as MCP server. To enable multi-agent orchestration:

**Key Architecture:**
- **Claude Code (Opus)** = Master orchestrator
- **go-ent** = ACP proxy
- **OpenCode** = Workers

## What Changes
### 1. go-ent as ACP Proxy
[Detailed breakdown with tables, code examples]
```

---

## 4. Tasks Template (tasks.md)

```markdown
# Tasks: <Same Title as Proposal>

## Status: pending

## Phase 1: <Phase Name>

### Task 1.1: <Task Title>
**Priority:** High
**Files:** `internal/path/file.go`

Steps:
1. First step
2. Second step

Validation:
- [ ] Acceptance criterion

### Task 1.2: <Next Task>
**Dependencies:** Task 1.1
...

## Phase 2: <Next Phase>
...

## Completion Summary
**Date:** YYYY-MM-DD

### Verification
- [ ] All tests pass
- [ ] Build succeeds
- [ ] Manual verification done
```

### Task Format Options

**Simple (checkboxes):**
```markdown
## 1. Section Name
- [x] 1.1 Task description (2026-01-28)
- [ ] 1.2 Another task
```

**Detailed (with metadata):**
```markdown
### Task 1: Task Title
**Priority:** High
**Complexity:** Medium
**Dependencies:** Task 2

**Steps:**
1. Step one
2. Step two

**Files Modified:**
- `path/to/file.go`
```

### Status Markers
- `[ ]` = pending
- `[x]` = completed
- `✓ YYYY-MM-DD` = completed with date
- `⏳` = in progress

### Dependency Graph (ASCII)
```
Task 1 (Setup)
  ↓
Task 2 (Core impl)
  ↓
Task 3 (Integration)
```

---

## 5. Design Document (design.md) - Optional

Use for complex architectural changes:

```markdown
## Context
<Why this design is needed>

## Goals
- Goal 1
- Goal 2

## Non-Goals
- What we're NOT solving

## Architecture
<ASCII diagrams, component breakdown>

## Decisions
- **D1:** Decision with rationale
- **D2:** Decision with rationale

## Risks
| Risk | Mitigation |
|------|------------|
| Risk 1 | How to handle |

## Open Questions
1. Question? (Leaning: answer)
```

---

## 6. Workflow Commands

| Command | Purpose |
|---------|---------|
| `/ent:plan <desc>` | Create new proposal (automated) |
| `/ent:task` or `/ent:apply` | Execute next task |
| `/ent:status` | View workflow state |
| `/ent:registry list` | Show all tasks |
| `/ent:archive <id>` | Archive completed change |

### Planning Workflow Phases

| Phase | Agent | Output |
|-------|-------|--------|
| 0-2 | `planner-fast` | Feasibility, clarification, research |
| 3 | `architect` | proposal.md, design.md |
| 4 | `planner` | Spec deltas |
| 5 | `decomposer` | tasks.md |
| 6 | (validation) | Final checks |

---

## 7. Manual Proposal Creation

1. **Create directory:**
   ```bash
   mkdir -p openspec/changes/my-refactor
   ```

2. **Write proposal.md** with:
   - Why (problem/motivation)
   - What Changes (solution)
   - Success Criteria (checkboxes)
   - Files affected

3. **Write tasks.md** with:
   - Phased task breakdown
   - Dependencies between tasks
   - Validation criteria per task

4. **Optional design.md** for:
   - Architecture diagrams
   - Decision rationale
   - Risk analysis

5. **Execute:**
   ```bash
   /ent:apply
   ```

---

## 8. Key Files Reference

| File | Purpose |
|------|---------|
| `docs/tools/claude-code-extension-guide.md` | Claude Code integration |
| `docs/tools/opencode-extension-guide.md` | OpenCode integration |
| `plugins/go-ent/agents/meta/*.yaml` | Agent configs |
| `plugins/go-ent/agents/meta/bases/*.yaml` | Base agent configs for inheritance |
| `plugins/go-ent/agents/presets/tools.yaml` | Tool preset definitions |
| `plugins/go-ent/platformspecs/agent.schema.json` | Agent validation schema |
| `plugins/go-ent/agents/prompts/shared/_openspec.md` | Workflow guide |
| `plugins/go-ent/agents/prompts/shared/_handoffs.md` | Delegation rules |
| `plugins/go-ent/agents/prompts/shared/_judgment.md` | Decision framework |
| `plugins/go-ent/agents/prompts/shared/_principals.md` | Priority hierarchy |
| `plugins/go-ent/commands/workflows/plan.md` | Planning workflow |
| `plugins/go-ent/commands/workflows/task.md` | Task execution workflow |
| `plugins/go-ent/commands/workflows/bug.md` | Bug fixing workflow |
| `openspec/config.yaml` | OpenSpec configuration |
| `openspec/registry.yaml` | Task registry with status tracking |
| `openspec/specs/` | Permanent specification library |

---

## 9. Registry Integration

### Registry File (`openspec/registry.yaml`)

```yaml
version: '1.0'
synced_at: '2026-01-27T12:00:00+03:00'

changes:
  my-feature:
    id: my-feature
    title: 'My Feature Title'
    status: ready          # pending | ready | in_progress | complete
    tasks_file: /path/to/tasks.md
    total: 12
    completed: 5
    in_progress: 1
    blocked: 0
    depends_on:            # Optional dependencies
      - prerequisite-change
    completed_at: null     # Set when status=complete

archived:
  old-feature:
    id: old-feature
    status: complete
    archived: true
    archive_path: /path/to/archive
    archived_at: '2026-01-20T15:34:53+03:00'
```

### Status Lifecycle

1. **pending** - Proposal created, waiting for dependencies
2. **ready** - Dependencies resolved, can start
3. **in_progress** - Tasks being executed
4. **complete** - All tasks done, validated
5. **archived** - Moved to `archive/`, specs merged

---

## 10. Real Example: Complete Proposal

From `openspec/changes/archive/simplify-01-delete-unused/`:

**proposal.md:**
```markdown
# Proposal: Delete Unused Packages (Phase 1)

## Status: complete

## Why
Simplification initiative. Remove orphaned packages to reduce maintenance.

## What Changes
- Delete 15 unused packages in internal/
- Update MCP server initialization
- Fix remaining imports

### Key Components
- `internal/aggregator/` - Unused, delete
- `internal/validator/` - Unused, delete
- ...

## Success Criteria
- [x] All packages deleted
- [x] No orphaned imports remain
- [x] `go build ./...` succeeds
- [x] `go test ./... -race` passes
```

**tasks.md:**
```markdown
# Tasks: Delete Unused Packages

## Status: ✅ COMPLETED
**Completion Date:** 2026-01-27
**Tasks Completed:** 19/19

### 1. Delete internal/aggregator/ directory
**Priority:** High
**Dependencies:** None

Steps:
1. Remove `internal/aggregator/` directory
2. Search for imports: `grep -r "go-ent/internal/aggregator"`
3. Remove any found imports

Validation:
- [x] Directory deleted
- [x] No imports remain

Files: `internal/aggregator/` (deleted)

---

### 16. Update MCP server initialization
**Priority:** High
**Dependencies:** Tasks 1-15

...

## Task Order
**Parallel:** Tasks 1-15
**Sequential:** Tasks 16-19 (after 1-15)

## Total Time: ~5.5 hours
```

---

## 11. Validation Checklist

Before executing:
- [ ] All TBD/TODO resolved
- [ ] Tasks have IDs and file paths
- [ ] Dependencies form valid graph (no cycles)
- [ ] Success criteria are testable
- [ ] Files affected are listed

After completing:
- [ ] `go build ./...` succeeds
- [ ] `go test ./... -race` passes
- [ ] Manual verification done
- [ ] Task statuses updated with dates
- [ ] Registry synced (`openspec/registry.yaml`)

---

## 12. Creating a New Proposal (Step-by-Step)

### Option A: Using `/ent:plan` Command

```bash
/ent:plan "Add Redis caching for user queries"
```

The command triggers the agent chain:
1. `@ent:planner-fast` - Feasibility check
2. `@ent:architect` - Design proposal.md, design.md
3. `@ent:planner` - Write spec deltas
4. `@ent:decomposer` - Break down into tasks.md

### Option B: Manual Creation

```bash
# 1. Create directory
mkdir -p openspec/changes/add-redis-caching

# 2. Write proposal.md
cat > openspec/changes/add-redis-caching/proposal.md << 'EOF'
# Proposal: Add Redis Caching

## Status: pending

## Why
User queries hit database directly, causing latency.

## What Changes
- Add Redis client package
- Implement cache wrapper for UserRepository
- Add cache invalidation on writes

## Success Criteria
- [ ] Cache hit rate > 80%
- [ ] P99 latency < 50ms
- [ ] Tests pass
EOF

# 3. Write tasks.md
cat > openspec/changes/add-redis-caching/tasks.md << 'EOF'
# Tasks: Add Redis Caching

## Phase 1: Infrastructure
- [ ] 1.1 Add rueidis dependency
- [ ] 1.2 Create Redis client wrapper

## Phase 2: Implementation
- [ ] 2.1 Implement cache layer
  - Dependencies: 1.1, 1.2
- [ ] 2.2 Add invalidation hooks

## Phase 3: Testing
- [ ] 3.1 Add integration tests
  - Dependencies: 2.1, 2.2
EOF

# 4. Execute
/ent:apply
```

---

## 13. Best Practices

### Proposal Writing

1. **Why section**: Focus on the problem, not the solution
2. **What Changes**: Be specific about files and components
3. **Success Criteria**: Make them testable and measurable
4. **Impact**: Call out breaking changes early

### Task Breakdown

1. **Granularity**: Each task should be completable in < 2 hours
2. **Dependencies**: Make them explicit and directed
3. **Validation**: Every task needs acceptance criteria
4. **Prioritization**: Mark blocking tasks as High priority

### Agent Delegation

1. **Trust the chain**: Let agents delegate to specialists
2. **Fast first**: Use `planner-fast` for quick checks
3. **Heavy for architecture**: Complex decisions need `architect` or `planner-heavy`
4. **Test everything**: Always delegate to `tester` before review

### Registry Management

1. **Sync frequently**: After each task completion
2. **Update status**: Keep registry.yaml accurate
3. **Track dependencies**: Use `depends_on` for change ordering
4. **Archive promptly**: Move completed changes to archive/

---

## 14. Troubleshooting

### Common Issues

**Issue**: Tasks not executing
- Check `registry.yaml` status field
- Verify `depends_on` dependencies are complete
- Ensure tasks.md format is correct

**Issue**: Agent delegation failing
- Verify agent name in `dependencies` field
- Check if target agent exists in `meta/`
- Review `_handoffs.md` for delegation rules

**Issue**: Build failures after task
- Run validation checklist before archiving
- Check for orphaned imports
- Verify all files in task are updated

### Debug Commands

```bash
# Check registry status
cat openspec/registry.yaml | grep -A 10 "my-change"

# Find orphaned imports
grep -r "deleted/package" . --include="*.go"

# Validate proposal format
cat openspec/changes/my-change/proposal.md

# List all pending tasks
/ent:registry list
```

---

## 15. Advanced Topics

### Custom Agent Creation

1. Create YAML in `plugins/go-ent/agents/meta/`
2. Define skills, tools, dependencies
3. Optionally extend from base config
4. Add to delegation chains in `_handoffs.md`

### Skill Development

1. Create skill in `plugins/go-ent/skills/`
2. Write SKILL.md with structured sections
3. Test with `/ent:skill-sync`
4. Reference in agent metadata

### Workflow Customization

1. Modify phase definitions in `_openspec.md`
2. Add custom commands in `plugins/go-ent/commands/`
3. Update registry schema in `openspec/config.yaml`
4. Extend agent prompts in `prompts/shared/`

---

## Appendix A: File Format Specifications

### proposal.md Structure

```markdown
# Proposal: <title>
## Status: <pending|ready|in_progress|complete>
## Why
## What Changes
### Key Components
## Success Criteria
## Impact
## Alternatives Considered
```

### tasks.md Structure

```markdown
# Tasks: <title>
## Status: <pending|in_progress|complete>
## Phase N: <name>
### Task N.M: <title>
**Priority:** <High|Medium|Low>
**Dependencies:** <task-ids>
**Files:** <file-paths>
Steps:
Validation:
## Completion Summary
```

### Agent Metadata (YAML)

```yaml
name: ent:agent-name
model: fast|main|heavy
role: planning|execution|validation|research
skills:
  - skill-name
toolPresets:
  - preset-name
disallowedToolPresets:
  - preset-name
dependencies:
  - ent:other-agent
extends: base-name  # optional
```

---

## Appendix B: Tool Preset Definitions

Full tool preset matrix from `presets/tools.yaml`:

```yaml
presets:
  read-only:
    - Read
    - Glob
    - Grep

  editing:
    - Read
    - Write
    - Edit
    - Bash
    - Glob
    - Grep

  serena-analysis:
    - find_symbol
    - find_referencing_symbols
    - get_symbols_overview
    - search_for_pattern
    - list_dir
    - read_file

  serena-editing:  # Typically disallowed
    - replace_symbol_body
    - insert_after_symbol
    - insert_before_symbol
    - create_text_file
    - replace_content
```

---

## Appendix C: Quick Command Reference

```bash
# Planning
/ent:plan <description>        # Create proposal
/ent:status                    # View workflow state
/ent:registry list             # List all tasks

# Execution
/ent:apply                     # Execute next task
/ent:task                      # Alias for apply

# Completion
/ent:archive <change-id>       # Archive completed change

# Agents
@ent:architect <prompt>        # Architecture design
@ent:planner <prompt>          # Task planning
@ent:coder <prompt>            # Implementation
@ent:debugger <prompt>         # Bug investigation
@ent:tester <prompt>           # Testing
@ent:reviewer <prompt>         # Code review

# Build & Test
make build                     # Build binary
make test                      # Run tests
make lint                      # Run linter
```

---

**Last Updated:** 2026-01-28
**Version:** 1.0.0
**Maintainer:** go-ent project
