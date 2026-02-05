# Functional Options

Clean, extensible API design pattern for optional parameters, introduced by Dave Cheney. Enables backward-compatible API evolution without breaking changes.

## Quick Reference

| Pattern                         | Purpose                          |
|---------------------------------|----------------------------------|
| `type Option func(*Config)`     | Option function signature        |
| `func WithX(v T) Option`        | Constructor for option           |
| `func New(opts ...Option)`      | Variadic constructor             |
| `Config{field: defaultValue}`   | Default values in struct         |
| `for _, opt := range opts`      | Apply options in constructor     |

## Basic Pattern

### Option Type

```go
type Server struct {
    addr         string
    timeout      time.Duration
    maxConns     int
    readTimeout  time.Duration
    writeTimeout time.Duration
}

type Option func(*Server)

func WithTimeout(d time.Duration) Option {
    return func(s *Server) {
        s.timeout = d
    }
}

func WithMaxConns(n int) Option {
    return func(s *Server) {
        s.maxConns = n
    }
}

func WithReadTimeout(d time.Duration) Option {
    return func(s *Server) {
        s.readTimeout = d
    }
}

func WithWriteTimeout(d time.Duration) Option {
    return func(s *Server) {
        s.writeTimeout = d
    }
}
```

### Constructor with Defaults

```go
func New(addr string, opts ...Option) *Server {
    // Sensible defaults
    s := &Server{
        addr:         addr,
        timeout:      30 * time.Second,
        maxConns:     100,
        readTimeout:  10 * time.Second,
        writeTimeout: 10 * time.Second,
    }

    // Apply options
    for _, opt := range opts {
        opt(s)
    }

    return s
}

// Usage
srv := New(":8080")
srv := New(":8080", WithTimeout(60*time.Second))
srv := New(":8080", WithMaxConns(200), WithTimeout(60*time.Second))
```

## Default Values

### Explicit Defaults in Constructor

```go
type Client struct {
    baseURL    string
    timeout    time.Duration
    retries    int
    userAgent  string
    headers    map[string]string
}

type ClientOption func(*Client)

func NewClient(baseURL string, opts ...ClientOption) *Client {
    c := &Client{
        baseURL:   baseURL,
        timeout:   30 * time.Second,  // Default
        retries:   3,                 // Default
        userAgent: "go-client/1.0",   // Default
        headers:   make(map[string]string),
    }

    for _, opt := range opts {
        opt(c)
    }

    return c
}

func WithClientTimeout(d time.Duration) ClientOption {
    return func(c *Client) {
        c.timeout = d
    }
}

func WithRetries(n int) ClientOption {
    return func(c *Client) {
        c.retries = n
    }
}

func WithUserAgent(ua string) ClientOption {
    return func(c *Client) {
        c.userAgent = ua
    }
}

func WithHeader(key, value string) ClientOption {
    return func(c *Client) {
        c.headers[key] = value
    }
}
```

### Zero Values as Defaults

```go
type Logger struct {
    level  string
    output io.Writer
    prefix string
}

type LoggerOption func(*Logger)

func NewLogger(opts ...LoggerOption) *Logger {
    l := &Logger{
        level:  "info",      // Default
        output: os.Stdout,   // Default
        prefix: "",          // Zero value acceptable
    }

    for _, opt := range opts {
        opt(l)
    }

    return l
}

func WithLevel(lvl string) LoggerOption {
    return func(l *Logger) {
        l.level = lvl
    }
}

func WithOutput(w io.Writer) LoggerOption {
    return func(l *Logger) {
        l.output = w
    }
}

func WithPrefix(prefix string) LoggerOption {
    return func(l *Logger) {
        l.prefix = prefix
    }
}
```

## Validation

### Validate After Applying Options

```go
type Config struct {
    workers    int
    bufferSize int
    timeout    time.Duration
}

type ConfigOption func(*Config)

func NewConfig(opts ...ConfigOption) (*Config, error) {
    cfg := &Config{
        workers:    runtime.NumCPU(),
        bufferSize: 100,
        timeout:    30 * time.Second,
    }

    for _, opt := range opts {
        opt(cfg)
    }

    // Validate after applying all options
    if err := cfg.validate(); err != nil {
        return nil, err
    }

    return cfg, nil
}

func (c *Config) validate() error {
    if c.workers < 1 {
        return fmt.Errorf("workers must be >= 1, got %d", c.workers)
    }
    if c.bufferSize < 0 {
        return fmt.Errorf("buffer size must be >= 0, got %d", c.bufferSize)
    }
    if c.timeout < 0 {
        return fmt.Errorf("timeout must be >= 0, got %v", c.timeout)
    }
    return nil
}

func WithWorkers(n int) ConfigOption {
    return func(c *Config) {
        c.workers = n
    }
}

func WithBufferSize(n int) ConfigOption {
    return func(c *Config) {
        c.bufferSize = n
    }
}

func WithConfigTimeout(d time.Duration) ConfigOption {
    return func(c *Config) {
        c.timeout = d
    }
}

// Usage
cfg, err := NewConfig(WithWorkers(4), WithBufferSize(500))
if err != nil {
    log.Fatal(err)
}
```

