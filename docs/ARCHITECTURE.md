# go-ent Architecture

System architecture for go-ent v0.3.0 (Architecture v2.0).

---

## Table of Contents

- [Overview](#overview)
- [System Components](#system-components)
- [Data Flow](#data-flow)
- [Layer Responsibilities](#layer-responsibilities)
- [Design Decisions](#design-decisions)
- [Technology Stack](#technology-stack)

---

## Overview

go-ent is an **enterprise Go development toolkit** that provides:

- **MCP Server**: Spec and workflow management via Model Context Protocol
- **Plugin System**: Claude Code integration with agents, skills, and commands
- **OpenSpec Workflow**: Spec-driven development with change tracking
- **Self-Hosted**: Uses itself for development (dogfooding)

### Architecture Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                    Claude Code Plugin                         │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐          │
│  │ 17 Agents    │ │ 16 Commands  │ │ 15 Skills    │          │
│  └──────────────┘ └──────────────┘ └──────────────┘          │
└─────────────────────────┬────────────────────────────────────┘
                          │ MCP Protocol (stdio)
                          ▼
┌──────────────────────────────────────────────────────────────┐
│                    MCP Server (ent binary)                    │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ Tools: spec_*, registry_*, workflow_*, etc.            │  │
│  └────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ Domain: Spec management, Registry, Workflow state      │  │
│  └────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ CLI: Standalone commands (init, config, skill, etc.)   │  │
│  └────────────────────────────────────────────────────────┘  │
└─────────────────────────┬────────────────────────────────────┘
                          │ File I/O + BoltDB
                          ▼
┌──────────────────────────────────────────────────────────────┐
│                    Project Structure                          │
│  openspec/                                                    │
│  ├── config.yaml         # OpenSpec configuration            │
│  ├── specs/              # Source of truth (deployed state)  │
│  ├── changes/            # Active change proposals           │
│  │   ├── <change-id>/                                        │
│  │   │   ├── proposal.md                                     │
│  │   │   ├── specs/     # Delta specs                        │
│  │   │   └── tasks.md                                        │
│  │   └── archive/       # Completed changes                  │
│  └── .go-ent/                                                 │
│      ├── config.yaml    # Project configuration              │
│      └── state.db       # BoltDB (registry, workflow state)  │
└──────────────────────────────────────────────────────────────┘
```

---

## System Components

### 1. MCP Server (`cmd/ent`)

**Purpose**: Provide tools for spec and workflow management.

**Key Files**:
- `cmd/ent/main.go` - Entry point with MCP server and CLI fallback
- `internal/mcp/server/` - MCP server setup
- `internal/mcp/tools/` - Tool handlers

**MCP Tools** (30+ tools):
- **Spec management**: `spec_init`, `spec_create`, `spec_list`, `spec_show`, `spec_update`, `spec_delete`, `spec_validate`, `spec_archive`
- **Registry**: `registry_list`, `registry_update`, `registry_next`
- **Workflow**: `workflow_state`, `workflow_start`, `workflow_complete`

**Protocol**: Uses stdin/stdout for communication with Claude Code.

---

### 2. Claude Code Plugin (`plugins/go-ent`)

**Purpose**: Integrate go-ent with Claude Code.

**Structure**:
```
plugins/go-ent/
├── .claude-plugin/
│   ├── plugin.json           # MCP configuration
│   └── marketplace.json      # Marketplace metadata
├── agents/
│   ├── meta/                 # Agent definitions (YAML)
│   └── prompts/              # Agent prompts (Markdown)
├── commands/                 # Slash commands
└── skills/                   # Skill definitions
```

**Components**:
- **17 Agents**: Specialized roles (architect, coder, planner, etc.)
- **16 Commands**: Workflow shortcuts (`/ent:plan`, `/ent:task`, etc.)
- **15 Skills**: Domain knowledge (go-code, go-api, go-arch, etc.)

---

### 3. Domain Layer (`internal/domain`)

**Purpose**: Core business logic and types.

**Key Packages**:
- `internal/domain/agent.go` - Agent types and resolution
- `internal/domain/skill.go` - Skill types and validation
- `internal/spec/` - Spec management domain
- `internal/config/` - Configuration management

**Domain Types**:
```go
type Agent struct {
    ID          string
    Name        string
    Model       string
    Skills      []string
    Temperature float64
}

type Skill struct {
    ID           string
    Name         string
    Category     string
    QualityScore int
    Triggers     []string
}

type Spec struct {
    Type    string // "spec" or "change"
    ID      string
    Content string
}
```

---

### 4. OpenSpec Structure

**Purpose**: Organize specifications and changes.

#### Specs (Source of Truth)

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

**Characteristics**:
- Reflect **current production state**
- Updated only when changes are deployed
- Versioned with codebase

#### Changes (Proposals)

```
openspec/changes/
├── add-auth-middleware/
│   ├── proposal.md       # Problem, solution, alternatives
│   ├── design.md         # Technical design (optional)
│   ├── tasks.md          # Task breakdown
│   └── specs/            # Delta specs
│       └── api/
│           └── middleware.md
└── archive/              # Completed changes
```

**Lifecycle**: Created → In Progress → Complete → Archived

---

## Data Flow

### 1. Spec Management Flow

```
┌─────────┐     MCP      ┌─────────┐     Domain     ┌──────────┐
│ Claude  │────Call─────>│   MCP   │────Logic──────>│  Spec    │
│  Code   │              │ Server  │                 │ Manager  │
└─────────┘     Result    └─────────┘     CRUD       └──────────┘
     ▲           <─────         │                          │
     │                          │                          │ File I/O
     └──────────────────────────┴──────────────────────────┴──────>
                                                      openspec/
```

**Steps**:
1. User calls command (`/ent:plan`) or agent uses tool
2. Claude Code invokes MCP tool (`spec_create`)
3. MCP server validates request
4. Domain layer performs business logic
5. File system updated
6. Result returned to Claude Code

---

### 2. Task Execution Flow

```
┌────────────┐                 ┌──────────┐                 ┌─────────┐
│  Agent     │───Get Next─────>│ Registry │───Query────────>│ BoltDB  │
│  (Coder)   │                 │  Tools   │                 │ State   │
└────────────┘                 └──────────┘                 └─────────┘
      │                              │
      │ Execute                      │ Update Status
      ▼                              ▼
┌────────────┐                 ┌──────────┐
│   Code     │◀────Read────────│  Spec    │
│ Repository │                 │  Files   │
└────────────┘                 └──────────┘
```

**Steps**:
1. Agent requests next task (`registry_next`)
2. Registry queries BoltDB for available tasks
3. Task loaded from change directory
4. Agent reads relevant specs
5. Agent executes task
6. Agent updates task status
7. Changes committed to repository

---

### 3. Workflow State Machine

```
        start
          │
          ▼
   ┌──────────┐
   │  Draft   │──────────────┐
   └──────────┘              │ reject
          │ approve           │
          ▼                   ▼
   ┌──────────┐         ┌──────────┐
   │  Review  │         │ Revise   │
   └──────────┘         └──────────┘
          │ approve           │
          ▼                   │
   ┌──────────┐              │
   │Implement │◀─────────────┘
   └──────────┘
          │ complete
          ▼
   ┌──────────┐
   │ Archive  │
   └──────────┘
```

**States**:
- **Draft**: Proposal being written
- **Review**: Awaiting human approval
- **Revise**: Rejected, needs changes
- **Implement**: Tasks being executed
- **Archive**: Completed, specs merged

---

## Layer Responsibilities

### Clean Architecture Layers

```
┌────────────────────────────────────────┐
│           Transport Layer              │  MCP tools, CLI commands
│  (internal/mcp/tools, internal/cli)   │
└────────────────┬───────────────────────┘
                 │ calls
┌────────────────▼───────────────────────┐
│           Use Case Layer               │  Workflow orchestration
│  (internal/workflow, internal/agent)  │  Agent execution
└────────────────┬───────────────────────┘
                 │ uses
┌────────────────▼───────────────────────┐
│           Domain Layer                 │  Business logic
│  (internal/spec, internal/domain)     │  Core types
└────────────────┬───────────────────────┘
                 │ depends on abstractions
┌────────────────▼───────────────────────┐
│         Infrastructure Layer           │  File I/O, BoltDB
│  (internal/store, internal/fs)        │  External systems
└────────────────────────────────────────┘
```

**Dependency Rule**: Dependencies flow inward. Inner layers never depend on outer layers.

---

### Layer Details

#### Transport Layer

**Responsibility**: Handle external requests (MCP, CLI).

**Examples**:
- `internal/mcp/tools/spec.go` - MCP tool handlers
- `internal/cli/init.go` - CLI commands

**Does NOT**:
- Contain business logic
- Access infrastructure directly

---

#### Use Case Layer

**Responsibility**: Orchestrate domain operations.

**Examples**:
- `internal/workflow/` - Workflow state management
- `internal/agent/executor.go` - Agent task execution

**Does**:
- Coordinate multiple domain operations
- Manage transactions
- Handle authorization

---

#### Domain Layer

**Responsibility**: Core business logic.

**Examples**:
- `internal/spec/manager.go` - Spec CRUD operations
- `internal/domain/agent.go` - Agent resolution

**Does**:
- Validate business rules
- Define core types
- Encapsulate domain knowledge

**Does NOT**:
- Depend on infrastructure
- Know about transport mechanisms

---

#### Infrastructure Layer

**Responsibility**: External system integration.

**Examples**:
- `internal/store/bolt.go` - BoltDB persistence
- `internal/fs/` - File system operations

**Does**:
- Implement repository interfaces
- Handle external dependencies
- Manage connections

---

## Design Decisions

### 1. MCP Server Architecture

**Decision**: Use MCP protocol for Claude Code integration.

**Rationale**:
- Industry standard (Anthropic, OpenAI, AWS support)
- Language-agnostic
- Bidirectional communication
- Tool discovery

**Trade-offs**:
- ✅ Pro: Wide compatibility
- ✅ Pro: Future-proof
- ❌ Con: Requires running process

---

### 2. OpenSpec Workflow

**Decision**: Separate specs (truth) from changes (proposals).

**Rationale**:
- **Brownfield-first**: Works with existing codebases
- **Audit trail**: Complete change history
- **Iterative**: Changes evolve through stages
- **Review gates**: Human oversight

**Trade-offs**:
- ✅ Pro: Clear separation of concerns
- ✅ Pro: Change tracking
- ❌ Con: More file structure

---

### 3. Embedded vs. Runtime Templates

**Decision**: Hybrid approach - embedded structure templates, AI-generated business logic.

**Rationale**:
- Templates ensure **structural consistency**
- AI generates **domain-specific logic**
- Best of both worlds

**Trade-offs**:
- ✅ Pro: Consistency + Flexibility
- ✅ Pro: Reduces token usage
- ❌ Con: More complex system

---

### 4. BoltDB for State

**Decision**: Use BoltDB for workflow and registry state.

**Rationale**:
- **Embedded**: No external database
- **ACID**: Reliable transactions
- **Fast**: In-process access
- **Simple**: Single file

**Trade-offs**:
- ✅ Pro: Zero configuration
- ✅ Pro: Portable
- ❌ Con: Single writer (not an issue for local dev)

---

### 5. Split-File Agent Format

**Decision**: Separate agent metadata (YAML) from prompts (Markdown).

**Rationale**:
- **Clarity**: Easy to read/edit
- **Reusability**: Share prompt fragments
- **Validation**: YAML schema validation
- **Version control**: Better diffs

**Trade-offs**:
- ✅ Pro: Maintainable
- ✅ Pro: Composable
- ❌ Con: Two files per agent

---

### 6. Progressive Skill Loading

**Decision**: Load skills in three stages (metadata → core → extended).

**Rationale**:
- **Token efficiency**: Load only what's needed
- **Performance**: Faster initial load
- **Scalability**: Support 100+ skills

**Stages**:
1. **Metadata** (150 tokens): ID, triggers, quality score
2. **Core** (500 tokens): Core expertise, common patterns
3. **Extended** (1500 tokens): Advanced topics, references

**Trade-offs**:
- ✅ Pro: 70-90% token reduction
- ✅ Pro: Faster responses
- ❌ Con: More complex loading logic

---

## Technology Stack

### Core Technologies

| Technology | Purpose | Version |
|------------|---------|---------|
| **Go** | Primary language | 1.24+ |
| **MCP Protocol** | AI tool integration | 2025 spec |
| **BoltDB** | State persistence | v1.3.7 |
| **Markdown** | Documentation format | CommonMark |
| **YAML** | Configuration format | 1.2 |

### Go Libraries

| Library | Purpose |
|---------|---------|
| `github.com/mark3labs/mcp-go` | MCP server SDK |
| `go.etcd.io/bbolt` | Embedded database |
| `github.com/caarlos0/env/v11` | Environment config |
| `gopkg.in/yaml.v3` | YAML parsing |

### Development Tools

| Tool | Purpose |
|------|---------|
| `golangci-lint` | Linting |
| `gofumpt` | Formatting |
| `testify` | Testing assertions |
| `make` | Build automation |

---

## Project Structure

```
go-ent/
├── cmd/
│   └── ent/                   # MCP server + CLI entry point
│       └── main.go
├── internal/
│   ├── cli/                   # CLI commands
│   ├── config/                # Configuration management
│   ├── domain/                # Core domain types
│   ├── mcp/
│   │   ├── server/            # MCP server setup
│   │   └── tools/             # Tool handlers
│   ├── spec/                  # Spec management domain
│   ├── store/                 # BoltDB persistence
│   └── workflow/              # Workflow orchestration
├── plugins/go-ent/            # Claude Code plugin
│   ├── agents/
│   │   ├── meta/              # Agent YAML definitions
│   │   └── prompts/           # Agent Markdown prompts
│   ├── commands/              # Slash commands
│   ├── skills/                # Skill definitions
│   └── .claude-plugin/
│       ├── plugin.json        # MCP configuration
│       └── marketplace.json   # Marketplace metadata
├── openspec/                  # Self-hosted development
│   ├── config.yaml
│   ├── specs/
│   ├── changes/
│   └── changes/archive/
├── docs/                      # Documentation
└── scripts/                   # Build/run scripts
```

---

## See Also

- [Architecture Review](./ARCHITECTURE_REVIEW.md) - Detailed analysis and recommendations
- [Development Guide](./DEVELOPMENT.md) - Development workflow
- [OpenSpec Workflow](./OPENSPEC_WORKFLOW.md) - Spec-driven development
- [MCP API Reference](./MCP_API.md) *(coming soon)* - Tool API documentation

---

**Version:** v0.3.0
**Last updated:** 2026-01-28
