# GraphQL Schema Design Patterns

## Complete GraphQL Schema with Relay Connections

```graphql
type Query {
  posts(first: Int, after: String): PostConnection!
  post(id: ID!): Post
  comments(postId: ID!): [Comment!]!
}

type Mutation {
  createPost(input: CreatePostInput!): Post!
  updatePost(id: ID!, input: UpdatePostInput!): Post!
  deletePost(id: ID!): Boolean!
  createComment(input: CreateCommentInput!): Comment!
}

type Post {
  id: ID!
  title: String!
  content: String!
  author: User!
  comments(first: Int, after: String): CommentConnection!
  createdAt: DateTime!
  updatedAt: DateTime!
}

type Comment {
  id: ID!
  post: Post!
  author: User!
  content: String!
  createdAt: DateTime!
}

type User {
  id: ID!
  email: String!
  name: String!
  posts(first: Int, after: String): PostConnection!
}

type PostConnection {
  edges: [PostEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type PostEdge {
  node: Post!
  cursor: String!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}

input CreatePostInput {
  title: String!
  content: String!
}

input UpdatePostInput {
  title: String
  content: String
}

input CreateCommentInput {
  postId: ID!
  content: String!
}

scalar DateTime
```

**Pattern**: Relay-style connections for pagination, clear input types, nested relationships, DateTime scalar.
