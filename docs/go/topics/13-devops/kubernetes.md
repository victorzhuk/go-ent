# Kubernetes

K8s deployment patterns with health checks and automaxprocs.

## Quick Reference

| Component             | Purpose                                    | Key Settings                          |
|-----------------------|-----------------------------------------------|---------------------------------------|
| `automaxprocs`        | Set GOMAXPROCS from CPU quota                 | Import side-effect only               |
| Resources             | CPU/memory requests and limits                | `requests` for scheduling, `limits` for caps |
| Probes                | Health monitoring                             | `livenessProbe`, `readinessProbe`, `startupProbe` |
| PodDisruptionBudget   | Prevent simultaneous pod termination          | `minAvailable` or `maxUnavailable`    |
| Lifecycle Hooks       | Graceful shutdown and initialization          | `preStop`, `postStart`                |
| `/healthz`            | Liveness endpoint (app alive)                 | Should return 200 if app running      |
| `/readyz`             | Readiness endpoint (app ready to serve)       | Check dependencies (DB, cache)        |

```go
import _ "go.uber.org/automaxprocs"

func main() {
    // automaxprocs sets GOMAXPROCS based on CPU quota
    // ...
}
```

## Health Checks

```go
func healthz(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
}

func readyz(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := db.Ping(); err != nil {
            http.Error(w, "not ready", http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ready"))
    }
}

// In main:
http.HandleFunc("/healthz", healthz)
http.HandleFunc("/readyz", readyz(db))
```

## Graceful Shutdown

```go
func main() {
    srv := &http.Server{Addr: ":8080", Handler: mux}

    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Resources

Set resource requests and limits to ensure proper scheduling and prevent resource exhaustion.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: api
        image: myapp:1.2.3
        resources:
          requests:
            cpu: 100m      # Guaranteed CPU for scheduling
            memory: 128Mi  # Guaranteed memory for scheduling
          limits:
            cpu: 500m      # Maximum CPU (throttled if exceeded)
            memory: 512Mi  # Maximum memory (OOMKilled if exceeded)
```

**CPU Sizing:**
- `100m` = 0.1 CPU core (100 millicores)
- Start with requests=50m-100m, limits=500m-1000m
- Monitor actual usage with metrics
- CPU limits cause throttling, not termination

**Memory Sizing:**
- Start with requests=128Mi, limits=512Mi
- Memory limits cause OOMKill if exceeded
- Leave headroom: `limits = requests * 2-4`
- Monitor with `kubectl top pods`

**Namespace Limits:**

```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: default-limits
spec:
  limits:
  - default:
      cpu: 500m
      memory: 512Mi
    defaultRequest:
      cpu: 100m
      memory: 128Mi
    type: Container
---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: namespace-quota
spec:
  hard:
    requests.cpu: "10"
    requests.memory: 20Gi
    limits.cpu: "20"
    limits.memory: 40Gi
    pods: "50"
```

## PodDisruptionBudget

Prevent simultaneous pod termination during voluntary disruptions (node drain, cluster upgrade).

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: api-server-pdb
spec:
  minAvailable: 2  # Always keep 2 pods running
  selector:
    matchLabels:
      app: api-server
---
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: worker-pdb
spec:
  maxUnavailable: 1  # Only terminate 1 pod at a time
  selector:
    matchLabels:
      app: worker
```

**Strategies:**
- `minAvailable: 2` — Always keep 2+ pods (for replicas=3, allows 1 drain)
- `maxUnavailable: 1` — Only 1 pod down at a time (safer for rolling updates)
- Use percentage: `minAvailable: "50%"`
- Requires `replicas >= minAvailable + 1`

## Lifecycle Hooks

Graceful shutdown with `preStop` and initialization with `postStart`.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
spec:
  template:
    spec:
      containers:
      - name: api
        image: myapp:1.2.3
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 5"]  # Wait for LB deregistration
          postStart:
            exec:
              command: ["/bin/sh", "-c", "/app/warmup.sh"]
        terminationGracePeriodSeconds: 30
```

**preStop Workflow:**
1. Pod marked for termination
2. Removed from Service endpoints
3. `preStop` hook executes (blocks for 5s)
4. SIGTERM sent to main process
5. App gracefully shuts down (30s timeout)
6. SIGKILL if still running after grace period

**Go App with preStop:**

```go
func main() {
    srv := &http.Server{Addr: ":8080", Handler: mux}

    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
    <-quit

    // Fresh context for shutdown (parent context cancelled)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal(err)
    }
}
```

**Timing:**
- `preStop sleep` = 5s (LB deregistration propagation)
- `terminationGracePeriodSeconds` = 30s (app shutdown time)
- Total shutdown window = 35s

## Probes Deep Dive

