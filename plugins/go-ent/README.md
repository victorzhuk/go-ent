# Go-Ent Plugin

Enterprise Go development toolkit with spec-driven workflows, Clean Architecture, and production-ready patterns.

## Installation

```bash
/plugin install go-ent@go-ent
```

## Quick Start

```bash
# Initialize project
/goent:init my-service

# Plan a feature (guided workflow with approval gates)
/goent:plan "Add user authentication"

# Apply tasks from registry
/goent:apply

# Run autonomous loop for repetitive fixes
/goent:loop "fix all linting errors" --max-iterations=10
```

## Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `/goent:init <name>` | Initialize new Go project |
| `/goent:scaffold <type> <name>` | Generate components |
| `/goent:lint` | Run linters |

### Workflow Commands

| Command | Description |
|---------|-------------|
| `/goent:plan <feature>` | Multi-phase planning with approval gates |
| `/goent:clarify <change-id>` | Ask clarifying questions |
| `/goent:research <change-id>` | Research unknowns and technology choices |
| `/goent:decompose <change-id>` | Task decomposition with dependencies |
| `/goent:analyze <change-id>` | Consistency validation |

### Execution Commands

| Command | Description |
|---------|-------------|
| `/goent:apply [change-id]` | Execute tasks from registry |
| `/goent:gen` | Generate code from OpenAPI/Proto |
| `/goent:tdd <feature>` | Red-Green-Refactor TDD cycle |
| `/goent:loop <task> [--max-iterations=N]` | Autonomous self-correction loop |
| `/goent:loop-cancel` | Cancel running loop |

### Registry Commands

| Command | Description |
|---------|-------------|
| `/goent:registry list [--filters]` | List all tasks |
| `/goent:registry next [count]` | Get next recommended task |
| `/goent:registry update <task-id> <field=value>` | Update task status |
| `/goent:registry deps <task-id> <op>` | Manage dependencies |
| `/goent:registry sync` | Sync from tasks.md files |

### Change Management

| Command | Description |
|---------|-------------|
| `/goent:status` | View registry and change status |
| `/goent:archive <change-id>` | Archive completed change |

## Scaffold Types

```bash
# Components
/goent:scaffold entity User
/goent:scaffold repository User pgx
/goent:scaffold usecase CreateUser
/goent:scaffold handler User

# Full stack (domain + repo + usecase + transport)
/goent:scaffold service Order
```

## Agents

Tiered by model for optimal performance and cost:

### Senior Tier (Opus)
- `@goent:architect` - System design and architecture
- `@goent:reviewer` - Code review with confidence filtering
- `@goent:lead` - Workflow orchestration

### Balanced Tier (Sonnet)
- `@goent:planner` - Feature planning and decomposition
- `@goent:dev` - Implementation and coding
- `@goent:debug` - Bug investigation

### Fast Tier (Haiku)
- `@goent:tester` - Quick test feedback

## Skills

Skills activate automatically based on context:

| Skill | Triggers |
|-------|----------|
| `go-api` | API design, OpenAPI, gRPC, protobuf |
| `go-arch` | Architecture, Clean Architecture, DDD |
| `go-code` | Go implementation, patterns, Go 1.25+ |
| `go-db` | Database, PostgreSQL, ClickHouse, Redis |
| `go-ops` | Operations, Docker, Kubernetes, CI/CD |
| `go-perf` | Performance, profiling, optimization |
| `go-sec` | Security, OWASP, authentication, crypto |
| `go-test` | Testing, testify, testcontainers, TDD |
| `go-review` | Code review patterns and checklists |

## Workflow Features

### Guided Planning with Approval Gates

The `/goent:plan` command provides a comprehensive planning workflow with **4 explicit wait points** where you approve before the agent continues:

**Phase 0: Clarification & Research**
1. **WAIT 1**: Clarifying questions - agent asks, you answer
2. **WAIT 2**: Research review - agent presents findings, you approve approach

**Phase 1: Design & Contracts**
3. **WAIT 3**: Design approval - review architecture decisions

**Phase 2: Task Generation**
4. **WAIT 4**: Final plan approval - review complete task breakdown

**Artifacts Created**:
- `proposal.md` - Why and what changes
- `design.md` - Technical decisions
- `research.md` - Research findings
- `tasks.md` - Enhanced with IDs and dependencies
- `specs/*/spec.md` - Requirement deltas

### Autonomous Loop

Self-correcting execution for repetitive tasks:

```bash
# Fix linting errors autonomously
/goent:loop "fix all linting errors" --max-iterations=10

# Implement feature with auto-correction
/goent:loop "add email validation to User entity" --max-iterations=15

# Cancel if stuck
/goent:loop-cancel
```

