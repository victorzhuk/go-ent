---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
version: "${VERSION}"
author: "${AUTHOR}"
tags: [${TAGS}]
triggers:
  - pattern: "debug|troubleshoot|investigate|diagnose|fix bug"
    weight: 0.9
  - keywords: ["debug", "troubleshoot", "investigate", "diagnose", "error", "bug", "issue", "problem"]
    weight: 0.8
  - filePattern: "*_debug.go"
    weight: 0.7
---

# ${SKILL_NAME}

<role>
Debugging expert specializing in systematic troubleshooting, effective logging, distributed tracing, and debugging strategies. 
Focus on root cause analysis, diagnostic techniques, and efficient problem-solving approaches.
</role>

<instructions>

## Systematic Debugging Methodology

Follow a structured approach to debugging:

### 1. Reproduce the Issue

```go
// Create minimal reproduction
func main() {
    // Isolate the problematic code
    result := ProcessInput("problematic_value")
    fmt.Printf("Result: %v\n", result)
    
    // Vary inputs to find patterns
    for _, input := range []string{"a", "b", "c"} {
        fmt.Printf("Input: %s, Result: %v\n", input, ProcessInput(input))
    }
}
```

**Key steps:**
- Identify exact conditions that trigger the bug
- Create minimal reproduction case
- Vary inputs to understand patterns
- Remove irrelevant code

### 2. Add Logging Strategically

```go
import "log/slog"

// Use structured logging with context
func Process(ctx context.Context, data string) error {
    logger := slog.FromContext(ctx)
    
    logger.Debug("starting processing",
        "data", data,
        "data_length", len(data),
    )
    
    if err := validate(data); err != nil {
        logger.Warn("validation failed",
            "error", err,
            "data", data,
        )
        return fmt.Errorf("validation: %w", err)
    }
    
    logger.Info("processing completed successfully")
    return nil
}
```

**Logging levels:**
- **Debug**: Detailed diagnostic information
- **Info**: Normal operation milestones
- **Warn**: Unexpected but recoverable conditions
- **Error**: Errors that prevent normal operation

**Best practices:**
- Use structured logging (key-value pairs)
- Include context (request ID, user ID, trace ID)
- Avoid logging sensitive data
- Use appropriate log levels

### 3. Add Tracing

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func Process(ctx context.Context, data string) error {
    tracer := otel.Tracer("processor")
    ctx, span := tracer.Start(ctx, "Process",
        trace.WithAttributes(
            attribute.String("data.length", strconv.Itoa(len(data))),
        ),
    )
    defer span.End()
    
    if err := validate(ctx, data); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "validation failed")
        return err
    }
    
    span.SetStatus(codes.Ok, "processed successfully")
    return nil
}
```

**Tracing benefits:**
- See flow across service boundaries
- Identify latency hotspots
- Understand failure propagation
- Correlate logs with traces

### 4. Use Debug Flags

```go
var debugMode = flag.Bool("debug", false, "Enable debug logging")

func Process(data string) {
    if *debugMode {
        fmt.Printf("DEBUG: Processing data: %q\n", data)
    }
    
    result := transform(data)
    
    if *debugMode {
        fmt.Printf("DEBUG: Transformed result: %q\n", result)
    }
}
```

### 5. Inspect State

```go
func debugPrintState(m map[string]int) {
    fmt.Println("Current state:")
    for k, v := range m {
        fmt.Printf("  %s: %d\n", k, v)
    }
}

func Process(m map[string]int) {
    debugPrintState(m)
    
    m["counter"]++
    
    debugPrintState(m)
}
```

## Common Debugging Techniques

### Printf Debugging

```go
func Calculate(a, b int) int {
    fmt.Printf("DEBUG: Calculate called with a=%d, b=%d\n", a, b)
    
    result := a * b
    
    fmt.Printf("DEBUG: Intermediate result: %d\n", result)
    fmt.Printf("DEBUG: Returning: %d\n", result)
    
    return result
}
```

### Assertion Debugging

```go
import "github.com/stretchr/testify/assert"

func TestProcess(t *testing.T) {
    // Add assertions to catch state changes
    initial := getState()
    Process(input)
    final := getState()
    
    assert.NotEqual(t, initial, final, "state should change")
    assert.Equal(t, expectedState, final, "state should match expected")
}
```

### Binary Search Debugging

```go
func ComplexLogic(data []string) error {
    midpoint := len(data) / 2
    fmt.Printf("DEBUG: Processing items 0-%d\n", midpoint)
    err := ProcessHalf(data[:midpoint])
    if err != nil {
        return err
    }
    
    fmt.Printf("DEBUG: Processing items %d-%d\n", midpoint, len(data))
    return ProcessHalf(data[midpoint:])
}
```

## Error Analysis Patterns

### Error Context Collection

```go
type DetailedError struct {
    Op    string
    Path  string
    Err   error
    Stack []byte
}

