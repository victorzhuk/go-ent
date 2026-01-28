# Environment Variable Configuration

Complete example showing environment variable configuration with caarlos0/env.

## Example

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
