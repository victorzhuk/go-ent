# Proposal: Research-Aligned Quality Scoring

## Summary

Update skill quality scoring to align with research findings from docs/research/SKILL.md, emphasizing examples, conciseness, and explicit triggers.

## Problem

Current scoring breakdown doesn't reflect research priorities:
- **Current**: Frontmatter (20), Structure (30), Content (30), Triggers (20)
- **Issues**:
  - No penalty for verbose skills (>5k tokens)
  - No scoring for example quality/diversity
  - Triggers weighted too high for description-based matching
  - Structure and Content categories overlap

Research shows:
- 3-5 diverse examples dramatically improve consistency
- Skills >5k tokens suffer from attention dilution
- Explicit triggers with weights are more valuable than keyword matching

## Solution

New scoring breakdown aligned with research (100 points total):

| Category | Points | Criteria |
|----------|--------|----------|
| **Structure** | 20 | XML sections present (role, instructions, constraints, examples, output_format, edge_cases) |
| **Content** | 25 | Role clarity, instruction actionability, constraint specificity |
| **Examples** | 25 | Count (3-5), diversity, edge cases, proper format |
| **Triggers** | 15 | Explicit triggers with weights vs description-only |
| **Conciseness** | 15 | Token count penalty curve (<3k=15pts, 3-5k=10pts, 5-8k=5pts, >8k=0pts) |

**Key changes**:
- New **Examples** category (25 points) - research priority
- New **Conciseness** category (15 points) - prevent attention dilution
- Reduced **Triggers** from 20→15 points - less critical until explicit triggers added
- Refined **Content** and **Structure** to avoid overlap

## Breaking Changes

- [ ] Quality scores will change for existing skills
- [ ] Skills previously passing may now score lower (if verbose or lacking examples)

**Migration**:
- Scores are informational, not blocking
- No functionality breaks
- Authors see new scores and can improve skills

## Affected Systems

- **Scorer** (`internal/skill/scorer.go`): Complete scoring algorithm rewrite
- **Validation** (`internal/skill/validator.go`): Add conciseness checks
- **CLI Output**: Display new score breakdown by category

## Alternatives Considered

1. **Keep current scoring**: Maintain consistency
   - ❌ Doesn't reflect research findings on what makes skills effective

2. **Binary pass/fail**: Skip scoring entirely
   - ❌ Loses valuable feedback mechanism for authors

3. **Research-aligned scoring** (chosen):
   - ✅ Incentivizes proven patterns
   - ✅ Provides actionable improvement guidance
   - ✅ Based on empirical research
