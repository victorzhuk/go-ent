# Secrets Handling with Redaction

Complete example showing secure secrets handling with logging redaction.

## Example

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
