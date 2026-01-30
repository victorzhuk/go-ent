# Constraints

- Include repository pattern with private models and public entities
- Include squirrel for complex queries (joins, dynamic WHERE, pagination)
- Include proper connection pooling with pgxpool
- Include error mapping (pgx.ErrNoRows → domain errors)
- Include transaction support for multi-operation writes
- Include migration management with goose
- Include caching strategy with Redis (cache-aside pattern)
- Include connection lifecycle management (begin, commit, rollback)
- Exclude raw SQL in application code (use squirrel or prepared statements)
- Exclude N+1 query problems (use joins or batch queries)
- Exclude unparameterized queries (use prepared statements to prevent injection)
- Exclude database-specific types leaking into domain layer
- Exclude running migrations in production without proper testing
- Bound to repository layer with entity types from domain
- Follow SQL best practices (indexes, constraints, proper data types)
- Use context for all database operations with proper timeouts