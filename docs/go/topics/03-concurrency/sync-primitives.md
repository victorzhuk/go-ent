# Sync Primitives

The `sync` package provides low-level synchronization primitives for protecting shared state and coordinating goroutines.

## Quick Reference

| Primitive        | Use When                                            |
|------------------|-----------------------------------------------------|
| `sync.Mutex`     | Protect shared state with exclusive access          |
| `sync.RWMutex`   | Many readers, few writers                           |
| `sync.WaitGroup` | Wait for multiple goroutines to complete            |
| `sync.Once`      | Initialize exactly once (lazy init, singleton)      |
| `sync.Pool`      | Reuse temporary objects (reduce GC pressure)        |
| `sync.Cond`      | Coordinate goroutines waiting for conditions (rare) |
| `sync.Map`       | Concurrent map (specific use cases only)            |

## Mutex

### Basic Mutex

```go
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock() // Always defer unlock
    c.value++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}
```

**Key points:**
- Always `defer unlock()` to handle panics
- Keep critical sections small
- Zero value is usable (no initialization needed)

### Mutex with Error Handling

```go
func (s *Service) Process(id string) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    item, exists := s.items[id]
    if !exists {
        return ErrNotFound // Unlock happens via defer
    }

    return item.Process()
}
```

## RWMutex

### Read-Heavy Workloads

```go
type Cache struct {
    mu    sync.RWMutex
    items map[string]Item
}

func (c *Cache) Get(key string) (Item, bool) {
    c.mu.RLock() // Multiple readers can hold RLock simultaneously
    defer c.mu.RUnlock()

    item, ok := c.items[key]
    return item, ok
}

func (c *Cache) Set(key string, item Item) {
    c.mu.Lock() // Exclusive lock (blocks readers and writers)
    defer c.mu.Unlock()

    c.items[key] = item
}
```

**When to use:**
- Read operations >> write operations (10:1 or higher ratio)
- Critical section is large enough that lock overhead matters
- Measure before using (RWMutex has more overhead than Mutex)

### Upgrading Read Lock to Write Lock (Anti-Pattern)

```go
// Bad - cannot upgrade read lock to write lock
func (c *Cache) bad(key string) {
    c.mu.RLock()
    _, exists := c.items[key]
    if !exists {
        c.mu.RUnlock()
        c.mu.Lock() // Deadlock risk if another goroutine acquired lock between unlock and lock
        defer c.mu.Unlock()
        c.items[key] = defaultItem
        return
    }
    c.mu.RUnlock()
}

// Good - decide lock type upfront
func (c *Cache) good(key string) {
    c.mu.Lock() // Start with write lock
    defer c.mu.Unlock()

    if _, exists := c.items[key]; !exists {
        c.items[key] = defaultItem
    }
}
```

## WaitGroup

### Basic Pattern

```go
func processBatch(items []Item) {
    var wg sync.WaitGroup

    for _, item := range items {
        wg.Add(1)
        go func(it Item) {
            defer wg.Done()
            it.Process()
        }(item)
    }

    wg.Wait() // Block until all goroutines call Done()
}
```

### With Error Handling (use errgroup instead)

```go
import "golang.org/x/sync/errgroup"

func processBatch(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)

    for _, item := range items {
        item := item
        g.Go(func() error {
            return item.Process(ctx)
        })
    }

    return g.Wait() // Returns first error
}
```

## sync.Once

### Lazy Initialization

```go
type Service struct {
    once   sync.Once
    client *http.Client
}

func (s *Service) getClient() *http.Client {
    s.once.Do(func() {
        s.client = &http.Client{
            Timeout: 10 * time.Second,
        }
    })
    return s.client
}
```

### Singleton Pattern

```go
var (
    instance *Database
    once     sync.Once
)

func GetDatabase() *Database {
    once.Do(func() {
        instance = &Database{
            conn: connectToDB(),
        }
    })
    return instance
}
```

**Key points:**
- `Do()` is called only once, even if multiple goroutines call it
- If `Do()` panics, `once.Do()` is considered complete (won't retry)
- Safe for concurrent access

## sync.Pool

### Object Reuse

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer) // Called when pool is empty
    },
}

func processData(data []byte) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)

    buf.Reset() // IMPORTANT: Reset state before use
    buf.Write(data)
    // Process buffer
}
```

**Key points:**
- Pool is safe for concurrent access
- Items may be removed from pool at any time (GC)
- Always reset object state before returning to pool
- Use for temporary objects created frequently (e.g., buffers)

### When to Use sync.Pool

```go
// Good - high-frequency temporary allocations
var encoderPool = sync.Pool{
    New: func() interface{} {
        return json.NewEncoder(new(bytes.Buffer))
    },
}

