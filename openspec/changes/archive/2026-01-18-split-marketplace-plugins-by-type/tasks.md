# Tasks: Split Marketplace Plugins by Type

**Change ID**: split-marketplace-plugins-by-type
**Total Estimated**: ~16 hours

---

## Phase 1: Plugin Manifest Enhancement (3h)

- [ ] **1.1** Add dependencies field to manifest (1h)
  - Files: `internal/plugin/manifest.go`
  - Add `Dependencies []string` field to `Manifest` struct
  - Update manifest parsing
  - Dependencies: none

- [ ] **1.2** Implement dependency validation (1h)
  - Files: `internal/plugin/validate.go` - **NEW**
  - Check for circular dependencies
  - Validate dependency format (`name@org`)
  - Dependencies: 1.1

- [ ] **1.3** Add version constraints (1h)
  - Files: `internal/plugin/manifest.go`
  - Support version ranges in dependencies (`skills@go-ent:^1.0.0`)
  - Parse and validate version constraints
  - Dependencies: 1.1

## Phase 2: Dependency Resolution (4h)

- [ ] **2.1** Create dependency resolver (2h)
  - Files: `internal/marketplace/resolve.go` - **NEW**
  - Topological sort for install order
  - Detect conflicts
  - Dependencies: 1.2

- [ ] **2.2** Update install command (1h)
  - Files: `internal/marketplace/install.go`
  - Install dependencies recursively
  - Use resolver for correct order
  - Dependencies: 2.1

- [ ] **2.3** Add dependency graph visualization (1h)
  - Files: `internal/plugin/graph.go` - **NEW**
  - Generate dependency graph (for debugging)
  - Export to DOT format
  - Dependencies: 2.1

## Phase 3: Plugin Package Creation (5h)

- [ ] **3.1** Create skills package (1h)
  - Files: `plugins/packages/skills/plugin.yaml`, `skills/`
  - Move skills from monolithic plugin
  - No dependencies
  - Dependencies: 2.2
  - Parallel with: 3.2, 3.4

- [ ] **3.2** Create hooks package (1h)
  - Files: `plugins/packages/hooks/plugin.yaml`, `hooks/`
  - Move hooks from monolithic plugin
  - No dependencies
  - Dependencies: 2.2
  - Parallel with: 3.1, 3.4

- [ ] **3.3** Create agents package (1.5h)
  - Files: `plugins/packages/agents/plugin.yaml`, `agents/`
  - Move agents from monolithic plugin
  - Add dependency: `skills@go-ent`
  - Dependencies: 3.1

- [ ] **3.4** Create commands package (1h)
  - Files: `plugins/packages/commands/plugin.yaml`, `commands/`
  - Move commands from monolithic plugin
  - Add dependency: `agents@go-ent`
  - Dependencies: 3.3

- [ ] **3.5** Create meta-package (0.5h)
  - Files: `plugins/packages/go-ent/plugin.yaml`
  - Empty plugin with dependencies on all 4 packages
  - Dependencies: 3.1, 3.2, 3.3, 3.4

## Phase 4: Cross-Plugin References (2h)

- [ ] **4.1** Update skill references in agents (1h)
  - Files: Agent meta YAML files
  - Change `skills: [go-code]` → `skills: [skills@go-ent:go-code]`
  - Fully qualified skill names
  - Dependencies: 3.3

- [ ] **4.2** Update agent references in commands (1h)
  - Files: Command markdown files
  - Change `@ent:architect` → `@agents@go-ent:architect`
  - Fully qualified agent names
  - Dependencies: 3.4

## Phase 5: Testing & Migration (2h)

- [ ] **5.1** Unit tests (1h)
  - Files: `internal/marketplace/resolve_test.go`, `internal/plugin/validate_test.go`
  - Test dependency resolution
  - Test circular dependency detection
  - Test version constraints
  - Dependencies: 2.1, 1.2

- [ ] **5.2** Integration tests (0.5h)
  - Files: `internal/marketplace/install_test.go`
  - Test installing `go-ent` installs all dependencies
  - Test installing individual packages
  - Dependencies: 3.5, 5.1

- [ ] **5.3** Migration guide (0.5h)
  - Files: `docs/MIGRATION.md` - **NEW**
  - Document migration from monolithic to split plugins
  - Provide uninstall/reinstall commands
  - Dependencies: 5.2

---

## Critical Path

```
1.1 → [1.2, 1.3] → 2.1 → 2.2 → [3.1, 3.2] → 3.3 → 3.4 → 3.5 → [4.1, 4.2] → 5.1 → 5.2 → 5.3
```

**Parallel work:**
- Tasks 1.2 and 1.3 can run in parallel after 1.1
- Tasks 3.1, 3.2, and 3.4 can run in parallel after 2.2
- Tasks 4.1 and 4.2 can run in parallel after 3.5
