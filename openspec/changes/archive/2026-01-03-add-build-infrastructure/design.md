# Design: Build Infrastructure

## Context

The go-ent project has a CLI tool that cannot build due to Go embed limitations. The CLI is currently structured as a separate Go module at `cli/` with its own `go.mod`, but `//go:embed` cannot access parent directories. Templates exist at `plugins/go-ent/templates/` but the CLI cannot reach them.

**Constraints:**
- Go embed directive `//go:embed` is module-relative and cannot access `../`
- Plugin consumers expect templates at `plugins/go-ent/templates/`
- CLI needs to be buildable and testable
- CI needs to validate Go code on every push/PR

## Goals / Non-Goals

**Goals:**
- Fix CLI build by restructuring to allow template embedding
- Add comprehensive build/test/lint infrastructure
- Enable CI validation for Go code
- Maintain backward compatibility for plugin template location

**Non-Goals:**
- Release workflow with goreleaser (future work)
- CLI refactoring for testability (inject fs.FS - future work)
- Integration tests (future work)
- Migration to cobra/kong (future work)

## Decisions

### Decision 1: Move CLI to Root Module

**What:** Restructure CLI from separate module at `cli/` to `cmd/goent/` under root `go.mod`.

**Why:**
- `//go:embed` can only access files within the same module
- Embedding from parent requires root module structure
- Standard Go project layout uses `cmd/` for multiple binaries
- Eliminates need for symlinks in embedded paths

**Alternatives considered:**
1. **Symlinks (`cli/templates → ../plugins/go-ent/templates`)**: Fails - embed follows symlinks but path must be module-relative
2. **Go workspace (`go.work`)**: Still requires symlinks or copy for embed since embed is module-relative
3. **Makefile copy target**: Violates DRY, templates drift risk, requires manual prep step

**Trade-offs:**
- ✅ Clean embed directive: `//go:embed ../../templates/*`
- ✅ Standard Go project structure
- ✅ No build-time copy steps
- ⚠️ Breaking change for anyone building from source
- ⚠️ Need to maintain symlink for plugin consumers

### Decision 2: Templates at Root with Plugin Symlink

**What:**
- Move `plugins/go-ent/templates/` → `templates/` at project root
- Create symlink `plugins/go-ent/templates → ../../templates`

**Why:**
- CLI embed can access `../../templates` from `cmd/goent/main.go`
- Symlink maintains plugin compatibility (plugin references stay valid)
- Single source of truth for templates

**Trade-offs:**
- ✅ DRY - single template source
- ✅ Plugin compatibility via symlink
- ✅ Embed works natively
- ⚠️ Symlink may not work on Windows (but plugin is for Unix-like dev environments)

### Decision 3: Makefile Build Targets

**What:** Create root `Makefile` with targets: `build`, `test`, `lint`, `fmt`, `clean`, `validate-plugin`

**Why:**
- Standard developer interface across Go projects
- Encapsulates build complexity (embed paths, output locations)
- CI can use same commands as local development

**Targets:**
```makefile
build:    Build CLI binary to dist/goent
test:     Run tests with race detector and coverage
lint:     Run golangci-lint
fmt:      Format with goimports
clean:    Remove dist/ and build artifacts
validate-plugin: Validate plugin.json structure
```

### Decision 4: Linting Configuration

**What:** Create `.golangci.yml` at root with curated linters

**Linters enabled:**
- **Critical**: errcheck, govet, staticcheck
- **Style**: gosimple, gofmt, goimports
- **Quality**: gocritic, revive
- **Production**: gosec (security), misspell

**Why:**
- Enforce code quality before CI
- Catch common bugs (errcheck)
- Consistent formatting (gofmt, goimports)
- Security best practices (gosec)

### Decision 5: CI Pipeline Structure

**What:** Add `go-cli` job to `.github/workflows/validate.yml` running in parallel with `validate-plugin`

**Steps:**
1. Set up Go 1.23
2. Build: `go build -v ./...`
3. Test: `go test -race -cover ./...`
4. Lint: golangci-lint-action (uses `.golangci.yml`)

**Why:**
- Parallel execution speeds up CI
- Fail fast on build errors
- Race detector catches concurrency bugs
- Coverage reporting for visibility

## Module Structure

**Before:**
```
cli/                                # Separate module
├── go.mod                          # github.com/victorzhuk/go-ent/cli
└── main.go                         # //go:embed templates/* (FAILS)

plugins/go-ent/templates/           # Templates here
```

**After:**
```
go.mod                              # github.com/victorzhuk/go-ent (root)
go.sum

cmd/goent/                          # CLI binary
├── main.go                         # //go:embed ../../templates/*
└── main_test.go                    # Tests

templates/                          # Moved from plugins/go-ent/
├── .gitignore.tmpl
├── .golangci.yml.tmpl
├── Makefile.tmpl
└── ...

plugins/go-ent/templates → ../../templates  # Symlink for plugin

Makefile                            # Build infrastructure
.golangci.yml                       # Linting config
```

## Embed Path Calculation

From `cmd/goent/main.go`:
```go
//go:embed ../../templates/*
var templates embed.FS
```

Path resolution:
- `cmd/goent/main.go` location
- `../../` goes up to project root
- `templates/` is at root
- Result: embeds `/home/zhuk/Projects/own/go-ent/templates/*`

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Symlink fails on Windows | Document requirement; plugin is for Unix dev environments |
| Breaking change for source builds | Update README with new build commands |
| Template path drift | Symlink ensures plugin sees same templates as CLI |
| CI takes longer | Parallel jobs, use Go cache |

## Migration Plan

**For users building from source:**

Before:
```bash
cd cli
go build .
```

After:
```bash
make build          # or
go build ./cmd/goent
```

**For plugin users:**
- No changes needed - symlink maintains `plugins/go-ent/templates/` path

**Rollback:**
If symlink causes issues, can fall back to Makefile copy target temporarily.

## Open Questions

None - design is complete and ready for implementation.
