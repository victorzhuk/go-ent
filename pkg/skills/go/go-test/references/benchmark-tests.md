# Benchmark Tests Quick Reference

Extracted from `docs/go/topics/08-testing/benchmarks.md` (451 lines) → 100 lines of actionable patterns.

## Quick Reference Table

| Operation                    | Example                        | Use Case                          |
|------------------------------|--------------------------------|-----------------------------------|
| b.N loop                     | `for i := 0; i < b.N; i++`     | Classic iteration (pre-1.24)      |
| b.Loop()                     | `for b.Loop() { ... }`         | Modern iteration (Go 1.24+)       |
| b.ResetTimer()               | After setup                    | Exclude initialization            |
| b.StopTimer()/StartTimer()   | Around excluded code           | Exclude cleanup/intermediate      |
| b.ReportAllocs()             | Report memory stats            | Track allocations per operation   |
| b.ReportMetric(n, unit)      | Custom metrics                 | Domain-specific measurements      |
| b.RunParallel(fn)            | Parallel execution             | Test concurrent scalability       |

## Basic Benchmark

```go
// Classic b.N loop (pre-1.24)
func BenchmarkCacheGet(b *testing.B) {
    cache := NewLRU(1000)
    cache.Put("key", "value")

    b.ResetTimer()  // Exclude setup time
    for i := 0; i < b.N; i++ {
        cache.Get("key")
    }
}

// Modern b.Loop() API (Go 1.24+) - recommended
func BenchmarkCacheGet(b *testing.B) {
    cache := NewLRU(1000)
    cache.Put("key", "value")

    b.ResetTimer()
    for b.Loop() {
        cache.Get("key")
    }
}

// Run: go test -bench=. -benchmem
```

## Prevent Dead Code Elimination

```go
var result string  // Package-level sink

func BenchmarkJSONMarshal(b *testing.B) {
    data := User{ID: "123", Email: "user@example.com"}
    var r []byte

    for b.Loop() {
        r, _ = json.Marshal(data)
    }
    result = string(r)  // Prevent compiler from eliminating Marshal call
}

// Without sink, compiler may optimize away the entire benchmark
```

## Memory Benchmarks

```go
func BenchmarkStringConcat(b *testing.B) {
    b.ReportAllocs()  // Report allocations per operation

    for b.Loop() {
        _ = "hello" + " " + "world"
    }
}

// Output:
// BenchmarkStringConcat-8    50000000    25.3 ns/op    48 B/op    3 allocs/op
//                                        ^^^^^^^^^^^   ^^^^^^^    ^^^^^^^^^^^^
//                                        time/op       bytes/op   allocs/op
```

## Parallel Benchmarks

```go
func BenchmarkParallelCache(b *testing.B) {
    cache := NewLRU(1000)
    cache.Put("key", "value")

    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            cache.Get("key")
        }
    })
}

// Tests concurrent access, scales with GOMAXPROCS
```

## Table-Driven Benchmarks

```go
func BenchmarkSizes(b *testing.B) {
    sizes := []int{10, 100, 1000, 10000}

    for _, size := range sizes {
        b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
            data := generateTestData(size)
            b.ResetTimer()

            for b.Loop() {
                Process(data)
            }
        })
    }
}

// Run: go test -bench=BenchmarkSizes
// Output:
// BenchmarkSizes/size-10-8       1000000    1234 ns/op
// BenchmarkSizes/size-100-8       100000   12345 ns/op
// BenchmarkSizes/size-1000-8       10000  123456 ns/op
```

## Benchmark Comparison

```bash
# Benchmark before optimization
go test -bench=. -benchmem > old.txt

# Make changes...

# Benchmark after optimization
go test -bench=. -benchmem > new.txt

# Compare results
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```
