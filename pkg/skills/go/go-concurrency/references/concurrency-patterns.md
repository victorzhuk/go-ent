# Concurrency Patterns Quick Reference

Extracted from `docs/go/topics/03-concurrency/patterns.md` (436 lines) → 120 lines.

## Worker Pool

```go
func WorkerPool(ctx context.Context, jobs []Job, workers int) []Result {
    jobCh := make(chan Job, workers)
    resultCh := make(chan Result, len(jobs))

    // Launch workers
    var wg sync.WaitGroup
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobCh {
                select {
                case resultCh <- processJob(job):
                case <-ctx.Done():
                    return
                }
            }
        }()
    }

    // Send jobs
    go func() {
        defer close(jobCh)
        for _, job := range jobs {
            select {
            case jobCh <- job:
            case <-ctx.Done():
                return
            }
        }
    }()

    // Collect results
    go func() {
        wg.Wait()
        close(resultCh)
    }()

    var results []Result
    for r := range resultCh {
        results = append(results, r)
    }
    return results
}
```

## Bounded Parallelism

```go
import "golang.org/x/sync/semaphore"

func ProcessWithLimit(ctx context.Context, items []Item, limit int64) error {
    sem := semaphore.NewWeighted(limit)

    var g errgroup.Group
    for _, item := range items {
        item := item
        if err := sem.Acquire(ctx, 1); err != nil {
            return err
        }

        g.Go(func() error {
            defer sem.Release(1)
            return process(ctx, item)
        })
    }

    return g.Wait()
}
```

## Rate Limiting

```go
import "golang.org/x/time/rate"

func ProcessWithRateLimit(ctx context.Context, items []Item, rps int) error {
    limiter := rate.NewLimiter(rate.Limit(rps), rps)

    for _, item := range items {
        if err := limiter.Wait(ctx); err != nil {
            return err
        }
        if err := process(ctx, item); err != nil {
            return err
        }
    }
    return nil
}
```

## Broadcast Pattern

```go
type Broadcaster struct {
    mu        sync.Mutex
    listeners []chan string
}

func (b *Broadcaster) Subscribe() <-chan string {
    b.mu.Lock()
    defer b.mu.Unlock()

    ch := make(chan string, 10)
    b.listeners = append(b.listeners, ch)
    return ch
}

func (b *Broadcaster) Broadcast(msg string) {
    b.mu.Lock()
    defer b.mu.Unlock()

    for _, ch := range b.listeners {
        select {
        case ch <- msg:
        default:
            // Drop if channel full
        }
    }
}
```

## Timeout Pattern

```go
func FetchWithTimeout(ctx context.Context, url string) ([]byte, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}
```

## Context Propagation

```go
func Handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Pass context to all operations
    data, err := fetchData(ctx)
    if err != nil {
        if ctx.Err() != nil {
            // Request was cancelled
            http.Error(w, "request cancelled", 499)
            return
        }
        http.Error(w, err.Error(), 500)
        return
    }

    json.NewEncoder(w).Encode(data)
}
```
