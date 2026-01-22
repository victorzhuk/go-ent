# Spec: Skill Templates

## ADDED Requirements

### REQ-TMPL-001: Basic template passes validation
**WHEN** user copies skill-basic.md and fills placeholders
**THEN** skill passes all validation rules

### REQ-TMPL-002: Complete template demonstrates all features
**WHEN** user opens skill-complete.md
**THEN** template shows all frontmatter options and XML sections

### REQ-TMPL-003: Delegating template shows composition
**WHEN** user opens skill-delegating.md
**THEN** template demonstrates depends_on and delegates_to usage
