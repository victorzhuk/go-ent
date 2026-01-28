# OpenCode Plugin & Marketplace System

## Overview

Plan to implement OpenCode plugin and marketplace system, making OpenCode compatible with Claude Code's plugin architecture. This enables community plugins that provide custom commands, agents, skills, and MCP tools.

**Status:** Planning Phase
**Author:** Victor Zhuk
**Date:** 2025-01-09

---

## Current State

### What go-ent Already Has

- **`RuntimeOpenCode`** domain type (`internal/domain/runtime.go:14`)
- **Execution infrastructure** (`internal/domain/execution.go`)
- **Multi-runtime support** in config system
- **Skill/agent registry** infrastructure in `internal/skill/`, `internal/agent/`

### What's Needed

- Plugin discovery and loading for OpenCode runtime
- Command/agent/skill parsing (like Claude Code)
- Marketplace integration
- Plugin configuration in `.opencode/` directory

---

## Architecture Overview

### Component 1: OpenCode Plugin Discovery

#### Directory Structure

```
.opencode/
├── marketplace.json          # Global marketplace metadata
├── plugins/                 # Installed plugins
│   └── <plugin-name>/
│       ├── manifest.yaml    # Plugin manifest
│       ├── commands/       # Slash commands
│       ├── agents/        # @-agents
│       ├── skills/        # Skills
│       └── resources/      # Resources (templates, configs)
└── settings.local.json      # Local configuration (permissions, enabled plugins)
```

#### `.opencode/marketplace.json`

Similar to `.claude-plugin/marketplace.json`:

```json
{
  "$schema": "https://opencode.org/plugin/schema.json",
  "name": "go-ent",
  "owner": {
    "name": "Victor Zhuk"
  },
  "metadata": {
    "description": "Enterprise Go development toolkit with Clean Architecture",
    "version": "0.3.0"
  },
  "plugins": [
    {
      "name": "go-ent",
      "source": "./plugins/go-ent",
      "description": "Complete Go enterprise development toolkit",
      "version": "0.3.0",
      "author": {
        "name": "Victor Zhuk",
        "url": "https://github.com/victorzhuk"
      },
      "homepage": "https://github.com/victorzhuk/go-ent",
      "repository": "https://github.com/victorzhuk/go-ent",
      "license": "MIT",
      "keywords": ["go", "golang", "backend", "microservices", "clean-architecture"],
      "category": "development",
      "tags": ["go", "enterprise", "clean-architecture", "ddd", "solid"]
    }
  ]
}
```

#### `.opencode/settings.local.json`

Similar to `.claude/settings.local.json`:

```json
{
  "permissions": {
    "allow": [
      "Bash(*)",
      "Write(*)",
      "Edit(*)",
      "Skill(go-code)",
      "mcp__go_ent__*"
    ]
  },
  "enableAllProjectMcpServers": true,
  "enabledMcpjsonServers": ["go-ent"],
  "enabledPlugins": {
    "go-ent@go-ent": true
  },
  "extraKnownMarketplaces": {
    "go-ent-local": {
      "source": {
        "source": "directory",
        "path": "."
      }
    }
  }
}
```

### Component 2: Plugin Manifest Format

#### `manifest.yaml` (in each plugin directory)

```yaml
name: my-opencode-plugin
version: 1.0.0
description: Custom commands, agents, and skills for OpenCode
author: Organization Name
homepage: https://github.com/org/my-opencode-plugin
repository: https://github.com/org/my-opencode-plugin
license: MIT
category: development
tags: [go, testing, code-review]

# Runtime compatibility
runtime:
  - open-code

# Provided capabilities
provides:
  commands: true
  agents: true
  skills: true
  mcp_tools: true

# Command definitions
commands:
  - name: opencode:plan
    description: Planning workflow
    path: commands/plan.md

# Agent definitions
agents:
  - name: opencode:architect
    description: System architect
    role: architect
    model: opus
    path: agents/architect.md
    skills: [go-arch, go-api]

# Skill definitions
skills:
  - name: go-code-opencode
    description: Go patterns for OpenCode
    path: skills/go-code/SKILL.md
    triggers: ["go code", "implement", "refactor"]
```

### Component 3: Plugin Loading System

#### Package: `internal/opencode/plugin/`

#### `loader.go`

