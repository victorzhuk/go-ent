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

	data, err := os.ReadFile(cfgPath) // #nosec G304 -- controlled config/template file path
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
//
// # Environment Variable Naming Convention
//
// All go-ent environment variables follow the pattern: GOENT_<SECTION>_<FIELD>
//
// Examples:
//   - GOENT_BUDGET_DAILY (maps to config.budget.daily)
//   - GOENT_RUNTIME_PREFERRED (maps to config.runtime.preferred)
//   - GOENT_AGENTS_DEFAULT (maps to config.agents.default)
//
// Naming Rules:
//   - Prefix: Always starts with GOENT_
//   - Section: Top-level config section (BUDGET, RUNTIME, AGENTS, etc.)
//   - Field: Specific field within that section
//   - Case: Always SCREAMING_SNAKE_CASE
//   - Separators: Underscores between prefix, section, and field
//
// Value Formats:
//   - Float values: "10.5" or "200.0" (for budget limits)
//   - String values: "claude-code" or "senior" (for runtime/agent)
//   - Boolean values: "true" or "false" (for tracking flags)
//
// # Supported Environment Variables
//
// Budget Section:
//   - GOENT_BUDGET_DAILY: Override daily budget limit (float, USD)
//   - GOENT_BUDGET_MONTHLY: Override monthly budget limit (float, USD)
//   - GOENT_BUDGET_PER_TASK: Override per-task budget limit (float, USD)
//
// Runtime Section:
//   - GOENT_RUNTIME_PREFERRED: Override preferred runtime (claude-code|opencode|cli)
//
// Agents Section:
//   - GOENT_AGENTS_DEFAULT: Override default agent role (architect|senior|developer|...)
//
// Metrics Section:
//   - GOENT_METRICS_ENABLED: Enable or disable metrics collection (true|false)
//
// # Error Handling
//
// Invalid environment variable values will cause LoadWithEnv to return an error:
//   - Type mismatch (e.g., "abc" for a float budget value)
//   - Invalid enum values (e.g., "invalid-runtime" for GOENT_RUNTIME_PREFERRED)
//   - Validation failures (e.g., negative budget values)
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

	if v := getenv("GOENT_METRICS_ENABLED"); v != "" {
		var val bool
		if _, err := fmt.Sscanf(v, "%t", &val); err != nil {
			return nil, fmt.Errorf("invalid GOENT_METRICS_ENABLED: %w", err)
		}
		cfg.Metrics.Enabled = val
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config after env overrides: %w", err)
	}

	return cfg, nil
}
