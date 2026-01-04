# cli-build Specification Delta

## MODIFIED Requirements

### Requirement: Template Embedding
The CLI tool SHALL embed all project templates including dotfiles at build time using Go embed, with templates located in `/internal/templates/`.

**What changed**:
1. Consolidated templates from `/templates/` (source) and `/cmd/goent/templates/` (copy) into `/internal/templates/`
2. Fixed embed directive to explicitly include dotfiles (`.gitignore.tmpl`, `.golangci.yml.tmpl`)
3. Eliminated build-time template copying step

#### Scenario: Templates embedded in binary
- **WHEN** CLI is built
- **THEN** all template files from `/internal/templates/` directory are embedded in the binary
- **AND** embedded files include dotfiles (`.gitignore.tmpl`, `.golangci.yml.tmpl`)
- **AND** embedded files are accessible via `embed.FS`
- **AND** no build-time copying is required

#### Scenario: Dotfiles properly embedded
- **WHEN** checking embedded files after build
- **THEN** `.gitignore.tmpl` is included in binary
- **AND** `.golangci.yml.tmpl` is included in binary
- **AND** all other dotfile templates (`.*.tmpl`) are included
- **AND** embed directive uses explicit paths for dotfiles

#### Scenario: Embedded templates accessible at runtime
- **WHEN** `goent_generate` tool is executed
- **THEN** CLI can read all template files from embedded filesystem
- **AND** dotfile templates are accessible
- **AND** template files are processed and written to target project directory
- **AND** generated projects include `.gitignore` and `.golangci.yml` files

#### Scenario: Single template source
- **WHEN** examining project structure
- **THEN** templates exist only at `/internal/templates/`
- **AND** no `/templates/` directory at project root
- **AND** no `/cmd/goent/templates/` copy directory
- **AND** embed directive references `/internal/templates/` directly

### Requirement: Build Artifacts
The CLI build process SHALL produce clean, versioned artifacts without requiring template preparation steps.

**What changed**: Removed `prepare-templates` Makefile target and template copying from build process.

#### Scenario: Build without preparation
- **WHEN** `make build` is executed
- **THEN** binary is built directly without template copying
- **AND** no intermediate template directories are created
- **AND** build completes successfully

#### Scenario: Clean build artifacts
- **WHEN** `make clean` is executed
- **THEN** `dist/` directory is removed
- **AND** no `/cmd/goent/templates/` artifacts remain
- **AND** build artifacts are fully cleaned

### Requirement: Template Structure
Templates SHALL be organized in `/internal/templates/` with proper subdirectory structure and explicit embedding directives.

**What changed**: Moved templates to `/internal/` and defined explicit embed patterns for all subdirectories.

#### Scenario: Template organization
- **WHEN** examining `/internal/templates/` structure
- **THEN** root-level templates exist (e.g., `CLAUDE.md.tmpl`, `Makefile.tmpl`)
- **AND** dotfile templates exist (e.g., `.gitignore.tmpl`, `.golangci.yml.tmpl`)
- **AND** subdirectories exist (`build/`, `cmd/`, `deploy/`, `internal/`, `mcp/`)
- **AND** each subdirectory contains appropriate `.tmpl` files

#### Scenario: Embed directive patterns
- **WHEN** reviewing `/internal/templates/embed.go`
- **THEN** root templates embedded with `//go:embed *.tmpl`
- **AND** dotfiles explicitly embedded: `//go:embed .gitignore.tmpl .golangci.yml.tmpl`
- **AND** subdirectories embedded with specific patterns (e.g., `//go:embed mcp/*.tmpl mcp/**/*.tmpl`)
- **AND** no glob patterns like `**/*.tmpl` used (they miss dotfiles)
