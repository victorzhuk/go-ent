# MongoDB

Production-grade MongoDB integration using go.mongodb.org/mongo-driver v1.17+ (latest stable). MongoDB is ideal for flexible schemas, document-oriented data, rapid prototyping, and hierarchical structures. Use PostgreSQL for strict schemas, complex transactions, and relational integrity.

## Quick Reference

| Pattern                                | Use When                          |
|----------------------------------------|-----------------------------------|
| `mongo.Connect(ctx, options.Client())` | Establish connection              |
| `collection.InsertOne(ctx, doc)`       | Insert single document            |
| `collection.Find(ctx, filter)`         | Query multiple documents          |
| `collection.UpdateOne(ctx, filter)`    | Update single document            |
| `collection.DeleteOne(ctx, filter)`    | Delete single document            |
| `session.StartTransaction()`           | Multi-document ACID transactions  |
| `collection.Indexes().CreateOne()`     | Create index                      |
| `collection.Aggregate(ctx, pipeline)`  | Aggregation pipeline              |

## Connection Setup

### Client with Connection Pool

```go
import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
)

type Config struct {
    URI          string
    Database     string
    MaxPoolSize  uint64
    MinPoolSize  uint64
    MaxConnIdle  time.Duration
    Timeout      time.Duration
}

func NewClient(ctx context.Context, cfg Config) (*mongo.Client, error) {
    clientOpts := options.Client().
        ApplyURI(cfg.URI).
        SetMaxPoolSize(cfg.MaxPoolSize).      // Default: 100
        SetMinPoolSize(cfg.MinPoolSize).      // Default: 0
        SetMaxConnIdleTime(cfg.MaxConnIdle).  // Default: 0 (no limit)
        SetTimeout(cfg.Timeout).              // Default: 30s
        SetServerSelectionTimeout(5 * time.Second)

    client, err := mongo.Connect(ctx, clientOpts)
    if err != nil {
        return nil, fmt.Errorf("connect: %w", err)
    }

    // Verify connection
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx, readpref.Primary()); err != nil {
        return nil, fmt.Errorf("ping: %w", err)
    }

    return client, nil
}

func (c *Client) Close(ctx context.Context) error {
    if err := c.client.Disconnect(ctx); err != nil {
        return fmt.Errorf("disconnect: %w", err)
    }
    return nil
}
```

### Connection String

```go
const (
    // Local development
    localURI = "mongodb://localhost:27017"

    // With credentials
    authURI = "mongodb://user:pass@localhost:27017/dbname?authSource=admin"

    // Replica set
    replicaURI = "mongodb://host1:27017,host2:27017,host3:27017/dbname?replicaSet=rs0"

    // MongoDB Atlas
    atlasURI = "mongodb+srv://user:pass@cluster.mongodb.net/dbname?retryWrites=true&w=majority"

    // With all options
    fullURI = "mongodb://user:pass@localhost:27017/dbname?" +
        "maxPoolSize=100&" +
        "minPoolSize=10&" +
        "maxIdleTimeMS=60000&" +
        "serverSelectionTimeoutMS=5000"
)
```

### Repository Pattern

```go
type Repository struct {
    client *mongo.Client
    db     *mongo.Database
}

func NewRepository(client *mongo.Client, dbName string) *Repository {
    return &Repository{
        client: client,
        db:     client.Database(dbName),
    }
}

func (r *Repository) users() *mongo.Collection {
    return r.db.Collection("users")
}

func (r *Repository) orders() *mongo.Collection {
    return r.db.Collection("orders")
}
```

## CRUD Operations

### Insert Operations

#### Insert Single Document

```go
import (
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Email     string             `bson:"email"`
    Name      string             `bson:"name"`
    Age       int                `bson:"age,omitempty"`
    Tags      []string           `bson:"tags,omitempty"`
    CreatedAt time.Time          `bson:"created_at"`
}

func (r *Repository) CreateUser(ctx context.Context, user User) (primitive.ObjectID, error) {
    user.CreatedAt = time.Now()

    result, err := r.users().InsertOne(ctx, user)
    if err != nil {
        return primitive.NilObjectID, fmt.Errorf("insert user: %w", err)
    }

    id := result.InsertedID.(primitive.ObjectID)
    return id, nil
}
```

#### Insert Multiple Documents

