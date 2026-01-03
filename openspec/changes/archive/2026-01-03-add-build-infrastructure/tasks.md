# Implementation Tasks

## 1. Restructure CLI Module

- [x] 1.1 Create root `go.mod` with module `github.com/victorzhuk/go-ent`
- [x] 1.2 Create `cmd/goent/` directory
- [x] 1.3 Move `cli/main.go` to `cmd/goent/main.go`
- [x] 1.4 Update embed directive in `cmd/goent/main.go` to `//go:embed templates/*` (Note: used cmd/goent/templates symlink workaround)
- [x] 1.5 Delete `cli/go.mod`
- [x] 1.6 Delete `cli/` directory (now empty)
- [x] 1.7 Run `go mod tidy` to ensure module is valid

## 2. Restructure Templates

- [x] 2.1 Move `plugins/go-ent/templates/` to `templates/` at project root
- [x] 2.2 Create symlink `plugins/go-ent/templates → ../../templates` for plugin compatibility
- [x] 2.3 Verify symlink works: `ls -la plugins/go-ent/templates`
- [x] 2.4 Test embed: `go build ./cmd/goent` (implemented via Makefile prepare-templates)

## 3. Add Build Infrastructure

- [x] 3.1 Create `Makefile` at project root with targets:
  - `build` - Build CLI to `dist/goent`
  - `test` - Run tests with `-race -cover`
  - `lint` - Run golangci-lint
  - `fmt` - Format with goimports
  - `clean` - Remove `dist/` and artifacts
  - `validate-plugin` - Validate plugin.json
  - `prepare-templates` - Copy templates for embedding
  - `help` - Show available targets
- [x] 3.2 Create `.golangci.yml` with linters: errcheck, gosimple, govet, staticcheck, gofmt, goimports, gocritic, revive, gosec, misspell
- [x] 3.3 Verify build: `make build`
- [x] 3.4 Verify lint: `make lint` (configuration created)

## 4. Add Tests

- [x] 4.1 Add testify dependency: `go get github.com/stretchr/testify`
- [x] 4.2 Run `go mod tidy`
- [x] 4.3 Create `cmd/goent/main_test.go` with test cases:
  - `TestCreateProject` - Verify directory/file creation
  - `TestCreateProject_TemplateReplacements` - Verify `{{PROJECT_NAME}}` and `{{MODULE_PATH}}` replaced
  - `TestCreateProject_NoTemplateMarkers` - Verify no `{{...}}` markers remain
  - `TestInitCmd_DefaultValues` - Verify default module path generation
  - `TestInitCmd_MissingProjectName` - Verify error handling
  - Additional tests for permissions, invalid paths, template replacement logic
- [x] 4.4 Run tests: `make test`
- [x] 4.5 Verify all tests pass (36.2% coverage)

## 5. Update CI Workflow

- [x] 5.1 Read current `.github/workflows/validate.yml`
- [x] 5.2 Fix plugin.json path bug: `plugins/*/claude-plugin/plugin.json` → `plugins/*/.claude-plugin/plugin.json`
- [x] 5.3 Add new `go-cli` job with steps:
  - Checkout code
  - Set up Go 1.23
  - Prepare templates: `make prepare-templates`
  - Build: `go build -v ./...`
  - Test: `go test -race -cover ./...`
  - Lint: Use `golangci-lint-action` with version `v4`
- [x] 5.4 Verify jobs run in parallel (both `validate-plugin` and `go-cli`)
- [ ] 5.5 Test locally: commit changes and push to verify CI passes (requires git push)

## 6. Documentation Updates

- [x] 6.1 Update README.md with new build instructions:
  - `make build` or `go build ./cmd/goent`
  - Removed old `cd cli && go build` instructions
- [x] 6.2 Document Makefile targets in README
- [x] 6.3 Add note about template preparation requirement for development

## 7. Verification

- [x] 7.1 Clean build: `make clean && make build`
- [x] 7.2 Run all tests: `make test`
- [x] 7.3 Run linter: configuration created (requires golangci-lint installation)
- [x] 7.4 Test CLI binary: `dist/goent init test-project`
- [x] 7.5 Verify templates are embedded: checked generated project structure
- [ ] 7.6 Push to GitHub and verify CI passes on both jobs (requires git push)

## Implementation Notes

### Template Embedding Solution

The original plan proposed using `//go:embed ../../templates/*` to access templates from the parent directory. However, Go's embed directive doesn't support `..` paths. The implemented solution:

1. **Templates Location**: Moved to `templates/` at project root
2. **Plugin Compatibility**: Created symlink `plugins/go-ent/templates → ../../templates`
3. **CLI Embedding**: Makefile `prepare-templates` target copies `templates/` to `cmd/goent/templates/` before build
4. **Embed Directive**: Uses `//go:embed templates/*` (relative to cmd/goent/)
5. **Gitignore**: Added `cmd/goent/templates/` to .gitignore (generated directory)

This approach:
- ✅ Maintains single source of truth for templates
- ✅ Plugin symlink works for Claude Code consumers
- ✅ Build process is automated via Makefile
- ✅ Templates are properly embedded in CLI binary
- ⚠️ Requires `make build` instead of direct `go build` (documented in README)

### Additional Improvements

- Added input validation to `createProject()` function (discovered via tests)
- Comprehensive test suite with 11 test cases covering edge cases
- CI workflow now validates both plugin structure and Go code
- All build artifacts properly ignored in git
