<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

# AGENTS.md

## Build Commands

```bash
make build          # Build binary to bin/go-ent
make test           # Run all tests with race detection and coverage
make lint           # Run golangci-lint
make fmt            # Format code with goimports
make clean          # Remove build artifacts
make validate-plugin  # Validate plugin JSON files
```

## Running Single Tests

```bash
# Run specific test function
go test -run TestAgentRole_String ./internal/domain

# Run tests for a package
go test ./internal/...

# Verbose mode
go test -v ./internal/domain

# Run with specific flags
go test -race -run TestAgentConfig_Valid ./internal/domain
```

## Code Style

### Imports
Standard lib (sorted) → Third-party → Internal with blank lines. Use `make fmt`.

### Naming
- Types/Constants: PascalCase (AgentRole, AgentRoleProduct)
- Functions: PascalCase exported, camelCase private
- Variables: Short but meaningful (cfg, repo, ctx)
- Receivers: Single letters (s *Store, c *AgentConfig)
- Files: lowercase_with_underscores.go

### Error Handling
- Package-level errors in `errors.go`
- Lowercase messages, no trailing punctuation
- Wrap with context: `fmt.Errorf("context: %w", err)`

### Types
- Public types with doc comments, PascalCase fields
- Enums: `type Role string`
- Validation: `Valid() bool`, `Validate() error`

### Organization
Package doc → Imports → Constants/Errors → Types → Variables → Public funcs → Private funcs

### Testing
- Table-driven with `t.Run()` and `t.Parallel()`
- Use `testify/assert`
- Files: `filename_test.go`

### Interfaces
- Minimal interfaces at consumer side
- Accept interfaces, return structs

### Context
First parameter, pass through all layers for cancellation/timeouts

### Linting & Formatting
```bash
make lint  # golangci-lint with errcheck, gosec, revive, staticcheck
make fmt    # goimports for import formatting
```

## Project Structure

```
internal/
├── agent/         # Agent types and selection logic
├── cli/           # CLI application
├── config/        # Configuration management
├── domain/        # Core domain types
├── mcp/           # MCP server and tools
├── skill/         # Skill registry and parsing
└── spec/          # Spec store and validation
```

## Quick Reference

- **Before coding**: Read existing code patterns in the relevant package
- **After changes**: Run `make lint` and `make test`
- **Testing**: Use table-driven tests with `t.Run()` and `t.Parallel()`
- **Errors**: Define in `errors.go`, wrap with context
- **Formatting**: Always run `make fmt`
