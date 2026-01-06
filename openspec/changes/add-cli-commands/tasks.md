# Tasks: Add CLI Commands

## 1. CLI Framework
- [ ] Add cobra/pflag dependencies
- [ ] Create internal/cli/root.go
- [ ] Set up command structure
- [ ] Add global flags (--config, --verbose)

## 2. Run Command
- [ ] Create internal/cli/run.go
- [ ] Implement `go-ent run <action>` command
- [ ] Add flags: --agent, --strategy, --budget, --dry-run
- [ ] Integration with execution engine

## 3. Agent Commands
- [ ] Create internal/cli/agent.go
- [ ] Implement `go-ent agent list`
- [ ] Implement `go-ent agent info <name>`
- [ ] Display agent capabilities

## 4. Skill Commands
- [ ] Create internal/cli/skill.go
- [ ] Implement `go-ent skill list`
- [ ] Implement `go-ent skill info <name>`

## 5. Spec Commands
- [ ] Create internal/cli/spec.go
- [ ] Implement `go-ent spec init/list/show`
- [ ] Reuse existing spec management code

## 6. Config Commands
- [ ] Create internal/cli/config.go
- [ ] Implement `go-ent config show`
- [ ] Implement `go-ent config set <key> <value>`
- [ ] Implement `go-ent config init`

## 7. Main Integration
- [ ] Update cmd/go-ent/main.go
- [ ] Add CLI mode detection (args vs stdin)
- [ ] Route to MCP server or CLI based on mode

## 8. Testing
- [ ] Integration tests for each command
- [ ] Test with real config files
- [ ] Test error handling

## 9. Documentation
- [ ] Add CLI usage examples
- [ ] Update README with CLI commands
