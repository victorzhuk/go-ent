# Tasks: Modernize Templates and CLI to Go 2026 Standards

## Dependencies

```
T001 -> T002, T003
T002, T003 -> T004
T004 -> T005
T005 -> T006
T004 -> T007 [P]
T006, T007 -> T008
T008 -> T009
T009 -> T010
```

**Legend**: `[P]` = Can run in parallel with previous task

## Phase 1: Project Root Upgrades

### T001: Update root go.mod to Go 1.25.5
**File**: `/go.mod`
**Depends**: None
**Parallel**: No

- [x] 1.1 Change `go 1.23.0` to `go 1.25.5`
- [x] 1.2 Run `go mod tidy`
- [x] 1.3 Verify build succeeds: `go build ./...`
- [x] 1.4 Commit: "chore: upgrade to Go 1.25.5"

### T002: Refactor cmd/goent/main.go to run() pattern
**File**: `/cmd/goent/main.go`
**Depends**: T001
**Parallel**: No

- [x] 2.1 Add imports: `fmt`, `io`, `log/slog`, `os`, `os/signal`, `syscall`, `time`
- [x] 2.2 Create `run(ctx context.Context, getenv func(string) string, stdout, stderr io.Writer) error`
- [x] 2.3 Add `setupLogger(level, format string, w io.Writer) *slog.Logger` function
- [x] 2.4 Move MCP server setup into `run()` function body
- [x] 2.5 Add `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)`
- [x] 2.6 Start MCP server in goroutine with error channel
- [x] 2.7 Add select block for error vs signal
- [x] 2.8 Add graceful shutdown with `context.WithTimeout(context.Background(), 30*time.Second)`
- [x] 2.9 Replace `log.Fatal` with proper error return
- [x] 2.10 Update `main()` to call `run()` and log error before `os.Exit(1)`
- [x] 2.11 Test with `kill -SIGTERM` to verify graceful shutdown
- [x] 2.12 Commit: "refactor: implement run() pattern in cmd/goent/main.go"

### T003: Enhance root Makefile with VERSION/VCS_REF
**File**: `/Makefile`
**Depends**: T001
**Parallel**: Yes (with T002)

- [x] 3.1 Add `VCS_REF ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")`
- [x] 3.2 Update VERSION line to include fallback: `VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")`
- [x] 3.3 Add LDFLAGS: `LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.vcsRef=$(VCS_REF)"`
- [x] 3.4 Update build target to use LDFLAGS: `go build $(LDFLAGS) -o dist/goent ./cmd/goent`
- [x] 3.5 Verify `make build` passes
- [x] 3.6 Verify `./dist/goent --version` shows version (if implemented)
- [x] 3.7 Commit: "build: add VERSION and VCS_REF to Makefile"

## Phase 2: Template Upgrades

### T004: Update templates/go.mod.tmpl to Go 1.25.5
**File**: `/templates/go.mod.tmpl`
**Depends**: T002, T003
**Parallel**: No

- [x] 4.1 Change `go 1.23` to `go 1.25.5`
- [x] 4.2 Keep `github.com/caarlos0/env/v11 v11.3.1` (already latest)
- [x] 4.3 Commit: "templates: upgrade to Go 1.25.5"

### T005: Enhance templates/Makefile.tmpl
**File**: `/templates/Makefile.tmpl`
**Depends**: T004
**Parallel**: No

- [x] 5.1 Add `VCS_REF ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")`
- [x] 5.2 Update VERSION line: `VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")`
- [x] 5.3 Update BUILD_DIR to `./bin`
- [x] 5.4 Update build target output: `$(BUILD_DIR)/$(APP_NAME)` → `bin/$(APP_NAME)`
- [x] 5.5 Update docker target: `docker build --build-arg VERSION=$(VERSION) --build-arg VCS_REF=$(VCS_REF) -t $(APP_NAME):$(VERSION) -f build/Dockerfile .`
- [x] 5.6 Ensure LDFLAGS includes VCS_REF: `-X main.vcsRef=$(VCS_REF)`
- [x] 5.7 Commit: "templates: add VERSION/VCS_REF to Makefile template"

### T006: Rewrite templates/build/Dockerfile.tmpl to distroless
**File**: `/templates/build/Dockerfile.tmpl`
**Depends**: T005
**Parallel**: No

- [x] 6.1 Change builder FROM: `golang:1.23-alpine` → `golang:1.25.5-trixie`
- [x] 6.2 Add ARG lines: `ARG VERSION=local` and `ARG VCS_REF=unknown`
- [x] 6.3 Add ENV lines: `ENV GOOS=linux` and `ENV GOARCH=amd64`
- [x] 6.4 Replace alpine apk with: `RUN apt-get update && apt-get install -y bash-static && rm -rf /var/lib/apt/lists/*`
- [x] 6.5 Update build command: `RUN VERSION=${VERSION} make build`
- [x] 6.6 Change runtime FROM: `alpine:3.19` → `gcr.io/distroless/static-debian13:nonroot`
- [x] 6.7 Remove alpine apk install (distroless has no apk)
- [x] 6.8 Add COPY for bash: `COPY --from=builder /bin/bash-static /usr/bin/bash`
- [x] 6.9 Update binary COPY: `COPY --from=builder /app/bin/{{PROJECT_NAME}} /{{PROJECT_NAME}}`
- [x] 6.10 Remove USER line (nonroot image already runs as non-root)
- [x] 6.11 Update ENTRYPOINT: `ENTRYPOINT ["/usr/bin/bash"]`
- [x] 6.12 Update CMD: `CMD ["-c", "/{{PROJECT_NAME}} serve"]`
- [x] 6.13 Remove HEALTHCHECK (use k8s probes instead)
- [x] 6.14 Commit: "templates: rewrite Dockerfile to use distroless with bash-static"

