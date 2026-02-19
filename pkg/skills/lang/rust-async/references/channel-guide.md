# Tokio Channel Selection Guide

## Choosing the Right Channel

| Need | Channel | Crate |
|---|---|---|
| One sender, one receiver, bounded | `mpsc::channel(n)` | tokio |
| Many senders, one receiver | `mpsc::channel(n)` | tokio |
| One sender, many receivers (fan-out) | `broadcast::channel(n)` | tokio |
| Single value, written once | `oneshot::channel()` | tokio |
| State shared as latest value | `watch::channel(init)` | tokio |
| Unbounded queue (backpressure risk) | `mpsc::unbounded_channel()` | tokio |

## broadcast — Fan-Out Pattern

```rust
use tokio::sync::broadcast;

// Publisher holds Sender; capacity = max buffered messages per lagged receiver
let (tx, _) = broadcast::channel::<Event>(128);

// Each subscriber gets their own Receiver with independent position
let mut rx_ui  = tx.subscribe();
let mut rx_log = tx.subscribe();

// Send to all
tx.send(Event::StateChanged(new_state)).ok();

// Receive (each receiver gets the message independently)
let event = rx_ui.recv().await?;
```

**Capacity advice**: set capacity to the maximum number of events you expect to buffer between polls. Too small → lagged errors. Too large → memory waste. 64–256 is typical for state events; 1024+ for log lines.

## watch — Latest Value Pattern

Use when consumers only care about the current value, not the history:

```rust
use tokio::sync::watch;

let (tx, rx) = watch::channel(ProcessState::Stopped);

// Publisher
tx.send(ProcessState::Running).ok();

// Consumer — blocks until value changes
let state = rx.borrow_and_update(); // or rx.changed().await?
```

## oneshot — Single Response Pattern

Use for request-response or signaling task completion:

```rust
use tokio::sync::oneshot;

let (result_tx, result_rx) = oneshot::channel::<Result<String, Error>>();

tokio::spawn(async move {
    let result = do_work().await;
    result_tx.send(result).ok();
});

match result_rx.await {
    Ok(Ok(data)) => println!("got: {data}"),
    Ok(Err(e))   => eprintln!("work failed: {e}"),
    Err(_)       => eprintln!("worker dropped without responding"),
}
```

## select! Patterns

### Race: first result wins

```rust
tokio::select! {
    result = operation_a() => handle_a(result),
    result = operation_b() => handle_b(result),
}
```

### Cancellation via channel

```rust
let (cancel_tx, mut cancel_rx) = tokio::sync::oneshot::channel::<()>();

tokio::spawn(async move {
    tokio::select! {
        _ = &mut cancel_rx => { /* cancelled */ }
        _ = do_long_work() => { /* completed */ }
    }
});

// Cancel from elsewhere
cancel_tx.send(()).ok();
```

### Loop with shutdown

```rust
loop {
    tokio::select! {
        Some(msg) = rx.recv() => process(msg).await,
        _ = shutdown_signal() => break,
    }
}
```

## Common Mistakes

| Mistake | Fix |
|---|---|
| Using `broadcast` for request-response | Use `oneshot` |
| Using `mpsc` for state observation | Use `watch` |
| Ignoring `RecvError::Lagged` | Log the gap, don't panic |
| Holding `std::sync::MutexGuard` across `.await` | Use `tokio::sync::Mutex` |
| Calling `recv()` in a tight loop without `select!` | Add a shutdown branch to `select!` |
| Forgetting `handle.await` after `abort()` | Always `await` aborted handles for cleanup |
