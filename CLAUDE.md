
## Self-Hosted Development

This project uses its own plugin system for development (dogfooding).

**Documentation**: See [docs/INDEX.md](docs/INDEX.md) for complete documentation index.

When working on go-ent itself, use `/opsx:` slash commands — the plugin is loaded via `ent init`.

### Key Workflow Commands

| Command | Purpose |
|---------|---------|
| `/opsx:explore` | Think through ideas, investigate problems |
| `/opsx:new` | Start a new change |
| `/opsx:continue` | Continue working on a change |
| `/opsx:apply` | Execute tasks from change |
| `/opsx:ff` | Fast-forward all artifacts |
| `/opsx:sync` | Sync delta specs |
| `/opsx:archive` | Archive completed change |
| `/opsx:verify` | Verify implementation |

### Skills Auto-Activate

Skills activate automatically based on task content. No manual invocation needed.

### Model Routing (via Task tool)

When spawning subagents for parallel or isolated work:

| Task Class | Model | Agent Type |
|---|---|---|
| Exploration, triage | haiku | Explore |
| Implementation, testing, planning | sonnet | general-purpose |
| Architecture, review, deep debug | opus | general-purpose or Explore |

### Quick Start

1. **Build the MCP server:**
   ```bash
   make build
   ```

2. **Install into Claude Code:**
   ```bash
   ent init --tools=claude
   ```

3. **Restart Claude Code** to load the plugin

4. **Create a new change:**
   ```
   /opsx:new Add new feature description
   ```

5. **Execute tasks:**
   ```
   /opsx:apply
   ```

6. **Archive when deployed:**
   ```
   /opsx:archive change-id
   ```

See [docs/INDEX.md](docs/INDEX.md) for full documentation including:

- Setup instructions
- Development workflows
- Adding new skills, commands, and MCP tools
