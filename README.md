<<p align="center">
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

> **Major Update**: v2.0 replaces the CLI code generator with an MCP server for spec-driven development. See [TRANSFORMATION.md](TRANSFORMATION.md) for details.

## Features

- 🏗️ **Clean Architecture** patterns and enforcement
- 📝 **SOLID principles** validation
- 🔍 **Automated code review** with enterprise standards
- 🧪 **Testing patterns** (unit, integration, benchmarks)
- 📋 **Spec-driven development** with `.spec` folder management
- 🤖 **MCP server** for spec/task management tools
- 🔧 **Hooks** for automatic formatting and safety
- 🤖 **Specialized agents** (reviewer, planner, test-runner)
- ⚡ **Slash commands** for common workflows

## Quick Start

### 1. Install Plugin

```bash
/plugin install go-ent@go-ent
```

### 2. Initialize Spec-Driven Development

Use MCP tools to manage your project specs:

```
# Initialize .spec folder in your project
Call spec_init tool with path to your project

# Create a new spec
Call spec_create tool with type="spec", id="user-auth", content="..."

# List all specs
Call spec_list tool with type="specs"
```

The LLM (Claude Code) will generate code based on specs and templates, not copy-paste them.

## Architecture v2.0

### MCP Server

The `go-ent` binary is now an MCP server that provides tools for managing `.spec` folders:

```
go-ent/
├── cmd/go-ent/              # MCP server
│   ├── main.go             # stdio transport
│   └── internal/
│       ├── server/         # MCP setup
│       ├── tools/          # Tool handlers
│       └── spec/           # Domain logic
├── internal/
│   └── templates/          # Reference patterns (embedded)
└── plugins/go-ent/
    └── .claude-plugin/
        └── plugin.json     # MCP configuration
```

### `.spec` Folder Structure

```
project/.spec/
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

| Tool | Description |
|------|-------------|
| `spec_init` | Initialize .spec folder in project |
| `spec_list` | List specs, changes, or tasks |
| `spec_show` | Show detailed content |
| `spec_create` | Create new spec/change/task |
| `spec_update` | Update existing item |
| `spec_delete` | Delete item |

## Available Commands

| Command | Description |
|---------|-------------|
| `/go-ent:init <name>` | Initialize new project with specs |
| `/go-ent:scaffold <type> <name>` | Scaffold components |
| `/go-ent:review` | Review code for enterprise standards |
| `/go-ent:plan <feature>` | Create implementation plan |
| `/go-ent:test [pkg]` | Run tests and analyze failures |
| `/go-ent:lint` | Run linters |

## Available Agents

| Agent | Description |
|-------|-------------|
| `@code-reviewer` | Senior Go code reviewer |
| `@go-planner` | Architecture and feature planning |
| `@test-runner` | Test analysis and fixes |

## Skills (Auto-activated)

| Skill | Triggers |
|-------|----------|
| `go-review` | "review code", "check quality" |
| `go-patterns` | "create repository", "implement handler" |
| `go-testing` | "write tests", "add coverage" |
| `go-architecture` | "design service", "plan architecture" |

## Building from Source

```bash
# Clone repository
git clone https://github.com/victorzhuk/go-ent.git
cd go-ent

# Build MCP server
make build

# Binary will be in dist/go-ent
./dist/go-ent  # runs as MCP server on stdio
```

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build MCP server to `dist/go-ent` |
| `make test` | Run tests with race detector and coverage |
| `make lint` | Run golangci-lint |
| `make fmt` | Format code with goimports |
| `make clean` | Remove build artifacts |
| `make help` | Show all available targets |

### Development Requirements

- Go 1.23 or later
- make
- golangci-lint (for `make lint`)

## Project Structure

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
├── .spec/             # Spec-driven development
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

1. **Specs First**: Create specs in `.spec/specs/`
2. **LLM Reads Templates**: Uses `internal/templates/` as reference patterns
3. **LLM Generates Code**: Writes code adapted to your project context
4. **Track Progress**: Manages tasks in `.spec/changes/` and `.spec/tasks/`

## Migration from v1.x

v1.x used template-based file generation (`go-ent init`). v2.0 uses:

- **MCP server** instead of CLI
- **Spec-driven development** instead of template copying
- **LLM code generation** instead of string replacement

See [TRANSFORMATION.md](TRANSFORMATION.md) for detailed migration guide.

## Contributing

1. Fork the repository
2. Create feature branch
3. Make changes following enterprise standards
4. Submit PR

## License

MIT

## References

- [MCP Specification](https://modelcontextprotocol.io)
- [Official Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk)
- [Transformation Guide](TRANSFORMATION.md)
