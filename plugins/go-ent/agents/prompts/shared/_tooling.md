# Tooling Reference

Common patterns for native tools, Serena semantic analysis, Git, Go, and Bash tools used across agents.

## Native File Operations

Use Claude Code native tools for all file CRUD operations:

- **Read**: Read file contents (always before editing)
- **Write**: Write new files or overwrite existing
- **Edit**: Make targeted replacements with exact string matches
- **Glob**: Find files by pattern (e.g., `**/*.go`)
- **Grep**: Search file contents with regex patterns
- **Bash**: Execute shell commands

## Serena Semantic Analysis

Use Serena tools **only** for semantic code analysis and understanding:

- **serena_find_symbol**: Find code symbols (classes, functions, methods) by name path
- **serena_find_referencing_symbols**: Find all references to a symbol across the codebase
- **serena_get_symbols_overview**: Get high-level overview of symbols in a file
- **serena_search_for_pattern**: Flexible pattern-based content search with filters
- **serena_list_dir**: Directory structure listing (use when Glob isn't sufficient)

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

| Task | Tool | Category | Notes |
|------|------|----------|-------|
| **File Operations** | | | |
| Read file | `Read` | Native | Always before editing |
| Write new file | `Write` | Native | Creates or overwrites |
| Edit file | `Edit` | Native | Exact string replacement |
| Find files | `Glob` | Native | Pattern matching (e.g., `**/*.go`) |
| Search content | `Grep` | Native | Regex search in files |
| Execute commands | `Bash` | Native | Shell commands |
| **Semantic Analysis** | | | |
| Find code symbols | `serena_find_symbol` | Serena | Classes, functions, methods |
| Find symbol usages | `serena_find_referencing_symbols` | Serena | Cross-file references |
| File structure | `serena_get_symbols_overview` | Serena | High-level analysis |
| Pattern search | `serena_search_for_pattern` | Serena | Advanced search with filters |
| **Version Control** | | | |
| Find changes | `git diff` | Bash | Via Bash tool |
| View history | `git log` | Bash | Via Bash tool |
