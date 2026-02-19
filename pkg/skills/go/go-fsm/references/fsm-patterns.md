# FSM Patterns Reference

## Transition Table as Data

Define valid transitions as a package-level `map[State][]State`. This separates the rules from the enforcement logic — easy to read, test, and extend.

```go
var valid = map[State][]State{
    StateStopped:  {StateStarting},
    StateStarting: {StateRunning, StateError},
    StateRunning:  {StateStopping, StateError},
    StateStopping: {StateStopped, StateError},
    StateError:    {StateStarting, StateStopped},
}
```

Alternative for small FSMs — `switch` in `CanTransitionTo`:

```go
func (s State) CanTransitionTo(target State) bool {
    switch s {
    case StateStopped:
        return target == StateStarting
    case StateRunning:
        return target == StateStopping || target == StateError
    // ...
    }
    return false
}
```

Use the map approach once you have more than 5 states — the linear scan over the slice is negligible at this scale.

---

## State Events With Payload

When state changes carry context (e.g., error reason, connection metadata), embed it in the event:

```go
type StateEvent struct {
    From, To State
    At       time.Time
    Err      error  // set when transitioning to StateError
    Meta     any    // optional payload (use sparingly)
}
```

Keep `Meta any` as a last resort — prefer typed fields.

---

## Testing State Machines

Test all edges of the transition graph, not just the happy path:

```go
func TestTransitions(t *testing.T) {
    t.Parallel()

    cases := []struct {
        from, to State
        ok       bool
    }{
        {StateStopped, StateStarting, true},
        {StateStopped, StateRunning, false},  // invalid
        {StateRunning, StateStopped, false},  // must go through Stopping
        {StateError, StateStarting, true},
        {StateError, StateRunning, false},
    }

    for _, tc := range cases {
        t.Run(fmt.Sprintf("%s->%s", tc.from, tc.to), func(t *testing.T) {
            t.Parallel()
            sm := NewStateMachine(tc.from, 0)
            err := sm.Transition(tc.to)
            if tc.ok {
                assert.NoError(t, err)
            } else {
                var te *TransitionError
                assert.ErrorAs(t, err, &te)
                assert.Equal(t, tc.from, te.From)
                assert.Equal(t, tc.to, te.To)
            }
        })
    }
}
```

Test concurrent transitions under `-race`:

```go
func TestConcurrentTransitions(t *testing.T) {
    sm := NewStateMachine(StateStopped, 10)
    var wg sync.WaitGroup
    for range 20 {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _ = sm.Transition(StateStarting)
        }()
    }
    wg.Wait()
    // Only one goroutine should succeed; rest get TransitionError
}
```

---

## Composing FSM With Subprocess Manager

The FSM and subprocess manager are separate concerns. The service layer composes them:

```go
type Service struct {
    fsm *StateMachine
    proc *process.Manager
}

func (s *Service) Start(ctx context.Context) error {
    if err := s.fsm.EnsureState(StateStopped, StateError); err != nil {
        return err
    }
    _ = s.fsm.Transition(StateStarting)

    if err := s.proc.Start(ctx); err != nil {
        _ = s.fsm.Transition(StateError)
        return fmt.Errorf("start: %w", err)
    }

    return s.fsm.Transition(StateRunning)
}
```

This keeps the FSM free of I/O and the manager free of state tracking.

---

## Event Fan-Out

If multiple consumers need state events, fan out from a single source channel:

```go
func fanOut(src <-chan StateEvent, n int) []<-chan StateEvent {
    outs := make([]chan StateEvent, n)
    for i := range outs {
        outs[i] = make(chan StateEvent, cap(src))
    }
    go func() {
        for ev := range src {
            for _, ch := range outs {
                select {
                case ch <- ev:
                default:
                }
            }
        }
    }()
    result := make([]<-chan StateEvent, n)
    for i, ch := range outs {
        result[i] = ch
    }
    return result
}
```

---

## Common Mistakes

| Mistake | Fix |
|---|---|
| Stringly-typed states (`"running"`) | `type State int` with `iota` |
| Mutating state outside `Transition()` | All state writes go through the FSM method |
| Blocking event channel send | Use buffered channel + `select/default` |
| Testing only happy path | Test every invalid transition explicitly |
| `sync.Mutex` on every `State()` read | Use `sync.RWMutex`; reads are frequent |
| Panic on invalid transition | Return `*TransitionError` — let caller decide |
