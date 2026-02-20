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
- [Architecture](#architecture)
- [MCP Tools](#mcp-tools)
- [CLI Commands](#cli-commands)
- [Available Commands](#available-commands)
- [Available Agents](#available-agents)
- [Skills](#skills-auto-activated)
- [Skill Template System](#skill-template-system)
- [Building from Source](#building-from-source)
- [Project Structure](#project-structure)
- [Standards Enforced](#standards-enforced)
- [How It Works](#how-it-works)
- [Contributing](#contributing)
- [License](#license)
- [References](#references)

---

## Features

- 🏗️ **Clean Architecture** patterns and enforcement
- 📝 **SOLID principles** validation
- 🔍 **Automated code review** with enterprise standards
- 🧪 **Testing patterns** (unit, integration, benchmarks)
- 📋 **Spec-driven development** with `openspec` folder management
- 🤖 **MCP server** for spec/task management tools
- 🔧 **Hooks** for automatic formatting and safety
- 🤖 **Specialized agents** (coder, planner, researcher, reviewer, scout)
- ⚡ **Slash commands** for common workflows

## Quick Start

### 1. Build and Install

```bash
# Clone repository
git clone https://github.com/victorzhuk/go-ent.git
cd go-ent

# Build the MCP server binary
make build

# Install into Claude Code (writes to ~/.claude/)
ent init --tools=claude
```

Then restart Claude Code to load the plugin.

### 2. Using ent

**Via slash commands (in Claude Code):**

```
# Start a new change
/opsx:new Add user authentication

# Execute tasks from a change
/opsx:apply

# Archive when complete
/opsx:archive change-id
```

**Via CLI (standalone):**

```bash
# View configuration
ent config show

# List available agents
ent agent list

# Initialize OpenSpec in your project
ent spec init
```

## Architecture

### Project Layout

```
go-ent/
├── cmd/ent/                 # MCP server binary entry point
│   └── main.go             # stdio transport
├── internal/               # Implementation packages
│   ├── agent/              # Agent types and selection logic
│   ├── cli/                # CLI application
│   ├── config/             # Configuration management
│   ├── hooks/              # Hook registry and execution
│   ├── mcp/                # MCP server and tool registration
│   ├── openspec/           # OpenSpec client (change management)
│   ├── skill/              # Skill registry and parsing
│   ├── spec/               # Spec store (BoltDB-backed task registry)
│   ├── template/           # Skill template system
│   ├── workspace/          # Multi-project workspace support
│   └── xdg/               # XDG base directory paths
├── pkg/                    # Embeddable plugin assets
│   ├── agents/meta/        # 5 agent definitions (YAML)
│   ├── commands/           # Slash command definitions
│   ├── skills/             # ~80 skill definitions (SKILL.md)
│   ├── templates/          # 12 skill templates
│   ├── hooks/              # Hook configurations
│   └── schemas/            # Agent schema + examples
└── openspec/               # Self-hosted development specs
    ├── project.yaml
    ├── specs/
    ├── changes/
    └── archive/
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
└── archive/                # Completed changes
```

## MCP Tools

### Skill

| Tool | Description |
|------|-------------|
| `skill_list` | List all available skills |
| `skill_info` | Get detailed skill information |
| `skill_validate` | Validate a skill against quality rules |
| `skill_match` | Match skills to a task description |

### OpenSpec

| Tool | Description |
|------|-------------|
| `openspec_new_change` | Create a new change proposal |
| `openspec_archive` | Archive a completed change |
| `openspec_validate` | Validate openspec documents |
| `openspec_instructions` | Get workflow instructions |

### Registry

| Tool | Description |
|------|-------------|
| `registry_list_changes` | List all changes |
| `registry_list_tasks` | List all tasks across changes |
| `registry_get_change` | Get detailed change content |
| `registry_status` | Show registry summary |
| `registry_next_task` | Get next recommended task |
| `registry_deps` | Show task dependencies |
| `registry_mark_done` | Mark a task as completed |
| `registry_start_task` | Mark a task as in-progress |
| `registry_sync` | Sync tasks across proposals |

### Agent

| Tool | Description |
|------|-------------|
| `agent_list` | List all available agents |
| `agent_info` | Get detailed agent information |
| `agent_generate` | Generate an agent definition |

### Configuration

| Tool | Description |
|------|-------------|
| `config_show` | Show current configuration |
| `config_set` | Update a configuration value |

### Workspace

| Tool | Description |
|------|-------------|
| `workspace_specs` | List specs across workspace projects |
| `workspace_projects` | List workspace projects |

### Discovery

| Tool | Description |
|------|-------------|
| `tool_list` | List all available MCP tools |

### Generation

| Tool | Description |
|------|-------------|
| `generate` | Generate code from templates |

## CLI Commands

The `ent` binary runs as an MCP server by default (stdio transport) and also supports CLI mode.

### Initialization

```bash
# Install into Claude Code
ent init --tools=claude

# Install into OpenCode
ent init --tools=opencode

# Install into both
ent init --tools=claude,opencode

# Preview without writing files
ent init --tools=claude --dry-run

# Force overwrite existing files
ent init --tools=claude --force
```

### Configuration Management

```bash
# Show current configuration
ent config show
ent config show --format summary

# Modify configuration
ent config set <key> <value>
ent config set budget.daily 25
```

### Agent Management

```bash
# List all agents
ent agent list
ent agent list --detailed

# Get agent information
ent agent info <name>
ent agent info coder
```

### Skill Management

```bash
# List all skills
ent skill list

# Get skill information
ent skill info go-code

# Create new skill from template
ent skill new <name>              # Interactive mode
ent skill new go-payment \
  --template go-basic \
  --description "Payment processing"

# List available templates
ent skill list-templates
ent skill list-templates --category go

# Show template details
ent skill show-template go-complete

# Add custom template
ent skill add-template ./my-template
```

### Spec Management

```bash
# Initialize OpenSpec in a project
ent spec init [path]

# List specs or changes
ent spec list spec
ent spec list change

# Show specific spec or change
ent spec show spec api
ent spec show change add-authentication
```

### Global Flags

```bash
ent --config /path/to/config.yaml <command>
ent --verbose <command>
ent version
```

## Available Commands

> **Note:** These are **slash commands** for use within Claude Code, installed via `ent init`.

| Command | Description |
|---------|-------------|
| `/opsx:explore` | Think through ideas and investigate problems |
| `/opsx:new` | Start a new change with spec-driven workflow |
| `/opsx:continue` | Continue working on an existing change |
| `/opsx:apply` | Execute tasks from a change |
| `/opsx:ff` | Fast-forward all change artifacts |
| `/opsx:sync` | Sync delta specs from a change to main specs |
| `/opsx:archive` | Archive a completed change |
| `/opsx:bulk-archive` | Archive multiple completed changes at once |
| `/opsx:verify` | Verify implementation matches change artifacts |
| `/opsx:onboard` | Guided onboarding walkthrough |

## Available Agents

| Agent | Description |
|-------|-------------|
| `coder` | Code implementation agent. Writes, edits, and refactors code following project patterns. |
| `planner` | Task planning and decomposition agent. Breaks down tasks, creates implementation plans. |
| `researcher` | Deep research and analysis agent. Explores architecture, investigates codebases. |
| `reviewer` | Code review and quality analysis agent. Reviews for bugs, security, architecture adherence. |
| `scout` | Quick exploration and triage agent. Fast file search, code lookup, simple Q&A. |

## Skills (Auto-activated)

Skills activate automatically based on conversation context. No manual invocation needed.

| Category | Description |
|----------|-------------|
| `go` | Go development — Clean Architecture, patterns, idioms |
| `backend` | Backend services, APIs, databases, messaging |
| `core` | Core development workflows and best practices |
| `agent` | Agent orchestration, multi-agent patterns |
| `infra` | Infrastructure, DevOps, deployment |
| `lang` | Language-specific patterns (TypeScript, Python, etc.) |
| `qa` | Quality assurance, testing, code review |
| `ent` | go-ent specific skills and workflows |

## Skill Template System

go-ent provides a template-based system for creating new skills quickly and consistently.

### Built-in Templates

Located in `pkg/templates/skills/`:

- `go-basic`, `go-complete` — Go development patterns
- `typescript-basic` — TypeScript-specific guidance
- `testing` — TDD and testing best practices
- `database` — SQL, migrations, and data access
- `api-design` — REST, GraphQL, and API patterns
- `core-basic`, `arch` — Architecture and system design
- `debugging-basic` — Troubleshooting and debugging
- `security` — Authentication, authorization, and security
- `review` — Code review practices

**Custom Templates**: User-defined templates in `~/.go-ent/templates/skills/`

### Creating Skills from Templates

**Interactive mode** (recommended):
```bash
ent skill new my-skill
```

**Non-interactive mode**:
```bash
ent skill new go-payment \
  --template go-basic \
  --description "Payment processing patterns" \
  --category go \
  --tags "payment,api"
```

### Template Structure

Each template consists of two files:

**config.yaml** — Template metadata:
```yaml
name: go-complete
category: go
description: "Comprehensive Go development template"
version: "1.0.0"
prompts:
  - key: NAME
    prompt: "Skill name"
    required: true
  - key: DESCRIPTION
    prompt: "Skill description"
    required: true
```

**template.md** — Skill with placeholders:
```markdown
---
name: ${NAME}
description: "${DESCRIPTION}"
---

# ${NAME}

<role>
Expert Go developer focused on clean architecture.
</role>

<instructions>
## Pattern 1
...
</instructions>
```

Output path: `pkg/skills/<category>/<skill-name>/SKILL.md`

## Building from Source

```bash
git clone https://github.com/victorzhuk/go-ent.git
cd go-ent
make build
./bin/ent  # runs as MCP server on stdio
```

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build MCP server to `bin/ent` |
| `make init` | Build and run `ent init` |
| `make generate` | Generate agent output for configured tools |
| `make validate` | Validate generated files against tool specs |
| `make test` | Run tests with race detector and coverage |
| `make test-templates` | Test all skill templates |
| `make lint` | Run golangci-lint |
| `make fmt` | Format code with goimports |
| `make clean` | Remove build artifacts |
| `make validate-plugin` | Validate plugin JSON files |
| `make validate-templates` | Validate all skill templates |
| `make skill-validate` | Validate all skills with strict mode |
| `make skill-sync` | Sync skills from pkg/ to .claude/ |
| `make skill-quality` | Generate quality report for all skills |
| `make help` | Show all available targets |

### Development Requirements

- Go 1.24 or later
- make
- golangci-lint (for `make lint`)

## Project Structure

```
go-ent/
├── cmd/ent/                 # MCP server binary
│   └── main.go
├── internal/                # Implementation (all unexported)
│   ├── agent/               # Agent types and selection
│   ├── cli/                 # CLI application
│   ├── config/              # Configuration system
│   ├── hooks/               # Hook registry and execution
│   ├── mcp/                 # MCP server setup and tool handlers
│   ├── openspec/            # OpenSpec client
│   ├── skill/               # Skill registry and parsing
│   ├── spec/                # BoltDB-backed task registry
│   ├── template/            # Skill template system
│   ├── workspace/           # Multi-project workspace
│   └── xdg/                # XDG base directory support
├── pkg/                     # Plugin assets (embedded, installed by ent init)
│   ├── agents/meta/         # 5 agent definitions (YAML)
│   ├── commands/            # Slash command definitions
│   ├── skills/              # ~80 skill definitions (SKILL.md)
│   ├── templates/           # 12 skill templates
│   ├── hooks/               # Hook configurations
│   └── schemas/             # Agent schema + examples
├── openspec/                # Self-hosted development specs
│   ├── project.yaml
│   ├── specs/
│   ├── changes/
│   └── archive/
├── docs/                    # Documentation
├── assets/                  # Logo and branding
└── Makefile
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

## How It Works

1. **Specs First**: Create change proposals in `openspec/changes/`
2. **LLM Reads Context**: Uses `pkg/skills/` and `pkg/templates/` as reference patterns
3. **LLM Generates Code**: Writes code adapted to your project context
4. **Track Progress**: Manages tasks via MCP registry tools

## Contributing

1. Fork the repository
2. Create feature branch
3. Make changes following enterprise standards
4. Submit PR

## License

MIT

## References

- **[Documentation Index](docs/INDEX.md)** — Central navigation hub
- **[Workspaces](docs/WORKSPACES.md)** — Multi-project workspace support
- **[MCP Specification](https://modelcontextprotocol.io)** — Model Context Protocol
- **[Official Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk)**
- **[Fission-AI/OpenSpec](https://github.com/Fission-AI/OpenSpec)** — OpenSpec standard
