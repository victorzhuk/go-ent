# Agents and Skills System

Comprehensive guide to go-ent's agent and skill architecture.

---

## Table of Contents

- [Overview](#overview)
- [Agent Architecture](#agent-architecture)
- [Available Agents](#available-agents)
- [Skill System](#skill-system)
- [Available Skills](#available-skills)
- [Tool Presets](#tool-presets)
- [Usage Patterns](#usage-patterns)

---

## Overview

go-ent uses a **multi-agent system** where specialized agents collaborate to accomplish development tasks. Each agent has:

- **Domain expertise** (skills)
- **Tool access** (tool presets)
- **Delegation chains** (dependencies)
- **Model assignment** (fast/main/heavy)

### Design Principles

1. **Separation of Concerns**: Each agent has a specific role
2. **Composability**: Agents delegate to other agents
3. **Progressive Complexity**: Fast/main/heavy models for different task types
4. **Skill-Based Expertise**: Agents activated based on required skills

---

## Agent Architecture

### Split-File Format

Agents use a **split-file architecture** for better maintainability:

```
plugins/go-ent/agents/
├── meta/                    # Agent metadata (YAML)
│   ├── architect.yaml
│   ├── coder.yaml
│   ├── planner.yaml
│   └── ...
├── prompts/                 # Agent prompts (Markdown)
│   ├── shared/              # Reusable sections
│   │   ├── _principals.md
│   │   ├── _judgment.md
│   │   └── _handoffs.md
│   └── agents/              # Agent-specific prompts
│       ├── architect.md
│       ├── coder.md
│       └── ...
└── templates/               # Runtime frontmatter templates
    ├── claude.yaml.tmpl
    └── opencode.yaml.tmpl
```

### Agent Metadata (YAML)

**File**: `meta/<agent>.yaml`

```yaml
name: ent:architect
description: System architect. Designs components, layers, data flow.
model: heavy              # fast, main, or heavy
color: "#4169E1"          # UI color
role: planning            # planning or execution
complexity: heavy         # fast, standard, or heavy

skills:                   # Skill IDs to load
  - go-arch
  - go-api
  - api-design

toolPresets:              # Tool access presets
  - read-only
  - serena-analysis

dependencies:             # Agents this can delegate to
  - ent:planner
  - ent:coder
```

**Fields**:
- `name`: Unique agent ID (format: `plugin:agent`)
- `description`: Short description (shown in UI)
- `model`: Model category (fast/main/heavy)
- `color`: Hex color for UI
- `role`: Agent role (planning/execution)
- `complexity`: Complexity level
- `skills`: Skills to activate
- `toolPresets`: Tools the agent can use
- `dependencies`: Agents that can be delegated to

### Agent Prompts (Markdown)

**File**: `prompts/agents/<agent>.md`

Prompts are composed from:
1. **Shared sections**: Reusable prompt fragments
2. **Agent-specific**: Unique instructions

**Example**:
```markdown
# Architect Agent

You are a system architect specializing in Go applications.

{{> _principals }}      <!-- Shared: Core principles -->
{{> _judgment }}        <!-- Shared: Decision framework -->

## Your Responsibilities

- Design system architecture
- Define component boundaries
- Plan data flow
- Review technical decisions

## Delegation

{{> _handoffs }}        <!-- Shared: Handoff patterns -->

When you need implementation, delegate to @ent:coder.
When you need planning, delegate to @ent:planner.
```

### Template Generation

At runtime, templates combine metadata and prompts for specific tools:

**Claude Code** (`.claude/agents/ent:architect.yaml`):
```yaml
name: ent:architect
description: System architect. Designs components, layers, data flow.
model: claude-opus-4.5
color: "#4169E1"
instructions: |
  # Architect Agent
  You are a system architect...
  [prompt content]
```

**OpenCode** (`.opencode/agents/ent:architect.yaml`):
```yaml
name: architect
role: planning
model: heavy
skills:
  - go-arch
  - go-api
prompt: |
  # Architect Agent
  [prompt content]
```

---

## Available Agents

### Planning Agents

#### `ent:architect`

**Purpose**: System architecture and design

**Model**: Heavy (Claude Opus)

**Skills**: `go-arch`, `go-api`, `api-design`

**Tools**: Read-only + Serena analysis

**Use Cases**:
- Design new system components
- Define layer boundaries
- Plan data flow
- Review architectural decisions

**Delegates To**: `planner`, `coder`

**Example**:
```
/ent:architect Design authentication system with JWT
```

---

#### `ent:planner`

**Purpose**: Task breakdown and planning

**Model**: Main (Claude Sonnet)

**Skills**: None (general planning)

**Tools**: Read-only + Serena analysis

**Use Cases**:
- Break features into tasks
- Analyze task dependencies
- Estimate complexity
- Create implementation plan

**Delegates To**: `coder`, `tester`

**Example**:
```
/ent:planner Break down user authentication feature
```

---

#### `ent:planner-fast`

**Purpose**: Quick feasibility assessment

**Model**: Fast (Claude Haiku)

**Skills**: None

**Tools**: Read-only + Serena analysis

**Use Cases**:
- Quick feasibility check
- Simple task triage
- Rapid estimates

**Delegates To**: `planner`, `coder`

**Example**:
```
/ent:planner-fast Can we add rate limiting?
```

---

#### `ent:planner-heavy`

**Purpose**: Complex architectural planning

**Model**: Heavy (Claude Opus)

**Skills**: `go-arch`, `api-design`

**Tools**: Read-only + Serena analysis

**Use Cases**:
- Complex system design
- Multi-component planning
- Architecture refactoring

**Delegates To**: `architect`, `coder`

**Example**:
```
/ent:planner-heavy Plan microservices migration
```

---

#### `ent:decomposer`

**Purpose**: Task breakdown with dependency analysis

**Model**: Main (Claude Sonnet)

**Skills**: None

**Tools**: Read-only + Serena analysis

**Use Cases**:
- Break down epics
- Identify task dependencies
- Create task DAG
- Estimate parallelization

**Delegates To**: `planner`

**Example**:
```
/ent:decomposer Decompose payment system integration
```

---

### Execution Agents

#### `ent:coder`

**Purpose**: Go implementation and coding

**Model**: Main (Claude Sonnet)

**Skills**: `go-code`, `go-db`

**Tools**: Editing + Serena analysis (but not Serena editing)

**Use Cases**:
- Implement features
- Write Go code
- Refactor code
- Add functionality

**Delegates To**: `tester`, `reviewer`, `debugger`

**Example**:
```
/ent:coder Implement User repository with CRUD operations
```

---

#### `ent:tester`

**Purpose**: Test engineering and TDD

**Model**: Main (Claude Sonnet)

**Skills**: `go-test`

**Tools**: Editing + Serena analysis + Bash

**Use Cases**:
- Write unit tests
- Write integration tests
- TDD red-green-refactor
- Test coverage improvement

**Delegates To**: `coder`, `debugger`

**Example**:
```
/ent:tester Write tests for User service
```

---

### Analysis Agents

#### `ent:debugger`

**Purpose**: Standard debugging

**Model**: Main (Claude Sonnet)

**Skills**: `debug-core`

**Tools**: Read-only + Serena analysis + Bash

**Use Cases**:
- Investigate bugs
- Analyze stack traces
- Debug failing tests
- Root cause analysis

**Delegates To**: `reproducer`, `coder`

**Example**:
```
/ent:debugger Why is authentication failing?
```

---

#### `ent:debugger-fast`

**Purpose**: Quick debugging for simple issues

**Model**: Fast (Claude Haiku)

**Skills**: None

**Tools**: Read-only + Bash

**Use Cases**:
- Simple bug fixes
- Obvious errors
- Quick troubleshooting

**Delegates To**: `debugger`, `coder`

---

#### `ent:debugger-heavy`

**Purpose**: Complex debugging

**Model**: Heavy (Claude Opus)

**Skills**: `debug-core`, `go-perf`

**Tools**: Read-only + Serena analysis + Bash

**Use Cases**:
- Concurrency issues
- Performance problems
- Multi-component bugs
- Race conditions

**Delegates To**: `reproducer`, `debugger`

---

#### `ent:reproducer`

**Purpose**: Create minimal bug reproductions

**Model**: Main (Claude Sonnet)

**Skills**: `go-test`

**Tools**: Editing + Bash

**Use Cases**:
- Create minimal repro
- Write failing test
- Isolate bug cause

**Delegates To**: `debugger`, `tester`

---

#### `ent:researcher`

**Purpose**: Research and investigation

**Model**: Main (Claude Sonnet)

**Skills**: None

**Tools**: Read-only + Serena analysis

**Use Cases**:
- Explore codebase
- Research patterns
- Understand architecture
- Gather context

**Delegates To**: `architect`, `planner`

---

### Quality Agents

#### `ent:reviewer`

**Purpose**: Code review

**Model**: Heavy (Claude Opus)

**Skills**: `review-core`, `security-core`

**Tools**: Read-only + Serena analysis

**Use Cases**:
- Review code quality
- Check security
- Validate patterns
- Ensure standards

**Delegates To**: `coder` (for fixes)

**Example**:
```
/ent:reviewer Review authentication implementation
```

---

#### `ent:acceptor`

**Purpose**: Validate acceptance criteria

**Model**: Main (Claude Sonnet)

**Skills**: None

**Tools**: Read-only + Bash

**Use Cases**:
- Verify acceptance criteria
- Check spec compliance
- Validate before archive

**Delegates To**: `tester`, `reviewer`

---

### Task Management Agents

#### `ent:task-fast`

**Purpose**: Quick task assessment

**Model**: Fast (Claude Haiku)

**Skills**: None

**Tools**: Read-only

**Use Cases**:
- Quick complexity check
- Fast triage
- Simple routing

**Delegates To**: `planner`, `coder`

---

#### `ent:task-heavy`

**Purpose**: Complex task analysis

**Model**: Heavy (Claude Opus)

**Skills**: `go-arch`

**Tools**: Read-only + Serena analysis

**Use Cases**:
- Complex task analysis
- Deep reasoning
- Architecture-heavy tasks

**Delegates To**: `architect`, `planner`

---

## Agent Delegation Chains

### Planning Workflow

```
User Request
    ↓
ent:planner-fast (triage)
    ↓
ent:planner (standard planning)
    ↓
ent:architect (if architectural)
    ↓
ent:coder (implementation)
```

### Implementation Workflow

```
ent:coder (write code)
    ↓
ent:tester (write tests)
    ↓
ent:reviewer (review code)
    ↓
ent:debugger (if issues found)
    ↓
ent:coder (fix issues)
```

### Debugging Workflow

```
ent:debugger-fast (simple check)
    ↓
ent:debugger (standard debugging) ──> ent:reproducer (minimal repro)
    ↓                                          ↓
ent:debugger-heavy (complex)              ent:tester (test)
    ↓
ent:coder (fix)
```

---

## Skill System

### Skill Architecture

Skills provide **domain-specific knowledge** that agents can activate.

**Structure**:
```
plugins/go-ent/skills/
├── core/                  # Cross-language skills
│   ├── api-design/
│   ├── arch-core/
│   ├── debug-core/
│   ├── review-core/
│   └── security-core/
├── go/                    # Go-specific skills
│   ├── go-api/
│   ├── go-arch/
│   ├── go-code/
│   └── ...
└── plugins/               # Plugin development
    └── go-ent/
```

### Skill Format (v2)

**File**: `<category>/<skill-id>/SKILL.md`

```markdown
---
name: go-code
description: "Modern Go implementation patterns..."
version: "2.0.0"
author: "go-ent"
tags: ["go", "code", "implementation"]
---

<triggers>
keywords:
  - "go code"
  - "golang"
  - "implementation"
file_pattern: "*.go"
weight: 0.8
</triggers>

<role>
Expert Go developer focused on clean architecture...
</role>

<instructions>
## Pattern 1
...

## Pattern 2
...
</instructions>

<references>
See references/ directory for complete examples
</references>
```

### Progressive Loading

Skills load in **three stages** to optimize token usage:

1. **Metadata** (150 tokens):
   - ID, triggers, quality score
   - Used for skill matching

2. **Core** (500 tokens):
   - `<role>` section
   - `<instructions>` section
   - Essential patterns

3. **Extended** (1500 tokens):
   - `<references>` section
   - Advanced examples
   - Complete documentation

**Token Savings**: 70-90% reduction compared to loading all content upfront.

### Quality Scoring

Each skill has a **quality score** (0-100) based on:

- **Completeness** (30%): Coverage of domain
- **Clarity** (25%): Clear, well-structured content
- **Accuracy** (25%): Correct, up-to-date information
- **Examples** (20%): Practical, working examples

**Minimum threshold**: 70 (configurable)

---

## Available Skills

### Core Skills (Cross-Language)

#### `api-design`

**Category**: core

**Quality**: 95

**Triggers**: API, endpoint, REST, GraphQL

**Description**: Expert in REST and GraphQL API design, versioning, documentation.

**Key Topics**:
- Resource modeling
- HTTP method semantics
- Status codes
- Versioning strategies
- API documentation

---

#### `arch-core`

**Category**: core

**Quality**: 92

**Triggers**: architecture, design, layers

**Description**: Clean Architecture, DDD, SOLID principles.

**Key Topics**:
- Layer separation
- Dependency inversion
- Domain-driven design
- Bounded contexts
- Hexagonal architecture

---

#### `debug-core`

**Category**: core

**Quality**: 88

**Triggers**: debug, troubleshoot, investigate

**Description**: Systematic debugging, root cause analysis.

**Key Topics**:
- Reproduction steps
- Stack trace analysis
- Hypothesis testing
- Logging strategies
- Performance profiling

---

#### `review-core`

**Category**: core

**Quality**: 90

**Triggers**: review, quality, standards

**Description**: Code review best practices, quality metrics.

**Key Topics**:
- Review checklist
- Common issues
- Constructive feedback
- Security review
- Performance review

---

#### `security-core`

**Category**: core

**Quality**: 93

**Triggers**: security, vulnerability, attack

**Description**: Security best practices, OWASP Top 10, threat modeling.

**Key Topics**:
- Authentication/Authorization
- Input validation
- SQL injection prevention
- XSS prevention
- Secrets management

---

### Go Skills

#### `go-api`

**Category**: go

**Quality**: 95

**Triggers**: API, handler, HTTP, REST

**Description**: Production-ready Go REST and GraphQL APIs.

**Key Topics**:
- HTTP handler patterns
- Middleware implementation
- Request validation
- Error responses
- OpenAPI/Swagger

---

#### `go-arch`

**Category**: go

**Quality**: 94

**Triggers**: architecture, clean, layers

**Description**: Go Clean Architecture with DDD.

**Key Topics**:
- Layer structure (Transport/UseCase/Domain/Infra)
- Dependency injection
- Repository pattern
- Interface design
- Package organization

---

#### `go-code`

**Category**: go

**Quality**: 92

**Triggers**: implement, code, golang

**Description**: Modern Go patterns, error handling, concurrency.

**Key Topics**:
- Error handling
- Concurrency patterns
- Configuration management
- Logging
- Testing patterns

---

#### `go-config`

**Category**: go

**Quality**: 88

**Triggers**: config, environment, settings

**Description**: Configuration management with env vars, validation.

**Key Topics**:
- env/v11 usage
- Validation
- Defaults
- Type safety
- Testability

---

#### `go-db`

**Category**: go

**Quality**: 91

**Triggers**: database, postgres, SQL

**Description**: Database access with pgx, squirrel, migrations.

**Key Topics**:
- Repository pattern
- Query building (squirrel)
- Transaction management
- Error mapping
- Connection pooling

---

#### `go-error`

**Category**: go

**Quality**: 87

**Triggers**: error, handling, wrap

**Description**: Go error handling patterns and best practices.

**Key Topics**:
- Error wrapping
- Custom errors
- Error checking
- Sentinel errors
- Error context

---

#### `go-migration`

**Category**: go

**Quality**: 85

**Triggers**: migration, schema, goose

**Description**: Database migrations with goose.

**Key Topics**:
- Migration structure
- Versioning
- Rollback strategies
- Testing migrations
- Production safety

---

#### `go-ops`

**Category**: go

**Quality**: 89

**Triggers**: deployment, docker, production

**Description**: Go application operations, Docker, health checks.

**Key Topics**:
- Graceful shutdown
- Health checks
- Metrics (Prometheus)
- Docker images
- Production readiness

---

#### `go-perf`

**Category**: go

**Quality**: 86

**Triggers**: performance, optimize, benchmark

**Description**: Go performance optimization and profiling.

**Key Topics**:
- Profiling (CPU, memory)
- Benchmarking
- Memory optimization
- Concurrency tuning
- Hot path optimization

---

#### `go-review`

**Category**: go

**Quality**: 88

**Triggers**: review, quality, Go

**Description**: Go-specific code review patterns.

**Key Topics**:
- Go idioms
- Common mistakes
- Performance issues
- Security concerns
- Style guide

---

#### `go-sec`

**Category**: go

**Quality**: 90

**Triggers**: security, Go, vulnerability

**Description**: Go security best practices.

**Key Topics**:
- Input validation
- SQL injection (with pgx)
- Authentication
- Secrets management
- Dependency security

---

#### `go-test`

**Category**: go

**Quality**: 91

**Triggers**: test, TDD, testing

**Description**: Go testing patterns, TDD, test organization.

**Key Topics**:
- Table-driven tests
- Mocking strategies
- Integration tests (testcontainers)
- Test coverage
- Benchmark tests

---

### Plugin Skills

#### `go-ent`

**Category**: plugins

**Quality**: 85

**Triggers**: go-ent, plugin, agent

**Description**: go-ent plugin development.

**Key Topics**:
- Agent creation
- Skill authoring
- Command development
- MCP tools
- Plugin structure

---

## Tool Presets

Tool presets define **what tools an agent can use**.

### Available Presets

#### `read-only`

**Tools**:
- `Read`, `Glob`, `Grep`
- Basic code exploration

**Use**: Planning, research, analysis

---

#### `editing`

**Tools**:
- `Read`, `Write`, `Edit`
- `Glob`, `Grep`
- File manipulation

**Use**: Implementation, coding

---

#### `serena-analysis`

**Tools**:
- All Serena read tools
  - `find_symbol`
  - `find_referencing_symbols`
  - `get_symbols_overview`
  - `search_for_pattern`

**Use**: Deep code analysis

---

#### `serena-editing`

**Tools**:
- Serena write tools
  - `replace_symbol_body`
  - `insert_after_symbol`
  - `rename_symbol`

**Use**: Symbol-level editing

**Note**: Most agents use standard `editing` instead.

---

## Usage Patterns

### Direct Agent Invocation

Call an agent directly:

```
/ent:architect Design user authentication system
/ent:coder Implement User repository
/ent:tester Write tests for UserService
/ent:reviewer Review auth implementation
```

### Workflow Commands with Agents

Agents are invoked automatically by workflows:

```
# Planning workflow uses: planner → architect → decomposer
/ent:plan Add user authentication

# Execution uses: coder → tester → reviewer
/ent:task

# Debugging uses: debugger → reproducer → coder
/ent:bug Auth returns 500 error
```

### Agent Selection Logic

Agents are selected based on:

1. **Explicit invocation** (`/ent:coder`)
2. **Required skills** (task needs `go-api` → agents with that skill)
3. **Complexity** (architectural → heavy model)
4. **Delegation chain** (coder → tester → reviewer)

### Skill Activation

Skills are activated when:

1. **Agent specifies** (in `skills` list)
2. **Context matches** (editing `*.go` files)
3. **User requests** (mentions skill triggers)

**Example**:
```
Task: "Implement REST API for users"

Matched skills:
- go-api (trigger: "REST API")
- go-code (trigger: "implement")
- api-design (trigger: "API")

Selected agent: ent:coder
  - Has: go-code skill ✓
  - Delegates to: ent:architect (has api-design) ✓
```

---

## Configuration

### Agent Configuration

**File**: `.go-ent/config.yaml`

```yaml
agents:
  roles:
    architect:
      model: heavy
      skills:
        - go-arch
        - api-design
      temperature: 0.3

    coder:
      model: main
      skills:
        - go-code
        - go-db
      temperature: 0.7
```

### Skill Configuration

```yaml
skills:
  enabled: true
  progressive_load: true
  min_quality: 70

  directories:
    - plugins/go-ent/skills
    - custom/skills
```

---

## See Also

- [OpenSpec Workflow](./OPENSPEC_WORKFLOW.md) - How agents use OpenSpec
- [Skill Authoring](./SKILL-AUTHORING.md) - Creating new skills
- [Development Guide](./DEVELOPMENT.md) - Adding new agents
- [Configuration Reference](./CONFIGURATION.md) - Agent/skill config

---

**Version:** v0.3.0
**Last updated:** 2026-01-28
