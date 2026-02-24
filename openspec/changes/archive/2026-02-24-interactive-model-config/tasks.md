# Tasks Checklist

## 1. Remove Legacy Configuration

- [x] 1.1 Delete internal/genconfig/ package
- [x] 1.2 Delete internal/config/model_config.go
- [x] 1.3 Delete internal/config/model_defaults.go
- [x] 1.4 Delete internal/config/resolver.go
- [x] 1.5 Delete internal/cli/model.go (model commands)
- [x] 1.6 Remove genconfig imports from all files
- [x] 1.7 Delete root ent.yaml
- [x] 1.8 Update .gitignore to include .claude/ and .opencode/

## 2. Create Runtime Configuration

- [x] 2.1 Create internal/config/runtime_config.go
- [x] 2.2 Define RuntimeConfig interface (Claude() ClaudeConfig, OpenCode() OpenCodeConfig)
- [x] 2.3 Define ClaudeConfig struct with short alias fields (Sonnet, Opus, Haiku)
- [x] 2.4 Define OpenCodeConfig struct with model ID fields
- [x] 2.5 Implement LoadRuntimeConfig(runtime string) function
- [x] 2.6 Implement config file loading from .claude/ent.yaml or .opencode/ent.yaml
- [x] 2.7 Implement default values when config file missing
- [x] 2.8 Write tests for runtime_config.go

## 3. Create OpenCode Discovery

- [x] 3.1 Create internal/config/opencode_discovery.go
- [x] 3.2 Implement DiscoverModels() function (runs `opencode models`)
- [x] 3.3 Implement cache at ~/.cache/go-ent/opencode-models.json
- [x] 3.4 Implement 24h cache expiry check
- [x] 3.5 Implement fallback defaults when CLI unavailable
- [x] 3.6 Add warning log when opencode CLI not found
- [x] 3.7 Write tests for opencode_discovery.go

## 4. Update Generator

- [x] 4.1 Update internal/generator/generator.go to use RuntimeConfig
- [x] 4.2 Remove ResolveModel() function if standalone
- [x] 4.3 Replace genconfig.Config usage with RuntimeConfig interface
- [x] 4.4 Update generator constructors to accept RuntimeConfig
- [x] 4.5 Update tests for generator changes

## 5. Update CLI Commands

- [x] 5.1 Update internal/cli/agent/generate.go to load runtime config
- [x] 5.2 Update internal/cli/skill/generate.go to load runtime config
- [x] 5.3 Update internal/cli/init.go to remove ent.yaml generation
- [x] 5.4 Add runtime detection (check .claude/ or .opencode/ dir exists)
- [x] 5.5 Update CLI tests

## 6. Verification

- [x] 6.1 Run make build - must compile
- [x] 6.2 Run make test - all tests pass
- [x] 6.3 Run make lint - no errors (pre-existing issues in skill-convert)
- [x] 6.4 Test `ent agent generate` without config files
- [x] 6.5 Test with .claude/ent.yaml present
- [x] 6.6 Test with .opencode/ent.yaml present
- [x] 6.7 Verify no ent.yaml created in root
- [x] 6.8 Verify no .go-ent/ directory created
