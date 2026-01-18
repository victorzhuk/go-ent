# Tasks: Skill Migration

## 1. Audit Existing Skills
- [ ] 1.1 Run quality analyzer on all 12 skills
- [ ] 1.2 Document current state (triggers, examples, token counts, scores)
- [ ] 1.3 Prioritize by quality score (lowest first)
- [ ] 1.4 Create migration checklist per skill

## 2. Migrate Go Skills

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

### 2.7 go-concurrency
- [ ] 2.7.1 Add explicit triggers (keywords: ["concurrent", "goroutine", "channel"])
- [ ] 2.7.2 Ensure diverse examples (simple, complex, edge cases)
- [ ] 2.7.3 Trim to <5k tokens if verbose

### 2.8 go-validation
- [ ] 2.8.1 Add explicit triggers (keywords: ["validate", "validation"], file_pattern: "**/*_validator.go")
- [ ] 2.8.2 Check examples

### 2.9 go-idiomatic
- [ ] 2.9.1 Add explicit triggers (keywords: ["idiomatic", "go style"])
- [ ] 2.9.2 Check examples

## 3. Migrate Core Skills

### 3.1 ent-task
- [ ] 3.1.1 Add explicit triggers
- [ ] 3.1.2 Ensure quality standards
- [ ] 3.1.3 Update examples if needed

### 3.2 ent-plan
- [ ] 3.2.1 Add explicit triggers
- [ ] 3.2.2 Ensure quality standards
- [ ] 3.2.3 Update examples if needed

### 3.3 ent-debug
- [ ] 3.3.1 Add explicit triggers
- [ ] 3.3.2 Ensure quality standards
- [ ] 3.3.3 Update examples if needed

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

### 4.4 debug-core
- [ ] 4.4.1 Create from template
- [ ] 4.4.2 Define role (general debugging expert, language-agnostic)
- [ ] 4.4.3 Add examples (log analysis, stack traces, reproduction, binary search debugging, hypothesis testing)
- [ ] 4.4.4 Add triggers (keywords with low weight 0.5, fallback skill)
- [ ] 4.4.5 Validate score

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
