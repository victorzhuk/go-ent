---
description: Initialize a new Go enterprise project with Clean Architecture structure
argument-hint: <project-name> [module-path] [--type=http|mcp]
---

# Initialize Go Ent Project

Bootstrap a new Go project with Clean Architecture: $ARGUMENTS

## Steps

1. Parse arguments:
   - Project name (required)
   - Module path (optional, defaults to github.com/org/<project-name>)
   - Template type (optional flag: --type=http or --type=mcp, defaults to http)

2. Select template directory based on type:
   - HTTP (--type=http or default): Use `templates/` directory
   - MCP (--type=mcp): Use `templates/mcp/` directory

3. Create directory structure:

   **HTTP Project Structure:**
   ```
   <project>/
   ├── cmd/server/main.go
   ├── internal/
   │   ├── app/
   │   │   ├── app.go
   │   │   ├── contract.go
   │   │   ├── di.go
   │   │   └── uc.go
   │   ├── config/config.go
   │   ├── domain/
   │   │   ├── entity/
   │   │   ├── contract/
   │   │   └── error/
   │   ├── usecase/
   │   ├── repository/
   │   ├── infrastructure/
   │   │   └── database/
   │   └── transport/
   │       └── http/v1/
   │           ├── handler/
   │           ├── dto/
   │           └── middleware/
   ├── database/migrations/
   ├── api/openapi/
   ├── build/Dockerfile
   ├── deploy/docker-compose.yml
   ├── test/integration/
   ├── CLAUDE.md
   ├── Makefile
   ├── .golangci.yml
   ├── .gitignore
   └── go.mod
   ```

   **MCP Project Structure:**
   ```
   <project>/
   ├── cmd/server/main.go
   ├── internal/
   │   └── server/
   │       └── server.go
   ├── build/Dockerfile
   ├── Makefile
   ├── .golangci.yml
   ├── .gitignore
   └── go.mod
   ```

4. Generate files using appropriate templates:
   - HTTP: Bootstrap pattern with app/DI, HTTP handlers, full Clean Architecture
   - MCP: Stdio transport, MCP SDK server setup, tool registration
   - Both: Makefile with VERSION/VCS_REF, distroless Docker, golangci-lint config

5. Initialize git and go modules:
   ```bash
   cd <project>
   go mod init <module-path>
   go mod tidy
   git init
   ```

6. Verify structure:
   ```bash
   make lint  # Should pass
   make test  # Should pass (empty)
   ```

## Generated Files

**Template Selection:**
- HTTP projects: Use templates from `templates/` directory
- MCP projects: Use templates from `templates/mcp/` directory

**Dependencies:**
- HTTP: github.com/caarlos0/env/v11 (config), plus app-specific deps
- MCP: github.com/modelcontextprotocol/go-sdk v1.2.0

**Common Standards:**
- Go 1.25.5
- Distroless Docker images with bash-static
- VERSION and VCS_REF build metadata
- Graceful shutdown with 30s timeout
- Signal handling (SIGTERM, SIGINT, SIGQUIT)
- Structured logging with slog

Follow all enterprise standards: naming, error handling, architecture layers.
