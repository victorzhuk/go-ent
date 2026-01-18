# Proposal: Add Dynamic MCP Discovery and Management

## Why

Currently, all MCP servers must be pre-configured in Claude Code settings before use. This creates brittleness and requires manual configuration updates whenever new MCP tools are needed.

**Problem**:
- Agents cannot discover MCP servers dynamically
- Every MCP server must be hardcoded in `claude_desktop_config.json`
- No runtime activation/deactivation of MCP tools
- Cannot leverage cloud-hosted MCP servers without local installation

**Solution**:
Enable agents to discover, activate, and manage MCP servers at runtime without manual configuration.

**Inspired by**:
- Docker's dynamic MCP Gateway pattern
- Runtime tool discovery in modern IDEs
- Plugin systems with hot-loading (VS Code, IntelliJ)

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                 CLAUDE CODE / Agent                               │
│  "I need to work with PostgreSQL"                                │
└────────────────────────────┬─────────────────────────────────────┘
                             │
                             │ MCP call: mcp_find("postgres")
                             ▼
┌──────────────────────────────────────────────────────────────────┐
│                      go-ent MCP Server                            │
│                                                                   │
│   ┌─────────────────┐   ┌─────────────────┐   ┌──────────────┐  │
│   │ MCP Discovery   │   │ MCP Manager     │   │ MCP Router   │  │
│   │                 │   │                 │   │              │  │
│   │ - Search local  │   │ - Load/unload   │   │ - Auto-      │  │
│   │ - Search        │   │ - Session-scoped│   │   activate   │  │
│   │   registry      │   │ - Permission    │   │ - Rules      │  │
│   │ - Search        │   │   management    │   │   engine     │  │
│   │   gateway       │   │                 │   │              │  │
│   └─────────────────┘   └─────────────────┘   └──────────────┘  │
│                                                                   │
└───────────────────────┬───────────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               ▼               ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ Local MCP   │  │ Docker MCP  │  │ Custom      │
│ Registry    │  │ Gateway     │  │ Registry    │
│             │  │             │  │             │
│ ~/.mcp/     │  │ Cloud       │  │ Project-    │
│ registry/   │  │ hosted      │  │ specific    │
└─────────────┘  └─────────────┘  └─────────────┘
```

---

## What Changes

### 1. MCP Discovery Tools

| Tool | Purpose | Parameters |
|------|---------|------------|
| `mcp_find` | Search for MCP servers | `query: string` (capability/keyword) |
| `mcp_add` | Activate MCP server in session | `name: string`, `config: object` |
| `mcp_remove` | Deactivate MCP server | `name: string` |
| `mcp_active` | List active MCP servers | None |

#### Example: Dynamic Discovery Workflow

```
Agent: "I need to analyze database schema"

1. Agent calls mcp_find("database schema migration")
   → Returns: mcp-server-postgres, mcp-server-prisma, mcp-server-sqlite

2. Agent calls mcp_add("mcp-server-postgres", {connection: "..."})
   → go-ent launches postgres MCP, registers its tools
   → Tools become available to agent

3. Agent uses postgres MCP tools (query_schema, etc.)

4. Task completes, mcp_remove("mcp-server-postgres") cleans up
```

---

### 2. MCP Registry Sources

**Local Registry** (`~/.mcp-server-registry/`):
```json
{
  "mcp-server-postgres": {
    "name": "PostgreSQL MCP Server",
    "description": "Database schema analysis and migrations",
    "capabilities": ["database", "migration", "schema", "sql"],
    "transport": "stdio",
    "command": "mcp-server-postgres",
    "installed": true
  }
}
```

**Docker MCP Gateway** (cloud-hosted):
```go
type GatewayClient struct {
    apiURL    string  // https://mcp-gateway.docker.com/v1
    apiKey    string
    transport *http.Client
}

func (g *GatewayClient) Search(query string) []RemoteMCP {
    // Query Docker's MCP registry
    // Return cloud-hosted MCP servers
}

func (g *GatewayClient) Proxy(server, tool string, params any) (any, error) {
    // Proxy tool calls to remote MCP
}
```

**Project-Specific Registry** (`.mcp-server-registry/`):
Project-local MCP servers defined in codebase.

---

### 3. Dynamic Tool Selection (Auto-Activation)

**Routing Rules** (`.go-ent/mcp-routing.yaml`):
```yaml
routing:
  # Database tasks → activate postgres MCP
  - match: { keywords: ["database", "migration", "schema"] }
    mcp: mcp-server-postgres
    auto_activate: true

  # Browser tasks → activate playwright MCP
  - match: { keywords: ["browser", "web", "scrape"] }
    mcp: mcp-server-playwright
    auto_activate: true

  # Cloud tasks → check Docker Gateway first
  - match: { keywords: ["cloud", "deploy", "container"] }
    prefer: gateway
