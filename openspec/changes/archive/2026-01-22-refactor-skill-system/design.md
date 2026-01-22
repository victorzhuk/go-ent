# Design: Skill System Refactoring

## Architecture Overview

The skill system refactoring extends the existing architecture with validation, quality scoring, and v2 format support while maintaining backward compatibility.

```
┌─────────────────────────────────────────────────────────────┐
│                       MCP Server                             │
├─────────────────────────────────────────────────────────────┤
│  MCP Tools                                                   │
│  ├── skill_list (existing)                                  │
│  ├── skill_info (existing)                                  │
│  ├── skill_validate (NEW)                                   │
│  └── skill_quality (NEW)                                    │
└────────────┬────────────────────────────────────────────────┘
             │
             ├─> Registry
             │   ├── Load() - scan directories
             │   ├── Register() - runtime skills
             │   ├── MatchForContext() - activation
             │   ├── ValidateSkill() - NEW
             │   ├── ValidateAll() - NEW
             │   └── GetQualityReport() - NEW
             │
             ├─> Parser
             │   ├── ParseSkillFile() - extract metadata
             │   ├── extractFrontmatter() - YAML parsing
             │   ├── extractTriggers() - "Auto-activates for:"
             │   ├── detectVersion() - NEW (v1 vs v2)
             │   └── parseFrontmatterV2() - NEW
             │
             ├─> Validator (NEW)
             │   ├── Validate(meta, content) - run rules
             │   └── Rules:
             │       ├── validateFrontmatter
             │       ├── validateVersion
             │       ├── validateXMLTags
             │       ├── validateRoleSection
             │       ├── validateInstructionsSection
             │       ├── validateExamples
             │       ├── validateConstraints
             │       ├── validateEdgeCases
             │       └── validateOutputFormat
             │
             └─> QualityScorer (NEW)
                 └── Score(meta, content) - compute 0-100
```

## Data Model

### Extended SkillMeta

**Current** (`internal/skill/parser.go:13-18`):
```go
type SkillMeta struct {
    Name        string
    Description string
    Triggers    []string
    FilePath    string
}
```

**Extended** (maintains backward compatibility):
```go
type SkillMeta struct {
    // Existing fields
    Name        string
    Description string
    Triggers    []string
    FilePath    string

    // New fields (v2 format)
    Version        string   // Semantic version (e.g., "2.0.0")
    Author         string   // Attribution (e.g., "go-ent")
    Tags           []string // Categorization (e.g., ["go", "code"])
    AllowedTools   []string // Security boundary (optional)

    // Computed fields
    StructureVersion string  // "v1" or "v2"
    QualityScore     float64 // 0-100
}
```

### SkillValidationContext

**New type** (`internal/skill/validator.go`):
```go
type ValidationContext struct {
    FilePath    string
    Content     string
    Lines       []string
    Meta        *SkillMeta
    Strict      bool
}
```

### ValidationResult

**New type** (`internal/skill/validator.go`):
```go
type ValidationResult struct {
    Valid   bool
    Issues  []ValidationIssue
    Score   float64
}

type ValidationIssue struct {
    Rule     string
    Severity Severity // error, warning, info
    Message  string
    Line     int
    Column   int
}

type Severity string

const (
    SeverityError   Severity = "error"
    SeverityWarning Severity = "warning"
    SeverityInfo    Severity = "info"
)
```

## Component Design

### 1. Parser Extension

**File**: `internal/skill/parser.go`

#### New Methods

```go
// detectVersion checks content for v2 markers (<role>, <instructions>)
func (p *Parser) detectVersion(content string) string

// parseFrontmatterV2 parses extended frontmatter fields
func (p *Parser) parseFrontmatterV2(frontmatter string) (*SkillMetaV2, error)

type SkillMetaV2 struct {
    Name         string   `yaml:"name"`
    Description  string   `yaml:"description"`
    Version      string   `yaml:"version"`
    Author       string   `yaml:"author"`
    Tags         []string `yaml:"tags"`
    AllowedTools []string `yaml:"allowed-tools"`
}
```

#### Modified ParseSkillFile Flow

```
1. Open file
2. Extract frontmatter (existing)
3. Parse base fields (name, description)
4. Read full content
5. Detect version (NEW)
   ├─> If contains "<role>" or "<instructions>" → v2
   └─> Else → v1
6. If v2: Parse extended frontmatter (NEW)
7. Extract triggers (existing)
8. Return SkillMeta with all fields
```

### 2. Validator

**New File**: `internal/skill/validator.go`

#### Architecture

