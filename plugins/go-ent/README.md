# Go-Ent Plugin

Enterprise Go development toolkit with spec-driven workflows, Clean Architecture, and production-ready patterns.

## Installation

```bash
/plugin install go-ent@go-ent
```

## Quick Start

```bash
# Initialize with all agents
/go-ent:init --tool=claude

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
| `/go-ent:init --tool=<claude\|opencode\|all>` | Initialize tool configuration with agents |
| `/go-ent:scaffold <type> <name>` | Generate components (via agents) |
| `/go-ent:lint` | Run linters |

**Init Flags**:
- `--tool` (required): `claude`, `opencode`, or `all`
- `--agents`: Comma-separated agent names
- `--include-deps`: Auto-resolve transitive dependencies
- `--no-deps`: Skip dependency validation
- `--dry-run`: Preview changes
- `--force`: Overwrite existing config
- `--model`: Override model by pattern

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

Use agents to generate components (delegated to `@go-ent:coder`):

```bash
# Component generation via agent interaction
@go-ent:coder Create User entity with email and password fields
@go-ent:coder Add User repository with pgx implementation
@go-ent:coder Implement CreateUser use case with validation
@go-ent:coder Add HTTP handler for user operations

# Full stack generation
@go-ent:coder Generate Order service with domain, repository, usecase, and transport layers
```

**Alternative**: Use agents directly for scaffolding via Serena tools instead of CLI commands.

## Agent Architecture

### File Structure

Agents use a split-file format for better organization and maintainability:

```
plugins/go-ent/agents/
├── meta/              # Agent metadata (name, model, dependencies)
│   ├── architect.yaml
│   ├── coder.yaml
│   └── ...
├── prompts/           # Agent prompts (shared + agent-specific)
│   ├── shared/
│   │   ├── _tooling.md
│   │   ├── _conventions.md
│   │   └── _handoffs.md
│   └── agents/
│       ├── architect.md
│       └── coder.md
└── templates/         # Tool-specific frontmatter templates
    ├── claude.yaml.tmpl
    └── opencode.yaml.tmpl
```

### Component Breakdown

**meta/*.yaml** - Agent configuration
```yaml
name: architect
description: System architect. Designs components, layers, data flow.
model: heavy
color: "#4169E1"
skills:
  - go-arch
  - go-api
tools:
  - read
  - glob
  - grep
  - mcp__plugin_serena_serena
dependencies:
  - planner
  - coder
tags:
  - "role:planning"
  - "complexity:heavy"
