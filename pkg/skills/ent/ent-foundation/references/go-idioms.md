# Go Idioms Quick Reference

Extracted from `docs/go/topics/01-fundamentals/idioms.md` (439 lines) → 100 lines of actionable patterns.

## Quick Reference Table

| Idiom                               | Principle                                |
|-------------------------------------|------------------------------------------|
| Accept interfaces, return structs   | Callers define interfaces they need      |
| Errors are values                   | Handle errors explicitly                 |
| Don't communicate by sharing memory | Share memory by communicating (channels) |
| Composition over inheritance        | Embed types, don't extend                |
| Clear is better than clever         | Simplicity wins                          |

## Accept Interfaces, Return Structs

```go
// Good - accept interface
func Save(w io.Writer, data []byte) error {
    _, err := w.Write(data)
    return err
}

// Good - return struct
func NewServer(addr string) *Server {
    return &Server{addr: addr}
}

// Bad - return interface (limits extensibility)
func NewServer(addr string) io.Closer {
    return &Server{addr: addr}
}
```

## Channels vs Mutexes

```go
// Channels: ownership transfer, work distribution, async results
func counter(inc <-chan struct{}) <-chan int {
    count := 0
    out := make(chan int)
    go func() {
        for {
            select {
            case <-inc: count++
            case out <- count:
            }
        }
    }()
    return out
}

// Mutexes: protecting shared state, performance-critical, simple structs
type Counter struct {
    mu    sync.Mutex
    value int
}
func (c *Counter) Inc() {
    c.mu.Lock()
    c.value++
    c.mu.Unlock()
}
```

## Errors are Values

```go
// Bad - ignoring errors
data, _ := os.ReadFile(path)

// Good - handle explicitly
data, err := os.ReadFile(path)
if err != nil {
    return fmt.Errorf("read file: %w", err)
}

// Good - custom error handling (check once at end)
type errWriter struct {
    w   io.Writer
    err error
}
func (ew *errWriter) write(buf []byte) {
    if ew.err != nil { return }
    _, ew.err = ew.w.Write(buf)
}
```

## Small Interfaces

```go
// Bad - kitchen sink (weak abstraction)
type Service interface {
    Create(User) error
    Read(id string) (User, error)
    Update(User) error
    Delete(id string) error
    List() ([]User, error)
    // ... 20 more methods
}

// Good - small focused interfaces
type Creator interface { Create(User) error }
type Reader interface { Read(id string) (User, error }

// Compose when needed
type UserService interface {
    Creator
    Reader
}
```

## Composition Over Inheritance

```go
// Good - embed types
type ReadWriter struct {
    *Reader
    *Writer
}

// Good - explicit embedding
type Logger struct {
    io.Writer
    prefix string
}
func (l *Logger) Log(msg string) {
    l.Write([]byte(l.prefix + msg))
}
```

## Make Zero Value Useful

```go
// Good - usable without initialization
var buf bytes.Buffer
buf.Write([]byte("data"))

var mu sync.Mutex
mu.Lock()
```
