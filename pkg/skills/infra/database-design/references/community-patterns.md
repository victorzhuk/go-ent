## Schema Design Principles
- Use UUIDs for primary keys (avoid sequential IDs for security)
- Add `created_at` and `updated_at` timestamps to all tables
- Use appropriate data types (don't store dates as strings)
- Normalize to 3NF by default; denormalize intentionally for read performance
- Use constraints: NOT NULL, UNIQUE, CHECK, FOREIGN KEY

## Indexing Strategy
- Index columns used in WHERE, JOIN, and ORDER BY clauses
- Use composite indexes for multi-column queries (leftmost prefix rule)
- Use partial indexes for filtered queries: `WHERE status = 'active'`
- Use covering indexes to avoid table lookups
- Don't over-index — each index slows writes and uses storage
- Use EXPLAIN ANALYZE to verify index usage

## Migrations
- Use forward-only migrations in production
- Make migrations backwards-compatible (expand-and-contract)
- Never modify existing migration files
- Test migrations against production-like data volumes
- Include both up and down scripts

## Query Optimization
- Avoid SELECT * — specify needed columns
- Use JOINs instead of subqueries where possible
- Use LIMIT for pagination
- Avoid N+1 queries (use eager loading in ORMs)
- Use batch operations for bulk inserts/updates
- Use connection pooling

## PostgreSQL Patterns
- Use `jsonb` for semi-structured data (with GIN indexes)
- Use `array` types for simple lists
- Use CTEs (`WITH`) for readable complex queries
- Use window functions for rankings and running totals
- Use `UPSERT` (INSERT ON CONFLICT) for idempotent operations
- Use advisory locks for distributed coordination
