# OpenSpec Workflow

go-ent implements the OpenSpec workflow system for managing software changes through a structured, auditable process.

## Philosophy

OpenSpec is designed for **iterative, brownfield development** with these principles:

- **Fluid not rigid**: Workflows adapt to project needs, not enforced waterfall
- **Iterative not waterfall**: Changes evolve through exploration and refinement
- **Brownfield-first**: Designed for real codebases with history and constraints
- **Spec-driven**: Single source of truth with tracked modifications
- **Auditable**: Complete history of proposals, decisions, and changes

---

## Core Concepts

### Specs (Source of Truth)

**Location:** `openspec/specs/`

Specs are the **authoritative documentation** of your system. They represent the current, deployed state.

**Structure:**
```
openspec/specs/
├── api/
│   ├── handlers.md
│   └── middleware.md
├── domain/
│   ├── models.md
│   └── services.md
└── infrastructure/
    ├── database.md
    └── cache.md
```

**Characteristics:**
- Reflect **current production state**
- Updated only when changes are deployed
- Versioned with codebase
- Never modified directly during development

---

### Changes (Proposed Modifications)

**Location:** `openspec/changes/`

Changes are **proposed modifications** to specs. They represent work in progress.

**Structure:**
```
openspec/changes/
├── add-auth-middleware/
│   ├── proposal.md       # Problem, solution, alternatives
│   ├── design.md         # Technical design (optional)
│   ├── tasks.md          # Task breakdown
│   └── specs/           # Delta specs
│       └── api/
│           └── middleware.md
└── optimize-db-queries/
    ├── proposal.md
    ├── tasks.md
    └── specs/
        └── infrastructure/
            └── database.md
```

**Lifecycle:**
1. **Created**: `/ent:plan` or `/ent:new` creates change directory
2. **In Progress**: Tasks executed via `/ent:task` or `/ent:apply`
3. **Complete**: All tasks done, implementation verified
4. **Archived**: Merged into main specs, moved to `changes/archive/`

---

### Delta Specs (ADDED/MODIFIED/REMOVED)

Delta specs describe **differences from current specs** using structured sections:

```markdown
# API Middleware

## ADDED

### Authentication Middleware

New JWT-based authentication middleware:

<SNIP>
func AuthMiddleware(secret []byte) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        claims, err := validateJWT(token, secret)
        if err != nil {
            c.AbortWithStatus(401)
            return
        }
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
</SNIP>

## MODIFIED

### Error Handler Middleware

**Before:**
- Only logged errors

**After:**
- Logs errors with request ID
- Returns structured error response
- Tracks error metrics

## REMOVED

### Debug Middleware

Removed debug middleware as it's no longer needed in production.
```

**Benefits:**
- Clear diff from current state
- Easy review of proposed changes
- Simplified merge when archiving

---

### Artifact Dependency Graph

Changes follow a **dependency graph** of artifacts:

```
proposal.md
    ↓
design.md (optional)
    ↓
specs/*.md (delta specs)
    ↓
tasks.md
```

**Artifact Types:**

| Artifact | Purpose | Required |
|----------|---------|----------|
| `proposal.md` | Problem statement, solution, alternatives | Yes |
| `design.md` | Technical design, architecture decisions | Optional |
| `specs/*.md` | Delta specifications | Yes |
| `tasks.md` | Concrete implementation tasks | Yes |

**Workflow:**
1. Write proposal (problem, solution, alternatives)
2. Write design doc (if architecture changes)
3. Write delta specs (what will change)
4. Break down into tasks
5. Execute tasks
6. Verify implementation matches specs
7. Archive change

---

## Commands

### `/ent:plan <description>`

**Full planning workflow**: clarify → research → design → decompose

```
/ent:plan Add JWT authentication to API
```

**What it does:**
1. **Clarify**: Ask questions to understand requirements
2. **Research**: Explore codebase for context
3. **Design**: Create proposal and design doc
4. **Decompose**: Break into concrete tasks

**Creates:**
- `openspec/changes/<change-id>/proposal.md`
- `openspec/changes/<change-id>/design.md` (if needed)
- `openspec/changes/<change-id>/specs/` (delta specs)
- `openspec/changes/<change-id>/tasks.md`

**Use when:**
- Starting a new feature
- Complex architectural changes
- Need to explore options

