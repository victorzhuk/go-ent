# Design: Interactive Model Configuration

## Context

The current configuration system is fragmented across multiple locations:
- Root `ent.yaml` for generation config
- `.go-ent/models.yaml` for model aliases
- Global XDG config for defaults

This creates confusion and doesn't align with how different AI tools manage their own configurations. Claude and OpenCode have different model discovery mechanisms and naming conventions.

## Goals / Non-Goals

**Goals:**
- Simplify configuration to per-runtime files in tool-native paths
- Support Claude's short aliases (sonnet, opus, haiku) with auto-resolution
- Support OpenCode's `opencode models` CLI with caching
- Provide sensible defaults that work without any configuration
- Remove all legacy configuration code and files

**Non-Goals:**
- Migration from old config format (pre-release, clean break)
- Backward compatibility with existing configs
- Shared configuration between runtimes
- Configuration validation beyond basic parsing

## Decisions

### Decision 1: Per-Runtime Config Files

**Decision:** Each runtime gets its own config file in its native directory (`.claude/ent.yaml` or `.opencode/ent.yaml`).

**Rationale:** 
- Tools already have their own conventions and directory structures
- Avoids config conflicts between tools with different capabilities
- Users can configure each tool independently
- Simpler mental model: one tool = one config location

**Alternative Considered:** Single config file with runtime sections. Rejected because it creates coupling between tools and doesn't match user mental models of "Claude config lives in .claude/".

### Decision 2: No Migration Path

**Decision:** Delete all legacy config without migration support.

**Rationale:**
- Project is pre-release (v0.x)
- Breaking changes are acceptable
- Migration code adds complexity for little benefit
- Clean slate allows better architecture

**Alternative Considered:** Automatic migration script. Rejected due to maintenance burden and pre-release status.

### Decision 3: OpenCode Discovery with 24h Cache

**Decision:** Run `opencode models` CLI and cache results for 24 hours at `~/.cache/go-ent/opencode-models.json`.

**Rationale:**
- OpenCode model list changes rarely
- CLI invocation has latency (~100-500ms)
- Cache provides fast subsequent access
- 24h balance between freshness and performance

**Alternative Considered:** No caching, always query CLI. Rejected due to latency impact on every generation.

### Decision 4: Claude Short Aliases Without Discovery

**Decision:** Claude config uses short aliases (sonnet, opus, haiku) that map to known model IDs internally.

**Rationale:**
- Claude model IDs are stable and well-known
- No CLI to query available models
- Short aliases are more user-friendly
- Internal resolution is trivial

**Alternative Considered:** Require full model IDs. Rejected because short aliases match Claude Code conventions and are more ergonomic.

### Decision 5: Optional Configuration with Defaults

**Decision:** All configuration is optional; defaults work without any config file.

**Rationale:**
- Reduces friction for new users
- Most users only need default models
- Config only needed for customization
- Follows "convention over configuration" principle

**Alternative Considered:** Require config file. Rejected because it creates unnecessary setup step.

## Risks / Trade-offs

**Risk: Cache staleness**
- OpenCode models could change within 24h window
- Mitigation: Users can delete cache file or wait for expiry

**Risk: Config file location confusion**
- Users might look for config in old locations
- Mitigation: Clear error messages and documentation

**Trade-off: No shared config**
- Duplication if user wants same config for both tools
- Acceptable: Tools have different capabilities, configs naturally diverge

**Trade-off: Hardcoded Claude model mappings**
- New Claude models require code update
- Acceptable: Model releases are infrequent, update is simple
