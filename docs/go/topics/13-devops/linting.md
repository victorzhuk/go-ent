# Linting

Production-grade Go linting using golangci-lint v2, the standard aggregator that runs 50+ linters in parallel with smart caching.

## Quick Reference

| Pattern                                | Use When                          |
|----------------------------------------|-----------------------------------|
| `golangci-lint run`                    | Run all enabled linters           |
| `golangci-lint run --fix`              | Auto-fix issues where possible    |
| `golangci-lint run ./...`              | Lint entire project recursively   |
| `golangci-lint run --config=.yml`      | Use specific config file          |
| `golangci-lint cache clean`            | Clear analysis cache              |
| `golangci-lint linters`                | List available linters            |

## Installation

### Binary (Recommended)

```bash
# Linux/macOS
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.0.0

# Verify installation
golangci-lint version
```

### Docker

```bash
# Run in container
docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:v2.0.0 golangci-lint run

# Add to docker-compose.yml
services:
  lint:
    image: golangci/golangci-lint:v2.0.0
    volumes:
      - .:/app
    working_dir: /app
    command: golangci-lint run
```

### CI Installation with Caching

```yaml
# .github/workflows/lint.yml
- name: Install golangci-lint
  uses: golangci/golangci-lint-action@v6
  with:
    version: v2.0.0
    args: --timeout=5m
    skip-cache: false # Enable caching
    skip-pkg-cache: false
    skip-build-cache: false
```

## Configuration

### .golangci.yml Structure

```yaml
# .golangci.yml (Production-grade v2 config)
run:
  timeout: 5m
  tests: true
  build-tags:
    - integration
  modules-download-mode: readonly
  allow-parallel-runners: true
  go: '1.24'

output:
  formats:
    - format: colored-line-number
  print-issued-lines: true
  print-linter-name: true
  uniq-by-line: true
  sort-results: true
  show-stats: true

linters:
  disable-all: true
  enable:
    # Essential (always enabled)
    - errcheck        # Unchecked errors
    - gosimple        # Simplify code
    - govet           # Go vet built-in
    - ineffassign     # Unused assignments
    - staticcheck     # Go static analysis
    - unused          # Unused code

    # Code quality
    - gocyclo         # Cyclomatic complexity
    - gofmt           # Code formatting
    - gofumpt         # Stricter gofmt
    - goimports       # Import organization
    - misspell        # Typos in comments
    - revive          # Fast golint replacement
    - stylecheck      # Style consistency

    # Bugs and correctness
    - bodyclose       # HTTP body close
    - dupl            # Code duplication
    - errname         # Error naming (ErrXxx)
    - errorlint       # Error wrapping
    - goconst         # Repeated strings
    - gocritic        # Comprehensive checks
    - nilerr          # Nil error returns
    - nilnil          # Nil,nil returns
    - noctx           # HTTP req without context
    - prealloc        # Slice preallocation

    # Security
    - gosec           # Security issues

    # Performance
    - gomoddirectives # go.mod directives
    - makezero        # Slice append issues
    - unconvert       # Unnecessary conversions

    # Formatting
    - whitespace      # Leading/trailing whitespace

linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true
    exclude-functions:
      - (io.Closer).Close
      - (*database/sql.Rows).Close
      - (*database/sql.Stmt).Close
      - (*os.File).Close

  govet:
    enable-all: true
    disable:
      - fieldalignment # Too noisy

  staticcheck:
    checks: ["all"]

  gosec:
    severity: medium
    confidence: medium
    excludes:
      - G104 # Covered by errcheck
      - G304 # File path from variable (too strict)

  gocyclo:
    min-complexity: 15

  gofumpt:
    lang-version: "1.24"
    extra-rules: true

  goimports:
    local-prefixes: github.com/yourusername/yourproject

  misspell:
    locale: US
    ignore-words:
      - cancelled

  revive:
    confidence: 0.8
    rules:
      - name: blank-imports
      - name: context-as-argument
      - name: context-keys-type
      - name: dot-imports
      - name: error-return
      - name: error-strings
      - name: error-naming
      - name: exported
      - name: if-return
      - name: increment-decrement
      - name: var-naming
      - name: var-declaration
      - name: package-comments
      - name: range
      - name: receiver-naming
      - name: time-naming
      - name: unexported-return
      - name: indent-error-flow
      - name: errorf
      - name: empty-block
      - name: superfluous-else
      - name: unused-parameter
      - name: unreachable-code
      - name: redefines-builtin-id

  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance
    disabled-checks:
      - commentedOutCode
      - whyNoLint

  stylecheck:
    checks: ["all", "-ST1003"] # Disable ALL_CAPS check

  dupl:
    threshold: 100

issues:
  exclude-rules:
    # Exclude some linters from test files
    - path: _test\.go
      linters:
        - gocyclo
        - errcheck
        - dupl
        - gosec
        - goconst

    # Exclude main.go from some checks
    - path: cmd/.*main\.go
      linters:
        - gocyclo

    # Exclude generated files
    - path: \.pb\.go$
      linters:
        - all

    - path: mock_.*\.go$
      linters:
        - all

  exclude-dirs:
    - vendor
    - third_party
    - testdata
    - examples
    - Godeps

  exclude-files:
    - ".*\\.pb\\.go$"
    - ".*_gen\\.go$"

  max-issues-per-linter: 0
  max-same-issues: 0

  # Show all issues
  new: false
  new-from-rev: ""
  new-from-patch: ""

  fix: false # Don't auto-fix in CI

severity:
  default-severity: error
  rules:
    - linters:
        - revive
        - stylecheck
      severity: warning

    - linters:
        - gosec
        - govet
        - errcheck
      severity: error
```

