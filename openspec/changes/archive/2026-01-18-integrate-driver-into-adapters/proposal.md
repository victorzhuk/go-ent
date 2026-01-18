# Proposal: Integrate Driver into Adapters

## Why

The "driver" (orchestrator) concept exists in `/prompts/agents/driver.md` but isn't integrated with the plugin system. This creates confusion and duplication - there are two separate prompt locations with overlapping functionality.

The driver/orchestrator pattern is inherently platform-specific:
- Claude Code has different agent delegation mechanisms than OpenCode
- Orchestration styles vary by platform capabilities
- Each platform needs tailored coordination prompts

## What Changes

1. **Remove legacy `/prompts/` directory** - Eliminate duplication
2. **Integrate driver into platform adapters** - Platform-specific orchestration
3. **Consolidate all prompts** into `plugins/sources/go-ent/agents/prompts/`
4. **Implement driver as adapter feature** - Each platform handles orchestration differently

**Before:**
```
prompts/                     # Legacy location
├── agents/driver.md
└── shared/tooling.md

plugins/go-ent/agents/
└── prompts/                 # Current location
    ├── agents/*.md
    └── shared/_*.md
```

**After:**
```
plugins/sources/go-ent/agents/prompts/  # Single source of truth
├── _base.md                             # Universal agent base
├── _driver.md                           # Orchestrator capabilities
├── shared/
│   ├── _tooling.md
│   ├── _conventions.md
│   └── _handoffs.md
└── agents/
    ├── architect.md
    ├── coder.md
    └── driver.md                        # Inherits _driver.md

plugins/platforms/claude/
└── driver.go                            # Claude-specific orchestration

plugins/platforms/opencode/
└── driver.go                            # OpenCode-specific orchestration
```

Key changes:
- Delete `/prompts/` directory
- Add `_base.md` and `_driver.md` shared prompts
- Implement driver logic in each platform adapter
- Consolidate tooling references

## Impact

**Affected specs:**
- `agent-system` - Agent orchestration and coordination

**Affected code:**
- `internal/toolinit/claude.go` - Add driver orchestration
- `internal/toolinit/opencode.go` - Add driver orchestration
- `plugins/sources/go-ent/agents/prompts/` - New prompt files
- `/prompts/` - **DELETED**

**Breaking changes:** None - internal only (no API changes)

## Dependencies

**Requires:** `reorganize-plugin-source-layout` (provides new directory structure)

**Blocks:** None (standalone enhancement)

## Success Criteria

- [ ] `/prompts/` directory removed
- [ ] All prompts in `plugins/sources/go-ent/agents/prompts/`
- [ ] `_base.md` and `_driver.md` created
- [ ] Platform adapters implement orchestration
- [ ] All tests pass
- [ ] Agent delegation works in both Claude and OpenCode
