# Design: Explicit Skill Triggers

## Data Model

```go
// In internal/skill/parser.go

type Frontmatter struct {
    Name        string
    Description string
    Version     string
    Author      string
    Tags        []string
    Triggers    []Trigger  // NEW
}

type Trigger struct {
    Pattern     string   `yaml:"pattern"`      // Regex pattern
    Keywords    []string `yaml:"keywords"`     // Exact keywords
    FilePattern string   `yaml:"file_pattern"` // File glob pattern
    Weight      float64  `yaml:"weight"`       // 0.0-1.0
}
```

## Parsing Implementation

```go
func (p *Parser) parseFrontmatter(content []byte) (*Frontmatter, error) {
    var fm Frontmatter
    if err := yaml.Unmarshal(content, &fm); err != nil {
        return nil, err
    }

    // Validate trigger weights
    for i := range fm.Triggers {
        if fm.Triggers[i].Weight == 0 {
            fm.Triggers[i].Weight = 0.7  // Default weight
        }
        if fm.Triggers[i].Weight < 0 || fm.Triggers[i].Weight > 1 {
            return nil, fmt.Errorf("trigger weight must be 0.0-1.0")
        }
    }

    return &fm, nil
}
```

## Backward Compatibility

```go
func (r *Registry) getTriggers(skill *Skill) []Trigger {
    // Use explicit triggers if present
    if len(skill.Frontmatter.Triggers) > 0 {
        return skill.Frontmatter.Triggers
    }

    // Fallback: extract from description
    keywords := extractTriggersFromDescription(skill.Frontmatter.Description)
    return []Trigger{{
        Keywords: keywords,
        Weight:   0.5,  // Lower weight for extracted triggers
    }}
}
```

## Migration Path

1. **Phase 1**: Add Trigger parsing (non-breaking)
2. **Phase 2**: Update registry to use explicit triggers
3. **Phase 3**: Add SK012 validation rule (info-level)
4. **Phase 4**: Gradually migrate skills to explicit triggers

## Testing

```go
func TestParseTriggers(t *testing.T) {
    yaml := `---
name: go-code
triggers:
  - keywords: ["go", "golang"]
    weight: 0.8
  - pattern: "implement.*go"
    weight: 0.9
---`

    skill, err := parser.Parse(yaml)
    assert.NoError(t, err)
    assert.Len(t, skill.Frontmatter.Triggers, 2)
    assert.Equal(t, 0.8, skill.Frontmatter.Triggers[0].Weight)
}
```
