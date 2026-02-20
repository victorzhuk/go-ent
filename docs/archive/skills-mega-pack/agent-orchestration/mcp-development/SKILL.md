---
name: mcp-development
description: Model Context Protocol server development, tool design, resource management, and integration patterns
---

# MCP (Model Context Protocol) Development

## What is MCP?
- Open protocol for AI agent ↔ tool communication
- Provides: Tools (actions), Resources (data), Prompts (templates)
- Transport: stdio (local) or SSE/HTTP (remote)
- Language SDKs: TypeScript, Python, Go, Rust

## Server Architecture
```typescript
import { Server } from '@modelcontextprotocol/sdk/server/index.js';

const server = new Server({
    name: 'my-mcp-server',
    version: '1.0.0',
}, { capabilities: { tools: {}, resources: {} } });

// Define tools
server.setRequestHandler(ListToolsRequestSchema, async () => ({
    tools: [{
        name: 'search_code',
        description: 'Search codebase for symbols or text',
        inputSchema: {
            type: 'object',
            properties: {
                query: { type: 'string', description: 'Search query' },
                type: { type: 'string', enum: ['symbol', 'text'] },
            },
            required: ['query'],
        },
    }],
}));

server.setRequestHandler(CallToolRequestSchema, async (request) => {
    const { name, arguments: args } = request.params;
    switch (name) {
        case 'search_code': return await searchCode(args);
        default: throw new Error(`Unknown tool: ${name}`);
    }
});
```

## Tool Design Principles
- Clear, descriptive tool names (verb_noun: `search_files`, `create_issue`)
- Detailed descriptions — the AI reads these to decide when to use tools
- Strict input schemas with required/optional fields
- Return structured, parseable output
- Handle errors gracefully — return error messages, don't throw

## Resource Design
- Use URIs for resource identification: `project://files/{path}`
- Support list + read operations
- Include metadata: MIME type, size, last modified
- Implement pagination for large resource lists

## Best Practices
- Keep tools focused — one tool, one job
- Use descriptive error messages the AI can understand
- Implement rate limiting for expensive operations
- Log tool usage for debugging and monitoring
- Version your MCP server for compatibility
- Test tools independently before integrating with agents

## Integration Patterns
- Claude Code: Configure in `.mcp.json` or `mcp_servers` in settings
- OpenCode: Configure in `opencode.json`
- Cursor/VS Code: Configure in settings or `.cursor/mcp.json`
