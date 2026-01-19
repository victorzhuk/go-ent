<p align="center">
  <img src="assets/go-ent-logo.png" alt="go-ent mascot" width="280">
</p>

<h1 align="center">Go Ent</h1>

<p align="center">
  <em>Enterprise Go development toolkit for Claude Code with Clean Architecture, SOLID principles, and spec-driven development via MCP.</em>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/victorzhuk/go-ent"><img src="https://pkg.go.dev/badge/github.com/victorzhuk/go-ent.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/victorzhuk/go-ent"><img src="https://goreportcard.com/badge/github.com/victorzhuk/go-ent" alt="Go Report Card"></a>
  <a href="https://github.com/victorzhuk/go-ent/actions/workflows/validate.yml"><img src="https://github.com/victorzhuk/go-ent/actions/workflows/validate.yml/badge.svg" alt="CI"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <img src="https://img.shields.io/badge/go-%3E%3D1.24-blue" alt="Go 1.24+">
</p>

---

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Architecture v2.0](#architecture-v20)
- [MCP Tools](#mcp-tools)
- [CLI Commands](#cli-commands)
- [Skill Template System](#skill-template-system)
- [Available Commands](#available-commands)
- [Available Agents](#available-agents)
- [Skills](#skills-auto-activated)
- [Building from Source](#building-from-source)
- [Project Structure](#project-structure)
- [Standards Enforced](#standards-enforced)
- [How It Works](#how-it-works-v20)
- [Migration from v1.x](#migration-from-v1x)
- [Contributing](#contributing)
- [License](#license)
- [References](#references)

---

> [!IMPORTANT]
> **Architecture v2.0** (Current release: v0.3.0) - MCP server for spec-driven development with multi-agent orchestration.

## Features

- 🏗️ **Clean Architecture** patterns and enforcement
- 📝 **SOLID principles** validation
- 🔍 **Automated code review** with enterprise standards
- 🧪 **Testing patterns** (unit, integration, benchmarks)
- 📋 **Spec-driven development** with `openspec` folder management
- 🤖 **MCP server** for spec/task management tools
- 🔧 **Hooks** for automatic formatting and safety
- 🤖 **Specialized agents** (architect, debug, dev, lead, planner, reviewer, tester)
- ⚡ **Slash commands** for common workflows

## Quick Start

### 1. Install Plugin

Add the plugin source to your Claude Code settings:

```json
{
  "extraKnownMarketplaces": {
    "go-ent": {
      "source": {
        "source": "directory",
        "path": "/path/to/go-ent/plugins/go-ent"
      }
    }
  },
  "enabledPlugins": {
    "go-ent@go-ent": true
  }
}
```

Then restart Claude Code.

### 2. Using go-ent

**Via MCP (in Claude Code):**

Use MCP tools to manage your project specs:

```
# Initialize openspec folder in your project
Call go_ent_spec_init tool with path to your project

# Create a new spec
Call go_ent_spec_create tool with type="spec", id="user-auth", content="..."

# List all specs
Call go_ent_spec_list tool with type="spec"
```

**Via CLI (standalone):**

The `go-ent` binary can also be used as a standalone CLI:

```bash
# Initialize configuration
go-ent config init

# View configuration
go-ent config show

# List available agents
go-ent agent list

# Initialize OpenSpec
go-ent spec init
```

See [CLI Examples](docs/CLI_EXAMPLES.md) for detailed usage.

The LLM (Claude Code) will generate code based on specs and templates, not copy-paste them.

## Architecture v2.0

### MCP Server

The `go-ent` binary is now an MCP server that provides tools for managing `openspec` folders:

```
go-ent/
├── cmd/go-ent/              # MCP server
│   └── main.go             # stdio transport
├── internal/
│   ├── mcp/
│   │   ├── server/         # MCP setup
│   │   └── tools/          # Tool handlers (25 tools)
│   ├── spec/               # Spec management domain
│   ├── templates/          # Reference patterns (embedded)
│   └── domain/             # Core domain types
├── plugins/go-ent/          # Claude Code plugin
│   ├── agents/             # 7 agent definitions
│   ├── commands/           # 16 slash commands
│   ├── skills/             # 10 skill definitions
│   └── .claude-plugin/
│       ├── plugin.json     # MCP configuration
│       └── marketplace.json
└── openspec/               # Self-hosted development
    ├── project.yaml
    ├── specs/
    ├── changes/
    └── tasks/
```

### `openspec` Folder Structure

```
project/openspec/
├── project.yaml            # Project metadata
├── specs/                  # Capability specs
│   └── {capability}/
│       ├── spec.md
│       └── design.md
├── changes/                # Active changes
│   └── {change-id}/
│       ├── proposal.md
│       ├── tasks.md
│       └── design.md
├── tasks/                  # Standalone tasks
└── archive/                # Completed changes
```

## MCP Tools

### Spec Management

| Tool | Description |
|------|-------------|
| `go_ent_spec_init` | Initialize openspec folder in project |
| `go_ent_spec_list` | List specs, changes, or tasks |
| `go_ent_spec_show` | Show detailed content |
| `go_ent_spec_create` | Create new spec/change/task |
| `go_ent_spec_update` | Update existing item |
| `go_ent_spec_delete` | Delete item |
| `go_ent_spec_validate` | Validate specs against rules |
| `go_ent_spec_archive` | Archive completed changes |

### Registry

| Tool | Description |
|------|-------------|
| `go_ent_registry_init` | Initialize task registry |
| `go_ent_registry_list` | List all tasks in registry |
| `go_ent_registry_next` | Get next recommended tasks |
| `go_ent_registry_update` | Update task status/priority |
| `go_ent_registry_deps` | Show task dependencies |
| `go_ent_registry_sync` | Sync tasks across proposals |

### Workflow

| Tool | Description |
|------|-------------|
| `go_ent_workflow_start` | Start planning workflow |
| `go_ent_workflow_status` | Get workflow status |
| `go_ent_workflow_approve` | Approve workflow step |

### Loop

| Tool | Description |
|------|-------------|
| `go_ent_loop_start` | Start autonomous loop |
| `go_ent_loop_get` | Get loop status |
| `go_ent_loop_set` | Update loop parameters |
| `go_ent_loop_cancel` | Cancel loop |

### Generation

| Tool | Description |
|------|-------------|
| `go_ent_generate` | Generate code from templates |
| `go_ent_generate_component` | Generate specific component |
| `go_ent_generate_from_spec` | Generate from OpenSpec |
| `go_ent_list_archetypes` | List available archetypes |

## CLI Commands

The `go-ent` binary can run in two modes:

1. **MCP Server Mode** (default): Communicates via stdio with Claude Code
2. **CLI Mode**: Standalone command-line interface for automation and scripting

### Configuration Management

```bash
# Initialize configuration
go-ent config init [path]

# Show current configuration
go-ent config show [path]
go-ent config show --format summary

# Modify configuration
go-ent config set <key> <value> [path]
go-ent config set budget.daily 25
go-ent config set agents.default architect
```

### Agent Management

```bash
# List all agents
go-ent agent list
go-ent agent list --detailed

# Get agent information
go-ent agent info <name>
go-ent agent info architect
```

### Skill Management

```bash
# List all skills
go-ent skill list
go-ent skill list --detailed

# Get skill information
go-ent skill info <name>
go-ent skill info go-arch

# Create new skill from template
go-ent skill new <name>              # Interactive mode
go-ent skill new go-payment \
  --template go-basic \
  --description "Payment processing"

# List available templates
go-ent skill list-templates
go-ent skill list-templates --category go
go-ent skill list-templates --built-in

# Show template details
go-ent skill show-template <name>
go-ent skill show-template go-complete

# Add custom template
go-ent skill add-template <path>
go-ent skill add-template ./my-template
```

### Spec Management

```bash
# Initialize OpenSpec
go-ent spec init [path]

# List specs or changes
go-ent spec list <type>  # type: spec, change
go-ent spec list spec
go-ent spec list change

# Show specific spec or change
go-ent spec show <type> <id>
go-ent spec show spec api
go-ent spec show change add-authentication
```

### Global Flags

```bash
# Use custom config file
go-ent --config /path/to/config.yaml <command>

# Verbose output
go-ent --verbose <command>

# Show version
go-ent version
```

**Full CLI documentation:** [CLI Examples](docs/CLI_EXAMPLES.md)

## Available Commands

> **Note:** The following are **slash commands** for use within Claude Code, not CLI commands.

### Planning & Workflow

| Command | Description |
|---------|-------------|
| `/plan <feature>` | Comprehensive planning workflow with research, design, and task decomposition |
| `/clarify <change-id>` | Ask focused questions to clarify underspecified requirements |
| `/research <change-id> [topic]` | Structured research phase for unknowns and technology decisions |
| `/decompose <change-id>` | Break proposal into dependency-aware, trackable tasks |
| `/analyze <change-id>` | Cross-document consistency validation (read-only) |

### Execution

| Command | Description |
|---------|-------------|
| `/apply` | Execute tasks from OpenSpec change proposal |
| `/loop <task-description> [--max-iterations=10]` | Start autonomous work loop with self-correction |
| `/loop-cancel` | Cancel running autonomous loop |
| `/tdd` | Test-driven development cycle (Red-Green-Refactor) |

### Project Management

| Command | Description |
|---------|-------------|
| `/init <project-name> [module-path] [--type=http\|mcp]` | Initialize a new Go enterprise project with Clean Architecture structure |
| `/scaffold <type> <name> [impl]` | Scaffold Go components (entity, repository, usecase, handler, service) |
| `/gen` | Generate code from OpenAPI/Proto specs |
| `/status` | View status of all OpenSpec changes |
| `/archive` | Archive completed OpenSpec change |
| `/registry` | Manage OpenSpec task registry |

### Quality

| Command | Description |
|---------|-------------|
| `/lint` | Run Go linters and fix issues |

## Available Agents

| Agent | Description |
|-------|-------------|
| `lead` (opus/gold) | Lead developer. Orchestrates workflow, delegates to specialists |
| `architect` (opus/blue) | System architect. Designs components, layers, data flow |
| `planner` (sonnet/green) | Task planner. Breaks features into actionable tasks |
| `dev` (sonnet/green) | Go developer. Implements features, writes code |
| `reviewer` (opus/blue) | Code reviewer. Reviews code for bugs, security, quality, and adherence to project conventions |
| `debug` (sonnet/red) | Debugger. Troubleshoots issues, analyzes errors |
| `tester` (haiku/cyan) | Test engineer. Writes tests, TDD cycles |

## Skill Template System

go-ent provides a template-based system for creating new skills quickly and consistently. Templates include pre-built patterns, validation, and quality standards for various programming languages and domains.

### Template Types

**Built-in Templates**: Shipped with go-ent in `plugins/go-ent/templates/skills/`
- go-basic, go-complete: Go development patterns
- typescript-basic: TypeScript-specific guidance
- testing: TDD and testing best practices
- database: SQL, migrations, and data access
- api-design: REST, GraphQL, and API patterns
- core-basic, arch: Architecture and system design
- debugging-basic: Troubleshooting and debugging
- security: Authentication, authorization, and security
- review: Code review practices

**Custom Templates**: User-defined templates in `~/.go-ent/templates/skills/`
- Add your own templates for team-specific patterns
- Share templates across projects
- Extend built-in functionality

### Creating Skills from Templates

**Interactive mode** (recommended):
```bash
go-ent skill new my-skill
```

This prompts for:
1. Template selection from available options
2. Skill description and metadata
3. Category (auto-detected from name prefix)

**Non-interactive mode**:
```bash
go-ent skill new go-payment \
  --template go-basic \
  --description "Payment processing patterns" \
  --category go \
  --author "your-name" \
  --tags "payment,api"
```

### Managing Templates

**List all available templates**:
```bash
go-ent skill list-templates
```

Filter by category:
```bash
go-ent skill list-templates --category go
```

Show only built-in or custom:
```bash
go-ent skill list-templates --built-in
go-ent skill list-templates --custom
```

**Show template details**:
```bash
go-ent skill show-template go-complete
```

Displays:
- Template metadata (name, category, version, author)
- Configuration prompts and defaults
- Preview of template content (first 20 lines)

**Add custom template**:
```bash
go-ent skill add-template ./my-custom-template
```

Template directory must contain:
- `template.md`: Skill template with v2 format
- `config.yaml`: Template metadata and prompt configuration

By default, templates are added to `~/.go-ent/templates/skills/`. Use `--built-in` flag to add to built-in directory (requires write permissions).

### Template Structure

Each template consists of two files:

**config.yaml** - Template metadata:
```yaml
name: go-complete
category: go
description: "Comprehensive Go development template"
version: "1.0.0"
author: "go-ent"
prompts:
  - key: NAME
    prompt: "Skill name"
    required: true
  - key: DESCRIPTION
    prompt: "Skill description"
    required: true
    default: "Custom Go skill"
```

**template.md** - V2 format skill with placeholders:
```markdown
---
name: ${NAME}
description: "${DESCRIPTION}"
version: "2.0.0"
author: "${AUTHOR}"
tags: ["go"]
---

# ${NAME}

<role>
Expert Go developer focused on clean architecture and patterns.
</role>

<instructions>
## Pattern 1

Code example...

**Why this pattern**:
- Reason 1
- Reason 2
</instructions>
...
```

### Auto-Detection

Category is automatically detected from skill name prefix:
- `go-payment` → `go` category
- `typescript-ui` → `typescript` category
- `db-migration` → `database` category

Output path: `plugins/go-ent/skills/<category>/<skill-name>/SKILL.md`

### Validation

Generated skills are automatically validated against:
- Frontmatter completeness
- XML tag structure
- Required sections (role, instructions, examples, edge_cases)
- Quality scoring (0-100 scale)

For detailed skill authoring guidance, see [SKILL-AUTHORING.md](docs/SKILL-AUTHORING.md).

## Skills (Auto-activated)

| Skill | Triggers |
|-------|----------|
| `go-hub` | Go development, backend services, Clean Architecture, OpenSpec workflow |
| `go-code` | Writing Go code, implementing features, refactoring, error handling, configuration |
| `go-arch` | Architecture decisions, system design, layer organization, dependency injection, bounded contexts |
| `go-api` | API design, OpenAPI specs, code generation, protobuf, REST endpoints, gRPC services |
| `go-db` | Database work, migrations, queries, repositories, caching |
| `go-test` | Writing tests, TDD, coverage, integration tests, mocks |
| `go-review` | Code review, quality checks, PR review, architecture validation |
| `go-perf` | Performance issues, profiling, optimization, memory leaks, benchmarking |
| `go-sec` | Security concerns, authentication, authorization, input validation, secrets |
| `go-ops` | Deployment, containerization, orchestration, CI/CD pipelines, infrastructure |

## Building from Source

```bash
# Clone repository
git clone https://github.com/victorzhuk/go-ent.git
cd go-ent

# Build MCP server
make build

# Binary will be in bin/go-ent
./bin/go-ent  # runs as MCP server on stdio
```

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build MCP server to `bin/go-ent` |
| `make test` | Run tests with race detector and coverage |
| `make lint` | Run golangci-lint |
| `make fmt` | Format code with goimports |
| `make clean` | Remove build artifacts |
| `make validate-plugin` | Validate plugin JSON files |
| `make help` | Show all available targets |

### Development Requirements

- Go 1.24 or later
- make
- golangci-lint (for `make lint`)
- jq (for `make validate-plugin`)

## Project Structure

### go-ent Repository

```
go-ent/
├── cmd/go-ent/              # MCP server binary
│   └── main.go
├── internal/
│   ├── mcp/
│   │   ├── server/           # MCP server setup
│   │   └── tools/           # 25 MCP tool handlers
│   ├── spec/                # OpenSpec domain logic
│   ├── templates/           # Code generation templates
│   ├── domain/              # Core domain types
│   ├── generation/          # Code generation engine
│   ├── config/              # Configuration system
│   └── version/             # Version metadata
├── plugins/go-ent/          # Claude Code plugin
│   ├── agents/              # Agent role definitions (7)
│   ├── commands/            # Slash commands (16)
│   ├── skills/              # Skill definitions (10)
│   └── .claude-plugin/      # Plugin config
├── openspec/                # Self-hosted development specs
│   ├── project.yaml
│   ├── specs/
│   ├── changes/
│   └── tasks/
├── docs/                    # Additional documentation
├── assets/                  # Logo and branding
└── Makefile                 # Build targets
```

### Generated Projects

Generated projects follow Clean Architecture:

```
project/
├── cmd/server/main.go
├── internal/
│   ├── app/           # Bootstrap, DI
│   ├── config/        # Configuration
│   ├── domain/        # Entities, contracts (ZERO external deps)
│   ├── usecase/       # Business logic
│   ├── repository/    # Data access
│   └── transport/     # HTTP handlers
├── openspec/          # Spec-driven development
│   ├── project.yaml
│   ├── specs/
│   ├── changes/
│   └── tasks/
├── database/migrations/
├── build/Dockerfile
├── CLAUDE.md
├── Makefile
└── .golangci.yml
```

## Standards Enforced

### Naming
- Variables: `cfg`, `repo`, `srv` (NOT `applicationConfiguration`)
- Constructors: `New()` public, `new*()` private
- Structs: private by default

### Error Handling
```go
// ✅ return fmt.Errorf("query user %s: %w", id, err)
// ❌ return fmt.Errorf("Failed to query: %w", err)
```

### Architecture
```
Transport → UseCase → Domain ← Repository ← Infrastructure
```
- Domain: ZERO external deps, NO struct tags
- Interfaces: defined at consumer side
- Repository: private models, mappers

## How It Works (v2.0)

1. **Specs First**: Create specs in `openspec/specs/`
2. **LLM Reads Templates**: Uses `internal/templates/` as reference patterns
3. **LLM Generates Code**: Writes code adapted to your project context
4. **Track Progress**: Manages tasks in `openspec/changes/` and `openspec/tasks/`

## Migration from v1.x

v1.x used template-based file generation (`go-ent init`). v2.0 uses:

- **MCP server** instead of CLI
- **Spec-driven development** instead of template copying
- **LLM code generation** instead of string replacement

See [MIGRATION_PLAN_GOENT_V3.md](docs/MIGRATION_PLAN_GOENT_V3.md) for the v3.0 multi-agent architecture migration plan.

## Contributing

1. Fork the repository
2. Create feature branch
3. Make changes following enterprise standards
4. Submit PR

## License

MIT

## References

- [CLI Examples](docs/CLI_EXAMPLES.md) - Comprehensive CLI usage guide
- [MCP Specification](https://modelcontextprotocol.io)
- [Official Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk)
- [Development Guide](docs/DEVELOPMENT.md)
- [Migration Plan v3.0](docs/MIGRATION_PLAN_GOENT_V3.md)
