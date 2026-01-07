# Tasks: Add Configuration System

## 1. Create Config Package

- [x] Create `internal/config/` directory
- [x] Create `internal/config/config.go` - Main Config struct
- [x] Create `internal/config/agents.go` - AgentsConfig section
- [x] Create `internal/config/runtime.go` - RuntimeConfig section (implemented in config.go)
- [x] Create `internal/config/budget.go` - BudgetConfig section (implemented in config.go)
- [x] Create `internal/config/model.go` - ModelsConfig section (implemented in config.go)
- [x] Create `internal/config/loader.go` - YAML loading with env override
- [x] Create `internal/config/defaults.go` - Default values
- [x] Add package documentation

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
- [x] Add environment variable override logic
- [x] Add config file existence check
- [x] Return defaults when config file missing
- [x] Add comprehensive error messages

## 4. Implement Default Configuration

- [x] Define default agent roles (architect, senior, developer)
- [x] Define default models (opus, sonnet)
- [x] Define default runtime (claude-code)
- [x] Define default budget limits
- [x] Add `DefaultConfig() *Config` function
- [x] Document default values

## 5. Add Environment Variable Support

- [x] Support `GOENT_BUDGET_DAILY` override
- [x] Support `GOENT_BUDGET_MONTHLY` override
- [x] Support `GOENT_RUNTIME_PREFERRED` override
- [x] Support `GOENT_AGENTS_DEFAULT` override
- [x] Add env var parsing helpers (GOENT_BUDGET_PER_TASK also supported)
- [x] Document env var naming convention

## 6. Integration with Existing Code

- [x] Update `internal/spec/store.go`
  - Add `ConfigPath() string` method
  - Add `LoadConfig() (*config.Config, error)` method
  - Add `SaveConfig(cfg *config.Config) error` method
  - Add config file path resolution
- [x] Update `internal/generation/config.go`
  - Import new config package
  - Add integration points
  - Maintain backward compatibility
- [x] Verify no circular dependencies

## 7. Config Validation

- [x] Validate agent role names match domain.AgentRole
- [x] Validate runtime names match domain.Runtime
- [x] Validate model mappings are non-empty
- [x] Validate budget values are positive
- [x] Validate skills list contains valid skill IDs
- [x] Add detailed validation error messages
- [x] Add `Validate() error` method on Config

## 8. Testing

- [x] Unit test config loading from YAML
- [x] Unit test default config generation
- [x] Unit test env var overrides
- [x] Unit test validation (valid cases)
- [x] Unit test validation (invalid cases)
- [x] Unit test error handling (missing file, bad YAML)
- [x] Integration test with spec store
- [x] Test config file creation

## 9. Example Configuration

- [x] Create `examples/go-ent/config.yaml` with full example
- [x] Create `examples/go-ent/config-minimal.yaml` with minimal example
- [x] Document each section in examples
- [x] Add comments explaining options

## 10. Documentation

- [x] Add package-level documentation
- [x] Document each config section
- [x] Add godoc examples for loading
- [x] Document env var naming convention
- [x] Add migration guide from no config
