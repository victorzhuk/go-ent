# Data Structures

Go's built-in and container package data structures.

## Quick Reference

| Type             | Use When                                 |
|------------------|------------------------------------------|
| Slice `[]T`      | Dynamic array, most common collection    |
| Map `map[K]V`    | Key-value lookup, K must be comparable   |
| Array `[N]T`     | Fixed size, value type, rare             |
| `container/list` | Doubly-linked list, insertions/deletions |
| `container/heap` | Priority queue                           |
| `container/ring` | Circular list                            |

## Slices

### Basic Operations

```go
// Creation
s := make([]int, 0, 10) // length 0, capacity 10
s := []int{1, 2, 3}     // literal

// Append
s = append(s, 4, 5, 6)
s = append(s, otherSlice...)

// Slicing
sub := s[1:3]  // elements 1, 2
sub := s[:3]   // first 3
sub := s[2:]   // from index 2 to end
sub := s[:]    // full slice

// Length and capacity
len(s)  // number of elements
cap(s)  // capacity

// Clear (Go 1.21+)
clear(s)  // sets all elements to zero value, keeps capacity
```

### Pre-allocate When Size Known

```go
// Good - pre-allocate
users := make([]User, 0, 100)
for i := 0; i < 100; i++ {
    users = append(users, User{ID: i})
}

// Bad - multiple allocations
var users []User
for i := 0; i < 100; i++ {
    users = append(users, User{ID: i})
}
```

### Slice Tricks

```go
// Delete element
s = append(s[:i], s[i+1:]...)

// Insert element
s = append(s[:i], append([]int{x}, s[i:]...)...)

// Reverse
for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
    s[i], s[j] = s[j], s[i]
}

// Filter (in-place)
n := 0
for _, x := range s {
    if keep(x) {
        s[n] = x
        n++
    }
}
s = s[:n]
```

## Maps

### Basic Operations

```go
// Creation
m := make(map[string]int)
m := map[string]int{"a": 1, "b": 2}

// Set
m["key"] = value

// Get
value := m["key"]           // returns zero value if not exists
value, ok := m["key"]       // ok is false if not exists

// Delete
delete(m, "key")

// Check existence
if _, ok := m["key"]; ok {
    // key exists
}

// Iterate
for k, v := range m {
    fmt.Println(k, v)
}

// Length
len(m)

// Clear (Go 1.21+)
clear(m)  // removes all entries
```

### Maps are Not Safe for Concurrent Access

```go
// Use sync.Map or sync.RWMutex
type SafeMap struct {
    mu sync.RWMutex
    m  map[string]int
}

func (sm *SafeMap) Get(key string) (int, bool) {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    val, ok := sm.m[key]
    return val, ok
}

func (sm *SafeMap) Set(key string, value int) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    sm.m[key] = value
}
```

## Container Packages

### container/list (Doubly-Linked List)

```go
import "container/list"

l := list.New()

// Add elements
e1 := l.PushFront(1)
e2 := l.PushBack(2)
l.InsertAfter(3, e1)

// Iterate
for e := l.Front(); e != nil; e = e.Next() {
    fmt.Println(e.Value)
}

// Remove
l.Remove(e1)
```

### container/heap (Priority Queue)

```go
import "container/heap"

type IntHeap []int

func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntHeap) Push(x interface{}) {
    *h = append(*h, x.(int))
}

func (h *IntHeap) Pop() interface{} {
    old := *h
    n := len(old)
    x := old[n-1]
    *h = old[0 : n-1]
    return x
}

// Usage
h := &IntHeap{2, 1, 5}
heap.Init(h)
heap.Push(h, 3)
min := heap.Pop(h)  // 1
```

## Swiss Tables (Go 1.24+)

```go
// Go 1.24+ uses Swiss Tables for faster map operations
// No API changes, just performance improvements
// ~15-30% faster for most workloads
```

## See Also

- [Generics](./generics.md) - Type-safe containers
- [Slice tricks](https://go.dev/wiki/SliceTricks)
