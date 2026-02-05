# Benchmarks

Go's benchmarking framework (`testing.B`) provides standardized performance measurement, statistical analysis, and profiling integration for identifying bottlenecks and validating optimizations.

## Quick Reference

| Operation | Example | Use Case |
|-----------|---------|----------|
| **b.N loop** | `for i := 0; i < b.N; i++` | Classic iteration pattern (pre-1.24) |
| **b.Loop()** | `for b.Loop() { ... }` | Modern iteration API (Go 1.24+), cleaner |
| **b.ResetTimer()** | Reset after setup | Exclude initialization from measurement |
| **b.StopTimer()/StartTimer()** | Pause/resume | Exclude cleanup or intermediate setup |
| **b.ReportAllocs()** | Report memory stats | Track allocations per operation |
| **b.ReportMetric(n, unit)** | Custom metrics | Domain-specific measurements |
| **b.RunParallel(fn)** | Parallel execution | Test concurrent scalability |
| **b.SetParallelism(p)** | Set GOMAXPROCS multiplier | Control parallel benchmark concurrency |

## Basic Benchmarks

Benchmarks measure operation performance over `b.N` iterations. The framework automatically adjusts `N` until results stabilize (typically 1-10 seconds runtime).

**File naming:** `*_test.go`, function `BenchmarkXxx(b *testing.B)`

**Run:** `go test -bench=. -benchmem`

### Classic b.N Loop (Pre-1.24)

```go
// internal/cache/lru_test.go
func BenchmarkCacheGet(b *testing.B) {
	cache := NewLRU(1000)
	cache.Put("key", "value")

	b.ResetTimer() // Exclude setup time
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}
```

### Modern b.Loop() API (Go 1.24+)

```go
func BenchmarkCacheGet(b *testing.B) {
	cache := NewLRU(1000)
	cache.Put("key", "value")

	b.ResetTimer()
	for b.Loop() {
		cache.Get("key")
	}
}
```

**Why b.Loop():**
- Cleaner syntax (no index variable needed)
- Prevents accidental misuse of loop index
- Compiler optimizations easier to apply
- Recommended for all new benchmarks

### Preventing Dead Code Elimination

```go
var result string // Package-level sink

func BenchmarkJSONMarshal(b *testing.B) {
	data := User{ID: "123", Email: "user@example.com"}
	var r []byte

	for b.Loop() {
		r, _ = json.Marshal(data)
	}
	result = string(r) // Prevent compiler from eliminating Marshal call
}
```

Without the sink, the compiler may optimize away the entire benchmark body.

## Memory Benchmarks

Track allocations to identify unnecessary heap escapes, boxing, or redundant allocations.

### Allocation Reporting

```go
func BenchmarkStringConcat(b *testing.B) {
	b.ReportAllocs()

	strs := []string{"hello", "world", "foo", "bar"}
	var s string

	for b.Loop() {
		s = strings.Join(strs, " ")
	}
	_ = s
}

func BenchmarkStringBuilder(b *testing.B) {
	b.ReportAllocs()

	strs := []string{"hello", "world", "foo", "bar"}
	var result string

	for b.Loop() {
		var b strings.Builder
		for i, s := range strs {
			if i > 0 {
				b.WriteString(" ")
			}
			b.WriteString(s)
		}
		result = b.String()
	}
	_ = result
}
```

**Output:**
```
BenchmarkStringConcat-8     5000000    240 ns/op    64 B/op    1 allocs/op
BenchmarkStringBuilder-8    8000000    180 ns/op    64 B/op    1 allocs/op
```

### Custom Metrics

```go
func BenchmarkBatchInsert(b *testing.B) {
	db := setupDB(b)
	defer db.Close()

	b.ResetTimer()
	var rowsInserted int64

	for b.Loop() {
		result, _ := db.Exec("INSERT INTO users (email) VALUES ($1)", "user@example.com")
		n, _ := result.RowsAffected()
		rowsInserted += n
	}

	b.ReportMetric(float64(rowsInserted)/float64(b.N), "rows/op")
}
```

**Output:**
```
BenchmarkBatchInsert-8    10000    120000 ns/op    1.00 rows/op
```

