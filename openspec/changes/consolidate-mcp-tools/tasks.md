## Tasks

### 1. Design Consolidated Tool API

- [ ] **1.1** Define unified spec tool interface
  - Files: `docs/MCP_TOOLS_V2.md`
  - Dependencies: none
  - Effort: 1h

- [ ] **1.2** Map old tools to new consolidated set
  - Files: `docs/MCP_TOOLS_MIGRATION.md`
  - Dependencies: 1.1
  - Effort: 1h

### 2. Implement Core Spec Tools

- [ ] **2.1** Implement spec_init tool
  - Files: `internal/mcp/tools/spec.go` (new)
  - Dependencies: 1.2
  - Effort: 1h

- [ ] **2.2** Implement spec_list tool (merge registry_list)
  - Files: `internal/mcp/tools/spec.go`
  - Dependencies: 2.1
  - Effort: 2h

- [ ] **2.3** Implement spec_show tool
  - Files: `internal/mcp/tools/spec.go`
  - Dependencies: 2.2
  - Effort: 1h

- [ ] **2.4** Implement spec_create tool
  - Files: `internal/mcp/tools/spec.go`
  - Dependencies: 2.3
  - Effort: 1h

- [ ] **2.5** Implement spec_update tool (merge registry_update, state_update)
  - Files: `internal/mcp/tools/spec.go`
  - Dependencies: 2.4
  - Effort: 2h

- [ ] **2.6** Implement spec_validate tool
  - Files: `internal/mcp/tools/spec.go`
  - Dependencies: 2.5
  - Effort: 1h

- [ ] **2.7** Implement spec_archive tool
  - Files: `internal/mcp/tools/spec.go`
  - Dependencies: 2.6
  - Effort: 1h

- [ ] **2.8** Implement spec_sync tool (merge registry_sync, state_sync)
  - Files: `internal/mcp/tools/spec.go`
  - Dependencies: 2.7
  - Effort: 2h

### 3. Remove Deprecated Tools

- [ ] **3.1** Remove registry tools (registry_list, registry_next, registry_update, registry_deps, registry_sync, registry_init)
  - Files: `internal/mcp/tools/registry.go`
  - Dependencies: 2.8
  - Effort: 1h

- [ ] **3.2** Remove workflow tools
  - Files: `internal/mcp/tools/workflow.go`
  - Dependencies: 3.1
  - Effort: 1h

- [ ] **3.3** Remove state tools
  - Files: `internal/mcp/tools/state.go`
  - Dependencies: 3.2
  - Effort: 1h

- [ ] **3.4** Remove skill tools
  - Files: `internal/mcp/tools/skill_*.go`
  - Dependencies: 3.3
  - Effort: 1h

- [ ] **3.5** Remove generate and loop tools
  - Files: `internal/mcp/tools/generate.go`, `internal/mcp/tools/loop.go`
  - Dependencies: 3.4
  - Effort: 1h

### 4. Update Server Registration

- [ ] **4.1** Update tool registration in server
  - Files: `internal/mcp/server/server.go`, `internal/mcp/tools/tools.go`
  - Dependencies: 3.5
  - Effort: 1h

### 5. Update Documentation

- [ ] **5.1** Update MCP_API.md with new tool set
  - Files: `docs/MCP_API.md`
  - Dependencies: 4.1
  - Effort: 1h

- [ ] **5.2** Update agent prompts to use new tools
  - Files: `prompts/shared/tooling.md`
  - Dependencies: 5.1
  - Effort: 1h

---

**Total Tasks**: 17
**Estimated Effort**: ~18 hours
