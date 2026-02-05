# Database Migrations

Production-grade database migrations using goose/v3. Goose provides a simple, reliable approach to schema versioning with both SQL and Go migration support.

**Why goose over alternatives:**
- Simple, focused tool (does one thing well)
- SQL-first approach (readable, reviewable migrations)
- Go code support for complex data transformations
- Embedded migrations support (ship migrations with binary)
- No ORM coupling (works with any database driver)
- Battle-tested in production at scale

## Quick Reference

| Command                         | Purpose                          |
|---------------------------------|----------------------------------|
| `goose create NAME sql`         | Create timestamped SQL migration |
| `goose create NAME go`          | Create Go code migration         |
| `goose up`                      | Apply all pending migrations     |
| `goose up-by-one`               | Apply next migration only        |
| `goose down`                    | Rollback last migration          |
| `goose down-to VERSION`         | Rollback to specific version     |
| `goose status`                  | Show migration status            |
| `goose version`                 | Show current schema version      |
| `goose fix`                     | Fix migration sequence numbers   |

## Installation

### Go Install (Recommended)

```bash
# Install as project tool
go install github.com/pressly/goose/v3/cmd/goose@latest

# Verify installation
goose -version
```

### As Project Dependency

```go
// tools.go
//go:build tools

package tools

import _ "github.com/pressly/goose/v3/cmd/goose"
```

```bash
# Install
go mod tidy

# Use via go run
go run github.com/pressly/goose/v3/cmd/goose@latest status
```

### Docker

```dockerfile
FROM golang:1.23-alpine AS builder
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine:latest
COPY --from=builder /go/bin/goose /usr/local/bin/
ENTRYPOINT ["goose"]
```

## SQL Migrations

### Create Migration

```bash
# Create timestamped migration
goose -dir ./migrations create add_users_table sql

# Creates: migrations/20240205120000_add_users_table.sql
```

### Migration Structure

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id         BIGSERIAL PRIMARY KEY,
    email      VARCHAR(255) NOT NULL UNIQUE,
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TABLE IF EXISTS users CASCADE;
-- +goose StatementEnd
```

### Transaction Handling

```sql
-- Automatic transaction (default)
-- +goose Up
CREATE TABLE products (id BIGSERIAL PRIMARY KEY);
ALTER TABLE products ADD COLUMN name VARCHAR(255);

-- +goose Down
DROP TABLE products;
```

```sql
-- Disable transaction for specific operations
-- +goose NO TRANSACTION
-- +goose Up
CREATE INDEX CONCURRENTLY idx_products_name ON products(name);

-- +goose Down
DROP INDEX CONCURRENTLY idx_products_name;
```

### Idempotent Migrations

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY
);

-- Safe column addition
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'orders' AND column_name = 'status'
    ) THEN
        ALTER TABLE orders ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'pending';
    END IF;
END $$;

-- +goose Down
ALTER TABLE orders DROP COLUMN IF EXISTS status;
DROP TABLE IF EXISTS orders;
```

## Go Migrations

### When to Use Go Code

Use Go migrations for:
- Complex data transformations
- Multi-step operations requiring logic
- Backfilling data based on business rules
- Integrating with external services
- Operations requiring transaction control

### Create Go Migration

```bash
goose -dir ./migrations create backfill_user_roles go

# Creates: migrations/20240205120001_backfill_user_roles.go
```

### Basic Go Migration

```go
package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upBackfillUserRoles, downBackfillUserRoles)
}

func upBackfillUserRoles(ctx context.Context, tx *sql.Tx) error {
	query := `
		UPDATE users
		SET role = 'user'
		WHERE role IS NULL OR role = ''
	`

	result, err := tx.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("update users: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	fmt.Printf("Updated %d users with default role\n", rows)
	return nil
}

func downBackfillUserRoles(ctx context.Context, tx *sql.Tx) error {
	// Backfill operations typically cannot be reversed
	return nil
}
```

### Complex Data Transformation

```go
func upMigrateEmailToLowercase(ctx context.Context, tx *sql.Tx) error {
	// Fetch users in batches
	const batchSize = 1000
	var lastID int64

	for {
		query := `
			SELECT id, email
			FROM users
			WHERE id > $1
			ORDER BY id
			LIMIT $2
		`

		rows, err := tx.QueryContext(ctx, query, lastID, batchSize)
		if err != nil {
			return fmt.Errorf("query users: %w", err)
		}

		var users []struct {
			ID    int64
			Email string
		}

		for rows.Next() {
			var u struct {
				ID    int64
				Email string
			}
			if err := rows.Scan(&u.ID, &u.Email); err != nil {
				rows.Close()
				return fmt.Errorf("scan user: %w", err)
			}
			users = append(users, u)
		}
		rows.Close()

		if len(users) == 0 {
			break
		}

		// Update batch
		stmt, err := tx.PrepareContext(ctx, `UPDATE users SET email = $1 WHERE id = $2`)
		if err != nil {
			return fmt.Errorf("prepare statement: %w", err)
		}

		for _, u := range users {
			if _, err := stmt.ExecContext(ctx, strings.ToLower(u.Email), u.ID); err != nil {
				stmt.Close()
				return fmt.Errorf("update user %d: %w", u.ID, err)
			}
		}
		stmt.Close()

		lastID = users[len(users)-1].ID
	}

	return nil
}
```

