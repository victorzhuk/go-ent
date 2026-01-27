# Proposal: Restore Minimal `init` Command

## Metadata
- **Change ID:** `restore-init-command`
- **Status:** complete
- **Type:** Feature Restoration
- **Priority:** High
- **Affects Specs:** CLI, Plugin System
- **Depends On:** `simplify-01-delete-unused` (completed)

## Problem

The `init` command was deleted in Phase 1 of the simplification effort (`simplify-01-delete-unused`) because it was over-engineered at 1500+ LOC. However, the `init` command is essential functionality:

1. **External projects** cannot scaffold go-ent agents into their Claude Code or OpenCode configuration
2. **Self-hosted development** loses the ability to regenerate agent configurations
3. **Documentation is stale** - SETUP_GUIDE.md references commands that no longer exist

The original implementation was over-engineered with:
- Complex dependency resolution (~300 LOC)
- Multiple update modes
- Excessive abstractions in `internal/toolinit/`

## Proposed Solution

Restore `init` command with minimal complexity (~300-400 LOC):

### Command Interface

```
ent init --help

Scaffold agents, commands, and skills for agentic tools.

Usage:
  ent init [flags]

Flags:
  --tool string     Target tools: claude, opencode, or comma-separated list (required)
  --prefix string   Agent name prefix (default: "ent")
  --force           Overwrite existing files
  --dry-run         Preview changes without writing

Examples:
  ent init --tool=claude
  ent init --tool=opencode
  ent init --tool=claude,opencode
  ent init --tool=claude --force
  ent init --tool=claude --dry-run
```

### Rename Binary

Rename `cmd/go-ent` → `cmd/ent` for cleaner CLI:
- Before: `go-ent init --tool=claude`
- After: `ent init --tool=claude`

### Output Structure

**Claude Code** (`.claude/`):
```
.claude/
├── agents/{prefix}/       # coder.md, planner.md, etc.
├── commands/{prefix}/     # plan.md, task.md, etc.
└── skills/{prefix}/       # core/, go/ subdirectories
```

**OpenCode** (`.opencode/`):
```
.opencode/
├── agent/                 # Note: singular, no prefix subdirectory
├── commands/{prefix}/
└── skills/{prefix}/
```

### Architecture

```
internal/cli/init.go (~300-400 LOC)
├── loadAgents()     - Parse plugins/go-ent/agents/meta/*.yaml
├── loadPrompts()    - Load plugins/go-ent/agents/prompts/agents/*.md
├── loadShared()     - Concatenate plugins/go-ent/agents/prompts/shared/*.md
├── loadTemplate()   - Load claude.yaml.tmpl or opencode.yaml.tmpl
├── renderAgent()    - Render frontmatter + shared + prompt
├── copyCommands()   - Copy plugins/go-ent/commands/*.md
└── copySkills()     - Copy plugins/go-ent/skills/**/*.md
```

Resources loaded from existing embedded `PluginFS` (no new dependencies).

## Impact

- **Breaking Changes:** Binary renamed `go-ent` → `ent`
- **API Changes:** None (MCP tools unchanged)
- **Migration Required:** Update scripts/aliases using `go-ent` binary name
- **Testing Required:** Yes
  - Unit tests for path generation
  - Integration test with temp directory
  - Both Claude Code and OpenCode output verification

## Risks

- **Low Risk:** Simple implementation
- **Mitigation:** `--dry-run` flag for preview
- **Rollback:** Git revert straightforward

## What's NOT Included (Future Enhancements)

- `--agents` flag to filter specific agents
- `--include-deps` for automatic dependency resolution
- `--update` mode for updating existing config
- `--model` override for model mapping

## Success Criteria

- [x] `ent init --tool=claude` scaffolds valid Claude Code config
- [x] `ent init --tool=opencode` scaffolds valid OpenCode config
- [x] `ent init --tool=claude,opencode` scaffolds both
- [x] `--dry-run` shows preview without writing
- [x] `--force` overwrites existing files
- [x] `make build` succeeds
- [x] `make test` passes
- [x] `make lint` clean
