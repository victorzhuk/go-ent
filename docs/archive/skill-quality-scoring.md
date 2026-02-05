# Skill Quality Scoring Guide

## Overview

This guide explains the skill quality scoring system used to evaluate skill effectiveness, completeness, and readiness for production use. The scoring system is aligned with research findings on what makes skills effective.

## Scoring Breakdown

Quality scores are calculated across 5 categories, totaling **100 points maximum**:

| Category | Points | Weight |
|----------|--------|--------|
| Frontmatter | 20 | 20% |
| Structure | 20 | 20% |
| Content | 25 | 25% |
| Examples | 25 | 25% |
| Triggers | 15 | 15% |
| Conciseness | 15 | 15% |

## Category Details

### 1. Frontmatter (20 points)

Evaluates the presence and completeness of skill metadata.

#### Criteria (20 points total)

| Field | Points | Requirement |
|-------|--------|------------|
| Name | 5pts | Skill name present (non-empty) |
| Description | 5pts | Description present (non-empty) |
| Version | 5pts | Version present (non-empty) |
| Tags | 5pts | Tags array has at least one element |

#### What Makes a Good Frontmatter

**Best practices:**
- ✅ Use descriptive, concise names (e.g., "go-code", not "my-skill")
- ✅ Write clear descriptions that explain when the skill activates
- ✅ Include version tags following semantic versioning (e.g., "1.2.0")
- ✅ Add relevant tags for discoverability (e.g., ["go", "database", "testing"])

**Avoid:**
- ❌ Empty or missing required fields
- ❌ Vague descriptions like "A skill for doing stuff"
- ❌ Generic version numbers like "1.0" without semver format

---

### 2. Structure (20 points)

Evaluates the presence of required XML sections in the skill body.

#### Criteria (20 points total)

| Section | Points | Requirement |
|---------|--------|------------|
| `<role>` | 4pts | Role section with opening/closing tags |
| `<instructions>` | 4pts | Instructions section with opening/closing tags |
| `<constraints>` | 3pts | Constraints section with opening/closing tags |
| `<examples>` | 3pts | Examples section with opening/closing tags |
| `<output_format>` | 3pts | Output format section with opening/closing tags |
| `<edge_cases>` | 3pts | Edge cases section with opening/closing tags |

#### What Makes Good Structure

**Best practices:**
- ✅ Include all 6 sections for completeness (18-20 pts)
- ✅ Each section contains meaningful content (not empty tags)
- ✅ Use proper XML nesting and closing tags

**Avoid:**
- ❌ Missing sections reduces discoverability
- ❌ Empty sections waste space without value
- ❌ Unclosed XML tags cause parsing errors

---

### 3. Content (25 points)

Evaluates the quality and actionability of skill content.

#### Criteria (25 points total)

| Sub-category | Points | Assessment |
|--------------|--------|------------|
| Role Clarity | 0-8pts | Expertise level, domain specificity, behavioral description |
| Instructions | 0-9pts | Actionability, specificity, organization |
| Constraints | 0-8pts | Positive/negative rules, specificity |

#### Role Clarity (0-8 points)

Assesses the `<role>` section quality:

| Indicator | Points | Assessment |
|-----------|--------|------------|
| Expertise keywords | 3pts | Contains: "expert", "specialist", "architect", "engineer", "developer" |
| Domain specificity | 2pts | Mentions domain: "go", "golang", "python", "rust", "api", "database", "security" |
| Behavioral description | 3pts | Contains: "focus", "prioritize", "specialize", "ensure", "implement" |

**Examples:**

**Good (8 pts):**
```
<role>
Expert Go developer specializing in clean architecture and domain-driven design.
Focus on production-grade quality with emphasis on SOLID principles and maintainability.
Ensure patterns are practical and battle-tested.
</role>
```

**Needs improvement (3-5 pts):**
```
<role>
A skill for Go coding patterns.
</role>
```

#### Instructions (0-9 points)

Assesses the `<instructions>` section quality:

| Indicator | Points | Assessment |
|-----------|--------|------------|
| Actionability | 3pts | Contains imperative verbs: "use", "implement", "create", "add", "define", "handle", "check", "verify" |
| Specificity | 3pts | Contains specific conditions: "for", "with", "when", "if", "ensure", "require", "must" |
| Structure | 3pts | Organized with headers, 5+ non-empty lines, 1+ section headers |

**Examples:**

**Good (9 pts):**
```
<instructions>

## Bootstrap Pattern

Use this pattern for all Go services:

```go
func main() {
    if err := run(context.Background()); err != nil {
        slog.Error("fatal", "error", err)
        os.Exit(1)
    }
}
```

### Repository Pattern

Implement repository with private models and public entities.

### Error Handling

Always wrap errors with context using `%w` verb.
</instructions>
```

