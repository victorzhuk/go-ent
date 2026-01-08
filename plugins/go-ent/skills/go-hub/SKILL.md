---
name: go-hub
description: "Go enterprise development hub with spec-driven workflow, MCP integrations, and automatic verification. Auto-activates for: Go development, backend services, Clean Architecture, OpenSpec workflow. Orchestrates: go-arch, go-api, go-code, go-db, go-test, go-perf, go-sec, go-ops, go-review."
---

# Go Hub — Spec-Driven Enterprise Development

## Commands

```
/go-ent:init {project}     → Initialize project + openspec
/go-ent:plan {feature}     → Create change (proposal + tasks)
/go-ent:apply {id}         → Execute tasks with verification
/go-ent:status             → View changes and progress
/go-ent:archive {id}       → Complete and archive change
/go-ent:review             → Code review
/go-ent:tdd {desc}         → Test-driven development
```

## Core Principles

```
ZERO comments explaining WHAT → fix naming instead
Natural names: cfg, repo, srv → NOT applicationConfiguration
Domain has ZERO external deps
Interfaces at consumer side
Errors wrapped lowercase: fmt.Errorf("create user: %w", err)
```

## Verification Cycle (Mandatory)

After ANY code change:

```bash
make build              # Must pass
make lint               # Must be clean
make test               # Must pass

task gen:api            # OpenAPI/ogen (when relevant)
task gen:proto          # Protobuf/buf (when relevant)
```

**NEVER mark task complete until verification passes.**

## MCP Tools

```
mcp__serena__find_symbol(name: "...")
mcp__serena__find_referencing_symbols(symbol: "...")
mcp__serena__get_project_structure()
mcp__context7__resolve(library: "pgx|ogen|testify|squirrel")
mcp__github__create_issue(...)
```

## OpenSpec Structure

```
openspec/
├── project.yaml
├── specs/{capability}/
├── changes/{change-id}/
│   ├── proposal.md
│   ├── tasks.md
│   └── design.md
└── archive/
```

## Domain Skills

| Skill | Triggers |
|-------|----------|
| go-arch | architecture, layers, DI |
| go-api | OpenAPI, ogen, gRPC |
| go-code | implementation, patterns |
| go-db | PostgreSQL, Redis, migrations |
| go-test | testing, TDD, coverage |
| go-perf | profiling, optimization |
| go-sec | security, OWASP |
| go-ops | Docker, K8s, CI/CD |
| go-review | code review |

## Layer Architecture

```
Transport → UseCase → Domain ← Repository ← Infrastructure
```

## Verification Checklist

- [ ] `make build` passes
- [ ] `make lint` clean
- [ ] `make test` passes
- [ ] ZERO WHAT comments
- [ ] Natural variable names
- [ ] Domain zero external deps
- [ ] Errors wrapped lowercase

**Mantra:** Work → Right → Fast | Simple > Complex | Good > Perfect
