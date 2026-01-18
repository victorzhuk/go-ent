# Proposal: Enhanced Skill Validation Error Messages

## Summary

Add actionable suggestions and examples to validation error messages to help skill authors fix issues faster.

## Problem

Current validation errors only describe what's wrong, but don't guide authors on how to fix issues:
- "Missing examples section" - but what should examples look like?
- "Name format invalid" - but what's the correct format?
- "Description too short" - but how long should it be?

Authors must consult documentation to understand fixes, slowing iteration cycles.

## Solution

Extend `ValidationError` struct with `Suggestion` and `Example` fields:

```go
type ValidationError struct {
    Field      string
    Message    string
    Suggestion string  // How to fix
    Example    string  // Correct usage
}
```

Each validation rule provides concrete fix guidance:
- **What's wrong**: "Missing examples section"
- **How to fix**: "Add <examples> section with 3-5 diverse examples"
- **Example**: Shows proper XML structure with example/input/output tags

## Breaking Changes

- [ ] None - additive only

## Affected Systems

- **Validator** (`internal/skill/validator.go`): Add new fields to error struct
- **Validation Rules** (`internal/skill/rules.go`): Update all rules to provide suggestions
- **CLI Output**: Enhanced error formatting to display suggestions

## Alternatives Considered

1. **External documentation links**: Point to docs instead of inline help
   - ❌ Requires context switching, slower iteration

2. **Auto-fix suggestions**: Generate fixes automatically
   - ❌ Too complex, may not match author intent

3. **Current approach**: Inline suggestions with examples
   - ✅ Fast feedback loop, self-documenting
