# Proposal: Skill Templates

## Summary
Create three skill templates (basic, complete, delegating) to help authors start with correct structure and best practices.

## Problem
New skill authors start from scratch or copy existing skills, leading to:
- Missing required sections
- Inconsistent formatting
- Unclear what makes a good skill

## Solution
Add templates in `plugins/go-ent/templates/`:
1. **skill-basic.md**: Minimal valid skill (passes validation)
2. **skill-complete.md**: Full example with all sections
3. **skill-delegating.md**: Shows skill composition with depends_on/delegates_to

## Breaking Changes
- [ ] None - additive only

## Alternatives Considered
1. **Interactive generator**: CLI tool that asks questions
   - ❌ Over-engineered for simple use case
2. **Templates only** (chosen):
   - ✅ Simple, copyable, self-documenting
