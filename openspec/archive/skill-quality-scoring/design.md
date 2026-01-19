# Design: Research-Aligned Quality Scoring

## Scoring Algorithm

### Structure Score (20 points max)

Award points for required XML sections present:

```go
type StructureScore struct {
    Role         bool  // 4 points
    Instructions bool  // 4 points
    Constraints  bool  // 3 points
    Examples     bool  // 3 points
    OutputFormat bool  // 3 points
    EdgeCases    bool  // 3 points
}

func calculateStructureScore(skill *Skill) int {
    score := 0
    if hasSection(skill, "role") { score += 4 }
    if hasSection(skill, "instructions") { score += 4 }
    if hasSection(skill, "constraints") { score += 3 }
    if hasSection(skill, "examples") { score += 3 }
    if hasSection(skill, "output_format") { score += 3 }
    if hasSection(skill, "edge_cases") { score += 3 }
    return score
}
```

### Content Score (25 points max)

Evaluate quality of key sections:

```go
type ContentScore struct {
    RoleClarity        int  // 0-8 points
    InstructionQuality int  // 0-9 points
    ConstraintQuality  int  // 0-8 points
}

func calculateContentScore(skill *Skill) int {
    roleScore := scoreRoleClarity(skill.Role)
    instructionScore := scoreInstructions(skill.Instructions)
    constraintScore := scoreConstraints(skill.Constraints)
    return roleScore + instructionScore + constraintScore
}

func scoreRoleClarity(role string) int {
    // 8 points max
    score := 0
    if hasExpertiseLevel(role) { score += 3 }  // "senior", "expert", "10+ years"
    if hasDomainFocus(role) { score += 3 }     // specific domain mentioned
    if hasBehaviorGuidelines(role) { score += 2 }  // behavioral constraints
    return score
}

func scoreInstructions(instructions string) int {
    // 9 points max
    score := 0
    if hasActionableSteps(instructions) { score += 4 }  // imperative verbs, numbered steps
    if hasSpecificity(instructions) { score += 3 }      // concrete, not vague
    if hasStructure(instructions) { score += 2 }        // organized, scannable
    return score
}

func scoreConstraints(constraints string) int {
    // 8 points max
    score := 0
    if hasWhatToInclude(constraints) { score += 3 }  // positive constraints
    if hasWhatToExclude(constraints) { score += 3 }  // boundaries
    if hasSpecificity(constraints) { score += 2 }    // concrete rules
    return score
}
```

### Examples Score (25 points max)

**Research finding**: 3-5 diverse examples dramatically improve consistency.

```go
type ExamplesScore struct {
    Count      int  // 0-10 points
    Diversity  int  // 0-8 points
    EdgeCases  int  // 0-4 points
    Format     int  // 0-3 points
}

func calculateExamplesScore(skill *Skill) int {
    examples := extractExamples(skill)

    // Count scoring (10 points max)
    countScore := 0
    switch len(examples) {
    case 0: countScore = 0
    case 1: countScore = 3
    case 2: countScore = 6
    case 3, 4, 5: countScore = 10  // Optimal range
    default: countScore = 8  // >5 examples, slight penalty for verbosity
    }

    // Diversity scoring (8 points max)
    diversityScore := scoreDiversity(examples)  // Different input types, edge cases

    // Edge case scoring (4 points max)
    edgeCaseScore := countEdgeCases(examples)  // Empty input, error cases, boundary conditions

    // Format scoring (3 points max)
    formatScore := 0
    if allHaveInputOutput(examples) { formatScore += 2 }
    if properXMLStructure(examples) { formatScore += 1 }

    return countScore + diversityScore + edgeCaseScore + formatScore
}

func scoreDiversity(examples []Example) int {
    // Check for variety in:
    // - Input complexity (simple, medium, complex)
    // - Input types (different data structures)
    // - Expected behaviors (success, error, edge case)
    // Return 0-8 based on diversity
}

func countEdgeCases(examples []Example) int {
    count := 0
    for _, ex := range examples {
        if isEdgeCase(ex) { count++ }
    }
    // 2 points per edge case, max 4 points
    return min(count*2, 4)
}
```

### Triggers Score (15 points max)

Favor explicit triggers over description-based:

```go
func calculateTriggersScore(skill *Skill) int {
    score := 0

    // Explicit triggers present (10 points)
    if len(skill.Frontmatter.Triggers) > 0 {
        score += 10

        // Bonus for weighted triggers (3 points)
        if allHaveWeights(skill.Frontmatter.Triggers) {
            score += 3
        }

        // Bonus for diverse trigger types (2 points)
        types := getTriggerTypes(skill.Frontmatter.Triggers)  // pattern, keywords, file_pattern
        if len(types) >= 2 {
            score += 2
        }
    } else {
        // Fallback: description-based triggers (5 points max)
        triggers := extractTriggersFromDescription(skill.Frontmatter.Description)
        if len(triggers) > 0 {
            score += 5
        }
    }

    return score
}
```

