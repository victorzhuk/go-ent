## Tasks

### 1. Design v4 Skill Format

- [ ] **1.1** Define v4 skill schema (frontmatter + body structure)
  - Files: `docs/SKILL_FORMAT_V4.md`
  - Dependencies: none
  - Effort: 1h

- [ ] **1.2** Create example v4 skills for all categories
  - Files: `examples/skills/v4/`
  - Dependencies: 1.1
  - Effort: 2h

### 2. Update Skill Parser

- [ ] **2.1** Simplify parser to handle v4 format
  - Files: `internal/skill/parser.go`
  - Dependencies: 1.2
  - Effort: 2h

- [ ] **2.2** Remove progressive loading logic
  - Files: `internal/skill/parser.go`, `internal/skill/registry.go`
  - Dependencies: 2.1
  - Effort: 1h

- [ ] **2.3** Update parser tests
  - Files: `internal/skill/parser_test.go`
  - Dependencies: 2.2
  - Effort: 1h

### 3. Simplify Validation

- [ ] **3.1** Reduce validation rules to 5 essentials
  - Files: `internal/skill/validator.go`, `internal/skill/rules.go`
  - Dependencies: 2.1
  - Effort: 2h

- [ ] **3.2** Remove quality scoring system
  - Files: `internal/skill/scorer.go`, `internal/skill/scorer_*_test.go`
  - Dependencies: 3.1
  - Effort: 1h

- [ ] **3.3** Remove overlap detection
  - Files: `internal/skill/overlap.go`, `internal/skill/overlap_test.go`
  - Dependencies: 3.2
  - Effort: 1h

- [ ] **3.4** Remove auto-fixer
  - Files: `internal/skill/fixer.go`, `internal/skill/fixer_test.go`
  - Dependencies: 3.3
  - Effort: 1h

### 4. Update MCP Tools

- [ ] **4.1** Simplify skill_list tool
  - Files: `internal/mcp/tools/skill_list.go`
  - Dependencies: 2.2
  - Effort: 1h

- [ ] **4.2** Remove skill_info, skill_validate, skill_quality tools
  - Files: `internal/mcp/tools/skill_*.go`
  - Dependencies: 4.1
  - Effort: 1h

### 5. Convert Existing Skills

- [ ] **5.1** Create conversion script v3 → v4
  - Files: `scripts/convert-skills.go`
  - Dependencies: 2.2
  - Effort: 2h

- [ ] **5.2** Convert all existing skills to v4
  - Files: `internal/cli/.claude/skills/**/*.md`, `internal/cli/.opencode/skills/**/*.md`
  - Dependencies: 5.1
  - Effort: 2h

### 6. Documentation

- [ ] **6.1** Update SKILL-AUTHORING.md for v4
  - Files: `docs/SKILL-AUTHORING.md`
  - Dependencies: 5.2
  - Effort: 1h

- [ ] **6.2** Update AGENTS_AND_SKILLS.md
  - Files: `docs/AGENTS_AND_SKILLS.md`
  - Dependencies: 6.1
  - Effort: 1h

---

**Total Tasks**: 14
**Estimated Effort**: ~16 hours
