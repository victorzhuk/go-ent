# Design: Weighted Trigger Matching

## Data Models

```go
type MatchResult struct {
    Skill     *SkillMetadata
    Score     float64
    MatchedBy []MatchReason
}

type MatchReason struct {
    Type   string   // "keyword", "pattern", "file_type"
    Value  string
    Weight float64
}
```

## Scoring Algorithm

```go
func (r *Registry) FindMatchingSkills(query string) []MatchResult {
    results := []MatchResult{}

    for _, skill := range r.skills {
        score, reasons := r.scoreSkill(skill, query)
        if score > 0 {
            results = append(results, MatchResult{
                Skill:     skill,
                Score:     score,
                MatchedBy: reasons,
            })
        }
    }

    // Sort by score descending
    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })

    return results
}

func (r *Registry) scoreSkill(skill *SkillMetadata, query string) (float64, []MatchReason) {
    reasons := []MatchReason{}
    maxScore := 0.0

    // Score explicit triggers
    if len(skill.Triggers) > 0 {
        for _, trigger := range skill.Triggers {
            if matched, reason := r.matchTrigger(trigger, query); matched {
                reasons = append(reasons, reason)
                if reason.Weight > maxScore {
                    maxScore = reason.Weight
                }
            }
        }
    } else {
        // Fallback: description-based
        if matched, reason := r.matchDescription(skill.Description, query); matched {
            reasons = append(reasons, reason)
            maxScore = max(maxScore, reason.Weight)
        }
    }

    return maxScore, reasons
}

func (r *Registry) matchTrigger(trigger Trigger, query string) (bool, MatchReason) {
    // Pattern matching
    if trigger.Pattern != "" {
        if matched := regexp.MustCompile(trigger.Pattern).MatchString(query); matched {
            return true, MatchReason{
                Type:   "pattern",
                Value:  trigger.Pattern,
                Weight: trigger.Weight,
            }
        }
    }

    // Keyword matching
    for _, keyword := range trigger.Keywords {
        if strings.Contains(strings.ToLower(query), strings.ToLower(keyword)) {
            return true, MatchReason{
                Type:   "keyword",
                Value:  keyword,
                Weight: trigger.Weight,
            }
        }
    }

    return false, MatchReason{}
}
```

## Migration Path

Update all callers:
```go
// Before
skills := registry.FindMatchingSkills(query)
for _, skill := range skills {
    fmt.Println(skill.Name)
}

// After
matches := registry.FindMatchingSkills(query)
for _, match := range matches {
    fmt.Printf("%s (score: %.2f)\n", match.Skill.Name, match.Score)
}
```

## Performance

- Regex compilation: Cache compiled patterns (10x speedup)
- Scoring: O(n*m) where n=skills, m=triggers per skill (acceptable for <1000 skills)
