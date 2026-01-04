# cli-build Specification Delta

## MODIFIED Requirements

### Requirement: Go Module Structure
The CLI SHALL follow clean code principles with no unused directories or files in the project structure.

**What changed**: Removed empty `/cmd/goent/internal/resources/` directory that had no purpose or usage.

#### Scenario: No unused directories
- **WHEN** examining `/cmd/goent/internal/` structure
- **THEN** only active packages with code exist
- **AND** no empty directories are present
- **AND** all directories serve a documented purpose

#### Scenario: Clean package structure
- **WHEN** listing directories in `/cmd/goent/internal/`
- **THEN** only `tools/` and `server/` directories exist (after refactoring)
- **AND** each directory contains Go source files
- **AND** no placeholder or empty packages exist
