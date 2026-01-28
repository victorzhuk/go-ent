# Configuration Merging

Complete example showing how to merge configuration from file and environment variables.

## Example

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
