# Archive: add-openspec-config-and-aliases

**Archived**: 2026-01-30
**Reason**: Partially Done

## Original Intent
Add OpenSpec schema support by:
1. Creating OpenSpec schema definition (`openspec/schemas/go-ent/schema.yaml`)
2. Creating templates for proposals, specs, designs, and tasks
3. Adding `ent init` command to scaffold OpenSpec in a Go project
4. Providing `/opsx:*` command aliases for OpenSpec workflow compatibility

## Why Archived
This proposal is partially completed. Some deliverables exist, others are missing:

**Completed:**
- ✅ `/opsx:*` command aliases exist in multiple locations:
  - `.claude/commands/ent/opsx-new.md`, `opsx-apply.md`, `opsx-archive.md`
  - `.opencode/commands/opsx-*.md` (9 commands)
  - `pkg/commands/aliases/opsx-*.md` (3 commands)
- ✅ `ent init` command exists in `internal/cli/init.go` (873 lines)
  - Supports `--tool=claude` and `--tool=opencode` flags
  - Copies agents, commands, and skills to appropriate directories
  - Validates agent definitions

**Missing:**
- ❌ No `openspec/schemas/` directory
- ❌ No `schema.yaml` file defining go-ent artifacts
- ❌ No templates directory with proposal.md, spec.md, design.md, tasks.md templates

## Actual State
The proposal succeeded in creating the CLI tooling and command aliases, but the OpenSpec schema standardization part was not completed:
- Command aliases are implemented and functional
- `ent init` command exists and works for both Claude Code and OpenCode
- OpenSpec schema definition and templates remain unimplemented

## Files Preserved
- proposal.md
- This archive.md

## Notes for Future
The OpenSpec schema standardization component of this proposal could be revisited if needed. The current implementation uses custom `/ent:*` commands and agents without the formal OpenSpec schema template system. The command aliases provide basic workflow compatibility with OpenSpec, but full schema-based artifact generation was not implemented.

If schema support becomes a priority:
1. Create `openspec/schemas/go-ent/schema.yaml` with artifact definitions
2. Create template files in `openspec/schemas/go-ent/templates/`
3. Integrate with OpenSpec CLI's standard workflow (if external users require it)
