# Feature Flags Implementation

Complete example showing thread-safe feature flag management.

## Example

<example>
<input>Implement feature flags with thread-safe enable/disable</input>
<output>
```go
package feature

import "sync"

type Flags struct {
    mu           sync.RWMutex
    enabledFlags map[string]bool
}

func New() *Flags {
    return &Flags{
        enabledFlags: make(map[string]bool),
    }
}

func (f *Flags) Enable(name string) {
    f.mu.Lock()
    defer f.mu.Unlock()
    f.enabledFlags[name] = true
}

func (f *Flags) Disable(name string) {
    f.mu.Lock()
    defer f.mu.Unlock()
    f.enabledFlags[name] = false
}

func (f *Flags) Enabled(name string) bool {
    f.mu.RLock()
    defer f.mu.RUnlock()
    return f.enabledFlags[name]
}

func (f *Flags) LoadFromMap(flags map[string]bool) {
    f.mu.Lock()
    defer f.mu.Unlock()
    for k, v := range flags {
        f.enabledFlags[k] = v
    }
}
```
</output>
</example>
