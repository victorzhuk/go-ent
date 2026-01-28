# Configuration Hot Reload

Complete example showing configuration file watching for hot reload.

## Example

<example>
<input>Implement configuration hot reload with file watcher</input>
<output>
```go
package config

import (
    "os"
    "time"
)

type Watcher struct {
    path     string
    interval time.Duration
    callback func(*Config)
    stop     chan struct{}
}

func NewWatcher(path string, interval time.Duration, callback func(*Config)) *Watcher {
    return &Watcher{
        path:     path,
        interval: interval,
        callback: callback,
        stop:     make(chan struct{}),
    }
}

func (w *Watcher) Start() {
    var lastMod time.Time

    for {
        select {
        case <-w.stop:
            return
        case <-time.After(w.interval):
            info, err := os.Stat(w.path)
            if err != nil {
                continue
            }
            if info.ModTime().After(lastMod) {
                lastMod = info.ModTime()
                if cfg, err := LoadFromFile(w.path); err == nil {
                    w.callback(cfg)
                }
            }
        }
    }
}

func (w *Watcher) Stop() {
    close(w.stop)
}
```
</output>
</example>
