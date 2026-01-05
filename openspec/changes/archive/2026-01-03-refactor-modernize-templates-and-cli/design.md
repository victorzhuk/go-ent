# Design: Modernize Templates and CLI to Go 2026 Standards

## Context

### Go 1.25.5 Release (August 2025)

Key features ([Go 1.25 Release Notes](https://go.dev/doc/go1.25)):
- **DWARF v5 debug information** - reduces binary debug info size, faster linking
- **Experimental JSON v2** - `encoding/json/v2` with better performance
- **testing/synctest** - deterministic concurrent test execution
- **New garbage collector** - 10-40% reduction in GC overhead
- **FlightRecorder API** - runtime tracing for production profiling

Security fixes in Go 1.25.5 ([Security Release](https://groups.google.com/g/golang-announce/c/8FJoBkPddm4)):
- CVE-2025-61729 - crypto/x509 wildcard SAN handling
- CVE-2025-61727 - crypto/x509 certificate validation

### Distroless Images

[GoogleContainerTools/distroless](https://github.com/GoogleContainerTools/distroless) provides:
- **No shell, package manager, or debug tools** - minimal attack surface
- **Non-root by default** - security best practice
- **Static variant** - ~2MB base image vs alpine's ~7MB
- **Debian 13 support** - latest stable Debian base

However, **bash-static is required** for:
- ENTRYPOINT script flexibility
- Health check commands
- Debugging capabilities when needed

## Goals

### Must Have
1. Go 1.25.5 across project and templates
2. Distroless production runtime images
3. Proper signal handling (SIGTERM, SIGINT, SIGQUIT)
4. 30-second graceful shutdown on fresh `context.Background()`
5. MCP server template option
6. Reproducible builds with VERSION/VCS_REF

### Nice to Have
- Automated migration tool for existing projects
- Multi-arch Docker builds (amd64, arm64)
- Health check probe templates

### Non-Goals
- HTTP/3 or QUIC support
- Breaking changes to MCP SDK interfaces
- Generic plugin system for templates
- Kubernetes manifests (future change)

## Architectural Decisions

### D1: Dockerfile Base Image Strategy

**Decision**: Use `golang:1.25.5-trixie` for builder, `gcr.io/distroless/static-debian13:nonroot` for runtime

**Options Considered**:
| Option | Pros | Cons | Decision |
|--------|------|------|----------|
| `golang:1.25.5-alpine` | Small (50MB), musl libc | CVE-prone, bash issues | ❌ Rejected |
| `golang:1.25.5-trixie` | Latest toolchain, glibc, bash-static | Larger builder (800MB) | ✅ **Selected** |
| `scratch` | Absolute minimum | No CA certs, no timezone | ❌ Rejected |
| `distroless/static-debian13` | Minimal (2MB), secure | No shell | ✅ **Selected for runtime** |

**Rationale**:
- Trixie (Debian 13) matches distroless-debian13 for glibc compatibility
- Builder size doesn't matter (discarded after build)
- Runtime security and size are critical

**Trade-offs**:
- Build time increases ~10s for bash-static installation
- Runtime image grows from 2MB to ~4MB with bash-static

### D2: bash-static for ENTRYPOINT Flexibility

**Decision**: Copy `/bin/bash-static` from builder to `/usr/bin/bash` in runtime image

**Rationale**:
```dockerfile
ENTRYPOINT ["/usr/bin/bash"]
CMD ["-c", "/{{PROJECT_NAME}} serve"]
```

Enables:
- Command-line argument parsing without rebuilding
- Environment variable substitution in CMD
- Health check scripts when needed
- Debugging with `docker exec` (bash available)

**Alternatives Rejected**:
- Direct binary ENTRYPOINT - inflexible for args
- Dynamic bash installation - violates immutability
- busybox - larger, more CVE surface

### D3: Template Structure for MCP vs HTTP

**Decision**: Separate `templates/mcp/` directory with distinct file structure

**Structure**:
```
templates/
├── http (default templates, no prefix)
│   ├── cmd/server/main.go.tmpl
│   ├── internal/app/app.go.tmpl
│   └── ...
└── mcp/ (NEW)
    ├── cmd/server/main.go.tmpl
    ├── internal/server/server.go.tmpl
    ├── go.mod.tmpl (with MCP SDK)
    ├── Makefile.tmpl (mcp-specific targets)
    └── build/Dockerfile.tmpl (stdio transport)
```

**Rationale**:
- Avoids conditional logic pollution in templates
- Clear separation of concerns
- Different dependencies (HTTP vs MCP SDK)
- Different ENTRYPOINT patterns

**Alternatives Rejected**:
- Single template with `{{if .IsMCP}}` conditionals - hard to maintain
- Separate repo - overkill, fragmentation

### D4: cmd/go-ent/main.go Refactor Pattern

**Decision**: Follow exact pattern from `templates/cmd/server/main.go.tmpl`

**Pattern**:
```go
func main() {
    if err := run(context.Background(), os.Getenv, os.Stdout, os.Stderr); err != nil {
        slog.Error("startup failed", "error", err)
        os.Exit(1)
    }
}

func run(ctx context.Context, getenv func(string) string, stdout, stderr io.Writer) error {
    // Setup logger
    logger := setupLogger(getenv("LOG_LEVEL"), getenv("LOG_FORMAT"), stdout)
    slog.SetDefault(logger)

    // Signal handling
    ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
    defer cancel()

    // Start server in goroutine
    errCh := make(chan error, 1)
    go func() { errCh <- server.Start(ctx) }()

    // Wait for error or signal
    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        logger.Info("shutdown signal received")
    }

    // Graceful shutdown on fresh context
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()

    return server.Shutdown(shutdownCtx)
}
```

**Rationale**:
- **Consistency** with generated projects
- **Testability** via dependency injection (`getenv`, `stdout`, `stderr`)
- **No globals** except `slog.SetDefault`
- **Proper shutdown** with fresh context (parent is cancelled)

**Key Details**:
- `context.Background()` for shutdown - parent context is already cancelled
- 30-second timeout matches Kubernetes default terminationGracePeriodSeconds
- `slog.Error` before `os.Exit(1)` - proper error logging

### D5: golangci-lint Updates for Go 1.25

**Decision**: Add copyloopvar linter, update go version to 1.25

**New Linters**:
```yaml
linters:
  enable:
    - copyloopvar  # Go 1.22+ loop variable semantics
```

**Configuration**:
```yaml
run:
  go: "1.25"
  timeout: 5m
```

**Rationale**:
- Go 1.22+ changed loop variable semantics (copy per iteration)
- copyloopvar detects potential issues in migrated code
- Timeout increased for larger codebases with Go 1.25

## Risks and Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Distroless breaks health checks | High | Medium | Use HTTP probes, not wget/curl |
| MCP SDK incompatibility | High | Low | Pin to stable v1.2.0, test thoroughly |
| golangci-lint version mismatch | Medium | Medium | Pin to v2.4.0+ in CI, document requirements |
| Build time increase | Low | High | Accept +10s for security benefits |
| Binary size increase (bash-static) | Low | High | Accept +2MB for flexibility |

## Migration Strategy

### For Existing Projects

#### Option 1: Manual Migration (Recommended for Production)
1. Update `go.mod`: `go 1.25.5`
2. Run `go mod tidy`
3. Replace `build/Dockerfile` with new template
4. Update `Makefile` with VERSION/VCS_REF
5. Rebuild Docker images
6. Test deployment

#### Option 2: Automated Migration (Future Enhancement)
```bash
goent upgrade --to-go 1.25.5
```
- Out of scope for this change
- Future proposal: `goent upgrade` command

### For New Projects

No action required - templates generate Go 1.25.5 projects automatically.

## Open Questions

### Q1: Should we provide multi-arch Docker builds?

**Status**: Deferred to future change

**Reasoning**:
- Adds complexity (buildx, QEMU)
- Most users deploy to amd64
- Can be added without breaking changes

**Recommendation**: Document manual multi-arch build in templates/README.md

### Q2: Should main.go expose version/vcsRef as flags?

**Status**: Deferred to future change

**Reasoning**:
- Requires flag parsing in every generated project
- Not all projects need version commands
- Can be added as optional template enhancement

**Recommendation**: Document in templates/CLAUDE.md.tmpl as pattern to follow

## Implementation Complexity

| Task | Complexity | Risk | Justification |
|------|------------|------|---------------|
| T001: go.mod upgrade | Low | Low | Simple version change |
| T002: main.go refactor | Medium | Low | Well-defined pattern exists |
| T003: Makefile VERSION | Low | Low | Standard Makefile pattern |
| T004-T005: Template updates | Low | Low | Version bumps, minor changes |
| T006: Dockerfile rewrite | **High** | **Medium** | Major base image change |
| T007: golangci.yml | Low | Low | Config file update |
| T008: MCP templates | Medium | Medium | New directory structure |
| T009: Scaffold tool | Medium | Low | Add template_type parameter |
| T010: Validation | Medium | Low | Integration testing |

**Total Estimated Effort**: 2-3 days for implementation + 1 day for validation