```go
type Loader struct {
    configDir    string
    pluginDir    string
    cmdRegistry  *CommandRegistry
    agentRegistry *agent.Registry
    skillRegistry *skill.Registry
}

// Load discovers and loads all plugins
func (l *Loader) Load(ctx context.Context) error

// LoadPlugin loads a single plugin
func (l *Loader) LoadPlugin(ctx context.Context, pluginPath string) (*Plugin, error)

// ParseCommands parses commands from plugin
func (l *Loader) ParseCommands(pluginDir string) ([]Command, error)

// ParseAgents parses agents from plugin
func (l *Loader) ParseAgents(pluginDir string) ([]Agent, error)

// ParseSkills parses skills from plugin
func (l *Loader) ParseSkills(pluginDir string) ([]Skill, error)
```

#### Command Parsing (similar to Claude Code)

- Read markdown files from `commands/*.md`
- Parse YAML frontmatter: `name`, `description`, `allowed-tools`
- Register as slash commands (e.g., `/opencode:plan`)

#### Agent Parsing

- Read markdown files from `agents/*.md`
- Parse YAML frontmatter: `name`, `description`, `tools`, `model`, `skills`, `role`
- Register as @-agents (e.g., `@opencode:architect`)

#### Skill Parsing

- Read markdown files from `skills/*/SKILL.md`
- Parse YAML frontmatter: `name`, `description`
- Parse triggers from `description` (e.g., "Auto-activates for: writing Go code")
- Register in skill registry

### Component 4: Command Registry for OpenCode

#### Package: `internal/opencode/command/`

#### `registry.go`

```go
type Registry struct {
    commands map[string]Command
}

// Register registers a command
func (r *Registry) Register(name string, cmd Command) error

// Get retrieves a command
func (r *Registry) Get(name string) (Command, error)

// List returns all commands
func (r *Registry) List() []Command

// Find searches commands by query
func (r *Registry) Find(query string) []Command
```

#### Command Structure

```go
type Command struct {
    Name        string
    Description string
    Category    string
    AllowedTools []string
    Handler     func(ctx context.Context, args []string) error
}
```

### Component 5: Agent Registry for OpenCode

#### Package: `internal/opencode/agent/` (extend existing)

#### `registry.go`

```go
type Registry struct {
    agents map[string]AgentConfig
}

// LoadFromPlugins loads agents from plugin manifests
func (r *Registry) LoadFromPlugins(plugins []Plugin) error

// SelectForTask selects agent for task
func (r *Registry) SelectForTask(ctx context.Context, task string) (AgentConfig, error)

// BuildPrompt builds agent prompt with skills
func (r *Registry) BuildPrompt(role AgentRole, skills []string) (string, error)
```

### Component 6: Marketplace Client

#### Package: `internal/marketplace/`

#### `client.go`

```go
type Client struct {
    baseURL    string
    httpClient  *http.Client
    authToken   string
}

// Search finds plugins
func (c *Client) Search(ctx context.Context, query string, filters SearchFilters) ([]PluginInfo, error)

// Download downloads plugin
func (c *Client) Download(ctx context.Context, name string, version string) ([]byte, error)

// Publish uploads plugin
func (c *Client) Publish(ctx context.Context, pluginPath string) error
```

#### Search Features

- Query by name, description, keywords
- Filter by category, language
- Sort by downloads, rating
- Pagination support

### Component 7: Plugin Manager

#### Package: `internal/opencode/plugin/` (extend existing)

#### `manager.go`

```go
type Manager struct {
    loader      *Loader
    marketplace *marketplace.Client
    config      *Config
}

// Install installs plugin from marketplace or local path
func (m *Manager) Install(ctx context.Context, source string) error

// Uninstall removes plugin
func (m *Manager) Uninstall(ctx context.Context, name string) error

// List returns all installed plugins
func (m *Manager) List() []PluginInfo

// Update updates plugin
func (m *Manager) Update(ctx context.Context, name string) error
```

### Component 8: Integration with Config

#### Extend `internal/config/config.go`

```go
type Config struct {
    // ... existing fields ...

    // OpenCode plugin configuration
    OpenCode OpenCodeConfig `yaml:"opencode"`
}

type OpenCodeConfig struct {
    PluginDir    string `yaml:"plugin_dir"`     // default: ".opencode/plugins"
    MarketplaceURL string `yaml:"marketplace_url"`
    AutoUpdate     bool   `yaml:"auto_update"`
    TrustedSources []string `yaml:"trusted_sources"`
}
```

#### Extend `internal/config/loader.go`

