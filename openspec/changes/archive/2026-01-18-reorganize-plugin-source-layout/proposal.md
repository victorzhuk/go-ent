# Proposal: Reorganize Plugin Source Layout

## Why

After `refactor-agent-command-skill-system`, the plugin structure mixes source definitions with platform-specific templates. This creates confusion about what's platform-agnostic vs platform-specific, and build artifacts aren't clearly separated from sources.

Current issues:
- Source plugins in `plugins/go-ent/` mixed with platform templates
- Generated configs (`.claude/`, `.opencode/`) lack clear source location
- No dedicated output directory for compiled plugins
- Legacy `/prompts/` directory duplicates functionality

## What Changes

Restructure the plugins directory to separate concerns:

**From:**
```
plugins/
└── go-ent/                  # Mixed sources + templates
    ├── agents/
    ├── commands/
    ├── skills/
    └── hooks/
```

**To:**
```
plugins/
├── sources/                 # Platform-agnostic definitions
│   └── go-ent/
│       ├── agents/
│       ├── commands/
│       ├── skills/
│       └── hooks/
├── platforms/               # Platform-specific templates
│   ├── claude/
│   │   └── templates/
│   └── opencode/
│       └── templates/
dist/                        # Build output (gitignored)
├── claude/go-ent/
└── opencode/go-ent/
```

Key changes:
- Move `plugins/go-ent/` → `plugins/sources/go-ent/`
- Extract platform templates to `plugins/platforms/{claude,opencode}/`
- Configure `dist/` as build output directory
- Update all loader and adapter code paths
- Add `dist/` to `.gitignore`

## Impact

**Affected specs:**
- `plugin-system` - Plugin source organization

**Affected code:**
- `internal/toolinit/adapter.go` - Base adapter paths
- `internal/toolinit/claude.go` - Claude source paths
- `internal/toolinit/opencode.go` - OpenCode source paths
- `internal/plugin/loader.go` - Plugin scanning logic
- `.gitignore` - Ignore build artifacts

**Breaking changes:** None - internal reorganization only

## Dependencies

**Requires:** None (foundation change)

**Blocks:**
- `integrate-driver-into-adapters` - Needs new layout
- `upgrade-template-engine` - Needs separated sources/platforms

## Success Criteria

- [ ] All plugins sources in `plugins/sources/`
- [ ] Platform templates in `plugins/platforms/`
- [ ] Build output in `dist/` (gitignored)
- [ ] All tests pass
- [ ] Plugin generation works for both Claude and OpenCode
- [ ] No broken paths in generated configs
