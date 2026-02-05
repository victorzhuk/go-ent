# gRPC

gRPC services with interceptors, keepalive, and OpenTelemetry.

## Quick Reference

| Feature | Server | Client |
|---------|--------|--------|
| **Basic** | `grpc.NewServer()` | `grpc.Dial(addr, opts...)` |
| **Unary** | `func(ctx, *Req) (*Resp, error)` | `client.Method(ctx, req)` |
| **Server Streaming** | `func(req, stream) error` | `stream, _ := client.Method(ctx, req)` |
| **Client Streaming** | `func(stream) error` | `stream, _ := client.Method(ctx)` |
| **Bidirectional** | `func(stream) error` | `stream, _ := client.Method(ctx)` |
| **Error Codes** | `status.Error(codes.NotFound, msg)` | `st, ok := status.FromError(err)` |
| **TLS** | `grpc.Creds(creds)` | `grpc.WithTransportCredentials(creds)` |
| **Health** | `grpc_health_v1.RegisterHealthServer` | `grpc_health_v1.NewHealthClient(conn)` |
| **Tracing** | `grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor())` | `grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor())` |

```go
import "google.golang.org/grpc"

// Server
s := grpc.NewServer()
pb.RegisterUserServiceServer(s, &userService{})
s.Serve(lis)

// Client
conn, _ := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
client := pb.NewUserServiceClient(conn)
```

## Server Setup

```go
func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatal(err)
    }

    s := grpc.NewServer(
        grpc.UnaryInterceptor(loggingInterceptor),
        grpc.KeepaliveParams(keepalive.ServerParameters{
            Time:    5 * time.Minute,
            Timeout: 10 * time.Second,
        }),
    )

    pb.RegisterUserServiceServer(s, &userService{})

    log.Fatal(s.Serve(lis))
}
```

## Interceptors (Middleware)

```go
func loggingInterceptor(ctx context.Context, req interface{},
    info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

    start := time.Now()
    resp, err := handler(ctx, req)
    duration := time.Since(start)

    log.Printf("method=%s duration=%v error=%v", info.FullMethod, duration, err)
    return resp, err
}
```

## Streaming

### Server Streaming

Server sends multiple responses for single request.

```go
func (s *userService) ListUsers(req *pb.ListUsersRequest, stream pb.UserService_ListUsersServer) error {
    users, err := s.repo.FindAll(stream.Context())
    if err != nil {
        return status.Errorf(codes.Internal, "query users: %v", err)
    }

    for _, u := range users {
        if err := stream.Send(&pb.User{Id: u.ID, Email: u.Email}); err != nil {
            return status.Errorf(codes.Internal, "send user: %v", err)
        }
    }

    return nil
}
```

**Client:**
```go
stream, err := client.ListUsers(ctx, &pb.ListUsersRequest{})
if err != nil {
    return fmt.Errorf("list users: %w", err)
}

for {
    user, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        return fmt.Errorf("receive user: %w", err)
    }
    fmt.Printf("User: %s\n", user.Email)
}
```

### Client Streaming

Client sends multiple requests, server responds once.

```go
func (s *userService) CreateUsers(stream pb.UserService_CreateUsersServer) error {
    var count int32

    for {
        req, err := stream.Recv()
        if err == io.EOF {
            return stream.SendAndClose(&pb.CreateUsersResponse{Count: count})
        }
        if err != nil {
            return status.Errorf(codes.Internal, "receive user: %v", err)
        }

        if err := s.repo.Create(stream.Context(), req.User); err != nil {
            return status.Errorf(codes.Internal, "create user: %v", err)
        }
        count++
    }
}
```

**Client:**
```go
stream, err := client.CreateUsers(ctx)
if err != nil {
    return fmt.Errorf("create users: %w", err)
}

for _, u := range users {
    if err := stream.Send(&pb.CreateUserRequest{User: u}); err != nil {
        return fmt.Errorf("send user: %w", err)
    }
}

resp, err := stream.CloseAndRecv()
if err != nil {
    return fmt.Errorf("close stream: %w", err)
}
fmt.Printf("Created %d users\n", resp.Count)
```

