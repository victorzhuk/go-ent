# CLI Usage Examples

This document provides practical examples of using the `go-ent` CLI commands.

## Table of Contents

- [Getting Started](#getting-started)
- [Configuration Management](#configuration-management)
- [Agent Management](#agent-management)
- [Skill Management](#skill-management)
- [Spec Management](#spec-management)
- [Advanced Usage](#advanced-usage)

## Getting Started

### Check Version

```bash
# Display version information
ent version
```

Output:
```
ent v1.0.0
  go: go1.25.5
```

### Get Help

```bash
# Root help
ent --help

# Command-specific help
ent config --help
ent agent --help
ent spec --help
```

## Configuration Management

### Initialize Configuration

```bash
# Initialize config in current directory
ent config init

# Initialize config in specific directory
ent config init /path/to/project
```

Output:
```
✅ Created config file at .go-ent/config.yaml

Default configuration:
  - Budget: $10.00/day, $200.00/month, $1.00/task
  - Runtime: claude-code (fallback: [cli])
  - Default agent: senior
  - Models: opus, sonnet, haiku
  - Skills: 5 enabled
```

### View Configuration

```bash
# Show full YAML config
ent config show

# Show config from specific directory
ent config show /path/to/project

# Show config summary
ent config show --format summary
```

YAML Output:
```yaml
version: "1.0"
agents:
  default: senior
  roles:
    architect:
      model: opus
      skills:
        - go-arch
        - go-api
budget:
  daily: 10
  monthly: 200
  per_task: 1
  tracking: true
```

Summary Output:
```
# ent Configuration

**Version**: 1.0

## Budget
- Daily: $10.00
- Monthly: $200.00
- Per Task: $1.00
- Tracking: true

## Runtime
- Preferred: claude-code
- Fallback: [cli]

## Agents
- Default: senior
- Roles configured: 3
```

### Modify Configuration

```bash
# Set daily budget
ent config set budget.daily 25

# Set monthly budget
ent config set budget.monthly 500

# Change default agent
ent config set agents.default architect

# Change preferred runtime
ent config set runtime.preferred cli

# Modify in specific directory
ent config set budget.daily 50 /path/to/project
```

Output:
```
✅ Updated budget.daily = 25
```

### Configuration Examples

**Budget Management:**
```bash
# Conservative budget
ent config set budget.daily 5
ent config set budget.monthly 100
ent config set budget.per_task 0.5

# Production budget
ent config set budget.daily 50
ent config set budget.monthly 1000
ent config set budget.per_task 5
```

**Agent Selection:**
```bash
# Use architect for design work
ent config set agents.default architect

# Use developer for implementation
ent config set agents.default developer

# Use senior for balanced work
ent config set agents.default senior
```

## Agent Management

### List Agents

```bash
# List all agents (compact)
ent agent list

# List with details
ent agent list --detailed
```

Compact Output:
```
NAME        MODEL   SKILLS
architect   opus    go-arch, go-api
developer   sonnet  go-code, go-test
senior      sonnet  go-code, go-db, go-test
```

Detailed Output:
```
Agent: architect
  Model: claude-opus-4-5-20251101
  Skills: go-arch, go-api
  Description: System architect for design and architecture decisions

Agent: developer
  Model: claude-sonnet-4-5-20251101
  Skills: go-code, go-test
  Description: Implementation specialist for coding tasks
```

### Get Agent Info

```bash
# Get specific agent details
ent agent info architect
ent agent info developer
ent agent info senior
```

Output:
```
Agent: architect

Model: claude-opus-4-5-20251101
Skills:
  - go-arch: Go architecture and design patterns
  - go-api: API design with OpenAPI and protobuf

Description:
  System architect responsible for high-level design decisions,
  architecture patterns, and API contracts.

Best for:
  - System design
  - API design
  - Architecture decisions
  - Breaking changes
```

## Skill Management

### List Skills

```bash
# List all skills (compact)
ent skill list

# List with details
ent skill list --detailed
```

Compact Output:
```
NAME       DESCRIPTION
go-arch    Go architecture and design patterns
go-api     API design with OpenAPI/protobuf
go-code    Go implementation and coding
go-db      Database integration (PostgreSQL, ClickHouse, Redis)
go-test    Testing with testify and testcontainers
```

Detailed Output:
```
Skill: go-arch
  Description: Go architecture and design patterns
  Auto-activates: architecture decisions, system design
  Tools: Read, Write, Edit, Grep, Glob

Skill: go-code
  Description: Go 1.25+ implementation patterns
  Auto-activates: writing Go code, implementing features
  Tools: All code tools
```

### Get Skill Info

```bash
# Get specific skill details
ent skill info go-arch
ent skill info go-code
ent skill info go-test
```

Output:
```
Skill: go-test

Description:
  Testing patterns with testify, testcontainers, table-driven tests

Auto-activates for:
  - writing tests
  - TDD
  - coverage
  - integration tests
  - mocks

Tools Available:
  - All code manipulation tools
  - Test execution tools
  - Coverage tools

Best Practices:
  - Table-driven tests
  - Use testify/assert and testify/require
  - Real implementations over mocks when simple
  - testcontainers for integration tests
```

## Spec Management

### Initialize OpenSpec

```bash
# Initialize in current directory
ent spec init

# Initialize in specific directory
ent spec init /path/to/project
```

Output:
```
✅ Initialized openspec at .spec

Next steps:
  1. Create specs: openspec/specs/{name}/spec.md
  2. Create changes: openspec/changes/{id}/proposal.md
  3. Run: ent spec list spec
```

### List Specs

```bash
# List all specs
ent spec list spec

# List all changes
ent spec list change

# List from specific directory
ent spec list spec /path/to/project
```

Output:
```
SPECS (3):
  - api
  - database
  - architecture

CHANGES (active: 2):
  - add-authentication
  - refactor-database
```

### Show Spec

```bash
# Show a specific spec
ent spec show spec api

# Show a specific change
ent spec show change add-authentication

# Show from specific directory
ent spec show spec api /path/to/project
```

## Advanced Usage

### Combining Commands

```bash
# Initialize project, create config, list agents
ent spec init && \
  ent config init && \
  ent agent list --detailed

# Check configuration and list available skills
ent config show --format summary && \
  ent skill list
```

### Using Global Flags

```bash
# Verbose output
ent --verbose config show

# Custom config file
ent --config /custom/path/config.yaml agent list

# Combine flags
ent --verbose --config ./my-config.yaml spec list spec
```

### Working with Multiple Projects

```bash
# Project A configuration
ent config show /projects/project-a
ent config set budget.daily 10 /projects/project-a

# Project B configuration
ent config show /projects/project-b
ent config set budget.daily 50 /projects/project-b

# Compare agent setup
ent agent list
```

### Scripting Examples

**Batch Configuration:**
```bash
#!/bin/bash
# Setup multiple projects

for project in project-a project-b project-c; do
  echo "Configuring $project..."
  ent config init "$project"
  ent config set budget.daily 20 "$project"
  ent config set agents.default senior "$project"
done
```

**Configuration Backup:**
```bash
#!/bin/bash
# Backup all configs

timestamp=$(date +%Y%m%d_%H%M%S)
backup_dir="config-backup-$timestamp"
mkdir -p "$backup_dir"

for dir in */; do
  if [ -f "$dir/.go-ent/config.yaml" ]; then
    cp "$dir/.go-ent/config.yaml" "$backup_dir/${dir%/}.yaml"
    echo "Backed up: $dir"
  fi
done

echo "Backup complete in $backup_dir"
```

**Agent Information Report:**
```bash
#!/bin/bash
# Generate agent capabilities report

echo "# Agent Capabilities Report"
echo "Generated: $(date)"
echo ""

for agent in architect developer senior; do
  echo "## Agent: $agent"
  ent agent info "$agent" | sed 's/^/  /'
  echo ""
done
```

### CI/CD Integration

**GitHub Actions:**
```yaml
name: Go-Ent Validation

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install go-ent
        run: |
          curl -L https://github.com/user/go-ent/releases/latest/download/go-ent-linux-amd64 -o /usr/local/bin/go-ent
          chmod +x /usr/local/bin/go-ent

      - name: Validate config
        run: ent config show

      - name: List agents
        run: ent agent list --detailed
```

**GitLab CI:**
```yaml
validate-config:
  stage: test
  script:
    - ent config show
    - ent spec list spec
  only:
    - main
    - develop
```

### Troubleshooting

**Config not found:**
```bash
# Check if config exists
ls -la .go-ent/config.yaml

# Initialize if missing
ent config init

# Verify it works
ent config show
```

**Permission errors:**
```bash
# Check directory permissions
ls -ld .go-ent

# Fix permissions
chmod 755 .go-ent
chmod 644 .go-ent/config.yaml
```

**View default configuration:**
```bash
# Even without init, you can see defaults
ent config show /tmp/nonexistent
```

## Quick Reference

| Command | Description |
|---------|-------------|
| `ent version` | Show version |
| `ent config init` | Initialize config |
| `ent config show` | Display config |
| `ent config set <key> <value>` | Update config |
| `ent agent list` | List agents |
| `ent agent info <name>` | Agent details |
| `ent skill list` | List skills |
| `ent skill info <name>` | Skill details |
| `ent spec init` | Initialize specs |
| `ent spec list <type>` | List specs/changes |
| `ent spec show <type> <id>` | Show spec/change |

## Next Steps

- Read the full [CLI Documentation](./CLI.md)
- Learn about [OpenSpec Workflow](../openspec/README.md)
- Explore [Agent System](./AGENTS.md)
- Review [Configuration Reference](./CONFIG.md)
- Check [Skill Lint CI/CD Integration](./SKILL_LINT_CI.md) for automated validation
