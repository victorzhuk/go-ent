# Spec: Skill Migration

## MODIFIED Requirements

### REQ-MIG-001: All existing skills have explicit triggers
**WHEN** loading any skill from plugins/go-ent/skills/
**THEN** skill has explicit triggers section in frontmatter
**AND** triggers have appropriate weights

### REQ-MIG-002: All existing skills meet example standards
**WHEN** validating any skill
**THEN** skill has 3-5 examples
**AND** examples show diversity (different input types, edge cases)

### REQ-MIG-003: All existing skills meet conciseness standards
**WHEN** measuring skill token count
**THEN** core instructions are <5000 tokens
**AND** detailed content moved to references/ if needed

### REQ-MIG-004: All existing skills score ≥80
**WHEN** calculating quality score for any skill
**THEN** score is ≥80 (target ≥85)
**AND** breakdown shows no category below 50% of max

## ADDED Requirements

### REQ-MIG-005: New go-migration skill covers database migrations
**WHEN** user queries about database migrations or schema changes
**THEN** go-migration skill matches
**AND** skill provides migration best practices

### REQ-MIG-006: New go-config skill covers configuration
**WHEN** user queries about configuration management
**THEN** go-config skill matches
**AND** skill covers env vars, files, flags, secrets

### REQ-MIG-007: New go-error skill covers error handling
**WHEN** user queries about error handling patterns
**THEN** go-error skill matches
**AND** skill covers wrapping, custom errors, sentinel errors

### REQ-MIG-008: New debug-core skill is fallback debugger
**WHEN** no language-specific debugger matches
**THEN** debug-core skill matches with lower score
**AND** skill provides language-agnostic debugging strategies

### REQ-MIG-009: Dependencies properly defined
**WHEN** loading go-db or go-testing skills
**THEN** go-code dependency is loaded first
**AND** registry resolves dependencies correctly
