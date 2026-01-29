# Migration Guide: v2 to v3

Complete guide for migrating go-ent plugins from v2 to v3 format (Claude Code native).

---

## Table of Contents

- [Overview](#overview)
- [Breaking Changes](#breaking-changes)
- [Agent Migration](#agent-migration)
- [Skill Migration](#skill-migration)
- [Command Migration](#command-migration)
- [Backward Compatibility](#backward-compatibility)
- [Migration Checklist](#migration-checklist)

---

## Overview

### What Changed

go-ent v3 aligns with Claude Code official patterns:

| Component | v2 Format | v3 Format |
|-----------|-----------|-----------|
| **Agents** | Split YAML + MD files | Single MD with YAML frontmatter |
| **Skills** | XML tags (`<role>`) | Markdown sections (`## Role`) |
| **Tool Presets** | YAML `toolPresets:` | Reference skills (preloaded) |
| **Shared Prompts** | Included via template | Reference skills (preloaded) |
| **Commands** | Mixed root/flows | Organized: workflows/, aliases/, utilities/ |

### Benefits

✅ **Claude Code alignment** - Uses official agent/skill format
✅ **Better maintainability** - Single source of truth
✅ **Cleaner structure** - Logical directory organization
✅ **Reference skills** - Reusable skill preloading
✅ **Backward compatible** - v2 skills still work

---

## Breaking Changes

### Removed Features

❌ **Split agent files** - No more `meta/` + `prompts/` separation
❌ **XML skill tags** - Replaced with Markdown sections
❌ **Tool preset syntax** - Converted to reference skills
❌ **Root command files** - Moved to workflows/ directory

### Removed Directories

```
plugins/go-ent/
├── agents/
│   ├── meta/           ❌ Removed
│   ├── prompts/        ❌ Removed
│   └── presets/        ❌ Removed
└── commands/
    ├── flows/          ❌ Renamed to workflows/
    └── domains/        ❌ Renamed to context/
```

### New Structure

```
plugins/go-ent/
├── .claude/
│   └── agents/         ✅ New: Claude Code native location
│       └── *.md        (16 agents)
├── skills/
│   ├── ent/           ✅ New: Reference skills (10 files)
│   ├── core/          (5 skills)
│   └── go/            (12 skills)
├── commands/
│   ├── workflows/     ✅ Renamed from flows/
│   ├── aliases/       ✅ New: OpenSpec aliases
│   ├── utilities/     ✅ New: Standalone tools
│   └── context/       ✅ Renamed from domains/
└── schemas/
    └── agent.schema.json  ✅ Unified schema
```

---

## Agent Migration

### Before (v2 - Split Files)

**meta/coder.yaml:**
```yaml
name: ent:coder
description: Go developer. Implements features, writes code.
model: main
color: "#32CD32"
skills:
  - go-code
  - go-db
toolPresets:
  - execution-full
disallowedToolPresets:
  - serena-editing
```

**prompts/agents/coder.md:**
```markdown
You are a senior Go backend developer.

## Responsibilities
- Implement features from tasks.md
- Write production-quality Go code
```

### After (v3 - Single File)

**.claude/agents/coder.md:**
```markdown
---
name: coder
description: Go developer. Implements features, writes code.
model: sonnet
skills:
  - go-code
  - go-db
  - ent-tools-editing
  - ent-tooling
  - ent-conventions
  - ent-handoffs
  - ent-judgment
  - ent-principals
  - ent-openspec
disallowedTools:
  - mcp__plugin_serena_serena__replace_symbol_body
  - mcp__plugin_serena_serena__insert_after_symbol
  - mcp__plugin_serena_serena__insert_before_symbol
  - mcp__plugin_serena_serena__replace_content
  - mcp__plugin_serena_serena__create_text_file
color: "#32CD32"
role: execution
complexity: standard
dependencies:
  - tester
  - reviewer
  - debugger
---

You are a senior Go backend developer.

## Responsibilities
- Implement features from tasks.md
- Write production-quality Go code
```

### Key Changes

1. **Single file** - Metadata + prompt combined
2. **Name without prefix** - `coder` not `ent:coder`
3. **Model normalization** - `main` → `sonnet`
4. **Tool presets → Reference skills**:
   - `execution-full` → `ent-tools-editing`
   - Shared prompts → `ent-tooling`, `ent-conventions`, etc.
5. **Explicit disallowed tools** - Serena editing tools listed individually
6. **Dependencies without prefix** - `tester` not `ent:tester`

---

## Skill Migration

### Before (v2 - XML Tags)

```markdown
---
name: go-error
description: "Error handling patterns"
tags: ["go", "error"]
---

<triggers>
  keywords:
    - "error handling"
  weight: 0.8
</triggers>

# Go Error Handling

<role>
Expert Go error handling engineer specializing in error design patterns.
</role>

<instructions>
## Error Handling Stack
- **Error Wrapping** — fmt.Errorf with %w
</instructions>

<constraints>
- Include proper error wrapping
- Exclude unwrapped errors
</constraints>
```

### After (v3 - Markdown Sections)

```markdown
---
name: go-error
description: "Error handling patterns"
tags: ["go", "error"]
triggers:
  keywords:
    - error handling
  file_pattern: "*.go"
  weight: 0.8
---

## Role

Expert Go error handling engineer specializing in error design patterns.

## Instructions

## Error Handling Stack
- **Error Wrapping** — fmt.Errorf with %w

## Constraints

- Include proper error wrapping
- Exclude unwrapped errors
```

### Key Changes

1. **Triggers in frontmatter** - Moved from body to YAML
2. **Markdown sections** - `## Role` instead of `<role>`
3. **Structured triggers** - Object with `keywords`, `file_pattern`, `weight`
4. **No title needed** - Section headings suffice

---

## Command Migration

### Before (v2 - Mixed Location)

```
commands/
├── plan.md             # Duplicate
├── task.md             # Duplicate
├── bug.md              # Duplicate
├── flows/
│   ├── plan.md         # Actual implementation
│   ├── task.md
│   └── bug.md
└── opsx-*.md           # Root level
```

### After (v3 - Organized)

```
commands/
├── workflows/          # Full workflows
│   ├── plan.md
│   ├── task.md
│   └── bug.md
├── aliases/           # OpenSpec aliases
│   ├── opsx-new.md
│   ├── opsx-apply.md
│   └── opsx-archive.md
├── utilities/         # Standalone tools
│   └── skill-sync.md
└── context/           # Shared context
    ├── generic.md
    └── openspec.md
```

---

## Backward Compatibility

### Skill Format Support

| Format | Support | Removal |
|--------|---------|---------|
| v3 (Markdown sections) | ✅ Primary | N/A |
| v2 (XML tags) | ✅ Full support | v4.0 |
| v1 (Basic frontmatter) | ✅ Full support | v4.0 |

### Skill Parser Behavior

The parser automatically detects format version:

```go
// v3: Markdown sections + triggers in frontmatter
if hasMarkdownSections && hasFrontmatterTriggers {
    return "v3"
}

// v2: XML tags
if hasXMLTags {
    return "v2"
}

// v1: Basic frontmatter
return "v1"
```

### Agent Format Support

| Format | Support | Removal |
|--------|---------|---------|
| Single MD (v3) | ✅ Primary | N/A |
| Split YAML+MD (v2) | ❌ Removed | v3.0 |

**Note**: Agent split-file format was removed in v3.0. Use the migration script to convert.

---

## Migration Checklist

### For Plugin Authors

- [ ] **Backup your plugin** before migration
- [ ] **Run migration scripts**:
  ```bash
  go run scripts/migrate-agents.go
  go run scripts/migrate-skills.go
  ```
- [ ] **Verify agent files** in `.claude/agents/`
- [ ] **Test skill parsing**:
  ```bash
  go test ./internal/skill -v
  ```
- [ ] **Update documentation** references
- [ ] **Remove old directories**:
  ```bash
  rm -rf plugins/*/agents/meta
  rm -rf plugins/*/agents/prompts
  rm -rf plugins/*/agents/presets
  ```
- [ ] **Test in Claude Code**:
  ```bash
  make build
  # Restart Claude Code
  /agents  # Should list all agents
  ```

### For End Users

No action required! The plugin maintainer handles migration.

### For Contributors

- [ ] Read [AGENTS_AND_SKILLS.md](./AGENTS_AND_SKILLS.md) for new format
- [ ] Read [SKILL-AUTHORING.md](./SKILL-AUTHORING.md) for v3 skills
- [ ] Follow [CLAUDE_CODE_COMPATIBILITY.md](./CLAUDE_CODE_COMPATIBILITY.md)
- [ ] Use new agent template from `schemas/agent.example.md`

---

## Troubleshooting

### Agents Not Loading

**Problem**: Agents don't appear in `/agents` list

**Solution**:
1. Check file location: `.claude/agents/*.md`
2. Verify YAML frontmatter is valid
3. Ensure `name` and `description` fields exist
4. Restart Claude Code to reload plugins

### Skills Not Activating

**Problem**: Skills not triggering when expected

**Solution**:
1. Check `triggers:` in frontmatter for v3 skills
2. Verify keywords match your use case
3. Check `file_pattern` if specified
4. Test trigger detection:
   ```bash
   go run scripts/test-skill-triggers.go
   ```

### Backward Compatibility Issues

**Problem**: v2 skills not working

**Solution**:
1. Parser should auto-detect v2 format
2. Check parser version detection:
   ```bash
   go test ./internal/skill -run TestParser_detectVersion
   ```
3. Verify XML tags are properly formatted
4. If needed, convert to v3:
   ```bash
   go run scripts/migrate-skills.go
   ```

---

## See Also

- [AGENTS_AND_SKILLS.md](./AGENTS_AND_SKILLS.md) - v3 agent and skill architecture
- [SKILL-AUTHORING.md](./SKILL-AUTHORING.md) - Writing v3 skills
- [CLAUDE_CODE_COMPATIBILITY.md](./CLAUDE_CODE_COMPATIBILITY.md) - Claude Code alignment
- [DEVELOPMENT.md](./DEVELOPMENT.md) - Development workflow