```go
func Load(projectRoot string) (*Config, error) {
    // ... existing load logic ...

    opencodeConfig := &OpenCodeConfig{
        PluginDir:    filepath.Join(projectRoot, ".opencode/plugins"),
        MarketplaceURL: "https://opencode.org/marketplace",
        AutoUpdate:     false,
        TrustedSources: []string{"https://opencode.org"},
    }

    cfg.OpenCode = *opencodeConfig

    return cfg, nil
}
```

### Component 9: OpenCode Runtime Execution

#### Package: `internal/execution/` (extend existing)

#### `opencode.go`

```go
type OpenCodeRunner struct {
    config   domain.RuntimeConfig
    cmdReg   *opencode.CommandRegistry
    agentReg *opencode.AgentRegistry
}

// Execute runs command in OpenCode runtime
func (r *OpenCodeRunner) Execute(ctx context.Context, cmd string, args []string) (*ExecutionResult, error)

// ExecuteAgent runs agent in OpenCode runtime
func (r *OpenCodeRunner) ExecuteAgent(ctx context.Context, agent domain.AgentRole, task string) (*ExecutionResult, error)

// ExecuteSkill runs skill in OpenCode runtime
func (r *OpenCodeRunner) ExecuteSkill(ctx context.Context, skill string, input SkillRequest) (*SkillResult, error)
```

---

## Implementation Phases

### Phase 1: Plugin Discovery (Week 1-2)

**Tasks:**

1. Create `internal/opencode/plugin/loader.go`
2. Implement command parsing from markdown
3. Implement agent parsing from markdown
4. Implement skill parsing from markdown
5. Create plugin manifest schema (`manifest.yaml`)
6. Write tests for plugin loading

**Deliverables:**
- Plugin loading infrastructure
- Markdown parser for commands/agents/skills
- Unit tests

### Phase 2: Registries (Week 3)

**Tasks:**

1. Create `internal/opencode/command/registry.go`
2. Extend `internal/opencode/agent/registry.go`
3. Implement registration for commands/agents/skills
4. Add support for @-agents and slash commands
5. Integration tests

**Deliverables:**
- Command registry
- Agent registry (extended)
- Integration tests

### Phase 3: Marketplace Client (Week 4-5)

**Tasks:**

1. Create `internal/marketplace/client.go`
2. Implement search functionality
3. Implement download functionality
4. Add checksum verification
5. Add authentication support
6. Tests with mock marketplace

**Deliverables:**
- Marketplace client
- Search/download functionality
- Mock marketplace server for testing

### Phase 4: Plugin Manager (Week 6)

**Tasks:**

1. Extend `internal/opencode/plugin/manager.go`
2. Implement install/uninstall/list/update
3. Add plugin validation
4. Add dependency checking
5. Conflict detection
6. Integration tests

**Deliverables:**
- Plugin manager
- Plugin lifecycle management
- Validation and conflict detection

### Phase 5: Configuration Integration (Week 7)

**Tasks:**

1. Extend `internal/config/config.go` with OpenCodeConfig
2. Extend `internal/config/loader.go` for loading
3. Add environment variable support (`OPENCODE_PLUGIN_DIR`, etc.)
4. Create example `.opencode/settings.local.json`
5. Update examples

**Deliverables:**
- Configuration support
- Environment variable integration
- Example configuration

### Phase 6: Execution Integration (Week 8)

**Tasks:**

1. Create `internal/execution/opencode.go`
2. Implement command execution
3. Implement agent execution
4. Implement skill execution
5. Integrate with existing execution engine
6. End-to-end tests

**Deliverables:**
- OpenCode runtime execution
- Command/agent/skill execution
- End-to-end tests

### Phase 7: Documentation (Week 9)

**Tasks:**

1. Create `docs/OPENCODE_PLUGINS.md`
2. Document plugin development
3. Document manifest format
4. Document marketplace usage
5. Update README with OpenCode support

**Deliverables:**
- Plugin development guide
- Manifest format reference
- Marketplace usage guide
- Updated README

### Phase 8: Migration & Examples (Week 10)

**Tasks:**

1. Migrate `plugins/go-ent` to OpenCode format
2. Create example plugin (simple skills)
3. Create example plugin (agents + commands)
4. Write migration guide
5. Test migration process

**Deliverables:**
- Migrated go-ent plugin
- Example plugins
- Migration guide
- Tests

---

## Key Design Decisions

### Q1: Plugin Storage Location?

