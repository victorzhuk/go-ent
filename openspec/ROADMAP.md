# Go-Ent Development Roadmap

**Last Updated:** 2026-01-30
**Status:** Post-Audit, Ready for Implementation

## Overview

After comprehensive audit of 35+ proposals, 30+ invalid/stale proposals were archived. 5 core proposals remain active and ready for implementation. The project is in a clean state with a clear path forward.

**Total Active Effort:** ~36 hours across 5 proposals
**Archived Proposals:** 30+ (invalid, stale, completed, or out-of-scope)

## Active Proposals

### Phase 1: Technical Debt Cleanup (Immediate, ~4h)

**Priority: HIGH - Low Risk, Immediate Benefit**

| Proposal | Effort | Status | Risk |
|----------|--------|--------|------|
| cleanup-deprecated-features | ~4h | Ready | Low |

**Why First:**
- Removes dead code and test artifacts
- Immediate build speed improvement
- No risk to existing functionality
- Clears mental overhead for future work

**Impact:**
- Smaller codebase
- Faster builds
- Clearer project structure

---

### Phase 2: Core Functionality Improvements (High Impact, ~26h)

**Priority: HIGH - High Impact on Core System**

| Proposal | Effort | Status | Impact |
|----------|--------|--------|--------|
| simplify-skill-format | ~16h | Ready | Simplifies skill system, easier maintenance |
| streamline-registry | ~10h | Ready | Improves registry performance and usability |

**Why These Second:**
- High impact on developer experience
- Core system improvements
- Moderate effort for significant benefit
- Clear business value

**Impact:**
- Simpler skill format (v4)
- Easier to understand codebase
- Better performance
- Reduced maintenance burden

---

### Phase 3: Agent Tool Polish (Medium Value, ~6h)

**Priority: MEDIUM - Quality of Life**

| Proposal | Effort | Status | Impact |
|----------|--------|--------|--------|
| refactor-align-agent-tools | ~6h | Ready | Consistent agent tooling |

**Why Last of Core:**
- Quality improvement, not functional change
- Lower priority than cleanup and core improvements
- Can be done incrementally

**Impact:**
- Consistent agent tool usage
- Better alignment between agents and tools
- Improved maintainability

---

## Deferred Enhancements

### Project-Aware Agent Generation

| Proposal | Status | Reasoning |
|----------|--------|-----------|
| generate-agent-configs | Partially Done | Core functionality exists, enhancements optional |

**Current State:**
- ✅ `ent agent generate` command exists (`internal/cli/agent/generate.go`)
- ✅ Generator package exists (`internal/generator/`)
- ✅ Meta format source files in `pkg/agents/meta/`
- ✅ Static agent prompts in `pkg/agents/prompts/`
- ✅ Generation works: `ent agent generate` → `.claude/agents/ent/` and `.opencode/agents/ent/`
- ✅ Supports Claude Code and OpenCode formats
- ✅ Project configuration in `ent.yaml`

**Missing from Proposal:**
- ❌ Automatic project type detection (Go vs Web vs Mixed)
- ❌ OpenSpec specs-based agent customization
- ❌ `.go-ent/config.yaml` agent configuration section
- ❌ Project-specific agent selection

**Decision:**
The current implementation provides a solid foundation. The proposed enhancements (project-aware generation, OpenSpec integration) are valuable but not critical. The static + generate workflow works well for current needs.

**Recommendation:**
Archive or defer as "future enhancement." Focus on completing Phase 1-3 proposals first. Revisit if project-aware generation becomes a blocking requirement.

---

## Proposal Details

### 1. cleanup-deprecated-features (~4h)

**What:** Remove dead code, test artifacts, unused packages

**Files to Remove:**
- `marketplace/` - Unused directory
- Coverage files/outdated test artifacts
- MCP tool generation artifacts

**Success Criteria:**
- Clean repository with zero dead code
- Faster build times
- No impact on existing functionality

**Risk:** Low - only removes unused code

---

### 2. simplify-skill-format (~16h)

**What:** Migrate from v3 to v4 skill format, remove progressive loading

**Changes:**
- Simplify skill YAML structure
- Remove progressive loading complexity
- Update all skill files to v4 format
- Simplify parser logic

**Success Criteria:**
- All skills use v4 format
- Parser significantly simplified
- Documentation updated
- All tests pass

**Risk:** Medium - touches core skill system, but well-tested

---

### 3. streamline-registry (~10h)

**What:** Improve registry performance and usability

**Changes:**
- Remove `registry.yaml` if unnecessary
- Optimize registry operations
- Simplify registry sync logic
- Improve error messages

**Success Criteria:**
- Faster registry operations
- Clearer user experience
- Simplified code
- All tests pass

**Risk:** Low to Medium - optimization work, clear test coverage

---

### 4. refactor-align-agent-tools (~6h)

