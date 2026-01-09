# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Tool discovery system with progressive disclosure
  - `tool_find` - Search tools by semantic query using TF-IDF scoring
  - `tool_describe` - Get detailed tool metadata and JSON schema
  - `tool_load` - Dynamically activate tools into the active set
  - `tool_active` - List currently loaded tools
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
See [docs/MIGRATION_TOOL_NAMES.md](docs/MIGRATION_TOOL_NAMES.md) for complete migration guide.

**Quick migration:**
```bash
# Update tool references in scripts
sed -i 's/go_ent_spec_/spec_/g' scripts/*.sh
sed -i 's/go_ent_registry_/registry_/g' scripts/*.sh
sed -i 's/go_ent_workflow_/workflow_/g' scripts/*.sh
sed -i 's/go_ent_loop_/loop_/g' scripts/*.sh
sed -i 's/go_ent_generate/generate/g' scripts/*.sh
sed -i 's/go_ent_agent_execute/agent_execute/g' scripts/*.sh
```

**No backward compatibility** - Tool name changes are breaking and require updates.

---

## [0.2.0] - 2025-XX-XX

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

## [0.1.0] - 2025-XX-XX

### Added
- Initial release of go-ent MCP server
- OpenSpec document management (init, create, update, delete, list, show)
- Code generation from templates (standard and MCP project types)
- Component scaffolding from spec files
- Archetype system for project templates
- MCP server integration with Claude Code
- Plugin system for self-hosted development

---

[Unreleased]: https://github.com/victorzhuk/go-ent/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/victorzhuk/go-ent/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/victorzhuk/go-ent/releases/tag/v0.1.0
