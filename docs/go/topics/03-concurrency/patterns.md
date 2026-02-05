# Concurrency Patterns

Advanced patterns for structuring concurrent Go programs using channels and goroutines.

## Quick Reference

| Pattern       | Use When                          |
|---------------|-----------------------------------|
| Pipeline      | Transform data through stages     |
| Fan-Out       | Parallelize CPU-intensive work    |
| Fan-In        | Merge multiple channels into one  |
| Worker Pool   | Limit concurrent processing       |
| Or-Channel    | Combine multiple done signals     |
| Tee-Channel   | Send values to multiple consumers |
| Rate Limiting | Control request rate              |

## Pipeline Pattern

### Basic Pipeline

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

func printer(ctx context.Context, in <-chan int) {
    for n := range in {
        select {
        case <-ctx.Done():
            return
        default:
            fmt.Println(n)
        }
    }
}

func main() {
    ctx := context.Background()

    // Build pipeline
    nums := generator(ctx, 1, 2, 3, 4)
    squared := square(ctx, nums)
    printer(ctx, squared)
}
```

**Key points:**
- Each stage receives from input channel, transforms, sends to output
- Each stage closes its output channel when done
- Context propagation enables graceful shutdown

## Fan-Out, Fan-In

### Fan-Out (Multiple Workers)

```go
func fanOut(ctx context.Context, input <-chan Task, numWorkers int) []<-chan Result {
    workers := make([]<-chan Result, numWorkers)

    for i := 0; i < numWorkers; i++ {
        workers[i] = worker(ctx, input)
    }

    return workers
}