```go
func (r *Repository) CreateUsers(ctx context.Context, users []User) ([]primitive.ObjectID, error) {
    docs := make([]interface{}, len(users))
    for i := range users {
        users[i].CreatedAt = time.Now()
        docs[i] = users[i]
    }

    result, err := r.users().InsertMany(ctx, docs)
    if err != nil {
        return nil, fmt.Errorf("insert users: %w", err)
    }

    ids := make([]primitive.ObjectID, len(result.InsertedIDs))
    for i, id := range result.InsertedIDs {
        ids[i] = id.(primitive.ObjectID)
    }

    return ids, nil
}
```

### Query Operations

#### Find Single Document

```go
func (r *Repository) GetUser(ctx context.Context, id primitive.ObjectID) (*User, error) {
    filter := bson.M{"_id": id}

    var user User
    err := r.users().FindOne(ctx, filter).Decode(&user)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("find user: %w", err)
    }

    return &user, nil
}
```

#### Find Multiple Documents

```go
func (r *Repository) ListUsers(ctx context.Context, limit int64) ([]User, error) {
    opts := options.Find().
        SetLimit(limit).
        SetSort(bson.D{{Key: "created_at", Value: -1}})

    cursor, err := r.users().Find(ctx, bson.M{}, opts)
    if err != nil {
        return nil, fmt.Errorf("find users: %w", err)
    }
    defer cursor.Close(ctx)

    var users []User
    if err := cursor.All(ctx, &users); err != nil {
        return nil, fmt.Errorf("decode users: %w", err)
    }

    return users, nil
}
```

#### Find with Filters

```go
func (r *Repository) FindUsersByAge(ctx context.Context, minAge, maxAge int) ([]User, error) {
    filter := bson.M{
        "age": bson.M{
            "$gte": minAge,
            "$lte": maxAge,
        },
    }

    cursor, err := r.users().Find(ctx, filter)
    if err != nil {
        return nil, fmt.Errorf("find: %w", err)
    }
    defer cursor.Close(ctx)

    var users []User
    if err := cursor.All(ctx, &users); err != nil {
        return nil, fmt.Errorf("decode: %w", err)
    }

    return users, nil
}

func (r *Repository) FindUsersByTags(ctx context.Context, tags []string) ([]User, error) {
    // Find users with ANY of the tags
    filter := bson.M{
        "tags": bson.M{"$in": tags},
    }

    // Find users with ALL of the tags
    // filter := bson.M{
    //     "tags": bson.M{"$all": tags},
    // }

    cursor, err := r.users().Find(ctx, filter)
    if err != nil {
        return nil, fmt.Errorf("find: %w", err)
    }
    defer cursor.Close(ctx)

    var users []User
    return users, cursor.All(ctx, &users)
}
```

### Update Operations

#### Update Single Document

```go
func (r *Repository) UpdateUserName(ctx context.Context, id primitive.ObjectID, name string) error {
    filter := bson.M{"_id": id}
    update := bson.M{
        "$set": bson.M{
            "name":       name,
            "updated_at": time.Now(),
        },
    }

    result, err := r.users().UpdateOne(ctx, filter, update)
    if err != nil {
        return fmt.Errorf("update user: %w", err)
    }

    if result.MatchedCount == 0 {
        return ErrNotFound
    }

    return nil
}
```

#### Update Multiple Fields

```go
func (r *Repository) UpdateUser(ctx context.Context, id primitive.ObjectID, updates map[string]interface{}) error {
    filter := bson.M{"_id": id}

    updates["updated_at"] = time.Now()
    update := bson.M{"$set": updates}

    result, err := r.users().UpdateOne(ctx, filter, update)
    if err != nil {
        return fmt.Errorf("update: %w", err)
    }

    if result.MatchedCount == 0 {
        return ErrNotFound
    }

    return nil
}
```

#### Update or Insert (Upsert)

```go
func (r *Repository) UpsertUser(ctx context.Context, email string, name string) error {
    filter := bson.M{"email": email}
    update := bson.M{
        "$set": bson.M{
            "name":       name,
            "updated_at": time.Now(),
        },
        "$setOnInsert": bson.M{
            "created_at": time.Now(),
        },
    }

    opts := options.Update().SetUpsert(true)
    _, err := r.users().UpdateOne(ctx, filter, update, opts)
    if err != nil {
        return fmt.Errorf("upsert: %w", err)
    }

    return nil
}
```

#### Atomic Operations

