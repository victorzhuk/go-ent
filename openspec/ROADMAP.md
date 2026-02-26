# Go-Ent Development Roadmap

**Last Updated:** 2026-02-27
**Status:** All v0.4.0 proposals complete

## Current State

All planned proposals from the pre-v0.4.0 audit have been implemented and archived:

| Proposal | Outcome |
|----------|---------|
| cleanup-deprecated-features | Archived (complete) |
| simplify-skill-format | Archived (complete — v4 format shipped) |
| streamline-registry | Archived (complete) |
| refactor-align-agent-tools | Archived (complete) |

See `openspec/changes/archive/` for full implementation details.

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-01-30 | Archive 30+ proposals | Invalid, stale, completed, or out-of-scope |
| 2026-01-30 | Defer generate-agent-configs | Core functionality exists, enhancements optional |
| 2026-02-21 | Release v0.4.0 | All Phase 1-3 proposals complete |
| 2026-02-27 | Change license to Apache 2.0 | Patent protection for open-source distribution |

## Deferred

**Project-aware agent generation** (`generate-agent-configs`): Core `ent agent generate` works. Project-type detection and OpenSpec-based customization deferred — revisit if it becomes blocking.

## Next Steps

No active proposals. Open `openspec/changes/` to start a new change:

```bash
/opsx:new <description>
```