### Accessing DB Connection

```go
func upWithCustomLogic(ctx context.Context, tx *sql.Tx) error {
	// Access transaction for operations
	rows, err := tx.QueryContext(ctx, "SELECT id FROM users")
	if err != nil {
		return err
	}
	defer rows.Close()

	// Use pgx if needed (convert *sql.Tx to pgx)
	// Note: goose v3 uses database/sql, not pgx directly

	return nil
}
```

## Migration Patterns

### Add Column with Default

```sql
-- +goose Up
-- Step 1: Add column as nullable
ALTER TABLE users ADD COLUMN status VARCHAR(50);

-- Step 2: Backfill existing rows
UPDATE users SET status = 'active' WHERE status IS NULL;

-- Step 3: Add NOT NULL constraint
ALTER TABLE users ALTER COLUMN status SET NOT NULL;

-- Step 4: Add default for new rows
ALTER TABLE users ALTER COLUMN status SET DEFAULT 'active';

-- +goose Down
ALTER TABLE users DROP COLUMN status;
```

### Backfill Data Safely

```sql
-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    batch_size INT := 10000;
    affected INT;
BEGIN
    LOOP
        UPDATE users
        SET normalized_email = LOWER(email)
        WHERE normalized_email IS NULL
        AND id IN (
            SELECT id FROM users
            WHERE normalized_email IS NULL
            LIMIT batch_size
        );

        GET DIAGNOSTICS affected = ROW_COUNT;
        EXIT WHEN affected = 0;

        RAISE NOTICE 'Updated % rows', affected;
        COMMIT;
    END LOOP;
END $$;
-- +goose StatementEnd

-- +goose Down
UPDATE users SET normalized_email = NULL;
```

### Zero-Downtime Column Removal

```sql
-- Migration 1: Stop writing to column (application deploy)
-- Application code change: Remove writes to old_column

-- Migration 2: Remove column
-- +goose Up
ALTER TABLE users DROP COLUMN old_column;

-- +goose Down
-- Cannot restore data, add column back as nullable
ALTER TABLE users ADD COLUMN old_column TEXT;
```

### Rename Column Safely

```sql
-- Migration 1: Add new column
-- +goose Up
ALTER TABLE users ADD COLUMN full_name VARCHAR(255);
UPDATE users SET full_name = name;

-- +goose Down
ALTER TABLE users DROP COLUMN full_name;
```

```sql
-- Migration 2 (after application supports both columns): Remove old column
-- +goose Up
ALTER TABLE users DROP COLUMN name;

-- +goose Down
ALTER TABLE users ADD COLUMN name VARCHAR(255);
UPDATE users SET name = full_name;
```

### Rollback Strategy

```sql
-- Always provide rollback path
-- +goose Up
ALTER TABLE orders ADD COLUMN tracking_number VARCHAR(100);
CREATE INDEX idx_orders_tracking ON orders(tracking_number);

-- +goose Down
DROP INDEX IF EXISTS idx_orders_tracking;
ALTER TABLE orders DROP COLUMN IF EXISTS tracking_number;
```

## Configuration

### Embed Migrations

```go
package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func RunMigrations(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
```

### Environment Variables

```bash
# PostgreSQL
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="postgres://user:pass@localhost:5432/dbname?sslmode=disable"
export GOOSE_MIGRATION_DIR=./migrations

# Run migrations
goose up

# Check status
goose status
```

### Programmatic Usage

```go
package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

type Config struct {
	MigrationDir string
	Driver       string
	DSN          string
}

func Migrate(ctx context.Context, cfg Config) error {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect(cfg.Driver); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	// Apply all pending migrations
	if err := goose.UpContext(ctx, db, cfg.MigrationDir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	// Get current version
	version, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}

	fmt.Printf("Database migrated to version: %d\n", version)
	return nil
}

func Rollback(ctx context.Context, cfg Config) error {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect(cfg.Driver); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	// Rollback last migration
	if err := goose.DownContext(ctx, db, cfg.MigrationDir); err != nil {
		return fmt.Errorf("rollback migration: %w", err)
	}

	return nil
}
```

