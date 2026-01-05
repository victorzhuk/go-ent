# Implementation Tasks

## Phase 1: File System Renames

Atomic operations - must complete first.

### 1.1 Directory Rename
- [x] 1.1.1 `git mv cmd/go-ent cmd/go-ent`
- [x] 1.1.2 Verify no broken imports: `find cmd/go-ent/ -name "*.go" | head`

### 1.2 Command Files (17 files)
- [x] 1.2.1 Rename all `plugins/go-ent/commands/go-ent:*.md` → `go-ent:*.md`
- [x] 1.2.2 Verify: `ls plugins/go-ent/commands/go-ent:*.md | wc -l` (expect 17)

### 1.3 Agent Files (7 files)
- [x] 1.3.1 Rename all `plugins/go-ent/agents/go-ent:*.md` → `go-ent:*.md`
- [x] 1.3.2 Verify: `ls plugins/go-ent/agents/go-ent:*.md | wc -l` (expect 7)

## Phase 2: Go Source Code

Depends on Phase 1 completion.

### 2.1 Import Paths (44 files)
- [x] 2.1.1 Update all imports: `cmd/go-ent` → `cmd/go-ent`
- [x] 2.1.2 Verify: `go mod tidy`
- [x] 2.1.3 Verify: `go build ./...`

### 2.2 MCP Server Name
- [x] 2.2.1 Update `cmd/go-ent/internal/server/server.go:12`: `Name: "go-ent"`

### 2.3 MCP Tool Names (24 tools)
- [x] 2.3.1 Update all tool registrations: `go_ent_*` → `go_ent_*`
- [x] 2.3.2 Update error messages referencing tool names
- [x] 2.3.3 Verify: `grep -r "go_ent_" cmd/go-ent/` returns 0 matches

### 2.4 Version Output
- [x] 2.4.1 Update `cmd/go-ent/main.go:23`: `fmt.Printf("go-ent %s\n", ...)`

### 2.5 Build Verification
- [x] 2.5.1 Run: `go build ./...`
- [x] 2.5.2 Run: `go test ./...`
- [x] 2.5.3 Run: `make lint`

## Phase 3: Configuration Files

### 3.1 Makefile
- [x] 3.1.1 Update VERSION_PKG path: `cmd/go-ent` → `cmd/go-ent`
- [x] 3.1.2 Update build output: `dist/go-ent` → `dist/go-ent`
- [x] 3.1.3 Update build input: `./cmd/go-ent` → `./cmd/go-ent`
- [x] 3.1.4 Update clean target: `goent` → `go-ent`
- [x] 3.1.5 Update messages referencing binary name

### 3.2 .goreleaser.yaml
- [x] 3.2.1 Update `project_name: goent` → `project_name: go-ent`
- [x] 3.2.2 Update build id: `goent` → `go-ent`
- [x] 3.2.3 Update main path: `./cmd/go-ent` → `./cmd/go-ent`
- [x] 3.2.4 Update binary name: `goent` → `go-ent`
- [x] 3.2.5 Update ldflags paths

### 3.3 Plugin Configuration
- [x] 3.3.1 Update `plugins/go-ent/.claude-plugin/plugin.json`:
  - `"name": "goent"` → `"name": "go-ent"`
  - `"command": "../../dist/go-ent"` → `"../../dist/go-ent"`
- [x] 3.3.2 Update `.claude-plugin/marketplace.json`: `"name": "goent"` → `"name": "go-ent"`

### 3.4 .gitignore
- [x] 3.4.1 Update: `cmd/go-ent/templates/` → `cmd/go-ent/templates/`

### 3.5 Build Verification
- [x] 3.5.1 Run: `make clean && make build`
- [x] 3.5.2 Test: `./dist/go-ent version`

## Phase 4: Documentation

### 4.1 Batch Pattern Replacements
- [x] 4.1.1 Commands: `/go-ent:` → `/go-ent:`
- [x] 4.1.2 MCP tools: `go_ent_` → `go_ent_`
- [x] 4.1.3 MCP namespace: `mcp__go_ent__` → `mcp__go_ent__`
- [x] 4.1.4 Binary paths: `dist/go-ent` → `dist/go-ent`
- [x] 4.1.5 Source paths: `cmd/go-ent` → `cmd/go-ent`

### 4.2 Primary Documentation (Manual Review)
- [x] 4.2.1 Update CLAUDE.md (lines 25-73)
- [x] 4.2.2 Update README.md
- [x] 4.2.3 Update docs/DEVELOPMENT.md
- [x] 4.2.4 Update plugins/go-ent/README.md
- [x] 4.2.5 Update openspec/AGENTS.md

### 4.3 Command/Agent Frontmatter
- [x] 4.3.1 Update `allowed-tools` in all command .md files
- [x] 4.3.2 Update `allowed-tools` in all agent .md files

### 4.4 Verification
- [x] 4.4.1 Grep check: `grep -r "goent:" . --exclude-dir=.git --exclude-dir=dist` (expect 0)
- [x] 4.4.2 Grep check: `grep -r "go_ent_" . --exclude-dir=.git --exclude-dir=dist --include="*.go"` (expect 0)
- [x] 4.4.3 Grep check: `grep -r "mcp__go_ent__" . --exclude-dir=.git --include="*.md"` (expect 0)

## Phase 5: Validation & Release

### 5.1 Integration Testing
- [x] 5.1.1 Restart Claude Code
- [x] 5.1.2 Verify: `/go-ent:` autocomplete shows 17 commands
- [x] 5.1.3 Verify: `@go-ent:` shows 7 agents
- [x] 5.1.4 Test: `/go-ent:init test-project` creates openspec/
- [x] 5.1.5 Verify MCP tools: `go_ent_spec_list` accessible

### 5.2 Migration Documentation
- [x] 5.2.1 Create `MIGRATION.md` with breaking changes and user actions
- [x] 5.2.2 Update `CHANGELOG.md` with v4.0.0 entry

### 5.3 Release
- [x] 5.3.1 Commit all changes
- [x] 5.3.2 Tag: `git tag -a v4.0.0 -m "BREAKING: Unify naming to go-ent"`
- [x] 5.3.3 Archive proposal: `openspec archive refactor-unify-naming-go-ent --skip-specs --yes`
