# Async Process Manager Patterns

## Full Process Manager Struct

From v2ray-rs `crates/process/src/manager.rs` — production reference:

```rust
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};
use nix::sys::signal::{kill, Signal};
use nix::unistd::Pid;
use tokio::io::{AsyncBufReadExt, BufReader};
use tokio::process::{Child, Command};
use tokio::sync::broadcast;
use tokio::task::JoinHandle;
use tokio::time::{sleep, timeout};
use std::process::Stdio;
use std::path::PathBuf;

const STOP_TIMEOUT:      Duration = Duration::from_secs(5);
const CRASH_RESTART_DELAY: Duration = Duration::from_secs(2);
const MAX_CRASHES:       usize    = 3;
const CRASH_WINDOW:      Duration = Duration::from_secs(60);

pub struct ProcessManager {
    binary_path:    PathBuf,
    config_path:    PathBuf,
    child:          Option<Child>,
    log_handles:    Vec<JoinHandle<()>>,
    log_tx:         broadcast::Sender<String>,
    crash_times:    Vec<Instant>,
    auto_restart:   bool,
}

impl ProcessManager {
    pub fn new(binary_path: PathBuf, config_path: PathBuf) -> Self {
        let (log_tx, _) = broadcast::channel(256);
        Self {
            binary_path,
            config_path,
            child: None,
            log_handles: vec![],
            log_tx,
            crash_times: vec![],
            auto_restart: true,
        }
    }

    pub fn subscribe_logs(&self) -> broadcast::Receiver<String> {
        self.log_tx.subscribe()
    }

    pub async fn start(&mut self) -> Result<(), ProcessError> {
        if self.child.is_some() {
            return Err(ProcessError::AlreadyRunning);
        }

        let mut child = Command::new(&self.binary_path)
            .args(["run", "-c"])
            .arg(&self.config_path)
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .map_err(ProcessError::Spawn)?;

        self.capture_output(&mut child);
        self.child = Some(child);
        Ok(())
    }

    pub async fn stop(&mut self) -> Result<(), ProcessError> {
        if self.child.is_none() {
            return Ok(());
        }
        self.auto_restart = false;
        self.graceful_stop().await;
        self.auto_restart = true;
        Ok(())
    }

    pub async fn restart(&mut self) -> Result<(), ProcessError> {
        self.stop().await?;
        self.start().await
    }

    fn capture_output(&mut self, child: &mut Child) {
        let tx = self.log_tx.clone();
        if let Some(stdout) = child.stdout.take() {
            let tx = tx.clone();
            self.log_handles.push(tokio::spawn(async move {
                let mut lines = BufReader::new(stdout).lines();
                while let Ok(Some(line)) = lines.next_line().await {
                    let _ = tx.send(line);
                }
            }));
        }
        if let Some(stderr) = child.stderr.take() {
            self.log_handles.push(tokio::spawn(async move {
                let mut lines = BufReader::new(stderr).lines();
                while let Ok(Some(line)) = lines.next_line().await {
                    let _ = tx.send(format!("[err] {line}"));
                }
            }));
        }
    }

    async fn graceful_stop(&mut self) {
        let Some(ref mut child) = self.child else { return };

        if let Some(pid) = child.id() {
            let _ = kill(Pid::from_raw(pid as i32), Signal::SIGTERM);
        }

        let wait_result = timeout(STOP_TIMEOUT, child.wait()).await;
        if wait_result.is_err() {
            child.kill().await.ok();
            child.wait().await.ok();
        }

        for handle in self.log_handles.drain(..) {
            handle.abort();
        }
        self.child = None;
    }

    async fn handle_unexpected_exit(&mut self) {
        if !self.auto_restart {
            return;
        }

        let now = Instant::now();
        self.crash_times.retain(|&t| now.duration_since(t) < CRASH_WINDOW);
        self.crash_times.push(now);

        if self.crash_times.len() >= MAX_CRASHES {
            // Crash loop — surface error to caller
            return;
        }

        sleep(CRASH_RESTART_DELAY).await;
        self.start().await.ok();
    }
}

#[derive(thiserror::Error, Debug)]
pub enum ProcessError {
    #[error("process already running")]
    AlreadyRunning,
    #[error("spawn: {0}")]
    Spawn(#[from] std::io::Error),
}
```

---

## Retry-on-Busy Spawn

Some Linux systems briefly set the `ETXTBSY` error when a binary is being updated:

```rust
async fn try_spawn(binary: &Path, args: &[&str]) -> Result<Child, std::io::Error> {
    const MAX_RETRIES: u32 = 5;
    for attempt in 0..MAX_RETRIES {
        match Command::new(binary).args(args)
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
        {
            Ok(child) => return Ok(child),
            Err(e) if e.kind() == std::io::ErrorKind::ExecutableFileBusy => {
                if attempt == MAX_RETRIES - 1 { return Err(e); }
                sleep(Duration::from_millis(50)).await;
            }
            Err(e) => return Err(e),
        }
    }
    unreachable!()
}
```

---

## PID File

```rust
use std::fs;
use std::path::{Path, PathBuf};

pub struct PidFile(PathBuf);

impl PidFile {
    pub fn write(&self, pid: u32) -> std::io::Result<()> {
        fs::write(&self.0, pid.to_string())
    }

    pub fn remove(&self) -> std::io::Result<()> {
        if self.0.exists() {
            fs::remove_file(&self.0)?;
        }
        Ok(())
    }

    pub fn read(&self) -> Option<u32> {
        fs::read_to_string(&self.0)
            .ok()?
            .trim()
            .parse()
            .ok()
    }
}
```

---

## tokio-util CancellationToken

For structured cancellation (alternative to raw context channels):

```toml
[dependencies]
tokio-util = { version = "0.7", features = ["rt"] }
```

```rust
use tokio_util::sync::CancellationToken;

let token = CancellationToken::new();
let child_token = token.child_token(); // cloned token for subtasks

// Pass to workers
tokio::spawn(async move {
    tokio::select! {
        _ = child_token.cancelled() => { /* cleanup */ }
        _ = do_work()               => { /* normal exit */ }
    }
});

// Cancel all children
token.cancel();
```

---

## Arc<Mutex<T>> for Shared Log Buffer

```rust
use std::sync::{Arc, Mutex};

#[derive(Default)]
pub struct LogBuffer {
    lines: Vec<String>,
    max:   usize,
}

impl LogBuffer {
    pub fn new(max: usize) -> Self {
        Self { lines: vec![], max }
    }

    pub fn push(&mut self, line: String) {
        if self.lines.len() >= self.max {
            self.lines.remove(0); // ring: drop oldest
        }
        self.lines.push(line);
    }

    pub fn snapshot(&self) -> Vec<String> {
        self.lines.clone()
    }
}

// Shared between spawn closure and manager
let buffer = Arc::new(Mutex::new(LogBuffer::new(1000)));

let buf = Arc::clone(&buffer);
tokio::spawn(async move {
    let mut lines = BufReader::new(stdout).lines();
    while let Ok(Some(line)) = lines.next_line().await {
        if let Ok(mut b) = buf.lock() {
            b.push(line);
        }
    }
});
```
