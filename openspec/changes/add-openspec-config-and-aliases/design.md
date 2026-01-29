# Design: OpenSpec Schema Support

## Context

go-ent currently uses custom `/ent:*` commands. The standard OpenSpec approach uses `openspec init` to scaffold skills and `/opsx:*` commands for workflow.

We will provide a simple go-ent schema that works with standard OpenSpec CLI.

## Goals / Non-Goals

**Goals:**
- Create `openspec/schemas/go-ent/schema.yaml` defining artifact types
- Create Go-specific templates for proposals, specs, design, tasks
- Add simple `ent init` command to scaffold `openspec/config.yaml`
- Work with standard `openspec` CLI and `/opsx:*` commands

**Non-Goals:**
- Config generators
- Schema generators  
- Template generators
- Custom `/opsx:*` command aliases
- Migration tools
- npm packages

## Decisions

### Decision 1: Simple Static Schema

**What:** Create static schema.yaml and templates, not generated.

**Why:**
- Simple to understand and maintain
- No build step required
- Works immediately with `openspec init`
- Can be versioned in git

### Decision 2: Simple Templates

Create markdown templates with Go-specific guidance in `openspec/schemas/go-ent/templates/`:
- proposal.md - Change proposal with Why/What/Impact
- spec.md - Delta specs (ADDED/MODIFIED/REMOVED)
- design.md - Design with Go patterns  
- tasks.md - Task checklist

### Decision 3: Simple `ent init` Command

Add `ent init` that creates basic openspec/config.yaml by reading go.mod.

### Decision 4: Standard OpenSpec Workflow

Use standard `/opsx:*` commands, no custom aliases.

User workflow:
1. `ent init` - creates openspec/config.yaml
2. `openspec init` - uses go-ent schema, creates .claude/skills/
3. `/opsx:new`, `/opsx:apply`, `/opsx:archive` - standard commands

## Implementation Plan

Phase 1: Create schema.yaml
Phase 2: Create templates
Phase 3: Add `ent init` command
Phase 4: Documentation

## Risks

| Risk | Mitigation |
|------|-----------|
| Schema outdated | Simple to update manually |
| Templates don't fit | Users customize after init |
