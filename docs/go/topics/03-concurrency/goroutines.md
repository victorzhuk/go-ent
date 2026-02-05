# Goroutines

Goroutines are lightweight threads managed by the Go runtime. Proper lifecycle management prevents resource leaks and ensures graceful shutdown.

## Quick Reference

| Pattern           | Use When                                |
|-------------------|-----------------------------------------|
| `go func()`       | Fire-and-forget background work         |
| `errgroup.Group`  | Multiple goroutines with error handling |
| `context.Context` | Need cancellation or timeout            |
| `sync.WaitGroup`  | Need to wait for completion (no errors) |

## Lifecycle Management

### Basic Pattern with WaitGroup

```go
var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        // Work here
    }(i) // Pass i as argument to avoid closure capture bug
}

wg.Wait() // Block until all goroutines complete
```

**Key points:**
- Always `Add(1)` before launching goroutine (not inside it)
- Always `defer Done()` at start of goroutine
- Pass loop variables as arguments to avoid closure capture issues

### Error Handling with errgroup

```go
import "golang.org/x/sync/errgroup"

func processItems(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)

    for _, item := range items {
        item := item // Capture loop variable (not needed in Go 1.22+)
        g.Go(func() error {
            return processItem(ctx, item)
        })
    }

    return g.Wait() // Returns first non-nil error
}
```

**Key points:**
- First error cancels context, stopping other goroutines
- `g.Wait()` returns first error encountered
- Derived context automatically cancelled on error
- Go 1.22+ automatically captures loop variables correctly

### Limited Concurrency with errgroup

```go
func processWithLimit(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(10) // Max 10 concurrent goroutines

    for _, item := range items {
        item := item
        g.Go(func() error {
            return processItem(ctx, item)
        })
    }

    return g.Wait()
}
```

## Leak Prevention

### Always Provide Exit Path

**Bad - goroutine leaks if channel never receives:**
```go
func bad() {
    ch := make(chan int)
    go func() {
        val := <-ch // Blocks forever if nothing sends
        process(val)
    }()
    // If we return here, goroutine leaks
}
```

**Good - use context for cancellation:**
```go
func good(ctx context.Context) {
    ch := make(chan int)
    go func() {
        select {
        case val := <-ch:
            process(val)
        case <-ctx.Done():
            return // Exit when context cancelled
        }
    }()
}
```

### Worker Pattern with Graceful Shutdown

```go
type Worker struct {
    tasks  chan Task
    done   chan struct{}
    wg     sync.WaitGroup
}

func NewWorker(numWorkers int) *Worker {
    w := &Worker{
        tasks: make(chan Task, 100),
        done:  make(chan struct{}),
    }

    for i := 0; i < numWorkers; i++ {
        w.wg.Add(1)
        go w.worker()
    }

    return w
}

func (w *Worker) worker() {
    defer w.wg.Done()

    for {
        select {
        case task := <-w.tasks:
            task.Execute()
        case <-w.done:
            return
        }
    }
}

func (w *Worker) Stop() {
    close(w.done)   // Signal all workers to stop
    w.wg.Wait()     // Wait for all to finish current task
}
```

## Context-Aware Goroutines

### Respecting Context Cancellation

```go
func worker(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err() // Return context error (Cancelled or DeadlineExceeded)

        default:
            // Do work
            if err := doWork(ctx); err != nil {
                return fmt.Errorf("do work: %w", err)
            }
        }
    }
}
```

### Long-Running Operation with Context

```go
func processItems(ctx context.Context, items []Item) error {
    for _, item := range items {
        // Check context before processing each item
        if err := ctx.Err(); err != nil {
            return err
        }

        if err := processItem(ctx, item); err != nil {
            return fmt.Errorf("process item %v: %w", item.ID, err)
        }
    }
    return nil
}
```

## Common Patterns

### Fan-Out Pattern

```go
func fanOut(ctx context.Context, input <-chan Task) (<-chan Result, <-chan error) {
    results := make(chan Result, 100)
    errs := make(chan error, 1)

    var wg sync.WaitGroup
    for i := 0; i < 10; i++ { // 10 workers
        wg.Add(1)
        go func() {
            defer wg.Done()
            for task := range input {
                select {
                case <-ctx.Done():
                    return
                default:
                    result, err := task.Process(ctx)
                    if err != nil {
                        select {
                        case errs <- err:
                        default: // Don't block if error already sent
                        }
                        return
                    }
                    results <- result
                }
            }
        }()
    }

    go func() {
        wg.Wait()
        close(results)
        close(errs)
    }()

    return results, errs
}
```

### Timeout Pattern

```go
func doWithTimeout(ctx context.Context, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    done := make(chan error, 1)
    go func() {
        done <- expensiveOperation(ctx)
    }()

    select {
    case err := <-done:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

## Common Mistakes

| Mistake                         | Fix                                     |
|---------------------------------|-----------------------------------------|
| `wg.Add(1)` inside goroutine    | Call `Add` before launching goroutine   |
| Not calling `defer wg.Done()`   | Always defer it at start of goroutine   |
| Closure capturing loop variable | Pass as argument or use Go 1.22+        |
| No exit condition               | Always provide cancellation via context |
| Blocking forever on channel     | Use `select` with `ctx.Done()` case     |
| Ignoring `ctx.Err()`            | Check and return context errors         |

## Performance Considerations

### Goroutine Stack Size

- Initial stack: ~2KB
- Grows/shrinks dynamically
- Can create millions of goroutines
- Prefer goroutine pools for high-frequency short tasks

### When Not to Use Goroutines

```go
// Bad - overhead exceeds benefit
for i := 0; i < 100; i++ {
    go func(x int) {
        result := x * 2
        fmt.Println(result)
    }(i)
}

// Good - simple loop is faster
for i := 0; i < 100; i++ {
    result := i * 2
    fmt.Println(result)
}
```

**Use goroutines when:**
- I/O-bound operations (network, disk)
- Independent tasks that can run concurrently
- Total work > goroutine creation overhead (~1-3µs)

## See Also

- [Channels](./channels.md) - Communication between goroutines
- [Sync Primitives](./sync-primitives.md) - Synchronization tools
- [Patterns](./patterns.md) - Advanced concurrency patterns
- [Context](https://pkg.go.dev/context) - Cancellation and deadlines