### Validation in Option

```go
type Pool struct {
    minConns int
    maxConns int
}

type PoolOption func(*Pool) error

func NewPool(opts ...PoolOption) (*Pool, error) {
    p := &Pool{
        minConns: 5,
        maxConns: 100,
    }

    for _, opt := range opts {
        if err := opt(p); err != nil {
            return nil, err
        }
    }

    return p, nil
}

func WithMinConns(n int) PoolOption {
    return func(p *Pool) error {
        if n < 0 {
            return fmt.Errorf("min conns must be >= 0")
        }
        p.minConns = n
        return nil
    }
}

func WithMaxConns(n int) PoolOption {
    return func(p *Pool) error {
        if n < 1 {
            return fmt.Errorf("max conns must be >= 1")
        }
        if n < p.minConns {
            return fmt.Errorf("max conns (%d) must be >= min conns (%d)", n, p.minConns)
        }
        p.maxConns = n
        return nil
    }
}
```

## Advanced Patterns

### Option Groups

```go
type HTTPServer struct {
    addr            string
    readTimeout     time.Duration
    writeTimeout    time.Duration
    maxHeaderBytes  int
    shutdownTimeout time.Duration
}

type ServerOption func(*HTTPServer)

func WithAddr(addr string) ServerOption {
    return func(s *HTTPServer) {
        s.addr = addr
    }
}

func WithTimeouts(read, write time.Duration) ServerOption {
    return func(s *HTTPServer) {
        s.readTimeout = read
        s.writeTimeout = write
    }
}

func WithShutdownTimeout(d time.Duration) ServerOption {
    return func(s *HTTPServer) {
        s.shutdownTimeout = d
    }
}

// Option group for production settings
func ProductionDefaults() ServerOption {
    return func(s *HTTPServer) {
        s.readTimeout = 10 * time.Second
        s.writeTimeout = 10 * time.Second
        s.maxHeaderBytes = 1 << 20 // 1 MB
        s.shutdownTimeout = 30 * time.Second
    }
}

// Option group for development
func DevelopmentDefaults() ServerOption {
    return func(s *HTTPServer) {
        s.readTimeout = 60 * time.Second
        s.writeTimeout = 60 * time.Second
        s.maxHeaderBytes = 1 << 22 // 4 MB
        s.shutdownTimeout = 5 * time.Second
    }
}

// Usage
srv := NewHTTPServer(":8080", ProductionDefaults())
srv := NewHTTPServer(":8080", DevelopmentDefaults(), WithAddr(":9000"))
```

### Composable Options

```go
type Database struct {
    host     string
    port     int
    user     string
    password string
    dbname   string
    sslMode  string
}

type DBOption func(*Database)

func WithHost(host string) DBOption {
    return func(db *Database) {
        db.host = host
    }
}

func WithPort(port int) DBOption {
    return func(db *Database) {
        db.port = port
    }
}

func WithCredentials(user, password string) DBOption {
    return func(db *Database) {
        db.user = user
        db.password = password
    }
}

func WithDatabase(name string) DBOption {
    return func(db *Database) {
        db.dbname = name
    }
}

func WithSSL(mode string) DBOption {
    return func(db *Database) {
        db.sslMode = mode
    }
}

// Composite option
func WithDefaults(host, user, password, dbname string) DBOption {
    return func(db *Database) {
        WithHost(host)(db)
        WithPort(5432)(db)
        WithCredentials(user, password)(db)
        WithDatabase(dbname)(db)
        WithSSL("require")(db)
    }
}

// Usage
db := NewDatabase(
    WithDefaults("localhost", "user", "pass", "mydb"),
    WithPort(5433), // Override default port
)
```

### Fluent Builder Alternative (Comparison)

```go
// Functional options (preferred for libraries)
type ClientOpts struct {
    timeout time.Duration
    retries int
}

type ClientOption func(*ClientOpts)

func WithClientTimeoutOpt(d time.Duration) ClientOption {
    return func(o *ClientOpts) {
        o.timeout = d
    }
}

func NewHTTPClient(opts ...ClientOption) *Client {
    o := &ClientOpts{
        timeout: 30 * time.Second,
        retries: 3,
    }
    for _, opt := range opts {
        opt(o)
    }
    return &Client{opts: o}
}

// Usage: immutable after creation
client := NewHTTPClient(WithClientTimeoutOpt(60*time.Second))

// Fluent builder (for complex config)
type ClientBuilder struct {
    timeout time.Duration
    retries int
}

func NewClientBuilder() *ClientBuilder {
    return &ClientBuilder{
        timeout: 30 * time.Second,
        retries: 3,
    }
}

func (b *ClientBuilder) Timeout(d time.Duration) *ClientBuilder {
    b.timeout = d
    return b
}

func (b *ClientBuilder) Retries(n int) *ClientBuilder {
    b.retries = n
    return b
}

func (b *ClientBuilder) Build() (*Client, error) {
    if b.timeout < 0 {
        return nil, fmt.Errorf("invalid timeout")
    }
    return &Client{timeout: b.timeout, retries: b.retries}, nil
}

// Usage: mutable during construction
client, err := NewClientBuilder().
    Timeout(60*time.Second).
    Retries(5).
    Build()
```

