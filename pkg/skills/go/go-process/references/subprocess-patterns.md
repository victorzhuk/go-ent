# Subprocess Patterns Reference

## Windowed Crash Counter

Track recent crashes using a `[]time.Time` slice. Before appending a new crash, filter out entries older than the window. This avoids timers, tickers, and external state.

```go
const (
    maxCrashes  = 3
    crashWindow = 60 * time.Second
)

func retainWithin(ts []time.Time, now time.Time, window time.Duration) []time.Time {
    cutoff := now.Add(-window)
    out := ts[:0] // reuse backing array
    for _, t := range ts {
        if t.After(cutoff) {
            out = append(out, t)
        }
    }
    return out
}

// In watchExit:
m.crashes = retainWithin(m.crashes, time.Now(), crashWindow)
m.crashes = append(m.crashes, time.Now())
if len(m.crashes) >= maxCrashes {
    // enter error state — crash loop
}
```

**Why `ts[:0]`?** Reuses the backing array memory without allocation. Safe because we only append existing elements back.

---

## Log Streaming

Spawn one goroutine per pipe. The `bufio.Scanner` exits naturally when the pipe closes (process exits), so no explicit cancellation is needed.

```go
func (m *Manager) streamLines(r io.Reader) {
    sc := bufio.NewScanner(r)
    for sc.Scan() {
        select {
        case m.logCh <- LogLine{Text: sc.Text(), Time: time.Now()}:
        default:
            // drop if consumer is slow — or use ring buffer
        }
    }
}
```

For high-volume output, replace the channel with a ring buffer:

```go
type RingBuffer struct {
    mu   sync.Mutex
    buf  []LogLine
    head int
    size int
}

func (rb *RingBuffer) Push(line LogLine) {
    rb.mu.Lock()
    rb.buf[rb.head%rb.size] = line
    rb.head++
    rb.mu.Unlock()
}
```

---

## Binary Discovery

Combine well-known paths with `exec.LookPath` as fallback. Always validate executability.

```go
func discover(names []string, wellKnown []string) (string, error) {
    for _, p := range wellKnown {
        if isExecutable(p) {
            return p, nil
        }
    }
    for _, name := range names {
        if p, err := exec.LookPath(name); err == nil {
            return p, nil
        }
    }
    return "", fmt.Errorf("binary not found: tried %v", names)
}

func isExecutable(path string) bool {
    info, err := os.Stat(path)
    return err == nil && info.Mode()&0o111 != 0
}
```

---

## Clean Exit Detection

Distinguish between signal-killed processes (context cancellation, manual stop) and unexpected exits:

```go
func isSignalExit(err error) bool {
    var exitErr *exec.ExitError
    if !errors.As(err, &exitErr) {
        return false
    }
    status, ok := exitErr.Sys().(syscall.WaitStatus)
    return ok && status.Signaled()
}

// In watchExit:
exitErr := cmd.Wait()
if ctx.Err() != nil || isSignalExit(exitErr) {
    return // intentional
}
// otherwise: unexpected crash → apply windowed counter
```

---

## Restart With Delay

Always restart in a separate goroutine, never in the `watchExit` goroutine call stack:

```go
go func() {
    time.Sleep(restartDelay)
    if err := m.Start(ctx); err != nil {
        // log or send to error channel
    }
}()
```

---

## PID File (Optional)

Write PID on start, remove on stop. Guards against stale processes from previous runs.

```go
func writePID(path string, pid int) error {
    return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0o644)
}

func removePID(path string) {
    _ = os.Remove(path)
}

func readPID(path string) (int, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return 0, err
    }
    return strconv.Atoi(strings.TrimSpace(string(data)))
}
```

---

## Environment Passthrough

By default `exec.Cmd.Env` is nil (inherits parent env). Set it explicitly for isolation:

```go
cmd.Env = append(os.Environ(),
    "MY_VAR=value",
    "LOG_LEVEL=debug",
)
```

---

## Common Mistakes

| Mistake | Fix |
|---|---|
| Using `cmd.Process.Kill()` immediately | Always try SIGTERM first, give grace period |
| Calling `cmd.Wait()` from multiple goroutines | Only one goroutine may call `Wait()` |
| Reusing `exec.Cmd` after exit | Create a new `exec.Cmd` for each start |
| Using parent context for stop timeout | Create `context.WithTimeout(context.Background(), ...)` |
| Blocking log channel | Use buffered channel or `select/default` drop |
| No crash window | Pure counter triggers on 3 crashes spread over hours |
