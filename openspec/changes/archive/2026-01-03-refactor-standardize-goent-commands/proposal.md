# Standardize Command Naming to goent:* Prefix

**ID:** refactor-standardize-goent-commands
**Status:** COMPLETED
**Type:** Refactor
**Priority:** HIGH
**Complexity:** MEDIUM

## Overview

Standardize all plugin commands to use the `goent:*` prefix pattern, removing naming inconsistencies and improving discoverability. Currently, the plugin has 20 commands split across modern `goent:*` (8), legacy non-prefixed utilities (3), planning workflow commands (6), and incomplete stubs (3).

## Problem Statement

The go-ent plugin has evolved organically, creating two naming patterns:
- **Modern** `goent:*` commands (apply, archive, gen, loop, loop-cancel, registry, status, tdd)
- **Legacy** non-prefixed commands (init, scaffold, lint, plan, plan-full, clarify, research, decompose, analyze, checklist, review, test)

This inconsistency creates:
1. **UX confusion** - Users don't know which pattern to use
2. **Discoverability issues** - Related commands don't appear grouped
3. **Maintenance overhead** - Two patterns to maintain and document
4. **Documentation fragmentation** - Command references scattered across naming styles

## Requirements

### Functional Requirements

- **REQ-1:** All commands must use `goent:*` prefix for namespace consistency
- **REQ-2:** Delete obsolete/incomplete stub commands (checklist, review, test)
- **REQ-3:** Rename utility commands (init, scaffold, lint) to `goent:*` pattern
- **REQ-4:** Consolidate planning workflow commands under `goent:*` namespace
- **REQ-5:** Merge `plan.md` logic into `plan-full.md` as primary `goent:plan`
- **REQ-6:** Update all internal command orchestration calls to use new names
- **REQ-7:** Preserve all existing functionality during migration
- **REQ-8:** Update all documentation references to new command names

### Non-Functional Requirements

- **NFR-1:** Zero breaking changes to `goent:*` commands (already correct)
- **NFR-2:** All commands must execute without errors after refactor
- **NFR-3:** Planning workflow orchestration must remain functional
- **NFR-4:** Documentation must be updated before release

## Acceptance Criteria

### Phase 1: Cleanup
- [ ] **AC-1.1:** Files `checklist.md`, `review.md`, `test.md` are deleted
- [ ] **AC-1.2:** Command count reduced from 20 to 17 files
- [ ] **AC-1.3:** No references to deleted commands remain in codebase

### Phase 2: Utility Renaming
- [ ] **AC-2.1:** `init.md` renamed to `goent:init.md` with updated frontmatter
- [ ] **AC-2.2:** `scaffold.md` renamed to `goent:scaffold.md` with updated frontmatter
- [ ] **AC-2.3:** `lint.md` renamed to `goent:lint.md` with updated frontmatter
- [ ] **AC-2.4:** All three commands execute successfully with new names

### Phase 3: Planning Workflow Consolidation
- [ ] **AC-3.1:** `clarify.md` renamed to `goent:clarify.md`
- [ ] **AC-3.2:** `research.md` renamed to `goent:research.md`
- [ ] **AC-3.3:** `decompose.md` renamed to `goent:decompose.md`
- [ ] **AC-3.4:** `analyze.md` renamed to `goent:analyze.md`
- [ ] **AC-3.5:** `plan.md` deleted (functionality merged into plan-full)
- [ ] **AC-3.6:** `plan-full.md` renamed to `goent:plan.md`
- [ ] **AC-3.7:** Internal orchestration calls in `goent:plan.md` updated to use `goent:*` prefix
- [ ] **AC-3.8:** Planning workflow executes end-to-end without errors

### Phase 4: Documentation Updates
- [ ] **AC-4.1:** README.md updated with all new command names
- [ ] **AC-4.2:** AGENTS.md updated with all new command names
- [ ] **AC-4.3:** TRANSFORMATION.md updated with all new command names
- [ ] **AC-4.4:** All agent prompts updated to reference new command names
- [ ] **AC-4.5:** Plugin hooks verified for command name dependencies

