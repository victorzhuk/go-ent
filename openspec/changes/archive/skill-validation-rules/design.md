# Design: Enhanced Validation Rules

## Rule Implementations

### SK010: example-diversity

```go
func checkExampleDiversity(skill *Skill) *ValidationWarning {
    examples := extractExamples(skill)
    if len(examples) < 3 {
        return nil  // Count check handled by SK004
    }

    diversityScore := calculateDiversityScore(examples)
    if diversityScore < 0.5 {  // <50% diversity
        return &ValidationWarning{
            Field:   "examples",
            Message: "Examples lack diversity",
            Suggestion: "Include examples with different input types, success/error cases, and edge cases",
            Example: "Mix simple inputs, complex inputs, empty inputs, and boundary cases",
        }
    }
    return nil
}

func calculateDiversityScore(examples []Example) float64 {
    if len(examples) == 0 {
        return 0
    }

    factors := []float64{
        checkInputComplexityVariety(examples),  // Simple, medium, complex
        checkBehaviorVariety(examples),         // Success, error, edge case
        checkDataTypeVariety(examples),         // Different data structures
    }

    sum := 0.0
    for _, f := range factors {
        sum += f
    }
    return sum / float64(len(factors))
}
```

### SK011: instruction-concise

```go
func checkInstructionConcise(skill *Skill) *ValidationWarning {
    tokens := countTokens(skill.Body)

    var severity string
    var message string

    switch {
    case tokens > 8000:
        severity = "critical"
        message = fmt.Sprintf("Skill is very verbose (%d tokens, recommended <5000)", tokens)
    case tokens > 5000:
        severity = "warning"
        message = fmt.Sprintf("Skill is verbose (%d tokens, recommended <5000)", tokens)
    default:
        return nil  // Under threshold
    }

    return &ValidationWarning{
        Field:   "body",
        Message: message,
        Suggestion: "Reduce content to prevent attention dilution. Move detailed examples to references/",
        Example: "Keep core instructions <5k tokens, use progressive disclosure for extended content",
    }
}

func countTokens(text string) int {
    // Approximate: split on whitespace and multiply by 1.3
    words := len(strings.Fields(text))
    return int(float64(words) * 1.3)
}
```

### SK012: trigger-explicit

```go
func checkTriggerExplicit(skill *Skill) *ValidationWarning {
    // Check if skill uses new explicit triggers format
    if len(skill.Frontmatter.Triggers) > 0 {
        return nil  // Using explicit triggers
    }

    // Check if description contains trigger keywords
    desc := strings.ToLower(skill.Frontmatter.Description)
    triggerKeywords := []string{"use when", "auto-activates for", "triggers on"}

    hasTriggerHints := false
    for _, kw := range triggerKeywords {
        if strings.Contains(desc, kw) {
            hasTriggerHints = true
            break
        }
    }

    if !hasTriggerHints {
        return &ValidationWarning{
            Field:   "triggers",
            Message: "No activation triggers defined",
            Suggestion: "Define explicit triggers in frontmatter for better skill matching",
            Example: `triggers:
  - keywords: ["go code", "golang"]
    weight: 0.8
  - file_pattern: "*.go"
    weight: 0.6`,
        }
    }

    return &ValidationWarning{
        Field:   "triggers",
        Message: "Using description-based triggers",
        Suggestion: "Consider defining explicit triggers in frontmatter for more precise matching",
        Example: `triggers:
  - keywords: ["go code", "golang"]
    weight: 0.8`,
    }
}
```

### SK013: redundancy-check