---

### `/ent:task [change-id]`

**Execute tasks** from a change.

```
# Execute next task from active change
/ent:task

# Execute task from specific change
/ent:task add-auth-middleware
```

**What it does:**
1. Loads task from registry
2. Reads relevant specs and code
3. Implements change
4. Updates task status
5. Commits work

**Use when:**
- Implementing planned changes
- Following TDD workflow
- Incremental development

---

### `/ent:bug <description>`

**Debug-focused workflow**: reproduce → research → fix → test

```
/ent:bug Auth middleware returns 500 instead of 401
```

**What it does:**
1. **Reproduce**: Create minimal reproduction
2. **Research**: Investigate root cause
3. **Fix**: Implement solution
4. **Test**: Verify fix and add regression test

**Creates:**
- `openspec/changes/<bug-id>/proposal.md` (minimal)
- `openspec/changes/<bug-id>/tasks.md`

**Use when:**
- Fixing bugs
- Need to understand failure first
- Creating regression tests

---

### `/ent:status`

**View workflow state** and progress.

```
/ent:status
```

**Shows:**
- Active changes
- Task progress per change
- Next available task
- Budget usage

---

### `/ent:archive <change-id>`

**Archive completed change**.

```
/ent:archive add-auth-middleware
```

**What it does:**
1. Validates all tasks complete
2. Merges delta specs into main specs
3. Moves change to `changes/archive/`
4. Updates change index
5. Creates archive commit

**Use when:**
- Change is deployed to production
- Implementation verified
- Ready to update source of truth

---

## Workflow Patterns

### Quick Feature: new → ff → apply → archive

**Scenario:** Simple, well-understood feature

```bash
# 1. Fast-forward through planning
/ent:plan Add health check endpoint --ff

# 2. Execute all tasks
while /ent:task; do
    echo "Task complete"
done

# 3. Archive
/ent:archive add-health-check
```

---

### Exploratory: explore → new → continue → apply

**Scenario:** Unclear requirements, need investigation

```bash
# 1. Explore problem space
/ent:explore How does authentication work?

# 2. Create change based on findings
/ent:plan Add OAuth support

# 3. Execute incrementally
/ent:task
/ent:task
...

# 4. Archive when done
/ent:archive add-oauth
```

---

### Parallel Changes

**Scenario:** Multiple independent features

```
# Start multiple changes
/ent:plan Add rate limiting
/ent:plan Add request logging
/ent:plan Optimize database queries

# Work on each incrementally
/ent:task add-rate-limiting
/ent:task add-request-logging
/ent:task optimize-db-queries

# Archive as each completes
/ent:archive add-rate-limiting
/ent:archive add-request-logging
/ent:archive optimize-db-queries
```

**Best practices:**
- Keep changes independent (no shared files)
- Archive completed changes before starting dependent work
- Use delta specs to avoid merge conflicts

---

## Project Configuration

### `openspec/config.yaml`

**Structure:**
```yaml
# Context injection
context:
  project_overview: |
    go-ent is a Go development toolkit with MCP integration.

  coding_standards: |
    - Follow standard Go conventions
    - Use pointer receivers for structs
    - Wrap errors with context

  architecture_principles: |
    - Clean Architecture with DDD
    - Repository pattern for data access
    - Dependency injection

# Artifact workflow
artifacts:
  sequence:
    - proposal
    - design    # Optional
    - specs
    - tasks

  templates:
    proposal: templates/proposal.md
    design: templates/design.md
    tasks: templates/tasks.md

# Per-artifact rules
rules:
  proposal:
    require_alternatives: true
    require_tradeoffs: true

  specs:
    format: delta  # delta or full
    require_examples: true

  tasks:
    require_acceptance_criteria: true
    require_test_plan: true

# Archive behavior
archive:
  require_verification: true
  merge_strategy: delta  # delta or replace
  create_commit: true
```

---

## Archive Process

### Delta Merge

When archiving, delta specs are merged into main specs:

**Before Archive:**
```
openspec/
├── specs/
│   └── api/
│       └── middleware.md  (current state)
└── changes/
    └── add-auth/
        └── specs/
            └── api/
                └── middleware.md  (ADDED/MODIFIED/REMOVED)
```

