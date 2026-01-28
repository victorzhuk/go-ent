# Spec: Research-Aligned Quality Scoring

## ADDED Requirements

### REQ-SCORE-001: Quality score reflects research priorities

Skill quality score must align with research findings on effective prompt engineering.

#### Scenario: Optimal skill structure
**WHEN** skill has all XML sections (role, instructions, constraints, examples, output_format, edge_cases)
**THEN** structure score is 20/20

#### Scenario: Missing optional sections
**WHEN** skill is missing edge_cases and output_format sections
**THEN** structure score is 14/20 (missing 6 points)

### REQ-SCORE-002: Examples scoring rewards quality and diversity

Examples score must prioritize count, diversity, and edge cases per research.

#### Scenario: Optimal example count (3-5)
**WHEN** skill has 4 diverse examples with 1 edge case
**THEN** examples score is close to maximum (23-25/25)

#### Scenario: Too few examples
**WHEN** skill has only 1 example
**THEN** examples score is low (≤12/25)

#### Scenario: Diverse examples with edge cases
**WHEN** skill has 5 examples covering different input types and 2 edge cases
**THEN** examples score is maximum (25/25)

### REQ-SCORE-003: Conciseness scoring prevents attention dilution

Conciseness score must penalize verbose skills per research on context window limits.

#### Scenario: Ideal skill length (<3k tokens)
**WHEN** skill body is 2500 tokens
**THEN** conciseness score is 15/15

#### Scenario: Acceptable length (3-5k tokens)
**WHEN** skill body is 4000 tokens
**THEN** conciseness score is 10/15

#### Scenario: Warning length (5-8k tokens)
**WHEN** skill body is 6000 tokens
**THEN** conciseness score is 5/15
**AND** warning message suggests reducing content

#### Scenario: Critical length (>8k tokens)
**WHEN** skill body is 10000 tokens
**THEN** conciseness score is 0/15
**AND** critical warning about attention dilution

### REQ-SCORE-004: Triggers scoring favors explicit definitions

Triggers score must reward explicit trigger definitions over description-based.

#### Scenario: Explicit triggers with weights
**WHEN** skill has explicit triggers section with weights and multiple types
**THEN** triggers score is maximum (15/15)

#### Scenario: Explicit triggers without weights
**WHEN** skill has explicit triggers without weights
**THEN** triggers score is 12/15

#### Scenario: Description-based triggers only
**WHEN** skill uses description-based trigger extraction only
**THEN** triggers score is maximum 5/15

### REQ-SCORE-005: Content scoring evaluates quality not just presence

Content score must evaluate quality of role, instructions, and constraints.

#### Scenario: High-quality role definition
**WHEN** role defines expertise level, domain focus, and behavioral guidelines
**THEN** role clarity score is 8/8

#### Scenario: Vague role definition
**WHEN** role is generic without specifics
**THEN** role clarity score is ≤3/8

#### Scenario: Actionable instructions
**WHEN** instructions use imperative verbs, numbered steps, and concrete actions
**THEN** instruction quality score is 8-9/9

#### Scenario: Vague instructions
**WHEN** instructions are abstract or unclear
**THEN** instruction quality score is ≤4/9

### REQ-SCORE-006: Total score combines all categories correctly

Total score must be sum of all category scores with proper validation.

#### Scenario: Score calculation
**WHEN** calculating total score
**THEN** total equals structure + content + examples + triggers + conciseness
**AND** total is between 0 and 100
**AND** each category score is within its maximum

### REQ-SCORE-007: CLI displays scores clearly

CLI must show score breakdown with visual indicators.

#### Scenario: Score display format
**WHEN** displaying skill quality score
**THEN** CLI shows total score (X/100)
**AND** CLI shows breakdown for each category with progress bars
**AND** CLI shows recommendations for improvement

#### Scenario: Low score warnings
**WHEN** category score is <50% of maximum
**THEN** CLI highlights category with warning indicator
**AND** CLI provides specific recommendation for improvement

## MODIFIED Requirements

### REQ-SCORE-008: Updated score ranges

**Old behavior**: Four categories (Frontmatter 20, Structure 30, Content 30, Triggers 20)
**New behavior**: Five categories (Structure 20, Content 25, Examples 25, Triggers 15, Conciseness 15)

**Reason**: Research shows examples and conciseness are critical quality factors

## REMOVED Requirements

None - old scoring is replaced, not removed from history.
