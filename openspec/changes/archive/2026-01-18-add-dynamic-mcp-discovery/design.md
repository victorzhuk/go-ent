# Design: Dynamic MCP Discovery Architecture

## Context

Enable runtime discovery and activation of MCP servers without pre-configuration in Claude Code settings.

## Goals

- Discover MCP servers from multiple sources (local, gateway, project)
- Activate/deactivate MCPs within session scope
- Auto-activate MCPs based on task context
- Integrate with Docker MCP Gateway for cloud-hosted servers
- Provide security boundaries and resource quotas

## Non-Goals

- Replacing existing static MCP configuration (both approaches coexist)
- Supporting non-MCP tools/protocols
- Building a full MCP marketplace (just discovery and activation)

---

## Architecture Components

### 1. MCP Discovery Service

```go
type DiscoveryService struct {
    localRegistry   *LocalRegistry
    gatewayClient   *GatewayClient
    projectRegistry *ProjectRegistry
}

type MCPServer struct {
    Name         string
    Description  string
    Capabilities []string
    Transport    string  // "stdio" | "sse" | "http"
    Command      string  // For stdio: command to launch
    Installed    bool
    Source       string  // "local" | "gateway" | "project"
}

func (d *DiscoveryService) Find(query string) []MCPServer {
    // 1. Search local registry (~/.mcp-server-registry/)
    local := d.localRegistry.Search(query)

    // 2. Search project registry (.mcp-server-registry/)
    project := d.projectRegistry.Search(query)

    // 3. Search Docker Gateway (if enabled)
    gateway := d.gatewayClient.Search(query)

    // 4. Merge, dedupe, and rank by relevance
    return d.mergeAndRank(local, project, gateway, query)
}
```

---

### 2. MCP Manager (Session-Scoped)

```go
type MCPManager struct {
    sessions     map[string]*MCPSession
    discovery    *DiscoveryService
    router       *MCPRouter
}

type MCPSession struct {
    SessionID    string
    ActiveMCPs   map[string]*ActiveMCP
    ActivatedBy  map[string]ActivationSource
}

type ActiveMCP struct {
    Name       string
    Server     *MCPServer
    Process    *exec.Cmd      // For stdio transport
    Client     *MCPClient     // MCP protocol client
    Tools      []ToolInfo
    StartedAt  time.Time
}

func (m *MCPManager) Add(sessionID, name string, cfg MCPConfig) error {
    // 1. Validate MCP exists
    server := m.discovery.GetByName(name)
    if server == nil {
        return ErrMCPNotFound
    }

    // 2. Request user approval if first-time
    if !m.isApproved(name) {
        approved := m.requestApproval(server)
        if !approved {
            return ErrMCPRejected
        }
        m.markApproved(name)
    }

    // 3. Launch MCP server
    mcp, err := m.launchMCP(server, cfg)
    if err != nil {
        return fmt.Errorf("launch mcp: %w", err)
    }

    // 4. Register tools with go-ent MCP bridge
    m.registerTools(sessionID, mcp.Tools)

    // 5. Send tools/list_changed notification
    m.notifyToolsChanged()

    return nil
}

func (m *MCPManager) Remove(sessionID, name string) error {
    session := m.sessions[sessionID]
    mcp := session.ActiveMCPs[name]

    // 1. Unregister tools
    m.unregisterTools(sessionID, name)

    // 2. Shutdown MCP server
    if mcp.Process != nil {
        mcp.Process.Kill()
    }

    // 3. Send tools/list_changed notification
    m.notifyToolsChanged()

    return nil
}
```

---

### 3. MCP Router (Auto-Activation)

```go
type MCPRouter struct {
    rules []RoutingRule
}

type RoutingRule struct {
    Match        MatchCondition
    MCP          string  // MCP server name
    AutoActivate bool
    Prefer       string  // "local" | "gateway" | ""
}

type MatchCondition struct {
    Keywords []string
    Pattern  *regexp.Regexp
}

func (r *MCPRouter) SelectForTask(task Task) []string {
    var recommendations []string

    for _, rule := range r.rules {
        if r.matches(rule.Match, task) {
            recommendations = append(recommendations, rule.MCP)
        }
    }

    return recommendations
}

func (r *MCPRouter) AutoActivate(sessionID string, task Task) error {
    mcps := r.SelectForTask(task)

    for _, mcpName := range mcps {
        rule := r.getRuleFor(mcpName)
        if rule.AutoActivate {
            // Auto-activate in background
            go r.manager.Add(sessionID, mcpName, MCPConfig{
                Source: rule.Prefer,
            })
        }
    }

    return nil
}
```

**Routing Rules File** (`.go-ent/mcp-routing.yaml`):
```yaml
routing:
  - match:
      keywords: ["database", "migration", "schema", "sql"]
    mcp: mcp-server-postgres
    auto_activate: true
    prefer: local

  - match:
      keywords: ["browser", "web", "scrape", "html"]
    mcp: mcp-server-playwright
    auto_activate: true
    prefer: local

  - match:
      keywords: ["cloud", "deploy", "container", "docker"]
    mcp: mcp-server-docker
    auto_activate: false  # Require explicit activation
    prefer: gateway
```

