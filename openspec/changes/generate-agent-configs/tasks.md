## Tasks

### 1. Design Generation System

- [ ] **1.1** Design agent generation architecture
  - Files: `docs/AGENT_GENERATION.md`
  - Dependencies: none
  - Effort: 1h

- [ ] **1.2** Define agent template structure
  - Files: `internal/generate/templates/`
  - Dependencies: 1.1
  - Effort: 2h

### 2. Project Detection

- [ ] **2.1** Implement project type detection
  - Files: `internal/generate/detect.go`
  - Dependencies: 1.2
  - Effort: 2h

- [ ] **2.2** Add detection for Go projects
  - Files: `internal/generate/detect.go`
  - Dependencies: 2.1
  - Effort: 1h

- [ ] **2.3** Add detection for web projects
  - Files: `internal/generate/detect.go`
  - Dependencies: 2.2
  - Effort: 1h

- [ ] **2.4** Write detection tests
  - Files: `internal/generate/detect_test.go`
  - Dependencies: 2.3
  - Effort: 1h

### 3. Claude Code Generator

- [ ] **3.1** Implement Claude Code agent template
  - Files: `internal/generate/claude.go`
  - Dependencies: 1.2
  - Effort: 2h

- [ ] **3.2** Implement agent generation for Claude Code
  - Files: `internal/generate/claude.go`
  - Dependencies: 3.1
  - Effort: 2h

- [ ] **3.3** Write Claude generator tests
  - Files: `internal/generate/claude_test.go`
  - Dependencies: 3.2
  - Effort: 1h

### 4. OpenCode Generator

- [ ] **4.1** Implement OpenCode agent template
  - Files: `internal/generate/opencode.go`
  - Dependencies: 1.2
  - Effort: 2h

- [ ] **4.2** Implement agent generation for OpenCode
  - Files: `internal/generate/opencode.go`
  - Dependencies: 4.1
  - Effort: 2h

- [ ] **4.3** Write OpenCode generator tests
  - Files: `internal/generate/opencode_test.go`
  - Dependencies: 4.2
  - Effort: 1h

### 5. Generate Command

- [ ] **5.1** Create generate command structure
  - Files: `internal/cli/generate.go`
  - Dependencies: 3.3, 4.3
  - Effort: 1h

- [ ] **5.2** Implement `ent generate agents` command
  - Files: `internal/cli/generate.go`
  - Dependencies: 5.1
  - Effort: 2h

- [ ] **5.3** Add platform flags (--claude, --opencode, --both)
  - Files: `internal/cli/generate.go`
  - Dependencies: 5.2
  - Effort: 1h

- [ ] **5.4** Add output directory flag
  - Files: `internal/cli/generate.go`
  - Dependencies: 5.3
  - Effort: 1h

- [ ] **5.5** Register generate command in CLI
  - Files: `internal/cli/root.go`
  - Dependencies: 5.4
  - Effort: 0.5h

### 6. Configuration

- [ ] **6.1** Add agent config section to Config
  - Files: `internal/config/config.go`
  - Dependencies: 5.5
  - Effort: 1h

- [ ] **6.2** Load agent configuration from .go-ent/config.yaml
  - Files: `internal/config/loader.go`
  - Dependencies: 6.1
  - Effort: 1h

### 7. Remove Static Agents

- [ ] **7.1** Generate agents for go-ent project
  - Files: `.claude/agents/`, `.opencode/agents/`
  - Dependencies: 5.5
  - Effort: 1h

- [ ] **7.2** Remove static agent files from internal/cli/
  - Files: `internal/cli/.claude/agents/`, `internal/cli/.opencode/agents/`
  - Dependencies: 7.1
  - Effort: 0.5h

- [ ] **7.3** Update embed.go if needed
  - Files: `embed.go`
  - Dependencies: 7.2
  - Effort: 0.5h

### 8. Documentation

- [ ] **8.1** Document generate command
  - Files: `docs/CLI_REFERENCE.md`
  - Dependencies: 7.3
  - Effort: 1h

- [ ] **8.2** Update AGENTS_AND_SKILLS.md
  - Files: `docs/AGENTS_AND_SKILLS.md`
  - Dependencies: 8.1
  - Effort: 1h

---

**Total Tasks**: 24
**Estimated Effort**: ~26 hours
