# Domain: Generic

Generic development patterns that apply across projects.

## Project Structure Patterns

### Standard Layered Architecture

```
internal/
├── domain/       # Business entities and interfaces
├── repository/   # Data access layer
├── usecase/      # Business logic orchestration
└── transport/    # HTTP/gRPC handlers
```

### Package Organization

```
package/
├── {concept}/
│   ├── {impl}/
│   │   ├── repo.go
│   │   ├── models.go
│   │   └── mappers.go
│   └── contract.go
```

## Code Standards

### Naming Conventions

| Type | Convention | Example |
|------|-----------|---------|
| Variables | Short, meaningful | `cfg`, `repo`, `ctx` |
| Types | PascalCase | `User`, `Config` |
| Functions | PascalCase (exported), camelCase (private) | `NewUser()`, `validate()` |
| Constants | PascalCase | `MaxRetries` |
| Files | lowercase_with_underscores | `user_service.go` |
| Interfaces | PascalCase, simple | `Repository`, `Service` |

### Error Handling

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("operation name: %w", err)
}

// Custom errors
var ErrNotFound = errors.New("not found")
var ErrInvalidInput = errors.New("invalid input")
```

### Context Usage

```go
// First parameter, always pass through
func (s *Service) DoWork(ctx context.Context, req Request) error {
    // Use context for cancellation/timeouts
    select {
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

## Testing Patterns

### Table-Driven Tests

```go
tests := []struct {
    name    string
    input   Input
    want    Output
    wantErr bool
}{
    {"valid input", validInput, expectedOutput, false},
    {"invalid input", invalidInput, nil, true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        got, err := DoWork(tt.input)
        if (err != nil) != tt.wantErr {
            t.Errorf("DoWork() error = %v, wantErr %v", err, tt.wantErr)
            return
        }
        if !reflect.DeepEqual(got, tt.want) {
            t.Errorf("DoWork() = %v, want %v", got, tt.want)
        }
    })
}
```

### Test Commands

```bash
# Run all tests
go test ./... -race

# Run specific test
go test -run TestUser_Create ./internal/domain

# Verbose output
go test -v ./internal/...

# Coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Organization

```
package/
├── file.go
└── file_test.go
```

## Build and Validation

### Standard Build Commands

```bash
# Build
go build ./...

# Run linter
golangci-lint run

# Format code
goimports -w .

# Run all tests
go test ./... -race
```

### Validation Checklist

Before committing:
- [ ] Code builds without errors
- [ ] All tests pass (`go test ./... -race`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] Code is formatted (`goimports -w .`)
- [ ] No TODO comments (or documented tickets)
- [ ] Error messages are wrapped with context
- [ ] Context is passed through all layers

## Code Review Patterns

### Review Checklist

- [ ] Code follows project conventions
- [ ] Naming is clear and concise
- [ ] Error handling is proper
- [ ] Tests are sufficient and pass
- [ ] No unnecessary complexity
- [ ] Documentation is adequate
- [ ] No commented-out code
- [ ] No magic numbers

### Review Outcomes

- **APPROVED** - Ready to merge
- **CHANGES_REQUESTED** - Specific fixes needed
- **NEEDS_REBASE** - Merge conflicts exist
- **REJECTED** - Fundamental issues

## Agent Escalation Patterns

### Escalation Triggers

Escalate when:
- Task complexity score > 0.8
- Multiple integration points (>2)
- Security-critical implementation
- Unclear requirements after initial analysis
- Previous attempt failed

### Escalation Flow

```
fast agent → standard agent → heavy agent
```

Example:
```
task-fast → task-heavy (if complex) → coder → reviewer → tester → acceptor
```

## TDD Workflow

### Red-Green-Refactor

1. **RED**: Write failing test
   ```go
   func TestUser_Create(t *testing.T) {
       _, err := user.Create("")
       if err == nil {
           t.Fatal("expected error for empty name")
       }
   }
   ```

2. **GREEN**: Implement minimal solution
   ```go
   func (u *User) Create(name string) (*User, error) {
       if name == "" {
           return nil, ErrEmptyName
       }
       return &User{Name: name}, nil
   }
   ```

3. **REFACTOR**: Clean up
   ```go
   // Extract validation logic if reused
   ```

## Best Practices

### Code Style

- **Zero comments** explaining WHAT (fix naming instead)
- **Only WHY comments** for non-obvious reasons
- **Happy path left** (handle errors first)
- **Early returns** over deep nesting
- **Short functions** (<50 lines typically)

### Dependencies

- **Accept interfaces, return structs**
- **Prefer stdlib** over external deps
- **Minimal external dependencies**
- **Version pins** for reproducibility

### Performance

- **Measure before optimizing**
- **Profile with pprof**
- **Avoid premature optimization**
- **Consider cache locality**

## Common Patterns

### Repository Pattern

```go
type Repository interface {
    Find(ctx context.Context, id string) (*Entity, error)
    Save(ctx context.Context, e *Entity) error
}
```

### Service Pattern

```go
type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) DoWork(ctx context.Context, req Request) error {
    // Business logic
}
```

### Factory Pattern

```go
func New(config Config) (*Service, error) {
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("validate: %w", err)
    }
    return &Service{config: config}, nil
}
```

## Documentation Patterns

### Package Documentation

```go
// Package user provides user entity and business logic.
//
// The user domain handles authentication, authorization,
// and profile management.
package user
```

### Function Documentation

```go
// CreateUser creates a new user with the given email.
// Returns ErrInvalidEmail if email format is invalid.
func CreateUser(email string) (*User, error)
```

## Output Format Standards

### Status Output

```
═════════════════════════════════════════
TASK: {task-id}
═════════════════════════════════════════

