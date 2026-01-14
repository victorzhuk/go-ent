# Tooling Reference

Common patterns for Serena, Git, Go, and Bash tools used across agents.

## Serena Tools

### File Operations
- **read**: Read file contents (always before editing)
- **write**: Write new files or overwrite existing
- **edit**: Make targeted replacements with exact matches
- **glob**: Find files by pattern (e.g., `**/*.go`)
- **grep**: Search file contents with regex patterns

### Serena-Specific
- **serena_list_dir**: Directory structure overview
- **serena_find_symbol**: Find code symbols (classes, functions)
- **serena_find_referencing_symbols**: Find where symbols are used
- **serena_search_for_pattern**: Pattern-based content search
- **serena_replace_content**: Regex-based content replacement
- **serena_get_symbols_overview**: High-level file understanding

## Git Commands

### Change Analysis
```bash
git diff --name-only HEAD~1           # Show changed files
git log --oneline -10 -- {path}       # Recent changes to path
git diff HEAD~1..HEAD                  # Full diff
```

### Status & History
```bash
git status                              # Working tree status
git log -p -1                           # Last commit with changes
git branch -a                           # All branches
```

## Go Commands

### Build & Test
```bash
go build ./...                          # Build all packages
go test ./...                           # Run all tests
go test ./... -race                     # Race detection
go test -run TestXxx -v ./pkg/...      # Specific test
go test -coverprofile=c.out ./...       # Coverage report
```

### Linting
```bash
golangci-lint run                       # Full lint
golangci-lint run --fast                # Fast lint
gofmt -s -w .                           # Format code
goimports -w .                          # Format imports
```

## Bash Search Patterns

### Grep Patterns
```bash
# Search across codebase
grep -rn "pattern" internal/            # Recursive search
grep -rn "func New" internal/repository/  # Find constructors

# Find specific patterns
grep -rn "applicationConfig\|userRepository" internal/  # Anti-patterns
grep -rn "// Create\|// Get\|// Set" internal/         # WHAT comments
grep -rn 'return err$' internal/                       # Unwrapped errors
```

### Find Patterns
```bash
# Directory structure
find internal -type d -depth 2           # Two-level deep dirs
find . -name "*.go" -type f              # All Go files
find . -name "*_test.go" -type f         # All test files
```

### Ripgrep (faster alternative)
```bash
rg "pattern" -g "*.go"                  # Go files only
rg "error:" internal/ --type go         # Type-filtered
```

## Debugging Commands

```bash
# Test with verbose output
go test -v -run TestName ./...

# Run with specific flags
go test -race -run TestAgentConfig_Valid ./internal/domain

# Check imports
go list -f '{{join .Deps "\n"}}' ./... | sort | uniq

# Build check
go build -o /dev/null ./...
```

## Database Commands

```bash
# Goose migrations
goose -dir ./migrations up
goose -dir ./migrations status
goose -dir ./migrations create name sql

# pgx direct (via testcontainers)
docker exec -it {container} psql -U user -d dbname
```

## Tool Selection Guide

| Task | Tool |
|------|------|
| Read file | `read` |
| Search symbols | `serena_find_symbol` |
| Find usages | `serena_find_referencing_symbols` |
| Pattern search | `serena_search_for_pattern` |
| Replace code | `serena_replace_content` or `edit` |
| Find files | `glob` |
| Command execution | `bash` |
| Find Git changes | `git diff` |