**After Archive:**
```
openspec/
├── specs/
│   └── api/
│       └── middleware.md  (updated with changes)
└── changes/
    └── archive/
        └── add-auth/
            ├── proposal.md
            ├── specs/
            │   └── api/
            │       └── middleware.md  (delta preserved)
            └── tasks.md
```

**Merge Algorithm:**
1. **ADDED sections**: Append to main spec
2. **MODIFIED sections**: Replace in main spec
3. **REMOVED sections**: Delete from main spec
4. Update modification date
5. Preserve delta spec in archive

---

### Audit Trail

Archived changes provide complete audit trail:

```
openspec/changes/archive/
├── 2025-12-15-add-auth-middleware/
│   ├── proposal.md        # Why this change
│   ├── design.md          # How it was designed
│   ├── specs/             # What changed (delta)
│   └── tasks.md           # Implementation steps
├── 2025-12-20-optimize-queries/
└── 2026-01-05-add-rate-limiting/
```

**Each archive includes:**
- Original proposal with problem statement
- Design decisions and alternatives considered
- Delta specs showing exact changes
- Task breakdown and completion status
- Commit references

---

## Integration with go-ent

### MCP Tools

OpenSpec workflow is exposed via MCP tools:

- `go_ent_spec_init`: Initialize openspec structure
- `go_ent_spec_list`: List specs/changes/tasks
- `go_ent_spec_show`: Show spec details
- `go_ent_spec_create`: Create new spec/change
- `go_ent_spec_update`: Update spec/change
- `go_ent_spec_archive`: Archive completed change

See [MCP API Reference](./MCP_API.md) for details.

---

### Agent Integration

Agents understand OpenSpec structure:

- **Architect**: Creates proposals and design docs
- **Planner**: Breaks proposals into tasks
- **Coder**: Implements tasks, updates delta specs
- **Reviewer**: Validates implementation matches specs
- **Acceptor**: Verifies acceptance criteria before archive

See [Agent System](./AGENTS_AND_SKILLS.md) for details.

---

## Best Practices

### Proposal Writing

✅ **Good:**
```markdown
## Problem
Authentication is currently handled in each handler, leading to code duplication
and inconsistent behavior. 12 handlers each validate tokens differently.

## Solution
Create centralized JWT authentication middleware that validates tokens,
extracts claims, and attaches user context to requests.

## Alternatives Considered
1. **OAuth2 proxy** - Too heavy for our needs, adds deployment complexity
2. **API Gateway auth** - Couples authentication to infrastructure
3. **Middleware** (chosen) - Centralized, testable, framework-idiomatic
```

❌ **Bad:**
```markdown
## Problem
Need authentication

## Solution
Add middleware
```

---

### Delta Spec Writing

✅ **Good:**
```markdown
## MODIFIED

### UserRepository.Create

**Before:**
- No transaction support
- No error wrapping

**After:**
- Accepts context with transaction
- Wraps errors with operation context
- Returns domain errors (ErrDuplicateUser)

<SNIP>
func (r *UserRepository) Create(ctx context.Context, user *User) error {
    return r.db.WithContext(ctx).Create(user).Error
}
</SNIP>
```

❌ **Bad:**
```markdown
## MODIFIED

Changed Create method
```

---

### Task Breakdown

✅ **Good:**
```markdown
## Tasks

### 1. Create JWT validation function
- [ ] Implement parseJWT with error handling
- [ ] Add token expiration check
- [ ] Validate signature with secret
- [ ] Extract and validate claims
- [ ] **Test**: Valid/expired/invalid tokens

### 2. Implement middleware
- [ ] Create AuthMiddleware function
- [ ] Extract token from header
- [ ] Call validation function
- [ ] Set user context on success
- [ ] Return 401 on failure
- [ ] **Test**: Integration test with test server

### 3. Apply middleware to routes
- [ ] Add to router configuration
- [ ] Update route documentation
- [ ] **Test**: End-to-end API test
```

❌ **Bad:**
```markdown
## Tasks

- Add auth
- Test it
```

---

## See Also

- [CLI Reference](./CLI_REFERENCE.md) - OpenSpec CLI commands
- [MCP API Reference](./MCP_API.md) - OpenSpec MCP tools
- [Agent System](./AGENTS_AND_SKILLS.md) - Agent workflows
- [Commands Reference](./COMMANDS_REFERENCE.md) - Slash commands