📋 Task: {description}
   Priority: {priority}
   Dependencies: {count}

🔨 Implementation:
   Files modified: {count}
   Lines added: +{num}
   Lines removed: -{num}

🧪 Testing:
   Tests written: {count}
   Coverage: {percent}%

✅ Validation:
   Build: PASS
   Tests: PASS
   Lint: PASS

<promise>COMPLETE</promise>
═════════════════════════════════════════
```

## Configuration Management

### Environment Variables

```go
type Config struct {
    Port    int    `env:"PORT" envDefault:"8080"`
    DBURL   string `env:"DATABASE_URL,required"`
    LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
}

func Load() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }
    return cfg, nil
}
```

## Logging

### Structured Logging

```go
log := slog.Default()
log.Info("user created",
    "user_id", user.ID,
    "email", user.Email,
)

log.Error("failed to create user",
    "email", req.Email,
    "error", err,
)
```

## Concurrency

### Goroutine Safety

```go
type SafeCounter struct {
    mu    sync.Mutex
    value int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}
```

### Context Cancellation

```go
func worker(ctx context.Context) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Do work
        }
    }
}
```

## Deployment Readiness

### Health Checks

```go
func (s *Service) Health(ctx context.Context) error {
    if err := s.db.Ping(ctx); err != nil {
        return fmt.Errorf("database ping: %w", err)
    }
    return nil
}
```

### Graceful Shutdown

```go
func (s *Service) Shutdown(ctx context.Context) error {
    shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    return s.server.Shutdown(shutdownCtx)
}
```

## Common Gotchas

### Time

```go
// Use time.UTC for storage
now := time.Now().UTC()

// Duration calculations
deadline := time.Now().Add(5 * time.Minute)
```

### JSON

```go
// Handle nil pointers gracefully
type User struct {
    Name *string `json:"name,omitempty"`
}
```

### SQL Injection

```go
// ALWAYS use parameterized queries
query := "SELECT * FROM users WHERE id = $1"
db.QueryRow(query, userID)
```

## Metrics and Observability

### Prometheus Metrics

```go
var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request latency",
        },
        []string{"method", "path"},
    )
)

func init() {
    prometheus.MustRegister(requestDuration)
}
```

## Security Patterns

### Input Validation

```go
func validateEmail(email string) error {
    if email == "" {
        return ErrEmptyEmail
    }
    if !strings.Contains(email, "@") {
        return ErrInvalidEmail
    }
    return nil
}
```

### Secret Management

```go
// NEVER log secrets
// Use environment variables or secret stores
func (s *Service) Connect() error {
    apiKey := os.Getenv("API_KEY")
    // Don't log this!
    return nil
}
```
