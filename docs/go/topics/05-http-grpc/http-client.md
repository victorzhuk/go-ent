# HTTP Client

## Overview

Go's `net/http` package provides a robust, production-ready HTTP client with sensible defaults. The stdlib client supports connection pooling, keep-alive, compression, and timeout management out of the box.

**When to use `http.DefaultClient`:**
- Quick scripts and tools
- Internal services with default timeouts acceptable
- Non-production code

**When to use custom `http.Client`:**
- Production services (always set timeouts)
- Fine-tuned connection pool settings
- Custom TLS configuration
- Retry logic requirements
- Request/response middleware

**Key principle:** Never use `http.DefaultClient` in production—it has no timeout, risking indefinite hangs.

## Quick Reference

| Task | Code |
|------|------|
| Simple GET | `resp, err := http.Get(url)` |
| GET with context | `req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)`<br>`resp, err := client.Do(req)` |
| POST JSON | `resp, err := http.Post(url, "application/json", bytes.NewReader(data))` |
| Custom client (timeout) | `client := &http.Client{Timeout: 10 * time.Second}` |
| Set headers | `req.Header.Set("Authorization", "Bearer "+token)` |
| Query params | `u.Query().Set("page", "1")` |
| Custom transport | `client := &http.Client{Transport: customTransport}` |
| Close response body | `defer resp.Body.Close()` (always after error check) |
| Read body | `body, err := io.ReadAll(resp.Body)` |
| Context deadline | `ctx, cancel := context.WithTimeout(ctx, 5*time.Second)` |

## Basic Requests

### GET Request

```go
func fetchUser(ctx context.Context, client *http.Client, userID string) (*User, error) {
	url := fmt.Sprintf("https://api.example.com/users/%s", userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &user, nil
}
```

### POST Request with JSON

```go
func createUser(ctx context.Context, client *http.Client, user *User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("marshal user: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.example.com/users", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create failed with status %d: %s", resp.StatusCode, body)
	}

	return nil
}
```

### Query Parameters

```go
func searchUsers(ctx context.Context, client *http.Client, query string, page int) ([]User, error) {
	u, err := url.Parse("https://api.example.com/users/search")
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	q := u.Query()
	q.Set("q", query)
	q.Set("page", strconv.Itoa(page))
	q.Set("limit", "50")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return users, nil
}
```

## Custom Client Configuration

### Production Client with Timeouts

```go
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			MaxConnsPerHost:       100,
			IdleConnTimeout:       90 * time.Second,
		},
	}
}
```

**Timeout breakdown:**
- `Timeout`: Total time for entire request/response cycle
- `DialContext.Timeout`: TCP connection establishment
- `TLSHandshakeTimeout`: TLS handshake completion
- `ResponseHeaderTimeout`: Waiting for response headers
- `ExpectContinueTimeout`: Waiting for 100-Continue response

### Custom TLS Configuration

```go
func newTLSClient(certFile, keyFile, caFile string) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load keypair: %w", err)
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read ca cert: %w", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS13,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}, nil
}
```

### Connection Pool Tuning

```go
// For high-throughput scenarios
func newHighThroughputClient() *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 50,
			MaxConnsPerHost:     100,
			IdleConnTimeout:     60 * time.Second,
			DisableKeepAlives:   false,
			DisableCompression:  false,
		},
	}
}

// For low-latency scenarios
func newLowLatencyClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   2,
			IdleConnTimeout:       30 * time.Second,
			ResponseHeaderTimeout: 2 * time.Second,
		},
	}
}
```

## Context and Cancellation

### Request with Timeout

```go
func fetchWithTimeout(baseURL string, userID string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := &http.Client{}

	url := fmt.Sprintf("%s/users/%s", baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("request timeout after 5s")
		}
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &user, nil
}
```

### Propagating Request Context

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Propagate incoming request context to downstream calls
	ctx := r.Context()

	users, err := fetchUsers(ctx, httpClient)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			http.Error(w, "request canceled", http.StatusRequestTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func fetchUsers(ctx context.Context, client *http.Client) ([]User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.example.com/users", nil)
	if err != nil {
		return nil, err
	}

	// Request inherits cancellation from parent context
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users, nil
}
```

## Retries and Backoff

### Exponential Backoff with Jitter

```go
func doWithRetry(ctx context.Context, client *http.Client, req *http.Request, maxRetries int) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			delay := backoff + jitter

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// Clone request for retry (body might be consumed)
		reqClone := req.Clone(ctx)

		resp, err = client.Do(reqClone)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				continue
			}
			return nil, err
		}

		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			resp.Body.Close()
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", err)
}
```

### Retry with Retry-After Header

```go
func doWithRetryAfter(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 429 {
		return resp, nil
	}

	resp.Body.Close()

	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return nil, fmt.Errorf("rate limited with no retry-after header")
	}

	delay, err := time.ParseDuration(retryAfter + "s")
	if err != nil {
		seconds, parseErr := strconv.Atoi(retryAfter)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid retry-after: %s", retryAfter)
		}
		delay = time.Duration(seconds) * time.Second
	}

	select {
	case <-time.After(delay):
		return client.Do(req.Clone(ctx))
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
```

### Idempotency Token Pattern

```go
func createOrderWithRetry(ctx context.Context, client *http.Client, order *Order) error {
	idempotencyKey := uuid.Must(uuid.NewV7()).String()

	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("marshal order: %w", err)
	}

	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			"https://api.example.com/orders", bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			return nil
		}

		if resp.StatusCode >= 500 {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create order failed: %d %s", resp.StatusCode, body)
	}

	return fmt.Errorf("max retries exceeded")
}
```

## Error Handling

### Network Error Classification

```go
func classifyError(err error) string {
	if err == nil {
		return "success"
	}

	if errors.Is(err, context.Canceled) {
		return "canceled"
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return "network_timeout"
		}
		return "network_error"
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "dns_error"
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Sprintf("url_error_%s", urlErr.Op)
	}

	return "unknown"
}