**Features**:
- Automatic error detection and adjustment
- Iteration tracking with adjustment history
- State persistence in `openspec/.loop-state.yaml`
- Smart stopping (success, max iterations, or same error 3x)
- Safe cancellation with state preservation

**Guardrails**:
- Never modifies critical files (go.mod, .git/)
- Always runs tests after changes
- Stops if same error repeats
- Never pushes to remote
- Documents all adjustments

### Registry Management

Centralized task tracking across all changes:

```bash
# Initialize from existing changes
/goent:registry sync

# Get next task recommendation
/goent:registry next

# Start working (auto-picks task)
/goent:apply

# Update task status
/goent:registry update add-auth/1.1 status=completed

# Manage cross-change dependencies
/goent:registry deps add-auth/2.1 add add-build/5.5
```

**Features**:
- Cross-change visibility and dependencies
- Priority-based recommendations (critical > high > medium > low)
- Dependency cycle detection
- Progress tracking and completion rates
- Smart next-task selection

## Code Standards

### Naming
```go
// ✅ Natural, concise
cfg, repo, srv, pool, ctx, req, resp

// ❌ AI-style verbose
applicationConfiguration, userRepositoryInstance
```

### Comments
```go
// ✅ WHY only (rare)
// Required by legacy API - remove after v2

// ❌ WHAT (fix naming instead)
// Create a new user
// Get user by ID
```

### Error Handling
```go
// ✅ Lowercase, concise context
return fmt.Errorf("query user %s: %w", id, err)
return fmt.Errorf("create order: %w", err)

// ❌ Verbose, capitalized
return fmt.Errorf("Failed to query user: %w", err)
```

### Architecture Layers
```
Transport → UseCase → Domain ← Repository ← Infrastructure
```

**Rules**:
- Domain has ZERO external dependencies
- Interfaces at consumer side
- Private by default (expose only what's needed)
- Repository models private, return domain entities

## Confidence-Based Code Review

The `@goent:reviewer` agent filters findings by confidence level:

- **95-100%**: Definite bugs, security vulnerabilities (always shown)
- **85-94%**: Strong code quality issues (always shown)
- **75-84%**: Style inconsistencies (always shown)
- **<75%**: Skipped (subjective preferences)

**Only issues ≥80% confidence are reported**, reducing noise and focusing on high-impact improvements.

## MCP Tools

The plugin provides 18 MCP tools for automation:

**Spec Management**:
- `goent_spec_init` - Initialize openspec
- `goent_spec_list` - List specs/changes
- `goent_spec_show` - Show details
- `goent_spec_create` - Create item
- `goent_spec_update` - Update item
- `goent_spec_delete` - Delete item

**Registry**:
- `goent_registry_list` - List tasks
- `goent_registry_next` - Get next task
- `goent_registry_update` - Update task
- `goent_registry_add_dep` - Add dependency
- `goent_registry_remove_dep` - Remove dependency
- `goent_registry_sync` - Sync from tasks.md

**Workflow**:
- `goent_workflow_start` - Start guided workflow
- `goent_workflow_approve` - Approve wait point
- `goent_workflow_status` - Check workflow state

**Loop**:
- `goent_loop_start` - Start autonomous loop
- `goent_loop_get` - Get loop state
- `goent_loop_set` - Update loop state
- `goent_loop_cancel` - Cancel loop

## Best Practices

### When to Use What

**Use `/goent:plan`** for:
- New features
- Breaking changes
- Architecture changes
- Performance optimizations

**Use `/goent:loop`** for:
- Fixing failing tests
- Resolving linting errors
- Straightforward implementations
- Iterative debugging

**Use `/goent:apply`** for:
- Executing planned tasks
- Following registry recommendations
- Structured implementation

**Use Direct Commands** for:
- Bug fixes (no proposal needed)
- Quick scaffolding
- Running tests/linters

### Planning Workflow

1. **Explore**: `openspec list`, `openspec list --specs`
2. **Plan**: `/goent:plan "feature description"`
3. **Approve**: Review at each of 4 wait points
4. **Sync**: `/goent:registry sync`
5. **Execute**: `/goent:apply` (or implement manually)
6. **Archive**: `/goent:archive <change-id>`

### Error Recovery

If autonomous loop gets stuck:
```bash
# Cancel and review state
/goent:loop-cancel
cat openspec/.loop-state.yaml

# Refine task description and retry
/goent:loop "more specific task description" --max-iterations=5
```

## License

MIT
