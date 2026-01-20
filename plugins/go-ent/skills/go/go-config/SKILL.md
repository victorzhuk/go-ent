---
name: go-config
description: "Handle configuration in Go applications (env, files, flags). Use for config setup or issues."
version: "2.0.0"
author: "go-ent"
tags: ["go", "config", "environment"]
triggers:
  - keywords: ["config", "configuration", "environment", "env var", "feature flag", "secret"]
    file_patterns: ["config.go", "**/config/*.go", "**/*config*.go"]
    weight: 0.8
---

# Go Configuration

<role>
Expert configuration management engineer specializing in Go applications. Focus on environment variables, config files, validation, secrets management, and feature flags with production-grade patterns.
</role>

<instructions>

## Configuration Stack

- **Environment Variables** — github.com/caarlos0/env/v11
- **Config Files** — YAML/JSON with os.ReadFile
- **Validation** — validator/v10 or custom Validate() methods
- **Secrets** — env vars + vault integration (defer to go-sec)
- **Feature Flags** — in-memory or remote flag provider

## Environment Variables Pattern

```go
package config

import (
    "time"

    "github.com/caarlos0/env/v11"
)

type Config struct {
    App AppConfig `envPrefix:"APP_"`
    DB  DBConfig  `envPrefix:"DB_"`
    API APIConfig `envPrefix:"API_"`
}

type AppConfig struct {
    Name    string `env:"NAME" envDefault:"myapp"`
    Port    int    `env:"PORT" envDefault:"8080"`
    Debug   bool   `env:"DEBUG" envDefault:"false"`
    Timeout int    `env:"TIMEOUT" envDefault:"30"`
}

type DBConfig struct {
    DSN         string        `env:"DSN,required"`
    MaxConns    int           `env:"MAX_CONNS" envDefault:"25"`
    MaxIdleTime time.Duration `env:"MAX_IDLE_TIME" envDefault:"5m"`
}

type APIConfig struct {
    Key     string        `env:"KEY" envDefault:""`
    Timeout time.Duration `env:"TIMEOUT" envDefault:"10s"`
}

func LoadFromEnv(getenv func(string) string) (*Config, error) {
    var cfg Config
    if err := env.ParseWithOptions(&cfg, env.Options{Environment: getenv}); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

**Key points**:
- Prefixes for namespacing (APP_, DB_)
- `envDefault` for sensible defaults
- `env:"VAR,required"` for mandatory values
- Injectable `getenv` for testing

## Validation Pattern

```go
func (c *Config) Validate() error {
    if c.App.Port < 1 || c.App.Port > 65535 {
        return fmt.Errorf("invalid port: %d", c.App.Port)
    }
    if c.DB.MaxConns < 1 {
        return fmt.Errorf("max conns must be > 0")
    }
    if c.API.Key == "" && !c.App.Debug {
        return fmt.Errorf("API key required in production")
    }
    return nil
}
```

**Usage**:
```go
cfg, err := config.LoadFromEnv(os.Getenv)
if err != nil {
    return fmt.Errorf("load config: %w", err)
}
if err := cfg.Validate(); err != nil {
    return fmt.Errorf("validate config: %w", err)
}
```

## Config Files (YAML)

```go
package config

import (
    "os"

    "gopkg.in/yaml.v3"
)

type FileConfig struct {
    Server ServerConfig `yaml:"server"`
    DB     DBConfig     `yaml:"database"`
    Features FeatureFlags `yaml:"features"`
}

type ServerConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

type FeatureFlags struct {
    NewCheckout bool `yaml:"new_checkout"`
    BetaAPI     bool `yaml:"beta_api"`
}

func LoadFromFile(path string) (*FileConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read file: %w", err)
    }

    var cfg FileConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse yaml: %w", err)
    }

    return &cfg, nil
}
```

**Pattern**: Env vars override file config
```go
cfgFile, _ := LoadFromFile("config.yaml")
cfgEnv, _ := LoadFromEnv(os.Getenv)

// Merge: env takes precedence
if cfgEnv.DB.DSN != "" {
    cfgFile.DB.DSN = cfgEnv.DB.DSN
}
```

## Feature Flags

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
```

**Usage**:
```go
flags := feature.New()
flags.Enable("new_checkout")

if flags.Enabled("new_checkout") {
    // Use new checkout flow
}
```

## Secrets Management

