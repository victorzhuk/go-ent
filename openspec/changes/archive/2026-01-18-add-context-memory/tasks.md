# Tasks: Add Context Memory System

## Dependencies
- Requires: add-config-system (completed)

## 1. Core Infrastructure

- [ ] 1.1 Create `internal/memory/store.go` - Memory storage interface
- [ ] 1.2 Create `internal/memory/sqlite.go` - SQLite-based pattern storage
- [ ] 1.3 Create `internal/memory/session.go` - In-session context tracking
- [ ] 1.4 Create `internal/memory/embeddings.go` - Text embedding for semantic search
- [ ] 1.5 Add memory configuration to config system

## 2. Pattern Learning

- [ ] 2.1 Create `internal/memory/patterns.go` - Pattern extraction from tasks
- [ ] 2.2 Create `internal/memory/capture.go` - Post-task success capture
- [ ] 2.3 Define pattern schema (trigger, context, solution, confidence)
- [ ] 2.4 Implement pattern matching algorithm

## 3. MCP Tools

- [ ] 3.1 Implement `go_ent_memory_store` - Store context/pattern
- [ ] 3.2 Implement `go_ent_memory_search` - Semantic search for relevant context
- [ ] 3.3 Implement `go_ent_memory_recall` - Retrieve specific memories
- [ ] 3.4 Implement `go_ent_memory_forget` - Remove outdated patterns
- [ ] 3.5 Implement `go_ent_memory_stats` - Memory usage statistics

## 4. Compression & Management

- [ ] 4.1 Implement memory summarization for old entries
- [ ] 4.2 Add TTL-based expiration for session memories
- [ ] 4.3 Implement storage limits with LRU eviction
- [ ] 4.4 Add manual cleanup commands

## 5. Integration

- [ ] 5.1 Hook into agent selector to inject relevant patterns
- [ ] 5.2 Add post-task hooks for pattern capture
- [ ] 5.3 Integrate with skill registry for context enrichment

## 6. Testing

- [ ] 6.1 Unit tests for memory storage
- [ ] 6.2 Integration tests for pattern learning
- [ ] 6.3 Test semantic search accuracy
