# Sync Primitives Quick Reference

Extracted from `docs/go/topics/03-concurrency/sync-primitives.md` (436 lines) → 100 lines.

## Mutex

```go
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}
```

## RWMutex

```go
type Cache struct {
    mu    sync.RWMutex
    items map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.items[key]
    return val, ok
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = value
}
```

## WaitGroup

```go
func ProcessAll(items []Item) {
    var wg sync.WaitGroup

    for _, item := range items {
        wg.Add(1)  // Before goroutine
        go func(item Item) {
            defer wg.Done()  // Always in defer
            process(item)
        }(item)
    }

    wg.Wait()  // Block until all done
}
```

## Once

```go
var (
    instance *Singleton
    once     sync.Once
)

func GetInstance() *Singleton {
    once.Do(func() {
        instance = &Singleton{}
    })
    return instance
}
```

## Atomic Operations

```go
import "sync/atomic"

type Counter struct {
    value int64
}

func (c *Counter) Inc() int64 {
    return atomic.AddInt64(&c.value, 1)
}

func (c *Counter) Value() int64 {
    return atomic.LoadInt64(&c.value)
}

// CAS (Compare-And-Swap)
func (c *Counter) CAS(old, new int64) bool {
    return atomic.CompareAndSwapInt64(&c.value, old, new)
}
```

## sync.Map

```go
var cache sync.Map

// Store
cache.Store("key", "value")

// Load
value, ok := cache.Load("key")

// LoadOrStore
actual, loaded := cache.LoadOrStore("key", "value")

// Delete
cache.Delete("key")

// Range
cache.Range(func(key, value interface{}) bool {
    // Return false to stop iteration
    return true
})
```
