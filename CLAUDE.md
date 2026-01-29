
## Self-Hosted Development

This project uses its own plugin system for development (dogfooding).

**Documentation**: See [docs/INDEX.md](docs/INDEX.md) for complete documentation index.

When working on go-ent itself:

- Use `/ent:plan <description>` to create change proposals
- Use `/ent:apply` to execute tasks from the registry
- Agents (`/ent:architect`, `/ent:coder`, `/ent:tester`, etc.) are available for specialized assistance
- Skills (`go-code`, `go-arch`, `go-api`, etc.) auto-activate for Go code work

### Key Workflow Commands

| Command                    | Purpose                                                 |
|----------------------------|---------------------------------------------------------|
| `/ent:plan`                | Full planning workflow (clarify → research → decompose) |
| `/ent:apply`               | Execute next task from registry                         |
| `/ent:status`              | View workflow state and progress                        |
| `/ent:registry list`       | Show all tasks across proposals                         |
| `/ent:archive <change-id>` | Archive completed change after deployment               |

### Available Agents

| Agent                | Purpose                                     | Model  |
|----------------------|---------------------------------------------|--------|
| `/ent:architect`     | System design and architecture              | Opus   |
| `/ent:planner`       | Task breakdown and planning                 | Sonnet |
| `/ent:planner-fast`  | Quick task assessment and routing           | Haiku  |
| `/ent:planner-heavy` | Deep architectural planning                 | Opus   |
| `/ent:coder`         | Go implementation and feature development   | Sonnet |
| `/ent:tester`        | Test writing and TDD cycles                 | Sonnet |
| `/ent:debugger`      | Standard debugging and investigation        | Sonnet |
| `/ent:debugger-fast` | Quick debugging for simple issues           | Haiku  |
| `/ent:debugger-heavy`| Complex debugging (concurrency, performance)| Opus   |
| `/ent:reviewer`      | Code review for bugs, quality, adherence    | Opus   |
| `/ent:researcher`    | Codebase research and deep code analysis    | Sonnet |
| `/ent:reproducer`    | Create minimal bug reproductions            | Sonnet |
| `/ent:acceptor`      | Validate acceptance criteria and requirements| Sonnet |
| `/ent:decomposer`    | Task breakdown and dependency analysis      | Sonnet |
| `/ent:task-fast`     | Quick task complexity evaluation            | Haiku  |
| `/ent:task-heavy`    | Complex task analysis with deep reasoning   | Opus   |

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

See `docs/DEVELOPMENT.md` for the complete development guide, including:

- Setup instructions
- Development workflows
- Hot-reload vs rebuild guidance
- Bootstrap problem and fallback layers
- Troubleshooting
- Adding new agents, skills, commands, and MCP tools
