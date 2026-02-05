# Memory Optimization

Escape analysis, sync.Pool, and memory efficiency.

## Quick Reference

| Technique | Use Case | Impact |
|-----------|----------|--------|
| `sync.Pool` | Reuse temporary objects | Reduce GC pressure, 30-50% alloc reduction |
| Preallocate slices | Known capacity | Avoid grow/copy, predictable memory |
| Struct padding | Cache-aligned structs | 20-40% perf gain on hot paths |
| Escape analysis | Optimize allocations | Stack vs heap, reduce escapes |
| Arena allocator | Bulk allocations (experimental) | Fast alloc/free, region-based |
| Memory patterns | Buffer reuse, pointer chains | Lower GC overhead, better locality |

```bash
# Escape analysis
go build -gcflags="-m -m" 2>&1 | grep "escapes to heap"

# Benchmark memory
go test -bench=. -benchmem

# Check struct size and alignment
go build -gcflags="-S" | grep "type.*size"
```

## sync.Pool

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func process(data []byte) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)

    buf.Reset() // Important: reset state
    buf.Write(data)
    // Process...
}
```

## Reduce Allocations

```go
// Bad - allocates every call
func bad() []byte {
    return []byte("constant")
}

// Good - static allocation
var constant = []byte("constant")

func good() []byte {
    return constant
}

// Bad - string concatenation
s := "hello" + " " + "world"

// Good - strings.Builder
var sb strings.Builder
sb.WriteString("hello")
sb.WriteString(" ")
sb.WriteString("world")
s := sb.String()
```

## Struct Padding

Compiler adds padding for alignment. Reorder fields for cache efficiency.

```go
// Bad - 40 bytes (5 words + padding)
type bad struct {
    flag    bool   // 1 byte + 7 padding
    counter int64  // 8 bytes
    active  bool   // 1 byte + 7 padding
    id      int64  // 8 bytes
    ready   bool   // 1 byte + 7 padding
}

// Good - 24 bytes (3 words, no padding)
type good struct {
    counter int64  // 8 bytes
    id      int64  // 8 bytes
    flag    bool   // 1 byte
    active  bool   // 1 byte
    ready   bool   // 1 byte
    // 5 bytes padding at end
}

// Cache-aligned struct (64-byte cache line)
type cacheAligned struct {
    hot1 int64  // Frequently accessed together
    hot2 int64
    hot3 int64
    _    [40]byte // Padding to 64 bytes
    cold int64  // Separate cache line
}
```

## Arena Allocator

Experimental Go 1.20+. Bulk allocate/free. Use for temporary data structures.

```go
// go:build goexperiment.arenas

import "arena"

func processRequests(reqs []Request) {
    a := arena.NewArena()
    defer a.Free() // Free all at once

    results := arena.MakeSlice[Result](a, len(reqs), len(reqs))
    for i, req := range reqs {
        // Allocate in arena
        r := arena.New[Result](a)
        *r = process(req)
        results[i] = *r
    }
    // All arena memory freed on defer
}
```

**When NOT to use:**
- Long-lived objects (use heap)
- Shared across goroutines (arena not thread-safe)
- Objects with finalizers
- Production code (still experimental)

## Memory Patterns

### Preallocate Slices

```go
// Bad - grows 4 times (1→2→4→8→16)
func bad(n int) []int {
    var result []int
    for i := 0; i < n; i++ {
        result = append(result, i)
    }
    return result
}

// Good - single allocation
func good(n int) []int {
    result := make([]int, 0, n)
    for i := 0; i < n; i++ {
        result = append(result, i)
    }
    return result
}
```

### Reuse Buffers

```go
var responsePool = sync.Pool{
    New: func() interface{} {
        return &bytes.Buffer{}
    },
}

func handler(w http.ResponseWriter, r *http.Request) {
    buf := responsePool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        responsePool.Put(buf)
    }()

    // Build response in pooled buffer
    json.NewEncoder(buf).Encode(data)
    w.Write(buf.Bytes())
}
```

### Avoid Pointer Chains

```go
// Bad - 4 indirections, poor cache locality
type bad struct {
    user    *User
    account *Account
    billing *Billing
}

// Good - embedded, single allocation
type good struct {
    user    User
    account Account
    billing Billing
}
```

## Escape Analysis

Compiler decides stack vs heap. Stack is faster (no GC).

```go
// Escapes to heap - returned pointer
func escapes() *int {
    x := 42
    return &x // x escapes
}

// Stack allocation - not returned
func stack() int {
    x := 42
    return x // x on stack
}

// Check with:
// go build -gcflags="-m" 2>&1 | grep escape
```

**Reduce escapes:**

```go
// Bad - interface causes escape
func bad(v interface{}) {
    fmt.Println(v) // v escapes to heap
}

// Good - concrete type stays on stack
func good(v int) {
    fmt.Println(v) // v on stack
}

// Bad - large array escapes
func badArray() {
    var buf [64 * 1024]byte // Escapes (too large)
    process(buf[:])
}

// Good - use slice on heap explicitly
func goodArray() {
    buf := make([]byte, 64*1024) // Controlled heap alloc
    process(buf)
}
```

## Common Mistakes

| Mistake | Why Bad | Fix |
|---------|---------|-----|
| `buf := make([]byte, 10)` repeatedly | Allocates every call | Use `sync.Pool` or preallocate |
| Not calling `buf.Reset()` | Pool returns dirty buffers | Always reset before returning to pool |
| Premature optimization | Waste time, complex code | Profile first, optimize hot paths only |
| Ignoring struct padding | 30-50% memory waste | Reorder fields (large→small) |
| Arena in production | Experimental, bugs possible | Wait for stable release |
| Small `sync.Pool` objects | Pool overhead > savings | Pool only buffers >1KB typically |

## See Also

- [Profiling](./profiling.md) - Memory profiling with pprof
- [GC Tuning](./gc-tuning.md) - Garbage collector optimization
- [Benchmarks](./benchmarks.md) - Measuring allocations with -benchmem
- [Sync Primitives](../03-concurrency/sync-primitives.md) - sync.Pool patterns
