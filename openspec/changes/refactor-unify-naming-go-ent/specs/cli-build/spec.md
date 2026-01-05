# CLI Build Spec Delta

## MODIFIED Requirements

### Requirement: Binary Naming
The go-ent CLI binary SHALL be named `go-ent` (with hyphen) for consistency with the repository and module names.

**Previous**: Binary was named `goent` (without hyphen)
**Reason for Change**: Unify naming across all project components

#### Scenario: Build produces correctly named binary
- **WHEN** `make build` is executed
- **THEN** the output binary SHALL be `dist/go-ent`
- **AND** running `./dist/go-ent version` SHALL display "go-ent vX.Y.Z"