Three probe types with different purposes and failure handling.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
spec:
  template:
    spec:
      containers:
      - name: api
        image: myapp:1.2.3
        ports:
        - containerPort: 8080
        startupProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 0
          periodSeconds: 1
          failureThreshold: 30  # 30s startup window
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 0
          periodSeconds: 10
          failureThreshold: 3   # 30s to respond or restart
          timeoutSeconds: 5
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 0
          periodSeconds: 5
          failureThreshold: 2   # 10s to respond or remove from LB
          successThreshold: 1
          timeoutSeconds: 3
```

**Probe Types:**

| Probe      | Purpose                        | On Failure                  | Check What                |
|------------|--------------------------------|-----------------------------|---------------------------|
| `startup`  | App finished starting          | Restart pod                 | App process running       |
| `liveness` | App still alive (not deadlocked)| Restart pod                | App not deadlocked        |
| `readiness`| App ready to serve traffic     | Remove from Service         | Dependencies healthy (DB) |

**Probe Settings:**

```go
// /healthz — Liveness (app alive)
func healthz(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
}

// /readyz — Readiness (dependencies healthy)
func readyz(db *sql.DB, cache *redis.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
        defer cancel()

        // Check DB
        if err := db.PingContext(ctx); err != nil {
            http.Error(w, "db unhealthy", http.StatusServiceUnavailable)
            return
        }

        // Check cache
        if err := cache.Ping(ctx).Err(); err != nil {
            http.Error(w, "cache unhealthy", http.StatusServiceUnavailable)
            return
        }

        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ready"))
    }
}

// In main:
http.HandleFunc("/healthz", healthz)
http.HandleFunc("/readyz", readyz(db, cache))
```

**Timing Calculations:**
- Startup: `periodSeconds * failureThreshold = 1 * 30 = 30s` to start
- Liveness: `periodSeconds * failureThreshold = 10 * 3 = 30s` to respond
- Readiness: `periodSeconds * failureThreshold = 5 * 2 = 10s` to respond

**Best Practices:**
- Use `startupProbe` for slow-starting apps (avoid long `initialDelaySeconds`)
- Liveness = simple check (app process alive)
- Readiness = dependency checks (DB, cache, downstream services)
- `timeoutSeconds` < database query timeout
- Never check external services in liveness (causes cascading failures)

## Common Mistakes

| Mistake                          | Impact                                  | Fix                                      |
|----------------------------------|-----------------------------------------|------------------------------------------|
| No resources defined             | Pods scheduled on overloaded nodes      | Set `requests` and `limits`              |
| Same liveness and readiness      | DB down = pod restart loop              | Liveness=simple, readiness=dependencies  |
| Missing PodDisruptionBudget      | All pods terminated during drain        | Add PDB with `minAvailable`              |
| No `preStop` hook                | In-flight requests dropped              | Add 5s sleep in `preStop`                |
| Wrong probe path                 | 404 = pod restart loop                  | Use `/healthz` and `/readyz`             |
| Liveness checks DB               | DB outage kills all pods                | Only check DB in readiness               |
| `limits` = `requests`            | No burst capacity                       | `limits = requests * 2-4`                |
| Long `initialDelaySeconds`       | Slow rolling updates                    | Use `startupProbe` instead               |
| Short `failureThreshold`         | Flaky network = restart                 | Use 3+ failures for liveness             |
| No `terminationGracePeriodSeconds`| SIGKILL before graceful shutdown       | Set to 30s minimum                       |

## Production Deployment

Complete production-ready manifest:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
  labels:
    app: api-server
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: api-server
  template:
    metadata:
      labels:
        app: api-server
    spec:
      terminationGracePeriodSeconds: 30
      containers:
      - name: api
        image: myapp:1.2.3
        ports:
        - name: http
          containerPort: 8080
        - name: metrics
          containerPort: 9090
        env:
        - name: GOMAXPROCS
          valueFrom:
            resourceFieldRef:
              resource: limits.cpu
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 5"]
        startupProbe:
          httpGet:
            path: /healthz
            port: 8080
          periodSeconds: 1
          failureThreshold: 30
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          periodSeconds: 10
          failureThreshold: 3
          timeoutSeconds: 5
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          periodSeconds: 5
          failureThreshold: 2
          timeoutSeconds: 3
---
apiVersion: v1
kind: Service
metadata:
  name: api-server
spec:
  selector:
    app: api-server
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
---
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: api-server-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: api-server
```

## See Also

- [Docker](./docker.md) — Container images and multi-stage builds
- [CI/CD](./ci-cd.md) — Build and deployment pipelines
- [Migrations](../04-database/migrations.md) — Database migrations in K8s
- [HTTP Server](../05-http-grpc/http-server.md) — Health check handlers
