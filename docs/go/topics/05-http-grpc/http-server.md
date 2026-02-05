# HTTP Server

Production HTTP servers using Go 1.22+ `net/http` with enhanced routing and graceful shutdown.

## Quick Reference

| Pattern            | Use When                               |
|--------------------|----------------------------------------|
| `http.HandleFunc`  | Simple routes                          |
| `http.NewServeMux` | Go 1.22+ routing with wildcards        |
| Middleware         | Cross-cutting concerns (logging, auth) |
| `http.Server`      | Production config (timeouts, TLS)      |
| Graceful shutdown  | Handle SIGTERM/SIGINT                  |

## Basic Server (Go 1.22+)

```go
func main() {
    mux := http.NewServeMux()
   
    // Simple routes
    mux.HandleFunc("GET /users", listUsers)
    mux.HandleFunc("GET /users/{id}", getUser)
    mux.HandleFunc("POST /users", createUser)
   
    // Wildcard
    mux.HandleFunc("GET /files/{path...}", serveFile)
   
    log.Fatal(http.ListenAndServe(":8080", mux))
}

func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // Get path parameter
    // ...
}
```

## Production Server

```go
func main() {
    srv := &http.Server{
        Addr:         ":8080",
        Handler:      router(),
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Middleware

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}
       
        next.ServeHTTP(wrapped, r)
       
        log.Printf("%s %s %d %v", r.Method, r.URL.Path,
            wrapped.statusCode, time.Since(start))
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
    w.statusCode = code
    w.ResponseWriter.WriteHeader(code)
}

// Usage
handler := loggingMiddleware(authMiddleware(mux))
```

## JSON API

```go
func respondJSON(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, message string) {
    respondJSON(w, status, map[string]string{"error": message})
}

func createUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        respondError(w, http.StatusBadRequest, "invalid request")
        return
    }

    if err := user.Validate(); err != nil {
        respondError(w, http.StatusBadRequest, err.Error())
        return
    }

    respondJSON(w, http.StatusCreated, user)
}
```

## See Also

- [HTTP Client](./http-client.md)
- [gRPC](./grpc.md)
- [OpenAPI](./openapi.md)
