# Proposal: Skill Templates

## Summary
Created eleven comprehensive skill templates to help authors start with correct structure and best practices.

## Problem
New skill authors start from scratch or copy existing skills, leading to:
- Missing required sections
- Inconsistent formatting
- Unclear what makes a good skill

## Solution
Added templates in `plugins/go-ent/templates/`:
1. **go-basic/**: Minimal valid Go skill (passes validation)
2. **go-complete/**: Full Go skill example with all sections and delegation patterns
3. **go-arch/**: Clean Architecture and DDD patterns
4. **go-api/**: OpenAPI/ogen and gRPC design
5. **go-code/**: Modern Go implementation patterns
6. **go-config/**: Environment and file configuration
7. **go-db/**: PostgreSQL, ClickHouse, Redis integration
8. **go-error/**: Custom error types and wrapping
9. **go-migration/**: Database migrations with goose
10. **go-ops/**: Docker, Kubernetes, CI/CD
11. **go-perf/**: Performance profiling and optimization

Delegation patterns integrated into go-complete edge_cases section (delegates to go-perf, go-arch, go-test, go-sec, go-db, go-api).

## Breaking Changes
- [x] None - additive only

## Alternatives Considered
1. **Interactive generator**: CLI tool that asks questions
   - ❌ Over-engineered for simple use case
2. **Templates only** (chosen):
   - ✅ Simple, copyable, self-documenting

## Status
Complete - All 11 templates implemented and validated with comprehensive test suite (quality scores >= 90). Documentation in docs/TEMPLATE-CREATION.md (948 lines) and updated docs/SKILL-AUTHORING.md.
