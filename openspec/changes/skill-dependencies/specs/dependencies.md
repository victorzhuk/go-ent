# Spec: Skill Dependencies

## ADDED Requirements

### REQ-DEP-001: Skills can depend on other skills
**WHEN** skill A depends_on skill B
**THEN** skill B loads before skill A

### REQ-DEP-002: Circular dependencies detected
**WHEN** skill A depends on B, B depends on A
**THEN** registry loading fails with clear error

### REQ-DEP-003: Delegation hints stored
**WHEN** skill has delegates_to metadata
**THEN** metadata available in skill selection UI
