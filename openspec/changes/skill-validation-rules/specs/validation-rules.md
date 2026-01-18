# Spec: Enhanced Validation Rules

## ADDED Requirements

### REQ-RULE-001: SK010 validates example diversity
**WHEN** skill has 3+ examples with low diversity (<50% score)
**THEN** validation returns warning
**AND** warning includes diversity improvement suggestion

### REQ-RULE-002: SK011 validates skill conciseness
**WHEN** skill body exceeds 5000 tokens
**THEN** validation returns warning
**WHEN** skill body exceeds 8000 tokens
**THEN** validation returns critical warning

### REQ-RULE-003: SK012 recommends explicit triggers
**WHEN** skill uses description-based triggers only
**THEN** validation returns info-level warning
**WHEN** skill has no triggers at all
**THEN** validation returns warning with explicit trigger example

### REQ-RULE-004: SK013 detects skill redundancy
**WHEN** skill has >70% overlap with another skill
**THEN** validation returns warning identifying overlapping skill
**AND** warning suggests merging or differentiation

### REQ-RULE-005: All new rules are non-blocking
**WHEN** any new rule (SK010-SK013) fails
**THEN** validation continues (warnings only, not errors)
**AND** skill can still be loaded and used