// Bad - pool overhead exceeds benefit
var intPool = sync.Pool{
    New: func() interface{} {
        return new(int) // int allocation is cheap, pool overhead higher
    },
}
```

**Use when:**
- Object allocation is expensive
- Object is created and destroyed frequently
- Object has significant size (>1KB)
- Profile shows allocation pressure

## sync.Map

### Specialized Concurrent Map

```go
var cache sync.Map

// Store
cache.Store("key", "value")

// Load
value, ok := cache.Load("key")

// LoadOrStore (atomic)
actual, loaded := cache.LoadOrStore("key", "value")

// Delete
cache.Delete("key")

// Range
cache.Range(func(key, value interface{}) bool {
    fmt.Println(key, value)
    return true // continue iteration
})
```

### When to Use sync.Map

**Use `sync.Map` when:**
1. Keys are written once and read many times
2. Multiple goroutines read, write, and overwrite different keys

**Use `map` + `sync.RWMutex` when:**
- General-purpose concurrent map
- Performance matters (measure first)

```go
// sync.Map is optimized for these patterns
var cache sync.Map

// Write once per key
cache.Store(key, value)

// Read many times
for i := 0; i < 1000; i++ {
    value, _ := cache.Load(key)
}

// Different goroutines work on different keys
go func() { cache.Store("key1", val1) }()
go func() { cache.Store("key2", val2) }()
```

## sync.Cond

### Condition Variables (Rare Use Case)

```go
type Queue struct {
    mu    sync.Mutex
    cond  *sync.Cond
    items []Item
}

func NewQueue() *Queue {
    q := &Queue{items: make([]Item, 0)}
    q.cond = sync.NewCond(&q.mu)
    return q
}

func (q *Queue) Enqueue(item Item) {
    q.mu.Lock()
    defer q.mu.Unlock()

    q.items = append(q.items, item)
    q.cond.Signal() // Wake one waiting goroutine
}

func (q *Queue) Dequeue() Item {
    q.mu.Lock()
    defer q.mu.Unlock()

    for len(q.items) == 0 { // Use loop, not if (spurious wakeups)
        q.cond.Wait() // Releases lock and blocks
    }

    item := q.items[0]
    q.items = q.items[1:]
    return item
}
```

**Key points:**
- Rarely needed in Go (channels usually better)
- `Wait()` must be called with lock held
- Always use loop, not `if` (handles spurious wakeups)

## Common Mistakes

| Mistake                         | Fix                                   |
|---------------------------------|---------------------------------------|
| Not deferring `Unlock()`        | Always `defer mu.Unlock()`            |
| Copying `sync.Mutex`            | Pass by pointer, never copy           |
| Locking in wrong goroutine      | Lock/unlock must be in same goroutine |
| Holding lock during I/O         | Keep critical section small           |
| Using RWMutex for write-heavy   | Profile first; Mutex may be faster    |
| Not resetting Pool objects      | Call `Reset()` before `Put()`         |
| Using sync.Map for general case | Measure; map+RWMutex often faster     |

## Performance Patterns

### Minimize Critical Section

```go
// Bad - holding lock during I/O
func (s *Service) badProcess(id string) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    data := s.cache[id]
    result := expensiveComputation(data) // Lock held too long
    return s.db.Save(result)             // I/O with lock held!
}

// Good - minimal critical section
func (s *Service) goodProcess(id string) error {
    s.mu.Lock()
    data := s.cache[id]
    s.mu.Unlock() // Release lock quickly

    result := expensiveComputation(data) // No lock needed
    return s.db.Save(result)             // No lock needed
}
```

### Reduce Lock Contention

```go
// Bad - single lock for entire cache
type BadCache struct {
    mu    sync.RWMutex
    items map[string]Item
}

// Good - sharded locks
type GoodCache struct {
    shards [256]struct {
        mu    sync.RWMutex
        items map[string]Item
    }
}

func (c *GoodCache) getShard(key string) *struct {
    mu    sync.RWMutex
    items map[string]Item
} {
    h := fnv.New32a()
    h.Write([]byte(key))
    idx := h.Sum32() % 256
    return &c.shards[idx]
}

func (c *GoodCache) Get(key string) (Item, bool) {
    shard := c.getShard(key)
    shard.mu.RLock()
    defer shard.mu.RUnlock()

    item, ok := shard.items[key]
    return item, ok
}
```

## See Also

- [Goroutines](./goroutines.md) - Goroutine lifecycle
- [Channels](./channels.md) - Alternative to shared memory
- [Patterns](./patterns.md) - Higher-level concurrency patterns
- [sync package](https://pkg.go.dev/sync) - Official documentation
