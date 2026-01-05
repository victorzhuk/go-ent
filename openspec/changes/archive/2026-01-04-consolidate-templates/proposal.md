# Change: Consolidate Templates and Fix Dotfile Embedding

## Why

The current template system has two critical issues:

### Issue 1: Template Duplication and Build Complexity

**Current workflow** (inefficient):
```
/templates/ (source)
    ├── .gitignore.tmpl
    ├── CLAUDE.md.tmpl
    └── ...

    ↓ make prepare-templates (copy step)

/cmd/go-ent/templates/ (copy for embedding)
    ├── CLAUDE.md.tmpl
    └── ... (missing dotfiles!)

    ↓ go build (embed)

dist/go-ent (binary with embedded templates)
```

**Problems**:
1. Build step requires copying 15 template files
2. Two template directories maintained (source vs copy)
3. Easy to get out of sync during development
4. Build artifacts (`cmd/go-ent/templates/`) tracked in git status

### Issue 2: Dotfile Embedding Bug

The `//go:embed **/*.tmpl` pattern in `/cmd/go-ent/templates/embed.go` does **not** match dotfiles:
- `.gitignore.tmpl` - **NOT embedded** ❌
- `.golangci.yml.tmpl` - **NOT embedded** ❌
- All other `.*.tmpl` files - **NOT embedded** ❌

**Impact**:
- Generated projects missing `.gitignore`
- Generated projects missing `.golangci.yml`
- Silent failure - no build error, templates just missing at runtime

**Root cause**: Go's `embed` directive doesn't match dotfiles with glob patterns like `**/*.tmpl`. Explicit paths needed.

### Proposed Solution

Consolidate templates into `/internal/templates/` with fixed embedding:

```
/internal/templates/
    ├── embed.go           ← Fixed embed directives
    ├── .gitignore.tmpl    ✓ Now embedded
    ├── .golangci.yml.tmpl ✓ Now embedded
    └── ... (all templates)
```

**Benefits**:
1. Single source of truth - no copying needed
2. Simpler build process - `go build` just works
3. All templates properly embedded (including dotfiles)
4. Templates live with other internal packages
5. Cleaner git status (no build artifacts)

## What Changes

### 1. Template Consolidation

Move templates from split locations to unified location:

```
Before:
/templates/                     ← Source templates (15 files)
/cmd/go-ent/templates/embed.go   ← Embed directive (broken)
Makefile: prepare-templates     ← Copy step

After:
/internal/templates/            ← All templates + embed.go
    ├── embed.go                ← Fixed directives
    ├── .gitignore.tmpl         ← Now included
    ├── .golangci.yml.tmpl      ← Now included
    └── ...
```

### 2. Fixed Embed Directive

Replace glob pattern with explicit paths to include dotfiles:

```go
// OLD (broken - misses dotfiles)
//go:embed **/*.tmpl
var TemplateFS embed.FS

// NEW (fixed - includes all templates)
//go:embed *.tmpl
//go:embed .gitignore.tmpl .golangci.yml.tmpl
//go:embed build/*.tmpl
//go:embed cmd/*.tmpl cmd/**/*.tmpl
//go:embed deploy/*.tmpl
//go:embed internal/*.tmpl internal/**/*.tmpl
//go:embed mcp/*.tmpl mcp/**/*.tmpl
var TemplateFS embed.FS
```

### 3. Import Path Updates

Update template imports (2 files):

```diff
-import "github.com/victorzhuk/go-ent/cmd/go-ent/templates"
+import "github.com/victorzhuk/go-ent/internal/templates"
```

### 4. Build Process Simplification

Remove template copying from Makefile:

```diff
-prepare-templates: ## Copy templates to cmd/go-ent for embedding
-	@mkdir -p cmd/go-ent/templates
-	@find cmd/go-ent/templates -name "*.tmpl" -delete
-	@rm -rf cmd/go-ent/templates/mcp cmd/go-ent/templates/build ...
-	@cp -r templates/* cmd/go-ent/templates/
-	@echo "Templates prepared for embedding"

-build: prepare-templates ## Build CLI binary
+build: ## Build CLI binary
	@mkdir -p dist
	@go build $(LDFLAGS) -o dist/go-ent ./cmd/go-ent

clean: ## Remove dist/ and build artifacts
	@rm -rf dist/
-	@rm -rf cmd/go-ent/templates
	@rm -f goent
```

### 5. Files Affected

**Moved**:
- `/templates/*.tmpl` → `/internal/templates/*.tmpl` (15 files)
- `/cmd/go-ent/templates/embed.go` → `/internal/templates/embed.go` (1 file)

**Modified**:
- `/cmd/go-ent/internal/tools/generate.go` - import path
- `/cmd/go-ent/internal/tools/generate_from_spec.go` - import path
- `/Makefile` - remove `prepare-templates`, update `build` and `clean`
- `/internal/templates/embed.go` - fix embedding directives

**Deleted**:
- `/templates/` directory (after move)
- `/cmd/go-ent/templates/` directory (no longer needed)

## Impact

- **Affected specs**: `cli-build` (template embedding requirements modified)
- **New directory**: `/internal/templates/` (consolidation target)
- **Removed directories**: `/templates/`, `/cmd/go-ent/templates/`
- **Code changes**: 3 files (2 import updates + Makefile)
- **Breaking changes**: None (internal refactoring)
- **Dependencies**: Should be done **after** `refactor-move-core-packages` to have `/internal/` established

## Success Criteria

1. All 15 templates moved to `/internal/templates/`
2. Dotfiles `.gitignore.tmpl` and `.golangci.yml.tmpl` properly embedded
3. `make build` succeeds without `prepare-templates` step
4. Generated projects include `.gitignore` and `.golangci.yml`
5. `make test` passes all tests
6. No template-related build artifacts in git status

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Missed templates in embedding | Medium | Explicit verification that all 15 files embed successfully |
| Template paths incorrect | Low | Go compiler will catch missing embeds at build time |
| Build breaks without prepare step | Low | Test build immediately after Makefile change |
| Generated projects malformed | Medium | Test `go_ent_generate` after changes |
