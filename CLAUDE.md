
## Self-Hosted Development

This project uses its own plugin system for development (dogfooding).

**Documentation**: See [docs/INDEX.md](docs/INDEX.md) for complete documentation index.

When working on go-ent itself:

- Use `/ent:plan <description>` to create change proposals
- Use `/ent:apply` to execute tasks from the registry
- Skills (`go-code`, `go-arch`, `go-api`, `task-router`, etc.) auto-activate based on task content

### Key Workflow Commands

| Command                    | Purpose                                                 |
|----------------------------|---------------------------------------------------------|
| `/ent:plan`                | Full planning workflow (clarify -> research -> decompose) |
| `/ent:apply`               | Execute next task from registry                         |
| `/ent:status`              | View workflow state and progress                        |
| `/ent:registry list`       | Show all tasks across proposals                         |
| `/ent:archive <change-id>` | Archive completed change after deployment               |

### Skills Auto-Activate

Skills activate automatically based on task content. No manual invocation needed.

### Model Routing (via Task tool)

When spawning subagents for parallel or isolated work:

| Task Class | Model | Agent Type |
|---|---|---|
| Exploration, triage | haiku | Explore |
| Implementation, testing, planning | sonnet | general-purpose |
| Architecture, review, deep debug | opus | general-purpose or Explore |

See the `task-router` skill for the full routing table and invocation patterns.

### Quick Start

1. **Build the MCP server:**
   ```bash
   make build
   ```

2. **Restart Claude Code** to load the plugin

3. **Create a new change:**
   ```
   /ent:plan Add new feature description
   ```

4. **Execute tasks:**
   ```
   /ent:apply
   ```

5. **Archive when deployed:**
   ```
   /ent:archive change-id
   ```

See [docs/INDEX.md](docs/INDEX.md) for full documentation including:

- Setup instructions
- Development workflows
- Adding new skills, commands, and MCP tools
