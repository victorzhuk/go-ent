# Channels

Channels are Go's built-in mechanism for safe communication between goroutines. Understanding channel axioms prevents deadlocks and race conditions.

## Quick Reference

| Pattern                 | Use When                     |
|-------------------------|------------------------------|
| `ch := make(chan T)`    | Unbuffered synchronization   |
| `ch := make(chan T, n)` | Buffered async communication |
| `close(ch)`             | Signal "no more values"      |
| `val, ok := <-ch`       | Check if channel closed      |
| `select`                | Wait on multiple channels    |

## Channel Axioms

Understanding these axioms prevents most channel bugs:

| Operation   | Nil Channel   | Closed Channel            | Open Channel           |
|-------------|---------------|---------------------------|------------------------|
| **Send**    | Block forever | Panic                     | Block or succeed       |
| **Receive** | Block forever | Return zero value + false | Block or receive value |
| **Close**   | Panic         | Panic                     | Succeed                |

## Basic Patterns

### Unbuffered Channel (Synchronization)

```go
func main() {
    done := make(chan struct{}) // Unbuffered

    go func() {
        doWork()
        done <- struct{}{} // Send blocks until receive happens
    }()

    <-done // Receive blocks until send happens
    fmt.Println("Work completed")
}
```

**Key points:**
- Unbuffered channels guarantee synchronization
- Send and receive happen simultaneously
- Use `struct{}` for signaling (zero memory)

### Buffered Channel (Async Communication)

```go
func producer(ctx context.Context, out chan<- int) {
    defer close(out) // Signal completion

    for i := 0; i < 100; i++ {
        select {
        case out <- i: // Non-blocking if buffer not full
        case <-ctx.Done():
            return
        }
    }
}

func consumer(ctx context.Context, in <-chan int) {
    for {
        select {
        case val, ok := <-in:
            if !ok {
                return // Channel closed
            }
            process(val)
        case <-ctx.Done():
            return
        }
    }
}

func main() {
    ctx := context.Background()
    ch := make(chan int, 10) // Buffer size 10

    go producer(ctx, ch)
    consumer(ctx, ch)
}
```

### Range Over Channel

```go
func consumer(in <-chan int) {
    for val := range in { // Exits when channel closed
        process(val)
    }
}

func main() {
    ch := make(chan int)

    go func() {
        defer close(ch) // MUST close for range to exit
        for i := 0; i < 10; i++ {
            ch <- i
        }
    }()

    consumer(ch)
}
```

**Key points:**
- `for val := range ch` exits only when channel is closed
- Producer MUST close channel or consumer blocks forever
- Multiple consumers on same channel is race-prone (see fan-out pattern)

## Select Statement

### Basic Select

```go
func worker(ctx context.Context, in <-chan Task, out chan<- Result) {
    for {
        select {
        case task := <-in:
            out <- task.Process()

        case <-ctx.Done():
            return

        case <-time.After(30 * time.Second):
            log.Warn("no tasks for 30s")
        }
    }
}
```

### Non-Blocking Send/Receive

```go
select {
case ch <- value:
    fmt.Println("sent value")
default:
    fmt.Println("channel full, dropped value")
}

select {
case val := <-ch:
    fmt.Println("received:", val)
default:
    fmt.Println("no value available")
}
```

### Preventing Goroutine Leaks with Select

```go
func fetchWithTimeout(ctx context.Context, url string) (string, error) {
    result := make(chan string, 1) // Buffered to prevent sender leak

    go func() {
        data := fetch(url) // May be slow
        result <- data     // Won't block even if no receiver
    }()

    select {
    case data := <-result:
        return data, nil
    case <-ctx.Done():
        return "", ctx.Err()
    }
}
```

**Key points:**
- Use buffered channel (size 1) when sender may outlive receiver
- Prevents goroutine blocking forever on send

## Channel Direction

### Restricting Channel Direction