### Minimal Config (Small Projects)

```yaml
# .golangci.yml
run:
  timeout: 3m

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofumpt
    - revive
    - gosec

linters-settings:
  gofumpt:
    extra-rules: true
```

## Key Linters

### errcheck - Unchecked Error Returns

```go
// ✗ Bad - errcheck violation
rows, _ := db.Query(ctx, query)

// ✓ Good - check error
rows, err := db.Query(ctx, query)
if err != nil {
    return fmt.Errorf("query: %w", err)
}
defer rows.Close()

// ✓ Good - explicit ignore
_, _ = fmt.Fprintf(w, "message")
```

### govet - Go Vet Built-in

```go
// ✗ Bad - printf format mismatch
fmt.Printf("user %d", userName) // Expects int, got string

// ✓ Good
fmt.Printf("user %s", userName)

// ✗ Bad - unreachable code
return nil
fmt.Println("never printed")

// ✓ Good
fmt.Println("printed")
return nil
```

### staticcheck - Static Analysis

```go
// ✗ Bad - ineffective append
s := []int{1, 2, 3}
append(s, 4) // Result not assigned

// ✓ Good
s = append(s, 4)

// ✗ Bad - deprecated function
ioutil.ReadFile(path) // Deprecated in Go 1.16

// ✓ Good
os.ReadFile(path)
```

### gosec - Security Issues

```go
// ✗ Bad - G104: unhandled error
file.Close()

// ✓ Good
if err := file.Close(); err != nil {
    log.Printf("close file: %v", err)
}

// ✗ Bad - G401: weak crypto
hash := md5.New()

// ✓ Good
hash := sha256.New()

// ✗ Bad - G404: weak random
rand.Int() // math/rand, not crypto/rand

// ✓ Good
rand.Int(rand.Reader, big.NewInt(100)) // crypto/rand
```

### gocyclo - Cyclomatic Complexity

```go
// ✗ Bad - complexity > 15
func ValidateUser(u User) error {
    if u.Name == "" {
        return ErrEmptyName
    }
    if len(u.Name) < 3 {
        return ErrNameTooShort
    }
    if len(u.Name) > 50 {
        return ErrNameTooLong
    }
    // ... 15 more branches
}

// ✓ Good - extract validation
func ValidateUser(u User) error {
    if err := validateName(u.Name); err != nil {
        return err
    }
    if err := validateEmail(u.Email); err != nil {
        return err
    }
    return nil
}
```

### ineffassign - Unused Assignments

```go
// ✗ Bad - x assigned but never used
func process() int {
    x := 10
    x = 20
    return 30
}

// ✓ Good
func process() int {
    x := 20
    return x
}
```

### misspell - Comment Typos

```go
// ✗ Bad - misspelled comment
// Connnect to databse and retrive user

// ✓ Good
// Connect to database and retrieve user
```

### gofmt/gofumpt - Code Formatting

```bash
# gofmt - standard formatter
gofmt -w .

# gofumpt - stricter version (recommended)
gofumpt -l -w .

# Integrated in golangci-lint
golangci-lint run --fix
```

## Custom Rules

### Project-Specific Exclude Patterns

```yaml
issues:
  exclude-rules:
    # Allow todos in development
    - linters:
        - godox
      source: "TODO|FIXME"

    # Ignore complexity in legacy code
    - path: internal/legacy/
      linters:
        - gocyclo
        - gocognit

    # Allow long lines in generated code
    - path: \.gen\.go$
      linters:
        - lll

    # Disable security checks in test code
    - path: _test\.go$
      text: "G404:" # Weak random is OK in tests
      linters:
        - gosec
```

### Severity Levels

```yaml
severity:
  default-severity: error

  rules:
    # Warnings for style issues
    - linters:
        - revive
        - stylecheck
        - misspell
      severity: warning

    # Errors for critical issues
    - linters:
        - gosec
        - govet
        - errcheck
        - staticcheck
      severity: error

    # Info for suggestions
    - linters:
        - gocritic
        - prealloc
      severity: info
```

## CI Integration

### GitHub Actions (Recommended)

