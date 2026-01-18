---
name: skill-sync
description: "Sync skills from plugins to Claude skills directory"
---

# Skill Sync

Synchronizes skill files from the plugin source directory to the Claude Code skills directory.

## What It Does

Copies all skill directories from `plugins/go-ent/skills/` to `.claude/skills/ent/`:

- **Source**: `plugins/go-ent/skills/` (14 skills in v2 format)
- **Destination**: `.claude/skills/ent/` (where Claude Code loads skills from)
- **Behavior**: Overwrites existing files, preserves directory structure

## Workflow

### Step 1: Prepare Destination

Create the destination directory if it doesn't exist:

```bash
mkdir -p .claude/skills/ent
```

### Step 2: Copy Skills

Copy all skills from source to destination:

```bash
cp -r plugins/go-ent/skills/* .claude/skills/ent/
```

This preserves the directory structure:
- `plugins/go-ent/skills/go/` → `.claude/skills/ent/go/`
- `plugins/go-ent/skills/core/` → `.claude/skills/ent/core/`

### Step 3: Validate Skills

Validate all synced skills using the MCP tool:

```bash
# This is done via MCP tool call
# No manual command needed - validation happens automatically
```

The validation step checks:
- Skill metadata format (SKILL.md frontmatter)
- Required sections present
- Markdown formatting valid

### Step 4: Report Results

Report which skills were synced:

```
Synced 14 skills:
  go-api
  go-arch
  go-code
  go-db
  go-ops
  go-perf
  go-review
  go-sec
  go-test
  api-design
  arch-core
  debug-core
  review-core
  security-core

Validation: ✅ PASS
```

## Error Handling

If any validation errors occur:

```
Synced 14 skills:

Validation: ❌ FAIL
  - go-ops/SKILL.md: Missing required section 'Examples'
  - go-test/SKILL.md: Invalid frontmatter format

Fix errors and re-run sync
```

## Usage

Execute via task system:

```
/ent:task 5.1 skill-sync
```

Or run manually:

```bash
mkdir -p .claude/skills/ent
cp -r plugins/go-ent/skills/* .claude/skills/ent/
```

## Verification

After sync, verify skills are available in Claude Code:

```bash
ls -la .claude/skills/ent/
```

Expected output:
```
go/
core/
```

Check individual skill directories:

```bash
ls -la .claude/skills/ent/go/
```

Expected output:
```
go-api/
go-arch/
go-code/
go-db/
go-ops/
go-perf/
go-review/
go-sec/
go-test/
```

---

## When to Use

Run `skill-sync` after:

- Adding or modifying skills in `plugins/go-ent/skills/`
- Converting skills to v2 format
- Updating skill metadata or content
- Setting up a new development environment

## Notes

- **Overwrites existing files**: Destination skills will be replaced with source versions
- **Preserves structure**: Directory hierarchy is maintained
- **No selective sync**: All skills are copied (use manual copy for selective updates)
- **Auto-validation**: Validation happens automatically after sync
