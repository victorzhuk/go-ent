# Tasks: Skill System Refactoring

## 1. Foundation - Infrastructure

### 1.1 Extend Domain Types
- [x] **1.1** Extend `internal/domain/skill.go` SkillMetadata ✓
  - Files: internal/domain/skill.go
  - Dependencies: none
  - Effort: 1h
  - Details:
    - Add Version, Author, Tags, AllowedTools fields
    - Add QualityScore, StructureVersion fields
    - Update documentation

### 1.2 Extend Parser
- [x] **1.2** Extend parser for v2 frontmatter ✓
  - Files: internal/skill/parser.go
  - Dependencies: 1.1
  - Effort: 2h
  - Details:
    - Add detectVersion() method
    - Add parseFrontmatterV2() method
    - Update ParseSkillFile() to handle both v1 and v2
    - Update SkillMeta struct with new fields

### 1.3 Create Validator
- [x] **1.3** Create skill validator with rules ✓
  - Files: internal/skill/validator.go, internal/skill/rules.go
  - Dependencies: 1.1, 1.2
  - Effort: 3h
  - Details:
    - Create Validator struct following spec/validator.go pattern
    - Implement ValidationContext and ValidationResult types
    - Create 9 validation rules:
      - validateFrontmatter
      - validateVersion
      - validateXMLTags
      - validateRoleSection
      - validateInstructionsSection
      - validateExamples
      - validateConstraints
      - validateEdgeCases
      - validateOutputFormat
    - Write unit tests for each rule

### 1.4 Create Quality Scorer
- [x] **1.4** Create quality scoring system ✓
  - Files: internal/skill/scorer.go
  - Dependencies: 1.1
  - Effort: 2h
  - Details:
    - Create QualityScorer struct
    - Implement scoreFrontmatter() (20 points)
    - Implement scoreStructure() (30 points)
    - Implement scoreContent() (30 points)
    - Implement scoreTriggers() (20 points)
    - Write unit tests

### 1.5 Extend Registry
- [x] **1.5** Extend registry with validation ✓
  - Files: internal/skill/registry.go
  - Dependencies: 1.2, 1.3, 1.4
  - Effort: 2h
  - Details:
    - Add validator and scorer fields to Registry
    - Implement ValidateSkill(name) method
    - Implement ValidateAll() method
    - Implement GetQualityReport() method
    - Update Load() to compute quality scores
    - Write integration tests

## 2. Tooling - MCP Integration

### 2.1 Skill Validation Tool
- [x] **2.1** Create skill_validate MCP tool ✓
  - Files: internal/mcp/tools/skill_validate.go
  - Dependencies: 1.5
  - Effort: 2h
  - Details:
    - Define SkillValidateInput struct (name, strict)
    - Define SkillValidateOutput struct (valid, score, issues)
    - Implement tool handler
    - Add output formatting
    - Write integration tests

### 2.2 Skill Quality Tool
- [x] **2.2** Create skill_quality MCP tool ✓
  - Files: internal/mcp/tools/skill_quality.go
  - Dependencies: 1.5
  - Effort: 1h
  - Details:
    - Define SkillQualityInput struct (threshold)
    - Define SkillQualityOutput struct (skills, avg_score, below_threshold)
    - Implement tool handler
    - Add output formatting
    - Write integration tests

### 2.3 Register Tools
- [x] **2.3** Register new MCP tools ✓
  - Files: internal/mcp/tools/register.go
  - Dependencies: 2.1, 2.2
  - Effort: 0.5h
  - Details:
    - Add registerSkillValidate call
    - Add registerSkillQuality call
    - Update tool documentation

## 3. Template - Go-Code Refactoring

### 3.1 Refactor Go-Code Skill
- [x] **3.1** Refactor go-code to v2 format ✓
  - Files: plugins/go-ent/skills/go/go-code/SKILL.md
  - Dependencies: 2.3
  - Effort: 3h
  - Details:
    - Update frontmatter (version: 2.0.0, author, tags)
    - Add `<role>` section: Expert Go developer persona
    - Convert content to `<instructions>` section
    - Extract constraints from CLAUDE.md into `<constraints>`
    - Add `<edge_cases>`: delegate to go-test, go-arch, go-perf
    - Add 2-3 `<examples>` with input/output pairs:
      - Bootstrap pattern example
      - Error handling example
    - Add `<output_format>` section
    - Preserve existing code blocks as reference

