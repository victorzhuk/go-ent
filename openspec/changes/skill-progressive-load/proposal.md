# Proposal: Progressive Skill Loading

## Summary
Implement three-level loading: metadata (~100 tokens), core (<5k tokens), extended (unlimited) to optimize context usage.

## Problem
Currently all skill content loads at once:
- Wastes context window on unused skills
- Slow startup for large skill sets
- No way to defer detailed content

## Solution
Three disclosure levels:
1. **Level 1 (Metadata)**: Name, description, triggers only (~100 tokens)
2. **Level 2 (Core)**: + role, instructions, constraints, examples (<5k tokens)
3. **Level 3 (Extended)**: + references/, scripts/, detailed docs (unlimited)

Registry loads Level 1 by default, upgrades to Level 2/3 on demand.

## Breaking Changes
- [ ] None - transparent to skill authors, registry implementation detail

## Alternatives
1. **Full loading**: Current approach
   - ❌ Wastes context
2. **Progressive** (chosen):
   - ✅ Optimizes context usage
   - ✅ Faster startup
