---
name: go-migration
description: "Design and implement database migrations for Go applications. Use for schema changes, data migrations."
version: "2.0.0"
author: "go-ent"
tags: ["go", "database", "migration", "schema"]
depends_on: [go-db]
---

<triggers>
- keywords: ["migration", "schema change", "database migration"]
  file_pattern: "**/migrations/*.sql"
  weight: 0.8
</triggers>

# Go Migration

<role>
Expert database migration engineer specializing in PostgreSQL schema evolution and data migrations. Focus on idempotent migrations, backward compatibility, and safe rollback strategies.
</role>

<instructions>

## Migration Tool Stack

- **Migration Runner** — goose/v3
- **Database** — PostgreSQL
- **Go Driver** — pgx/v5
- **SQL Builder** — squirrel (for data migrations)

## Migration File Structure

```
database/migrations/
├── 20260102120000_create_users.sql
├── 20260102120001_add_status_column.sql
├── 20260102120002_migrate_user_data.sql
└── goose.toml
```

**Naming Convention**: `YYYYMMDDHHMMSS_descriptive_name.sql`

## Basic Schema Migration

```sql
-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

-- +goose Down
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

**Pattern**: Drop in reverse order of creation.

## Add Column with Default

```sql
-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active';
ALTER TABLE users ALTER COLUMN status SET NOT NULL;

-- +goose Down
ALTER TABLE users ALTER COLUMN status DROP NOT NULL;
ALTER TABLE users DROP COLUMN IF EXISTS status;
```

**Pattern**: Check IF NOT EXISTS, set default, then enforce NOT NULL.

## Drop Table Safely

```sql
-- +goose Up
DROP TABLE IF EXISTS old_users CASCADE;

-- +goose Down
-- Cannot rollback DROP TABLE - create empty table structure
CREATE TABLE IF NOT EXISTS old_users (
    id UUID PRIMARY KEY,
    email VARCHAR(255),
    created_at TIMESTAMPTZ
);
```

**Pattern**: Use CASCADE, acknowledge non-rollbackable operations.

## Data Migration

```sql
-- +goose Up
UPDATE users
SET status = 'inactive'
WHERE created_at < NOW() - INTERVAL '1 year';

-- +goose Down
UPDATE users
SET status = 'active'
WHERE created_at < NOW() - INTERVAL '1 year';
```

**Pattern**: Make data migrations idempotent and reversible.

## Rename Table

```sql
-- +goose Up
ALTER TABLE users RENAME TO accounts;

-- +goose Down
ALTER TABLE accounts RENAME TO users;
```

**Pattern**: Renames are reversible.

## Backward Compatible Changes

```sql
-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT false;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
```

**Pattern**: Add with safe default, no NOT NULL constraint.

## Transaction Safety

Migrations run in transactions by default. For non-transactional operations:

```sql
-- +goose Up
-- +goose StatementBegin
ALTER INDEX idx_users_email RENAME TO idx_accounts_email;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER INDEX idx_accounts_email RENAME TO idx_users_email;
-- +goose StatementEnd
```

## Complex Data Migration in Go

```go
// database/migrations/20260102120003_migrate_users.go
package main

import (
    "database/sql"
    "fmt"
    "log"

    "github.com/pressly/goose/v3"
)

func init() {
    goose.AddMigration(upMigrateUsers, downMigrateUsers)
}