### Final Validation
- [ ] **AC-5.1:** All 15 commands listed in `/goent` namespace
- [ ] **AC-5.2:** 100% `goent:*` prefix standardization achieved
- [ ] **AC-5.3:** Zero non-prefixed commands remain
- [ ] **AC-5.4:** Full workflow test passes: init → plan → apply → archive
- [ ] **AC-5.5:** No broken references in documentation or code

## Current State Analysis

### Command Inventory (20 total)

**goent:* prefixed (8 - KEEP AS IS):**
- goent:apply.md
- goent:archive.md
- goent:gen.md
- goent:loop.md
- goent:loop-cancel.md
- goent:registry.md
- goent:status.md
- goent:tdd.md

**Utilities (3 - RENAME):**
- init.md → goent:init.md
- scaffold.md → goent:scaffold.md
- lint.md → goent:lint.md

**Planning Workflow (6 - CONSOLIDATE):**
- plan.md → DELETE (merge into plan-full)
- plan-full.md → goent:plan.md
- clarify.md → goent:clarify.md
- research.md → goent:research.md
- decompose.md → goent:decompose.md
- analyze.md → goent:analyze.md

**Stubs (3 - DELETE):**
- checklist.md → DELETE
- review.md → DELETE
- test.md → DELETE

## Target State (15 commands)

All commands under `goent:*` namespace:
```
goent:analyze
goent:apply
goent:archive
goent:clarify
goent:decompose
goent:gen
goent:init
goent:lint
goent:loop
goent:loop-cancel
goent:plan
goent:registry
goent:research
goent:scaffold
goent:status
goent:tdd
```

## Impact Analysis

### Breaking Changes
- **User-facing:** Command invocations change from `/init` to `/go-ent:init`
- **Orchestration:** Internal calls in `plan-full.md` must update to new names
- **Documentation:** All references need update

### Risk Mitigation
- **Low Risk:** `goent:*` commands already correct (8/20 unchanged)
- **Medium Risk:** Renaming utilities requires doc updates
- **High Risk:** Planning orchestration - must update internal calls correctly

### Dependencies
- Planning workflow orchestration depends on sub-commands
- Documentation references scattered across multiple files
- Potential agent prompt references to old command names

## Open Questions

1. **Migration Strategy:** Should we provide deprecation aliases or hard cutover?
   - **Option A:** Version 1.1.0 adds goent:* versions, keeps legacy as aliases with warnings
   - **Option B:** Hard cutover in next version with breaking change notice
   - **Recommendation:** Option B (clean break, simpler maintenance)

2. **plan.md vs plan-full.md:** How to merge?
   - **Option A:** Keep both, plan as lightweight mode, plan-full as comprehensive
   - **Option B:** Delete plan.md, make plan-full.md the only `goent:plan`
   - **Recommendation:** Option B (reduce cognitive load, plan-full is superior)

3. **Sub-command Exposure:** Should planning phases be standalone?
   - **Current:** clarify, research, decompose, analyze are orchestrated by plan-full
   - **Option A:** Keep as separate commands (reusable independently)
   - **Option B:** Inline into goent:plan (simpler)
   - **Recommendation:** Option A (advanced users can call independently)

## Success Metrics

- **Quantitative:**
  - 100% commands use `goent:*` prefix
  - 25% reduction in command count (20 → 15)
  - Zero broken references in docs

- **Qualitative:**
  - Improved command discoverability
  - Consistent user experience
  - Simplified maintenance

## References

- Project conventions: `/openspec/AGENTS.md`
- Current commands: `/plugins/go-ent/commands/`
- Documentation: `/README.md`, `/TRANSFORMATION.md`
- Ultrathink analysis: (conversation context)
