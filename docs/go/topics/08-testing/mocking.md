# Mocking

Test doubles using mockery and manual mocks.

## Quick Reference

| Approach                | Use When                                           |
|-------------------------|----------------------------------------------------|
| mockery                 | Auto-generate mocks from interfaces                |
| Manual mocks            | Simple interfaces, full control                    |
| Real implementation     | Cheap to create (in-memory storage, net.Listen)    |
| .mockery.yaml           | Project-wide mockery config                        |
| assert.*                | Assertions that continue test on failure           |
| require.*               | Assertions that stop test on failure               |
| testify/mock.Mock       | Call verification, argument matching               |

## Manual Mocks

```go
// Production code
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*User, error)
}

// Test mock
type MockUserRepository struct {
    GetByIDFunc func(ctx context.Context, id string) (*User, error)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
    if m.GetByIDFunc != nil {
        return m.GetByIDFunc(ctx, id)
    }
    return nil, errors.New("not implemented")
}

// Usage in test
func TestUserService(t *testing.T) {
    mockRepo := &MockUserRepository{
        GetByIDFunc: func(ctx context.Context, id string) (*User, error) {
            if id == "123" {
                return &User{ID: "123", Name: "Alice"}, nil
            }
            return nil, ErrNotFound
        },
    }

    svc := NewUserService(mockRepo)
    user, err := svc.GetUser(context.Background(), "123")

    require.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
}
```

## Mockery Configuration

Create `.mockery.yaml` in project root:

```yaml
with-expecter: true
dir: "internal/mocks"
packages:
  github.com/yourorg/project/internal/repository:
    config:
      all: true
      inpackage: false
      outpkg: mocks
    interfaces:
      UserRepository:
      OrderRepository:
  github.com/yourorg/project/internal/service:
    config:
      all: true
      inpackage: false
      outpkg: servicemocks
```

Generate mocks:

```bash
# Install
go install github.com/vektra/mockery/v2@latest

# Generate all configured mocks
mockery

# Generate specific interface
mockery --name=UserRepository
```

## Using Mockery

```go
import "myproject/internal/mocks"

func TestWithMockery(t *testing.T) {
    mockRepo := mocks.NewUserRepository(t)

    // Setup expectations
    mockRepo.EXPECT().
        GetByID(mock.Anything, "123").
        Return(&User{ID: "123", Name: "Alice"}, nil).
        Once()

    mockRepo.EXPECT().
        GetByID(mock.Anything, "456").
        Return(nil, ErrNotFound).
        Maybe()

    svc := NewUserService(mockRepo)
    user, err := svc.GetUser(context.Background(), "123")

    require.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)

    // Automatically verified at test end via t.Cleanup
}
```

### Advanced Mockery Features

```go
func TestAdvancedMocking(t *testing.T) {
    mockRepo := mocks.NewUserRepository(t)

    // Argument matching
    mockRepo.EXPECT().
        Save(mock.Anything, mock.MatchedBy(func(u *User) bool {
            return u.Email != "" && strings.Contains(u.Email, "@")
        })).
        Return(nil).
        Once()

    // Custom return logic
    mockRepo.EXPECT().
        GetByID(mock.Anything, mock.AnythingOfType("string")).
        RunAndReturn(func(ctx context.Context, id string) (*User, error) {
            if id == "" {
                return nil, errors.New("empty id")
            }
            return &User{ID: id}, nil
        })

    // Multiple calls
    mockRepo.EXPECT().
        Delete(mock.Anything, mock.Anything).
        Return(nil).
        Times(3)
}

## Testify Assertions

### assert vs require

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAssertions(t *testing.T) {
    // assert: test continues on failure
    assert.Equal(t, 42, result)
    assert.NoError(t, err)
    assert.True(t, condition)

    // require: test stops on failure (use for critical preconditions)
    require.NoError(t, err) // If this fails, no point continuing
    require.NotNil(t, user)

    // Use require when subsequent code depends on the assertion
    user, err := repo.GetByID(ctx, "123")
    require.NoError(t, err)          // Must succeed
    require.NotNil(t, user)          // Must exist
    assert.Equal(t, "Alice", user.Name) // Can fail independently
}
```

### Common Assertions with Mocks

```go
func TestMockAssertions(t *testing.T) {
    mockRepo := mocks.NewUserRepository(t)

    mockRepo.EXPECT().
        GetByID(mock.Anything, "123").
        Return(&User{ID: "123"}, nil)

    svc := NewUserService(mockRepo)
    user, err := svc.GetUser(context.Background(), "123")

    // Assertions
    require.NoError(t, err)
    require.NotNil(t, user)
    assert.Equal(t, "123", user.ID)
    assert.NotEmpty(t, user.Name)
    assert.Contains(t, user.Email, "@")
    assert.Greater(t, user.Age, 0)
    assert.Len(t, user.Orders, 2)

    // Mock-specific assertions
    mockRepo.AssertExpectations(t)
    mockRepo.AssertNumberOfCalls(t, "GetByID", 1)
    mockRepo.AssertCalled(t, "GetByID", mock.Anything, "123")
    mockRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}
```

## When to Mock

