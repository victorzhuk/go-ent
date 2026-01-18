# Design: Enhanced Validation Error Messages

## Data Model Changes

### ValidationError Struct

```go
// Before
type ValidationError struct {
    Field   string
    Message string
}

// After
type ValidationError struct {
    Field      string
    Message    string
    Suggestion string  // NEW: How to fix the issue
    Example    string  // NEW: Example of correct usage
}
```

### ValidationWarning Struct

```go
// Extend similarly for consistency
type ValidationWarning struct {
    Field      string
    Message    string
    Suggestion string  // NEW
    Example    string  // NEW
}
```

## Implementation Strategy

### Phase 1: Extend Error Types

Update struct definitions in `internal/skill/validator.go`:
- Add `Suggestion` and `Example` fields
- Both fields optional (empty string if not provided)
- Backward compatible: existing code continues to work

### Phase 2: Update Validation Rules

For each validation rule in `internal/skill/rules.go`, add:

**SK001 (name-required)**:
```go
Suggestion: "Add a 'name' field to the frontmatter"
Example: `---
name: your-skill-name
---`
```

**SK002 (name-format)**:
```go
Suggestion: "Use lowercase letters, numbers, and hyphens only"
Example: "valid-skill-name-123"
```

**SK003 (description-required)**:
```go
Suggestion: "Add a 'description' field explaining what the skill does and when to use it"
Example: "Analyzes Go code for common issues. Use when reviewing Go files or debugging Go applications."
```

**SK004 (examples-section)**:
```go
Suggestion: "Add <examples> section with 3-5 diverse examples showing input/output pairs"
Example: `<examples>
  <example>
    <input>sample input</input>
    <output>expected output</output>
  </example>
</examples>`
```

**SK005 (role-section)**:
```go
Suggestion: "Add <role> section defining the skill's expertise and behavior"
Example: `<role>
You are an expert Go developer with 10+ years experience.
</role>`
```

**SK006 (instructions-section)**:
```go
Suggestion: "Add <instructions> section with clear, actionable steps"
Example: `<instructions>
1. Analyze the Go code structure
2. Identify anti-patterns
3. Suggest improvements
</instructions>`
```

**SK007 (constraints-section)**:
```go
Suggestion: "Add <constraints> section defining boundaries and limitations"
Example: `<constraints>
- Focus only on idiomatic Go
- Do not suggest external dependencies
</constraints>`
```

**SK008 (output-format)**:
```go
Suggestion: "Add <output_format> section specifying expected response structure"
Example: `<output_format>
Provide findings in a markdown table with columns: Issue, Location, Suggestion
</output_format>`
```

**SK009 (edge-cases)**:
```go
Suggestion: "Add <edge_cases> section handling unusual inputs"
Example: `<edge_cases>
If input is empty: Return validation error
If file is too large: Process first 10k lines only
</edge_cases>`
```

### Phase 3: Enhanced CLI Output

Update error display in `cmd/go-ent/skill_validate.go`:

```go
func formatValidationError(err ValidationError) string {
    var b strings.Builder

    fmt.Fprintf(&b, "❌ %s: %s\n", err.Field, err.Message)

    if err.Suggestion != "" {
        fmt.Fprintf(&b, "   💡 %s\n", err.Suggestion)
    }

    if err.Example != "" {
        fmt.Fprintf(&b, "\n   Example:\n")
        for _, line := range strings.Split(err.Example, "\n") {
            fmt.Fprintf(&b, "      %s\n", line)
        }
    }

    return b.String()
}
```

**Output example**:
```
❌ examples: Missing examples section
   💡 Add <examples> section with 3-5 diverse examples showing input/output pairs

   Example:
      <examples>
        <example>
          <input>sample input</input>
          <output>expected output</output>
        </example>
      </examples>
```

## Migration Path

**Backward Compatibility**: 100% compatible
- Empty `Suggestion` and `Example` fields work like before
- Existing validation rules continue functioning
- Enhanced rules provide better UX without breaking changes

**Rollout Strategy**:
1. Add fields to structs (non-breaking)
2. Update CLI formatter to display new fields (non-breaking)
3. Gradually enhance validation rules with suggestions
4. All rules enhanced by completion

## Performance Considerations

- **Memory**: Each error adds ~100-500 bytes for suggestion/example strings
- **Impact**: Negligible - validation runs once, errors are rare
- **Optimization**: Not needed - validation is not performance-critical

## Testing Strategy

**Unit Tests**:
```go
func TestValidationErrorWithSuggestion(t *testing.T) {
    err := ValidationError{
        Field:      "name",
        Message:    "Name is required",
        Suggestion: "Add a 'name' field to frontmatter",
        Example:    "name: my-skill",
    }

    output := formatValidationError(err)
    assert.Contains(t, output, "Add a 'name' field")
    assert.Contains(t, output, "name: my-skill")
}
```

**Integration Tests**:
- Validate skill with missing name → check suggestion appears
- Validate skill with invalid format → check example appears
- Validate skill with all issues → check all suggestions appear
