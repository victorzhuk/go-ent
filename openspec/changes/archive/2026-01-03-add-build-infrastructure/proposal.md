# Change: Add Build Infrastructure

## Why

The CLI tool cannot build due to missing templates directory (`//go:embed templates/*` fails). The project lacks:
- Go build/test infrastructure (no tests, no linting)
- CI validation for Go code (only plugin JSON validation exists)
- Development workflow tooling (no Makefile)

This prevents the CLI from being built, tested, or distributed.

## What Changes

- **BREAKING**: Restructure CLI module from `cli/` to `cmd/goent/` under root go.mod
- Move templates from `plugins/go-ent/templates/` to `templates/` at project root
- Create symlink `plugins/go-ent/templates → ../../templates` for plugin compatibility
- Update CLI embed directive to `//go:embed ../../templates/*`
- Add root `go.mod` with module `github.com/victorzhuk/go-ent`
- Add `Makefile` with targets: build, test, lint, clean
- Add `.golangci.yml` with linters: errcheck, gosimple, govet, staticcheck, gofmt, goimports
- Add CLI tests in `cmd/goent/main_test.go` using testify
- Update `.github/workflows/validate.yml` to add Go build/test/lint job
- Fix existing CI bug: `plugins/*/claude-plugin/` → `plugins/*/.claude-plugin/`

## Impact

- **Affected specs**: cli-build (NEW), ci-pipeline (NEW)
- **Affected code**:
  - `cli/` → `cmd/goent/` (restructure)
  - `plugins/go-ent/templates/` → `templates/` (move + symlink)
  - New files: `go.mod`, `Makefile`, `.golangci.yml`, `cmd/goent/main_test.go`
  - Modified: `.github/workflows/validate.yml`
- **Breaking change**: CLI build location changes from `cli/` to `cmd/goent/`
- **Migration**: Users building from source need to use `make build` or `go build ./cmd/goent`
