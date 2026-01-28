# Proposal: Skill Dependencies and Delegation

## Summary
Add `depends_on` and `delegates_to` fields to enable skill composition and hierarchical relationships.

## Status
complete

## Problem
Skills can't reference or delegate to other skills, leading to:
- Duplication of common functionality
- No way to compose specialized skills
- Flat skill hierarchy

## Solution
Add to frontmatter:
```yaml
depends_on: [base-skill, common-utils]
delegates_to:
  specialized-skill: "For complex cases, delegate to X"
```

Registry handles:
- Dependency resolution and loading
- Circular dependency detection
- Delegation hints in skill selection

## Breaking Changes
- [x] None - optional fields, backward compatible

## Alternatives
1. **Duplication**: Copy common content
   - ❌ Maintenance burden
2. **Composition** (chosen):
   - ✅ DRY principle
   - ✅ Clear specialization
