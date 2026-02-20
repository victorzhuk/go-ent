# 🚀 Skills Mega Pack — Enterprise-Grade AI Agent Skills

**45 curated, enhanced skills for AI coding agents (Claude Code, OpenCode, Cursor, Cline, etc.)**

Sourced from the top skills on [skills.sh](https://skills.sh) leaderboard, enhanced with enterprise patterns, modern best practices, and production-ready guidelines.

## Installation

### Claude Code
```bash
# Copy all skills
cp -r skills-mega-pack/* ~/.claude/skills/

# Or per-project
cp -r skills-mega-pack/* /path/to/project/.claude/skills/
```

### OpenCode
```bash
cp -r skills-mega-pack/* ~/.opencode/skills/
```

### Any Agent (npx skills)
```bash
# Individual install from skills.sh
npx skills add <owner/repo>
```

## Skill Categories

### 🔹 Go / Golang (7 skills)
| Skill | Description |
|-------|-------------|
| `go-core` | Go 1.22+ idiomatic patterns, error handling, concurrency |
| `go-api` | Spec-first API design with stdlib, OpenAPI/ogen, gRPC |
| `go-database` | PostgreSQL/pgx, squirrel, goose migrations, Redis |
| `go-testing` | Table-driven tests, testcontainers, benchmarks, fuzzing, TDD |
| `go-microservices` | Clean architecture, observability, graceful shutdown |
| `go-security` | OWASP, auth, input validation, secrets management |
| `go-performance` | Profiling, benchmarks, memory optimization, PGO |

### 🦀 Rust (2 skills)
| Skill | Description |
|-------|-------------|
| `rust-core` | Ownership, lifetimes, error handling, async, traits |
| `rust-web` | Axum, SQLx, tower middleware, production deployment |

### 🐍 Python (3 skills)
| Skill | Description |
|-------|-------------|
| `python-core` | Python 3.12+ with type hints, async, uv, pytest |
| `python-fastapi` | FastAPI with Pydantic v2, async SQLAlchemy, DI |
| `python-django` | Django 5, DRF, ORM optimization, async views |

### 🐘 PHP (2 skills)
| Skill | Description |
|-------|-------------|
| `php-core` | PHP 8.3+ with PSR standards, type safety, Composer |
| `php-laravel` | Laravel with Eloquent, queues, events, testing |

### 🟢 Node.js (2 skills)
| Skill | Description |
|-------|-------------|
| `nodejs-backend` | Express/Fastify with TypeScript, async patterns |
| `nodejs-nestjs` | NestJS modules, DI, guards, interceptors, clean arch |

### ⚛️ React.js (2 skills)
| Skill | Description |
|-------|-------------|
| `react-core` | React 19 hooks, Server Components, performance, testing |
| `react-nextjs` | Next.js 15 App Router, Server Actions, caching |

### 🦋 Flutter (1 skill)
| Skill | Description |
|-------|-------------|
| `flutter-core` | Dart, Riverpod, GoRouter, widget patterns, testing |

### 🏗️ Software Development (7 skills)
| Skill | Description |
|-------|-------------|
| `clean-architecture` | Clean Architecture, DDD, SOLID, dependency injection |
| `tdd` | Test-Driven Development workflow and patterns |
| `git-workflow` | Conventional commits, branching, code review, CI/CD |
| `api-design` | RESTful API design, versioning, pagination, OpenAPI |
| `docker-devops` | Docker, Kubernetes, CI/CD pipelines, IaC |
| `security-best-practices` | OWASP Top 10, auth, input validation, secrets |
| `database-design` | Schema design, indexing, migrations, query optimization |

### 🤖 Agent Orchestration (3 skills)
| Skill | Description |
|-------|-------------|
| `agent-patterns` | Multi-agent orchestration, delegation, routing, workflows |
| `mcp-development` | MCP server development, tool design, integration |
| `prompt-engineering` | Prompt patterns for agents: XML, chain-of-thought, few-shot |

### 🔧 Backend Patterns (4 skills)
| Skill | Description |
|-------|-------------|
| `graphql-development` | GraphQL schema-first, DataLoader, subscriptions |
| `grpc-development` | Protocol Buffers, streaming, interceptors |
| `message-queues` | Kafka, RabbitMQ, NATS, event-driven patterns |
| `microservices-patterns` | Service design, sagas, circuit breakers, data management |

### 🎨 Frontend (3 skills)
| Skill | Description |
|-------|-------------|
| `vue-development` | Vue 3 Composition API, Pinia, TypeScript |
| `tailwind-css` | Tailwind v4 utility patterns, responsive, theming |
| `web-design` | Typography, color, spacing, layout, accessibility |

### ⭐ High-Value General (6 skills)
| Skill | Description |
|-------|-------------|
| `observability` | OpenTelemetry, structured logging, tracing, metrics |
| `performance-optimization` | Profiling, caching, database and API optimization |
| `typescript-advanced` | Utility types, generics, type guards, strict config |
| `postgresql-advanced` | Indexing, JSONB, window functions, CTEs, tuning |
| `redis-patterns` | Caching strategies, data structures, distributed locks |
| `systematic-debugging` | Reproduce, isolate, diagnose, fix, prevent |

### 📋 Workflow (3 skills)
| Skill | Description |
|-------|-------------|
| `code-review` | Review checklist, giving feedback, automation |
| `writing-plans` | Implementation plans, RFCs, ADRs |
| `session-handoff` | Context preservation for AI agent session continuity |

## Key Differences from skills.sh

These skills are **enhanced** over the cursor-rule conversions on skills.sh:

1. **Enterprise-grade**: Production patterns, not just tutorials
2. **Opinionated**: Clear recommendations vs wishy-washy guidelines
3. **Modern**: Latest language/framework versions (Go 1.22+, Python 3.12+, etc.)
4. **Code-heavy**: Practical examples, not just prose descriptions
5. **Agent-optimized**: Written for AI agent consumption with clear structure

## Sources & Credits

- [skills.sh](https://skills.sh) leaderboard — top community skills
- [mindrally/skills](https://github.com/Mindrally/skills) — 240+ cursor rule conversions
- [obra/superpowers](https://skills.sh/obra/superpowers) — TDD, debugging, planning
- [wshobson/agents](https://skills.sh/wshobson/agents) — API design, backend patterns
- [anthropics/skills](https://skills.sh/anthropics/skills) — official Anthropic skills
- [softaworks/agent-toolkit](https://skills.sh/softaworks/agent-toolkit) — agent tooling

## License

MIT — Use freely in your projects and agent configurations.