### Conciseness Score (15 points max)

**Research finding**: Skills >5k tokens suffer from attention dilution.

```go
func calculateConcisenessScore(skill *Skill) int {
    tokens := countTokens(skill.Body)  // Approximate: words * 1.3

    // Scoring curve
    switch {
    case tokens < 3000:
        return 15  // Ideal: concise and focused
    case tokens < 5000:
        return 10  // Acceptable: moderate length
    case tokens < 8000:
        return 5   // Warning: approaching attention limits
    default:
        return 0   // Critical: likely suffering from attention dilution
    }
}
```

## Total Score Calculation

```go
type QualityScore struct {
    Total       int
    Structure   int
    Content     int
    Examples    int
    Triggers    int
    Conciseness int
}

func CalculateQualityScore(skill *Skill) QualityScore {
    return QualityScore{
        Structure:   calculateStructureScore(skill),      // max 20
        Content:     calculateContentScore(skill),        // max 25
        Examples:    calculateExamplesScore(skill),       // max 25
        Triggers:    calculateTriggersScore(skill),       // max 15
        Conciseness: calculateConcisenessScore(skill),    // max 15
        Total:       structure + content + examples + triggers + conciseness,
    }
}
```

## CLI Output Format

```
Quality Score: 82/100

Breakdown:
  Structure    ████████████████████ 20/20
  Content      ████████████████░░░░ 20/25
  Examples     ██████████████░░░░░░ 18/25
  Triggers     ████████░░░░░░░░░░░░ 10/15
  Conciseness  ██████████░░░░░░░░░░ 14/15

Recommendations:
  • Add 1-2 more diverse examples (currently 3, optimal is 4-5)
  • Define explicit triggers in frontmatter for better activation
```

## Migration Strategy

### Impact Analysis

Run scorer on all existing skills:

```bash
go run cmd/go-ent/main.go skill analyze --all
```

Generate report:
```
Total skills: 12
Score distribution:
  90-100: 2 skills (excellent)
  80-89:  4 skills (good)
  70-79:  3 skills (needs improvement)
  <70:    3 skills (requires attention)

Most common issues:
  • 8 skills missing edge case examples
  • 5 skills >5k tokens (attention dilution risk)
  • 7 skills using description-based triggers only
```

### Rollout Plan

1. **Phase 1**: Deploy new scoring (informational only)
2. **Phase 2**: Update documentation with new scoring criteria
3. **Phase 3**: Improve existing skills based on new scores
4. **Phase 4**: Set minimum score threshold (optional, future)

## Testing Strategy

### Unit Tests

```go
func TestStructureScore(t *testing.T) {
    tests := []struct {
        name     string
        skill    *Skill
        expected int
    }{
        {"all sections", skillWithAllSections(), 20},
        {"missing role", skillWithoutRole(), 16},
        {"minimal", skillMinimal(), 11},
    }
    // ...
}

func TestExamplesScore(t *testing.T) {
    tests := []struct {
        name     string
        examples []Example
        expected int
    }{
        {"optimal count (4)", fourDiverseExamples(), 25},
        {"too few (1)", oneExample(), 12},
        {"no edge cases", threeExamplesNoEdgeCases(), 18},
    }
    // ...
}

func TestConcisenessScore(t *testing.T) {
    tests := []struct {
        tokens   int
        expected int
    }{
        {2000, 15},  // Ideal
        {4000, 10},  // Acceptable
        {7000, 5},   // Warning
        {10000, 0},  // Critical
    }
    // ...
}
```

### Integration Tests

```go
func TestFullSkillScoring(t *testing.T) {
    // Load real skill files
    skills := loadSkills("testdata/skills/")

    for _, skill := range skills {
        score := CalculateQualityScore(skill)

        // Verify score is in valid range
        assert.GreaterOrEqual(t, score.Total, 0)
        assert.LessOrEqual(t, score.Total, 100)

        // Verify breakdown matches total
        assert.Equal(t, score.Total,
            score.Structure + score.Content + score.Examples +
            score.Triggers + score.Conciseness)
    }
}
```

## Performance Considerations

- **Token counting**: Cache results per skill
- **Diversity scoring**: O(n²) comparison of examples - acceptable for n<10
- **Regex matching**: Compile patterns once, reuse
- **Estimated time**: <10ms per skill (acceptable for validation workflow)