Follows existing `internal/spec/validator.go` pattern:

```go
type Validator struct {
    rules []ValidationRule
}

type ValidationRule func(ctx *ValidationContext) []ValidationIssue

func NewValidator() *Validator {
    return &Validator{
        rules: []ValidationRule{
            validateFrontmatter,
            validateVersion,
            validateXMLTags,
            validateRoleSection,
            validateInstructionsSection,
            validateExamples,
            validateConstraints,
            validateEdgeCases,
            validateOutputFormat,
        },
    }
}

func (v *Validator) Validate(meta *SkillMeta, content string) *ValidationResult {
    ctx := &ValidationContext{
        FilePath: meta.FilePath,
        Content:  content,
        Lines:    strings.Split(content, "\n"),
        Meta:     meta,
    }

    var issues []ValidationIssue
    for _, rule := range v.rules {
        issues = append(issues, rule(ctx)...)
    }

    // Compute quality score
    scorer := NewQualityScorer()
    score := scorer.Score(meta, content)

    return &ValidationResult{
        Valid:  !hasErrors(issues),
        Issues: issues,
        Score:  score,
    }
}
```

### 3. Validation Rules

**New File**: `internal/skill/rules.go`

Each rule is self-contained and testable:

```go
// validateFrontmatter checks required fields
func validateFrontmatter(ctx *ValidationContext) []ValidationIssue {
    var issues []ValidationIssue

    if ctx.Meta.Name == "" {
        issues = append(issues, ValidationIssue{
            Rule:     "frontmatter",
            Severity: SeverityError,
            Message:  "missing required field: name",
            Line:     1,
        })
    }

    if ctx.Meta.Description == "" {
        issues = append(issues, ValidationIssue{
            Rule:     "frontmatter",
            Severity: SeverityError,
            Message:  "missing required field: description",
            Line:     1,
        })
    }

    return issues
}

// validateVersion checks semantic version format
func validateVersion(ctx *ValidationContext) []ValidationIssue {
    if ctx.Meta.Version == "" {
        return nil // optional field
    }

    if !semverRegex.MatchString(ctx.Meta.Version) {
        return []ValidationIssue{{
            Rule:     "version",
            Severity: SeverityError,
            Message:  fmt.Sprintf("invalid semantic version: %s", ctx.Meta.Version),
            Line:     findLineNumber(ctx.Lines, "version:"),
        }}
    }

    return nil
}

// validateXMLTags checks well-formed XML sections
func validateXMLTags(ctx *ValidationContext) []ValidationIssue {
    tags := []string{"role", "instructions", "constraints", "edge_cases", "examples", "output_format"}
    var issues []ValidationIssue

    for _, tag := range tags {
        openTag := "<" + tag + ">"
        closeTag := "</" + tag + ">"

        openCount := strings.Count(ctx.Content, openTag)
        closeCount := strings.Count(ctx.Content, closeTag)

        if openCount != closeCount {
            issues = append(issues, ValidationIssue{
                Rule:     "xml-tags",
                Severity: SeverityError,
                Message:  fmt.Sprintf("unbalanced <%s> tags: %d open, %d close", tag, openCount, closeCount),
            })
        }
    }

    return issues
}

// Additional rules follow similar pattern...
```

### 4. Quality Scorer

**New File**: `internal/skill/scorer.go`

```go
type QualityScorer struct{}

func NewQualityScorer() *QualityScorer {
    return &QualityScorer{}
}

func (s *QualityScorer) Score(meta *SkillMeta, content string) float64 {
    frontmatterScore := s.scoreFrontmatter(meta)     // 20 points
    structureScore := s.scoreStructure(content)      // 30 points
    contentScore := s.scoreContent(content)          // 30 points
    triggerScore := s.scoreTriggers(meta)            // 20 points

    return frontmatterScore + structureScore + contentScore + triggerScore
}

func (s *QualityScorer) scoreFrontmatter(meta *SkillMeta) float64 {
    score := 0.0

    if meta.Name != "" {
        score += 5.0
    }
    if meta.Description != "" {
        score += 5.0
    }
    if meta.Version != "" {
        score += 5.0
    }
    if len(meta.Tags) > 0 {
        score += 5.0
    }

    return score // max 20
}

func (s *QualityScorer) scoreStructure(content string) float64 {
    score := 0.0

    requiredSections := []string{"<role>", "<instructions>", "<examples>"}
    for _, section := range requiredSections {
        if strings.Contains(content, section) {
            score += 10.0
        }
    }

    return score // max 30
}

func (s *QualityScorer) scoreContent(content string) float64 {
    score := 0.0

    // Check for examples with input/output
    exampleCount := strings.Count(content, "<example>")
    if exampleCount >= 2 {
        score += 15.0
    } else if exampleCount == 1 {
        score += 10.0
    }

    // Check for edge cases
    if strings.Contains(content, "<edge_cases>") {
        score += 15.0
    }

    return score // max 30
}

func (s *QualityScorer) scoreTriggers(meta *SkillMeta) float64 {
    if len(meta.Triggers) == 0 {
        return 0.0
    }

    if len(meta.Triggers) >= 3 {
        return 20.0
    }

    return float64(len(meta.Triggers)) * 6.67 // max 20
}
```