**Needs improvement (3-6 pts):**
```
<instructions>
Write clean code. Follow patterns. Handle errors.
</instructions>
```

#### Constraints (0-8 points)

Assesses the `<constraints>` section quality:

| Indicator | Points | Assessment |
|-----------|--------|------------|
| Positive rules | 3pts | Contains: "include", "must" |
| Negative rules | 3pts | Contains: "exclude", "don't", "avoid", "never" |
| Specificity | 2pts | Contains detailed criteria: "bound to", "follow", "ensure", "verify", "use" |

**Examples:**

**Good (8 pts):**
```
<constraints>
- Include clean architecture patterns
- Exclude global state and singletons
- Must follow SOLID principles
- Bound to transport/usecase/domain layers
- Ensure proper error handling
- Verify context propagation
</constraints>
```

**Needs improvement (3-5 pts):**
```
<constraints>
Be good with code.
</constraints>
```

---

### 4. Examples (25 points)

Evaluates the quality, diversity, and completeness of examples.

#### Criteria (25 points total)

| Sub-category | Points | Assessment |
|--------------|--------|------------|
| Count | 0-10pts | Number of examples (0=0, 1=3, 2=6, 3-5=10, >5=8) |
| Diversity | 0-8pts | Variety of input types, behaviors, and scenarios |
| Edge Cases | 0-4pts | Coverage of edge cases (2pts per edge case, max 4pts) |
| Format | 0-3pts | Proper input/output pairs with XML structure |

#### Count Scoring

| Examples | Points | Notes |
|----------|--------|-------|
| 0 | 0pts | No examples - skill is incomplete |
| 1 | 3pts | Minimal - add more examples |
| 2 | 6pts | Below threshold - need diversity |
| 3-5 | 10pts | Optimal range - ideal number of examples |
| >5 | 8pts | Too many - consider consolidating |

#### Diversity Scoring (0-8 points)

Evaluates variety in examples:

| Factor | Points | Assessment |
|--------|--------|------------|
| Input type variety | Up to 8pts | Different input types: code, config, API calls, database queries, etc. |
| Behavior variety | Included in input type assessment | Success paths, error cases, edge scenarios |

**Scoring:**
- 3+ different input types = 8pts (excellent)
- 2 different input types = 5pts (good)
- 1 input type only = 2pts (needs improvement)

**Examples:**

**High diversity (8pts):**
- Code refactoring example
- Database query example
- API endpoint example
- Configuration setup example

**Low diversity (2pts):**
- Three code examples showing similar patterns

#### Edge Case Scoring (0-4 points)

Evaluates coverage of edge cases in `<edge_cases>` section:

| Edge case type | Points |
|----------------|--------|
| Empty/null inputs | 2pts |
| Error conditions | 2pts |
| Boundary conditions | 2pts |
| Invalid data | 2pts |

**Scoring:**
- 2 edge cases = 4pts (full coverage)
- 1 edge case = 2pts (partial coverage)
- 0 edge cases = 0pts (no coverage)

**Common edge cases:**
- Empty strings, nil values
- Zero/negative numbers
- Maximum values
- Malformed input
- Network timeouts
- Database connection failures

#### Format Scoring (0-3 points)

Evaluates proper XML structure in examples:

| Requirement | Points |
|-------------|--------|
| All examples have `<input>` and `<output>` tags | 2pts |
| Tags are properly nested in `<example>` wrapper | 1pt |

**Good format:**
```
<examples>
<example>
<input>Refactor main() to use bootstrap pattern</input>
<output>```go
func main() {
    if err := run(context.Background()); err != nil {
        slog.Error("fatal", "error", err)
        os.Exit(1)
    }
}
```</output>
</example>
</examples>
```

---

### 5. Triggers (15 points)

Evaluates presence and quality of explicit trigger definitions.

#### Criteria (15 points total)

| Factor | Points | Assessment |
|--------|--------|------------|
| Base score | 10pts | Has explicit triggers (v2 format) |
| Weights | 3pts | Triggers have defined weights |
| Diversity | 2pts | Multiple trigger types (keywords, patterns, file patterns) |
| Fallback | 5pts max | Legacy description-based triggers only |

#### Explicit Triggers (v2 format)

**New format** (max 15 points):

```yaml
triggers:
  - keywords: ["go code", "golang"]
    weight: 0.8
  - patterns: [".*go.*implementation", "implement.*go"]
    weight: 0.9
  - file_patterns: ["**/*.go"]
    weight: 0.7
```