```go
// Producer can only send
func producer(out chan<- int) {
    defer close(out)
    for i := 0; i < 10; i++ {
        out <- i
    }
}

// Consumer can only receive
func consumer(in <-chan int) {
    for val := range in {
        fmt.Println(val)
    }
}

func main() {
    ch := make(chan int) // Bidirectional

    go producer(ch) // Implicitly converts to send-only
    consumer(ch)    // Implicitly converts to receive-only
}
```

**Benefits:**
- Compile-time safety (can't accidentally receive on send-only channel)
- Documents intent in function signature
- Prevents closing receive-only channels

## Advanced Patterns

### Fan-In (Multiple Producers, Single Consumer)

```go
func fanIn(ctx context.Context, channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for {
                select {
                case val, ok := <-c:
                    if !ok {
                        return
                    }
                    out <- val
                case <-ctx.Done():
                    return
                }
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

### Pipeline Pattern

```go
func generator(ctx context.Context, nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            select {
            case out <- n:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

func square(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            select {
            case out <- n * n:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

func main() {
    ctx := context.Background()

    // Pipeline: generate -> square
    nums := generator(ctx, 1, 2, 3, 4, 5)
    squared := square(ctx, nums)

    for val := range squared {
        fmt.Println(val) // 1, 4, 9, 16, 25
    }
}
```

### Or-Done Pattern

```go
func orDone(ctx context.Context, ch <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for {
            select {
            case <-ctx.Done():
                return
            case val, ok := <-ch:
                if !ok {
                    return
                }
                select {
                case out <- val:
                case <-ctx.Done():
                }
            }
        }
    }()
    return out
}

// Usage - simplifies downstream consumers
for val := range orDone(ctx, ch) {
    // No need to select on ctx.Done() here
    process(val)
}
```

## Common Mistakes

| Mistake                          | Fix                                               |
|----------------------------------|---------------------------------------------------|
| Send on closed channel           | Only sender should close; check before send       |
| Close receive-only channel       | Use bidirectional or send-only type               |
| Receive without checking `ok`    | Always use `val, ok := <-ch` or range             |
| Unbuffered send without receiver | Use buffered channel or ensure receiver ready     |
| Select without default blocks    | Add `default:` for non-blocking behavior          |
| Forgetting to close channel      | Producer must close when done (for range to exit) |
| Multiple goroutines closing      | Use `sync.Once` if multiple closers possible      |

## Closing Channels Safely

### Only Sender Closes

```go
// Good
func producer(out chan<- int) {
    defer close(out) // Producer closes
    for i := 0; i < 10; i++ {
        out <- i
    }
}
```

### Multiple Senders (Use Done Channel)

```go
func workers(ctx context.Context, tasks <-chan Task, results chan<- Result) {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for task := range tasks {
                results <- task.Process(ctx)
            }
        }()
    }

    go func() {
        wg.Wait()
        close(results) // Close only after all senders done
    }()
}
```

### Close Only Once (sync.Once)

```go
type SafeChannel struct {
    ch    chan int
    once  sync.Once
}

func (s *SafeChannel) Close() {
    s.once.Do(func() {
        close(s.ch)
    })
}
```

## Performance Considerations

### Buffer Size Selection

```go
// No buffer - synchronous, guaranteed delivery
ch := make(chan int)

// Small buffer - smooth out bursts
ch := make(chan int, 10)

// Large buffer - decouple producer/consumer rates
ch := make(chan int, 1000)

// Buffer = num producers - prevent blocking
ch := make(chan int, runtime.NumCPU())
```

**Guidelines:**
- Unbuffered: when synchronization required
- Buffer = expected burst size: when smoothing spikes
- Large buffer: when producer/consumer rates vary significantly
- Measure: profile before optimizing buffer size

## See Also

- [Goroutines](./goroutines.md) - Goroutine lifecycle management
- [Patterns](./patterns.md) - Advanced concurrency patterns
- [Sync Primitives](./sync-primitives.md) - Mutexes and WaitGroups