**What:** Ensure consistent agent tooling across all agents

**Changes:**
- Align agent tool configurations
- Standardize tool access patterns
- Update agent definitions
- Ensure tool consistency

**Success Criteria:**
- All agents have consistent tool access
- Tool configurations aligned
- Documentation updated
- All tests pass

**Risk:** Low - consistency improvement, no functional changes

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-01-30 | Archive 30+ proposals | Invalid, stale, completed, or out-of-scope |
| 2026-01-30 | Defer generate-agent-configs | Core functionality exists, enhancements are optional |
| 2026-01-30 | Prioritize cleanup | Low risk, immediate benefit, clears technical debt |
| 2026-01-30 | Sequence by value | Core improvements first, then polish |

---

## Next Steps

### Immediate (This Week)
1. ✅ Review and archive invalid proposals
2. 🔄 **Implement cleanup-deprecated-features**
3. 🔄 Test cleanup impact

### Short Term (Next 1-2 Weeks)
4. 🔄 **Implement simplify-skill-format**
5. 🔄 **Implement streamline-registry**
6. 🔄 Test all changes together

### Medium Term (Next Month)
7. 🔄 **Implement refactor-align-agent-tools**
8. 🔄 Decide on generate-agent-configs enhancements
9. 🔄 Complete any deferred improvements

---

## Execution Plan

### Week 1: Cleanup and Foundation
- [x] Audit all proposals
- [x] Archive invalid/stale proposals
- [ ] **Do:** Implement cleanup-deprecated-features
- [ ] **Do:** Run full test suite

### Week 2-3: Core Improvements
- [ ] **Do:** Implement simplify-skill-format
- [ ] **Do:** Implement streamline-registry
- [ ] **Do:** Integration testing
- [ ] **Do:** Documentation updates

### Week 4: Polish and Review
- [ ] **Do:** Implement refactor-align-agent-tools
- [ ] **Do:** Decide on generate-agent-configs
- [ ] **Do:** End-to-end testing
- [ ] **Do:** Release notes

---

## Archived Proposals

30+ proposals archived in `openspec/changes/archive/`:

### Categories

**Invalid Proposals:**
- Referenced non-existent packages or files
- Assumed incorrect system architecture
- Outdated assumptions about project structure

**Completed Proposals:**
- Work already done in earlier commits
- Features already implemented
- Outdated by newer implementations

**Stale Proposals:**
- Based on outdated assumptions
- Obsoleted by system changes
- No longer relevant to current priorities

**Out-of-Scope Proposals:**
- Not aligned with current project direction
- Features not needed for current use cases
- Overly complex for current needs

### Notable Examples

See individual `archive.md` files in `openspec/changes/archive/*/` for details:
- `2026-01-03-add-build-infrastructure` - Infrastructure already exists
- `2026-01-08-add-config-system` - System implemented differently
- Various refactoring proposals - Work completed or superseded

---

## Success Metrics

**Post-Implementation Goals:**
- Build time improvement: ≥10% (from cleanup)
- Skill system complexity: ↓50% (from simplify)
- Registry performance: ↑20% (from streamline)
- Code coverage: Maintain ≥80%
- Test pass rate: 100%

**Quality Gates:**
- All tests pass (including race detector)
- No new lint warnings
- Documentation complete
- No breaking changes to public APIs

---

## Risk Assessment

**Low Risk:**
- cleanup-deprecated-features - Only removes dead code
- refactor-align-agent-tools - Consistency improvement only

**Medium Risk:**
- simplify-skill-format - Core system change, but well-tested
- streamline-registry - Optimization with clear scope

**Deferred (No Risk):**
- generate-agent-configs - Existing functionality works

---

## Questions & Answers

**Q: Why defer generate-agent-configs when it's partially done?**
A: The core functionality exists and works. The proposed enhancements (project-aware generation, OpenSpec integration) are valuable but not blocking. Focus on core improvements first.

**Q: Can proposals be done in parallel?**
A: Some can (refactor-align-agent-tools can happen anytime), but sequence recommended: cleanup → core improvements → polish. Reduces merge conflicts and ensures clean foundation.

**Q: What if new issues arise during implementation?**
A: Address as they come. The roadmap is flexible. Create new proposals if needed for unexpected work.

**Q: How to track progress?**
A: Update this roadmap weekly. Mark tasks complete in individual `tasks.md` files. Use `/ent:status` for workflow tracking.

---

## Contact & Support

For questions about roadmap or proposal implementation:
- Review proposal files in `openspec/changes/*/`
- Check individual `tasks.md` for detailed breakdowns
- Use `/ent:architect` for design questions
- Use `/ent:planner` for implementation planning
- Use `/ent:coder` for implementation work

---

**Version:** 1.0
**Owner:** Go-Ent Maintainers
**Next Review:** 2026-02-06 (after Phase 1 completion)