### Migration Status Check

```go
func CheckMigrationStatus(ctx context.Context, db *sql.DB, dir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	current, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	migrations, err := goose.CollectMigrations(dir, 0, goose.MaxVersion)
	if err != nil {
		return fmt.Errorf("collect migrations: %w", err)
	}

	latest := int64(0)
	if len(migrations) > 0 {
		latest = migrations[len(migrations)-1].Version
	}

	if current < latest {
		return fmt.Errorf("database schema outdated: current=%d, latest=%d", current, latest)
	}

	return nil
}
```

## CI/CD Integration

### Init Container (Kubernetes)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
spec:
  template:
    spec:
      initContainers:
        - name: migrations
          image: myapp:latest
          command:
            - /app/migrate
          env:
            - name: DB_HOST
              valueFrom:
                secretKeyRef:
                  name: db-credentials
                  key: host
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-credentials
                  key: password
      containers:
        - name: api
          image: myapp:latest
```

### Migration Binary

```go
// cmd/migrate/main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("set dialect: %v", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	version, _ := goose.GetDBVersionContext(ctx, db)
	fmt.Printf("Database at version: %d\n", version)
}
```

### Health Check After Migration

```go
func HealthCheck(ctx context.Context, db *sql.DB) error {
	// Check database is accessible
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}

	// Verify expected schema version
	version, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}

	expectedVersion := int64(20240205120000)
	if version < expectedVersion {
		return fmt.Errorf("schema outdated: have=%d, want=%d", version, expectedVersion)
	}

	return nil
}
```

### Rollback on Failure

```bash
#!/bin/bash
set -e

# Run migrations
if ! goose up; then
    echo "Migration failed, rolling back..."
    goose down
    exit 1
fi

# Deploy application
if ! ./deploy.sh; then
    echo "Deployment failed, rolling back migration..."
    goose down
    exit 1
fi

echo "Deployment successful"
```

### GitHub Actions Example

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install goose
        run: go install github.com/pressly/goose/v3/cmd/goose@latest

      - name: Run migrations
        env:
          GOOSE_DRIVER: postgres
          GOOSE_DBSTRING: ${{ secrets.DATABASE_URL }}
        run: goose -dir ./migrations up

      - name: Verify migration
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: |
          version=$(goose -dir ./migrations version)
          echo "Database at version: $version"
```

## Common Mistakes

| Mistake                              | Fix                                                  |
|--------------------------------------|------------------------------------------------------|
| Non-idempotent migrations            | Use `IF NOT EXISTS`, check before alter              |
| Missing rollback path                | Always provide `-- +goose Down` implementation       |
| Data loss on down migration          | Document irreversible operations, backup first       |
| Breaking changes without versioning  | Use multi-step migrations for backward compatibility |
| Not testing rollback                 | Test `up` and `down` in dev environment              |
| Large migrations without batching    | Process data in chunks, commit between batches       |
| Missing transaction control          | Default is transactional, use `NO TRANSACTION` only when needed |
| Concurrent index creation in TX      | Use `CREATE INDEX CONCURRENTLY` with `NO TRANSACTION`|

## Best Practices

```sql
-- ✓ Good - idempotent
CREATE TABLE IF NOT EXISTS users (id BIGSERIAL PRIMARY KEY);

-- ✗ Bad - fails on retry
CREATE TABLE users (id BIGSERIAL PRIMARY KEY);

-- ✓ Good - safe default addition
ALTER TABLE users ADD COLUMN status VARCHAR(50);
UPDATE users SET status = 'active' WHERE status IS NULL;
ALTER TABLE users ALTER COLUMN status SET NOT NULL;

-- ✗ Bad - breaks existing rows
ALTER TABLE users ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'active';

-- ✓ Good - reversible
-- +goose Down
DROP TABLE IF EXISTS users CASCADE;

-- ✗ Bad - data loss without warning
-- +goose Down
DROP TABLE users;

-- ✓ Good - documented irreversible operation
-- +goose Down
-- WARNING: This migration deletes data. Restore from backup if needed.
-- Cannot automatically reverse data deletion.
```

## See Also

- [PostgreSQL](./postgresql.md) - Database operations and patterns
- [MongoDB](./mongodb.md) - NoSQL migration strategies
- [Docker](../13-devops/docker.md) - Containerized migration workflows
- [CI/CD](../13-devops/ci-cd.md) - Automated deployment pipelines
- [Kubernetes](../13-devops/kubernetes.md) - Init container patterns
- [goose documentation](https://github.com/pressly/goose)