**Scoring breakdown:**
- 10pts: Base score for having explicit triggers
- 3pts: Triggers include `weight` fields
- 2pts: Triggers use multiple types (keywords + patterns, or keywords + file_patterns)

#### Legacy Triggers (fallback - 5 points max)

**Old format** (5 points only):

```yaml
triggers:
  - "go code"
  - "golang"
  - "implement go"
```

**Scoring:**
- 5pts: Has triggers (even if description-based only)
- 0pts: No triggers at all

#### Best Practices

**✅ Use explicit triggers (v2):**
- Higher precision matching
- Weighted relevance scoring
- File pattern matching
- Better activation control

**✅ Diversify trigger types:**
- Keywords for common phrases
- Patterns for complex queries
- File patterns for context-specific activation

**❌ Avoid:**
- Description-based triggers only (lower precision)
- Missing weights (default to 1.0)
- Single trigger type only

---

### 6. Conciseness (15 points)

Evaluates skill length to prevent attention dilution in AI models.

#### Criteria (0-15 points)

| Token Range | Points | Assessment |
|-------------|--------|------------|
| < 3,000 tokens | 15pts | Optimal length - full marks |
| 3,000 - 5,000 tokens | 10pts | Acceptable - minor penalty |
| 5,000 - 8,000 tokens | 5pts | Verbose - significant penalty |
| > 8,000 tokens | 0pts | Too verbose - no credit |

#### Token Estimation

Tokens are approximated as: `word_count × 1.3`

**Example:**
- 1000 words ≈ 1300 tokens
- 2000 words ≈ 2600 tokens
- 4000 words ≈ 5200 tokens (approaches verbose threshold)

#### Best Practices