```go
// Mock external dependencies
type PaymentGateway interface {
    Charge(ctx context.Context, amount int) error
}

type EmailSender interface {
    Send(ctx context.Context, to, subject, body string) error
}

// Mock third-party APIs
type WeatherAPI interface {
    GetForecast(ctx context.Context, city string) (*Forecast, error)
}

// Mock slow operations
type ImageProcessor interface {
    Resize(ctx context.Context, img []byte, width, height int) ([]byte, error)
}

// Mock non-deterministic behavior
type TimeProvider interface {
    Now() time.Time
    After(d time.Duration) <-chan time.Time
}

type IDGenerator interface {
    Generate() (string, error)
}
```

## When NOT to Mock

Real implementations are better when they're fast, deterministic, and isolated:

```go
// In-memory storage - use real implementation
type InMemoryUserRepository struct {
    mu    sync.RWMutex
    users map[string]*User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
    return &InMemoryUserRepository{
        users: make(map[string]*User),
    }
}

func (r *InMemoryUserRepository) Save(ctx context.Context, u *User) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.users[u.ID] = u
    return nil
}

// Filesystem - use afero
import "github.com/spf13/afero"

func TestFileOperations(t *testing.T) {
    fs := afero.NewMemMapFs()
    svc := NewFileService(fs)

    err := svc.WriteConfig("config.json", []byte(`{"key":"value"}`))
    require.NoError(t, err)

    data, err := afero.ReadFile(fs, "config.json")
    require.NoError(t, err)
    assert.Contains(t, string(data), "value")
}

// Network listeners - use real net.Listen
func TestHTTPServer(t *testing.T) {
    ln, err := net.Listen("tcp", "127.0.0.1:0")
    require.NoError(t, err)
    defer ln.Close()

    srv := &http.Server{Handler: NewHandler()}
    go srv.Serve(ln)
    defer srv.Close()

    addr := ln.Addr().String()
    resp, err := http.Get("http://" + addr + "/health")
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Simple interfaces - use real implementation
type Logger interface {
    Info(msg string)
    Error(msg string)
}

type TestLogger struct {
    mu   sync.Mutex
    logs []string
}

func (l *TestLogger) Info(msg string) {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.logs = append(l.logs, "INFO: "+msg)
}

func (l *TestLogger) Error(msg string) {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.logs = append(l.logs, "ERROR: "+msg)
}
```

## Common Mistakes

| Mistake                        | Problem                                    | Solution                                  |
|--------------------------------|--------------------------------------------|-------------------------------------------|
| Mocking concrete types         | Can't mock, only interfaces                | Define interface, mock that               |
| Over-mocking                   | Tests become brittle, hard to refactor     | Use real implementations when simple      |
| Not resetting mocks            | State leaks between tests                  | Create new mock per test                  |
| Order-dependent assertions     | Fragile tests coupled to implementation    | Use `mock.Anything` or `Times(n)` not order |
| Missing mock verification      | Expectations never checked                 | Use `AssertExpectations(t)` or EXPECT()   |
| Mocking internal packages      | Tight coupling to implementation details   | Mock at service boundaries                |
| No argument validation         | Mocks accept invalid data                  | Use `mock.MatchedBy()` for validation     |
| Returning wrong error types    | Tests don't match production behavior      | Return actual error types, not generic    |

### Example: Fixing Common Mistakes

```go
// Bad: Mocking concrete type
func TestBad(t *testing.T) {
    // Can't mock UserService - it's a concrete type
    // mockSvc := &MockUserService{} // Won't work
}

// Good: Mock interface
type UserService interface {
    GetUser(ctx context.Context, id string) (*User, error)
}

func TestGood(t *testing.T) {
    mockSvc := mocks.NewUserService(t)
    mockSvc.EXPECT().GetUser(mock.Anything, "123").Return(&User{}, nil)
}

// Bad: Over-mocking simple cache
func TestOverMocked(t *testing.T) {
    mockCache := mocks.NewCache(t)
    mockCache.EXPECT().Set("key", "value").Return(nil)
    mockCache.EXPECT().Get("key").Return("value", true)
    // Fragile, coupled to implementation
}

// Good: Use real in-memory cache
func TestRealCache(t *testing.T) {
    cache := NewInMemoryCache()
    cache.Set("key", "value")
    val, ok := cache.Get("key")
    assert.True(t, ok)
    assert.Equal(t, "value", val)
}

// Bad: No argument validation
mockRepo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)

// Good: Validate arguments
mockRepo.EXPECT().
    Save(mock.Anything, mock.MatchedBy(func(u *User) bool {
        return u.ID != "" && u.Email != "" && strings.Contains(u.Email, "@")
    })).
    Return(nil)

// Bad: Order-dependent
mockRepo.On("GetByID", "1").Return(user1, nil)
mockRepo.On("GetByID", "2").Return(user2, nil)
// Expects exact call order

// Good: Use Times() or accept any order
mockRepo.EXPECT().GetByID(mock.Anything, "1").Return(user1, nil).Maybe()
mockRepo.EXPECT().GetByID(mock.Anything, "2").Return(user2, nil).Maybe()
```

## See Also

- [Table-Driven Tests](./table-driven.md) - Combine mocks with table-driven patterns
- [Integration Tests](./integration.md) - When to use real implementations with testcontainers
- [Benchmarks](./benchmarks.md) - Measure mock overhead vs real implementations