## When to Use

### Use Functional Options When

- **Library API**: Public API with optional configuration
- **Many Options**: 3+ optional parameters
- **Backward Compatibility**: Need to add options without breaking API
- **Sensible Defaults**: Most users need only subset of options
- **Immutable Objects**: Options set once at construction

```go
// Good fit: many optional parameters
type Cache struct {
    ttl         time.Duration
    maxSize     int
    eviction    string
    compression bool
    metrics     bool
}

func NewCache(opts ...CacheOption) *Cache {
    // 5 optional parameters - functional options shine here
}
```

### Use Simple Struct When

- **Internal API**: Package-private constructors
- **Few Options**: 1-2 optional parameters
- **No Defaults**: All fields must be provided
- **Mutable**: Configuration changes after creation

```go
// Simple struct better: few required fields, internal use
type Request struct {
    Method string
    URL    string
    Body   io.Reader
}

// No need for options
func NewRequest(method, url string, body io.Reader) *Request {
    return &Request{Method: method, URL: url, Body: body}
}
```

### Use Config Struct When

- **Application Config**: Loading from env/file
- **Complex Nested**: Multiple configuration sections
- **Validation Needed**: Extensive cross-field validation

```go
// Config struct better: loaded from environment
type AppConfig struct {
    Server   ServerConfig
    Database DatabaseConfig
    Cache    CacheConfig
}

func LoadConfig() (*AppConfig, error) {
    // Load from env, validate, return
}
```

## Common Mistakes

| Mistake                        | Why It's Bad                         | Fix                                    |
|--------------------------------|--------------------------------------|----------------------------------------|
| Options that panic             | Crashes at runtime                   | Return error from constructor          |
| No validation                  | Invalid state silently accepted      | Validate after applying options        |
| Exposing config struct         | Breaks encapsulation                 | Keep struct private                    |
| Inconsistent naming            | `WithX`, `SetY`, `EnableZ` mixed     | Use `With*` prefix consistently        |
| Mutating options post-creation | Breaks immutability                  | Apply options only in constructor      |
| Options modifying other options| Order-dependent behavior             | Keep options independent               |
| Complex option logic           | Hard to understand and test          | Keep option functions simple           |

### Bad: Options That Panic

```go
// Bad
func WithPort(port int) Option {
    return func(s *Server) {
        if port < 1 || port > 65535 {
            panic("invalid port") // Never panic in library
        }
        s.port = port
    }
}

// Good
func WithPort(port int) Option {
    return func(s *Server) {
        if port >= 1 && port <= 65535 {
            s.port = port
        }
        // Invalid values silently ignored, validation in New()
    }
}

// Better: validate in constructor
func New(opts ...Option) (*Server, error) {
    s := &Server{port: 8080}
    for _, opt := range opts {
        opt(s)
    }
    if err := s.validate(); err != nil {
        return nil, err
    }
    return s, nil
}
```

### Bad: No Validation

```go
// Bad
func NewPool(opts ...PoolOption) *Pool {
    p := &Pool{maxConns: 100, minConns: 5}
    for _, opt := range opts {
        opt(p)
    }
    return p // What if minConns > maxConns?
}

// Good
func NewPool(opts ...PoolOption) (*Pool, error) {
    p := &Pool{maxConns: 100, minConns: 5}
    for _, opt := range opts {
        opt(p)
    }
    if p.minConns > p.maxConns {
        return nil, fmt.Errorf("minConns (%d) > maxConns (%d)", p.minConns, p.maxConns)
    }
    return p, nil
}
```

### Bad: Exposing Config Struct

```go
// Bad: exposes internal structure
type ServerConfig struct {
    Timeout time.Duration // Public
    MaxConn int          // Public
}

func New(cfg *ServerConfig) *Server {
    // Users can mutate cfg after creation
}

// Good: encapsulated
type server struct {
    timeout time.Duration
    maxConn int
}

type Option func(*server)

func New(opts ...Option) *server {
    // Internal state hidden
}
```

## See Also

- [Configuration](./configuration.md) - Environment-based config
- [Cobra](./cobra.md) - CLI applications
- [Idioms](../01-fundamentals/idioms.md) - Go design patterns
- [Error Handling](../02-language/error-handling.md) - Validation errors
- [Dave Cheney's Blog](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis) - Original pattern description