```go
type Secrets struct {
    APIKey     string `env:"API_KEY,required"`
    DBPassword string `env:"DB_PASSWORD,required"`
}

func LoadSecrets(getenv func(string) string) (*Secrets, error) {
    var s Secrets
    if err := env.ParseWithOptions(&s, env.Options{Environment: getenv}); err != nil {
        return nil, err
    }
    return &s, nil
}
```

**Important**:
- Never commit secrets to git
- Use vault/secret manager for production
- Rotate keys regularly
- Log redaction for sensitive values

## Config Watching (Hot Reload)

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

## Configuration Hierarchy

1. **Defaults** — Hardcoded in struct tags
2. **Config File** — YAML/JSON loaded at startup
3. **Environment Variables** — Override file values
4. **Command-Line Flags** — Override everything (rare)

**Example**:
```go
cfg := DefaultConfig()
LoadFromFile(cfg, "config.yaml")
LoadFromEnv(cfg, os.Getenv)
cfg.Port = flag.Int("port", cfg.Port, "server port")
```

## Best Practices

1. **Injectable getenv** — Pass `os.Getenv` or test function
2. **Validate early** — Fail fast on invalid config
3. **Type safety** — Use struct tags, avoid string maps
4. **Documentation** — Comment config fields
5. **Defaults** — Provide sensible defaults
6. **Namespacing** — Use prefixes (APP_, DB_)
7. **Redaction** — Log config without secrets
8. **No secrets in code** — Always from env/vault

## Redaction for Logging

```go
func (c *Config) String() string {
    return fmt.Sprintf(
        "{App: {Name: %s, Port: %d}, DB: {DSN: %s, MaxConns: %d}, API: {Key: <REDACTED>, Timeout: %s}}",
        c.App.Name, c.App.Port,
        maskDSN(c.DB.DSN), c.DB.MaxConns,
        c.API.Timeout,
    )
}

func maskDSN(dsn string) string {
    if dsn == "" {
        return ""
    }
    if i := strings.LastIndex(dsn, "@"); i != -1 {
        return dsn[:i+1] + "***"
    }
    return "***"
}
```

</instructions>

<constraints>
- Include environment variable parsing with caarlos0/env/v11
- Include validation with custom Validate() method or validator/v10
- Include configuration file loading (YAML/JSON)
- Include feature flags implementation
- Include secrets handling (defer security details to go-sec)
- Include config redaction for logging
- Include configuration hierarchy (defaults → file → env → flags)
- Exclude hardcoding secrets in code
- Exclude committing secrets to version control
- Exclude using global config objects (pass explicitly)
- Exclude parsing environment variables directly with os.Getenv
- Exclude mixing config loading with business logic
- Include injectable getenv for testing
- Always validate config after loading
</constraints>

<edge_cases>
If config validation fails: Fail fast with clear error messages indicating which field is invalid and why.

If required environment variable is missing: Return descriptive error with the variable name and suggest setting it.

If config file doesn't exist: Use defaults or treat as error based on application requirements.

If environment variable is invalid type: Return parse error with expected type and actual value.

If secrets are missing in non-debug mode: Fail startup to prevent running with incomplete security.

If multiple config sources conflict: Define clear precedence order (defaults < file < env < flags) and document it.

If hot-reload is required: Use file watcher pattern and notify dependent components of changes.

If config grows too large: Split into logical groups (AppConfig, DBConfig, APIConfig) with nested structs.

If environment-specific configs are needed: Use environment name (dev/staging/prod) to load different config files or defaults.

If feature flags need remote management: Integrate with flag provider (LaunchDarkly, Unleash) instead of in-memory flags.

If secrets require encryption: Delegate to go-sec skill for vault integration and encryption patterns.

</edge_cases>

<examples>

<example>
<input>Create config structure with environment variables for a web server</input>
<output>
```go
package config

import (
    "time"

    "github.com/caarlos0/env/v11"
)

type Config struct {
    Server ServerConfig `envPrefix:"SERVER_"`
    DB     DBConfig     `envPrefix:"DB_"`
}

type ServerConfig struct {
    Host    string        `env:"HOST" envDefault:"0.0.0.0"`
    Port    int           `env:"PORT" envDefault:"8080"`
    Timeout time.Duration `env:"TIMEOUT" envDefault:"30s"`
    Debug   bool          `env:"DEBUG" envDefault:"false"`
}

type DBConfig struct {
    DSN      string        `env:"DSN,required"`
    MaxConns int           `env:"MAX_CONNS" envDefault:"25"`
    Timeout  time.Duration `env:"TIMEOUT" envDefault:"5s"`
}

func LoadFromEnv(getenv func(string) string) (*Config, error) {
    var cfg Config
    if err := env.ParseWithOptions(&cfg, env.Options{Environment: getenv}); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```
