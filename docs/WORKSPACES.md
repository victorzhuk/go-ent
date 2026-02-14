# Workspaces

Workspaces enable sharing skills, specs, and configuration across multiple projects — ideal for enterprise teams running related services.

## Concepts

A workspace has two parts:

1. **Shared directory** (version-controlled, team-shared) — skills and openspec at a common path
2. **Per-user XDG state** — workspace config, project registry, and BoltDB cache

## Quick Start

```bash
# Create a workspace
ent workspace init /path/to/team-workspace --name=enterprise-apps

# Add projects
cd /path/to/app-api
ent workspace add --workspace=enterprise-apps

cd /path/to/app-worker
ent workspace add --workspace=enterprise-apps

# List projects
ent workspace list --workspace=enterprise-apps

# Sync specs
ent workspace sync --all

# View workspace info from any project
ent workspace info
```

## Directory Structure

### Workspace shared directory (git-tracked)

```
/workspace-root/
├── skills/                  # Shared skills
│   ├── domain-payments/
│   │   └── SKILL.md
│   └── domain-auth/
│       └── SKILL.md
└── openspec/                # Workspace-level specs
    ├── project.yaml         # Workspace metadata
    ├── specs/               # Shared architecture specs
    └── changes/             # Workspace-level changes
```

### Per-user state (XDG directories)

```
~/.config/go-ent/
├── config.yaml              # Global user config
├── models.yaml              # Global model mappings
└── workspaces.yaml          # Registry: maps workspace names → paths

~/.local/share/go-ent/
└── workspaces/
    └── <workspace-name>/
        ├── config.yaml      # Workspace config
        └── projects.yaml    # Project registry

~/.cache/go-ent/
└── workspaces/
    └── <workspace-name>/
        └── workspace.db     # BoltDB cache
```

### Project enhancement

Each project in a workspace gets a `.go-ent/workspace.yaml`:

```yaml
workspace: enterprise-apps
```

## CLI Commands

### `ent workspace init <path> [--name=<name>]`

Creates a workspace at the given path. Creates `skills/` and `openspec/` directories, registers in XDG workspace registry, and creates per-user config.

### `ent workspace add [project-path] [--workspace=<name>]`

Adds a project to a workspace. Creates `.go-ent/workspace.yaml` in the project and registers it in the workspace's project list.

### `ent workspace list [--workspace=<name>]`

Without `--workspace`, lists all registered workspaces. With `--workspace`, lists projects in that workspace.

### `ent workspace set <key> <value> [--workspace=<name>]`

Sets workspace configuration values. Currently supports `models.<name>` keys.

### `ent workspace info`

Shows workspace details for the current project.

### `ent workspace sync [--project=<name>] [--all] [--dry-run]`

Synchronizes project specs with the workspace:

1. Indexes project specs into workspace BoltDB cache
2. Copies shared workspace specs to projects (with `ws-` prefix)
3. Projects get a unified view of workspace-level architectural decisions

## Skill Resolution

Skills are resolved in priority order (first match wins):

1. Project-local skills (highest priority)
2. Workspace `skills/`
3. Built-in embedded skills

## Config Resolution

Configuration is merged from lowest to highest priority:

1. Built-in defaults
2. Workspace config
3. XDG global config
4. Project config
5. Environment variables (`GOENT_*`)

## Agent Awareness

When a project is in a workspace, agents receive workspace context through three layers:

1. **Static context** — workspace overview is inlined into agent prompts during `ent init`
2. **Dynamic context** — MCP tools (`workspace_specs`, `workspace_projects`) for querying workspace data at runtime
3. **Event-driven** — `onWorkspaceSync` hook triggers prompt regeneration

## XDG Base Directory Standard

go-ent follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/latest/):

| Type | Default Path | Environment Variable |
|------|-------------|---------------------|
| Config | `~/.config/go-ent/` | `XDG_CONFIG_HOME` |
| Data | `~/.local/share/go-ent/` | `XDG_DATA_HOME` |
| Cache | `~/.cache/go-ent/` | `XDG_CACHE_HOME` |

### Migration from `~/.go-ent/`

If `~/.go-ent/` exists, files are automatically migrated to XDG locations on first access. The legacy directory is not removed — verify the migration and remove it manually.
