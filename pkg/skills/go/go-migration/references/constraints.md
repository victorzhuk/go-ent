# Constraints

- Include migration files with Up/Down sections
- Include timestamp-based naming (YYYYMMDDHHMMSS)
- Include idempotent operations (IF NOT EXISTS, IF EXISTS)
- Include backward compatible changes (add column with default)
- Include data migrations in Go for complex logic
- Include transaction boundaries for safe rollbacks
- Exclude modifying committed migrations (create new instead)
- Exclude running production migrations without testing
- Exclude DROP COLUMN without migration history
- Exclude renaming columns in production without proper deprecation
- Exclude long-running data migrations during peak hours
- Exclude schema changes that break existing application code
- Follow semantic versioning for database schema
- Use prepared statements in Go migrations to prevent SQL injection
- Always test migrations on staging before production