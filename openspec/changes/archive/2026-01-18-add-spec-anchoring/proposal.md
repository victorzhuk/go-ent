# Proposal: Add Spec Anchoring and Evolution

## Overview

Implement spec-as-source workflow with anchoring modes, automatic spec evolution tracking, and code-to-spec synchronization to ensure specs remain the source of truth throughout development lifecycle.

## Rationale

### Problem

- Specs and code drift apart over time
- No automated tracking of spec changes from code modifications
- Agents make ad-hoc code changes without updating specs
- Difficult to see what changed between spec versions

### Solution

- **Spec Anchoring Mode**: Agents work in "anchored" mode where code changes automatically generate spec deltas
- **Evolution Tracking**: Version control for specs with diff/merge capabilities
- **Code-to-Spec Sync**: Analyze code changes and propose spec updates
- **Anchoring Workflow**: Three modes (free, review, strict) for different development phases

## Key Components

### Anchoring Modes

| Mode | Behavior | Use Case |
|------|----------|----------|
| **Free** | Code changes don't require spec updates | Exploration, prototyping |
| **Review** | Code changes suggest spec updates for review | Active development |
| **Strict** | Code changes blocked until spec updated | Production, compliance |

### Implementation Files

1. `internal/spec/anchor.go` - Anchoring mode enforcement
2. `internal/spec/evolution.go` - Spec versioning and diff
3. `internal/spec/sync.go` - Code-to-spec synchronization
4. `internal/spec/analyzer.go` - Code analysis for spec inference

### New MCP Tools

| Tool | Description |
|------|-------------|
| `spec_anchor_set` | Set anchoring mode (free/review/strict) |
| `spec_anchor_status` | Check current anchoring mode and violations |
| `spec_diff` | Show differences between spec versions |
| `spec_sync` | Analyze code and propose spec updates |

## Dependencies

- Requires: None (Phase 4 - Independent)
- Blocks: None
- Complements: add-tool-discovery, add-execution-engine

## Success Criteria

- [ ] Three anchoring modes implemented
- [ ] Spec versioning with git-like diff
- [ ] Code analysis detects API/schema changes
- [ ] Automatic spec delta generation from code
- [ ] CI integration for strict mode enforcement

## Impact

### Development Workflow

- **Spec-First**: Specs drive development (strict mode)
- **Spec-Tracked**: Code changes tracked in specs (review mode)
- **Spec-Free**: Rapid prototyping (free mode)

### Quality Assurance

- Specs stay synchronized with code
- Breaking changes automatically flagged
- Compliance requirements enforced via strict mode

## Architecture

```
Anchoring System
├── Mode Manager
│   ├── Free (no enforcement)
│   ├── Review (suggest updates)
│   └── Strict (block on violations)
├── Evolution Tracker
│   ├── Spec versioning
│   ├── Diff generation
│   └── Merge capabilities
└── Sync Engine
    ├── Code analyzer
    ├── Spec inferencer
    └── Delta generator
```

## Migration

**For Projects:**
1. Start in **Free** mode (default, no disruption)
2. Opt-in to **Review** mode for active development
3. Enable **Strict** mode for production code

**For CI/CD:**
1. Add `spec_anchor_status --mode=strict` to CI pipeline
2. Fail builds with unanchored changes
3. Require spec updates before code merge