</output>
</example>

<example>
<input>Load config from YAML file with validation</input>
<output>
```go
package config

import (
    "fmt"
    "os"

    "gopkg.in/yaml.v3"
)

type Config struct {
    Server ServerConfig `yaml:"server"`
    DB     DBConfig     `yaml:"database"`
}

type ServerConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

type DBConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Database string `yaml:"database"`
}

func LoadFromFile(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read file: %w", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse yaml: %w", err)
    }

    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("validate: %w", err)
    }

    return &cfg, nil
}

func (c *Config) Validate() error {
    if c.Server.Port < 1 || c.Server.Port > 65535 {
        return fmt.Errorf("invalid server port: %d", c.Server.Port)
    }
    if c.DB.Host == "" {
        return fmt.Errorf("database host required")
    }
    if c.DB.Database == "" {
        return fmt.Errorf("database name required")
    }
    return nil
}
```
</output>
</example>

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

<example>
<input>Handle secrets with environment variables and redaction for logging</input>
<output>
```go
package config

import (
    "fmt"
    "strings"

    "github.com/caarlos0/env/v11"
)

type Secrets struct {
    APIKey     string `env:"API_KEY,required"`
    DBPassword string `env:"DB_PASSWORD,required"`
    Token      string `env:"TOKEN,required"`
}

func LoadSecrets(getenv func(string) string) (*Secrets, error) {
    var s Secrets
    if err := env.ParseWithOptions(&s, env.Options{Environment: getenv}); err != nil {
        return nil, err
    }
    return &s, nil
}

func (s *Secrets) String() string {
    return fmt.Sprintf(
        "{APIKey: %s, DBPassword: %s, Token: %s}",
        s.mask(s.APIKey),
        s.mask(s.DBPassword),
        s.mask(s.Token),
    )
}

func (s *Secrets) mask(val string) string {
    if val == "" {
        return "<empty>"
    }
    if len(val) <= 8 {
        return "***"
    }
    return val[:4] + strings.Repeat("*", len(val)-8) + val[len(val)-4:]
}
```
</output>
</example>

<example>
<input>Merge config from file and environment with env taking precedence</input>
<output>
```go
package config

import (
    "os"

    "github.com/caarlos0/env/v11"
    "gopkg.in/yaml.v3"
)

type Config struct {
    Server ServerConfig `yaml:"server" envPrefix:"SERVER_"`
    DB     DBConfig     `yaml:"database" envPrefix:"DB_"`
}

type ServerConfig struct {
    Host string `yaml:"host" env:"HOST" envDefault:"0.0.0.0"`
    Port int    `yaml:"port" env:"PORT" envDefault:"8080"`
}

type DBConfig struct {
    DSN string `yaml:"dsn" env:"DSN,required"`
}

func Load(getenv func(string) string, file string) (*Config, error) {
    cfg, err := LoadFromFile(file)
    if err != nil {
        cfg = &Config{}
    }

    if err := env.ParseWithOptions(cfg, env.Options{Environment: getenv}); err != nil {
        return nil, fmt.Errorf("parse env: %w", err)
    }

    return cfg, nil
}

func LoadFromFile(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil && !os.IsNotExist(err) {
        return nil, err
    }

    if os.IsNotExist(err) {
        return &Config{}, nil
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse yaml: %w", err)
    }

    return &cfg, nil
}
```
</output>
</example>

</examples>

<output_format>
Provide configuration guidance with the following structure:

1. **Config Structure**: Nested structs with env tags and prefixes
2. **Environment Variables**: caarlos0/env/v11 with defaults and validation
3. **Config Files**: YAML/JSON loading with file watching if needed
4. **Feature Flags**: Thread-safe flag management
5. **Secrets**: Secure handling with redaction for logging
6. **Validation**: Early validation with clear error messages
7. **Merging**: Config hierarchy (defaults → file → env → flags)
8. **Examples**: Complete, runnable code with proper error handling

Focus on production-ready configuration patterns with testability and security.
</output_format>
