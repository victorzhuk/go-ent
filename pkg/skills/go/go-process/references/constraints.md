# go-process Constraints

## Include

- `exec.CommandContext` — always pass context so the OS handles cancellation
- SIGTERM before SIGKILL — allow graceful shutdown for the child
- Fresh `context.Background()` for stop timeout — the parent context is usually already cancelled
- Windowed crash counter using `time.Time` slice — `retainWithin` filter pattern
- Per-pipe goroutines for stdout/stderr streaming
- Single `sync.Mutex` around `cmd *exec.Cmd` and crash timestamp slice
- Domain error types: `ErrBinaryNotFound`, `ErrAlreadyRunning`, `ErrCrashLoop`
- Buffered log channel to avoid blocking the streaming goroutine

## Exclude

- `os.Process.Kill()` without first attempting SIGTERM
- Shared state across goroutines without a mutex
- Unbuffered channels for high-volume log lines
- `time.Sleep` in the critical path (only in watchExit goroutine before restart)
- Global mutable state for process handles
- `log.Fatal` / `os.Exit` outside of `main()`
- Inline magic numbers — use named constants for timeouts and thresholds

## Bound To

- `os/exec`, `syscall`, `bufio`, `sync`, `context`, `time` — stdlib only
- No external process management libraries
- Crash recovery only for unexpected exits — clean exit (code 0) does not trigger restart
- Log streaming goroutines must terminate when the pipe closes (scanner loop exits naturally)