### Bidirectional Streaming

Both sides send messages independently.

```go
func (s *userService) Chat(stream pb.UserService_ChatServer) error {
    for {
        msg, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return status.Errorf(codes.Internal, "receive message: %v", err)
        }

        reply := &pb.ChatMessage{Text: fmt.Sprintf("Echo: %s", msg.Text)}
        if err := stream.Send(reply); err != nil {
            return status.Errorf(codes.Internal, "send reply: %v", err)
        }
    }
}
```

**Client:**
```go
stream, err := client.Chat(ctx)
if err != nil {
    return fmt.Errorf("chat: %w", err)
}

go func() {
    for _, msg := range messages {
        if err := stream.Send(msg); err != nil {
            log.Printf("send: %v", err)
            return
        }
    }
    stream.CloseSend()
}()

for {
    msg, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        return fmt.Errorf("receive: %w", err)
    }
    fmt.Printf("Received: %s\n", msg.Text)
}
```

## Error Handling

Use `google.golang.org/grpc/status` and `codes` for structured errors.

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func (s *userService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "user id required")
    }

    user, err := s.repo.FindByID(ctx, req.Id)
    if err != nil {
        if errors.Is(err, domain.ErrNotFound) {
            return nil, status.Errorf(codes.NotFound, "user %s not found", req.Id)
        }
        return nil, status.Errorf(codes.Internal, "query user: %v", err)
    }

    return &pb.User{Id: user.ID, Email: user.Email}, nil
}
```

**Common Status Codes:**

| Code | When to Use |
|------|-------------|
| `codes.OK` | Success (default, don't return explicitly) |
| `codes.Canceled` | Request canceled (context done) |
| `codes.InvalidArgument` | Invalid input (validation failed) |
| `codes.NotFound` | Resource doesn't exist |
| `codes.AlreadyExists` | Duplicate resource |
| `codes.PermissionDenied` | No permission for operation |
| `codes.Unauthenticated` | Missing or invalid credentials |
| `codes.ResourceExhausted` | Rate limit, quota exceeded |
| `codes.Internal` | Server error (don't expose details) |
| `codes.Unavailable` | Service temporarily unavailable |
| `codes.DeadlineExceeded` | Timeout |

**Rich Error Details:**

```go
import (
    "google.golang.org/genproto/googleapis/rpc/errdetails"
)

func (s *userService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
    if req.Email == "" {
        st := status.New(codes.InvalidArgument, "validation failed")
        v := &errdetails.BadRequest_FieldViolation{
            Field:       "email",
            Description: "email is required",
        }
        br := &errdetails.BadRequest{FieldViolations: []*errdetails.BadRequest_FieldViolation{v}}
        st, _ = st.WithDetails(br)
        return nil, st.Err()
    }

    // create user...
    return user, nil
}
```

**Client Error Extraction:**

```go
resp, err := client.GetUser(ctx, &pb.GetUserRequest{Id: "123"})
if err != nil {
    st, ok := status.FromError(err)
    if !ok {
        return fmt.Errorf("unexpected error: %w", err)
    }

    switch st.Code() {
    case codes.NotFound:
        return fmt.Errorf("user not found")
    case codes.InvalidArgument:
        return fmt.Errorf("invalid input: %s", st.Message())
    default:
        return fmt.Errorf("grpc error: %s", st.Message())
    }
}
```

## TLS Configuration

### Server with TLS

```go
import (
    "crypto/tls"
    "google.golang.org/grpc/credentials"
)

func newServer() (*grpc.Server, error) {
    cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
    if err != nil {
        return nil, fmt.Errorf("load keypair: %w", err)
    }

    creds := credentials.NewTLS(&tls.Config{
        Certificates: []tls.Certificate{cert},
        MinVersion:   tls.VersionTLS13,
    })

    s := grpc.NewServer(grpc.Creds(creds))
    return s, nil
}
```

### Mutual TLS (mTLS)

```go
import (
    "crypto/x509"
    "os"
)

