## Profiling First
- ALWAYS profile before optimizing — measure, don't guess
- Use language-specific profilers: Go pprof, Python cProfile, Node clinic
- Identify bottlenecks: CPU, memory, I/O, network
- Focus on the hottest code paths (Pareto: 20% of code = 80% of time)
- Compare with benchmarks before and after changes

## Caching Strategies
- **Cache-Aside**: Check cache → miss → query source → set cache
- **Write-Through**: Write to cache and source simultaneously
- **Write-Behind**: Write to cache, async flush to source
- **TTL with Jitter**: Prevent thundering herd with randomized expiry
- Cache at multiple layers: CDN, reverse proxy, application, database

## Database Optimization
- Index columns used in WHERE, JOIN, ORDER BY
- Use EXPLAIN ANALYZE to verify query plans
- Avoid N+1 queries — use JOINs or batch loading
- Use connection pooling
- Partition large tables by date or key range
- Use read replicas for read-heavy workloads

## HTTP/API Performance
- Enable HTTP/2 for multiplexing
- Compress responses (gzip/brotli)
- Use CDN for static assets
- Implement request coalescing for duplicate concurrent requests
- Use pagination with cursors for large datasets
- Set appropriate cache headers (ETag, Cache-Control)

## Scalability Patterns
- Horizontal scaling: add more instances behind load balancer
- Vertical scaling: increase instance resources (quick fix)
- Async processing: offload work to queues
- Sharding: distribute data across multiple databases
- Event-driven: decouple services with message queues
- Circuit breaker: prevent cascade failures
