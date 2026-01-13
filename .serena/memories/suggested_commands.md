# Go Ent - Suggested Commands

## Build Commands

```bash
make build          # Build MCP server to bin/ent
make test           # Run tests with race detector and coverage
make lint           # Run golangci-lint
make fmt            # Format code with goimports
make clean          # Remove build artifacts
make validate-plugin # Validate plugin JSON files
```

## Running Tests

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

## Development Workflow

```bash
# Build and restart MCP server
make build
# Then restart Claude Code to load changes

# Validate plugin JSON files
make validate-plugin

# Format code
make fmt

# Lint code
make lint

# Run tests
make test
```

## CLI Commands (Standalone)

```bash
# Initialize configuration
go-ent config init

# View configuration
go-ent config show
go-ent config show --format summary

# Modify configuration
go-ent config set budget.daily 25
go-ent config set agents.default architect

# List agents
go-ent agent list
go-ent agent list --detailed

# Get agent info
go-ent agent info architect

# List skills
go-ent skill list
go-ent skill list --detailed

# Get skill info
go-ent skill info go-arch

# Initialize OpenSpec
go-ent spec init

# List specs
go-ent spec list spec
go-ent spec list change

# Show spec or change
go-ent spec show spec api
go-ent spec show change add-authentication
```

## Claude Code Slash Commands

### Planning & Workflow
- `/plan <feature>` - Comprehensive planning workflow
- `/clarify <change-id>` - Ask focused questions
- `/research <change-id> [topic]` - Structured research
- `/decompose <change-id>` - Break into tasks
- `/analyze <change-id>` - Cross-document validation

### Execution
- `/apply` - Execute tasks from OpenSpec
- `/loop <task-description>` - Start autonomous loop
- `/loop-cancel` - Cancel running loop
- `/tdd` - Test-driven development cycle

### Project Management
- `/init <project-name>` - Initialize new Go project
- `/scaffold <type> <name>` - Scaffold Go components
- `/gen` - Generate from OpenAPI/Proto specs
- `/status` - View OpenSpec changes status
- `/archive` - Archive completed change
- `/registry` - Manage task registry

### Quality
- `/lint` - Run Go linters

## Git Commands

```bash
# Standard git workflow
git status
git add .
git commit -m "message"
git push

# View diff
git diff
git diff --cached

# View history
git log --oneline
git show HEAD
```

## System Commands (Linux)

```bash
# File operations
ls -la
pwd
cd /path/to/dir
mkdir -p dirname
rm file
rm -rf dirname

# Search files
find . -name "*.go"
grep -r "pattern" .

# View files
cat file.txt
less file.txt
head -n 20 file.txt
tail -n 20 file.txt

# Process management
ps aux
kill PID
kill -9 PID

# System info
df -h
du -sh .
top
htop
```

## Go-specific Commands

```bash
# Build
go build -o bin/ent ./cmd/go-ent

# Run
go run ./cmd/go-ent

# Test
go test ./...
go test -v ./internal/...
go test -race ./...

# Format
goimports -w .
gofmt -w .

# Vet
go vet ./...

# Tidy
go mod tidy

# Download
go mod download
```

## Quick Reference

**After changes:**
```bash
make fmt
make lint
make test
```

**Before coding:**
- Read existing patterns in the package
- Check if task needs clarification
- Understand dependencies and integration points

**Code review checklist:**
- [ ] `make lint` passes
- [ ] `make test` passes
- [ ] Zero comments except WHY
- [ ] Natural variable names
- [ ] Domain has zero external deps
- [ ] Interfaces at consumer side
- [ ] Errors wrapped with context
- [ ] Context propagated (`ctx` first)
- [ ] No magic numbers
- [ ] Happy path left
- [ ] Code looks human-written