func newServerMTLS() (*grpc.Server, error) {
    cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
    if err != nil {
        return nil, fmt.Errorf("load keypair: %w", err)
    }

    ca, err := os.ReadFile("ca.crt")
    if err != nil {
        return nil, fmt.Errorf("read ca cert: %w", err)
    }

    certPool := x509.NewCertPool()
    if !certPool.AppendCertsFromPEM(ca) {
        return nil, fmt.Errorf("append ca cert")
    }

    creds := credentials.NewTLS(&tls.Config{
        Certificates: []tls.Certificate{cert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    certPool,
        MinVersion:   tls.VersionTLS13,
    })

    s := grpc.NewServer(grpc.Creds(creds))
    return s, nil
}
```

**Client with TLS:**

```go
func newClient(addr string) (*grpc.ClientConn, error) {
    creds := credentials.NewClientTLSFromCert(nil, "")

    conn, err := grpc.Dial(addr,
        grpc.WithTransportCredentials(creds),
        grpc.WithBlock(),
    )
    if err != nil {
        return nil, fmt.Errorf("dial: %w", err)
    }

    return conn, nil
}
```

## OpenTelemetry

```go
import (
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
    "go.opentelemetry.io/otel"
)

func newServerWithTracing() *grpc.Server {
    s := grpc.NewServer(
        grpc.StatsHandler(otelgrpc.NewServerHandler()),
    )
    return s
}

func newClientWithTracing(addr string) (*grpc.ClientConn, error) {
    conn, err := grpc.Dial(addr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
    )
    if err != nil {
        return nil, fmt.Errorf("dial: %w", err)
    }

    return conn, nil
}
```

**Manual Span Creation:**

```go
import (
    "go.opentelemetry.io/otel/trace"
)

func (s *userService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    tracer := otel.Tracer("user-service")
    ctx, span := tracer.Start(ctx, "repo.FindByID")
    defer span.End()

    user, err := s.repo.FindByID(ctx, req.Id)
    if err != nil {
        span.RecordError(err)
        return nil, status.Errorf(codes.Internal, "query user: %v", err)
    }

    return toProto(user), nil
}
```

## Health Service

```go
import (
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
    s := grpc.NewServer()

    healthSrv := health.NewServer()
    grpc_health_v1.RegisterHealthServer(s, healthSrv)

    pb.RegisterUserServiceServer(s, &userService{})

    healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
    healthSrv.SetServingStatus("user.UserService", grpc_health_v1.HealthCheckResponse_SERVING)

    log.Fatal(s.Serve(lis))
}
```

**Custom Health Check:**

```go
type healthChecker struct {
    pool *pgxpool.Pool
}

func (h *healthChecker) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
    if err := h.pool.Ping(ctx); err != nil {
        return &grpc_health_v1.HealthCheckResponse{
            Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
        }, nil
    }

    return &grpc_health_v1.HealthCheckResponse{
        Status: grpc_health_v1.HealthCheckResponse_SERVING,
    }, nil
}

func (h *healthChecker) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
    return status.Error(codes.Unimplemented, "watch not implemented")
}
```

## Common Mistakes

| Mistake | Why It's Bad | Fix |
|---------|--------------|-----|
| **Not checking stream.Recv() error** | Client disconnect causes infinite loop | Check for `io.EOF` and other errors |
| **Blocking in interceptors** | Delays all requests, causes cascading failures | Use goroutines for async work |
| **Missing context deadline** | Requests hang forever on timeout | Always set `context.WithTimeout` |
| **Wrong error codes** | Clients can't handle errors properly | Use semantic codes (NotFound, InvalidArgument) |
| **Returning generic errors** | Leaks implementation details | Map domain errors to gRPC codes |
| **Not closing streams** | Resource leak, connection buildup | Always `defer stream.CloseSend()` (client) |
| **Using insecure creds in prod** | Man-in-the-middle attacks | Use TLS credentials |
| **Ignoring context cancellation** | Wastes resources on canceled requests | Check `ctx.Err()` in loops |

## See Also

- [HTTP Server](./http-server.md)
- [OpenAPI](./openapi.md)
- [Tracing](../07-observability/tracing.md)
- [Correlation IDs](../07-observability/correlation.md)
