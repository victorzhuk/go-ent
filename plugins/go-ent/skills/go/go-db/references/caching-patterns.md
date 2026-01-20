# Redis Cache-Aside Pattern

<example>
<input>Implement cached repository with cache-aside pattern</input>
<output>
```go
package userrepo

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type cachedRepository struct {
    repo  *repository
    redis *redis.Client
    ttl   time.Duration
}

func NewCached(repo *repository, redis *redis.Client, ttl time.Duration) *cachedRepository {
    return &cachedRepository{
        repo:  repo,
        redis: redis,
        ttl:   ttl,
    }
}

func (r *cachedRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    key := fmt.Sprintf("user:%s", id)

    // Try cache first
    data, err := r.redis.Get(ctx, key).Bytes()
    if err == nil {
        var user entity.User
        if err := json.Unmarshal(data, &user); err == nil {
            return &user, nil
        }
    }

    // Cache miss - get from DB
    user, err := r.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("db query: %w", err)
    }

    // Populate cache
    data, _ = json.Marshal(user)
    _ = r.redis.Set(ctx, key, data, r.ttl).Err()

    return user, nil
}

func (r *cachedRepository) Save(ctx context.Context, user *entity.User) error {
    // Save to DB first
    if err := r.repo.Save(ctx, user); err != nil {
        return fmt.Errorf("db save: %w", err)
    }

    // Invalidate cache
    key := fmt.Sprintf("user:%s", user.ID)
    _ = r.redis.Del(ctx, key).Err()

    return nil
}
```

**Pattern**: Cache-aside ensures cache is only populated on demand, invalidation happens on writes.
</output>
</example>