func handleHTTPError(resp *http.Response, err error) error {
	if err != nil {
		errType := classifyError(err)
		return fmt.Errorf("request failed (%s): %w", errType, err)
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http error %d: %s", resp.StatusCode, body)
	}

	return nil
}
```

### Status Code Handling

```go
func callAPI(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusNotFound:
		return nil, fmt.Errorf("resource not found")
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized")
	case http.StatusForbidden:
		return nil, fmt.Errorf("forbidden")
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("rate limited")
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return nil, fmt.Errorf("server error %d: %s", resp.StatusCode, body)
	default:
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}
}
```

### Proper Resource Cleanup

```go
func fetchSafe(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body) // Drain body before close
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return body, nil
}
```

## Testing

### Using httptest.Server

```go
func TestFetchUser(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/123", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(User{ID: "123", Name: "Test User"})
	}))
	defer server.Close()

	client := server.Client()

	user, err := fetchUser(context.Background(), client, "123")
	assert.NoError(t, err)
	assert.Equal(t, "123", user.ID)
	assert.Equal(t, "Test User", user.Name)
}
```

### Testing Error Scenarios

```go
func TestFetchUser_NetworkError(t *testing.T) {
	t.Parallel()

	// Server that closes connection immediately
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("server does not support hijacking")
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}))
	defer server.Close()

	client := server.Client()

	_, err := fetchUser(context.Background(), client, "123")
	assert.Error(t, err)
}

func TestFetchUser_Timeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 100 * time.Millisecond}

	_, err := fetchUser(context.Background(), client, "123")
	assert.Error(t, err)

	var netErr net.Error
	assert.True(t, errors.As(err, &netErr))
	assert.True(t, netErr.Timeout())
}
```

### Custom RoundTripper for Mocking

```go
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFetchUser_MockTransport(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "https://api.example.com/users/123", req.URL.String())

			user := User{ID: "123", Name: "Mock User"}
			data, _ := json.Marshal(user)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(data)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	user, err := fetchUser(context.Background(), client, "123")
	assert.NoError(t, err)
	assert.Equal(t, "Mock User", user.Name)
}
```

### Recording HTTP Interactions

```go
type recordingTransport struct {
	transport http.RoundTripper
	requests  []*http.Request
	responses []*http.Response
}

func (rt *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.requests = append(rt.requests, req)

	resp, err := rt.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	rt.responses = append(rt.responses, resp)
	return resp, nil
}

func TestMultipleRequests(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		transport: http.DefaultTransport,
	}

	client := &http.Client{Transport: recorder}

	// Make requests...

	assert.Len(t, recorder.requests, 3)
	assert.Len(t, recorder.responses, 3)
}
```

## Common Mistakes

| Mistake | Impact | Fix |
|---------|--------|-----|
| Not closing response body | Connection pool exhaustion | Always `defer resp.Body.Close()` after error check |
| Using `http.DefaultClient` | Infinite hangs on timeout | Create custom client with `Timeout` set |
| Ignoring status codes | Silent failures | Check `resp.StatusCode` before reading body |
| Not using context | Cannot cancel long requests | Use `http.NewRequestWithContext(ctx, ...)` |
| Retrying non-idempotent requests | Duplicate operations (double charge) | Only retry GET/HEAD/PUT/DELETE or use idempotency tokens |
| Reading body without limit | Memory exhaustion from large response | Use `io.LimitReader(resp.Body, maxBytes)` |
| Reusing request body | Empty body on retry | Clone request or store body bytes for retry |
| Not draining body before close | Connection not reused | `io.Copy(io.Discard, resp.Body)` before close |
| Connection pool leaks | File descriptor exhaustion | Set `MaxIdleConns` and `IdleConnTimeout` |
| Infinite retry loops | Service degradation | Always set max retry count |
| Not handling `context.Canceled` | Confusing error messages | Check `errors.Is(err, context.Canceled)` |
| Missing `Content-Type` header | Server rejects request | Set `req.Header.Set("Content-Type", "application/json")` |

## See Also

- [HTTP Server](http-server.md) - Building HTTP servers with net/http
- [OpenAPI](openapi.md) - Type-safe API clients with code generation
- [Tracing](../07-observability/tracing.md) - Distributed tracing for HTTP requests
- [Correlation](../07-observability/correlation.md) - Request ID propagation
- [Error Handling](../02-language/error-handling.md) - Error wrapping and classification
