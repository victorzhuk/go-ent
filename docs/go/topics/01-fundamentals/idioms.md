# Go Idioms

Core Go idioms from Effective Go and Go Proverbs. These patterns define idiomatic Go code.

## Quick Reference

| Idiom                               | Principle                                |
|-------------------------------------|------------------------------------------|
| Accept interfaces, return structs   | Callers define interfaces they need      |
| Errors are values                   | Handle errors explicitly                 |
| Don't communicate by sharing memory | Share memory by communicating (channels) |
| Composition over inheritance        | Embed types, don't extend                |
| Clear is better than clever         | Simplicity wins                          |

## Go Proverbs

### Don't communicate by sharing memory, share memory by communicating

```go
// Bad - shared memory with mutex
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    c.value++
    c.mu.Unlock()
}

// Good - communicate via channels
func counter(inc <-chan struct{}) <-chan int {
    count := 0
    out := make(chan int)
    go func() {
        for {
            select {
            case <-inc:
                count++
            case out <- count:
            }
        }
    }()
    return out
}
```

**When to use channels:**
- Ownership transfer
- Distributing work
- Communicating async results

**When to use mutexes:**
- Protecting shared state
- Performance-critical sections
- Simple data structures

### Errors are values

```go
// Bad - ignoring errors
data, _ := os.ReadFile(path)

// Good - handle errors
data, err := os.ReadFile(path)
if err != nil {
    return fmt.Errorf("read file: %w", err)
}

// Good - custom error handling
type errWriter struct {
    w   io.Writer
    err error
}

func (ew *errWriter) write(buf []byte) {
    if ew.err != nil {
        return
    }
    _, ew.err = ew.w.Write(buf)
}

// Usage - check error once at end
ew := &errWriter{w: w}
ew.write([]byte("header\n"))
ew.write([]byte("body\n"))
ew.write([]byte("footer\n"))
if ew.err != nil {
    return ew.err
}
```

### Accept interfaces, return structs

```go
// Good - function accepts interface
func Save(w io.Writer, data []byte) error {
    _, err := w.Write(data)
    return err
}

// Good - function returns struct
func NewServer(addr string) *Server {
    return &Server{addr: addr}
}

// Bad - function returns interface
func NewServer(addr string) io.Closer {
    return &Server{addr: addr}
}
```

**Why:**
- Callers define interfaces they need (implicit satisfaction)
- Returning structs allows adding methods without breaking compatibility
- Smaller interfaces are easier to implement

### The bigger the interface, the weaker the abstraction

```go
// Bad - kitchen sink interface
type Service interface {
    Create(User) error
    Read(id string) (User, error)
    Update(User) error
    Delete(id string) error
    List() ([]User, error)
    Search(query string) ([]User, error)
    Validate(User) error
    // ... 20 more methods
}

// Good - small focused interfaces
type Creator interface {
    Create(User) error
}

type Reader interface {
    Read(id string) (User, error)
}

type Validator interface {
    Validate(User) error
}

// Compose when needed
type UserService interface {
    Creator
    Reader
}
```

### Make the zero value useful

```go
// Good - zero value is usable
var buf bytes.Buffer
buf.WriteString("hello") // Works immediately

var mu sync.Mutex
mu.Lock() // Works immediately

// Good - design types with useful zero values
type Server struct {
    Addr string // Empty string is valid
    mu   sync.Mutex // Zero value usable
}

func main() {
    var srv Server // Usable immediately
    srv.mu.Lock()
}
```

### Gofmt's style is no one's favorite, yet gofmt is everyone's favorite

**Message:** Consistency > personal preference

```bash
# Always run before committing
gofmt -w .

# Or use gofumpt (stricter)
gofumpt -w .
```

### A little copying is better than a little dependency

```go
// Bad - import entire library for one function
import "github.com/huge/library"

func process(data string) string {
    return library.OneSmallFunction(data)
}

// Good - copy small function if it avoids dependency
func process(data string) string {
    // Copied from github.com/huge/library
    // License: MIT
    return strings.ToUpper(data)
}
```

### Clear is better than clever

```go
// Clever - hard to understand
func calc(a, b int) int {
    return a&b + (a^b)>>1
}

// Clear - easy to understand
func average(a, b int) int {
    return (a + b) / 2
}

// Clever - single line
users := filter(map_(parse(strings.Split(input, ",")), toUser), active)

// Clear - multiple steps
lines := strings.Split(input, ",")
parsedUsers := parseUsers(lines)
activeUsers := filterActive(parsedUsers)
```

### Don't panic

```go
// Bad - panic for expected errors
func getUser(id string) User {
    user, err := db.Query(id)
    if err != nil {
        panic(err) // DON'T
    }
    return user
}

// Good - return error
func getUser(id string) (User, error) {
    user, err := db.Query(id)
    if err != nil {
        return User{}, fmt.Errorf("query user: %w", err)
    }
    return user, nil
}
```

**When to panic:**
- Programming errors (nil pointer, index out of range)
- Init failures (can't load config, can't connect to required service)
- Unrecoverable situations

## Effective Go Patterns

### Happy path left, errors right

```go
// Good
func process() error {
    data, err := load()
    if err != nil {
        return err
    }

    if err := validate(data); err != nil {
        return err
    }

    return save(data)
}

// Bad - nested if
func process() error {
    data, err := load()
    if err == nil {
        if validate(data) == nil {
            return save(data)
        }
    }
    return err
}
```

### Use defer for cleanup

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close() // Cleanup guaranteed

    // Process file
    return nil
}
```

### Getters don't need "Get" prefix

```go
// Good
type User struct {
    name string
}

func (u *User) Name() string {
    return u.name
}

// Bad
func (u *User) GetName() string {
    return u.name
}
```

### Use package names wisely

```go
// Bad - stuttering
user.UserService
http.HTTPServer
json.JSONEncoder

// Good
user.Service
http.Server
json.Encoder

// Bad - generic names
utils.Helper
common.Util

// Good - specific names
email.Validator
time.Parser
```

## Common Patterns

### Constructor pattern

```go
// Simple constructor
func NewServer(addr string) *Server {
    return &Server{addr: addr}
}

// Constructor with initialization
func NewPool(size int) (*Pool, error) {
    if size <= 0 {
        return nil, errors.New("size must be positive")
    }

    p := &Pool{
        workers: make(chan struct{}, size),
    }

    p.start()
    return p, nil
}
```

### Functional options

```go
type Server struct {
    addr    string
    timeout time.Duration
    logger  *log.Logger
}

type Option func(*Server)

func WithTimeout(d time.Duration) Option {
    return func(s *Server) {
        s.timeout = d
    }
}

func WithLogger(l *log.Logger) Option {
    return func(s *Server) {
        s.logger = l
    }
}

func NewServer(addr string, opts ...Option) *Server {
    s := &Server{
        addr:    addr,
        timeout: 30 * time.Second, // default
    }

    for _, opt := range opts {
        opt(s)
    }

    return s
}

// Usage
srv := NewServer(":8080",
    WithTimeout(10*time.Second),
    WithLogger(logger),
)
```

### Table-driven tests

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 2, 3, 5},
        {"negative", -2, -3, -5},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("got %d, want %d", got, tt.want)
            }
        })
    }
}
```

## See Also

- [Style Guide](./style-guide.md) - Uber + Google style consensus
- [Naming](./naming.md) - Naming conventions
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Proverbs](https://go-proverbs.github.io/)
