# Constraints

- Include profiling before optimization (measure first, then optimize)
- Include benchmarks with meaningful comparison data
- Include pre-allocation of slices/maps when size is known
- Include connection pooling for databases, HTTP clients, etc.
- Include batch operations for bulk inserts/updates
- Include concurrent processing with errgroup or worker pools
- Include rate limiting for external service calls
- Include memory reuse with sync.Pool for hot paths
- Exclude premature optimization without profiling data
- Exclude micro-optimizations with negligible impact
- Exclude sacrificing readability for minor performance gains
- Exclude ignoring error handling for performance
- Exclude hard-coding limits without measurement
- Bound to data-driven performance improvements
- Follow "measure → optimize → verify" cycle