func worker(ctx context.Context, input <-chan Task) <-chan Result {
    out := make(chan Result)
    go func() {
        defer close(out)
        for task := range input {
            select {
            case out <- task.Process():
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}
```

### Fan-In (Merge Results)

```go
func fanIn(ctx context.Context, channels ...<-chan Result) <-chan Result {
    var wg sync.WaitGroup
    out := make(chan Result)

    multiplex := func(c <-chan Result) {
        defer wg.Done()
        for result := range c {
            select {
            case out <- result:
            case <-ctx.Done():
                return
            }
        }
    }

    wg.Add(len(channels))
    for _, c := range channels {
        go multiplex(c)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}

// Usage
func main() {
    ctx := context.Background()
    input := generator(ctx, tasks)

    // Fan-out to 10 workers
    workers := fanOut(ctx, input, 10)

    // Fan-in results
    results := fanIn(ctx, workers...)

    for result := range results {
        fmt.Println(result)
    }
}
```

## Worker Pool

### Fixed Pool Size

```go
type WorkerPool struct {
    tasks   chan Task
    results chan Result
    workers int
    wg      sync.WaitGroup
}

func NewWorkerPool(numWorkers int) *WorkerPool {
    return &WorkerPool{
        tasks:   make(chan Task, numWorkers*2), // Buffer for smooth operation
        results: make(chan Result, numWorkers*2),
        workers: numWorkers,
    }
}

func (p *WorkerPool) Start(ctx context.Context) {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(ctx)
    }
}

func (p *WorkerPool) worker(ctx context.Context) {
    defer p.wg.Done()

    for {
        select {
        case task, ok := <-p.tasks:
            if !ok {
                return
            }
            result := task.Process(ctx)
            select {
            case p.results <- result:
            case <-ctx.Done():
                return
            }
        case <-ctx.Done():
            return
        }
    }
}

func (p *WorkerPool) Submit(task Task) {
    p.tasks <- task
}

func (p *WorkerPool) Results() <-chan Result {
    return p.results
}

func (p *WorkerPool) Shutdown() {
    close(p.tasks)
    p.wg.Wait()
    close(p.results)
}
```

### Dynamic Pool with errgroup

```go
func processWithPool(ctx context.Context, tasks []Task, poolSize int) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(poolSize) // Limit concurrent goroutines

    for _, task := range tasks {
        task := task
        g.Go(func() error {
            return task.Process(ctx)
        })
    }

    return g.Wait()
}
```

## Or-Channel Pattern

### Combine Multiple Done Signals

```go
func or(channels ...<-chan struct{}) <-chan struct{} {
    switch len(channels) {
    case 0:
        return nil
    case 1:
        return channels[0]
    }

    orDone := make(chan struct{})
    go func() {
        defer close(orDone)

        switch len(channels) {
        case 2:
            select {
            case <-channels[0]:
            case <-channels[1]:
            }
        default:
            select {
            case <-channels[0]:
            case <-channels[1]:
            case <-channels[2]:
            case <-or(append(channels[3:], orDone)...):
            }
        }
    }()
    return orDone
}

// Usage
func main() {
    sig := make(chan struct{})
    timeout := time.After(5 * time.Second)
    ctx := context.Background()

    // Returns when ANY signal fires
    <-or(sig, timeout, ctx.Done())
}
```

## Tee-Channel Pattern

### Split Stream to Multiple Consumers

```go
func tee(ctx context.Context, in <-chan int) (<-chan int, <-chan int) {
    out1 := make(chan int)
    out2 := make(chan int)

    go func() {
        defer close(out1)
        defer close(out2)

        for val := range in {
            var out1, out2 = out1, out2 // Shadow variables
            for i := 0; i < 2; i++ {
                select {
                case <-ctx.Done():
                    return
                case out1 <- val:
                    out1 = nil // Prevent sending to same channel twice
                case out2 <- val:
                    out2 = nil
                }
            }
        }
    }()

    return out1, out2
}
```

## Rate Limiting

### Token Bucket with time/rate

```go
import "golang.org/x/time/rate"

func processWithRateLimit(ctx context.Context, tasks []Task) error {
    limiter := rate.NewLimiter(rate.Limit(10), 1) // 10 requests/sec, burst 1

    for _, task := range tasks {
        if err := limiter.Wait(ctx); err != nil {
            return fmt.Errorf("rate limit wait: %w", err)
        }

        if err := task.Process(ctx); err != nil {
            return fmt.Errorf("process task: %w", err)
        }
    }

    return nil
}
```

### Semaphore Pattern

```go
type Semaphore struct {
    semaCh chan struct{}
}

func NewSemaphore(max int) *Semaphore {
    return &Semaphore{
        semaCh: make(chan struct{}, max),
    }
}

func (s *Semaphore) Acquire(ctx context.Context) error {
    select {
    case s.semaCh <- struct{}{}:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (s *Semaphore) Release() {
    <-s.semaCh
}

// Usage
func main() {
    sem := NewSemaphore(5) // Max 5 concurrent

    for _, task := range tasks {
        if err := sem.Acquire(ctx); err != nil {
            break
        }

        go func(t Task) {
            defer sem.Release()
            t.Process()
        }(task)
    }
}
```

### Leaky Bucket Pattern

```go
func leakyBucket(ctx context.Context, input <-chan Request, rate time.Duration) <-chan Request {
    out := make(chan Request)

    go func() {
        defer close(out)
        ticker := time.NewTicker(rate)
        defer ticker.Stop()

        for req := range input {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                out <- req
            }
        }
    }()

    return out
}
```

## Timeout and Cancellation

### Per-Request Timeout

```go
func processWithTimeout(ctx context.Context, task Task, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    done := make(chan error, 1)
    go func() {
        done <- task.Process(ctx)
    }()

    select {
    case err := <-done:
        return err
    case <-ctx.Done():
        return fmt.Errorf("timeout: %w", ctx.Err())
    }
}
```

### Heartbeat Pattern

```go
func doWorkWithHeartbeat(ctx context.Context, work func(context.Context) error) (<-chan struct{}, <-chan error) {
    heartbeat := make(chan struct{})
    errs := make(chan error, 1)

    go func() {
        defer close(heartbeat)
        defer close(errs)

        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                errs <- ctx.Err()
                return
            case <-ticker.C:
                heartbeat <- struct{}{}
            default:
                if err := work(ctx); err != nil {
                    errs <- err
                    return
                }
            }
        }
    }()

    return heartbeat, errs
}
```

## Bridge-Channel Pattern

### Flatten Channel of Channels

```go
func bridge(ctx context.Context, chanStream <-chan <-chan int) <-chan int {
    out := make(chan int)

    go func() {
        defer close(out)

        for {
            var stream <-chan int
            select {
            case maybeStream, ok := <-chanStream:
                if !ok {
                    return
                }
                stream = maybeStream
            case <-ctx.Done():
                return
            }

            for val := range stream {
                select {
                case out <- val:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()

    return out
}
```

## Common Mistakes

| Mistake                | Fix                                  |
|------------------------|--------------------------------------|
| Not closing channels   | Producer must close when done        |
| Unbounded goroutines   | Use worker pool or errgroup.SetLimit |
| No context propagation | Pass context through all stages      |
| Blocking on send       | Use select with ctx.Done()           |
| Ignoring backpressure  | Add buffer or limit producers        |
| Not draining channels  | Can cause goroutine leaks            |

## Performance Considerations

### When to Use Patterns

| Pattern       | Overhead | Use When                           |
|---------------|----------|------------------------------------|
| Pipeline      | Low      | Sequential transformations         |
| Fan-Out       | Medium   | CPU-intensive work, multiple cores |
| Worker Pool   | Low      | Limit concurrency, I/O-bound       |
| Or-Channel    | Low      | Multiple cancellation sources      |
| Rate Limiting | Low      | API rate limits, backpressure      |

### Buffering Strategy

```go
// No buffer - synchronous, backpressure
ch := make(chan T)

// Small buffer - smooth bursts
ch := make(chan T, runtime.NumCPU())

// Large buffer - decouple producer/consumer
ch := make(chan T, 1000)

// Unbounded (use with caution)
ch := make(chan T, len(items))
```

## See Also

- [Goroutines](./goroutines.md) - Goroutine basics
- [Channels](./channels.md) - Channel fundamentals
- [Sync Primitives](./sync-primitives.md) - Locks and synchronization
- [Go Concurrency Patterns (blog)](https://go.dev/blog/pipelines) - Official blog post
