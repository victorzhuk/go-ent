# Tasks: Add CLI Commands

## 1. CLI Framework
- [x] Add cobra/pflag dependencies
- [x] Create internal/cli/root.go
- [x] Set up command structure
- [x] Add global flags (--config, --verbose)

## 2. Run Command
- [x] Create internal/cli/run.go
- [x] Implement `go-ent run <action>` command
- [x] Add flags: --agent, --strategy, --budget, --dry-run
- [x] Integration with execution engine (stub implemented, blocked by add-execution-engine)

## 3. Agent Commands
- [x] Create internal/cli/agent.go
- [x] Implement `go-ent agent list`
- [x] Implement `go-ent agent info <name>`
- [x] Display agent capabilities

## 4. Skill Commands
- [x] Create internal/cli/skill.go
- [x] Implement `go-ent skill list`
- [x] Implement `go-ent skill info <name>`

## 5. Spec Commands
- [x] Create internal/cli/spec.go
- [x] Implement `go-ent spec init/list/show`
- [x] Reuse existing spec management code

## 6. Config Commands
- [x] Create internal/cli/config.go
- [x] Implement `go-ent config show`
- [x] Implement `go-ent config set <key> <value>`
- [x] Implement `go-ent config init`

## 7. Main Integration
- [x] Update cmd/go-ent/main.go
- [x] Add CLI mode detection (args vs stdin)
- [x] Route to MCP server or CLI based on mode

## 8. Testing
- [x] Integration tests for each command
- [x] Test with real config files
- [x] Test error handling

## 9. Documentation
- [x] Add CLI usage examples
- [x] Update README with CLI commands
