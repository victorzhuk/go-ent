# Proposal: Explicit Skill Triggers

## Summary

Add explicit trigger definitions to skill frontmatter with weights and pattern matching support.

## Status: complete

## Problem

Current trigger system relies on keyword extraction from description strings:
- Triggers buried in description text
- No way to specify trigger weights
- No pattern matching support
- No file-type triggers

This leads to:
- Inconsistent skill activation
- No control over matching precision
- Can't leverage weighted scoring (until implemented)

## Solution

Add explicit `triggers` section to frontmatter:

```yaml
---
name: go-code
description: "Expert Go developer..."
triggers:
  - pattern: "implement.*go"
    weight: 0.9
  - keywords: ["go code", "golang"]
    weight: 0.8
  - file_pattern: "*.go"
    weight: 0.6
---
```

**Features**:
- Pattern matching: Regex patterns for flexible matching
- Keywords: Exact keyword lists
- File patterns: Match on file extensions/paths
- Weights: Control relevance (0.0-1.0)

**Backward Compatibility**: Description-based trigger extraction continues working as fallback.

## Breaking Changes

- [ ] None - additive only, old format continues working

## Affected Systems

- **Parser** (`internal/skill/parser.go`): Parse triggers section
- **Registry** (`internal/skill/registry.go`): Use explicit triggers when available, fallback to description
- **Validator** (`internal/skill/validator.go`): Add SK012 rule recommending explicit triggers

## Alternatives Considered

1. **Keep description-based only**: No changes
   - ❌ No control over matching precision

2. **AI-learned triggers**: Train model on skill usage
   - ❌ Over-engineered, opaque

3. **Explicit triggers** (chosen):
   - ✅ Precise control
   - ✅ Human-readable
   - ✅ Backward compatible
