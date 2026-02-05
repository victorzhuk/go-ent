# Contributing to go-ent

Thank you for your interest in contributing to go-ent! This document provides guidelines and information for contributors.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Standards](#code-standards)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Documentation](#documentation)
- [Issue Guidelines](#issue-guidelines)

---

## Code of Conduct

### Our Standards

- **Be respectful**: Treat all contributors with respect
- **Be collaborative**: Work together to improve the project
- **Be constructive**: Provide helpful feedback
- **Be inclusive**: Welcome diverse perspectives

### Enforcement

Violations of the code of conduct can be reported to the project maintainers. All complaints will be reviewed and investigated.

---

## Getting Started

### Prerequisites

- **Go 1.24+**: Install from [golang.org](https://golang.org/dl/)
- **Make**: Standard build tool
- **Git**: Version control
- **Claude Code**: For testing the plugin (optional)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-ent.git
   cd go-ent
   ```

3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/victorzhuk/go-ent.git
   ```

### Build the Project

```bash
# Install dependencies
make deps

# Build MCP server
make build

# Run tests
make test

# Run linter
make lint
```

### Verify Installation

```bash
# Check version
./dist/go-ent --version

# Initialize test project
mkdir test-project && cd test-project
../dist/go-ent init
```

---

## Development Workflow

### Using go-ent to Develop go-ent

go-ent uses **self-hosted development** (dogfooding):

```bash
# 1. Build the MCP server
make build

# 2. Restart Claude Code to load the plugin

# 3. Use go-ent commands for development
/ent:plan <feature description>
/ent:task
/ent:bug <bug description>
```

See [DEVELOPMENT.md](./DEVELOPMENT.md) for detailed self-hosted workflow.

### Branch Strategy

- `master`: Stable, production-ready code
- `develop`: Integration branch for features (optional)
- `feature/*`: Feature branches
- `fix/*`: Bug fix branches
- `docs/*`: Documentation branches

### Creating a Feature Branch

```bash
# Update your fork
git fetch upstream
git checkout master
git merge upstream/master

# Create feature branch
git checkout -b feature/your-feature-name
```

### Making Changes

1. **Write code** following [Code Standards](#code-standards)
2. **Add tests** for new functionality
3. **Update documentation** as needed
4. **Run linter and tests**:
   ```bash
   make lint
   make test
   ```

5. **Commit with conventional commits**:
   ```bash
   git commit -m "feat(cli): Add new command for XYZ"
   git commit -m "fix(mcp): Fix tool registration bug"
   git commit -m "docs: Update configuration reference"
   ```

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance

**Scopes:**
- `cli`: CLI commands
- `mcp`: MCP server
- `skill`: Skill system
- `agent`: Agent system
- `openspec`: OpenSpec workflow
- `core`: Core functionality

**Examples:**
```
feat(cli): Add spec validation command

Add `ent spec validate` command to check spec format and content.

Closes #123
```

```
fix(mcp): Fix tool registration in Claude Code

Tool names weren't being registered correctly due to missing prefix.

Fixes #456
```

---

## Code Standards

### Go Code Style

Follow these standards (see [CLAUDE.md](../CLAUDE.md) for details):

**Naming:**
- Variables: `cfg`, `repo`, `srv`, `pool` (concise, not verbose)
- Constructors: `New()` for public, `new*()` for private
- Structs: Private by default (`type app struct`), public for domain
- Receivers: Short (`s *service`, `u *User`)

**Error Handling:**
```go
// Good
return fmt.Errorf("query user %s: %w", id, err)

// Bad
return fmt.Errorf("Failed to query user: %w", err)  // uppercase
return err  // no context
```

**Code Organization:**
1. `init()` (if needed)
2. `main()` or public types
3. Public functions
4. Private functions
5. Constants/errors at package top

**Imports:**
```go
import (
    "context"
    "fmt"

    "github.com/victorzhuk/go-ent/internal/domain"

    "github.com/google/uuid"
)
```

### Avoid

- Magic numbers (use named constants)
- `helper`/`util` packages (use domain names)
- Panic in production code
- `any` without reason
- Global mutable state
- Comments explaining WHAT (comments explain WHY only)

### Architecture

- **Clean Architecture** with DDD bounded contexts
- **Repository pattern** for data access
- **Dependency injection** via constructors
- **Interfaces at consumer side**

Layers:
```
Transport → UseCase → Domain ← Repository ← Infrastructure
```

Dependencies flow **inward only**.

---

## Testing Requirements

### Test Coverage

- **Unit tests**: Required for all new functions
- **Integration tests**: Required for MCP tools
- **E2E tests**: Required for CLI commands

### Writing Tests

**Table-driven tests:**
```go
func TestParseConfig(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Config
        wantErr bool
    }{
        {"valid config", "...", &Config{...}, false},
        {"invalid yaml", "...", nil, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got, err := ParseConfig(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Running Tests

```bash
# All tests
make test

# Specific package
go test ./internal/spec/...

# With coverage
make test-coverage

# Integration tests only
go test -tags=integration ./...
```

### Test Standards

- Use `testify/assert` for assertions
- Use `t.Parallel()` for independent tests
- Context is `t.Context()`
- Real implementations over mocks when simple
- `testcontainers` for integration tests

---

## Pull Request Process

### Before Submitting

1. **Ensure tests pass**:
   ```bash
   make test
   ```

2. **Run linter**:
   ```bash
   make lint
   ```

3. **Update documentation**:
   - Add/update relevant docs
   - Update CHANGELOG.md (Unreleased section)

4. **Commit with conventional commits**

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

### Creating Pull Request

1. Go to GitHub and create PR from your fork to `victorzhuk/go-ent:master`
2. Fill out the PR template:
   - **Description**: What does this PR do?
   - **Motivation**: Why is this change needed?
   - **Testing**: How was this tested?
   - **Checklist**: Complete the checklist

**PR Title Format:**
```
feat(scope): Brief description
```

**Example PR Description:**
```markdown
## Description
Adds a new `ent spec validate` command to validate OpenSpec structure and content.

## Motivation
Users need a way to verify their specs before committing. Currently validation only happens during `ent spec create`.

## Testing
- Added unit tests for validation logic
- Added integration test for CLI command
- Tested manually with various spec formats

## Checklist
- [x] Tests pass
- [x] Linter passes
- [x] Documentation updated
- [x] CHANGELOG.md updated
```

### Review Process

1. **Automated checks** run (tests, lint, build)
2. **Maintainer review** (1-2 business days)
3. **Address feedback** if requested
4. **Approval and merge** by maintainer

### After Merge

1. **Delete your branch**:
   ```bash
   git branch -d feature/your-feature-name
   git push origin --delete feature/your-feature-name
   ```

2. **Update your fork**:
   ```bash
   git checkout master
   git fetch upstream
   git merge upstream/master
   git push origin master
   ```

---

## Documentation

### Documentation Standards

- **Markdown**: GitHub-flavored markdown
- **Code examples**: Include syntax highlighting
- **Cross-references**: Link related docs
- **Up-to-date**: Keep examples current with codebase

### Documentation Files

| File | Purpose |
|------|---------|
| `README.md` | Project overview |
| `docs/INDEX.md` | Documentation hub |
| `docs/*.md` | Specific topics |
| `CHANGELOG.md` | Version history |
| Code comments | Why, not what |

### Updating Documentation

When adding features:

1. Update relevant docs in `docs/`
2. Add examples to `docs/CLI_EXAMPLES.md`
3. Update `docs/INDEX.md` if adding new doc
4. Update `CHANGELOG.md` (Unreleased section)

---

## Issue Guidelines

### Reporting Bugs

Use the bug report template:

```markdown
**Describe the bug**
Clear description of the bug.

**To Reproduce**
Steps to reproduce:
1. Run command '...'
2. See error

**Expected behavior**
What should happen.

**Environment:**
- go-ent version: [e.g., v0.3.0]
- Go version: [e.g., 1.24.0]
- OS: [e.g., macOS 15.2]
- Claude Code version: [e.g., 1.2.3]

**Additional context**
Any other relevant information.
```

### Requesting Features

Use the feature request template:

```markdown
**Is your feature request related to a problem?**
Clear description of the problem.

**Describe the solution you'd like**
What you want to happen.

**Describe alternatives you've considered**
Other solutions you've considered.

**Additional context**
Any other context or screenshots.
```

### Issue Labels

| Label | Purpose |
|-------|---------|
| `bug` | Something isn't working |
| `feature` | New feature request |
| `docs` | Documentation improvement |
| `good first issue` | Good for newcomers |
| `help wanted` | Extra attention needed |
| `wontfix` | Will not be fixed |

---

## Development Tips

### Hot Reload

During development, rebuild and restart Claude Code:

```bash
make build
# Restart Claude Code
```

For faster iteration, use the MCP inspector:
```bash
make mcp-inspector
```

### Debugging

Enable debug logging:

```bash
LOG_LEVEL=debug ./dist/go-ent
```

Use MCP inspector for tool testing:
```bash
make mcp-inspector
# Test tools interactively
```

### Common Tasks

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests with coverage
make test-coverage

# Build for all platforms
make build-all

# Clean build artifacts
make clean

# View available commands
make help
```

---

## Recognition

Contributors are recognized in:

- README.md Contributors section
- Release notes
- Git history

Thank you for contributing to go-ent! 🎉

---

## Questions?

- **Documentation**: See [docs/INDEX.md](./INDEX.md)
- **Development**: See [DEVELOPMENT.md](./DEVELOPMENT.md)
- **Issues**: [GitHub Issues](https://github.com/victorzhuk/go-ent/issues)
- **Discussions**: [GitHub Discussions](https://github.com/victorzhuk/go-ent/discussions)

---

**Version:** 2026-01-28
