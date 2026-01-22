# V2 Format Support Capability

## ADDED Requirements

### REQ-V2-001: Extended Frontmatter Parsing

The system shall parse extended frontmatter fields for v2 skills.

#### Scenario: Parse Version Field
**WHEN** skill frontmatter contains `version: "2.0.0"`
**THEN** SkillMeta.Version is set to "2.0.0"

#### Scenario: Parse Author Field
**WHEN** skill frontmatter contains `author: "go-ent"`
**THEN** SkillMeta.Author is set to "go-ent"

#### Scenario: Parse Tags Field
**WHEN** skill frontmatter contains `tags: [go, code, patterns]`
**THEN** SkillMeta.Tags is set to ["go", "code", "patterns"]

#### Scenario: Parse AllowedTools Field
**WHEN** skill frontmatter contains `allowed-tools: [read, write]`
**THEN** SkillMeta.AllowedTools is set to ["read", "write"]

### REQ-V2-002: XML-Tagged Content Structure

The system shall support XML-tagged content sections.

#### Scenario: Role Section
**WHEN** skill contains `<role>Expert persona</role>`
**THEN** content is recognized as v2 format

#### Scenario: Instructions Section
**WHEN** skill contains `<instructions>Task guidelines</instructions>`
**THEN** section is validated for presence and non-empty content

#### Scenario: Constraints Section
**WHEN** skill contains `<constraints>Boundaries</constraints>`
**THEN** section is validated for list items

#### Scenario: Edge Cases Section
**WHEN** skill contains `<edge_cases>Special handling</edge_cases>`
**THEN** section is validated for coverage of unclear/missing/out-of-scope

#### Scenario: Examples Section
**WHEN** skill contains `<examples><example>...</example></examples>`
**THEN** examples are validated for input/output structure

#### Scenario: Output Format Section
**WHEN** skill contains `<output_format>Structure spec</output_format>`
**THEN** section is recognized and validated

### REQ-V2-003: Trigger Extraction Compatibility

The system shall continue extracting triggers from description for v2 skills.

#### Scenario: Extract V2 Triggers
**WHEN** v2 skill description contains "Auto-activates for: coding, testing"
**THEN** triggers ["coding", "testing"] are extracted

### REQ-V2-004: Template Skill Creation

The system shall provide go-code as exemplary v2 template.

#### Scenario: Template Structure
**WHEN** go-code skill is refactored to v2
**THEN** includes all required XML sections

#### Scenario: Template Quality
**WHEN** go-code template is validated
**THEN** achieves quality score >= 90

#### Scenario: Template Examples
**WHEN** go-code template is reviewed
**THEN** includes 2-3 input/output examples

### REQ-V2-005: Migration Path

The system shall support gradual migration from v1 to v2 format.

#### Scenario: Mixed Format Support
**WHEN** registry loads skills directory with v1 and v2 skills
**THEN** both formats load successfully

#### Scenario: Format Detection
**WHEN** skill is parsed
**THEN** StructureVersion field indicates "v1" or "v2"

### REQ-V2-006: Content Preservation

The system shall preserve existing skill content value during migration.

#### Scenario: Migrate Code Examples
**WHEN** v1 skill is migrated to v2
**THEN** all code blocks are preserved in appropriate sections

#### Scenario: Migrate Tables
**WHEN** v1 skill contains reference tables
**THEN** tables are preserved after XML sections

### REQ-V2-007: Role Definition Standards

The system shall enforce role definition patterns.

#### Scenario: Expert Persona
**WHEN** role is defined for technical skill
**THEN** includes expertise area and experience level

#### Scenario: Behavioral Guidelines
**WHEN** role is defined
**THEN** includes constraints from project CLAUDE.md

### REQ-V2-008: Example Format Standards

The system shall enforce input/output example structure.

#### Scenario: Input Tag Required
**WHEN** example is provided
**THEN** must include `<input>` tag with user request or scenario

#### Scenario: Output Tag Required
**WHEN** example is provided
**THEN** must include `<output>` tag with expected response

#### Scenario: Representative Examples
**WHEN** skill has multiple examples
**THEN** at least one demonstrates typical use case

#### Scenario: Edge Case Examples
**WHEN** skill has multiple examples
**THEN** at least one demonstrates boundary condition

### REQ-V2-009: Constraint Definition Standards

The system shall enforce constraint specificity.

#### Scenario: Include Constraints
**WHEN** constraints section is present
**THEN** specifies what to include with concrete criteria

#### Scenario: Exclude Constraints
**WHEN** constraints section is present
**THEN** specifies what to exclude with concrete criteria

#### Scenario: Boundary Constraints
**WHEN** constraints section is present
**THEN** specifies limits and boundaries

## MODIFIED Requirements

### REQ-PARSE-001: Skill File Parsing
~~Parser extracts only name and description from frontmatter~~

Parser extracts name, description, version, author, tags, and allowed-tools from frontmatter. Detects format version (v1/v2) based on content structure.

**Reason**: Support v2 extended metadata

### REQ-LOAD-001: Skill Loading
~~Registry loads skills with minimal metadata~~

Registry loads skills with full metadata including version, author, tags, quality score, and structure version.

**Reason**: Enable quality tracking and version management

## Cross-References

- Related to REQ-VALID-001 (Structure Validation)
- Related to REQ-SCORE-001 (Quality Scoring)
- Implements design from `design.md`
- Template design in `proposal.md`
