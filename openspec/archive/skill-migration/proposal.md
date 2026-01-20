# Proposal: Skill Migration to New Standards

## Summary
Update all existing skills (14 files) to meet new quality standards and add missing skills identified during refactor.

**Current state:** 9 Go skills + 5 Core skills
**Missing skills confirmed:** go-migration, go-config, go-error
**Additional missing skills to be confirmed after audit:** go-concurrency, go-validation, go-idiomatic

**Note:** Quality analysis includes 15 skills (14 existing + 1 referenced), migration targets 14 core skills.

## Problem

Existing skills in `plugins/go-ent/skills/` were created before:
- Explicit trigger system
- Research-aligned quality standards
- Enhanced validation rules
- Quality scoring system

**Current state (audit completed):**
- 0/14 skills have explicit triggers
- Examples count: All skills have 3 examples (target: 3-5)
- Token counts: 1 skill verbose (debug-core at 639 lines)
- Quality scoring: Average 102.2 (all 15 skills scored, 14 for migration, threshold 80)
- Missing critical skills: go-migration, go-config, go-error

**Audit findings (15 skills analyzed, 14 for migration):**
- Quality distribution: 97-106, average 102.2
- Examples: All skills have 3 examples
- Triggers: 0/14 skills have explicit triggers
- Token concerns: 1 skill verbose (debug-core at 639 lines)

**Quality scores by skill (ascending priority for migration):**
- command-mastery: 97
- skill-mastery: 99
- go-code: 100
- go-test: 101
- go-db: 102
- go-api: 103
- go-perf: 104
- go-sec: 105
- go-review: 105
- debug-core: 106
- workflow-orchestrator: 106
- agent-mastery: 106
- arch-core: 106
- api-design: 106
- go-arch: 106

**Note:** Migration targets 14 skills. Quality analysis includes reference skills (command-mastery, skill-mastery, agent-mastery, workflow-orchestrator) for baseline comparison.

## Solution

### Phase 0: Complete Audit (NEW - must be done before migration)

**Audit completed - baseline established:**
1. **Inventory verification**
   - 14 skills for migration (9 Go + 5 Core)
   - 9 Go skills: go-code, go-arch, go-api, go-db, go-test, go-perf, go-sec, go-review, go-ops
   - 5 Core skills: api-design, arch-core, debug-core, review-core, security-core
   - Excluded: test-api-design-gen (test skill)

2. **Current state documented**
   - Examples per skill: All 14 have 3 examples
   - Token counts: 1 skill exceeds target (debug-core)
   - Explicit triggers: 0/14 skills
   - Quality scoring: Complete - average 102.2, all passing

3. **Missing skills identified**
   - Confirmed: go-migration, go-config, go-error
   - Decision: Skip go-concurrency, go-validation, go-idiomatic (covered by go-code)

4. **Audit results documented**
   - Baseline metrics established from 15 scored skills
   - Migration order: Lowest quality score first (go-code at 100, up to go-arch at 106)
   - Proposal updated with accurate data

### Phase 1: Migrate Existing Skills (9 Go skills + 5 Core skills)

**Migration order (lowest quality score first):**
1. go-code (100)
2. go-test (101)
3. go-db (102)
4. go-api (103)
5. go-perf (104)
6. go-sec (105)
7. go-review (105)
8. debug-core (106)
9. go-arch (106)
10. arch-core (106)
11. api-design (106)
12. go-ops (106) ✓
13. review-core (pending score)
14. security-core (pending score)

For each skill:
1. **Add explicit triggers** with weights
   ```yaml
   triggers:
     - keywords: ["go code", "golang"]
       weight: 0.8
     - file_pattern: "*.go"
       weight: 0.6
   ```

2. **Ensure 3-5 diverse examples**
   - Add examples if <3
   - Diversify if all similar
   - Include edge cases

3. **Trim verbose content**
   - Move detailed docs to references/
   - Keep core instructions <5k tokens

4. **Add dependencies** where appropriate
   - go-db → depends_on: [go-code]
   - go-testing → depends_on: [go-code]

5. **Validate quality score ≥80**

### Phase 2: Add Missing Skills

**go-migration**: Database migration patterns
```yaml
description: "Design and implement database migrations for Go applications. Use for schema changes, data migrations."
triggers:
  - keywords: ["migration", "schema change", "database migration"]
  - file_pattern: "**/migrations/*.sql"
```

**go-config**: Configuration management
```yaml
description: "Handle configuration in Go applications (env, files, flags). Use for config setup or issues."
triggers:
  - keywords: ["config", "configuration", "environment"]
  - file_pattern: "config.go"
```

**go-error**: Error handling patterns
```yaml
description: "Implement Go error handling patterns (wrapping, custom errors, error types). Use for error design."
triggers:
  - keywords: ["error handling", "error wrapping"]
```

**debug-core**: General debugging
```yaml
description: "Debug issues across languages. Use when no language-specific debugger matches."
triggers:
  - keywords: ["debug", "troubleshoot"]
  - weight: 0.5  # Lower weight, fallback skill
```

## Dependencies

**Must be completed BEFORE this proposal:**

1. **skill-quality-scoring** - Provides scoring system for audit
   - Status: Proposal exists, implementation pending
   - Required for: Baseline quality scores, migration prioritization

2. **skill-validation-rules** - Provides enhanced validation (SK010-SK013)
   - Status: Proposal exists, implementation pending
   - Required for: Example diversity checks, conciseness warnings

3. **skill-lint-tool** - Provides automated quality checks
   - Status: Proposal exists, implementation pending
   - Required for: Fast validation, CI/CD integration

**Execution Order:**
1. Implement skill-quality-scoring
2. Implement skill-validation-rules
3. Implement skill-lint-tool
4. Run Phase 0 audit with new tools
5. Execute migration phases 1-5 based on audit results

## Breaking Changes

- [ ] None - existing skills remain backward compatible
- Improved activation from explicit triggers

## Success Metrics

| Metric | Before (baseline) | After Target | Verification |
|--------|-------------------|--------------|---------------|
| Skills with explicit triggers | 0/14 (0%) | 14/14 (100%) | Audit + validation |
| Skills with 3+ examples | 14/14 (100%) | 14/14 (100%) | Phase 0 audit |
| Average quality score | 102.2 (all passing) | ≥85 | Quality scoring |
| Skills >5k tokens | 1/14 (7.1%) | <10% (1-2 skills max) | Phase 0 audit |

## Alternatives Considered

1. **Gradual migration**: Update skills as issues arise
   - ❌ Inconsistent quality across codebase

2. **Full migration** (chosen):
   - ✅ Consistent high quality
   - ✅ Demonstrates new standards
   - ✅ Validates tooling