func (e *DetailedError) Error() string {
    return fmt.Sprintf("%s %s: %v", e.Op, e.Path, e.Err)
}

func (e *DetailedError) Unwrap() error {
    return e.Err
}

func ReadFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return &DetailedError{
            Op:    "read",
            Path:  path,
            Err:   err,
            Stack: debug.Stack(),
        }
    }
    
    return ProcessData(data)
}
```

### Error Type Inspection

```go
func handleError(err error) {
    switch {
    case errors.Is(err, ErrNotFound):
        fmt.Println("Resource not found")
    case errors.Is(err, ErrPermission):
        fmt.Println("Permission denied")
    case errors.As(err, &validationErr):
        fmt.Printf("Validation error: %v\n", validationErr)
    default:
        fmt.Printf("Unexpected error: %v\n", err)
    }
}
```

## Performance Debugging

### Timing Operations

```go
import "time"

func ProcessWithTiming(data string) error {
    start := time.Now()
    
    defer func() {
        duration := time.Since(start)
        fmt.Printf("DEBUG: Process took %v\n", duration)
    }()
    
    if err := Step1(data); err != nil {
        return err
    }
    
    step1Duration := time.Since(start)
    fmt.Printf("DEBUG: Step1 took %v\n", step1Duration)
    
    return Step2(data)
}
```

### Memory Profiling

```go
import (
    "os"
    "runtime/pprof"
)

func ProcessWithProfiling(data string) error {
    if os.Getenv("ENABLE_MEMORY_PROFILE") == "1" {
        f, err := os.Create("mem.prof")
        if err != nil {
            return err
        }
        defer f.Close()
        
        if err := pprof.StartCPUProfile(f); err != nil {
            return err
        }
        defer pprof.StopCPUProfile()
    }
    
    return Process(data)
}
```

## Concurrency Debugging

### Race Condition Detection

```bash
# Run tests with race detector
go test -race ./...

# Run binary with race detector
go run -race main.go
```

### Goroutine Dump

```go
import (
    "os"
    "runtime/debug"
)

func DumpGoroutines() {
    stack := debug.Stack()
    os.Stdout.Write(stack)
}

func HandlePanic() {
    if r := recover(); r != nil {
        fmt.Printf("Recovered from panic: %v\n", r)
        DumpGoroutines()
    }
}
```

### Channel Debugging

```go
func debugChannel[T any](ch <-chan T, name string) {
    fmt.Printf("DEBUG: Channel %s state: %d buffered, %d waiting\n",
        name, len(ch), cap(ch)-len(ch))
}

func Worker(ctx context.Context, in <-chan int, out chan<- int) {
    for {
        debugChannel(in, "in")
        debugChannel(out, "out")
        
        select {
        case <-ctx.Done():
            return
        case data, ok := <-in:
            if !ok {
                return
            }
            out <- data * 2
        }
    }
}
```

## Network Debugging

### Request/Response Logging

```go
import (
    "net/http/httputil"
)

type debugTransport struct {
    http.RoundTripper
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    dump, err := httputil.DumpRequestOut(req, true)
    if err != nil {
        return nil, err
    }
    fmt.Printf("DEBUG: Request:\n%s\n", dump)
    
    resp, err := t.RoundTripper.RoundTrip(req)
    if err != nil {
        return nil, err
    }
    
    dump, err = httputil.DumpResponse(resp, true)
    if err != nil {
        return nil, err
    }
    fmt.Printf("DEBUG: Response:\n%s\n", dump)
    
    return resp, nil
}

func NewDebugClient() *http.Client {
    return &http.Client{
        Transport: &debugTransport{},
    }
}
```

### Connection State Debugging

```go
func debugConnectionState(conn net.Conn) {
    fmt.Printf("DEBUG: Local: %s, Remote: %s\n",
        conn.LocalAddr(), conn.RemoteAddr())
    
    if tcpConn, ok := conn.(*net.TCPConn); ok {
        state, err := tcpConn.ConnectionState()
        if err == nil {
            fmt.Printf("DEBUG: State: %v\n", state)
        }
    }
}
```

## Database Debugging

### Query Logging

```go
import "database/sql"

type debugDB struct {
    *sql.DB
}

func (db *debugDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    fmt.Printf("DEBUG: Query: %s, Args: %v\n", query, args)
    return db.DB.QueryContext(ctx, query, args...)
}

