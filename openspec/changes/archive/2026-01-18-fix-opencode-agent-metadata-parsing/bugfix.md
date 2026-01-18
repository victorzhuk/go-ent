## Bug Fix: OpenCode agent metadata parsing

**Symptom:** Running `go-ent init --tool opencode` failed with unmarshal error:
```
Error: generate opencode config: generate agents: load metadata for architect: parse agent metadata: yaml: unmarshal errors:
  line 18: cannot unmarshal !!seq into toolinit.AgentTags
```

**Root Cause:** The OpenCode adapter was directly unmarshaling YAML with `tags: []string` format into `AgentTags` struct, which expects a mapping. Claude adapter correctly uses `ParseAgentMetaYAML()` which transforms `tags: ["role:planning", "complexity:heavy"]` into the struct fields, but OpenCode had its own `loadAgentMetadata()` that bypassed this transformation.

**Fix:**
- Updated `OpenCodeAdapter.loadAgentMetadata()` to use shared `ParseAgentMetaYAML()` function instead of direct YAML unmarshaling
- Removed unused `gopkg.in/yaml.v3` import from `opencode.go`

**Files Changed:**
- `internal/toolinit/opencode.go`: Updated `loadAgentMetadata()` to call `ParseAgentMetaYAML()`

**Verification:**
```bash
# Build succeeds
go build ./...

# Error without --tool flag (as expected)
go run ./cmd/go-ent init
# Error: --tool flag is required. Use --tool=claude or --tool=opencode

# Claude tool init works
go run ./cmd/go-ent init --tool claude --dry-run --force
# DRY RUN - would create: commands/ent/..., agents/ent/..., skills/ent/...

# OpenCode tool init now works
go run ./cmd/go-ent init --tool opencode --dry-run --force
# DRY RUN - would create: command/ent/..., agent/ent/..., skill/ent/...

# All tool init works
go run ./cmd/go-ent init --tool all --dry-run --force
# ✅ Preview complete

# Tests pass
make test
# ok  	github.com/victorzhuk/go-ent/internal/toolinit	1.016s	coverage: 12.6% of statements
```

**Prevention:** Share YAML parsing logic across adapters to avoid inconsistent data structures. Use transformation functions that handle different YAML representations uniformly.
