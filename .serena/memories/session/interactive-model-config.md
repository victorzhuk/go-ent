# Interactive Model Config Change

## Goal

The user wants to consolidate model configuration in go-ent by:
1. Eliminating fragmented config locations (`ent.yaml`, `.go-ent/models.yaml`, XDG global config)
2. Moving to **per-runtime configuration** in tool-native locations (`.claude/ent.yaml`, `.opencode/ent.yaml`)
3. Using Claude short aliases (`sonnet`, `opus`, `haiku`) that auto-resolve to latest
4. Discovering OpenCode models via `opencode models` CLI with 24h cache
5. Creating all required OpenSpec change artifacts following the project's conventions

## Instructions

- **No migration/legacy support** - this is pre-release, clean slate
- **No XDG global config fallback** - only per-runtime configs
- **No `ent model` CLI commands** - users edit YAML directly
- **Configs are optional** - defaults work out of the box
- Follow OpenSpec conventions from `openspec/project.md` and archived changes

## Discoveries

1. **Claude Code model naming**: Uses short aliases (`sonnet`, `opus`, `haiku`) that auto-resolve to latest versions - no discovery needed
2. **OpenCode model naming**: Uses provider-prefixed IDs (`opencode/glm-4.7`, `kimi-for-coding/k2p5`)
3. **OpenCode discovery**: `opencode models` CLI lists available models, can cache to `~/.cache/go-ent/opencode-models.json`
4. **OpenSpec structure** (from `openspec/changes/archive/2026-02-06-refactor-core-architecture/`):
   - `proposal.md`: Why, What Changes, Capabilities, Impact, Success Criteria
   - `tasks.md`: Checklist with numbered sections and `- [ ]` items
   - `design.md`: Context, Goals/Non-Goals, Decisions with rationale, Risks/Trade-offs
   - `specs/*/spec.md`: Requirements with Given/When/Then scenarios, Level (MUST/SHOULD)
5. **Current architecture has two separate resolution paths** that don't talk to each other:
   - `ent init` uses `config.ModelResolver` reading `.go-ent/models.yaml`
   - `ent agent generate` uses `generator.ResolveModel()` reading `ent.yaml`

## Accomplished

**Completed:**
- Reviewed original proposal and identified gaps/issues
- Researched Claude Code and OpenCode model conventions
- Rewrote `proposal.md` to follow OpenSpec format
- Cleaned up proposal.md to remove duplicate Impact section
- Created `tasks.md` with 6 sections, 38 tasks
- Created `design.md` with 5 decisions and rationale
- Created `specs/configuration/spec.md` with 5 requirements

**Ready for Approval:** All OpenSpec artifacts complete, waiting for user approval to begin implementation.

## Implementation Complete ✅

All 38 tasks completed and verified:
- Build: ✅ `make build` passes
- Tests: ✅ `internal/config/...` all pass
- Legacy removed: ✅ No genconfig, model_*.go, root ent.yaml
- New files: ✅ runtime_config.go, opencode_discovery.go with tests

## Relevant files / directories

**Change directory:**
- `openspec/changes/interactive-model-config/.openspec.yaml` - minimal (schema: spec-driven, created: 2026-02-24)
- `openspec/changes/interactive-model-config/proposal.md` - **COMPLETE**
- `openspec/changes/interactive-model-config/tasks.md` - **COMPLETE** - 38 tasks in 6 sections
- `openspec/changes/interactive-model-config/design.md` - **COMPLETE** - 5 decisions documented
- `openspec/changes/interactive-model-config/specs/configuration/spec.md` - **COMPLETE** - 5 requirements

**Code to be modified/deleted:**
- `internal/genconfig/config.go` - DELETE (entire package)
- `internal/config/model_config.go` - DELETE
- `internal/config/model_defaults.go` - DELETE
- `internal/config/resolver.go` - DELETE
- `internal/cli/model.go` - DELETE (model commands)
- `internal/generator/generator.go` - MODIFY (use RuntimeConfig)
- `internal/cli/agent/generate.go` - MODIFY
- `internal/cli/skill/generate.go` - MODIFY
- `internal/cli/init.go` - MODIFY
- `ent.yaml` (root) - DELETE

**Code to be created:**
- `internal/config/runtime_config.go` - NEW (~80 lines)
- `internal/config/opencode_discovery.go` - NEW (~60 lines)
