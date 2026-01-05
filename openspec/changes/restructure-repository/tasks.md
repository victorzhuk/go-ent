# Tasks: Repository Restructure

## 1. File Movement

### 1.1 Move tool handlers
- [ ] Move `cmd/go-ent/internal/tools/` to `internal/mcp/tools/`
- [ ] Verify all files moved (15+ tool handler files)

### 1.2 Move server package
- [ ] Move `cmd/go-ent/internal/server/` to `internal/mcp/server/`
- [ ] Verify server.go moved

### 1.3 Move version package
- [ ] Move `cmd/go-ent/internal/version/` to `internal/version/`
- [ ] Verify version.go moved

### 1.4 Remove old directories
- [ ] Delete empty `cmd/go-ent/internal/` directory
- [ ] Verify no files left behind

## 2. Import Path Updates

### 2.1 Update cmd/go-ent/main.go
- [ ] Replace `cmd/go-ent/internal/server` with `internal/mcp/server`
- [ ] Replace `cmd/go-ent/internal/version` with `internal/version`
- [ ] Verify imports

### 2.2 Update internal/mcp/server/server.go
- [ ] Replace `cmd/go-ent/internal/tools` with `internal/mcp/tools`
- [ ] Verify imports

### 2.3 Update internal/mcp/tools/register.go
- [ ] Update package imports if needed
- [ ] Verify all tool registrations work

### 2.4 Update all tool handler files
- [ ] Update imports in init.go
- [ ] Update imports in list.go
- [ ] Update imports in show.go
- [ ] Update imports in crud.go
- [ ] Update imports in registry.go
- [ ] Update imports in workflow.go
- [ ] Update imports in loop.go
- [ ] Update imports in validate.go
- [ ] Update imports in archive.go
- [ ] Update imports in generate*.go files
- [ ] Verify all tool files compile

### 2.5 Update test files
- [ ] Find all *_test.go files with old imports
- [ ] Update import paths
- [ ] Verify tests compile

## 3. Verification

### 3.1 Build verification
- [ ] Run `go build ./cmd/go-ent`
- [ ] Verify no compilation errors
- [ ] Check binary size is similar

### 3.2 Test verification
- [ ] Run `go test ./internal/mcp/...`
- [ ] Run `go test ./...` (full suite)
- [ ] Verify all tests pass

### 3.3 Runtime verification
- [ ] Start MCP server: `./dist/go-ent serve`
- [ ] Test with Claude Code plugin
- [ ] Verify tools respond correctly
- [ ] Check logs for errors

### 3.4 Code search verification
- [ ] Search for `cmd/go-ent/internal/` in codebase
- [ ] Verify no references remain
- [ ] Check for broken imports

## 4. Documentation

### 4.1 Update README
- [ ] Update import examples if present
- [ ] Update directory structure documentation

### 4.2 Update CLAUDE.md
- [ ] Update any references to internal structure
- [ ] Verify openspec instructions still accurate

### 4.3 Update go.mod
- [ ] Run `go mod tidy`
- [ ] Verify module dependencies

## 5. Commit

### 5.1 Stage changes
- [ ] Stage all moved files
- [ ] Stage all modified files
- [ ] Review diff

### 5.2 Commit
- [ ] Create commit with clear message
- [ ] Reference this proposal in commit message
