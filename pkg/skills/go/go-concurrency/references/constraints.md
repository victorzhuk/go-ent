# Constraints

- Include context cancellation for all long-running goroutines
- Include proper channel close semantics (close by sender, range by receiver)
- Include WaitGroup or errgroup for goroutine lifecycle management
- Include buffered channels for producer-consumer patterns
- Include mutex/RWMutex for protecting shared state
- Include race detection in CI/CD (`go test -race`)
- Include timeout handling with context.WithTimeout
- Exclude launching goroutines without cancellation mechanism
- Exclude accessing shared state without synchronization
- Exclude closing channels from receiver side
- Exclude ignoring context cancellation in loops
- Exclude unbounded goroutine spawning without limits
- Bound to safe, deadlock-free concurrent code
- Follow "communicate by sharing memory" principle (prefer channels)
