# Spec: Weighted Trigger Matching

## ADDED Requirements

### REQ-MATCH-001: Skills ranked by relevance score
**WHEN** multiple skills match a query
**THEN** results are sorted by score (highest first)

### REQ-MATCH-002: Pattern matching with weights
**WHEN** skill has pattern trigger with weight 0.9
**AND** query matches pattern
**THEN** match score is 0.9

### REQ-MATCH-003: Fallback to description-based matching
**WHEN** skill has no explicit triggers
**THEN** system falls back to description-based keyword extraction
**AND** assigns default weight (0.5)

## MODIFIED Requirements

### REQ-MATCH-004: FindMatchingSkills return type changed
**Old**: Returns `[]*SkillMetadata`
**New**: Returns `[]MatchResult` with scores and match reasons
**Reason**: Enable relevance ranking and debugging
