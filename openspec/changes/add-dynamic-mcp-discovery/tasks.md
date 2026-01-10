# Tasks: Add Dynamic MCP Discovery and Management

## Dependencies

- May share patterns with **add-acp-agent-mode** (worker management)
- Requires MCP protocol knowledge (reuse existing MCP implementation)
- No hard blockers (can proceed independently)

---

## Phase 1: Core Discovery (MVP)

### 1. MCP Discovery Service

- [ ] 1.1 Create `internal/mcp/discovery.go` - MCP discovery service
- [ ] 1.2 Create `internal/mcp/registry.go` - Local registry scanner
- [ ] 1.3 Implement local registry format (`~/.mcp-server-registry/`)
- [ ] 1.4 Implement project registry scanner (`.mcp-server-registry/`)
- [ ] 1.5 Implement search/ranking algorithm (keyword matching, TF-IDF)
- [ ] 1.6 Add MCP metadata validation

### 2. MCP Manager (Session-Scoped)

- [ ] 2.1 Create `internal/mcp/manager.go` - Session-scoped MCP manager
- [ ] 2.2 Create `internal/mcp/session.go` - MCP session state
- [ ] 2.3 Implement MCP lifecycle (launch, monitor, shutdown)
- [ ] 2.4 Implement stdio transport for local MCPs
- [ ] 2.5 Implement tool registration bridge
- [ ] 2.6 Send `tools/list_changed` notifications on activation/deactivation
- [ ] 2.7 Implement session cleanup on session end

### 3. MCP Tools for Claude Code

- [ ] 3.1 Add MCP tool `mcp_find` - Search for MCP servers
  - Parameters: query (string)
  - Returns: list of MCPServer with metadata

- [ ] 3.2 Add MCP tool `mcp_add` - Activate MCP server
  - Parameters: name (string), config (object)
  - Returns: success, available tools list

- [ ] 3.3 Add MCP tool `mcp_remove` - Deactivate MCP server
  - Parameters: name (string)
  - Returns: success

- [ ] 3.4 Add MCP tool `mcp_active` - List active MCPs
  - Returns: list of active MCP servers with status

### 4. Security & Permissions

- [ ] 4.1 Create `internal/mcp/approval.go` - MCP approval system
- [ ] 4.2 Implement first-time approval workflow
- [ ] 4.3 Store approved MCPs in `~/.config/go-ent/mcp-approved.json`
- [ ] 4.4 Implement resource quotas (max concurrent, memory limits)
- [ ] 4.5 Add quota enforcement checks

### 5. Testing

- [ ] 5.1 Unit tests for discovery service
- [ ] 5.2 Unit tests for MCP manager
- [ ] 5.3 Integration tests for tool registration
- [ ] 5.4 Test session cleanup
- [ ] 5.5 Test approval workflow

---

## Phase 2: Auto-Activation

### 6. MCP Router

- [ ] 6.1 Create `internal/mcp/router.go` - MCP routing engine
- [ ] 6.2 Create `internal/mcp/rules.go` - Routing rule definitions
- [ ] 6.3 Load routing rules from `.go-ent/mcp-routing.yaml`
- [ ] 6.4 Implement keyword-based matching
- [ ] 6.5 Implement pattern-based matching (regex)
- [ ] 6.6 Implement auto-activation logic
- [ ] 6.7 Add task context analyzer

### 7. Task Context Analysis

- [ ] 7.1 Extract keywords from user prompt
- [ ] 7.2 Analyze file types in working directory
- [ ] 7.3 Match against routing rules
- [ ] 7.4 Return recommended MCPs

### 8. Testing

- [ ] 8.1 Unit tests for router
- [ ] 8.2 Unit tests for rule matching
- [ ] 8.3 Integration tests for auto-activation
- [ ] 8.4 Test context analyzer

---

## Phase 3: Gateway Integration

### 9. Docker MCP Gateway Client

- [ ] 9.1 Create `internal/mcp/gateway.go` - Gateway client
- [ ] 9.2 Implement gateway search API
- [ ] 9.3 Implement gateway proxy API
- [ ] 9.4 Add authentication (API key)
- [ ] 9.5 Implement caching for search results (TTL: 5 min)
- [ ] 9.6 Handle gateway errors gracefully

### 10. Gateway Configuration

- [ ] 10.1 Load gateway config from `.go-ent/mcp-gateway.yaml`
- [ ] 10.2 Support enabling/disabling gateway
- [ ] 10.3 Support API key via environment variable
- [ ] 10.4 Add gateway health check

### 11. Remote MCP Proxying

- [ ] 11.1 Implement tool call proxying to gateway
- [ ] 11.2 Handle streaming responses from gateway
- [ ] 11.3 Add retry logic for gateway failures
- [ ] 11.4 Track gateway usage/costs

### 12. Testing

- [ ] 12.1 Unit tests for gateway client
- [ ] 12.2 Mock gateway for integration tests
- [ ] 12.3 Test proxy mechanism
- [ ] 12.4 Test fallback to local when gateway unavailable

---

## Phase 4: Polish & Documentation

### 13. User Experience

- [ ] 13.1 Add helpful error messages
- [ ] 13.2 Log MCP activation/deactivation
- [ ] 13.3 Add metrics (activation count, success rate)
- [ ] 13.4 Create example registry entries

### 14. Documentation

- [ ] 14.1 Document MCP discovery workflow
- [ ] 14.2 Document registry format
- [ ] 14.3 Document routing rules format
- [ ] 14.4 Document gateway configuration
- [ ] 14.5 Create examples for common MCPs

---

## Configuration Files

### `.go-ent/mcp-routing.yaml`
```yaml
routing:
  - match:
      keywords: ["database", "migration", "schema"]
    mcp: mcp-server-postgres
    auto_activate: true
    prefer: local

  - match:
      keywords: ["browser", "web"]
    mcp: mcp-server-playwright
    auto_activate: true
```

### `.go-ent/mcp-gateway.yaml`
```yaml
gateway:
  enabled: true
  api_url: https://mcp-gateway.docker.com/v1
  api_key: ${DOCKER_MCP_KEY}

  discovery:
    - pattern: "docker*"
      source: gateway
    - pattern: "*"
      source: local
```

### `~/.mcp-server-registry/<server>.json`
```json
{
  "name": "mcp-server-postgres",
  "description": "PostgreSQL database tools",
  "capabilities": ["database", "migration", "schema", "sql"],
  "transport": "stdio",
  "command": "mcp-server-postgres",
  "installed": true
}
```
