# Tasks: Reorganize Plugin Source Layout

**Change ID**: reorganize-plugin-source-layout
**Total Estimated**: ~12 hours

---

## Phase 1: Directory Restructuring (3h)

- [ ] **1.1** Create new directory structure (0.5h)
  - Files: `plugins/sources/`, `plugins/platforms/`, `dist/`
  - Create placeholder directories
  - Dependencies: none

- [ ] **1.2** Move plugin sources to new location (1h)
  - Files: Move `plugins/go-ent/` → `plugins/sources/go-ent/`
  - Preserve git history with `git mv`
  - Dependencies: 1.1

- [ ] **1.3** Extract platform templates (1h)
  - Files: `plugins/go-ent/agents/templates/*.tmpl` → `plugins/platforms/{claude,opencode}/templates/`
  - Move Claude and OpenCode templates to respective platform directories
  - Dependencies: 1.2

- [ ] **1.4** Update .gitignore (0.5h)
  - Files: `.gitignore`
  - Add `dist/` to ignore build artifacts
  - Dependencies: 1.1

## Phase 2: Adapter Updates (4h)

- [ ] **2.1** Update adapter interface paths (1h)
  - Files: `internal/toolinit/adapter.go`
  - Update `SourcesDir()` method to point to `plugins/sources/`
  - Add `PlatformsDir()` method for platform-specific files
  - Add `DistDir()` method for build output
  - Dependencies: 1.2, 1.3

- [ ] **2.2** Update Claude adapter (1.5h)
  - Files: `internal/toolinit/claude.go`
  - Update source paths: `plugins/sources/go-ent/`
  - Update template paths: `plugins/platforms/claude/templates/`
  - Update output paths: `dist/claude/`
  - Dependencies: 2.1

- [ ] **2.3** Update OpenCode adapter (1.5h)
  - Files: `internal/toolinit/opencode.go`
  - Update source paths: `plugins/sources/go-ent/`
  - Update template paths: `plugins/platforms/opencode/templates/`
  - Update output paths: `dist/opencode/`
  - Dependencies: 2.1
  - Parallel with: 2.2

## Phase 3: Loader Updates (2h)

- [ ] **3.1** Update plugin loader (2h)
  - Files: `internal/plugin/loader.go`
  - Scan `plugins/sources/` instead of `plugins/`
  - Update plugin path resolution
  - Update manifest validation
  - Dependencies: 2.2, 2.3

## Phase 4: Testing & Validation (3h)

- [ ] **4.1** Update tests (1.5h)
  - Files: `internal/toolinit/*_test.go`, `internal/plugin/*_test.go`
  - Update test fixtures with new paths
  - Update test expectations
  - Dependencies: 3.1

- [ ] **4.2** Manual testing (1h)
  - Build and run MCP server
  - Test plugin generation for Claude
  - Test plugin generation for OpenCode
  - Verify generated files in `dist/`
  - Dependencies: 4.1

- [ ] **4.3** Documentation updates (0.5h)
  - Files: `docs/DEVELOPMENT.md`, `README.md`
  - Update plugin structure documentation
  - Update development workflow
  - Dependencies: 4.2

---

## Critical Path

```
1.1 → 1.2 → 1.3 → 2.1 → [2.2, 2.3] → 3.1 → 4.1 → 4.2 → 4.3
```

**Parallel work:** Tasks 2.2 and 2.3 can run in parallel
