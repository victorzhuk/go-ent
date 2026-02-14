---
name: graphql-development
description: GraphQL API design with schema-first development, resolvers, DataLoader, subscriptions, and security
---

# GraphQL Development

## Schema-First Design
```graphql
type Query {
  user(id: ID!): User
  users(filter: UserFilter, pagination: PaginationInput): UserConnection!
}

type Mutation {
  createUser(input: CreateUserInput!): CreateUserPayload!
  updateUser(id: ID!, input: UpdateUserInput!): UpdateUserPayload!
}

type User {
  id: ID!
  name: String!
  email: String!
  posts(first: Int, after: String): PostConnection!
  createdAt: DateTime!
}

type UserConnection {
  edges: [UserEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

input UserFilter {
  status: UserStatus
  search: String
}
```

## Resolver Patterns
- Keep resolvers thin — delegate to service/data layer
- Use DataLoader for batching N+1 queries
- Implement field-level resolvers for expensive computed fields
- Use context for authentication and request-scoped dependencies

## DataLoader (N+1 Prevention)
```javascript
const userLoader = new DataLoader(async (ids) => {
  const users = await db.users.findMany({ where: { id: { in: ids } } });
  const map = new Map(users.map(u => [u.id, u]));
  return ids.map(id => map.get(id) ?? null);
});
```

## Security
- Implement query depth limiting (max 10 levels)
- Set query complexity analysis with cost limits
- Use persisted queries in production
- Disable introspection in production
- Rate limit by query complexity, not just request count
- Validate and sanitize all input arguments

## Best Practices
- Use Relay Connection spec for pagination
- Return union types for error handling (`... on UserError`)
- Use enums for fixed value sets
- Implement proper null semantics (null = missing, not empty)
- Version through schema evolution, not URL versioning
- Use custom scalars for Date, DateTime, Email, URL
