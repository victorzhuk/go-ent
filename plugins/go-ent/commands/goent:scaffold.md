---
description: Scaffold Go components (entity, repository, usecase, handler, service)
argument-hint: <type> <name> [impl]
---

# Scaffold Go Component

Generate boilerplate for: $ARGUMENTS

## Types

- `entity User` → Domain entity + tests
- `repository User pgx` → Repository with models, mappers, schema + contract
- `usecase CreateUser` → UseCase with request/response
- `handler User` → HTTP handler + DTOs
- `service Order` → Full stack (all above)

## Rules

- Follow Clean Architecture
- Private by default
- `New()` public, `new*()` private
- No AI-style verbose names
- Zero comments explaining WHAT
- Context as first parameter
- Proper error wrapping

Generate files using patterns from go-patterns skill.
