# Memory Optimization Quick Reference

Extracted from `docs/go/topics/12-performance/memory.md` → 80 lines of actionable patterns.

## Quick Reference Table

| Technique         | Use Case                | Impact                    |
|-------------------|-------------------------|---------------------------|
| sync.Pool         | Reuse temporary objects | 30-50% alloc reduction    |
| Preallocate slices| Known capacity          | Avoid grow/copy           |
| Struct padding    | Cache-aligned structs   | 20-40% perf gain          |
| Escape analysis   | Optimize allocations    | Stack vs heap             |

## sync.Pool Pattern

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func process(data []byte) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)

    buf.Reset()  // Critical: reset state before reuse
    buf.Write(data)
}
```

## Reduce Allocations

```go
// Bad - allocates every call
func bad() []byte { return []byte("constant") }

// Good - static allocation
var constant = []byte("constant")
func good() []byte { return constant }

// Bad - string concatenation
s := "hello" + " " + "world"

// Good - strings.Builder with capacity hint
var sb strings.Builder
sb.Grow(11)  // preallocate
sb.WriteString("hello")
sb.WriteString(" ")
sb.WriteString("world")
```

## Struct Padding Optimization

```go
// Bad - 40 bytes (padding between fields)
type bad struct {
    flag    bool   // 1 byte + 7 padding
    counter int64  // 8 bytes
    active  bool   // 1 byte + 7 padding
    id      int64  // 8 bytes
}

// Good - 24 bytes (group by size)
type good struct {
    counter int64  // 8 bytes
    id      int64  // 8 bytes
    flag    bool   // 1 byte
    active  bool   // 1 byte
    // 6 bytes padding at end
}

// Cache-aligned (64-byte cache line)
type cacheAligned struct {
    // Hot fields (frequently accessed together)
    hot1 int64
    hot2 int64
    _ [48]byte  // padding to 64 bytes
    // Cold fields (separate cache line)
    cold1 int64
}
```

## Escape Analysis

```bash
# Check what escapes to heap
go build -gcflags="-m -m" 2>&1 | grep "escapes to heap"

# Benchmark memory
go test -bench=. -benchmem
```

```go
// Stays on stack (good)
func stack() int {
    x := 42
    return x
}

// Escapes to heap (bad if avoidable)
func heap() *int {
    x := 42
    return &x  // pointer escapes
}

// Stays on stack (good - slice doesn't escape)
func noEscape(data []int) int {
    return data[0]
}
```

## Preallocate Slices

```go
// Bad - multiple allocations as slice grows
items := []Item{}
for i := 0; i < 1000; i++ {
    items = append(items, Item{})
}

// Good - single allocation
items := make([]Item, 0, 1000)
for i := 0; i < 1000; i++ {
    items = append(items, Item{})
}
```
