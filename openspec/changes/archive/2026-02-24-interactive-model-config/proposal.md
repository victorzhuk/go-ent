## Why

Model configuration is fragmented across multiple locations with inconsistent behavior:

| Location | Type | Issue |
|----------|------|-------|
| `ent.yaml` (root) | Shared | Committed to git, cannot store user-specific preferences |
| `.go-ent/models.yaml` | User | Extra directory, not tool-native, conflicts with tool configs |
| `~/.config/go-ent/models.yaml` | Global | Overrides project needs, discovered late in load order |

**Root Cause**: Tools have different model naming conventions:
- Claude Code uses short aliases: `sonnet`, `opus`, `haiku` (auto-resolves to latest)
- OpenCode uses provider-prefixed IDs: `opencode/glm-4.7`, `kimi-for-coding/k2p5`

The current abstraction tries to be runtime-agnostic, forcing hardcoded mappings that get stale as models update and don't match either tool's native format.

## What Changes

Move to **per-runtime configuration** in tool-native locations:

```
.claude/ent.yaml      → Claude-specific (uses sonnet/opus/haiku aliases)
.opencode/ent.yaml    → OpenCode-specific (uses discovered model IDs)
```

**Claude Code**: Use short aliases (`sonnet`, `opus`, `haiku`) that auto-resolve to latest versions via Claude Code CLI.

**OpenCode**: Discover available models via `opencode models` CLI, cache results for 24h, fallback to sensible defaults if CLI unavailable.

**Both**: Config files are optional, user-created, and gitignored.

**Delete**: Remove `ent.yaml`, `.go-ent/` directory, and all global XDG config support entirely.

## Capabilities

### New Capabilities

- `runtime-config`: Per-tool configuration in `.claude/ent.yaml` and `.opencode/ent.yaml`
- `model-discovery`: Auto-discover OpenCode models with caching
- `claude-aliases`: Use short names (sonnet/opus/haiku) that resolve to latest versions

### Removed Capabilities

- `ent.yaml` (root) — replaced by per-runtime configs
- `.go-ent/models.yaml` — replaced by per-runtime configs  
- `~/.config/go-ent/models.yaml` — no longer needed
- `ent model set/list/reset` — users edit YAML directly

## Impact

### Code Impact

**Remove**:
- `internal/genconfig/` package (entire directory, ~90 lines)
- `internal/config/model_config.go` (LoadModelConfig, SaveModelConfig, ~125 lines)
- `internal/config/model_defaults.go` (DefaultModelConfig, ~28 lines)
- `internal/config/resolver.go` (ModelResolver, ~54 lines)
- `internal/cli/model.go` (model list/set/reset commands, ~192 lines)
- Root `ent.yaml` file

**Add**:
- `internal/config/runtime_config.go` (~80 lines) - RuntimeConfig interface, ClaudeConfig, OpenCodeConfig
- `internal/config/opencode_discovery.go` (~60 lines) - model discovery with caching

**Modify**:
- `internal/generator/generator.go` - use RuntimeConfig instead of genconfig.Config
- `internal/cli/agent/generate.go` - load runtime-specific config
- `internal/cli/skill/generate.go` - load runtime-specific config
- `internal/cli/init.go` - remove ent.yaml generation

### Architecture Impact

- Simplified config loading: single source per runtime (no merging logic)
- Tool-native configs: each tool has config in its directory (`.claude/`, `.opencode/`)
- Discovery pattern: OpenCode models discovered via CLI, not hardcoded

### User Impact

- No config required: Defaults work without any configuration
- Optional customization: Create `.claude/ent.yaml` or `.opencode/ent.yaml` to override
- Breaking change: Root `ent.yaml` deleted, `.go-ent/` ignored

### Migration Impact

- No migration: Pre-release clean slate
- Users create configs manually if customization desired

## Success Criteria

- `ent agent generate` works without any config files (uses defaults)
- Creating `.claude/ent.yaml` overrides Claude models
- Creating `.opencode/ent.yaml` overrides OpenCode models
- OpenCode model discovery caches for 24h
- Warning printed when `opencode` CLI not found
- No `ent.yaml` in project root
- No `.go-ent/` directory created
- No global XDG config loaded

## Artifacts

- `proposal.md` - This document
- `tasks.md` - Implementation checklist
- `design.md` - Technical design for runtime config system
- `specs/configuration/spec.md` - Delta spec for model configuration