### 3.2 Validate Template
- [x] **3.2** Validate go-code template ✓
  - Files: none (validation only)
  - Dependencies: 3.1
  - Effort: 0.5h
  - Details:
    - Run `skill_validate name=go-code strict=true`
    - Verify quality score >= 90
    - Fix any validation errors
    - Document as template reference

## 4. Migration - Remaining Skills

### 4.1 Migrate Tier 1 Skills
- [x] **4.1.1** Migrate go-arch skill ✓
  - Files: plugins/go-ent/skills/go/go-arch/SKILL.md
  - Dependencies: 3.2
  - Effort: 1h
  - Parallel with: 4.1.2, 4.1.3, 4.1.4

- [x] **4.1.2** Migrate go-test skill ✓
  - Files: plugins/go-ent/skills/go/go-test/SKILL.md
  - Dependencies: 3.2
  - Effort: 1h
  - Parallel with: 4.1.1, 4.1.3, 4.1.4

- [x] **4.1.3** Migrate go-db skill ✓
  - Files: plugins/go-ent/skills/go/go-db/SKILL.md
  - Dependencies: 3.2
  - Effort: 1h
  - Parallel with: 4.1.1, 4.1.2, 4.1.4

- [x] **4.1.4** Migrate go-api skill ✓
  - Files: plugins/go-ent/skills/go/go-api/SKILL.md
  - Dependencies: 3.2
  - Effort: 1h
  - Parallel with: 4.1.1, 4.1.2, 4.1.3

### 4.2 Migrate Tier 2 Skills
- [x] **4.2.1** Migrate go-sec skill ✓
  - Files: plugins/go-ent/skills/go/go-sec/SKILL.md
  - Dependencies: 4.1.4
  - Effort: 1h
  - Parallel with: 4.2.2, 4.2.3, 4.2.4

- [x] **4.2.2** Migrate go-perf skill ✓
  - Files: plugins/go-ent/skills/go/go-perf/SKILL.md
  - Dependencies: 4.1.4
  - Effort: 1h
  - Parallel with: 4.2.1, 4.2.3, 4.2.4

- [x] **4.2.3** Migrate go-review skill ✓
  - Files: plugins/go-ent/skills/go/go-review/SKILL.md
  - Dependencies: 4.1.4
  - Effort: 1h
  - Parallel with: 4.2.1, 4.2.2, 4.2.4

- [x] **4.2.4** Migrate go-ops skill ✓
  - Files: plugins/go-ent/skills/go/go-ops/SKILL.md
  - Dependencies: 4.1.4
  - Effort: 1h
  - Parallel with: 4.2.1, 4.2.2, 4.2.3

### 4.3 Migrate Tier 3 Skills
- [x] **4.3.1** Migrate api-design skill ✓
  - Files: plugins/go-ent/skills/core/api-design/SKILL.md
  - Dependencies: 4.2.4
  - Effort: 1h
  - Parallel with: 4.3.2, 4.3.3, 4.3.4, 4.3.5

- [x] **4.3.2** Migrate arch-core skill ✓
  - Files: plugins/go-ent/skills/core/arch-core/SKILL.md
  - Dependencies: 4.2.4
  - Effort: 1h
  - Parallel with: 4.3.1, 4.3.3, 4.3.4, 4.3.5

- [x] **4.3.3** Migrate debug-core skill ✓
  - Files: plugins/go-ent/skills/core/debug-core/SKILL.md
  - Dependencies: 4.2.4
  - Effort: 1h
  - Parallel with: 4.3.1, 4.3.2, 4.3.4, 4.3.5

- [x] **4.3.4** Migrate review-core skill ✓
  - Files: plugins/go-ent/skills/core/review-core/SKILL.md
  - Dependencies: 4.2.4
  - Effort: 1h
  - Parallel with: 4.3.1, 4.3.2, 4.3.3, 4.3.5

- [x] **4.3.5** Migrate security-core skill ✓
  - Files: plugins/go-ent/skills/core/security-core/SKILL.md
  - Dependencies: 4.2.4
  - Effort: 1h
  - Parallel with: 4.3.1, 4.3.2, 4.3.3, 4.3.4

## 5. Tooling - Build Integration

### 5.1 Skill Sync Command
- [x] **5.1** Create skill-sync command ✓
  - Files: plugins/go-ent/commands/skill-sync.md
  - Dependencies: 4.3.5
  - Effort: 1h
  - Details:
    - Create command to sync plugins/go-ent/skills/ to .claude/skills/ent/
    - Validate all skills after sync
    - Report any validation errors
    - Add error handling