```go
func (r *Repository) IncrementUserAge(ctx context.Context, id primitive.ObjectID, delta int) error {
    filter := bson.M{"_id": id}
    update := bson.M{
        "$inc": bson.M{"age": delta},
    }

    result, err := r.users().UpdateOne(ctx, filter, update)
    if err != nil {
        return fmt.Errorf("increment: %w", err)
    }

    if result.MatchedCount == 0 {
        return ErrNotFound
    }

    return nil
}

func (r *Repository) AddUserTag(ctx context.Context, id primitive.ObjectID, tag string) error {
    filter := bson.M{"_id": id}
    update := bson.M{
        "$addToSet": bson.M{"tags": tag}, // Only add if not exists
    }

    _, err := r.users().UpdateOne(ctx, filter, update)
    return err
}

func (r *Repository) RemoveUserTag(ctx context.Context, id primitive.ObjectID, tag string) error {
    filter := bson.M{"_id": id}
    update := bson.M{
        "$pull": bson.M{"tags": tag},
    }

    _, err := r.users().UpdateOne(ctx, filter, update)
    return err
}
```

### Delete Operations

#### Delete Single Document

```go
func (r *Repository) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
    filter := bson.M{"_id": id}

    result, err := r.users().DeleteOne(ctx, filter)
    if err != nil {
        return fmt.Errorf("delete user: %w", err)
    }

    if result.DeletedCount == 0 {
        return ErrNotFound
    }

    return nil
}
```

#### Delete Multiple Documents

```go
func (r *Repository) DeleteInactiveUsers(ctx context.Context, before time.Time) (int64, error) {
    filter := bson.M{
        "last_login": bson.M{"$lt": before},
    }

    result, err := r.users().DeleteMany(ctx, filter)
    if err != nil {
        return 0, fmt.Errorf("delete: %w", err)
    }

    return result.DeletedCount, nil
}
```

## Transactions

### Multi-Document ACID Transactions

MongoDB supports multi-document ACID transactions in replica sets and sharded clusters (MongoDB 4.0+).

```go
func (r *Repository) TransferBalance(ctx context.Context, fromID, toID primitive.ObjectID, amount float64) error {
    session, err := r.client.StartSession()
    if err != nil {
        return fmt.Errorf("start session: %w", err)
    }
    defer session.EndSession(ctx)

    callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
        // Deduct from sender
        filter := bson.M{"_id": fromID, "balance": bson.M{"$gte": amount}}
        update := bson.M{"$inc": bson.M{"balance": -amount}}

        result, err := r.users().UpdateOne(sessCtx, filter, update)
        if err != nil {
            return nil, fmt.Errorf("deduct balance: %w", err)
        }
        if result.MatchedCount == 0 {
            return nil, fmt.Errorf("insufficient balance")
        }

        // Add to receiver
        filter = bson.M{"_id": toID}
        update = bson.M{"$inc": bson.M{"balance": amount}}

        _, err = r.users().UpdateOne(sessCtx, filter, update)
        if err != nil {
            return nil, fmt.Errorf("add balance: %w", err)
        }

        return nil, nil
    }

    _, err = session.WithTransaction(ctx, callback)
    if err != nil {
        return fmt.Errorf("transaction: %w", err)
    }

    return nil
}
```

### Manual Transaction Control

```go
func (r *Repository) CreateOrderWithInventory(ctx context.Context, order Order) error {
    session, err := r.client.StartSession()
    if err != nil {
        return fmt.Errorf("start session: %w", err)
    }
    defer session.EndSession(ctx)

    err = session.StartTransaction()
    if err != nil {
        return fmt.Errorf("start transaction: %w", err)
    }

    // Insert order
    _, err = r.orders().InsertOne(mongo.NewSessionContext(ctx, session), order)
    if err != nil {
        session.AbortTransaction(ctx)
        return fmt.Errorf("insert order: %w", err)
    }

    // Update inventory
    filter := bson.M{"_id": order.ProductID}
    update := bson.M{"$inc": bson.M{"stock": -order.Quantity}}

    result, err := r.db.Collection("inventory").UpdateOne(
        mongo.NewSessionContext(ctx, session),
        filter,
        update,
    )
    if err != nil {
        session.AbortTransaction(ctx)
        return fmt.Errorf("update inventory: %w", err)
    }

    if result.MatchedCount == 0 {
        session.AbortTransaction(ctx)
        return fmt.Errorf("product not found")
    }

    if err := session.CommitTransaction(ctx); err != nil {
        return fmt.Errorf("commit: %w", err)
    }

    return nil
}
```