```yaml
# .github/workflows/lint.yml
name: Lint

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

jobs:
  golangci:
    name: lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: v2.0.0
          args: --timeout=5m

          # Performance
          skip-cache: false
          skip-pkg-cache: false
          skip-build-cache: false

          # Output
          only-new-issues: true

          # Annotations
          annotations: true
```

### With Test Coverage

```yaml
- name: Run tests
  run: go test -race -coverprofile=coverage.txt -covermode=atomic ./...

- name: Lint
  uses: golangci/golangci-lint-action@v6
  with:
    version: v2.0.0

- name: Upload coverage
  uses: codecov/codecov-action@v4
  with:
    file: ./coverage.txt
```

### Fail-Fast CI

```yaml
- name: Lint (fail-fast)
  uses: golangci/golangci-lint-action@v6
  with:
    version: v2.0.0
    args: --max-issues-per-linter=0 --max-same-issues=0
    only-new-issues: false
```

### Performance Tuning

```yaml
# .golangci.yml
run:
  concurrency: 4 # Number of CPUs
  timeout: 5m
  allow-parallel-runners: true

  # Skip heavy linters for speed
  skip-dirs:
    - vendor
    - third_party

  skip-files:
    - ".*\\.pb\\.go$"

linters-settings:
  govet:
    enable-all: false # Enable only specific checks
    enable:
      - assign
      - atomic
      - bools
      - buildtag
```

## IDE Integration

### VSCode

```json
// .vscode/settings.json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": [
    "--fast",
    "--config=${workspaceFolder}/.golangci.yml"
  ],
  "go.lintOnSave": "package",
  "go.formatTool": "gofumpt",
  "editor.formatOnSave": true,
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": "explicit"
    }
  }
}
```

### GoLand / IntelliJ IDEA

```
Settings → Tools → File Watchers → Add golangci-lint

Program: golangci-lint
Arguments: run --fix $FilePathRelativeToProjectRoot$
Working directory: $ProjectFileDir$
```

Or use built-in integration:

```
Settings → Tools → golangci-lint
☑ Enable golangci-lint
Path: /path/to/golangci-lint
Config: .golangci.yml
☑ Run on save
```

### Vim/Neovim

```vim
" Using ALE
let g:ale_linters = {
\   'go': ['golangci-lint'],
\}
let g:ale_fixers = {
\   'go': ['gofumpt'],
\}
let g:ale_go_golangci_lint_options = '--fast'
let g:ale_fix_on_save = 1
```

## Common Mistakes

| Mistake                         | Fix                                    |
|---------------------------------|----------------------------------------|
| Using deprecated linters        | Disable golint, use revive instead     |
| Too strict for legacy code      | Add exclude rules for old packages     |
| Not caching in CI               | Enable all cache options in actions    |
| Ignoring gosec warnings         | Review security issues, fix or justify |
| Running without config          | Always use .golangci.yml               |
| Enabling all linters            | Start with essentials, add gradually   |
| Not fixing auto-fixable issues  | Run with --fix flag                    |
| Ignoring exit codes             | Fail CI on linter errors               |

## Best Practices

```bash
# ✓ Good - run on entire project
golangci-lint run ./...

# ✗ Bad - missing directories
golangci-lint run

# ✓ Good - auto-fix where possible
golangci-lint run --fix

# ✗ Bad - manual fixes only
golangci-lint run

# ✓ Good - clear cache after config changes
golangci-lint cache clean
golangci-lint run

# ✗ Bad - stale cache
golangci-lint run

# ✓ Good - fail CI on issues
golangci-lint run --max-issues-per-linter=0

# ✗ Bad - allow issues
golangci-lint run || true
```

## Workflow Integration

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running golangci-lint..."
golangci-lint run --fix

if [ $? -ne 0 ]; then
    echo "Linting failed. Commit aborted."
    exit 1
fi

git add -u
```

### Make Targets

```makefile
# Makefile
.PHONY: lint
lint:
	golangci-lint run --timeout=5m

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix --timeout=5m

.PHONY: lint-fast
lint-fast:
	golangci-lint run --fast

.PHONY: lint-new
lint-new:
	golangci-lint run --new-from-rev=origin/main
```

### Docker Integration

```dockerfile
# Dockerfile.lint
FROM golangci/golangci-lint:v2.0.0

WORKDIR /app
COPY . .

RUN golangci-lint run --timeout=5m
```

## See Also

- [CI/CD](./ci-cd.md) - Continuous integration patterns
- [Docker](./docker.md) - Container configuration
- [Style Guide](../01-fundamentals/style-guide.md) - Go style conventions
- [Error Handling](../02-language/error-handling.md) - Error patterns
- [Idioms](../01-fundamentals/idioms.md) - Go best practices
- [golangci-lint documentation](https://golangci-lint.run/)
- [Linter list](https://golangci-lint.run/usage/linters/)
