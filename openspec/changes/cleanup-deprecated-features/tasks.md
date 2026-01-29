## Tasks

### 1. Archive Cleanup

- [ ] **1.1** Backup archived changes
  - Files: `openspec/changes/archive/` (backup)
  - Dependencies: none
  - Effort: 0.5h

- [ ] **1.2** Add openspec/changes/archive/ to .gitignore
  - Files: `.gitignore`
  - Dependencies: 1.1
  - Effort: 0.5h

- [ ] **1.3** Remove archived changes from git tracking
  - Command: `git rm -r openspec/changes/archive/`
  - Dependencies: 1.2
  - Effort: 0.5h

### 2. Remove Unused Packages

- [ ] **2.1** Remove internal/execution/ package
  - Files: `internal/execution/` (entire directory)
  - Dependencies: none
  - Effort: 0.5h

- [ ] **2.2** Remove internal/marketplace/ package
  - Files: `internal/marketplace/` (entire directory)
  - Dependencies: none
  - Effort: 0.5h

- [ ] **2.3** Clean up internal/template/testdata/
  - Files: `internal/template/testdata/`
  - Dependencies: none
  - Effort: 0.5h

- [ ] **2.4** Remove internal/mcp/tools/exports/
  - Files: `internal/mcp/tools/exports/`
  - Dependencies: none
  - Effort: 0.5h

- [ ] **2.5** Review and clean internal/mcp/tools/testdata/
  - Files: `internal/mcp/tools/testdata/`
  - Dependencies: none
  - Effort: 0.5h

### 3. Remove Deprecated Code

- [ ] **3.1** Remove old skill format parsers (v1, v2)
  - Files: `internal/skill/parser.go` (legacy code)
  - Dependencies: simplify-skill-format completion
  - Effort: 1h

- [ ] **3.2** Remove deprecated MCP tool handlers
  - Files: Various in `internal/mcp/tools/`
  - Dependencies: consolidate-mcp-tools completion
  - Effort: 1h

- [ ] **3.3** Remove unused domain types
  - Files: `internal/domain/`
  - Dependencies: none
  - Effort: 1h

- [ ] **3.4** Remove old validation rules
  - Files: `internal/skill/rules.go`
  - Dependencies: simplify-skill-format completion
  - Effort: 0.5h

### 4. Clean up Test Artifacts

- [ ] **4.1** Remove coverage files from git
  - Files: `coverage.out`, `reg_coverage.out`, `coverage_skill.out`, `coverage_template.out`, `combined_coverage.out`
  - Dependencies: none
  - Effort: 0.5h

- [ ] **4.2** Remove temporary test files from mcp/tools
  - Files: `internal/mcp/tools/*.json`, `internal/mcp/tools/*.csv`, `internal/mcp/tools/*.prom`
  - Dependencies: none
  - Effort: 0.5h

- [ ] **4.3** Add test artifacts to .gitignore
  - Files: `.gitignore`
  - Dependencies: 4.1, 4.2
  - Effort: 0.5h

### 5. Update .gitignore

- [ ] **5.1** Add comprehensive test artifact patterns
  - Files: `.gitignore`
  - Dependencies: 4.3
  - Effort: 0.5h

- [ ] **5.2** Add temporary directory patterns
  - Files: `.gitignore`
  - Dependencies: 5.1
  - Effort: 0.5h

- [ ] **5.3** Add IDE patterns
  - Files: `.gitignore`
  - Dependencies: 5.2
  - Effort: 0.5h

### 6. Verify Build

- [ ] **6.1** Run go mod tidy
  - Command: `go mod tidy`
  - Dependencies: 2.1, 2.2, 3.1, 3.2, 3.3
  - Effort: 0.5h

- [ ] **6.2** Build binary
  - Command: `make build`
  - Dependencies: 6.1
  - Effort: 0.5h

- [ ] **6.3** Run tests
  - Command: `make test`
  - Dependencies: 6.2
  - Effort: 1h

- [ ] **6.4** Run linter
  - Command: `make lint`
  - Dependencies: 6.3
  - Effort: 0.5h

### 7. Documentation

- [ ] **7.1** Update CHANGELOG.md
  - Files: `CHANGELOG.md`
  - Dependencies: 6.4
  - Effort: 0.5h

- [ ] **7.2** Document cleanup in development guide
  - Files: `docs/DEVELOPMENT.md`
  - Dependencies: 7.1
  - Effort: 0.5h

---

**Total Tasks**: 22
**Estimated Effort**: ~14 hours