## Indexes

### Create Index

```go
import "go.mongodb.org/mongo-driver/mongo/options"

func (r *Repository) CreateUserIndexes(ctx context.Context) error {
    indexes := []mongo.IndexModel{
        {
            Keys:    bson.D{{Key: "email", Value: 1}},
            Options: options.Index().SetUnique(true),
        },
        {
            Keys: bson.D{
                {Key: "name", Value: 1},
                {Key: "created_at", Value: -1},
            },
        },
        {
            Keys: bson.D{{Key: "name", Value: "text"}},
        },
    }

    _, err := r.users().Indexes().CreateMany(ctx, indexes)
    if err != nil {
        return fmt.Errorf("create indexes: %w", err)
    }

    return nil
}
```

### Index Types

```go
// Unique index
uniqueIdx := mongo.IndexModel{
    Keys:    bson.D{{Key: "email", Value: 1}},
    Options: options.Index().SetUnique(true),
}

// Compound index
compoundIdx := mongo.IndexModel{
    Keys: bson.D{
        {Key: "status", Value: 1},
        {Key: "created_at", Value: -1},
    },
}

// Text index for full-text search
textIdx := mongo.IndexModel{
    Keys: bson.D{
        {Key: "name", Value: "text"},
        {Key: "description", Value: "text"},
    },
}

// TTL index (auto-delete after expiration)
ttlIdx := mongo.IndexModel{
    Keys:    bson.D{{Key: "expires_at", Value: 1}},
    Options: options.Index().SetExpireAfterSeconds(0),
}

// Partial index (conditional)
partialIdx := mongo.IndexModel{
    Keys: bson.D{{Key: "email", Value: 1}},
    Options: options.Index().
        SetPartialFilterExpression(bson.M{"status": "active"}),
}
```

### List and Drop Indexes

```go
func (r *Repository) ListIndexes(ctx context.Context) ([]string, error) {
    cursor, err := r.users().Indexes().List(ctx)
    if err != nil {
        return nil, fmt.Errorf("list indexes: %w", err)
    }
    defer cursor.Close(ctx)

    var indexes []bson.M
    if err := cursor.All(ctx, &indexes); err != nil {
        return nil, fmt.Errorf("decode: %w", err)
    }

    names := make([]string, 0, len(indexes))
    for _, idx := range indexes {
        names = append(names, idx["name"].(string))
    }

    return names, nil
}

func (r *Repository) DropIndex(ctx context.Context, name string) error {
    _, err := r.users().Indexes().DropOne(ctx, name)
    if err != nil {
        return fmt.Errorf("drop index: %w", err)
    }
    return nil
}
```

## Aggregation Pipeline

### Basic Aggregation

```go
func (r *Repository) GetUserStats(ctx context.Context) ([]bson.M, error) {
    pipeline := mongo.Pipeline{
        {{Key: "$match", Value: bson.M{"status": "active"}}},
        {{Key: "$group", Value: bson.M{
            "_id":   "$country",
            "count": bson.M{"$sum": 1},
            "avgAge": bson.M{"$avg": "$age"},
        }}},
        {{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
        {{Key: "$limit", Value: 10}},
    }

    cursor, err := r.users().Aggregate(ctx, pipeline)
    if err != nil {
        return nil, fmt.Errorf("aggregate: %w", err)
    }
    defer cursor.Close(ctx)

    var results []bson.M
    if err := cursor.All(ctx, &results); err != nil {
        return nil, fmt.Errorf("decode: %w", err)
    }

    return results, nil
}
```

### Complex Aggregation with Lookup (Join)

```go
type OrderWithUser struct {
    ID        primitive.ObjectID `bson:"_id"`
    UserID    primitive.ObjectID `bson:"user_id"`
    Amount    float64            `bson:"amount"`
    Status    string             `bson:"status"`
    User      []User             `bson:"user"`
    CreatedAt time.Time          `bson:"created_at"`
}

func (r *Repository) GetOrdersWithUsers(ctx context.Context, status string) ([]OrderWithUser, error) {
    pipeline := mongo.Pipeline{
        {{Key: "$match", Value: bson.M{"status": status}}},
        {{Key: "$lookup", Value: bson.M{
            "from":         "users",
            "localField":   "user_id",
            "foreignField": "_id",
            "as":           "user",
        }}},
        {{Key: "$unwind", Value: "$user"}},
        {{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}}}},
    }

    cursor, err := r.orders().Aggregate(ctx, pipeline)
    if err != nil {
        return nil, fmt.Errorf("aggregate: %w", err)
    }
    defer cursor.Close(ctx)

    var results []OrderWithUser
    if err := cursor.All(ctx, &results); err != nil {
        return nil, fmt.Errorf("decode: %w", err)
    }

    return results, nil
}
```