## Parallel Benchmarks

Measure concurrent performance and scalability under contention.

### Basic Parallel Benchmark

```go
func BenchmarkCacheGetParallel(b *testing.B) {
	cache := NewLRU(10000)
	for i := 0; i < 1000; i++ {
		cache.Put(fmt.Sprintf("key-%d", i), i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get("key-42")
		}
	})
}
```

**b.RunParallel:**
- Spawns `GOMAXPROCS` goroutines by default
- `pb.Next()` distributes `b.N` iterations across goroutines
- Reports ops/sec across all goroutines

### Controlling Parallelism

```go
func BenchmarkCacheMixed(b *testing.B) {
	cache := NewLRU(10000)

	b.SetParallelism(4) // 4 × GOMAXPROCS goroutines
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%10 == 0 {
				cache.Put(fmt.Sprintf("key-%d", i), i)
			} else {
				cache.Get("key-42")
			}
			i++
		}
	})
}
```

### Parallel with b.Loop() (Go 1.24+)

```go
func BenchmarkCacheParallel(b *testing.B) {
	cache := NewLRU(10000)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() { // Use pb.Next() in parallel, not b.Loop()
			cache.Get("key-42")
		}
	})
}
```

**Note:** Use `pb.Next()` inside `RunParallel`, not `b.Loop()`. The `b.Loop()` API is for sequential benchmarks only.

## Table-Driven Benchmarks

Benchmark multiple scenarios with statistical isolation via sub-benchmarks.

### Input Size Scaling

```go
func BenchmarkJSONEncode(b *testing.B) {
	sizes := []struct {
		name  string
		users int
	}{
		{"Small", 10},
		{"Medium", 100},
		{"Large", 1000},
		{"XLarge", 10000},
	}

	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			users := make([]User, sz.users)
			for i := range users {
				users[i] = User{
					ID:    uuid.Must(uuid.NewV7()).String(),
					Email: fmt.Sprintf("user%d@example.com", i),
				}
			}

			b.ResetTimer()
			b.ReportAllocs()

			var result []byte
			for b.Loop() {
				result, _ = json.Marshal(users)
			}
			_ = result
		})
	}
}
```

**Run specific sub-benchmark:**
```bash
go test -bench=JSONEncode/Large -benchmem
```

### Algorithm Comparison

```go
func BenchmarkSortAlgorithms(b *testing.B) {
	algos := []struct {
		name string
		sort func([]int)
	}{
		{"QuickSort", quickSort},
		{"MergeSort", mergeSort},
		{"HeapSort", heapSort},
	}

	data := make([]int, 10000)
	for i := range data {
		data[i] = rand.Intn(100000)
	}

	for _, algo := range algos {
		b.Run(algo.name, func(b *testing.B) {
			for b.Loop() {
				b.StopTimer()
				input := make([]int, len(data))
				copy(input, data)
				b.StartTimer()

				algo.sort(input)
			}
		})
	}
}
```

**Pattern:** Use `b.StopTimer()/StartTimer()` to exclude setup from measurement, not `b.ResetTimer()` inside loop.

## b.Loop() API

Go 1.24 introduces `b.Loop()` as the preferred iteration API, replacing the traditional `for i := 0; i < b.N; i++` pattern.

### Migration from b.N

```go
// Before (Go 1.23 and earlier)
func BenchmarkOld(b *testing.B) {
	cache := setupCache()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}

// After (Go 1.24+)
func BenchmarkNew(b *testing.B) {
	cache := setupCache()
	b.ResetTimer()

	for b.Loop() {
		cache.Get("key")
	}
}
```

### Benefits

**Cleaner syntax:**
```go
// No unused index variable
for b.Loop() {          // Clean
	doWork()
}

for i := 0; i < b.N; i++ {  // i often unused, linter warnings
	doWork()
}
```

**Prevents misuse:**
```go
// BAD: Using loop index incorrectly
for i := 0; i < b.N; i++ {
	cache.Put(fmt.Sprintf("key-%d", i), i) // Creates N unique keys!
}

// GOOD: b.Loop() makes it obvious this is wrong
for b.Loop() {
	cache.Put("key", "value") // Same key each iteration (correct)
}
```

