# Skill Validation Capability

## ADDED Requirements

### REQ-VALID-001: Skill Content Validation

The system shall validate skill content structure against research-backed patterns.

#### Scenario: Validate Complete V2 Skill
**WHEN** a v2 skill with all required sections is validated
**THEN** validation passes with no errors and quality score >= 90

#### Scenario: Detect Missing Required Section
**WHEN** a v2 skill missing `<role>` section is validated
**THEN** validation returns error "missing required section: <role>"

#### Scenario: Detect Unbalanced XML Tags
**WHEN** a skill has `<role>` but no `</role>` closing tag
**THEN** validation returns error "unbalanced <role> tags: 1 open, 0 close"

### REQ-VALID-002: Frontmatter Validation

The system shall validate frontmatter fields are present and correctly formatted.

#### Scenario: Validate Required Fields
**WHEN** a skill is missing the `name` field in frontmatter
**THEN** validation returns error "missing required field: name"

#### Scenario: Validate Semantic Version
**WHEN** a skill has version "2.x" (invalid semver)
**THEN** validation returns error "invalid semantic version: 2.x"

#### Scenario: Valid Semantic Version
**WHEN** a skill has version "2.0.0"
**THEN** version validation passes

### REQ-VALID-003: Example Validation

The system shall validate that examples follow input/output pairing pattern.

#### Scenario: Validate Example Structure
**WHEN** an `<examples>` section contains `<example>` without `<input>` tag
**THEN** validation returns error "example missing <input> tag"

#### Scenario: Validate Example Count
**WHEN** a v2 skill has fewer than 2 examples in strict mode
**THEN** validation returns warning "should have at least 2 examples"

### REQ-VALID-004: Edge Case Validation

The system shall validate that edge cases cover common failure scenarios.

#### Scenario: Validate Edge Case Coverage
**WHEN** `<edge_cases>` section exists but doesn't mention "unclear" or "missing"
**THEN** validation returns warning "edge cases should handle unclear input and missing information"

### REQ-VALID-005: Version Detection

The system shall automatically detect skill format version.

#### Scenario: Detect V2 Format
**WHEN** skill content contains `<role>` tag
**THEN** skill is detected as version "v2"

#### Scenario: Detect V1 Format
**WHEN** skill content has no XML tags
**THEN** skill is detected as version "v1"

### REQ-VALID-006: Validation Modes

The system shall support both strict and lenient validation modes.

#### Scenario: Strict Mode Errors
**WHEN** validation runs in strict mode
**THEN** missing optional sections generate errors

#### Scenario: Lenient Mode Warnings
**WHEN** validation runs in lenient mode (default)
**THEN** missing optional sections generate warnings

### REQ-VALID-007: MCP Tool Integration

The system shall expose validation via MCP tool for Claude Code integration.

#### Scenario: Validate Single Skill
**WHEN** `skill_validate` tool is called with `name: "go-code"`
**THEN** returns validation result for go-code skill only

#### Scenario: Validate All Skills
**WHEN** `skill_validate` tool is called with no name parameter
**THEN** returns combined validation result for all skills

#### Scenario: Strict Mode Flag
**WHEN** `skill_validate` tool is called with `strict: true`
**THEN** runs validation in strict mode

### REQ-VALID-008: Issue Reporting

The system shall report validation issues with precise location information.

#### Scenario: Report Line Numbers
**WHEN** validation finds issue in skill content
**THEN** issue includes line number where problem occurs

#### Scenario: Report Severity Levels
**WHEN** validation completes
**THEN** issues categorized as "error", "warning", or "info"

### REQ-VALID-009: Backward Compatibility

The system shall continue loading v1 skills without errors.

#### Scenario: Load V1 Skill
**WHEN** a v1 format skill is loaded
**THEN** skill loads successfully with warning "v1 format detected"

#### Scenario: Skip V1 Validation
**WHEN** v1 skill is validated
**THEN** validation skips v2-specific rules

## Cross-References

- Related to REQ-SCORE-001 (Quality Scoring)
- Related to REQ-V2-001 (V2 Format Support)
- Implements design from `design.md`