```

**prompts/shared/*.md** - Reusable prompt sections
- Tooling reference (Serena, Git, Go, Bash)
- Code conventions (naming, error handling, architecture)
- Handoff patterns between agents
- OpenSpec workflows

**prompts/agents/*.md** - Agent-specific instructions
- Responsibilities and outputs
- Design templates
- Handoff instructions

**templates/*.tmpl** - Tool-specific frontmatter
- Transforms metadata into tool format
- Claude Code format vs OpenCode format

### Dependency System

Agents declare explicit dependencies in `meta/*.yaml`:

```yaml
# architect.yaml
dependencies:
  - planner    # architect needs planner for task breakdown
  - coder      # architect needs coder for implementation
```

**Resolution**:
- Topological sort ensures proper execution order
- Cycle detection prevents circular dependencies
- Transitive dependencies resolved automatically
- Missing dependencies error with helpful messages

**Example Dependency Chain**:
```
architect → [planner, coder]
           planner → []
           coder → [tester, reviewer, debugger]
                  tester → []
                  reviewer → []
                  debugger → []

Execution order: planner, tester, reviewer, debugger, coder, architect
```

### CLI Flags for Agent Management

```bash
# Initialize all agents (auto-resolves dependencies)
go-ent init --tool=claude

# Initialize specific agents only
go-ent init --tool=claude --agents=planner,tester --no-deps

# Initialize agents with their dependencies
go-ent init --tool=claude --agents=architect --include-deps
# Results in: planner, coder, tester, reviewer, debugger, architect

# Dry-run to preview
go-ent init --tool=claude --agents=architect --include-deps --dry-run
```

**Flags**:
- `--tool` (required): `claude`, `opencode`, or `all`
- `--agents`: Comma-separated agent names (e.g., `planner,tester`)
- `--include-deps`: Auto-resolve transitive dependencies
- `--no-deps`: Skip dependency validation
- `--dry-run`: Preview changes without writing
- `--force`: Overwrite existing configuration
- `--model`: Override model by pattern (e.g., `heavy=opus`)

## Agents

Tiered by model for optimal performance and cost:

### Senior Tier (Heavy/Opus)
- `@go-ent:architect` - System design and architecture
- `@go-ent:reviewer` - Code review with confidence filtering
- `@go-ent:lead` - Workflow orchestration

### Balanced Tier (Main/Sonnet)
- `@go-ent:planner` - Feature planning and decomposition
- `@go-ent:dev` - Implementation and coding
- `@go-ent:debug` - Bug investigation

### Fast Tier (Fast/Haiku)
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

## Development Workflow

### Creating a New Agent

**Step 1: Create metadata file**
```bash
# File: plugins/go-ent/agents/meta/youragent.yaml
name: youragent
description: Brief description of what this agent does
model: main        # heavy, main, or fast
color: "#FF6B6B"
skills:
  - go-code
  - go-arch
tools:
  - read
  - write
  - edit
  - bash
  - glob
  - grep
  - mcp__plugin_serena_serena
dependencies:
  - tester        # agents this agent depends on
tags:
  - "role:execution"
  - "complexity:standard"
```

**Step 2: Create agent prompt**
```bash
# File: plugins/go-ent/agents/prompts/agents/youragent.md

You are a [role description].

## Responsibilities

- [primary responsibility 1]
- [primary responsibility 2]

## Outputs

Create in [appropriate location]:
- `filename.md` - Description of output
- `filename.go` - Description of output

## Template

```markdown
# Template Name

[template structure for the agent to follow]
```

## Handoff

After completion, delegate to:
- @ent:otheragent - Reason for handoff
```

**Step 3: Test the agent**
```bash
# Initialize with just your agent (no deps)
go-ent init --tool=claude --agents=youragent --no-deps

# Or with dependencies
go-ent init --tool=claude --agents=youragent --include-deps
```

### Modifying Existing Agents

**Update metadata**:
```bash
# Edit: plugins/go-ent/agents/meta/architect.yaml
# Add new dependencies, skills, or tools
dependencies:
  - planner
  - coder
  - researcher  # newly added
```

**Update prompts**:
```bash
# Edit: plugins/go-ent/agents/prompts/agents/architect.md
# Modify responsibilities, templates, or handoffs
```

**Re-initialize to apply changes**:
```bash
go-ent init --tool=claude --update
```

### Adding Shared Prompt Sections

Shared sections in `prompts/shared/*.md` are automatically included in agent prompts:

```bash
# File: plugins/go-ent/agents/prompts/shared/_newsection.md

# New Section Name

[content that multiple agents might need]

## Examples

[examples of usage]
```

These sections can then be referenced or included in agent-specific prompts.

### Creating Tool Templates

Tool templates transform metadata into tool-specific frontmatter:

```bash
# File: plugins/go-ent/agents/templates/toolname.yaml.tmpl
---
name: {{.Name}}
description: "{{.Description}}"
tools:
{{- range .Tools }}
  {{.}}: true
{{- end }}
model: {{.Model}}
color: "{{.Color}}"
tags:
  - "role:{{.Role}}"
  - "complexity:{{.Complexity}}"
skills:
{{- range .Skills }}
  - {{.}}
{{- end }}
---
```

### Testing Dependency Resolution

```bash
# Dry-run to see dependency chain
go-ent init --tool=claude --agents=architect --include-deps --dry-run

# Should show: planner, coder, tester, reviewer, debugger, architect

# Test without dependencies
go-ent init --tool=claude --agents=planner,coder --no-deps --dry-run

# Should show: planner, coder (no transitive deps)
```

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

### CLI Examples

**Initialize all agents**:
```bash
go-ent init --tool=claude
```

**Initialize specific agents**:
```bash
go-ent init --tool=claude --agents=planner,tester
```

**Initialize with dependencies**:
```bash
# Architect requires planner and coder
go-ent init --tool=claude --agents=architect --include-deps
# Generates: planner, coder, tester, reviewer, debugger, architect
```

**Preview changes**:
```bash
go-ent init --tool=claude --agents=architect --include-deps --dry-run
```

**Update existing configuration**:
```bash
go-ent init --tool=claude --update
```

**Custom model overrides**:
```bash
go-ent init --tool=claude --model heavy=opus --model main=sonnet
```

**Multiple tools**:
```bash
go-ent init --tool=all
```

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

1. **Initialize**: `/go-ent:init --tool=claude`
2. **Explore**: `openspec list`, `openspec list --specs`
3. **Plan**: `/go-ent:plan "feature description"`
4. **Approve**: Review at each of 4 wait points
5. **Sync**: `/go-ent:registry sync`
6. **Execute**: `/go-ent:apply` (or implement manually)
7. **Archive**: `/go-ent:archive <change-id>`

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
