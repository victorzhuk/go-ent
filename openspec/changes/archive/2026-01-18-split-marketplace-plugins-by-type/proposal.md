# Proposal: Split Marketplace Plugins by Type

## Why

The current `go-ent` plugin is monolithic, containing agents, skills, commands, and hooks. This creates several problems:
- Users must install everything even if they only need skills
- Independent versioning is impossible (can't update skills without agents)
- Marketplace discoverability is poor (one large package vs focused ones)
- Dependency tracking is unclear (which agents need which skills?)

Splitting by type enables fine-grained installation and better marketplace organization.

## What Changes

Split `go-ent` into focused packages:

**Before:**
```
plugins/sources/
└── go-ent/               # Monolithic
    ├── agents/
    ├── skills/
    ├── commands/
    └── hooks/
```

**After:**
```
plugins/packages/
├── agents/               # agents@go-ent
│   ├── plugin.yaml
│   ├── agents/
│   └── dependencies: [skills@go-ent]
├── skills/               # skills@go-ent
│   ├── plugin.yaml
│   └── skills/
├── commands/             # commands@go-ent
│   ├── plugin.yaml
│   ├── commands/
│   └── dependencies: [agents@go-ent]
├── hooks/                # hooks@go-ent
│   ├── plugin.yaml
│   └── hooks/
└── go-ent/               # go-ent (meta-package)
    ├── plugin.yaml
    └── dependencies: [agents@go-ent, skills@go-ent, commands@go-ent, hooks@go-ent]
```

**Marketplace structure:**
- `agents@go-ent` - Agent definitions only
- `skills@go-ent` - Skill definitions only
- `commands@go-ent` - Command definitions only
- `hooks@go-ent` - Hook definitions only
- `go-ent` - Meta-package (installs all)

**Key features:**
1. **Plugin dependencies** - `agents@go-ent` depends on `skills@go-ent`
2. **Cross-plugin references** - Agents reference skills by fully qualified name
3. **Independent versioning** - Update skills without touching agents
4. **Selective installation** - `/plugin install skills@go-ent` only

**BREAKING**: Existing users with `go-ent` installed need migration to `go-ent@latest` (meta-package)

## Impact

**Affected specs:**
- `marketplace` - Plugin packaging and dependencies

**Affected code:**
- `internal/plugin/manifest.go` - Add `dependencies` field
- `internal/plugin/manager.go` - Dependency resolution
- `internal/marketplace/install.go` - Install dependencies recursively
- `internal/marketplace/resolve.go` - **NEW** - Dependency resolver
- `plugins/packages/*/plugin.yaml` - **NEW** - Split plugin manifests

**Breaking changes:**
- **YES** - Marketplace plugin IDs change
- Users must reinstall plugins
- Migration: `/plugin uninstall go-ent && /plugin install go-ent@latest`

## Dependencies

**Requires:**
- `reorganize-plugin-source-layout` (provides structure)
- `upgrade-template-engine` (enables better cross-references)

**Blocks:** None (end of dependency chain)

## Success Criteria

- [ ] 5 plugin packages created (`agents@go-ent`, `skills@go-ent`, `commands@go-ent`, `hooks@go-ent`, `go-ent`)
- [ ] Plugin dependency system implemented
- [ ] Dependency resolution works (recursive install)
- [ ] Cross-plugin skill references work
- [ ] Independent versioning possible
- [ ] Migration tooling tested
- [ ] All tests pass
- [ ] Marketplace can serve split packages
