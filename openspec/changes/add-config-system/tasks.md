# Tasks: Add Configuration System

## 1. Create Config Package

- [x] Create `internal/config/` directory
- [x] Create `internal/config/config.go` - Main Config struct
- [ ] Create `internal/config/agents.go` - AgentsConfig section
- [ ] Create `internal/config/runtime.go` - RuntimeConfig section
- [ ] Create `internal/config/budget.go` - BudgetConfig section
- [ ] Create `internal/config/model.go` - ModelsConfig section
- [ ] Create `internal/config/loader.go` - YAML loading with env override
- [ ] Create `internal/config/defaults.go` - Default values
- [ ] Add package documentation

## 2. Implement Core Config Types

- [x] Define `Config` struct with all sections
- [x] Define `AgentsConfig` with roles map
- [x] Define `AgentRoleConfig` struct
- [x] Define `RuntimeConfig` with preferred/fallback
- [x] Define `BudgetConfig` with limits
- [x] Define `ModelsConfig` as string map
- [x] Define `SkillsConfig` with enabled list
- [x] Add validation methods for each struct

## 3. Implement Config Loader

- [x] Implement `Load(projectRoot string) (*Config, error)`
- [x] Implement `LoadWithEnv(projectRoot, getenv) (*Config, error)`
- [x] Add support for `.go-ent/config.yaml` path resolution
- [x] Add YAML unmarshaling with validation
- [ ] Add environment variable override logic
- [ ] Add config file existence check
- [ ] Return defaults when config file missing
- [ ] Add comprehensive error messages

## 4. Implement Default Configuration

- [ ] Define default agent roles (architect, senior, developer)
- [ ] Define default models (opus, sonnet)
- [ ] Define default runtime (claude-code)
- [ ] Define default budget limits
- [ ] Add `DefaultConfig() *Config` function
- [ ] Document default values

## 5. Add Environment Variable Support

- [ ] Support `GOENT_BUDGET_DAILY` override
- [ ] Support `GOENT_BUDGET_MONTHLY` override
- [ ] Support `GOENT_RUNTIME_PREFERRED` override
- [ ] Support `GOENT_AGENTS_DEFAULT` override
- [ ] Add env var parsing helpers
- [ ] Document env var naming convention

## 6. Integration with Existing Code

- [ ] Update `internal/spec/store.go`
  - Add `ConfigPath() string` method
  - Add `LoadConfig() (*config.Config, error)` method
  - Add `SaveConfig(cfg *config.Config) error` method
  - Add config file path resolution
- [ ] Update `internal/generation/config.go`
  - Import new config package
  - Add integration points
  - Maintain backward compatibility
- [ ] Verify no circular dependencies

## 7. Config Validation

- [ ] Validate agent role names match domain.AgentRole
- [ ] Validate runtime names match domain.Runtime
- [ ] Validate model mappings are non-empty
- [ ] Validate budget values are positive
- [ ] Validate skills list contains valid skill IDs
- [ ] Add detailed validation error messages
- [ ] Add `Validate() error` method on Config

## 8. Testing

- [ ] Unit test config loading from YAML
- [ ] Unit test default config generation
- [ ] Unit test env var overrides
- [ ] Unit test validation (valid cases)
- [ ] Unit test validation (invalid cases)
- [ ] Unit test error handling (missing file, bad YAML)
- [ ] Integration test with spec store
- [ ] Test config file creation

## 9. Example Configuration

- [ ] Create `examples/go-ent/config.yaml` with full example
- [ ] Create `examples/go-ent/config-minimal.yaml` with minimal example
- [ ] Document each section in examples
- [ ] Add comments explaining options

## 10. Documentation

- [ ] Add package-level documentation
- [ ] Document each config section
- [ ] Add godoc examples for loading
- [ ] Document env var naming convention
- [ ] Add migration guide from no config
