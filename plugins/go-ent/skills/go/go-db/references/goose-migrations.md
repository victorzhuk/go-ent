# Goose Migrations with Rollback Strategy

<example>
<input>Write goose migration with proper rollback strategy</input>
<output>
```sql
-- +goose Up
BEGIN;

CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    status user_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at DESC);

CREATE TRIGGER update_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;

-- +goose Down
BEGIN;

DROP TRIGGER IF EXISTS update_updated_at ON users;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_status;

COMMIT;
```

**Best practices**:
- Wrap in transactions
- Drop in reverse order of creation
- Use IF EXISTS for safe rollback
- Include indexes and triggers
- Add updated_at trigger for audit trail
</output>
</example>
