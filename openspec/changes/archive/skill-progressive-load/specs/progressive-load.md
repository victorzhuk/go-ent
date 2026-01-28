# Spec: Progressive Loading

## ADDED Requirements

### REQ-LOAD-001: Three disclosure levels
**WHEN** registry loads skills
**THEN** starts with Level 1 (metadata only, ~100 tokens)

### REQ-LOAD-002: Lazy upgrade to core
**WHEN** skill matches query
**THEN** registry upgrades to Level 2 (core content, <5k tokens)

### REQ-LOAD-003: Full load on execution
**WHEN** skill is executed
**THEN** registry upgrades to Level 3 (full content including references)