**✅ Maintain conciseness:**
- Focus on core patterns (don't include every edge case)
- Move detailed examples to separate reference docs
- Link to external documentation for extended explanations
- Use bullet points and tables for dense information

**❌ Avoid:**
- Repeating information across sections
- Including complete API documentation inline
- Long tutorial-style explanations
- Copying code from external sources unchanged

**Optimization Tips:**
1. **Extract reference material**: Move detailed guides, tutorials, and docs to `references/` directory and link with "See: `references/database-patterns.md`"
2. **Use hierarchical structure**: Overview → Key patterns → Reference links
3. **Leverage examples**: Use `<examples>` to show usage rather than explaining at length
4. **Remove redundancy**: Delete duplicate information between sections

---

## Score Interpretation

### Score Ranges

| Score Range | Interpretation | Action |
|-------------|----------------|---------|
| 90-100 | **Excellent** | Production-ready, minimal improvements needed |
| 80-89 | **Good** | Minor improvements possible |
| 70-79 | **Acceptable** | Moderate improvements needed |
| 60-69 | **Needs Improvement** | Several areas require attention |
| 40-59 | **Poor** | Significant work needed |
| 0-39 | **Inadequate** | Skill needs major rework |

### Category Analysis

**Prioritize improvements by lowest category scores:**

1. **Score < 12 out of 20 (60%)**: Major gap - address immediately
2. **Score < 15 out of 25 (60%)**: Significant gap - needs work
3. **Score < 9 out of 15 (60%)**: Important gap - add content
4. **Score < 9 out of 15 (60%)**: Important gap - add content

### Category Improvement Guide

#### Structure Score < 12

**Common issues:**
- Missing `<output_format>` or `<edge_cases>` sections
- Empty sections without content
- Unclosed XML tags

**Actions:**
1. Add missing sections with appropriate content
2. Ensure each section has meaningful content
3. Verify all XML tags are properly closed

#### Content Score < 15

**Role Clarity < 5:**
- Define expertise level and domain
- Describe behavioral focus
- Add specific technical domain (e.g., "PostgreSQL", "REST API")

**Instructions < 5:**
- Use imperative verbs ("implement", "create")
- Add specific conditions and examples
- Organize with headers and bullet points

**Constraints < 5:**
- Add positive rules (what to include)
- Add negative rules (what to exclude)
- Specify layer boundaries (e.g., "bound to transport layer")

#### Examples Score < 15

**Count < 7:** Add 2-3 more examples with diversity
- Mix input types (code, config, API)
- Include error cases
- Add edge cases to `<edge_cases>` section

**Diversity < 5:** Add variety in examples
- Different operation types (CRUD, query, transformation)
- Different complexity levels (simple, complex)
- Different scenarios (success, failure, partial)

**Edge Cases < 2:** Expand `<edge_cases>` section
- Add common edge cases from production
- Include handling strategies
- Reference to error handling patterns

#### Triggers Score < 9

**No triggers:** Add explicit triggers
- Start with keywords (3-5 common phrases)
- Add 1-2 regex patterns for complex matching
- Define weights (0.6-0.9 based on relevance)
- Consider file patterns (e.g., `*.go`)

**No weights:** Add weight fields
- Weight most relevant triggers higher (0.9)
- Weight general triggers lower (0.6)
- Use fractional weights for fine-tuning

**No diversity:** Add multiple trigger types
- Combine keywords + patterns
- Add file patterns for context
- Each type adds 2pts (max 4pts total with weights)

#### Conciseness Score < 9

**Verbose (>5k tokens):** Reduce content length
- Move detailed explanations to references/
- Use concise descriptions
- Replace text with tables/bullets
- Remove redundant information

**Very verbose (>8k tokens):** Major reduction needed
- Split skill into specialized sub-skills
- Focus on one narrow domain
- Reference external documentation heavily

---

## Optimization Tips

### Quick Wins (1-2 hours)

1. **Add missing sections** (2-3 points each)
   - Add `<output_format>` and `<edge_cases>` if missing
   - Fill with brief, actionable content

2. **Add explicit triggers** (5-15 points)
   - Convert description-based triggers to explicit format
   - Add 3-5 keywords with weights

3. **Add 1-2 examples** (3-7 points each)
   - Focus on different input types
   - Include error handling cases

### Medium Efforts (3-5 hours)

1. **Improve content clarity** (5-10 points)
   - Rewrite role to include expertise and domain
   - Structure instructions with headers and imperatives
   - Add specific positive/negative constraints

2. **Increase example diversity** (3-8 points)
   - Add examples for different operations
   - Include edge cases in examples or `<edge_cases>`
   - Ensure input/output variety

3. **Extract reference material** (5-15 conciseness points)
   - Move detailed guides to `references/`
   - Replace with brief overviews and links

### Long-term Projects (1-2 days)

1. **Complete skill overhaul** (10-30 points)
   - Restructure missing sections
   - Rewrite content for clarity
   - Add comprehensive examples (5-8 total)
   - Create detailed triggers with multiple types
   - Optimize for conciseness

2. **Create specialized sub-skills** (if applicable)
   - Split broad skills into focused domains
   - Each sub-skill scores independently
   - Improves precision and relevance

---

## Verification

### Run Analysis

To analyze all skills:

```bash
# Console output with visual bars
go-ent skill analyze

# JSON export for CI/CD
go-ent skill analyze --json > quality-report.json

# CSV export for spreadsheets
go-ent skill analyze --csv > quality-scores.csv
```

### Validate Individual Skill

To check a specific skill:

```bash
# Run validate-skill command
cmd/validate-skill/main.go plugins/go-ent/skills/go/go-code/SKILL.md
```

### Continuous Improvement

1. **Analyze regularly**: Run `skill analyze` weekly or after significant changes
2. **Track progress**: Monitor score improvements over time
3. **Set targets**: Aim for ≥80 average across all skills
4. **Prioritize**: Focus on lowest-scoring skills first

---

## Appendix

### Token Count Reference

| Word Count | Estimated Tokens | Conciseness Score |
|-------------|------------------|------------------|
| 1000 | 1,300 | 15pts |
| 1500 | 1,950 | 15pts |
| 2000 | 2,600 | 15pts |
| 2308 | 3,000 | 10pts (threshold) |
| 2500 | 3,250 | 10pts |
| 3000 | 3,900 | 10pts |
| 3846 | 5,000 | 5pts (threshold) |
| 4000 | 5,200 | 5pts |
| 5000 | 6,500 | 5pts |
| 6154 | 8,000 | 0pts (threshold) |
| 7000 | 9,100 | 0pts |
| 8000 | 10,400 | 0pts |
| 10000 | 13,000 | 0pts |

### Quality Score Calculator

To estimate score improvement impact:

```
Current Score + (Target Score - Current Score) = New Target Score

Example: Score 65 → Add 10pts from triggers = Score 75
         Score 75 → Add 8pts from examples = Score 83
```

### Common Scoring Patterns

| Pattern | Likely Cause | Solution |
|---------|---------------|----------|
| High structure, low examples | Missing examples or poor diversity | Add 3-5 diverse examples |
| High examples, low structure | Missing sections | Add `<output_format>` or `<edge_cases>` |
| High content, low triggers | No explicit triggers defined | Add explicit triggers with weights |
| Low conciseness (<9) | Verbose skill >8k tokens | Extract references, shorten content |
| All categories ~60% | Multiple areas need attention | Follow prioritized improvement guide |

---

**Version**: 2.0.0  
**Last Updated**: 2025-01-19  
**Related**: Skill Validation Rules, Skill Migration Proposal