```

**Auto-Activation Flow**:
```
User: "Create database migration for new user table"

1. go-ent analyzes keywords: ["database", "migration"]
2. Matches routing rule → mcp-server-postgres
3. Checks if postgres MCP is active
4. If not active and auto_activate: true:
   → Automatically call mcp_add("mcp-server-postgres")
5. Tool becomes available to agent
```

---

### 4. Session-Scoped MCP Management

**Session Lifecycle**:
```go
type MCPSession struct {
    SessionID      string
    ActiveMCPs     map[string]*MCPServer
    ActivatedBy    map[string]ActivationSource  // "manual" | "auto"
}

// When session ends:
func (s *MCPSession) Cleanup() {
    for name, mcp := range s.ActiveMCPs {
        if s.ActivatedBy[name] == "auto" {
            mcp.Shutdown()
        }
        // Manual activations persist for reuse
    }
}
```

**Benefits**:
- Different tasks can use different MCP combinations
- No global state pollution
- Automatic cleanup after task completion

---

### 5. Security & Permission Model

**MCP Approval System**:
```go
type MCPApprovalRequest struct {
    Name         string
    Capabilities []string
    Transport    string
    Source       string  // "local" | "gateway" | "project"
    FirstTime    bool    // Requires user approval on first use
}

func (m *Manager) RequestActivation(req MCPApprovalRequest) error {
    if req.FirstTime {
        // Send approval request to user
        approved := m.requestUserApproval(req)
        if !approved {
            return ErrMCPRejected
        }
    }
    // Proceed with activation
}
```

**Resource Quotas**:
```yaml
# .go-ent/mcp-limits.yaml
quotas:
  max_concurrent_mcps: 5
  max_gateway_mcps: 2
  max_memory_per_mcp: "500MB"
  max_execution_time: "30s"
```

---

## Benefits

1. **No Manual Configuration**: Agents discover and activate MCPs as needed
2. **Gateway Access**: Use cloud MCPs without local installation
3. **Context-Aware Loading**: Only load MCPs relevant to current task
4. **Session Isolation**: Different tasks can use different MCP combinations
5. **Reduced Setup Friction**: New users don't need complex MCP configuration

---

## Implementation

### Phase 1: Core Discovery (MVP)
- `mcp_find` searches local registry
- `mcp_add` / `mcp_remove` for manual activation
- Session-scoped MCP management
- Basic permission system

### Phase 2: Auto-Activation
- MCP routing rules engine
- Keyword-based auto-activation
- Task context analysis

### Phase 3: Gateway Integration
- Docker MCP Gateway client
- Remote MCP proxy
- Cloud MCP registry search

---

## Dependencies

- **add-acp-agent-mode** - Worker management infrastructure (may share patterns)
- MCP protocol knowledge (reuse existing MCP server implementation)

---

## Impact

- Affected specs: mcp-discovery (new capability)
- Affected code: internal/mcp/, cmd/mcp/
- Breaking: None (additive feature)

---

## Example Workflows

### Workflow 1: Manual Discovery

```
User: "Help me work with PostgreSQL"

Agent: I'll search for PostgreSQL MCP servers.
       [Calls mcp_find("PostgreSQL")]

System: Found 2 servers:
        - mcp-server-postgres (local, installed)
        - mcp-server-prisma (gateway, remote)

Agent: I'll activate mcp-server-postgres.
       [Calls mcp_add("mcp-server-postgres", {...})]

System: mcp-server-postgres activated. Tools available:
        - query_schema
        - create_migration
        - execute_sql

Agent: [Uses tools to help user]

[Task completes]

System: [Auto-cleanup via mcp_remove if session ends]
```

### Workflow 2: Auto-Activation

```
User: "Analyze the database schema and suggest improvements"

System: [Detects keywords: "database", "schema"]
        [Matches routing rule for mcp-server-postgres]
        [auto_activate: true]
        [Automatically calls mcp_add("mcp-server-postgres")]
        [No user intervention needed]

Agent: I've activated PostgreSQL tools. Analyzing schema...
       [Uses query_schema tool]
```

### Workflow 3: Gateway Access

```
User: "I need to scrape this website"

Agent: [Calls mcp_find("web scraping")]

System: Found 2 servers:
        - mcp-server-playwright (local, not installed)
        - mcp-server-puppeteer (gateway, cloud-hosted)

Agent: Local playwright not installed. Using cloud gateway.
       [Calls mcp_add("mcp-server-puppeteer", {source: "gateway"})]

System: Connected to cloud MCP via Docker Gateway.
        Tools available: navigate, screenshot, extract_text

Agent: [Proxies tool calls through gateway]
```
