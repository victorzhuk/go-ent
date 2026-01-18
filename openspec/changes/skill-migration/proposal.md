# Proposal: Skill Migration to New Standards

## Summary
Update all existing skills (12 files) to meet new quality standards and add 4 missing skills identified during refactor.

## Problem
Existing skills in `plugins/go-ent/skills/` were created before:
- Explicit trigger system
- Research-aligned quality standards
- Enhanced validation rules
- Quality scoring system

Current state (estimated):
- 0% have explicit triggers
- ~50% have 3+ examples
- ~30% are >5k tokens (verbose)
- Missing critical skills: go-migration, go-config, go-error, debug-core

## Solution

### Phase 1: Migrate Existing Skills (9 Go skills + 3 core skills)

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

## Breaking Changes

- [ ] None - existing skills remain backward compatible
- Improved activation from explicit triggers

## Success Metrics

| Metric | Before | After Target |
|--------|--------|--------------|
| Skills with explicit triggers | 0% | 100% |
| Skills with 3+ examples | 50% | 100% |
| Average quality score | ~70 | >85 |
| Skills >5k tokens | 30% | <10% |

## Alternatives Considered

1. **Gradual migration**: Update skills as issues arise
   - ❌ Inconsistent quality across codebase

2. **Full migration** (chosen):
   - ✅ Consistent high quality
   - ✅ Demonstrates new standards
   - ✅ Validates tooling
