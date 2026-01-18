# Proposal: Enhanced Validation Rules

## Summary

Add four new validation rules aligned with research findings: example diversity, instruction conciseness, explicit triggers, and redundancy detection.

## Problem

Current 9 validation rules (SK001-SK009) cover basic structure but miss critical quality factors identified in research:
- **No diversity check**: Skills can pass with 5 identical examples
- **No conciseness check**: Skills can be 20k tokens and pass validation
- **No trigger format check**: Description-based triggers are silent failures
- **No redundancy check**: Skills with 90% overlap go undetected

These gaps allow low-quality skills that technically "pass" but perform poorly in practice.

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

## Alternatives Considered

1. **Make rules errors instead of warnings**: Block skills that fail
   - ❌ Too restrictive, existing skills would break

2. **Add more granular rules**: 20+ specific checks
   - ❌ Over-engineering, diminishing returns

3. **Four targeted rules** (chosen):
   - ✅ Addresses research-identified gaps
   - ✅ Non-blocking (warnings/info)
   - ✅ Actionable guidance
