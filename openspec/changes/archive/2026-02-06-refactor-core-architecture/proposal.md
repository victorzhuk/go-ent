## Why

The go-ent codebase has accumulated technical debt from rapid prototyping, resulting in god objects, mixed responsibilities, brittle tests, and non-idiomatic Go patterns. These issues impact maintainability, testability, and onboarding for new developers. Refactoring now establishes a solid foundation for future development and follows Go best practices (SOLID, DRY, KISS, YAGNI) to ensure long-term sustainability.

## What Changes

- **Extract task parsing logic**: Remove parsing responsibilities from `BoltStore` in `spec/boltdb.go` into a dedicated task parser component
- **Split skill.Registry**: Decompose `skill/registry.go` into focused components with single responsibilities:
  - Skill metadata management
  - Runtime skill execution
  - Skill matching and selection logic
- **Rewrite brittle tests**: Replace implementation-detail testing with behavior-focused tests that use interfaces and test real functionality
- **Standardize error handling**: Establish consistent error wrapping patterns, define package-level errors, and ensure all errors provide context
- **Replace custom string functions**: Remove custom string utilities in favor of Go standard library functions
- **Simplify configuration structure**: Flatten deeply nested configuration to improve readability and maintainability
- **Apply Clean Architecture principles**: Ensure clear separation of concerns with domain, usecase, repository, and transport layers

## Capabilities

### New Capabilities

- `task-parsing`: Parse task definitions from markdown documents with support for multi-task files, task metadata extraction, and structured task representation
- `skill-registry`: Manage skill metadata, registration, and lifecycle with separation between static metadata and runtime execution
- `spec-storage`: Persist and retrieve spec data from BoltDB with a clean API that separates storage concerns from business logic
- `configuration`: Load and validate application configuration from environment variables with flattened structure and clear validation rules

### Modified Capabilities

*None - This is an internal refactoring with no specification-level requirement changes. The existing spec storage, skill management, and configuration behaviors remain the same from a user perspective, only implementation details change.*

## Impact

**Code Impact**:
- `spec/boltdb.go`: Reduce from ~300 lines to focused storage operations (~100 lines)
- `skill/registry.go`: Split into 3-4 focused files with clear boundaries
- Test suite: Replace brittle implementation tests with behavior-driven tests
- Configuration: Flatten nested structures for better readability
- All packages: Apply consistent error handling patterns throughout

**Architecture Impact**:
- Introduce Clean Architecture layers where missing
- Define clear interfaces at consumer boundaries
- Remove god objects and mixed responsibilities
- Establish domain layer with zero external dependencies

**Build and Test Impact**:
- Build output: No changes expected
- Test coverage: Maintain or improve existing coverage
- Test stability: Reduce flakiness by testing behavior not implementation

**User Impact**:
- CLI commands: No breaking changes to public CLI API
- MCP tools: No changes to MCP tool interfaces
- Documentation: Update architecture documentation to reflect new structure

**Migration Impact**:
- No database migrations required (BoltDB schema unchanged)
- No configuration changes required for existing users
- Existing spec files remain compatible

**Dependencies**:
- No new external dependencies introduced
- May remove unused dependencies from simplification
- Standard library usage increased
