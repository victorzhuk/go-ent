# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

---

## [0.5.0] - 2026-02-27

### Changed
- Change license from MIT to Apache 2.0
- Consolidate model configuration into runtime-specific files
- Actualize and clean up all project documentation

### Build
- Improve build and test process in CI

---

## [0.4.0] - 2026-02-21

### Added
- Tool discovery system with progressive disclosure
  - `tool_find` - Search tools by semantic query using TF-IDF scoring
  - `tool_describe` - Get detailed tool metadata and JSON schema
  - `tool_load` - Dynamically activate tools into the active set
  - `tool_active` - List currently loaded tools
- New agent and runtime tools
  - `agent_execute` - Execute tasks with automatic agent selection
  - `skill_info` - Get detailed skill information
  - `runtime_list` - List available runtimes and their capabilities
  - `runtime_status` - Get current runtime status and configuration
- Tool registry with lazy loading architecture
- TF-IDF search implementation (stdlib only, no external dependencies)
- Comprehensive tool discovery documentation in `openspec/AGENTS.md`
- Migration guide for tool name changes at `docs/MIGRATION_TOOL_NAMES.md`
- Token reduction: 70-90% for typical workflows (2,385 → 200-500 tokens)
- Search accuracy: 100% top-3 accuracy on 25 diverse test queries
- Thread-safe concurrent access for all discovery operations

### Changed
- **BREAKING**: All MCP tool names simplified by removing `go_ent_` prefix
  - Spec tools: `go_ent_spec_*` → `spec_*` (9 tools)
  - Registry tools: `go_ent_registry_*` → `registry_*` (6 tools)
  - Workflow tools: `go_ent_workflow_*` → `workflow_*` (3 tools)
  - Loop tools: `go_ent_loop_*` → `loop_*` (4 tools)
  - Generation tools: `go_ent_generate*` → `generate*` (4 tools)
  - Agent tool: `go_ent_agent_execute` → `agent_execute` (1 tool)
- Tool loading now happens progressively based on agent needs
- Initial MCP context reduced from ~2,385 to ~147 tokens (meta tools only)
- Updated all agent instructions and command documentation with new tool names

### Fixed
- Context bloat for simple tasks that only need 2-3 tools
- Lack of tool discoverability for agents

### Migration
Tool names were shortened (e.g. `go_ent_skill_list` → `skill_list`).

---

## [0.3.0] - 2026-01-27

### Added
- **MCP Server Entry Point**: `ent` binary now serves as MCP server with CLI fallback
- **Execution Engine v2**: Complete rewrite of task execution system
  - Improved state management
  - Better error handling
  - Parallel task execution support
- **Constitutional AI Alignment**: Agent prompts aligned with Anthropic's Constitutional AI principles
- **Automated Releases**: GitHub Actions workflow for automated releases
- **Skill Validation Refactor**: Enhanced skill validation for v2 format

### Changed
- MCP server architecture for better Claude Code integration
- Agent prompt structure for improved reliability
- Skill loading system with progressive enhancement

### Fixed
- MCP tool registration in Claude Code
- Skill validation edge cases
- CLI command parsing

---

## [0.2.1] - 2026-01-22

### Added
- **Plugin System**: Modular architecture for extending functionality
- **Skill Quality Scoring**: Research-backed quality assessment (0-100)
- **Progressive Skill Loading**: Three-stage loading (metadata → core → extended)
- **Skill Linting Tool**: Automated validation with `ent skill validate`
- **Skill Templates**: Reusable patterns for common development tasks
- **Context-Aware Matching**: Skills matched based on conversation context
- **Background Agents**: Long-running agent processes
- **AST Operations**: AST-based code querying and refactoring

### Changed
- Skill format to v2 with XML sections
- Agent system to support delegation chains
- Command system to support workflow composition

### Deprecated
- Skill v1 format (use `ent skill migrate` to upgrade)

---

## [0.2.0] - 2025-12-15

### Added
- Agent execution system with automatic complexity-based selection
- Task registry with cross-change dependency tracking
- Workflow orchestration with wait points and approval gates
- Autonomous loop with self-correction
- Spec validation with strict mode
- Archive command for completed changes

### Changed
- OpenSpec structure refined with change proposals and deltas
- Project initialization includes conventions support
- Validation reports enhanced with line numbers and context

---

## [0.1.0] - 2025-12-01

### Added
- Initial release of go-ent MCP server
- OpenSpec document management (init, create, update, delete, list, show)
- Code generation from templates (standard and MCP project types)
- Component scaffolding from spec files
- Archetype system for project templates
- MCP server integration with Claude Code
- Plugin system for self-hosted development

---

[Unreleased]: https://github.com/victorzhuk/go-ent/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/victorzhuk/go-ent/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/victorzhuk/go-ent/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/victorzhuk/go-ent/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/victorzhuk/go-ent/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/victorzhuk/go-ent/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/victorzhuk/go-ent/releases/tag/v0.1.0
