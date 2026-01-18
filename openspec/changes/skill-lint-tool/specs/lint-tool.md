# Spec: Skill Linting Tool

## ADDED Requirements

### REQ-LINT-001: Lint command validates skills
**WHEN** running `skill lint path/to/skill.md`
**THEN** command validates skill and displays errors

### REQ-LINT-002: Auto-fix applies corrections
**WHEN** running `skill lint --fix path/to/skill.md`
**THEN** command fixes auto-fixable issues and reports results

### REQ-LINT-003: JSON output for CI
**WHEN** running `skill lint --json path/`
**THEN** command outputs structured JSON with all findings

### REQ-LINT-004: Exit codes indicate status
**WHEN** all skills pass: exit 0
**WHEN** errors found: exit 1
**WHEN** command error: exit 2
