# Generics

Go 1.18+ generics enable type-safe generic data structures and algorithms. Use when you'd otherwise need `interface{}` or code duplication.

## Quick Reference

| Pattern                 | Use When                             |
|-------------------------|--------------------------------------|
| `func F[T any](v T)`    | Generic function                     |
| `type Stack[T any]`     | Generic type                         |
| `T comparable`          | Need `==` or `!=` operations         |
| `T constraints.Ordered` | Need `<`, `>`, `<=`, `>=`            |
| `iter.Seq[T]`           | Go 1.23+ iteration (range over func) |

## Basic Generic Functions

### Simple Generic Function

```go
func First[T any](slice []T) (T, bool) {
    if len(slice) == 0 {
        var zero T
        return zero, false
    }
    return slice[0], true
}

// Usage - type inference
val, ok := First([]int{1, 2, 3})        // T inferred as int
str, ok := First([]string{"a", "b"})    // T inferred as string
```

### Multiple Type Parameters

```go
func Map[T any, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

// Usage
nums := []int{1, 2, 3}
strings := Map(nums, func(n int) string {
    return fmt.Sprintf("num-%d", n)
})
// Result: ["num-1", "num-2", "num-3"]
```

## Type Constraints

### comparable Constraint

```go
func Contains[T comparable](slice []T, target T) bool {
    for _, v := range slice {
        if v == target { // requires comparable
            return true
        }
    }
    return false
}

// Works with any comparable type
Contains([]int{1, 2, 3}, 2)              // true
Contains([]string{"a", "b"}, "c")        // false
```

### constraints.Ordered

```go
import "golang.org/x/exp/constraints"

func Min[T constraints.Ordered](a, b T) T {
    if a < b { // requires Ordered (< operator)
        return a
    }
    return b
}

Min(10, 20)        // 10
Min(3.14, 2.71)    // 2.71
Min("foo", "bar")  // "bar"
```

### Custom Constraints

```go
type Number interface {
    ~int | ~int64 | ~float64
}

func Sum[T Number](values []T) T {
    var sum T
    for _, v := range values {
        sum += v
    }
    return sum
}

Sum([]int{1, 2, 3})           // 6
Sum([]float64{1.5, 2.5})      // 4.0
```

**Key points:**
- `~` allows underlying type (e.g., `type MyInt int`)
- Without `~`, only exact types match
- Use for numeric operations, custom comparisons

## Generic Types

### Generic Stack

```go
type Stack[T any] struct {
    items []T
}

func NewStack[T any]() *Stack[T] {
    return &Stack[T]{
        items: make([]T, 0),
    }
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }

    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}

// Usage
stack := NewStack[int]()
stack.Push(1)
stack.Push(2)
val, ok := stack.Pop() // val = 2
```

### Generic Map

```go
type SafeMap[K comparable, V any] struct {
    mu    sync.RWMutex
    items map[K]V
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
    return &SafeMap[K, V]{
        items: make(map[K]V),
    }
}

func (m *SafeMap[K, V]) Set(key K, value V) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.items[key] = value
}

func (m *SafeMap[K, V]) Get(key K) (V, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    val, ok := m.items[key]
    return val, ok
}

// Usage
cache := NewSafeMap[string, User]()
cache.Set("user:1", User{ID: 1})
user, ok := cache.Get("user:1")
```

## Standard Library Generics

### slices Package (Go 1.21+)

```go
import "slices"

// Contains
slices.Contains([]int{1, 2, 3}, 2) // true

// Index
idx := slices.Index([]string{"a", "b", "c"}, "b") // 1

// Sort
nums := []int{3, 1, 2}
slices.Sort(nums) // [1, 2, 3]

// SortFunc with custom comparator
slices.SortFunc(users, func(a, b User) int {
    return strings.Compare(a.Name, b.Name)
})

// Compact (remove consecutive duplicates)
nums = []int{1, 1, 2, 2, 3}
slices.Compact(nums) // [1, 2, 3]

// Clone
original := []int{1, 2, 3}
copy := slices.Clone(original)

// Delete
nums = slices.Delete(nums, 1, 3) // Remove elements [1:3]

// Insert
nums = slices.Insert(nums, 1, 99, 100)

// Replace
nums = slices.Replace(nums, 0, 2, 7, 8, 9)
```

### maps Package (Go 1.21+)

```go
import "maps"

// Clone
original := map[string]int{"a": 1, "b": 2}
copy := maps.Clone(original)

// Copy (merge into existing map)
dst := map[string]int{"a": 1}
src := map[string]int{"b": 2}
maps.Copy(dst, src) // dst now {"a": 1, "b": 2}

// DeleteFunc
m := map[string]int{"a": 1, "b": 2, "c": 3}
maps.DeleteFunc(m, func(k string, v int) bool {
    return v > 1 // Delete values > 1
})
// m is now {"a": 1}

// Equal
maps.Equal(map[string]int{"a": 1}, map[string]int{"a": 1}) // true

// EqualFunc with custom equality
maps.EqualFunc(m1, m2, func(v1, v2 User) bool {
    return v1.ID == v2.ID
})
```