```go
func checkRedundancy(skill *Skill, allSkills []*Skill) *ValidationWarning {
    var highestOverlap float64
    var overlapSkill string

    for _, other := range allSkills {
        if other.Name == skill.Name {
            continue
        }

        overlap := calculateOverlap(skill, other)
        if overlap > highestOverlap {
            highestOverlap = overlap
            overlapSkill = other.Name
        }
    }

    if highestOverlap > 0.7 {  // >70% overlap
        return &ValidationWarning{
            Field:   "content",
            Message: fmt.Sprintf("High overlap (%.0f%%) with skill '%s'",
                highestOverlap*100, overlapSkill),
            Suggestion: "Consider merging skills or clarifying distinct use cases in descriptions",
            Example: "If both handle Go code, differentiate by trigger patterns or use depends_on",
        }
    }

    return nil
}

func calculateOverlap(skill1, skill2 *Skill) float64 {
    // Calculate trigger overlap
    triggerOverlap := calculateTriggerOverlap(skill1, skill2)

    // Calculate description similarity (basic token overlap)
    descOverlap := calculateTextSimilarity(
        skill1.Frontmatter.Description,
        skill2.Frontmatter.Description,
    )

    // Weighted average
    return 0.7*triggerOverlap + 0.3*descOverlap
}

func calculateTriggerOverlap(skill1, skill2 *Skill) float64 {
    triggers1 := extractAllTriggers(skill1)
    triggers2 := extractAllTriggers(skill2)

    if len(triggers1) == 0 || len(triggers2) == 0 {
        return 0
    }

    intersection := 0
    for _, t1 := range triggers1 {
        for _, t2 := range triggers2 {
            if strings.EqualFold(t1, t2) {
                intersection++
                break
            }
        }
    }

    union := len(triggers1) + len(triggers2) - intersection
    return float64(intersection) / float64(union)
}
```

## Rule Registration

```go
// In internal/skill/validator.go
func NewValidator() *Validator {
    return &Validator{
        rules: []ValidationRule{
            // Existing rules SK001-SK009
            checkNameRequired,
            checkNameFormat,
            // ...

            // New rules SK010-SK013
            checkExampleDiversity,
            checkInstructionConcise,
            checkTriggerExplicit,
            // checkRedundancy registered separately (needs all skills)
        },
    }
}

func (v *Validator) Validate(skill *Skill) ValidationResult {
    result := ValidationResult{}

    // Run standard rules
    for _, rule := range v.rules {
        if err := rule(skill); err != nil {
            result.AddError(err)
        }
        if warn := rule(skill); warn != nil {
            result.AddWarning(warn)
        }
    }

    return result
}

func (v *Validator) ValidateWithContext(skill *Skill, allSkills []*Skill) ValidationResult {
    result := v.Validate(skill)

    // Run context-aware rules
    if warn := checkRedundancy(skill, allSkills); warn != nil {
        result.AddWarning(warn)
    }

    return result
}
```

## Testing Strategy

```go
func TestSK010_ExampleDiversity(t *testing.T) {
    tests := []struct {
        name      string
        examples  []Example
        expectErr bool
    }{
        {"diverse examples", diverseExamples(), false},
        {"identical examples", identicalExamples(), true},
        {"mixed diversity", mixedExamples(), false},
    }
    // ...
}

func TestSK011_InstructionConcise(t *testing.T) {
    tests := []struct {
        tokens    int
        expectErr bool
    }{
        {3000, false},
        {6000, true},
        {10000, true},
    }
    // ...
}

func TestSK012_TriggerExplicit(t *testing.T) {
    tests := []struct {
        name      string
        skill     *Skill
        expectErr bool
    }{
        {"explicit triggers", skillWithExplicitTriggers(), false},
        {"description-based", skillWithDescTriggers(), true},
        {"no triggers", skillWithoutTriggers(), true},
    }
    // ...
}

func TestSK013_Redundancy(t *testing.T) {
    skill1 := &Skill{Name: "go-code", /* ... */}
    skill2 := &Skill{Name: "go-quality", /* 70% overlap */}
    skill3 := &Skill{Name: "python-code", /* 10% overlap */}

    warn := checkRedundancy(skill1, []*Skill{skill2, skill3})
    assert.NotNil(t, warn)
    assert.Contains(t, warn.Message, "go-quality")
}
```
