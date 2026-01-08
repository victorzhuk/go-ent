# Tasks: Repository Restructure

## 1. File Movement

### 1.1 Move tool handlers
- [x] Move `cmd/go-ent/internal/tools/` to `internal/mcp/tools/`
- [x] Verify all files moved (14 tool handler files)

### 1.2 Move server package
- [x] Move `cmd/go-ent/internal/server/` to `internal/mcp/server/`
- [x] Verify server.go moved

### 1.3 Move version package
- [x] Move `cmd/go-ent/internal/version/` to `internal/version/`
- [x] Verify version.go moved

### 1.4 Remove old directories
- [x] Delete empty `cmd/go-ent/internal/` directory
- [x] Verify no files left behind

## 2. Import Path Updates

### 2.1 Update cmd/go-ent/main.go
- [x] Replace `cmd/go-ent/internal/server` with `internal/mcp/server`
- [x] Replace `cmd/go-ent/internal/version` with `internal/version`
- [x] Verify imports

### 2.2 Update internal/mcp/server/server.go
- [x] Replace `cmd/go-ent/internal/tools` with `internal/mcp/tools`
- [x] Verify imports

### 2.3 Update internal/mcp/tools/register.go
- [x] Update package imports if needed
- [x] Verify all tool registrations work

### 2.4 Update all tool handler files
- [x] Update imports in init.go
- [x] Update imports in list.go
- [x] Update imports in show.go
- [x] Update imports in crud.go
- [x] Update imports in registry.go
- [x] Update imports in workflow.go
- [x] Update imports in loop.go
- [x] Update imports in validate.go
- [x] Update imports in archive.go
- [x] Update imports in generate*.go files
- [x] Verify all tool files compile

### 2.5 Update test files
- [x] Find all *_test.go files with old imports
- [x] Update import paths
- [x] Verify tests compile

## 3. Verification

### 3.1 Build verification
- [x] Run `go build ./cmd/go-ent`
- [x] Verify no compilation errors
- [x] Check binary size is similar

### 3.2 Test verification
- [x] Run `go test ./internal/mcp/...`
- [x] Run `go test ./...` (full suite)
- [x] Verify all tests pass

### 3.3 Runtime verification
- [x] Start MCP server: `./dist/go-ent serve`
- [x] Test with Claude Code plugin
- [x] Verify tools respond correctly
- [x] Check logs for errors

### 3.4 Code search verification
- [x] Search for `cmd/go-ent/internal/` in codebase
- [x] Verify no references remain
- [x] Check for broken imports

## 4. Documentation

### 4.1 Update README
- [x] Update import examples if present
- [x] Update directory structure documentation

### 4.2 Update CLAUDE.md
- [x] Update any references to internal structure
- [x] Verify openspec instructions still accurate

### 4.3 Update go.mod
- [x] Run `go mod tidy`
- [x] Verify module dependencies

## 5. Commit

### 5.1 Stage changes
- [x] Stage all moved files
- [x] Stage all modified files
- [x] Review diff

### 5.2 Commit
- [x] Create commit with clear message
- [x] Reference this proposal in commit message