### Aggregation Stages Reference

```go
// $match - Filter documents
{{Key: "$match", Value: bson.M{"status": "active"}}}

// $group - Group and aggregate
{{Key: "$group", Value: bson.M{
    "_id":   "$category",
    "total": bson.M{"$sum": "$amount"},
    "count": bson.M{"$sum": 1},
}}}

// $project - Select/transform fields
{{Key: "$project", Value: bson.M{
    "name":  1,
    "email": 1,
    "age":   0, // Exclude
}}}

// $sort - Sort results
{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}}}}

// $limit - Limit results
{{Key: "$limit", Value: 10}}

// $skip - Skip documents
{{Key: "$skip", Value: 20}}

// $lookup - Join collections
{{Key: "$lookup", Value: bson.M{
    "from":         "collection",
    "localField":   "field",
    "foreignField": "field",
    "as":           "alias",
}}}

// $unwind - Deconstruct array
{{Key: "$unwind", Value: "$arrayField"}}
```

## Query Patterns

### Pagination with Cursor

```go
type PaginationParams struct {
    LastID primitive.ObjectID
    Limit  int64
}

func (r *Repository) ListUsersPaginated(ctx context.Context, params PaginationParams) ([]User, error) {
    filter := bson.M{}
    if !params.LastID.IsZero() {
        filter["_id"] = bson.M{"$gt": params.LastID}
    }

    opts := options.Find().
        SetLimit(params.Limit).
        SetSort(bson.D{{Key: "_id", Value: 1}})

    cursor, err := r.users().Find(ctx, filter, opts)
    if err != nil {
        return nil, fmt.Errorf("find: %w", err)
    }
    defer cursor.Close(ctx)

    var users []User
    if err := cursor.All(ctx, &users); err != nil {
        return nil, fmt.Errorf("decode: %w", err)
    }

    return users, nil
}
```

### Projection (Select Fields)

```go
func (r *Repository) GetUserEmails(ctx context.Context) ([]bson.M, error) {
    projection := bson.M{
        "_id":   1,
        "email": 1,
        "name":  1,
    }

    opts := options.Find().SetProjection(projection)

    cursor, err := r.users().Find(ctx, bson.M{}, opts)
    if err != nil {
        return nil, fmt.Errorf("find: %w", err)
    }
    defer cursor.Close(ctx)

    var results []bson.M
    if err := cursor.All(ctx, &results); err != nil {
        return nil, fmt.Errorf("decode: %w", err)
    }

    return results, nil
}
```

### Sorting and Limiting

```go
func (r *Repository) GetTopUsers(ctx context.Context, limit int64) ([]User, error) {
    opts := options.Find().
        SetSort(bson.D{{Key: "score", Value: -1}}).  // Descending
        SetLimit(limit)

    cursor, err := r.users().Find(ctx, bson.M{}, opts)
    if err != nil {
        return nil, fmt.Errorf("find: %w", err)
    }
    defer cursor.Close(ctx)

    var users []User
    return users, cursor.All(ctx, &users)
}
```

### Count Documents

```go
func (r *Repository) CountActiveUsers(ctx context.Context) (int64, error) {
    filter := bson.M{"status": "active"}

    count, err := r.users().CountDocuments(ctx, filter)
    if err != nil {
        return 0, fmt.Errorf("count: %w", err)
    }

    return count, nil
}

func (r *Repository) EstimateUserCount(ctx context.Context) (int64, error) {
    // Faster but less accurate (uses metadata)
    count, err := r.users().EstimatedDocumentCount(ctx)
    if err != nil {
        return 0, fmt.Errorf("estimate: %w", err)
    }

    return count, nil
}
```

## Testing

### Using testcontainers-go