func (db *debugDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    fmt.Printf("DEBUG: Exec: %s, Args: %v\n", query, args)
    return db.DB.ExecContext(ctx, query, args...)
}
```

### Transaction Debugging

```go
func debugTransaction(ctx context.Context, db *sql.DB) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer func() {
        if err != nil {
            fmt.Printf("DEBUG: Rolling back transaction\n")
            tx.Rollback()
        } else {
            fmt.Printf("DEBUG: Committing transaction\n")
            tx.Commit()
        }
    }()
    
    // Operations...
    return err
}
```

## Debugging Checklists

### Before Debugging
- [ ] Can I reproduce the issue consistently?
- [ ] Do I have the right error messages and stack traces?
- [ ] Do I understand the expected vs actual behavior?
- [ ] Do I have access to relevant logs and metrics?

### During Debugging
- [ ] Am I using minimal reproduction?
- [ ] Are my log statements at appropriate levels?
- [ ] Am I verifying each assumption?
- [ ] Am I documenting my findings?

### After Debugging
- [ ] Have I identified the root cause?
- [ ] Have I written a test that catches this bug?
- [ ] Have I cleaned up debug code?
- [ ] Have I documented the solution?

</instructions>

<constraints>
- Use structured logging with context and key-value pairs
- Include relevant context (request ID, trace ID) in logs
- Avoid logging sensitive data (passwords, tokens, PII)
- Use appropriate log levels (debug, info, warn, error)
- Add tracing for distributed systems
- Create minimal reproduction cases
- Document findings during debugging
- Clean up debug code before committing
- Write tests that reproduce bugs
- Verify fixes don't introduce regressions
- Use race detector for concurrent code
- Profile performance bottlenecks
</constraints>

<edge_cases>
If the issue is intermittent: Add extensive logging and monitoring to capture the next occurrence. Use stress testing to reproduce more frequently.

If logs are overwhelming: Use log sampling, adjust log levels dynamically, or add filtering mechanisms.

If debugging in production: Use feature flags to enable debug mode safely, ensure debug logging has minimal performance impact.

If error messages are generic: Add context to errors with wrap messages, include request/response snippets where safe.

If reproducing requires specific data: Create test data that matches production patterns, anonymize if needed.

If debugging distributed systems: Use distributed tracing (OpenTelemetry), correlate logs with trace IDs, check inter-service communication.

If memory leaks are suspected: Use pprof for heap profiling, check for goroutine leaks, monitor GC metrics.

If race conditions occur: Run with race detector (`-race`), add mutex logging, check shared state access patterns.

If performance is degraded: Profile with pprof, check for N+1 queries, analyze database query plans, monitor resource usage.
</edge_cases>

<examples>
<example>
<input>Debug a function that sometimes returns nil values</input>
<output>
```go
func FindUser(ctx context.Context, db *sql.DB, email string) (*User, error) {
    logger := slog.FromContext(ctx)
    
    logger.Debug("finding user", "email", email)
    
    var user User
    err := db.QueryRowContext(ctx,
        "SELECT id, name, email FROM users WHERE email = $1",
        email,
    ).Scan(&user.ID, &user.Name, &user.Email)
    
    if errors.Is(err, sql.ErrNoRows) {
        logger.Warn("user not found", "email", email)
        return nil, ErrNotFound
    }
    
    if err != nil {
        logger.Error("database error",
            "email", email,
            "error", err,
        )
        return nil, fmt.Errorf("query user: %w", err)
    }
    
    logger.Info("user found", "user_id", user.ID)
    
    // Validate we have all required fields
    if user.ID == "" || user.Email == "" {
        logger.Error("invalid user state",
            "user_id", user.ID,
            "email", user.Email,
        )
        return nil, fmt.Errorf("invalid user state: missing required fields")
    }
    
    return &user, nil
}
```
</output>
</example>

<example>
<input>Add debugging to a concurrent processor to identify race conditions</input>
<output>
```go
import (
    "sync"
    "sync/atomic"
)

type Processor struct {
    mu       sync.Mutex
    counter  atomic.Int64
    active   atomic.Int64
}

func (p *Processor) Process(data string) {
    p.active.Add(1)
    defer p.active.Add(-1)
    
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        fmt.Printf("DEBUG: Process completed in %v, active: %d\n",
            duration, p.active.Load())
    }()
    
    p.mu.Lock()
    defer p.mu.Unlock()
    
    fmt.Printf("DEBUG: Starting process for data: %q\n", data)
    
    result := transform(data)
    
    count := p.counter.Add(1)
    fmt.Printf("DEBUG: Processed item #%d, result: %q\n", count, result)
    
    store(result)
}

// Use with race detector
// go run -race main.go

// Test for race conditions
func TestProcessor_Concurrent(t *testing.T) {
    p := &Processor{}
    
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            p.Process(fmt.Sprintf("item-%d", id))
        }(i)
    }
    
    wg.Wait()
    
    fmt.Printf("Total processed: %d\n", p.counter.Load())
}
```
</output>
</example>
</examples>

<output_format>
Provide debugging solutions following established patterns:

1. **Logging**: Structured logging with context and appropriate levels
2. **Tracing**: Distributed tracing for cross-service issues
3. **Minimal Reproduction**: Isolate the issue with minimal code
4. **Context**: Include relevant context (IDs, timestamps, states)
5. **Error Handling**: Detailed error messages with wrap context
6. **Profiling**: Use pprof for performance and memory issues
7. **Race Detection**: Use `-race` flag for concurrent code
8. **Documentation**: Record findings and solutions

Focus on systematic, efficient debugging approaches.
</output_format>
