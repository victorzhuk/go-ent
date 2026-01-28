# Proposal: Context-Aware Skill Matching

## Summary
Enhance matching algorithm to consider full context: query + file types + task type + active skills.

## Status: complete

## Problem
Current matching only considers query string, ignoring:
- Files being worked on (e.g., *.go files open)
- Task type (implement vs review vs debug)
- Already active skills

## Solution
```go
type MatchContext struct {
    Query        string
    FileTypes    []string  // [".go", ".md"]
    TaskType     string    // "implement", "review", "debug"
    ActiveSkills []string  // Currently loaded skills
}
```

Boost scores based on context:
- File-type match: +0.2 score
- Task-type match: +0.15 score
- Skill affinity: +0.1 score

## Breaking Changes
- [x] `FindMatchingSkills` adds optional context parameter

## Alternatives
1. **Query-only matching**: Current approach
   - ❌ Misses valuable context
2. **Context-aware** (chosen):
   - ✅ More accurate skill selection
