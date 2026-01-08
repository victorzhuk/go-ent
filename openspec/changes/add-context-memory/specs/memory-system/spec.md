## ADDED Requirements

### Requirement: Session Memory Storage

The system SHALL persist important context within a session for retrieval by subsequent agent invocations.

#### Scenario: Store session context
- **WHEN** `go_ent_memory_store` is called with `scope: "session"`, `key: "auth-patterns"`, `value: "JWT with refresh tokens"`
- **THEN** the context is stored for the current session
- **AND** subsequent agents can retrieve it

#### Scenario: Retrieve session context
- **WHEN** `go_ent_memory_recall` is called with `scope: "session"`, `key: "auth-patterns"`
- **THEN** the stored value is returned
- **AND** access timestamp is updated

#### Scenario: Session context expires
- **WHEN** session terminates
- **THEN** session-scoped memories are cleared
- **AND** storage is released

### Requirement: Project Memory Storage

The system SHALL persist project-level patterns and decisions in SQLite for cross-session retrieval.

#### Scenario: Store project pattern
- **WHEN** `go_ent_memory_store` is called with `scope: "project"`, `type: "pattern"`, `content: {...}`
- **THEN** the pattern is stored in `.go-ent/memory.db`
- **AND** pattern is available in future sessions

#### Scenario: Database initialization
- **WHEN** memory system starts and no database exists
- **THEN** SQLite database is created with schema
- **AND** indexes are created for efficient search

### Requirement: Semantic Memory Search

The system SHALL provide semantic search to find relevant past context.

#### Scenario: Search by similarity
- **WHEN** `go_ent_memory_search` is called with `query: "how to handle authentication errors"`
- **THEN** memories semantically similar to the query are returned
- **AND** results are ranked by relevance score

#### Scenario: Filter by type
- **WHEN** `go_ent_memory_search` is called with `type: "pattern"`
- **THEN** only pattern-type memories are searched

#### Scenario: Limit results
- **WHEN** `go_ent_memory_search` is called with `limit: 5`
- **THEN** at most 5 results are returned

### Requirement: Pattern Learning

The system SHALL capture successful task completion patterns for future reuse.

#### Scenario: Capture success pattern
- **WHEN** task completes successfully
- **AND** pattern capture is enabled
- **THEN** task context, approach, and solution are stored as pattern
- **AND** pattern includes confidence score

#### Scenario: Pattern schema
- **WHEN** pattern is stored
- **THEN** it includes: trigger (when to apply), context (prerequisites), solution (approach), confidence (0.0-1.0)

#### Scenario: Pattern matching
- **WHEN** new task is analyzed
- **THEN** similar patterns are retrieved
- **AND** high-confidence patterns (>0.7) are suggested

### Requirement: Memory Compression

The system SHALL compress old memories to manage storage efficiently.

#### Scenario: Summarize old memories
- **WHEN** memory is older than configured threshold (default 7 days)
- **THEN** detailed content is summarized
- **AND** original is replaced with compressed version

#### Scenario: LRU eviction
- **WHEN** storage limit is exceeded
- **THEN** least recently accessed memories are evicted
- **AND** critical patterns are preserved

#### Scenario: Manual cleanup
- **WHEN** `go_ent_memory_forget` is called with `older_than: "30d"`
- **THEN** memories older than 30 days are removed
- **AND** count of removed items is returned

### Requirement: Memory Statistics

The system SHALL provide visibility into memory usage.

#### Scenario: Get memory stats
- **WHEN** `go_ent_memory_stats` is called
- **THEN** total count, storage size, and breakdown by type are returned

#### Scenario: Stats by scope
- **WHEN** `go_ent_memory_stats` is called with `scope: "project"`
- **THEN** only project-scoped memory stats are returned

### Requirement: Agent Integration

The system SHALL integrate with agent selector to enrich context automatically.

#### Scenario: Auto-inject patterns
- **WHEN** agent is selected for a task
- **THEN** relevant patterns from memory are retrieved
- **AND** high-confidence patterns are included in agent context

#### Scenario: Post-task capture hook
- **WHEN** task completes successfully
- **AND** pattern_capture is enabled in config
- **THEN** success pattern is automatically captured
