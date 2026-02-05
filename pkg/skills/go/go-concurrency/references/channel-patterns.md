# Channel Patterns Quick Reference

Extracted from `docs/go/topics/03-concurrency/channels.md` (425 lines) → 100 lines.

## Buffered vs Unbuffered

```go
// Unbuffered - synchronous, blocks until receiver ready
ch := make(chan int)

// Buffered - asynchronous up to capacity
ch := make(chan int, 100)
```

## Close Semantics

```go
// Good - close by sender
func producer(out chan<- int) {
    defer close(out)  // Always close when done
    for i := 0; i < 10; i++ {
        out <- i
    }
}

// Good - receive until closed
func consumer(in <-chan int) {
    for v := range in {  // Exits when channel closed
        process(v)
    }
}
```

## Select Patterns

```go
// Timeout
select {
case result := <-ch:
    return result
case <-time.After(5 * time.Second):
    return ErrTimeout
}

// Non-blocking send
select {
case ch <- value:
    // Sent
default:
    // Would block, handle it
}

// Non-blocking receive
select {
case value := <-ch:
    // Received
default:
    // No data available
}

// Context cancellation
select {
case result := <-ch:
    return result, nil
case <-ctx.Done():
    return nil, ctx.Err()
}
```

## Fan-Out Pattern

```go
func FanOut(in <-chan int, n int) []<-chan int {
    outs := make([]<-chan int, n)
    for i := 0; i < n; i++ {
        out := make(chan int)
        outs[i] = out

        go func(out chan int) {
            defer close(out)
            for v := range in {
                out <- process(v)
            }
        }(out)
    }
    return outs
}
```

## Fan-In Pattern

```go
func FanIn(channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(ch <-chan int) {
            defer wg.Done()
            for v := range ch {
                out <- v
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

## Pipeline Pattern

```go
func Pipeline(ctx context.Context) {
    // Generator
    gen := func(nums ...int) <-chan int {
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

    // Stage
    square := func(in <-chan int) <-chan int {
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

    // Usage
    numbers := gen(1, 2, 3, 4)
    squared := square(numbers)
    for result := range squared {
        fmt.Println(result)
    }
}
```
