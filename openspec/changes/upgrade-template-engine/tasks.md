# Tasks: Upgrade Template Engine

**Change ID**: upgrade-template-engine
**Total Estimated**: ~10 hours

---

## Phase 1: Template Engine Core (4h)

- [ ] **1.1** Create template package structure (0.5h)
  - Files: `internal/template/engine.go`, `funcs.go`, `parse.go`
  - Define package structure and interfaces
  - Dependencies: none

- [ ] **1.2** Implement template engine (2h)
  - Files: `internal/template/engine.go`
  - Template loading and parsing
  - File system integration
  - Error handling
  - Dependencies: 1.1

- [ ] **1.3** Implement custom functions (1h)
  - Files: `internal/template/funcs.go`
  - `include(path string)` - File inclusion
  - `indent(n int, s string)` - Indentation
  - `default(fallback, value)` - Default values
  - Dependencies: 1.2

- [ ] **1.4** Add template validation (0.5h)
  - Files: `internal/template/parse.go`
  - Compile-time template syntax validation
  - Dependency checking
  - Dependencies: 1.2

## Phase 2: Integration (2h)

- [ ] **2.1** Update PromptComposer (1.5h)
  - Files: `internal/toolinit/transform.go`
  - Replace custom `{{include}}` with template engine
  - Maintain backward compatibility with `.md` files
  - Dependencies: 1.3, 1.4

- [ ] **2.2** Update adapter template processing (0.5h)
  - Files: `internal/toolinit/claude.go`, `opencode.go`
  - Use new template engine for frontmatter templates
  - Dependencies: 2.1

## Phase 3: Template Migration (2h)

- [ ] **3.1** Create base template (0.5h)
  - Files: `plugins/sources/go-ent/agents/prompts/_base.md.tmpl`
  - Convert `_base.md` to template format
  - Define common template blocks
  - Dependencies: 2.2

- [ ] **3.2** Migrate driver template (0.5h)
  - Files: `plugins/sources/go-ent/agents/prompts/_driver.md.tmpl`
  - Convert with template inheritance
  - Dependencies: 3.1

- [ ] **3.3** Migrate one agent as proof-of-concept (1h)
  - Files: `plugins/sources/go-ent/agents/prompts/agents/architect.md.tmpl`
  - Demonstrate inheritance from `_base.md.tmpl`
  - Use `{{define}}` and `{{template}}` directives
  - Dependencies: 3.1

## Phase 4: Testing (2h)

- [ ] **4.1** Unit tests for template engine (1h)
  - Files: `internal/template/engine_test.go`, `funcs_test.go`, `parse_test.go`
  - Test template parsing
  - Test custom functions
  - Test error handling
  - Dependencies: 1.3, 1.4

- [ ] **4.2** Integration tests (0.5h)
  - Files: `internal/toolinit/transform_test.go`
  - Test template engine integration with adapters
  - Test backward compatibility with `.md` files
  - Dependencies: 2.2, 3.3

- [ ] **4.3** Manual testing (0.5h)
  - Build and generate agent configs
  - Verify template inheritance works
  - Check backward compatibility
  - Dependencies: 4.2

---

## Critical Path

```
1.1 → 1.2 → [1.3, 1.4] → 2.1 → 2.2 → 3.1 → [3.2, 3.3] → 4.1 → 4.2 → 4.3
```

**Parallel work:**
- Tasks 1.3 and 1.4 can run in parallel after 1.2
- Tasks 3.2 and 3.3 can run in parallel after 3.1
