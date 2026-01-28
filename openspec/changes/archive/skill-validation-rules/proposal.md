# Proposal: Enhanced Validation Rules

## Summary

**Status: complete**

Add four new validation rules (SK010-SK013) aligned with research findings and the research-based quality scoring system. These rules provide actionable feedback across five key quality categories: Structure, Content, Examples, Triggers, and Conciseness.

## Problem

Current 9 validation rules (SK001-SK009) cover basic structure but miss critical quality factors identified in research and the skill quality scoring system:
- **No diversity check**: Skills can pass with 5 identical examples, missing the **Examples** category's diversity requirement (up to 8pts)
- **No conciseness check**: Skills can be 20k tokens and pass validation, despite the **Conciseness** category penalizing >5k tokens
- **No trigger format check**: Description-based triggers are silent failures, even though **Triggers** category rewards explicit format (15pts max)
- **No redundancy check**: Skills with 90% overlap go undetected, causing issues in both **Structure** and **Content** categories

These gaps allow low-quality skills that technically "pass" but perform poorly in practice and score low on the quality metrics.

## Solution

Add 4 new validation rules:

### SK010: example-diversity (warning)
Checks that examples demonstrate variety in inputs, behaviors, and edge cases.
```
⚠️  Examples lack diversity
   💡 Include examples with different input types, success/error cases, and edge cases
   Example: Mix simple inputs, complex inputs, empty inputs, and boundary cases
```

### SK011: instruction-concise (warning)
Warns when skill body exceeds research-recommended token limits.
```
⚠️  Skill is verbose (8500 tokens, recommended <5000)
   💡 Reduce content to prevent attention dilution
   Example: Move detailed examples to separate reference files
```

### SK012: trigger-explicit (info)
Recommends explicit triggers over description-based extraction.
```
ℹ️  Using description-based triggers
   💡 Define explicit triggers in frontmatter for better matching
   Example:
      triggers:
        - keywords: ["go code", "golang"]
          weight: 0.8
```

### SK013: redundancy-check (warning)
Detects high overlap (>70%) with other skills in the system.
```
⚠️  High overlap (85%) with skill 'go-code'
   💡 Consider merging skills or clarifying distinct use cases
```

## Breaking Changes

- [ ] None - all new rules are warnings/info, not errors

## Affected Systems

- **Rules** (`internal/skill/rules.go`): Add 4 new rule implementations
- **Validator** (`internal/skill/validator.go`): Register new rules
- **Registry** (`internal/skill/registry.go`): Add overlap detection for SK013

## Alignment with Quality Scoring Categories

The new validation rules (SK010-SK013) align directly with the research-based quality scoring categories:

| Validation Rule | Quality Category | Scoring Impact | Alignment |
|----------------|------------------|----------------|-----------|
| **SK010: example-diversity** | Examples | 0-8pts for diversity | Warns when examples lack variety, directly addressing the diversity scoring criteria |
| **SK011: instruction-concise** | Conciseness | 0-15pts (penalty >5k tokens) | Warns when skills exceed token limits, matching the conciseness scoring thresholds |
| **SK012: trigger-explicit** | Triggers | 15pts max for explicit format | Encourages explicit triggers which score higher than description-based (5pts max) |
| **SK013: redundancy-check** | Structure + Content | Impacts multiple categories | Detects overlap that dilutes scores across structure, content, and examples |

### How Validation Rules Complement Scoring

**Validation rules** (SK001-SK013) provide immediate, actionable feedback during skill development:
- Binary pass/warn/info checks
- Specific, targeted guidance
- Fast feedback loop for authors

**Quality scoring** provides comprehensive evaluation across 100 points:
- nuanced assessment (0-100 scale)
- Multiple dimensions (5 categories)
- Benchmarking and progress tracking

Together they form a two-tiered quality system:
1. **Validation**: Quick checks for common issues (prevents obvious problems)
2. **Scoring**: Comprehensive evaluation (identifies improvement opportunities)

## Alternatives Considered

1. **Make rules errors instead of warnings**: Block skills that fail
   - ❌ Too restrictive, existing skills would break

2. **Add more granular rules**: 20+ specific checks
   - ❌ Over-engineering, diminishing returns

3. **Four targeted rules** (chosen):
   - ✅ Addresses research-identified gaps
   - ✅ Aligns with quality scoring categories
   - ✅ Non-blocking (warnings/info)
   - ✅ Actionable guidance

## Completion Summary

All 11 tasks completed successfully:
- SK010-SK013 validation rules implemented in `internal/skill/rules.go`
- Rules registered and integrated with validator
- CLI updated to display warnings and info messages
- Tests passing for all validation rules
- Non-blocking feedback system providing actionable guidance during skill development

The implementation aligns with the quality scoring system by warning about:
- Example diversity (directly addressing the Examples category)
- Skill conciseness (matching Conciseness category thresholds)
- Trigger format (encouraging explicit triggers for higher Triggers scores)
- Content redundancy (impacting multiple quality categories)
