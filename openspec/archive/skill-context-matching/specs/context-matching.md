# Spec: Context-Aware Matching

## ADDED Requirements

### REQ-CTX-001: File-type boosting
**WHEN** query matches skill and current files are *.go
**AND** skill has file_pattern trigger for "*.go"
**THEN** skill score boosted by +0.2

### REQ-CTX-002: Task-type boosting
**WHEN** query is "implement X" (task type: implement)
**AND** skill description mentions "implement"
**THEN** skill score boosted by +0.15

### REQ-CTX-003: Graceful degradation
**WHEN** no context provided
**THEN** matching works with query-only scoring