### 5. Registry Extension

**File**: `internal/skill/registry.go`

#### New Fields

```go
type Registry struct {
    skills        []SkillMeta
    runtimeSkills map[string]domain.Skill
    parser        *Parser
    validator     *Validator  // NEW
    scorer        *QualityScorer  // NEW
}
```

#### New Methods

```go
func (r *Registry) ValidateSkill(name string) (*ValidationResult, error) {
    meta, err := r.Get(name)
    if err != nil {
        return nil, err
    }

    content, err := os.ReadFile(meta.FilePath)
    if err != nil {
        return nil, fmt.Errorf("read file: %w", err)
    }

    return r.validator.Validate(meta, string(content)), nil
}

func (r *Registry) ValidateAll() (*ValidationResult, error) {
    var allIssues []ValidationIssue
    totalScore := 0.0

    for _, meta := range r.skills {
        result, err := r.ValidateSkill(meta.Name)
        if err != nil {
            return nil, err
        }
        allIssues = append(allIssues, result.Issues...)
        totalScore += result.Score
    }

    avgScore := totalScore / float64(len(r.skills))

    return &ValidationResult{
        Valid:  !hasErrors(allIssues),
        Issues: allIssues,
        Score:  avgScore,
    }, nil
}

func (r *Registry) GetQualityReport() map[string]float64 {
    report := make(map[string]float64)

    for _, meta := range r.skills {
        report[meta.Name] = meta.QualityScore
    }

    return report
}
```

### 6. MCP Tools

#### skill_validate Tool

**New File**: `internal/mcp/tools/skill_validate.go`

```go
type SkillValidateInput struct {
    Name   string `json:"name,omitempty"`
    Strict bool   `json:"strict,omitempty"`
}

type SkillValidateOutput struct {
    Valid   bool                `json:"valid"`
    Score   float64             `json:"score"`
    Issues  []ValidationIssue   `json:"issues"`
}

func registerSkillValidate(s *server.MCPServer, registry *skill.Registry) {
    s.AddTool(mcp.Tool{
        Name: "skill_validate",
        Description: "Validate skill structure and content quality",
        InputSchema: generateSchema(SkillValidateInput{}),
    }, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
        var input SkillValidateInput
        if err := mapToStruct(args, &input); err != nil {
            return nil, err
        }

        var result *skill.ValidationResult
        var err error

        if input.Name != "" {
            result, err = registry.ValidateSkill(input.Name)
        } else {
            result, err = registry.ValidateAll()
        }

        if err != nil {
            return nil, err
        }

        output := SkillValidateOutput{
            Valid:  result.Valid,
            Score:  result.Score,
            Issues: result.Issues,
        }

        return &mcp.CallToolResult{
            Content: []mcp.Content{{
                Type: "text",
                Text: formatValidationOutput(output),
            }},
        }, nil
    })
}
```

#### skill_quality Tool

**New File**: `internal/mcp/tools/skill_quality.go`

```go
type SkillQualityInput struct {
    Threshold float64 `json:"threshold,omitempty"`
}

type SkillQualityOutput struct {
    Skills      []SkillScore `json:"skills"`
    AvgScore    float64      `json:"avg_score"`
    BelowThresh []string     `json:"below_threshold"`
}

type SkillScore struct {
    Name  string  `json:"name"`
    Score float64 `json:"score"`
}

func registerSkillQuality(s *server.MCPServer, registry *skill.Registry) {
    s.AddTool(mcp.Tool{
        Name: "skill_quality",
        Description: "Get quality scores for all skills",
        InputSchema: generateSchema(SkillQualityInput{}),
    }, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
        var input SkillQualityInput
        if err := mapToStruct(args, &input); err != nil {
            return nil, err
        }

        report := registry.GetQualityReport()

        var skills []SkillScore
        var total float64
        var belowThresh []string

        for name, score := range report {
            skills = append(skills, SkillScore{Name: name, Score: score})
            total += score

            if input.Threshold > 0 && score < input.Threshold {
                belowThresh = append(belowThresh, name)
            }
        }

        output := SkillQualityOutput{
            Skills:      skills,
            AvgScore:    total / float64(len(skills)),
            BelowThresh: belowThresh,
        }

        return &mcp.CallToolResult{
            Content: []mcp.Content{{
                Type: "text",
                Text: formatQualityOutput(output),
            }},
        }, nil
    })
}
```

