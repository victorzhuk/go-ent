# Go-Ent Plugin

Enterprise Go development toolkit with spec-driven workflows, Clean Architecture, and production-ready patterns.

## Installation

```bash
/plugin install go-ent@go-ent
```

## Quick Start

```bash
# Initialize project
/go-ent:init my-service

# Plan a feature (guided workflow with approval gates)
/go-ent:plan "Add user authentication"

# Apply tasks from registry
/go-ent:apply

# Run autonomous loop for repetitive fixes
/go-ent:loop "fix all linting errors" --max-iterations=10
```

## Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `/go-ent:init <name>` | Initialize new Go project |
| `/go-ent:scaffold <type> <name>` | Generate components |
| `/go-ent:lint` | Run linters |

### Workflow Commands

| Command | Description |
|---------|-------------|
| `/go-ent:plan <feature>` | Multi-phase planning with approval gates |
| `/go-ent:clarify <change-id>` | Ask clarifying questions |
| `/go-ent:research <change-id>` | Research unknowns and technology choices |
| `/go-ent:decompose <change-id>` | Task decomposition with dependencies |
| `/go-ent:analyze <change-id>` | Consistency validation |

### Execution Commands

| Command | Description |
|---------|-------------|
| `/go-ent:apply [change-id]` | Execute tasks from registry |
| `/go-ent:gen` | Generate code from OpenAPI/Proto |
| `/go-ent:tdd <feature>` | Red-Green-Refactor TDD cycle |
| `/go-ent:loop <task> [--max-iterations=N]` | Autonomous self-correction loop |
| `/go-ent:loop-cancel` | Cancel running loop |

### Registry Commands

| Command | Description |
|---------|-------------|
| `/go-ent:registry list [--filters]` | List all tasks |
| `/go-ent:registry next [count]` | Get next recommended task |
| `/go-ent:registry update <task-id> <field=value>` | Update task status |
| `/go-ent:registry deps <task-id> <op>` | Manage dependencies |
| `/go-ent:registry sync` | Sync from tasks.md files |

### Change Management

| Command | Description |
|---------|-------------|
| `/go-ent:status` | View registry and change status |
| `/go-ent:archive <change-id>` | Archive completed change |

## Scaffold Types

```bash
# Components
/go-ent:scaffold entity User
/go-ent:scaffold repository User pgx
/go-ent:scaffold usecase CreateUser
/go-ent:scaffold handler User

# Full stack (domain + repo + usecase + transport)
/go-ent:scaffold service Order
```

## Agents

Tiered by model for optimal performance and cost:

### Senior Tier (Opus)
- `@go-ent:architect` - System design and architecture
- `@go-ent:reviewer` - Code review with confidence filtering
- `@go-ent:lead` - Workflow orchestration

### Balanced Tier (Sonnet)
- `@go-ent:planner` - Feature planning and decomposition
- `@go-ent:dev` - Implementation and coding
- `@go-ent:debug` - Bug investigation

### Fast Tier (Haiku)
- `@go-ent:tester` - Quick test feedback

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

The `/go-ent:plan` command provides a comprehensive planning workflow with **4 explicit wait points** where you approve before the agent continues:

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
/go-ent:loop "fix all linting errors" --max-iterations=10

# Implement feature with auto-correction
/go-ent:loop "add email validation to User entity" --max-iterations=15

# Cancel if stuck
/go-ent:loop-cancel
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
/go-ent:registry sync

# Get next task recommendation
/go-ent:registry next

# Start working (auto-picks task)
/go-ent:apply

# Update task status
/go-ent:registry update add-auth/1.1 status=completed

# Manage cross-change dependencies
/go-ent:registry deps add-auth/2.1 add add-build/5.5
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

The `@go-ent:reviewer` agent filters findings by confidence level:

- **95-100%**: Definite bugs, security vulnerabilities (always shown)
- **85-94%**: Strong code quality issues (always shown)
- **75-84%**: Style inconsistencies (always shown)
- **<75%**: Skipped (subjective preferences)

**Only issues ≥80% confidence are reported**, reducing noise and focusing on high-impact improvements.

## MCP Tools

The plugin provides 18 MCP tools for automation:

**Spec Management**:
- `go_ent_spec_init` - Initialize openspec
- `go_ent_spec_list` - List specs/changes
- `go_ent_spec_show` - Show details
- `go_ent_spec_create` - Create item
- `go_ent_spec_update` - Update item
- `go_ent_spec_delete` - Delete item

**Registry**:
- `go_ent_registry_list` - List tasks
- `go_ent_registry_next` - Get next task
- `go_ent_registry_update` - Update task
- `go_ent_registry_add_dep` - Add dependency
- `go_ent_registry_remove_dep` - Remove dependency
- `go_ent_registry_sync` - Sync from tasks.md

**Workflow**:
- `go_ent_workflow_start` - Start guided workflow
- `go_ent_workflow_approve` - Approve wait point
- `go_ent_workflow_status` - Check workflow state

**Loop**:
- `go_ent_loop_start` - Start autonomous loop
- `go_ent_loop_get` - Get loop state
- `go_ent_loop_set` - Update loop state
- `go_ent_loop_cancel` - Cancel loop

## Best Practices

### When to Use What

**Use `/go-ent:plan`** for:
- New features
- Breaking changes
- Architecture changes
- Performance optimizations

**Use `/go-ent:loop`** for:
- Fixing failing tests
- Resolving linting errors
- Straightforward implementations
- Iterative debugging

**Use `/go-ent:apply`** for:
- Executing planned tasks
- Following registry recommendations
- Structured implementation

**Use Direct Commands** for:
- Bug fixes (no proposal needed)
- Quick scaffolding
- Running tests/linters

### Planning Workflow

1. **Explore**: `openspec list`, `openspec list --specs`
2. **Plan**: `/go-ent:plan "feature description"`
3. **Approve**: Review at each of 4 wait points
4. **Sync**: `/go-ent:registry sync`
5. **Execute**: `/go-ent:apply` (or implement manually)
6. **Archive**: `/go-ent:archive <change-id>`

### Error Recovery

If autonomous loop gets stuck:
```bash
# Cancel and review state
/go-ent:loop-cancel
cat openspec/.loop-state.yaml

# Refine task description and retry
/go-ent:loop "more specific task description" --max-iterations=5
```

## License

MIT
