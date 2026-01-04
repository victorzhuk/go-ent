# Change: Add MCP Template Generation and Validation Tools

## Why

The go-ent MCP server currently provides spec/task management tools but lacks critical functionality:

1. **Template Generation Gap**: 15 template files exist in `/templates/` but are never embedded or used. The `cli-build` spec requires Template Embedding and Template Processing, but implementation is missing.

2. **No Validation Tool**: AGENTS.md documents `openspec validate [change-id] --strict` extensively, but no MCP tool exists to validate specs, changes, or proposals.

3. **No Archive Tool**: The three-stage OpenSpec workflow requires archiving completed changes, but no tool automates this process.

4. **Hardcoded Plugin Path**: `plugin.json` contains an absolute path that only works on the original developer's machine.

Without these tools, go-ent cannot:
- Scaffold new Go projects from templates (primary use case)
- Validate spec proposals before implementation
- Complete the spec-driven workflow lifecycle

## What Changes

### New MCP Tools

| Tool | Purpose |
|------|---------|
| `goent_generate` | Generate Go project files from embedded templates |
| `goent_spec_validate` | Validate specs, changes, and proposals |
| `goent_spec_archive` | Archive completed changes and update specs |

### Template System

- Add `//go:embed` directive to bundle templates in binary
- Create template engine with variable substitution (`{{MODULE_PATH}}`, `{{PROJECT_NAME}}`)
- Support two project types: `standard` (web service) and `mcp` (MCP server)

### Plugin Configuration

- Fix `plugin.json` to use relative path or plugin resolution
- Update marketplace.json if needed

## Impact

- **Affected specs**: `cli-build` (implements existing requirements)
- **New spec**: `mcp-tools` (new capability for MCP tools)
- **Affected code**:
  - `cmd/goent/internal/tools/` - new tool handlers
  - `cmd/goent/templates/` - embed directive
  - `plugins/go-ent/.claude-plugin/plugin.json` - path fix
- **Breaking changes**: None (additive only)

## Success Criteria

1. `goent_generate` creates working Go project from templates
2. `goent_spec_validate` catches common spec errors
3. `goent_spec_archive` moves changes to archive and updates specs
4. Plugin works when installed via Claude Code marketplace