## Migration Strategy

### Phase 1: Infrastructure (No User-Facing Changes)

**Goal**: Build foundation without breaking existing functionality

1. Extend `SkillMeta` struct with new fields
2. Update parser to detect version and parse v2 frontmatter
3. Create validator, scorer, validation rules
4. Extend registry with validation methods
5. Add MCP tools

**Test**: All existing skills still work, `go test ./internal/skill/...` passes

### Phase 2: Template Creation

**Goal**: Create reference implementation

1. Refactor `go-code` to v2 format:
   - Add version: "2.0.0"
   - Define `<role>`: Expert Go developer with clean architecture focus
   - Convert sections to `<instructions>`
   - Extract constraints from CLAUDE.md
   - Add 2-3 `<examples>` with input/output
   - Define `<edge_cases>`: delegate to go-test, go-arch, go-perf
   - Specify `<output_format>`
2. Validate: `skill_validate name=go-code strict=true`
3. Ensure quality score >= 90

**Test**: go-code validates, original content preserved, new structure added

### Phase 3: Migration

**Goal**: Migrate remaining 13 skills following template

**Process per skill**:
1. Read current content
2. Use go-code as structural reference
3. Define role appropriate to skill domain
4. Convert content to XML-tagged sections
5. Add missing sections (examples, edge cases)
6. Validate and iterate until score >= 80

**Test**: Each skill validates before moving to next

## Backward Compatibility Design

### Version Detection

```go
func (p *Parser) detectVersion(content string) string {
    if strings.Contains(content, "<role>") ||
       strings.Contains(content, "<instructions>") {
        return "v2"
    }
    return "v1"
}
```

### Format Support

**During migration**:
- Parser handles both v1 and v2 skills
- Registry loads both formats
- Validation only runs on v2 skills (v1 gets warning)
- Quality scoring works for both (v1 scores lower due to missing structure)

**After migration**:
- All skills are v2
- Validation runs on all skills
- Can deprecate v1 support in future version

## Performance Considerations

### Validation Cost

- Validation runs on-demand via MCP tools
- Not executed during skill loading (keeps startup fast)
- Quality scoring is cached in `SkillMeta.QualityScore`

### Memory

- Extended `SkillMeta` adds ~100 bytes per skill (14 skills = ~1.4KB)
- Negligible impact

## Security Considerations

### AllowedTools Field

Future capability to restrict which MCP tools a skill can invoke:

```yaml
allowed-tools: [read, write, bash]
```

Not enforced in this phase, but field added for future use.

### Content Validation

- XML parsing uses standard library (safe)
- No arbitrary code execution
- File paths validated before reading

## Testing Strategy

### Unit Tests

- Parser: v1 detection, v2 detection, frontmatter parsing
- Validator: each rule independently
- Scorer: each component (frontmatter, structure, content, triggers)
- Registry: validation methods

### Integration Tests

- Load mixed v1/v2 skills
- Validate all skills
- Generate quality report
- MCP tools end-to-end

### Acceptance Tests

- go-code achieves score >= 90
- All 14 skills validate without errors
- Average score >= 80
- `make skill-validate` passes

## File Organization

```
internal/skill/
├── parser.go         # Extended for v2
├── registry.go       # Extended with validation
├── validator.go      # NEW: validation engine
├── rules.go          # NEW: validation rules
├── scorer.go         # NEW: quality scoring
├── parser_test.go
├── validator_test.go
└── scorer_test.go

internal/mcp/tools/
├── skill_validate.go # NEW: validation MCP tool
├── skill_quality.go  # NEW: quality MCP tool
└── register.go       # Updated to register new tools
```

## Dependencies

**No new external dependencies** - uses only standard library and existing project dependencies:
- `gopkg.in/yaml.v3` (already used)
- `github.com/google/uuid` (already used)
- Standard library: `strings`, `regexp`, `os`, `fmt`

## Rollback Plan

If issues arise:
1. Revert parser to detect and skip v2 validation
2. Skills remain functional (backward compatible)
3. Can fix issues and re-enable validation
4. No data loss (skills are text files in git)