### 5.2 Makefile Targets
- [x] **5.2** Add Makefile targets ✓
  - Files: Makefile
  - Dependencies: 5.1
  - Effort: 0.5h
  - Details:
    - Add `make skill-validate` target
    - Add `make skill-sync` target
    - Add `make skill-quality` target
    - Update help text

## 6. Documentation

### 6.1 Update Development Docs
- [x] **6.1** Update docs/DEVELOPMENT.md ✓
  - Files: docs/DEVELOPMENT.md
  - Dependencies: 5.2
  - Effort: 1h
  - Details:
    - Add skill authoring section
    - Document required XML sections
    - Explain quality scoring
    - Add migration checklist
    - Document MCP tools usage

### 6.2 Create Authoring Guide
- [x] **6.2** Create docs/SKILL-AUTHORING.md ✓
  - Files: docs/SKILL-AUTHORING.md
  - Dependencies: 6.1
  - Effort: 2h
  - Details:
    - Complete skill template with all sections
    - Validation rules explained
    - Quality scoring rubric
    - Examples of good/bad patterns
    - Migration guide
    - Best practices from research guide

## Task Summary

**Total Tasks**: 31
**Completed**: 31/31 (100%) ✓
**Total Effort**: ~34 hours

### Critical Path
```
1.1 → 1.2 → 1.3 → 1.5 → 2.1 → 2.3 → 3.1 → 3.2 → 4.1.* → 4.2.* → 4.3.* → 5.1 → 5.2 → 6.1 → 6.2
```

### Parallelizable Tasks
- Phase 1: All can run sequentially (dependencies)
- Phase 2: 2.1 and 2.2 can run in parallel after 1.5
- Phase 4:
  - Tier 1: 4.1.1-4.1.4 in parallel
  - Tier 2: 4.2.1-4.2.4 in parallel
  - Tier 3: 4.3.1-4.3.5 in parallel

### Dependencies Graph
```
Phase 1 (Foundation)
    1.1 (Domain Types)
     ├─> 1.2 (Parser)
     │    └─> 1.5 (Registry)
     ├─> 1.3 (Validator)
     │    └─> 1.5 (Registry)
     └─> 1.4 (Scorer)
          └─> 1.5 (Registry)

Phase 2 (MCP Tools)
    1.5 (Registry)
     ├─> 2.1 (Validate Tool)
     │    └─> 2.3 (Register)
     └─> 2.2 (Quality Tool)
          └─> 2.3 (Register)

Phase 3 (Template)
    2.3 (Register)
     └─> 3.1 (go-code refactor)
          └─> 3.2 (Validate template)

Phase 4 (Migration)
    3.2 (Validate template)
     └─> 4.1.* (Tier 1 - parallel)
          └─> 4.2.* (Tier 2 - parallel)
               └─> 4.3.* (Tier 3 - parallel)

Phase 5 (Build)
    4.3.5 (Last migration)
     └─> 5.1 (Sync command)
          └─> 5.2 (Makefile)

Phase 6 (Docs)
    5.2 (Makefile)
     └─> 6.1 (DEVELOPMENT.md)
          └─> 6.2 (SKILL-AUTHORING.md)
```

## Validation Checklist

After each phase:

### Phase 1 Complete
- [x] `go test ./internal/skill/...` passes ✓
- [x] Parser handles v1 and v2 format detection ✓
- [x] Validator produces meaningful issues with line numbers ✓
- [x] Quality scorer produces scores 0-100 ✓
- [x] All unit tests pass ✓

### Phase 2 Complete
- [x] `skill_validate` tool works via MCP ✓
- [x] `skill_quality` tool lists all skills with scores ✓
- [x] MCP tools return proper JSON responses ✓

### Phase 3 Complete
- [x] go-code validates strict mode with 0 errors ✓
- [x] go-code quality score >= 90 ✓
- [x] All XML sections present and well-formed ✓
- [x] Template documented and ready for reference ✓

### Phase 4 Complete
- [x] All 14 skills validate with 0 errors in strict mode ✓
- [x] Average quality score >= 80 ✓
- [x] All skills synced to `.claude/skills/ent/` ✓
- [x] No regression in skill functionality ✓

### Phase 5 Complete
- [x] `make skill-validate` passes ✓
- [x] `make skill-sync` works ✓
- [x] `make skill-quality` generates report ✓

### Phase 6 Complete
- [x] Documentation complete and accurate ✓
- [x] Migration guide published ✓
- [x] Examples provided for all patterns ✓
