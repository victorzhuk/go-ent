# Proposal: Weighted Trigger Matching

## Summary

Replace simple keyword matching with weighted scoring system that ranks skills by relevance using explicit triggers with pattern matching and file-type awareness.

## Problem

Current matching in `registry.go`:
- Binary yes/no match (keyword found or not)
- No relevance ranking (all matches equal)
- No confidence scores
- Can't differentiate between perfect and weak matches

This leads to:
- Wrong skills activating for ambiguous queries
- No way to show "best match" to users
- Can't leverage trigger weights from explicit triggers

## Solution

Replace `FindMatchingSkills(query) []*Skill` with `FindMatchingSkills(query) []MatchResult`:

```go
type MatchResult struct {
    Skill     *SkillMetadata
    Score     float64      // 0.0-1.0
    MatchedBy []MatchReason
}

type MatchReason struct {
    Type   string    // "keyword", "pattern", "file_type"
    Value  string    // What matched
    Weight float64   // Trigger weight
}
```

**Scoring algorithm**:
1. Pattern matching: Regex patterns with configurable weights (0.7-1.0)
2. Keyword matching: Exact/fuzzy keyword matches (0.6-0.9)
3. File-type matching: File patterns like "*.go" (0.5-0.8)
4. Return sorted by score (highest first)

## Breaking Changes

- [x] API change: `FindMatchingSkills` return type changes
- Migration: Update all callers to handle `MatchResult` instead of `*Skill`

## Affected Systems

- **Registry** (`internal/skill/registry.go`): Complete matching rewrite
- **CLI** (`cmd/go-ent/main.go`): Display match scores
- **Skills** (`plugins/go-ent/skills/`): Can now use weighted triggers

## Alternatives Considered

1. **Keep simple matching**: No changes
   - ❌ Doesn't leverage explicit triggers

2. **ML-based matching**: Embeddings and semantic search
   - ❌ Over-engineered, adds dependencies

3. **Weighted scoring** (chosen):
   - ✅ Precise control via trigger weights
   - ✅ No external dependencies
   - ✅ Fast and deterministic
