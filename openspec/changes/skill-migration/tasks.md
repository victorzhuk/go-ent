# Tasks: Skill Migration

## 0. Complete Audit

### 0.1 Inventory verification
- [ ] 0.1.1 List all 14 existing skills
- [ ] 0.1.2 Verify 9 Go skills
- [ ] 0.1.3 Verify 5 Core skills
- [ ] 0.1.4 Document in proposal

### 0.2 Current state documentation
- [ ] 0.2.1 Count examples per skill
- [ ] 0.2.2 Measure token counts per skill
- [ ] 0.2.3 Check for explicit triggers
- [ ] 0.2.4 Run quality scoring (when available)
- [ ] 0.2.5 Document baseline metrics

### 0.3 Identify missing skills
- [ ] 0.3.1 Confirm go-migration is needed
- [ ] 0.3.2 Confirm go-config is needed
- [ ] 0.3.3 Confirm go-error is needed
- [ ] 0.3.4 Decide: create go-concurrency or skip
- [ ] 0.3.5 Decide: create go-validation or skip
- [ ] 0.3.6 Decide: create go-idiomatic or skip
- [ ] 0.3.7 Remove non-existent skills from migration tasks

### 0.4 Update proposal with accurate data
- [ ] 0.4.1 Update summary with confirmed skill counts
- [ ] 0.4.2 Update problem section with audit findings
- [ ] 0.4.3 Update success metrics with real numbers
- [ ] 0.4.4 Prioritize migration by quality score

## 1. Migrate Go Skills

### 2.1 go-code
- [ ] 2.1.1 Add explicit triggers (keywords: ["go code", "golang"], file_pattern: "*.go")
- [ ] 2.1.2 Review examples (ensure has 3, ensure diversity)
- [ ] 2.1.3 Check token count, trim if needed
- [ ] 2.1.4 Validate score ≥85

### 2.2 go-arch
- [ ] 2.2.1 Add explicit triggers (keywords: ["architecture", "go design"])
- [ ] 2.2.2 Add 2 more examples (currently has 3)
- [ ] 2.2.3 Trim verbose architecture discussion (likely >5k tokens)
- [ ] 2.2.4 Move detailed patterns to references/

### 2.3 go-api
- [ ] 2.3.1 Add explicit triggers (keywords: ["go api", "rest", "grpc"], file_pattern: "**/api/*.go")
- [ ] 2.3.2 Check examples (add edge cases)
- [ ] 2.3.3 Validate score

### 2.4 go-db
- [ ] 2.4.1 Add explicit triggers (keywords: ["database", "sql"], file_pattern: "**/*_repo.go")
- [ ] 2.4.2 Add depends_on: [go-code]
- [ ] 2.4.3 Ensure 4-5 examples with various database scenarios
- [ ] 2.4.4 Trim if verbose

### 2.5 go-testing
- [ ] 2.5.1 Add explicit triggers (keywords: ["test", "testing"], file_pattern: "**/*_test.go")
- [ ] 2.5.2 Add depends_on: [go-code]
- [ ] 2.5.3 Check examples

### 2.6 go-perf
- [ ] 2.6.1 Add explicit triggers (keywords: ["performance", "optimize"])
- [ ] 2.6.2 Add examples (likely needs more)
- [ ] 2.6.3 Trim verbose optimization guides

## 3. Migrate Core Skills

### 3.1 api-design
- [ ] 3.1.1 Add explicit triggers (keywords: ["api design", "rest api"])
- [ ] 3.1.2 Review examples (ensure has 3, ensure diversity)
- [ ] 3.1.3 Check token count, trim if needed
- [ ] 3.1.4 Validate score ≥85

### 3.2 arch-core
- [ ] 3.2.1 Add explicit triggers (keywords: ["architecture", "clean architecture", "ddd"])
- [ ] 3.2.2 Review examples (ensure has 3, ensure diversity)
- [ ] 3.2.3 Check token count, trim if needed
- [ ] 3.2.4 Validate score ≥85

### 3.3 debug-core
- [ ] 3.3.1 Add explicit triggers (keywords: ["debug", "troubleshoot"], weight: 0.5 - fallback skill)
- [ ] 3.3.2 Review examples (ensure has 3, ensure diversity)
- [ ] 3.3.3 Check token count (currently verbose at 639 lines, trim to <5k tokens)
- [ ] 3.3.4 Validate score ≥85

### 3.4 review-core
- [ ] 3.4.1 Add explicit triggers (keywords: ["code review", "pull request"])
- [ ] 3.4.2 Review examples (ensure has 3, ensure diversity)
- [ ] 3.4.3 Check token count, trim if needed
- [ ] 3.4.4 Validate score ≥85

### 3.5 security-core
- [ ] 3.5.1 Add explicit triggers (keywords: ["security", "authentication", "authorization"])
- [ ] 3.5.2 Review examples (ensure has 3, ensure diversity)
- [ ] 3.5.3 Check token count, trim if needed
- [ ] 3.5.4 Validate score ≥85

## 4. Create New Skills

### 4.1 go-migration
- [ ] 4.1.1 Create from skill-complete.md template
- [ ] 4.1.2 Define role (database migration expert)
- [ ] 4.1.3 Add 4-5 migration examples (add column, drop table, data migration, rollback, etc.)
- [ ] 4.1.4 Add explicit triggers
- [ ] 4.1.5 Validate score ≥85

### 4.2 go-config
- [ ] 4.2.1 Create from template
- [ ] 4.2.2 Define role (configuration management expert)
- [ ] 4.2.3 Add examples (env vars, config files, feature flags, secrets, validation)
- [ ] 4.2.4 Add triggers (keywords + file_pattern "config.go")
- [ ] 4.2.5 Validate score

### 4.3 go-error
- [ ] 4.3.1 Create from template
- [ ] 4.3.2 Define role (Go error handling expert)
- [ ] 4.3.3 Add examples (wrapping, custom errors, error types, sentinel errors, error chains)
- [ ] 4.3.4 Add triggers (keywords ["error handling", "error wrapping"])
- [ ] 4.3.5 Validate score

### 4.4 [SKIP] debug-core
- [x] 4.4.1 Already exists in plugins/go-ent/skills/core/debug-core/SKILL.md
- [ ] 4.4.2 Will be migrated in Phase 3.3
## 5. Validation and Quality Check

### 5.1 Run quality analyzer on all skills
- [ ] 5.1.1 Generate distribution report
- [ ] 5.1.2 Verify all scores ≥80 (target ≥85)

### 5.2 Run full validation suite
- [ ] 5.2.1 Ensure all skills pass validation
- [ ] 5.2.2 Verify no errors, minimal warnings

### 5.3 Test skill activation
- [ ] 5.3.1 Test queries against new trigger system
- [ ] 5.3.2 Verify correct skills activate
- [ ] 5.3.3 Test file-type matching
- [ ] 5.3.4 Test skill dependencies
