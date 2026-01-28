# Spec: Explicit Skill Triggers

## ADDED Requirements

### REQ-TRIG-001: Skills can define explicit triggers in frontmatter

Skills may define explicit triggers with weights and pattern types.

#### Scenario: Skill with explicit keyword triggers
**WHEN** skill frontmatter contains:
```yaml
triggers:
  - keywords: ["go code", "golang"]
    weight: 0.8
```
**THEN** parser extracts triggers correctly
**AND** triggers have weight 0.8

#### Scenario: Skill with pattern triggers
**WHEN** skill frontmatter contains:
```yaml
triggers:
  - pattern: "implement.*go"
    weight: 0.9
```
**THEN** parser compiles regex pattern
**AND** trigger has weight 0.9

#### Scenario: Skill with file pattern triggers
**WHEN** skill frontmatter contains:
```yaml
triggers:
  - file_pattern: "*.go"
    weight: 0.6
```
**THEN** parser stores file pattern
**AND** trigger can match file types

### REQ-TRIG-002: Default weight applied if not specified

#### Scenario: Trigger without explicit weight
**WHEN** trigger does not specify weight
**THEN** parser assigns default weight 0.7

### REQ-TRIG-003: Weight validation

#### Scenario: Invalid weight value
**WHEN** trigger weight is <0.0 or >1.0
**THEN** parser returns validation error

### REQ-TRIG-004: Backward compatibility maintained

Skills without explicit triggers continue working.

#### Scenario: Skill with no explicit triggers
**WHEN** skill has no triggers section
**THEN** system extracts triggers from description
**AND** extracted triggers have weight 0.5 (lower than explicit)

### REQ-TRIG-005: SK012 validation rule added

#### Scenario: Skill using description-based triggers
**WHEN** validating skill without explicit triggers
**THEN** SK012 rule returns info-level warning
**AND** warning suggests adding explicit triggers section

#### Scenario: Skill with explicit triggers
**WHEN** validating skill with explicit triggers
**THEN** SK012 rule passes without warning
