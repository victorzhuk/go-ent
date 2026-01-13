# Go Ent - Project Overview

## Purpose

Go Ent is an enterprise Go development toolkit for Claude Code with Clean Architecture, SOLID principles, and spec-driven development via MCP. It provides:

- MCP server for spec/task management tools
- Specialized agents (lead, architect, planner, dev, tester, debug, reviewer)
- Code generation and scaffolding
- OpenSpec workflow for spec-driven development

## Tech Stack

- **Language**: Go 1.24+
- **MCP Protocol**: github.com/modelcontextprotocol/go-sdk v1.2.0
- **CLI**: github.com/spf13/cobra v1.10.2
- **Config**: github.com/spf13/pflag v1.0.9
- **YAML**: gopkg.in/yaml.v3
- **UUID**: github.com/google/uuid v1.6.0
- **Testing**: github.com/stretchr/testify v1.11.1

## Project Structure

```
go-ent/
├── cmd/go-ent/              # MCP server binary
│   └── main.go
├── internal/
│   ├── agent/              # Agent types and selection logic
│   ├── cli/                # CLI application
│   ├── config/             # Configuration management
│   ├── domain/             # Core domain types (ZERO external deps)
│   ├── execution/          # Agent execution engine
│   ├── generation/         # Code generation from specs
│   ├── mcp/               # MCP server and tools (25 tools)
│   ├── skill/              # Skill registry and parsing
│   ├── spec/               # Spec store, registry, workflow, archiver
│   ├── template/           # Template engine
│   ├── templates/          # Embedded reference templates
│   ├── plugin/             # Plugin manifest, manager, loader
│   ├── marketplace/        # MCP marketplace client
│   └── version/           # Version metadata
├── plugins/go-ent/         # Claude Code plugin
│   ├── agents/             # 7 agent definitions
│   ├── commands/           # 16 slash commands
│   ├── skills/             # 10 skill definitions
│   └── .claude-plugin/    # MCP configuration
├── openspec/               # Self-hosted development specs
│   ├── specs/              # Capability specs
│   ├── changes/            # Active changes with tasks.md
│   └── archive/            # Completed changes
├── docs/                  # Documentation
├── scripts/               # Build and utility scripts
└── Makefile              # Build targets
```

## Build Commands

```bash
make build          # Build MCP server to bin/ent
make test           # Run tests with race detector and coverage
make lint           # Run golangci-lint
make fmt            # Format code with goimports
make clean          # Remove build artifacts
make validate-plugin # Validate plugin JSON files
```

## Architecture

Clean Layered Architecture:
```
Transport → UseCase → Domain ← Repository ← Infrastructure
```

- **Domain**: ZERO external deps, NO struct tags, pure business logic
- **Interfaces**: Defined at consumer side, minimal
- **Repository**: Private models with tags, mappers to domain entities
- **MCP Tools**: 25 tools for spec, registry, workflow, generation, execution

## Self-Hosted Development

Project uses its own plugin system (dogfooding):

1. Build MCP server: `make build`
2. Restart Claude Code to load plugin
3. Use slash commands: `/go-ent:plan`, `/go-ent:apply`, `/go-ent:status`
4. Agents auto-activate based on task type

## Development Workflow

OpenSpec-driven:
1. `/go-ent:plan <description>` - Create change proposal
2. Registry tracks tasks across changes
3. `/go-ent:apply` - Execute next unblocked task
4. Agents specialize: architect, planner, dev, tester, reviewer
5. Skills auto-activate: go-code, go-arch, go-api, go-db, etc.
6. `/go-ent:archive <change-id>` - Archive completed changes

## Key Files

- `openspec/registry.yaml` - Task registry synced from tasks.md files
- `openspec/changes/{id}/tasks.md` - Task lists per change
- `openspec/specs/{id}/spec.md` - Capability specifications
- `AGENTS.md` - Build commands, code style, testing
- `CLAUDE.md` - Self-hosted development guide
- `Makefile` - Build targets