### T007: Update templates/.golangci.yml.tmpl for Go 1.25
**File**: `/templates/.golangci.yml.tmpl`
**Depends**: T004
**Parallel**: Yes (with T005, T006)

- [x] 7.1 Update `run.go` to `"1.25"`
- [x] 7.2 Add `run.go-version` to `"1.25.5"`
- [x] 7.3 Add copyloopvar to enabled linters list
- [x] 7.4 Review deprecated linters (check golangci-lint v2.4.0 docs)
- [x] 7.5 Commit: "templates: update golangci-lint config for Go 1.25"

## Phase 3: MCP Server Templates (NEW)

### T008: Create MCP server template structure
**Files**:
- `/templates/mcp/go.mod.tmpl`
- `/templates/mcp/Makefile.tmpl`
- `/templates/mcp/build/Dockerfile.tmpl`
- `/templates/mcp/cmd/server/main.go.tmpl`
- `/templates/mcp/internal/server/server.go.tmpl`

**Depends**: T006, T007
**Parallel**: No

- [x] 8.1 Create directory structure: `mkdir -p templates/mcp/{build,cmd/server,internal/server}`
- [x] 8.2 Create `mcp/go.mod.tmpl` with MCP SDK dependency:
  ```
  module {{MODULE_PATH}}

  go 1.25.5

  require (
      github.com/modelcontextprotocol/go-sdk v1.2.0
  )
  ```
- [x] 8.3 Create `mcp/Makefile.tmpl` with MCP-specific targets (similar to HTTP Makefile)
- [x] 8.4 Create `mcp/build/Dockerfile.tmpl` (same as HTTP Dockerfile, stdio transport focus)
- [x] 8.5 Create `mcp/cmd/server/main.go.tmpl` with MCP-specific run() pattern:
  - Import `github.com/modelcontextprotocol/go-sdk/mcp`
  - Create MCP server in run()
  - Use StdioTransport
  - Handle graceful shutdown
- [x] 8.6 Create `mcp/internal/server/server.go.tmpl` with MCP SDK setup:
  - `type Server struct` with mcp.Server
  - Tool registration function
  - New() constructor
- [x] 8.7 Test template generation manually
- [x] 8.8 Commit: "templates: add MCP server template variant"

### T009: Update scaffold logic to support template_type
**File**: Implementation-specific (tools/scaffold.go or equivalent)
**Depends**: T008
**Parallel**: No

- [x] 9.1 Identify scaffold/init tool implementation location
- [x] 9.2 Add `template_type` parameter (enum: `http`, `mcp`)
- [x] 9.3 Implement template directory selection logic:
  - `http` → use `templates/` (default)
  - `mcp` → use `templates/mcp/`
- [x] 9.4 Update tool documentation/help text
- [x] 9.5 Test HTTP template generation: verify files created correctly
- [x] 9.6 Test MCP template generation: verify MCP-specific files created
- [x] 9.7 Commit: "feat: add template_type parameter for HTTP vs MCP scaffolding"

## Phase 4: Validation

### T010: Verify all changes
**Depends**: T009
**Parallel**: No

- [x] 10.1 Run `make lint` in project root - must pass
- [x] 10.2 Run `make test` in project root - must pass
- [x] 10.3 Generate HTTP project with templates:
  - [ ] 10.3.1 Verify `go.mod` shows `go 1.25.5`
  - [ ] 10.3.2 Verify `make build` succeeds
  - [ ] 10.3.3 Verify binary name is `{{PROJECT_NAME}}`
  - [ ] 10.3.4 Verify `make docker` builds successfully
  - [ ] 10.3.5 Verify Docker image uses distroless base
  - [ ] 10.3.6 Verify bash-static is present: `docker run <image> /usr/bin/bash --version`
- [x] 10.4 Generate MCP project with templates:
  - [ ] 10.4.1 Verify `go.mod` includes MCP SDK v1.2.0
  - [ ] 10.4.2 Verify `make build` succeeds
  - [ ] 10.4.3 Verify binary name is `{{PROJECT_NAME}}`
  - [ ] 10.4.4 Verify `make docker` builds successfully
  - [ ] 10.4.5 Test MCP server connects via stdio
- [x] 10.5 Test graceful shutdown:
  - [ ] 10.5.1 Start goent binary
  - [ ] 10.5.2 Send SIGTERM: `kill -TERM <pid>`
  - [ ] 10.5.3 Verify "shutdown signal received" log
  - [ ] 10.5.4 Verify process exits within 30s
- [x] 10.6 Run `openspec validate refactor-modernize-templates-and-cli --strict`
- [x] 10.7 Verify all validation errors resolved
- [x] 10.8 Final commit: "chore: validate modernization complete"

## Completion Criteria

All tasks marked complete AND:
- [x] No golangci-lint errors
- [x] No test failures
- [x] HTTP template generates working project
- [x] MCP template generates working project
- [x] Docker images use distroless/static-debian13:nonroot
- [x] Graceful shutdown works correctly
- [x] OpenSpec validation passes
