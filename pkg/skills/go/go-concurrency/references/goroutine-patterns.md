# Goroutine Patterns Quick Reference

Extracted from `docs/go/topics/03-concurrency/goroutines.md` (398 lines) → 80 lines.

## Goroutine Lifecycle

```go
// Good - with context cancellation
func worker(ctx context.Context) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            doWork()
        case <-ctx.Done():
            return  // Clean exit
        }
    }
}

// Bad - goroutine leak (no exit condition)
func worker() {
    ticker := time.NewTicker(time.Second)
    for range ticker.C {
        doWork()  // Runs forever
    }
}
```

## WaitGroup Pattern

```go
func ProcessConcurrently(items []Item) {
    var wg sync.WaitGroup

    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()
            process(item)
        }(item)
    }

    wg.Wait()
}
```

## errgroup Pattern

```go
import "golang.org/x/sync/errgroup"

func ProcessWithErrors(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)

    for _, item := range items {
        item := item  // Capture for closure (Go <1.22)
        g.Go(func() error {
            return process(ctx, item)
        })
    }

    return g.Wait()  // Returns first error
}

// With limit
func ProcessWithLimit(ctx context.Context, items []Item, limit int) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(limit)  // Max concurrent goroutines

    for _, item := range items {
        item := item
        g.Go(func() error {
            return process(ctx, item)
        })
    }

    return g.Wait()
}
```

## Goroutine Pool

```go
func WorkerPool(ctx context.Context, jobs <-chan Job, results chan<- Result, workers int) {
    var wg sync.WaitGroup

    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                select {
                case results <- processJob(job):
                case <-ctx.Done():
                    return
                }
            }
        }()
    }

    wg.Wait()
    close(results)
}
```