**Options:**
- A) `.opencode/plugins/` (new directory, like Claude Code's `.claude-plugin/`)
- B) Reuse `plugins/` directory (shared between runtimes)

**Recommendation:** Use `.opencode/plugins/`

**Rationale:**
- Clear separation between Claude Code and OpenCode runtimes
- Follows Claude Code's pattern
- Easier to manage runtime-specific plugins

### Q2: Command Ambiguity?

**Problem:** `go-ent` has commands like `/go-ent:plan`. If OpenCode uses same format, how to distinguish?

**Options:**
- A) Different prefix: `/opencode:plan`
- B) Same prefix, different config file
- C) Allow configurable prefix

**Recommendation:** Same prefix, different config file

**Rationale:**
- Commands work identically across runtimes
- User experience is consistent
- `.claude/` vs `.opencode/` handles configuration

### Q3: Marketplace Hosting?

**Options:**
- A) Separate service (opencode.org)
- B) GitHub-based (like go-ent)
- C) Decentralized (IPFS, IPNS)

**Recommendation:** GitHub-based initially, option for external marketplace

**Rationale:**
- Faster to implement
- Uses existing infrastructure
- Can add external marketplace later
- go-ent can be first plugin!

### Q4: Plugin Validation?

**Options:**
- A) Strict validation (fail fast)
- B) Lenient validation (warn but load)
- C) User-controlled validation level

**Recommendation:** Strict validation by default, user override option

**Rationale:**
- Security best practice
- Prevents bad plugins
- Advanced users can disable if needed

---

## Benefits

1. **Unified Plugin System:** Same plugin format works for Claude Code and OpenCode
2. **Marketplace Distribution:** Easy sharing and discovery of plugins
3. **Runtime Agnostic:** Plugins work in any supported runtime
4. **Existing Infrastructure:** Leverages go-ent's agent/skill registries
5. **Backward Compatible:** Doesn't break existing go-ent functionality
6. **Extensible:** Easy to add new runtime support in future

---

## Dependencies

### Existing Dependencies

- `internal/domain/` - Agent, skill, runtime, execution types
- `internal/mcp/` - MCP server infrastructure
- `internal/config/` - Configuration system
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/modelcontextprotocol/go-sdk/mcp` - MCP SDK

### New Dependencies (TBD)

- HTTP client for marketplace (standard library or custom)
- Checksum verification (crypto/sha256)

---

## Open Questions

1. **Runtime Priority:** Should OpenCode be the default preferred runtime or keep Claude Code as default?

2. **Plugin Sharing:** Should go-ent plugin work in both Claude Code and OpenCode automatically, or require separate builds?

3. **Marketplace MVP:** Should we:
   - Start with GitHub-only distribution?
   - Build a dedicated marketplace service immediately?
   - Support both from day 1?

4. **Command Prefix:** Use `/go-ent:` for both runtimes, or `/opencode:` for OpenCode-specific commands?

5. **Migration Strategy:** Should we:
   - Auto-migrate existing plugins on first run?
   - Provide manual migration tools?
   - Keep both formats supported indefinitely?

---

## Success Criteria

- [ ] Plugin discovery system working for `.opencode/plugins/`
- [ ] Commands parse from markdown and register correctly
- [ ] Agents parse from markdown and register correctly
- [ ] Skills parse from markdown and register correctly
- [ ] Plugin manager can install/uninstall/list/update plugins
- [ ] Marketplace client can search and download plugins
- [ ] Configuration system supports OpenCode plugins
- [ ] Execution engine runs commands/agents/skills in OpenCode runtime
- [ ] go-ent plugin works in both Claude Code and OpenCode
- [ ] Documentation complete
- [ ] All tests passing

---

## Timeline Estimate

- **Phase 1 (Discovery):** 2 weeks
- **Phase 2 (Registries):** 1 week
- **Phase 3 (Marketplace):** 2 weeks
- **Phase 4 (Plugin Manager):** 1 week
- **Phase 5 (Config):** 1 week
- **Phase 6 (Execution):** 1 week
- **Phase 7 (Documentation):** 1 week
- **Phase 8 (Migration):** 1 week

**Total:** 10 weeks

---

## References

- Claude Code Plugin Documentation: https://docs.anthropic.com/en/docs/claude-code/plugins
- go-ent Existing Plugin: `plugins/go-ent/`
- MCP Protocol Specification: https://modelcontextprotocol.io
- OpenSpec System: `openspec/AGENTS.md`
- Domain Types: `internal/domain/`
