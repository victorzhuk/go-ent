# Change: Modernize Templates and CLI to Go 2026 Standards

## Status

**DRAFT** - Awaiting review

## Why

The project and generated templates use outdated Go 1.23 and alpine-based Docker images. Production Go applications in 2025/2026 require:

- **Go 1.25.5** with DWARF v5 debug info, new GC, and testing/synctest support
- **Distroless runtime images** for minimal attack surface (~2MB vs alpine's ~7MB)
- **Proper signal handling** with SIGTERM/SIGINT/SIGQUIT support
- **MCP server templates** for modern AI tooling ecosystems
- **Reproducible builds** with VERSION and VCS_REF metadata

### Current Issues

1. **Project** (`go-ent`):
   - `cmd/go-ent/main.go` uses `log.Fatal` - violates project standards
   - No graceful shutdown on signals
   - Missing structured logging setup
   - Go 1.23.0 lacks newest features

2. **Templates**:
   - Alpine-based Docker images have larger attack surface
   - Go 1.23 in generated projects is outdated
   - No MCP server template option
   - Missing build metadata (VCS_REF)

## What Changes

### Project Root Modifications

| File | Change | Impact |
|------|--------|--------|
| `go.mod` | `go 1.23.0` → `go 1.25.5` | Use latest Go features |
| `cmd/go-ent/main.go` | Refactor to `run(ctx, getenv, stdout, stderr)` pattern | Testable, follows standards |
| `Makefile` | Add `VERSION` and `VCS_REF` build args | Reproducible builds |

### Template Modifications

| File | Change | Impact |
|------|--------|--------|
| `templates/go.mod.tmpl` | `go 1.23` → `go 1.25.5` | Generated projects use Go 1.25.5 |
| `templates/Makefile.tmpl` | Add VERSION/VCS_REF, use `bin/` directory | Build metadata tracking |
| `templates/build/Dockerfile.tmpl` | **MAJOR REWRITE**: distroless + bash-static | Security, size reduction |
| `templates/.golangci.yml.tmpl` | Go 1.25 linters, add copyloopvar | Latest linting standards |

### New Templates

| Directory | Purpose |
|-----------|---------|
| `templates/mcp/` | MCP server project templates |
| `templates/mcp/cmd/server/` | MCP main.go with stdio transport |
| `templates/mcp/internal/server/` | MCP SDK setup patterns |

### Binary Naming

All binaries use `{{PROJECT_NAME}}` pattern (no `srv_` prefix) across:
- Makefile build targets
- Docker COPY commands
- CMD/ENTRYPOINT directives

## Affected Specs

- **templates** (NEW) - MCP templates, distroless, build metadata
- **cli-build** (MODIFIED) - run() pattern, Go 1.25.5

## Migration Impact

### For go-ent Maintainers

No breaking changes. After upgrade:
- `make build` continues to work
- Tests continue to pass
- MCP protocol remains unchanged

### For Template Users

**BREAKING**: Existing Docker images must be rebuilt.

New projects get:
- Go 1.25.5 automatically
- Distroless runtime images
- Optional MCP template type

Existing projects can upgrade by:
1. Updating `go.mod` to `go 1.25.5`
2. Replacing `build/Dockerfile` with new template
3. Updating `Makefile` with VERSION/VCS_REF
4. Rebuilding Docker images

## Implementation Phases

1. **Phase 1**: Project root upgrades (go.mod, main.go, Makefile)
2. **Phase 2**: Template upgrades (go.mod, Makefile, Dockerfile, golangci.yml)
3. **Phase 3**: MCP server templates (new directory structure)
4. **Phase 4**: Validation (lint, test, generate, build)

## Success Criteria

- [ ] `make lint` passes with Go 1.25.5
- [ ] `make test` passes with Go 1.25.5
- [ ] Generated HTTP project builds and runs
- [ ] Generated MCP project builds and runs
- [ ] Docker images use distroless runtime
- [ ] Graceful shutdown works with SIGTERM
- [ ] Binary names follow `{{PROJECT_NAME}}` pattern
- [ ] `openspec validate refactor-modernize-templates-and-cli --strict` passes
