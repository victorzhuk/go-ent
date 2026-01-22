# Tasks: Clean Up Dead Code

## 1. Delete orphaned packages
- [ ] 1.1 Delete `internal/rules/` directory (~284 lines)
- [ ] 1.2 Delete `internal/tool/` directory (~213 lines)
- [ ] 1.3 Delete `internal/embedded/` empty directory
- [ ] 1.4 Delete `internal/spec/cmd/` empty directory

## 2. Remove deprecated functions
- [ ] 2.1 Remove `Save()` from `internal/spec/registry_store.go`
- [ ] 2.2 Remove `parseTasksFile()` from `internal/spec/registry_store.go`
- [ ] 2.3 Remove `validateExplicitTriggers()` from `internal/skill/rules.go`

## 3. Clean up incomplete TODOs
- [ ] 3.1 Remove stub code in `internal/agent/background/manager.go`

## 4. Verify build
- [ ] 4.1 Run `make build`
- [ ] 4.2 Run `make test`
- [ ] 4.3 Run `make lint`
