---
name: postgresql-advanced
description: PostgreSQL advanced patterns, indexing, JSON, partitioning, full-text search, and administration
---

# PostgreSQL Advanced

## Index Types
```sql
-- B-tree (default): equality, range, sorting
CREATE INDEX idx_users_email ON users (email);

-- GIN: arrays, JSONB, full-text search
CREATE INDEX idx_users_tags ON users USING GIN (tags);

-- Partial: filtered index for common queries
CREATE INDEX idx_active_users ON users (email) WHERE status = 'active';

-- Composite: multi-column (leftmost prefix rule)
CREATE INDEX idx_user_status_created ON users (status, created_at DESC);

-- Covering: include extra columns to avoid table lookup
CREATE INDEX idx_users_email_cover ON users (email) INCLUDE (name, status);
```

## JSONB Patterns
```sql
-- Query nested values
SELECT * FROM events WHERE data->>'type' = 'purchase';
SELECT * FROM events WHERE data @> '{"type": "purchase"}';

-- GIN index on JSONB
CREATE INDEX idx_events_data ON events USING GIN (data jsonb_path_ops);

-- Update nested values
UPDATE events SET data = jsonb_set(data, '{status}', '"completed"');
```

## Window Functions
```sql
-- Running total
SELECT date, amount,
    SUM(amount) OVER (ORDER BY date) as running_total
FROM transactions;

-- Rank within groups
SELECT department, name, salary,
    RANK() OVER (PARTITION BY department ORDER BY salary DESC) as dept_rank
FROM employees;
```

## CTEs and Recursive Queries
```sql
-- Recursive: org hierarchy
WITH RECURSIVE org_tree AS (
    SELECT id, name, manager_id, 0 as depth
    FROM employees WHERE manager_id IS NULL
    UNION ALL
    SELECT e.id, e.name, e.manager_id, t.depth + 1
    FROM employees e JOIN org_tree t ON e.manager_id = t.id
)
SELECT * FROM org_tree ORDER BY depth, name;
```

## Performance
- Use `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)` for query analysis
- Monitor with `pg_stat_statements` extension
- Vacuum regularly (autovacuum should be enabled)
- Use `pg_stat_user_tables` to check dead tuple ratio
- Set `work_mem` appropriately for sort/hash operations
- Use connection pooling (PgBouncer) in production
