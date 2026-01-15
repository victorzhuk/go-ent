# Proposal: Upgrade Template Engine

## Why

The current template system uses a simple `{{include "path"}}` directive for file inclusion. While functional, it lacks:
- Template inheritance (DRY principle)
- Compile-time validation
- Conditional logic
- Code reuse patterns

Go's standard `text/template` package provides these capabilities without external dependencies, enabling more maintainable and composable prompts.

## What Changes

Upgrade from custom `{{include}}` processing to Go's `text/template` engine:

**Current approach:**
```go
// Simple string replacement
content = strings.ReplaceAll(content, "{{include \"domains/openspec.md\"}}", includedContent)
```

**New approach:**
```go
// Full template engine with inheritance
tmpl := template.Must(template.ParseFiles("base.tmpl", "agent.tmpl"))
tmpl.ExecuteTemplate(&buf, "agent", data)
```

**Template features to add:**
1. **Inheritance** - `{{define}}`, `{{template}}`, `{{block}}`
2. **Custom functions** - Include, extend, indent
3. **Compile-time validation** - Catch errors at build time
4. **Data context** - Pass structured data to templates

**Example template structure:**
```
{{/* _base.md.tmpl */}}
{{define "agent-base"}}
You are {{.Role}}. {{.Description}}

{{template "tooling" .}}
{{template "conventions" .}}
{{block "prompt" .}}{{end}}
{{end}}

{{/* architect.md.tmpl */}}
{{template "agent-base" .}}
{{define "prompt"}}
You are a senior Go systems architect...
{{end}}
```

Key changes:
- Create `internal/template/` package
- Implement template engine with custom functions
- Update `internal/toolinit/transform.go` to use new engine
- Convert prompt files to `.tmpl` format (gradual migration)
- Add compile-time template validation

## Impact

**Affected specs:**
- `template-system` - Template processing and composition

**Affected code:**
- `internal/template/engine.go` - **NEW** - Template engine
- `internal/template/funcs.go` - **NEW** - Custom functions
- `internal/template/parse.go` - **NEW** - Template parsing utilities
- `internal/toolinit/transform.go` - Use new engine
- `plugins/sources/go-ent/agents/prompts/*.md` - Convert to `.tmpl` (optional migration)

**Breaking changes:** None - backward compatible with existing `.md` files

## Dependencies

**Requires:** `integrate-driver-into-adapters` (provides consolidated prompts)

**Blocks:** None (enhancement)

## Success Criteria

- [ ] `internal/template/` package implemented
- [ ] Template inheritance works (`{{define}}`, `{{template}}`, `{{block}}`)
- [ ] Custom functions: `include`, `indent`, `default`
- [ ] Compile-time validation catches template errors
- [ ] Backward compatible with existing `.md` files
- [ ] All tests pass
- [ ] At least one agent converted to `.tmpl` as proof-of-concept