---

### 4. Docker MCP Gateway Integration

```go
type GatewayClient struct {
    apiURL    string
    apiKey    string
    transport *http.Client
}

func (g *GatewayClient) Search(query string) []MCPServer {
    req := &GatewaySearchRequest{
        Query: query,
        Limit: 10,
    }

    resp, err := g.transport.Post(
        fmt.Sprintf("%s/search", g.apiURL),
        "application/json",
        jsonEncode(req),
    )

    var servers []MCPServer
    json.NewDecoder(resp.Body).Decode(&servers)

    for i := range servers {
        servers[i].Source = "gateway"
        servers[i].Installed = false  // Cloud-hosted
    }

    return servers
}

func (g *GatewayClient) Proxy(server, tool string, params any) (any, error) {
    req := &GatewayProxyRequest{
        Server: server,
        Tool:   tool,
        Params: params,
    }

    resp, err := g.transport.Post(
        fmt.Sprintf("%s/proxy", g.apiURL),
        "application/json",
        jsonEncode(req),
    )

    var result any
    json.NewDecoder(resp.Body).Decode(&result)

    return result, nil
}
```

---

### 5. MCP Tool Registration Bridge

When an MCP is activated, its tools must be registered with go-ent's MCP server so Claude Code can see them.

```go
type ToolBridge struct {
    mcpServer    *MCPServer  // go-ent's MCP server
    activeMCPs   map[string]*ActiveMCP
}

func (b *ToolBridge) RegisterTools(sessionID string, tools []ToolInfo) {
    for _, tool := range tools {
        // Wrap tool with session-aware handler
        handler := b.createSessionHandler(sessionID, tool)

        // Register with go-ent MCP server
        b.mcpServer.RegisterTool(tool.Name, tool.Schema, handler)
    }

    // Notify Claude Code that tools changed
    b.mcpServer.SendNotification("tools/list_changed", nil)
}

func (b *ToolBridge) createSessionHandler(sessionID string, tool ToolInfo) ToolHandler {
    return func(params any) (any, error) {
        session := b.getSession(sessionID)
        mcp := session.ActiveMCPs[tool.MCPName]

        // Proxy call to actual MCP server
        return mcp.Client.CallTool(tool.Name, params)
    }
}
```

---

## Security Model

### 1. First-Time Approval

```go
type ApprovalRequest struct {
    Name         string
    Description  string
    Capabilities []string
    Transport    string
    Source       string
    Command      string  // For stdio: what will be executed
}

func (m *MCPManager) requestApproval(server *MCPServer) bool {
    req := ApprovalRequest{
        Name:         server.Name,
        Description:  server.Description,
        Capabilities: server.Capabilities,
        Transport:    server.Transport,
        Source:       server.Source,
        Command:      server.Command,
    }

    // Send to user via MCP notification or callback
    return m.sendApprovalRequest(req)
}
```

### 2. Resource Quotas

```go
type ResourceQuotas struct {
    MaxConcurrentMCPs  int
    MaxGatewayMCPs     int
    MaxMemoryPerMCP    int64
    MaxExecutionTime   time.Duration
}

func (m *MCPManager) enforceQuotas(sessionID string) error {
    session := m.sessions[sessionID]

    if len(session.ActiveMCPs) >= m.quotas.MaxConcurrentMCPs {
        return ErrQuotaExceeded{Type: "concurrent_mcps"}
    }

    gatewayCount := 0
    for _, mcp := range session.ActiveMCPs {
        if mcp.Server.Source == "gateway" {
            gatewayCount++
        }
    }

    if gatewayCount >= m.quotas.MaxGatewayMCPs {
        return ErrQuotaExceeded{Type: "gateway_mcps"}
    }

    return nil
}
```

---

## Decisions

### D1: Session-Scoped vs Global MCPs

**Decision**: MCPs are session-scoped by default.

**Rationale**:
- Different tasks may need different MCPs
- Automatic cleanup when session ends
- No global state pollution
- Optional: Allow "pinned" MCPs that persist across sessions

### D2: Auto-Activation

**Decision**: Support auto-activation via routing rules, but require user approval on first use.

**Rationale**:
- Reduces friction for common tasks
- Security: User must approve new MCPs once
- Transparency: Agent should announce auto-activation

### D3: Docker Gateway Integration

**Decision**: Support Docker Gateway as optional feature (requires API key).

**Rationale**:
- Access to cloud MCPs without installation
- Not everyone has Docker MCP account
- Can be disabled via config

---

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Security: Malicious MCPs | First-time approval, sandboxing, quotas |
| Performance: Too many MCPs | Concurrent MCP limit, lazy loading |
| Complexity: Tool name conflicts | Namespace tools by MCP name |
| Gateway availability | Fallback to local, error handling |

---

## Open Questions

1. Should auto-activated MCPs auto-deactivate after task?
   - **Leaning**: Yes, for session-scoped auto-activations

2. How to handle tool name conflicts?
   - **Leaning**: Prefix with MCP name: `postgres:query_schema`

3. Should we cache gateway search results?
   - **Leaning**: Yes, with TTL (5 minutes)
