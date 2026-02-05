# Configuration

Environment-based configuration using caarlos0/env/v11.

## Quick Reference

| Feature | Tag/Option | Example |
|---------|-----------|---------|
| Basic | `env:"NAME"` | `Port int \`env:"PORT"\`` |
| Default | `envDefault:"value"` | `Port int \`env:"PORT" envDefault:"8080"\`` |
| Required | `,required` | `URL string \`env:"DATABASE_URL,required"\`` |
| Separator | `envSeparator:","` | `Hosts []string \`env:"HOSTS" envSeparator:","\`` |
| Nested | Struct embedding | `Server ServerConfig \`envPrefix:"SERVER_"\`` |
| Custom parser | `env.ParseWithOptions` | See Custom Parsers section |
| Validation | Post-parse check | See Validation section |

```go
import "github.com/caarlos0/env/v11"

type Config struct {
    Port     int    `env:"PORT" envDefault:"8080"`
    Database string `env:"DATABASE_URL,required"`
    LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
}

var cfg Config
env.Parse(&cfg)
```

## Full Example

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
}

type ServerConfig struct {
    Port         int           `env:"SERVER_PORT" envDefault:"8080"`
    ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"10s"`
    WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"10s"`
}

type DatabaseConfig struct {
    URL      string `env:"DATABASE_URL,required"`
    MaxConns int    `env:"DATABASE_MAX_CONNS" envDefault:"10"`
}

func LoadConfig() (*Config, error) {
    var cfg Config
    if err := env.Parse(&cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }
    return &cfg, nil
}
```

## Testable Configuration

```go
func LoadFromEnv(getenv func(string) string) (*Config, error) {
    var cfg Config
    if err := env.ParseWithOptions(&cfg, env.Options{
        Environment: getenv,
    }); err != nil {
        return nil, err
    }
    return &cfg, nil
}

// Production
cfg, _ := LoadFromEnv(os.Getenv)

// Testing
cfg, _ := LoadFromEnv(func(key string) string {
    return map[string]string{
        "PORT": "9000",
    }[key]
})
```

## Nested Configuration

```go
type Config struct {
    Server   ServerConfig   `envPrefix:"SERVER_"`
    Database DatabaseConfig `envPrefix:"DB_"`
    Redis    RedisConfig    `envPrefix:"REDIS_"`
}

type ServerConfig struct {
    Host string `env:"HOST" envDefault:"0.0.0.0"`
    Port int    `env:"PORT" envDefault:"8080"`
}

type DatabaseConfig struct {
    Host     string `env:"HOST,required"`
    Port     int    `env:"PORT" envDefault:"5432"`
    Name     string `env:"NAME,required"`
    User     string `env:"USER,required"`
    Password string `env:"PASSWORD,required"`
}

// Environment variables:
// SERVER_HOST=0.0.0.0
// SERVER_PORT=8080
// DB_HOST=postgres
// DB_PORT=5432
// DB_NAME=myapp
// DB_USER=admin
// DB_PASSWORD=secret
```

## Custom Parsers

```go
type LogLevel string

func (l *LogLevel) UnmarshalText(text []byte) error {
    level := strings.ToLower(string(text))
    switch level {
    case "debug", "info", "warn", "error":
        *l = LogLevel(level)
        return nil
    default:
        return fmt.Errorf("invalid log level: %s", level)
    }
}

type Config struct {
    Level LogLevel `env:"LOG_LEVEL" envDefault:"info"`
}

// With options
cfg := Config{}
err := env.ParseWithOptions(&cfg, env.Options{
    Environment: os.Getenv,
    OnSet: func(tag string, value interface{}, isDefault bool) {
        fmt.Printf("Set %s to %v (default: %t)\n", tag, value, isDefault)
    },
})
```

## Validation

```go
type Config struct {
    Port     int    `env:"PORT" envDefault:"8080"`
    Database string `env:"DATABASE_URL,required"`
    MaxConns int    `env:"MAX_CONNS" envDefault:"10"`
}

func LoadConfig() (*Config, error) {
    var cfg Config
    if err := env.Parse(&cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }

    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("validate config: %w", err)
    }

    return &cfg, nil
}

func (c *Config) Validate() error {
    if c.Port < 1 || c.Port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
    }

    if c.MaxConns < 1 {
        return fmt.Errorf("max connections must be positive, got %d", c.MaxConns)
    }

    if !strings.HasPrefix(c.Database, "postgres://") {
        return fmt.Errorf("database URL must start with postgres://")
    }

    return nil
}
```

## YAML Fallback

```go
import (
    "gopkg.in/yaml.v3"
    "github.com/caarlos0/env/v11"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
}

func LoadConfig(path string) (*Config, error) {
    var cfg Config

    // Load from file if exists
    if path != "" {
        data, err := os.ReadFile(path)
        if err != nil && !os.IsNotExist(err) {
            return nil, fmt.Errorf("read config file: %w", err)
        }

        if err == nil {
            if err := yaml.Unmarshal(data, &cfg); err != nil {
                return nil, fmt.Errorf("parse yaml: %w", err)
            }
        }
    }

    // Override with env vars (takes precedence)
    if err := env.Parse(&cfg); err != nil {
        return nil, fmt.Errorf("parse env: %w", err)
    }

    return &cfg, nil
}

// config.yaml:
// server:
//   port: 3000
// database:
//   url: postgres://localhost/dev
//
// Override with: DATABASE_URL=postgres://prod/myapp
```

## Common Mistakes

| Mistake | Problem | Solution |
|---------|---------|----------|
| Missing `env` tag | Field not loaded | Add `env:"VAR_NAME"` tag |
| Wrong default type | `envDefault:"true"` for int | Use correct type: `envDefault:"1"` |
| No validation | Invalid values crash at runtime | Add `Validate()` method |
| Hardcoding secrets | Credentials in code | Always use env vars for secrets |
| Not documenting vars | Team doesn't know what to set | Add `.env.example` or README |
| Ignoring parse errors | Silent failures | Always check `env.Parse()` error |
| Missing required fields | App starts with nil values | Use `,required` tag |
| No testable injection | Hard to unit test | Use `LoadFromEnv(getenv)` pattern |

## See Also

- [Cobra CLI](./cobra.md)
- [Functional Options](./functional-options.md)
- [Docker](../13-devops/docker.md)
