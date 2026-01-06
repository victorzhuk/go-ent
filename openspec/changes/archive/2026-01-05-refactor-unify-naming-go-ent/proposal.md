# Refactor: Unify Naming to go-ent

## Overview

Standardize all naming to `go-ent` (with hyphen) across the entire codebase for consistency with repository and module names.

## Rationale

### Problem

The project has inconsistent naming between components:
- Repository: `go-ent` (correct)
- Go module: `github.com/victorzhuk/go-ent` (correct)
- Binary: `dist/go-ent` (inconsistent)
- Source directory: `cmd/go-ent/` (inconsistent)
- Plugin name: `goent` (inconsistent)
- MCP server: `goent-spec` (inconsistent)
- MCP tools: `go_ent_*` (inconsistent)
- Commands: `goent:*` (inconsistent)
- Agents: `goent:*` (inconsistent)

This creates cognitive overhead and confusion for users and developers.

### Solution

Rename all occurrences of `goent` to `go-ent` (or `go_ent` for MCP tools) to match the authoritative naming from the repository and Go module.

### Benefits

- **Consistency:** Single naming convention across all components
- **Discoverability:** Users find commands, tools, and docs more easily
- **Reduced Confusion:** No need to remember which context uses which variant
- **Professional:** Unified branding across the project

## Key Changes

### File System Changes
- Directory: `cmd/go-ent/` → `cmd/go-ent/`
- Binary: `dist/go-ent` → `dist/go-ent`
- Commands: 17 files `goent:*.md` → `go-ent:*.md`
- Agents: 7 files `goent:*.md` → `go-ent:*.md`

### Code Changes
- Import paths: `cmd/go-ent` → `cmd/go-ent` (44 Go files)
- MCP server name: `goent-spec` → `go-ent`
- MCP tools: `go_ent_*` → `go_ent_*` (24 tools)
- Version output: `"goent %s"` → `"go-ent %s"`
- Error messages referencing tool names

### Configuration Changes
- Makefile: Binary paths and version package
- .goreleaser.yaml: Project name, build IDs, binary name
- plugin.json: Plugin name and binary command
- marketplace.json: Plugin name
- .gitignore: Template path

### Documentation Changes
- CLAUDE.md: All command/agent examples
- README.md: Installation, commands, examples
- DEVELOPMENT.md: Tool permissions, setup
- plugins/go-ent/README.md: Full documentation
- openspec/AGENTS.md: Registry commands
- All command/agent frontmatter: allowed-tools

## Impact

- **Breaking Change:** v4.0.0 release required
- **Files Affected:** ~130 files
- **User Action Required:** Rebuild, restart Claude Code, use new command names
- **Migration Guide:** Required for users

## Dependencies

- None (self-contained refactoring)

## Risks

| Risk | Mitigation |
|------|------------|
| Import paths break | Test `go build ./...` before commit |
| Old binary cached | Document clean step in migration guide |
| MCP connection fails | Test plugin reload before final commit |
| Documentation drift | Automated grep checks for old references |
