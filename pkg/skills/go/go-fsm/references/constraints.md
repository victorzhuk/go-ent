# go-fsm Constraints

## Include

- `type State int` with `iota` — type-safe, comparable, usable as map key
- `[...]string` array for `String()` — O(1) lookup, no map allocation
- `map[State][]State` adjacency list for transition rules — easy to read and extend
- `sync.RWMutex` — read-lock for state queries, write-lock only for transitions
- `TransitionError` as a named struct with `From`/`To` fields — callers can inspect via `errors.As`
- Buffered event channel — never block the transition caller
- `select/default` when sending to event channel — drop if no consumer
- `EnsureState(...State) error` guard before state-dependent operations
- State constants exported; StateMachine struct private (`type stateMachine struct`)

## Exclude

- String-based states (`state == "running"`) — typo-prone, not comparable
- Global mutable FSM state
- Panicking on invalid transitions — return `error` instead
- Blocking event channel sends — callers must not be held hostage by slow consumers
- Third-party FSM libraries for simple lifecycle machines (< 10 states)
- Methods that mutate state without going through `Transition()` — bypass-free design

## Bound To

- `sync`, `fmt`, `time` — stdlib only for the FSM core
- FSM emits events; it does not act on them — separation of detection and reaction
- Transition table is defined at package level (var) — not dynamically modified at runtime