```go
import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/mongodb"
)

func setupMongoDB(t *testing.T) (*mongo.Client, func()) {
    ctx := context.Background()

    mongoContainer, err := mongodb.Run(ctx,
        "mongo:7",
        mongodb.WithUsername("test"),
        mongodb.WithPassword("test"),
    )
    require.NoError(t, err)

    uri, err := mongoContainer.ConnectionString(ctx)
    require.NoError(t, err)

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    require.NoError(t, err)

    cleanup := func() {
        client.Disconnect(ctx)
        mongoContainer.Terminate(ctx)
    }

    return client, cleanup
}

func TestRepository_CreateUser(t *testing.T) {
    client, cleanup := setupMongoDB(t)
    defer cleanup()

    ctx := context.Background()
    repo := NewRepository(client, "testdb")

    user := User{
        Email: "test@example.com",
        Name:  "Test User",
        Age:   30,
    }

    id, err := repo.CreateUser(ctx, user)
    require.NoError(t, err)
    assert.False(t, id.IsZero())

    // Verify user was created
    found, err := repo.GetUser(ctx, id)
    require.NoError(t, err)
    assert.Equal(t, user.Email, found.Email)
    assert.Equal(t, user.Name, found.Name)
}
```

### Seed Test Data

```go
func seedTestData(t *testing.T, repo *Repository) {
    ctx := context.Background()

    users := []User{
        {Email: "user1@test.com", Name: "User One", Age: 25},
        {Email: "user2@test.com", Name: "User Two", Age: 30},
        {Email: "user3@test.com", Name: "User Three", Age: 35},
    }

    _, err := repo.CreateUsers(ctx, users)
    require.NoError(t, err)
}

func TestRepository_ListUsers(t *testing.T) {
    client, cleanup := setupMongoDB(t)
    defer cleanup()

    ctx := context.Background()
    repo := NewRepository(client, "testdb")

    seedTestData(t, repo)

    users, err := repo.ListUsers(ctx, 10)
    require.NoError(t, err)
    assert.Len(t, users, 3)
}
```

### Cleanup Between Tests

```go
func cleanupCollection(t *testing.T, repo *Repository, collection string) {
    ctx := context.Background()

    err := repo.db.Collection(collection).Drop(ctx)
    require.NoError(t, err)
}

func TestRepository_UpdateUser(t *testing.T) {
    client, cleanup := setupMongoDB(t)
    defer cleanup()

    ctx := context.Background()
    repo := NewRepository(client, "testdb")
    defer cleanupCollection(t, repo, "users")

    // Test implementation
    user := User{Email: "test@test.com", Name: "Test"}
    id, err := repo.CreateUser(ctx, user)
    require.NoError(t, err)

    err = repo.UpdateUserName(ctx, id, "Updated Name")
    require.NoError(t, err)

    updated, err := repo.GetUser(ctx, id)
    require.NoError(t, err)
    assert.Equal(t, "Updated Name", updated.Name)
}
```

## Common Mistakes

| Mistake                           | Fix                                           |
|-----------------------------------|-----------------------------------------------|
| Not closing cursor                | Always `defer cursor.Close(ctx)`              |
| Missing context timeout           | Use `context.WithTimeout` for queries         |
| N+1 queries in loops              | Use aggregation pipeline with `$lookup`       |
| Using wrong BSON type             | Use `bson.M` for maps, `bson.D` for ordering  |
| Converting ObjectID to string     | Use `primitive.ObjectIDFromHex()` to convert  |
| No indexes on frequent queries    | Create indexes for common filters             |
| Not checking `MatchedCount`       | Verify updates/deletes affected documents     |
| Using `cursor.Next()` without All | Use `cursor.All()` for simpler code           |
| Forgetting `$` in update ops      | Use `$set`, `$inc`, not direct assignment     |
| No transaction for multi-doc ops  | Use sessions for ACID guarantees              |

## BSON Types Reference

```go
import (
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// bson.M - Map (unordered)
filter := bson.M{
    "age":    bson.M{"$gt": 18},
    "status": "active",
}

// bson.D - Document (ordered, preserves key order)
pipeline := bson.D{
    {Key: "name", Value: 1},
    {Key: "age", Value: -1},
}

// bson.A - Array
tags := bson.A{"golang", "mongodb", "backend"}

// primitive.ObjectID
id := primitive.NewObjectID()
idFromHex, err := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
```

## See Also

- [PostgreSQL](./postgresql.md) - SQL database alternative
- [Migrations](./migrations.md) - Schema migration patterns
- [Redis](./redis.md) - Caching patterns
- [Integration Testing](../08-testing/integration.md) - Testing strategies
- [MongoDB Go Driver Docs](https://www.mongodb.com/docs/drivers/go/current/)
- [BSON Specification](https://bsonspec.org/)