**Compiler optimizations:**
- Simpler control flow for optimizer analysis
- Reduced register pressure (no index variable)
- Better inlining decisions

### When to Use b.N Directly

Some benchmarks legitimately need the index:

```go
func BenchmarkCyclicAccess(b *testing.B) {
	data := make([]int, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = data[i%len(data)] // Need index for cyclic access
	}
}
```

**Rule:** Use `b.Loop()` unless you genuinely need the iteration counter.

## Profiling Integration

Benchmarks integrate with `pprof` for detailed CPU/memory analysis.

### CPU Profiling

```bash
go test -bench=BenchmarkCacheGet -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

**Interactive commands:**
```
(pprof) top10        # Top 10 CPU consumers
(pprof) list Get     # Source code with timings
(pprof) web          # Visual call graph (requires Graphviz)
```

### Memory Profiling

```bash
go test -bench=BenchmarkJSONEncode -memprofile=mem.prof -benchmem
go tool pprof mem.prof
```

**Analyze allocations:**
```
(pprof) top -alloc_space    # Total allocated bytes
(pprof) top -alloc_objects  # Total allocated objects
(pprof) list Marshal
```

### Mutex Profiling

```bash
go test -bench=BenchmarkCacheMixed -mutexprofile=mutex.prof
go tool pprof mutex.prof
```

**Identify contention:**
```
(pprof) top -cum         # Cumulative contention time
(pprof) list Lock
```

### Combined Analysis

```go
func BenchmarkComplexOperation(b *testing.B) {
	b.ReportAllocs()

	svc := setupService(b)
	defer svc.Close()

	b.ResetTimer()
	for b.Loop() {
		ctx := context.Background()
		_, _ = svc.ProcessRequest(ctx, &Request{
			UserID: "test-user",
			Data:   []byte("payload"),
		})
	}
}
```

**Run with all profiles:**
```bash
go test -bench=Complex \
  -cpuprofile=cpu.prof \
  -memprofile=mem.prof \
  -mutexprofile=mutex.prof \
  -benchmem
```

## Common Mistakes

| Mistake | Problem | Solution |
|---------|---------|----------|
| **Timer not reset** | Setup time included in measurements | Always `b.ResetTimer()` after initialization |
| **Allocations in loop** | Benchmark measures setup, not operation | Move allocations outside loop, use `b.StopTimer()` if unavoidable |
| **Insufficient iterations** | High variance, unreliable results | Let framework run longer: `go test -bench=. -benchtime=10s` |
| **Comparing across machines** | Hardware differences invalidate comparison | Use `benchstat` tool, compare relative performance on same machine |
| **Ignoring compiler optimizations** | Dead code elimination skews results | Store results in package-level variable sink |
| **No memory reporting** | Missing allocation regressions | Always use `-benchmem` or `b.ReportAllocs()` |
| **Mixing b.N and iterations** | Double iteration or incorrect timing | Use `b.Loop()` (1.24+) OR `for i := 0; i < b.N`, never both |
| **Using b.Loop() in parallel** | Runtime panic or incorrect results | Use `pb.Next()` inside `b.RunParallel()` |

### Example: Timer Not Reset

```go
// BAD: Setup time included
func BenchmarkBad(b *testing.B) {
	cache := NewLRU(10000) // Expensive setup
	for i := 0; i < 1000; i++ {
		cache.Put(fmt.Sprintf("key-%d", i), i)
	}

	for b.Loop() {
		cache.Get("key-42")
	}
}

// GOOD: Setup excluded
func BenchmarkGood(b *testing.B) {
	cache := NewLRU(10000)
	for i := 0; i < 1000; i++ {
		cache.Put(fmt.Sprintf("key-%d", i), i)
	}

	b.ResetTimer() // Critical
	for b.Loop() {
		cache.Get("key-42")
	}
}
```

## See Also

- [profiling.md](../12-performance/profiling.md) - CPU/memory/mutex profiling with pprof
- [memory.md](../12-performance/memory.md) - Memory optimization and allocation analysis
- [table-driven.md](table-driven.md) - Table-driven test patterns
- [fuzzing.md](fuzzing.md) - Fuzz testing for edge cases