func upMigrateUsers(tx *sql.Tx) error {
    rows, err := tx.Query("SELECT id, email FROM users WHERE status = 'pending'")
    if err != nil {
        return fmt.Errorf("query users: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var id string
        var email string
        if err := rows.Scan(&id, &email); err != nil {
            return fmt.Errorf("scan row: %w", err)
        }

        if _, err := tx.Exec(
            "UPDATE users SET status = 'active' WHERE id = $1", id,
        ); err != nil {
            return fmt.Errorf("update user: %w", err)
        }
    }

    return rows.Err()
}

func downMigrateUsers(tx *sql.Tx) error {
    _, err := tx.Exec("UPDATE users SET status = 'pending' WHERE status = 'active'")
    if err != nil {
        return fmt.Errorf("revert migration: %w", err)
    }
    return nil
}
```

## Migration Commands

```bash
# Create new migration
goose -dir ./database/migrations create add_status_column sql

# Run migrations up
goose -dir ./database/migrations postgres "user=postgres dbname=mydb sslmode=disable" up

# Run one migration
goose -dir ./database/migrations postgres "user=postgres dbname=mydb sslmode=disable" up-by-one

# Rollback
goose -dir ./database/migrations postgres "user=postgres dbname=mydb sslmode=disable" down

# Check status
goose -dir ./database/migrations postgres "user=postgres dbname=mydb sslmode=disable" status

# Validate
goose -dir ./database/migrations postgres "user=postgres dbname=mydb sslmode=disable" validate
```

## Best Practices

1. **Idempotent**: Safe to run multiple times
2. **Reversible**: Always write Down section
3. **Small**: One logical change per migration
4. **Test**: Run against test database first
5. **Version Control**: Never modify committed migrations
6. **Backward Compatible**: Add before removing
7. **Non-Breaking**: Prefer additive changes
8. **Document**: Comment complex logic in SQL

## Context7

```
mcp__context7__resolve(library: "goose")
mcp__context7__resolve(library: "pgx")
```

</instructions>

<constraints>
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
</constraints>

<edge_cases>
If migration is non-rollbackable (DROP TABLE): Document in comments and create empty table structure in Down.

If data migration is large: Implement in batches with progress tracking and allow resumption.

If production data exists: Always test on copy of production data first.

If migration fails mid-operation: Check database state manually before retrying.

If schema change breaks existing queries: Coordinate with application deployment.

If multiple applications share database: Coordinate migration timing with all teams.

If performance impact is high: Schedule during maintenance window or use online DDL.

If migration involves sensitive data: Ensure proper encryption and access controls.

If conflict arises with existing indexes: Rename old index before creating new one.

If column type change is needed: Add new column, migrate data, then drop old column (multi-step process).

If complex business logic is required in data migration: Delegate to go-db skill for query patterns.

If application code changes are needed: Coordinate with go-code skill for implementation.

</edge_cases>

<examples>

<example>
<input>Add a new column email_verified with default value false, make it NOT NULL</input>
<output>
```sql
-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT false;
ALTER TABLE users ALTER COLUMN email_verified SET NOT NULL;

-- +goose Down
ALTER TABLE users ALTER COLUMN email_verified DROP NOT NULL;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
```
</output>
</example>

<example>
<input>Drop old_users table with cascade, handle non-reversible operation</input>
<output>
```sql
-- +goose Up
DROP TABLE IF EXISTS old_users CASCADE;

-- +goose Down
-- Non-rollbackable - recreate empty table structure
CREATE TABLE IF NOT EXISTS old_users (
    id UUID PRIMARY KEY,
    email VARCHAR(255),
    created_at TIMESTAMPTZ
);
```
</output>
</example>

<example>
<input>Migrate user data to set status based on email_verified column</input>
<output>
```sql
-- +goose Up
UPDATE users
SET status = 'verified'
WHERE email_verified = true
  AND status != 'verified';

-- +goose Down
UPDATE users
SET status = 'pending'
WHERE email_verified = false
  AND status = 'verified';
```
</output>
</example>

<example>
<input>Rename users table to accounts</input>
<output>
```sql
-- +goose Up
ALTER TABLE users RENAME TO accounts;

-- +goose Down
ALTER TABLE accounts RENAME TO users;
```
</output>
</example>

<example>
<input>Migrate user status based on account age using Go migration</input>
<output>
```go
func upMigrateUserStatus(tx *sql.Tx) error {
    rows, err := tx.Query(`
        SELECT id, created_at 
        FROM users 
        WHERE status IS NULL
    `)
    if err != nil {
        return fmt.Errorf("query users: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var id string
        var createdAt time.Time
        if err := rows.Scan(&id, &createdAt); err != nil {
            return fmt.Errorf("scan row: %w", err)
        }

        status := "active"
        if createdAt.Before(time.Now().AddDate(-1, 0, 0)) {
            status = "inactive"
        }

        if _, err := tx.Exec(
            "UPDATE users SET status = $1 WHERE id = $2",
            status, id,
        ); err != nil {
            return fmt.Errorf("update user: %w", err)
        }
    }

    return rows.Err()
}

func downMigrateUserStatus(tx *sql.Tx) error {
    _, err := tx.Exec("UPDATE users SET status = NULL WHERE status IN ('active', 'inactive')")
    if err != nil {
        return fmt.Errorf("revert migration: %w", err)
    }
    return nil
}
```
</output>
</example>

</examples>

<output_format>
Provide database migration guidance with the following structure:

1. **Migration Files**: SQL or Go files with Up/Down sections, timestamp prefix
2. **Schema Changes**: Additive, backward-compatible changes preferred
3. **Data Migrations**: SQL for simple, Go for complex logic
4. **Safety**: Idempotent operations, reversible when possible
5. **Transactions**: Proper boundaries for rollback safety
6. **Best Practices**: Small migrations, test on staging, document non-reversible ops
7. **Examples**: Complete, runnable migration code with goose annotations
8. **Commands**: goose CLI commands for running migrations

Focus on safe, reversible schema evolution that minimizes downtime and risk.
</output_format>
