package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/victorzhuk/go-ent/internal/domain"
)

const (
	configDir  = ".go-ent"
	configFile = "config.yaml"
)

// Load loads configuration from the project root's .go-ent/config.yaml file.
// If the config file doesn't exist, returns default configuration.
// If the file exists but is invalid, returns an error.
func Load(projectRoot string) (*Config, error) {
	cfgPath := filepath.Join(projectRoot, configDir, configFile)

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidYAML, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

// LoadWithEnv loads configuration with environment variable overrides.
// Environment variables take precedence over config file values.
// Supported env vars:
//   - GOENT_BUDGET_DAILY: Override daily budget limit
//   - GOENT_BUDGET_MONTHLY: Override monthly budget limit
//   - GOENT_BUDGET_PER_TASK: Override per-task budget limit
//   - GOENT_RUNTIME_PREFERRED: Override preferred runtime
//   - GOENT_AGENTS_DEFAULT: Override default agent role
func LoadWithEnv(projectRoot string, getenv func(string) string) (*Config, error) {
	cfg, err := Load(projectRoot)
	if err != nil {
		return nil, err
	}

	if v := getenv("GOENT_BUDGET_DAILY"); v != "" {
		var val float64
		if _, err := fmt.Sscanf(v, "%f", &val); err != nil {
			return nil, fmt.Errorf("invalid GOENT_BUDGET_DAILY: %w", err)
		}
		cfg.Budget.Daily = val
	}

	if v := getenv("GOENT_BUDGET_MONTHLY"); v != "" {
		var val float64
		if _, err := fmt.Sscanf(v, "%f", &val); err != nil {
			return nil, fmt.Errorf("invalid GOENT_BUDGET_MONTHLY: %w", err)
		}
		cfg.Budget.Monthly = val
	}

	if v := getenv("GOENT_BUDGET_PER_TASK"); v != "" {
		var val float64
		if _, err := fmt.Sscanf(v, "%f", &val); err != nil {
			return nil, fmt.Errorf("invalid GOENT_BUDGET_PER_TASK: %w", err)
		}
		cfg.Budget.PerTask = val
	}

	if v := getenv("GOENT_RUNTIME_PREFERRED"); v != "" {
		runtime := domain.Runtime(v)
		if !runtime.Valid() {
			return nil, fmt.Errorf("invalid GOENT_RUNTIME_PREFERRED: %s", v)
		}
		cfg.Runtime.Preferred = runtime
	}

	if v := getenv("GOENT_AGENTS_DEFAULT"); v != "" {
		role := domain.AgentRole(v)
		if !role.Valid() {
			return nil, fmt.Errorf("invalid GOENT_AGENTS_DEFAULT: %s", v)
		}
		cfg.Agents.Default = role
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config after env overrides: %w", err)
	}

	return cfg, nil
}

// DefaultConfig returns the default configuration when no config file exists.
// Provides sensible defaults for all configuration sections.
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0",
		Agents: AgentsConfig{
			Default: domain.AgentRoleSenior,
			Roles: map[string]AgentRoleConfig{
				string(domain.AgentRoleArchitect): {
					Model:  "opus",
					Skills: []string{"go-arch", "go-api"},
				},
				string(domain.AgentRoleSenior): {
					Model:  "sonnet",
					Skills: []string{"go-code", "go-db", "go-test"},
				},
				string(domain.AgentRoleDeveloper): {
					Model:  "sonnet",
					Skills: []string{"go-code", "go-test"},
				},
			},
			Delegation: DelegationConfig{
				Auto: false,
			},
		},
		Runtime: RuntimeConfig{
			Preferred: domain.RuntimeClaudeCode,
			Fallback:  []domain.Runtime{domain.RuntimeCLI},
		},
		Budget: BudgetConfig{
			Daily:    10.0,
			Monthly:  200.0,
			PerTask:  1.0,
			Tracking: true,
		},
		Models: ModelsConfig{
			"opus":   "claude-opus-4-5-20251101",
			"sonnet": "claude-sonnet-4-5-20251101",
			"haiku":  "claude-haiku-3-5-20241022",
		},
		Skills: SkillsConfig{
			Enabled: []string{
				"go-code",
				"go-arch",
				"go-api",
				"go-db",
				"go-test",
			},
		},
	}
}
