# Implementation Tasks

## 1. Create Schema

- [ ] Create `openspec/schemas/go-ent/schema.yaml`
  - [ ] Define proposal artifact
  - [ ] Define specs artifact
  - [ ] Define design artifact
  - [ ] Define tasks artifact
- [ ] Test with `openspec schema validate go-ent`

## 2. Create Templates

- [ ] Create `openspec/schemas/go-ent/templates/proposal.md`
- [ ] Create `openspec/schemas/go-ent/templates/spec.md`
- [ ] Create `openspec/schemas/go-ent/templates/design.md`
- [ ] Create `openspec/schemas/go-ent/templates/tasks.md`

## 3. Add `ent init` Command

- [ ] Create `internal/cli/init.go`
- [ ] Parse go.mod for module path and version
- [ ] Write `openspec/config.yaml`
- [ ] Test command

## 4. Documentation

- [ ] Update CLAUDE.md with OpenSpec workflow
- [ ] Document `ent init` command