## Iteration (Go 1.23+)

### iter.Seq and range over func

```go
import "iter"

// Iterator function
func Count(n int) iter.Seq[int] {
    return func(yield func(int) bool) {
        for i := 0; i < n; i++ {
            if !yield(i) {
                return // Consumer stopped iteration
            }
        }
    }
}

// Usage with range
for i := range Count(5) {
    fmt.Println(i) // 0, 1, 2, 3, 4
}
```

### iter.Seq2 (Key-Value Pairs)

```go
func Enumerate[T any](slice []T) iter.Seq2[int, T] {
    return func(yield func(int, T) bool) {
        for i, v := range slice {
            if !yield(i, v) {
                return
            }
        }
    }
}

// Usage
for idx, val := range Enumerate([]string{"a", "b", "c"}) {
    fmt.Printf("%d: %s\n", idx, val)
}
```

### Pull Iterators

```go
import "iter"

func main() {
    next, stop := iter.Pull(Count(5))
    defer stop()

    val, ok := next() // 0, true
    val, ok = next()  // 1, true
    // ...
}
```

## Advanced Patterns

### Generic Result Type

```go
type Result[T any] struct {
    value T
    err   error
}

func (r Result[T]) Unwrap() (T, error) {
    return r.value, r.err
}

func (r Result[T]) OrDefault(def T) T {
    if r.err != nil {
        return def
    }
    return r.value
}

// Usage
func fetchUser(id string) Result[User] {
    user, err := db.GetUser(id)
    return Result[User]{value: user, err: err}
}

result := fetchUser("123")
user, err := result.Unwrap()
// Or
user := result.OrDefault(User{})
```

### Generic Option Type

```go
type Option[T any] struct {
    value T
    valid bool
}

func Some[T any](value T) Option[T] {
    return Option[T]{value: value, valid: true}
}

func None[T any]() Option[T] {
    return Option[T]{}
}

func (o Option[T]) IsSome() bool {
    return o.valid
}

func (o Option[T]) Unwrap() T {
    if !o.valid {
        panic("unwrap on None")
    }
    return o.value
}

func (o Option[T]) UnwrapOr(def T) T {
    if !o.valid {
        return def
    }
    return o.value
}

// Usage
func findUser(id string) Option[User] {
    user, err := db.GetUser(id)
    if err != nil {
        return None[User]()
    }
    return Some(user)
}
```

### Generic Constraints Interface

```go
type Stringer interface {
    String() string
}

func Format[T Stringer](items []T) []string {
    result := make([]string, len(items))
    for i, item := range items {
        result[i] = item.String()
    }
    return result
}
```

## Common Mistakes

| Mistake                           | Fix                                    |
|-----------------------------------|----------------------------------------|
| `func F[T any](v interface{})`    | Use `T` parameter type directly        |
| Over-using generics               | Only use when avoiding duplication     |
| Missing `comparable` constraint   | Add when using `==` or `!=`            |
| `T comparable` with structs       | Only works with comparable fields      |
| Type parameter on method receiver | Not supported, use type-level generics |

## When to Use Generics

### Use Generics When:
```go
// ✓ Container types (slice, map, set)
type Set[T comparable] map[T]struct{}

// ✓ Utility functions operating on any type
func Filter[T any](slice []T, pred func(T) bool) []T

// ✓ Avoiding interface{} with type assertions
func GetOrDefault[T any](m map[string]T, key string, def T) T
```

### Don't Use Generics When:
```go
// ✗ Function only used with one type
func addInts(a, b int) int { return a + b }

// ✗ Type-specific logic required
func processUser(u User) error { /* user-specific logic */ }

// ✗ Interface methods already provide abstraction
func Write(w io.Writer, data []byte) error
```

## Performance Considerations

### Generic Function Overhead

```go
// Generics use dictionary-passing (Go 1.18-1.22)
// Small overhead compared to specialized functions

// Compiler may specialize hot paths (Go 1.23+)
// Performance approaching hand-written code
```

### When Performance Matters

```go
// Profile before optimizing
// Consider monomorphization (duplicating code per type) for hot paths

//go:inline
func fastPath(nums []int) int {
    // Type-specific implementation
}

func genericPath[T constraints.Integer](nums []T) T {
    // Generic implementation
}
```

## See Also

- [Type Parameters Proposal](https://go.googlesource.com/proposal/+/HEAD/design/43651-type-parameters.md)
- [slices package](https://pkg.go.dev/slices)
- [maps package](https://pkg.go.dev/maps)
- [iter package](https://pkg.go.dev/iter) (Go 1.23+)
- [constraints package](https://pkg.go.dev/golang.org/x/exp/constraints)
