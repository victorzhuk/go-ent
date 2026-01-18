<!-- OPENSPEC:START -->

# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:

- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:

- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

## Self-Hosted Development

This project uses its own plugin system for development (dogfooding).

When working on go-ent itself:

- Use `/ent:plan <description>` to create change proposals
- Use `/ent:apply` to execute tasks from the registry
- Agents (`/ent:architect`, `/ent:dev`, etc.) are available for specialized assistance
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

| Agent            | Purpose                               | Model  |
|------------------|---------------------------------------|--------|
| `/ent:lead`      | Orchestration and delegation          | Opus   |
| `/ent:architect` | System design and architecture        | Opus   |
| `/ent:planner`   | Task breakdown and planning           | Sonnet |
| `/ent:dev`       | Implementation and coding             | Sonnet |
| `/ent:tester`    | Testing and TDD cycles                | Haiku  |
| `/ent:debug`     | Bug investigation and troubleshooting | Sonnet |
| `/ent:reviewer`  | Code review with confidence filtering | Opus   |

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