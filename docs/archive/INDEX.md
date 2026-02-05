# go-ent Documentation Index

Central navigation hub for all go-ent documentation.

---

## Getting Started

### New Users

Start here if you're new to go-ent:

1. **[Quick Start](../README.md#quick-start)** - Install and use go-ent in 5 minutes
2. **[CLI Examples](./CLI_EXAMPLES.md)** - Common workflows and patterns
3. **[OpenSpec Workflow](./OPENSPEC_WORKFLOW.md)** - Understand spec-driven development
4. **[Configuration](./CONFIGURATION.md)** - Configure for your project

### Plugin Authors

Building plugins or extending go-ent:

1. **[Plugin Development](./PLUGIN_DEVELOPMENT.md)** - Create custom plugins
2. **[Skill Authoring](./SKILL-AUTHORING.md)** - Write v3 skill definitions
3. **[Agent System](./AGENTS_AND_SKILLS.md)** - v3 agent architecture
4. **[Claude Code Compatibility](./CLAUDE_CODE_COMPATIBILITY.md)** - Alignment guide
5. **[Migration Guide v3](./MIGRATION_V3.md)** - Migrate from v2 to v3

### Contributors

Contributing to go-ent itself:

1. **[Development Guide](./DEVELOPMENT.md)** - Self-hosted development
2. **[Contributing Guide](./CONTRIBUTING.md)** - How to contribute
3. **[Architecture](./ARCHITECTURE.md)** - System design
4. **[Changelog](../CHANGELOG.md)** - Version history

---

## Core Concepts

### Architecture & Design

- **[Architecture](./ARCHITECTURE.md)** - System overview, components, data flow
- **[Architecture Review](./ARCHITECTURE_REVIEW.md)** - Design analysis and recommendations
- **[Flow Diagrams](./FLOW.md)** - Visual workflow representations
- **[Hooks System](./HOOKS.md)** - Lifecycle event hooks and automation

### OpenSpec System

- **[OpenSpec Workflow](./OPENSPEC_WORKFLOW.md)** - Spec-driven development process
- **[Agent System](./AGENTS_AND_SKILLS.md)** - Multi-agent orchestration v3

### Skills & Agents

- **[Skill Authoring](./SKILL-AUTHORING.md)** - Complete skill v3 format guide
- **[Agents and Skills](./AGENTS_AND_SKILLS.md)** - Agent and skill architecture
- **[Claude Code Compatibility](./CLAUDE_CODE_COMPATIBILITY.md)** - Alignment guide
- **[Skill Quality Scoring](./skill-quality-scoring.md)** - Quality metrics
- **[Skill Lint CI/CD](./SKILL_LINT_CI.md)** - Automated validation

---

## Development

### Guides

- **[Development Guide](./DEVELOPMENT.md)** - Self-hosted development workflow
- **[CLI Examples](./CLI_EXAMPLES.md)** - Command usage patterns
- **[Migration Guide v3](./MIGRATION_V3.md)** - Migrate from v2 to v3

---

## Reference

### Command Line Interface

- **[CLI Reference](./CLI_REFERENCE.md)** - Complete CLI command reference
- **[CLI Examples](./CLI_EXAMPLES.md)** - Usage patterns and workflows
- **[Commands Reference](./COMMANDS_REFERENCE.md)** *(coming soon)* - Slash command reference

### Configuration

- **[Configuration Reference](./CONFIGURATION.md)** - All configuration options

### API Reference

- **[MCP API Reference](./MCP_API.md)** *(coming soon)* - MCP tool API documentation
- **[Go API Reference](https://pkg.go.dev/github.com/victorzhuk/go-ent)** - Go package documentation

---

## By Topic

### Agents

| Document | Description |
|----------|-------------|
| [Agent System](./AGENTS_AND_SKILLS.md) | Agent architecture and workflows *(coming soon)* |
| [Available Agents](../README.md#available-agents) | Agent list in README |
| [Development](./DEVELOPMENT.md) | Adding new agents |

### Skills

| Document | Description |
|----------|-------------|
| [Skill Authoring](./SKILL-AUTHORING.md) | Complete v3 format guide |
| [Skill Quality Scoring](./skill-quality-scoring.md) | Quality metrics and evaluation |
| [Available Skills](../README.md#skills-auto-activated) | Skill list in README |

### OpenSpec

| Document | Description |
|----------|-------------|
| [OpenSpec Workflow](./OPENSPEC_WORKFLOW.md) | Complete workflow guide |
| [CLI Examples](./CLI_EXAMPLES.md) | Spec management examples |

### Configuration

| Document | Description |
|----------|-------------|
| [Configuration Reference](./CONFIGURATION.md) | All config options |
| [CLI Reference](./CLI_REFERENCE.md#ent-config) | Config commands |
| [Development Guide](./DEVELOPMENT.md) | Self-hosted config |

---

## By Role

### As a Developer

**I want to...**

- **Start using go-ent** → [Quick Start](../README.md#quick-start)
- **Run common workflows** → [CLI Examples](./CLI_EXAMPLES.md)
- **Configure my project** → [Configuration Reference](./CONFIGURATION.md)
- **Understand the workflow** → [OpenSpec Workflow](./OPENSPEC_WORKFLOW.md)

### As a Plugin Author

**I want to...**

- **Create a skill** → [Skill Authoring](./SKILL-AUTHORING.md)
- **Build a plugin** → [Plugin Development](./PLUGIN_DEVELOPMENT.md) *(coming soon)*
- **Add an agent** → [Development Guide](./DEVELOPMENT.md#adding-new-agents)
- **Create MCP tools** → [MCP API Reference](./MCP_API.md) *(coming soon)*
- **Validate skills** → [Skill Quality Scoring](./skill-quality-scoring.md)

### As a Contributor

**I want to...**

- **Contribute code** → [Contributing Guide](./CONTRIBUTING.md) *(coming soon)*
- **Understand architecture** → [Architecture](./ARCHITECTURE.md)
- **Set up development** → [Development Guide](./DEVELOPMENT.md)
- **See changelog** → [Changelog](../CHANGELOG.md)

---

## Quick Links

### Most Common

- [CLI Examples](./CLI_EXAMPLES.md) - Usage patterns
- [Configuration Reference](./CONFIGURATION.md) - All config options
- [OpenSpec Workflow](./OPENSPEC_WORKFLOW.md) - Spec-driven development
- [Skill Authoring](./SKILL-AUTHORING.md) - Write skills

### Project Files

- [README](../README.md) - Project overview
- [CLAUDE.md](../CLAUDE.md) - Claude Code instructions
- [AGENTS.md](../AGENTS.md) - Agent system overview
- [LICENSE](../LICENSE) - MIT License

---

## Documentation Status

| Document | Status | Priority |
|----------|--------|----------|
| [CLI Reference](./CLI_REFERENCE.md) | ✅ Complete | Critical |
| [Configuration Reference](./CONFIGURATION.md) | ✅ Complete | Critical |
| [OpenSpec Workflow](./OPENSPEC_WORKFLOW.md) | ✅ Complete | High |
| [CLI Examples](./CLI_EXAMPLES.md) | ✅ Complete | High |
| [Skill Authoring](./SKILL-AUTHORING.md) | ✅ Complete | High |
| [Development Guide](./DEVELOPMENT.md) | ✅ Complete | High |
| [Architecture](./ARCHITECTURE.md) | 📝 Planned | High |
| [Contributing](./CONTRIBUTING.md) | 📝 Planned | High |
| [Changelog](../CHANGELOG.md) | 📝 Planned | High |
| [Agents & Skills](./AGENTS_AND_SKILLS.md) | 📝 Planned | Medium |
| [MCP API Reference](./MCP_API.md) | 📝 Planned | Medium |
| [Commands Reference](./COMMANDS_REFERENCE.md) | 📝 Planned | Medium |

---

## Contributing to Documentation

Found an issue or want to improve the docs?

1. Check [Contributing Guide](./CONTRIBUTING.md) *(coming soon)*
2. Open an issue on [GitHub](https://github.com/victorzhuk/go-ent/issues)
3. Submit a pull request

**Documentation standards:**
- Use markdown with GitHub-flavored extensions
- Include code examples with syntax highlighting
- Cross-reference related documents
- Keep examples up-to-date with current version
- Follow the existing structure and style

---

## Version

This documentation is for **go-ent v0.3.0**.

For older versions, see [Changelog](../CHANGELOG.md) *(coming soon)*.

---

**Last updated:** 2026-01-28
