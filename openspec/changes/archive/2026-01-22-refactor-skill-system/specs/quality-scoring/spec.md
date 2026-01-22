# Quality Scoring Capability

## ADDED Requirements

### REQ-SCORE-001: Quality Score Computation

The system shall compute quality scores (0-100) based on research-backed rubric.

#### Scenario: Score Complete V2 Skill
**WHEN** a skill has all frontmatter fields, all XML sections, 3 examples, and edge cases
**THEN** quality score is 100

#### Scenario: Score V1 Skill
**WHEN** a v1 skill with no XML structure is scored
**THEN** quality score is <= 40 (missing structure and content quality)

#### Scenario: Score Partial V2 Skill
**WHEN** a skill has role and instructions but no examples
**THEN** quality score is approximately 50-60

### REQ-SCORE-002: Frontmatter Scoring

The system shall allocate 20 points for frontmatter completeness.

#### Scenario: Complete Frontmatter
**WHEN** skill has name, description, version, and tags
**THEN** frontmatter score is 20

#### Scenario: Minimal Frontmatter
**WHEN** skill has only name and description
**THEN** frontmatter score is 10

### REQ-SCORE-003: Structure Scoring

The system shall allocate 30 points for XML structure compliance.

#### Scenario: All Required Sections
**WHEN** skill has `<role>`, `<instructions>`, and `<examples>` tags
**THEN** structure score is 30

#### Scenario: Partial Structure
**WHEN** skill has `<role>` and `<instructions>` but no `<examples>`
**THEN** structure score is 20

### REQ-SCORE-004: Content Scoring

The system shall allocate 30 points for content quality (examples and edge cases).

#### Scenario: Excellent Content
**WHEN** skill has >= 2 examples with input/output and edge cases section
**THEN** content score is 30

#### Scenario: Good Content
**WHEN** skill has 1 example but no edge cases
**THEN** content score is 10

### REQ-SCORE-005: Trigger Scoring

The system shall allocate 20 points for auto-activation trigger clarity.

#### Scenario: Multiple Triggers
**WHEN** skill has >= 3 trigger keywords extracted
**THEN** trigger score is 20

#### Scenario: Few Triggers
**WHEN** skill has 1 trigger keyword
**THEN** trigger score is approximately 6.67

### REQ-SCORE-006: Quality Report Generation

The system shall generate quality reports for all skills.

#### Scenario: Generate Report
**WHEN** quality report is requested
**THEN** returns map of skill names to quality scores

#### Scenario: Calculate Average
**WHEN** quality report includes multiple skills
**THEN** computes and returns average quality score

### REQ-SCORE-007: MCP Tool Integration

The system shall expose quality scoring via MCP tool.

#### Scenario: List All Scores
**WHEN** `skill_quality` tool is called with no parameters
**THEN** returns scores for all skills with average

#### Scenario: Filter by Threshold
**WHEN** `skill_quality` tool is called with `threshold: 80`
**THEN** returns list of skills scoring below 80

### REQ-SCORE-008: Score Caching

The system shall cache computed quality scores in SkillMeta.

#### Scenario: Cache on Load
**WHEN** skill is loaded and parsed
**THEN** quality score is computed and stored in SkillMeta.QualityScore

#### Scenario: Retrieve Cached Score
**WHEN** quality report is requested
**THEN** uses cached scores without recomputation

### REQ-SCORE-009: Threshold Validation

The system shall support quality gates in CI/CD.

#### Scenario: All Skills Above Threshold
**WHEN** all skills score >= 80
**THEN** validation passes

#### Scenario: Skills Below Threshold
**WHEN** any skill scores < 80
**THEN** returns list of failing skills for CI/CD

## Cross-References

- Related to REQ-VALID-001 (Skill Validation)
- Related to REQ-V2-002 (Template Creation)
- Implements design from `design.md`
