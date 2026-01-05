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
- Use `/go-ent:plan <description>` to create change proposals
- Use `/go-ent:apply` to execute tasks from the registry
- Agents (`/go-ent:architect`, `/go-ent:dev`, etc.) are available for specialized assistance
- Skills (`go-code`, `go-arch`, `go-api`, etc.) auto-activate for Go code work

### Key Workflow Commands

| Command | Purpose |
|---------|---------|
| `/go-ent:plan` | Full planning workflow (clarify → research → decompose) |
| `/go-ent:apply` | Execute next task from registry |
| `/go-ent:status` | View workflow state and progress |
| `/go-ent:registry list` | Show all tasks across proposals |
| `/go-ent:archive <change-id>` | Archive completed change after deployment |

### Available Agents

| Agent | Purpose | Model |
|-------|---------|-------|
| `/go-ent:lead` | Orchestration and delegation | Opus |
| `/go-ent:architect` | System design and architecture | Opus |
| `/go-ent:planner` | Task breakdown and planning | Sonnet |
| `/go-ent:dev` | Implementation and coding | Sonnet |
| `/go-ent:tester` | Testing and TDD cycles | Haiku |
| `/go-ent:debug` | Bug investigation and troubleshooting | Sonnet |
| `/go-ent:reviewer` | Code review with confidence filtering | Opus |

### Quick Start

1. **Build the MCP server:**
   ```bash
   make build
   ```

2. **Restart Claude Code** to load the plugin

3. **Create a new change:**
   ```
   /go-ent:plan Add new feature description
   ```

4. **Execute tasks:**
   ```
   /go-ent:apply
   ```

5. **Archive when deployed:**
   ```
   /go-ent:archive change-id
   ```

See `docs/DEVELOPMENT.md` for the complete development guide, including:
- Setup instructions
- Development workflows
- Hot-reload vs rebuild guidance
- Bootstrap problem and fallback layers
- Troubleshooting
- Adding new agents, skills, commands, and MCP tools