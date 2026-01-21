package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/victorzhuk/go-ent/internal/opencode"
	"github.com/victorzhuk/go-ent/internal/provider"

	"gopkg.in/yaml.v3"
)

const (
	defaultProvidersFile      = "providers.yaml"
	defaultGoentDir           = ".goent"
	defaultOpenCodeConfigPath = "~/.config/opencode/opencode.json"
)

type CommunicationMethod string

const (
	MethodACP CommunicationMethod = "acp"
	MethodCLI CommunicationMethod = "cli"
	MethodAPI CommunicationMethod = "api"
)

type ResetPeriod string

const (
	ResetHourly  ResetPeriod = "hourly"
	ResetDaily   ResetPeriod = "daily"
	ResetWeekly  ResetPeriod = "weekly"
	ResetMonthly ResetPeriod = "monthly"
)

func (r ResetPeriod) String() string {
	if r == "" {
		return "unknown"
	}
	return string(r)
}

func (r ResetPeriod) Valid() bool {
	switch r {
	case ResetHourly, ResetDaily, ResetWeekly, ResetMonthly:
		return true
	default:
		return false
	}
}

func (m CommunicationMethod) String() string {
	if m == "" {
		return "unknown"
	}
	return string(m)
}

func (m CommunicationMethod) Valid() bool {
	switch m {
	case MethodACP, MethodCLI, MethodAPI:
		return true
	default:
		return false
	}
}

type CostConfig struct {
	Per1kTokens float64            `yaml:"per_1k_tokens"`
	PerHour     float64            `yaml:"per_hour,omitempty"`
	PerDay      float64            `yaml:"per_day,omitempty"`
	PerMonth    float64            `yaml:"per_month,omitempty"`
	Multipliers map[string]float64 `yaml:"multipliers,omitempty"`
}

func (c *CostConfig) Validate() error {
	if c.Per1kTokens < 0 {
		return fmt.Errorf("per_1k_tokens cannot be negative")
	}
	if c.PerHour < 0 {
		return fmt.Errorf("per_hour cannot be negative")
	}
	if c.PerDay < 0 {
		return fmt.Errorf("per_day cannot be negative")
	}
	if c.PerMonth < 0 {
		return fmt.Errorf("per_month cannot be negative")
	}

	for method, multiplier := range c.Multipliers {
		if multiplier < 0 {
			return fmt.Errorf("multiplier for %s cannot be negative", method)
		}
		switch CommunicationMethod(method) {
		case MethodACP, MethodCLI, MethodAPI:
		default:
			return fmt.Errorf("invalid method in multipliers: %s", method)
		}
	}

	return nil
}

type CostTrackingConfig struct {
	Enabled        bool        `yaml:"enabled"`
	GlobalBudget   float64     `yaml:"global_budget,omitempty"`
	ResetPeriod    ResetPeriod `yaml:"reset_period,omitempty"`
	PersistHistory bool        `yaml:"persist_history,omitempty"`
	HistoryFile    string      `yaml:"history_file,omitempty"`
}

func (c *CostTrackingConfig) Validate() error {
	if c.Enabled {
		if c.GlobalBudget < 0 {
			return fmt.Errorf("global_budget cannot be negative")
		}
		if c.ResetPeriod != "" && !c.ResetPeriod.Valid() {
			return fmt.Errorf("invalid reset_period: %s", c.ResetPeriod)
		}
		if c.PersistHistory && c.HistoryFile == "" {
			return fmt.Errorf("history_file required when persist_history is enabled")
		}
	}
	return nil
}

type ProviderDefinition struct {
	Method       CommunicationMethod `yaml:"method"`
	Provider     string              `yaml:"provider"`
	Model        string              `yaml:"model"`
	BestFor      []string            `yaml:"best_for,omitempty"`
	Cost         *CostConfig         `yaml:"cost,omitempty"`
	ContextLimit int                 `yaml:"context_limit,omitempty"`
}

type ProvidersConfig struct {
	Providers          map[string]ProviderDefinition `yaml:"providers"`
	Defaults           Defaults                      `yaml:"defaults,omitempty"`
	Health             *HealthConfig                 `yaml:"health,omitempty"`
	OpenCodeConfigPath string                        `yaml:"opencode_config_path,omitempty"`
	CostTracking       *CostTrackingConfig           `yaml:"cost_tracking,omitempty"`
}

type Defaults struct {
	Implementation string `yaml:"implementation,omitempty"`
	LargeContext   string `yaml:"large_context,omitempty"`
	SimpleTasks    string `yaml:"simple_tasks,omitempty"`
	Research       string `yaml:"research,omitempty"`
	Planning       string `yaml:"planning,omitempty"`
	Review         string `yaml:"review,omitempty"`
}

type HealthConfig struct {
	CheckInterval int `yaml:"check_interval,omitempty"`
	WorkerTimeout int `yaml:"worker_timeout,omitempty"`
	MaxRetries    int `yaml:"max_retries,omitempty"`
	RetryDelay    int `yaml:"retry_delay,omitempty"`
}

func DefaultHealthConfig() *HealthConfig {
	return &HealthConfig{
		CheckInterval: 30,
		WorkerTimeout: 300,
		MaxRetries:    3,
		RetryDelay:    5,
	}
}

func substituteEnvVars(cfg *ProvidersConfig, logger *slog.Logger) error {
	totalSubs := 0
	allUsedVars := make(map[string]bool)

	// Substitute opencode_config_path
	if cfg.OpenCodeConfigPath != "" {
		newValue, usedVars, err := ApplySubstitution(cfg.OpenCodeConfigPath)
		if err != nil {
			return fmt.Errorf("substitute opencode_config_path: %w", err)
		}
		if newValue != cfg.OpenCodeConfigPath {
			logField := "value"
			if IsSecret("api_key") {
				logField = "value (redacted)"
			}
			logger.Info("environment variable substitution", "field", "opencode_config_path", logField, RedactSecret(newValue, usedVars))
			for _, v := range usedVars {
				allUsedVars[v] = true
			}
			cfg.OpenCodeConfigPath = newValue
			totalSubs += len(usedVars)
			if len(usedVars) == 0 {
				totalSubs++
			}
		}
	}

	// Substitute provider fields
	for name, provider := range cfg.Providers {
		modified := false

		// Substitute provider name
		if provider.Provider != "" {
			newValue, usedVars, err := ApplySubstitution(provider.Provider)
			if err != nil {
				return fmt.Errorf("substitute provider %s field provider: %w", name, err)
			}
			if newValue != provider.Provider {
				logField := "provider"
				logger.Info("environment variable substitution", "provider", name, "field", logField, "value", RedactSecret(newValue, usedVars))
				for _, v := range usedVars {
					allUsedVars[v] = true
				}
				provider.Provider = newValue
				modified = true
				totalSubs += len(usedVars)
				if len(usedVars) == 0 {
					totalSubs++
				}
			}
		}

		// Substitute model name
		if provider.Model != "" {
			newValue, usedVars, err := ApplySubstitution(provider.Model)
			if err != nil {
				return fmt.Errorf("substitute provider %s field model: %w", name, err)
			}
			if newValue != provider.Model {
				logField := "model"
				logger.Info("environment variable substitution", "provider", name, "field", logField, "value", RedactSecret(newValue, usedVars))
				for _, v := range usedVars {
					allUsedVars[v] = true
				}
				provider.Model = newValue
				modified = true
				totalSubs += len(usedVars)
				if len(usedVars) == 0 {
					totalSubs++
				}
			}
		}

		// Update map if modified
		if modified {
			cfg.Providers[name] = provider
		}
	}

	if totalSubs > 0 {
		vars := make([]string, 0, len(allUsedVars))
		for v := range allUsedVars {
			vars = append(vars, v)
		}
		logger.Info("environment variable substitution complete", "variables", vars, "substitutions", totalSubs)
	}

	return nil
}

func (h *HealthConfig) Validate() error {
	if h.CheckInterval <= 0 {
		return fmt.Errorf("check_interval must be positive")
	}
	if h.WorkerTimeout <= 0 {
		return fmt.Errorf("worker_timeout must be positive")
	}
	if h.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}
	if h.RetryDelay < 0 {
		return fmt.Errorf("retry_delay cannot be negative")
	}
	return nil
}

func LoadProviders(projectRoot string) (*ProvidersConfig, error) {
	cfgPath := filepath.Join(projectRoot, defaultGoentDir, defaultProvidersFile)

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultProvidersConfig(), nil
		}
		return nil, fmt.Errorf("read providers config: %w", err)
	}

	var cfg ProvidersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal providers config: %w", err)
	}

	if cfg.OpenCodeConfigPath == "" {
		cfg.OpenCodeConfigPath = defaultOpenCodeConfigPath
	}

	logger := slog.Default()
	if err := substituteEnvVars(&cfg, logger); err != nil {
		return nil, fmt.Errorf("substitute environment variables: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate providers config: %w", err)
	}

	return &cfg, nil
}

func LoadProvidersFromFile(path string) (*ProvidersConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read providers config: %w", err)
	}

	var cfg ProvidersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal providers config: %w", err)
	}

	if cfg.OpenCodeConfigPath == "" {
		cfg.OpenCodeConfigPath = defaultOpenCodeConfigPath
	}

	logger := slog.Default()
	if err := substituteEnvVars(&cfg, logger); err != nil {
		return nil, fmt.Errorf("substitute environment variables: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate providers config: %w", err)
	}

	return &cfg, nil
}

func DefaultOpenCodeConfigPath() string {
	return defaultOpenCodeConfigPath
}

func DefaultCostTrackingConfig() *CostTrackingConfig {
	return &CostTrackingConfig{
		Enabled:        false,
		GlobalBudget:   1000.00,
		ResetPeriod:    ResetMonthly,
		PersistHistory: false,
		HistoryFile:    ".goent/cost_history.json",
	}
}

func DefaultProvidersConfig() *ProvidersConfig {
	return &ProvidersConfig{
		Providers:          make(map[string]ProviderDefinition),
		Defaults:           Defaults{},
		Health:             DefaultHealthConfig(),
		OpenCodeConfigPath: defaultOpenCodeConfigPath,
		CostTracking:       DefaultCostTrackingConfig(),
	}
}

func (c *ProvidersConfig) Validate() error {
	if len(c.Providers) == 0 {
		return ErrNoProviders
	}

	for name, provider := range c.Providers {
		if err := provider.Validate(); err != nil {
			return fmt.Errorf("provider %s: %w", name, err)
		}
	}

	if err := c.Defaults.Validate(c.Providers); err != nil {
		return fmt.Errorf("validate defaults: %w", err)
	}

	if c.Health == nil {
		c.Health = DefaultHealthConfig()
	} else {
		if err := c.Health.Validate(); err != nil {
			return fmt.Errorf("validate health: %w", err)
		}
	}

	if c.CostTracking == nil {
		c.CostTracking = DefaultCostTrackingConfig()
	} else {
		if err := c.CostTracking.Validate(); err != nil {
			return fmt.Errorf("validate cost_tracking: %w", err)
		}
	}

	if c.OpenCodeConfigPath != "" {
		if err := c.validateOpenCodeConfigPath(); err != nil {
			return fmt.Errorf("validate opencode_config_path: %w", err)
		}
	}

	return nil
}

func (p *ProviderDefinition) Validate() error {
	if p.Method == "" {
		return fmt.Errorf("method is required")
	}

	if !p.Method.Valid() {
		return fmt.Errorf("invalid method: %s", p.Method)
	}

	if p.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	if p.Model == "" {
		return fmt.Errorf("model is required")
	}

	validProviders := map[string]bool{
		"anthropic": true,
		"moonshot":  true,
		"deepseek":  true,
	}
	if !validProviders[p.Provider] {
		return fmt.Errorf("unsupported provider: %s", p.Provider)
	}

	if p.ContextLimit < 0 {
		return fmt.Errorf("context_limit cannot be negative")
	}

	if p.Cost != nil {
		if err := p.Cost.Validate(); err != nil {
			return fmt.Errorf("cost config: %w", err)
		}
	}

	return nil
}

func (c *ProvidersConfig) validateOpenCodeConfigPath() error {
	if c.OpenCodeConfigPath == "" {
		return nil
	}

	expandedPath := os.ExpandEnv(c.OpenCodeConfigPath)
	if expandedPath[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("expand home directory: %w", err)
		}
		expandedPath = filepath.Join(homeDir, expandedPath[1:])
	}

	if _, err := os.Stat(expandedPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist: %s", c.OpenCodeConfigPath)
		}
		return fmt.Errorf("check config file: %w", err)
	}

	return nil
}

func (d *Defaults) Validate(providers map[string]ProviderDefinition) error {
	defaultsMap := map[string]string{
		"implementation": d.Implementation,
		"large_context":  d.LargeContext,
		"simple_tasks":   d.SimpleTasks,
		"research":       d.Research,
		"planning":       d.Planning,
		"review":         d.Review,
	}

	for key, providerName := range defaultsMap {
		if providerName == "" {
			continue
		}
		if _, exists := providers[providerName]; !exists {
			return fmt.Errorf("default %s references unknown provider: %s", key, providerName)
		}
	}

	return nil
}

func (c *ProvidersConfig) GetProvider(name string) (ProviderDefinition, bool) {
	provider, exists := c.Providers[name]
	return provider, exists
}

func (c *ProvidersConfig) ListProviders() []string {
	names := make([]string, 0, len(c.Providers))
	for name := range c.Providers {
		names = append(names, name)
	}
	return names
}

func (c *ProvidersConfig) GetDefaultProvider(taskType string) (string, bool) {
	switch taskType {
	case "implementation":
		if c.Defaults.Implementation != "" {
			return c.Defaults.Implementation, true
		}
	case "large-context", "large_context":
		if c.Defaults.LargeContext != "" {
			return c.Defaults.LargeContext, true
		}
	case "simple-tasks", "simple_tasks":
		if c.Defaults.SimpleTasks != "" {
			return c.Defaults.SimpleTasks, true
		}
	case "research":
		if c.Defaults.Research != "" {
			return c.Defaults.Research, true
		}
	case "planning":
		if c.Defaults.Planning != "" {
			return c.Defaults.Planning, true
		}
	case "review":
		if c.Defaults.Review != "" {
			return c.Defaults.Review, true
		}
	}
	return "", false
}

func (c *ProvidersConfig) ValidateProviders(ctx context.Context, logger *slog.Logger) {
	for name, provider := range c.Providers {
		var err error

		switch provider.Method {
		case MethodAPI:
			err = c.validateAPIProvider(ctx, provider, logger)
		case MethodACP:
			err = c.validateACPProvider(ctx, provider, logger)
		case MethodCLI:
			err = c.validateCLIProvider(ctx, provider, logger)
		}

		if err != nil {
			logger.Warn("provider validation failed", "provider", name, "method", provider.Method, "error", err)
		} else {
			logger.Info("provider validation successful", "provider", name, "method", provider.Method)
		}
	}
}

func (c *ProvidersConfig) validateAPIProvider(ctx context.Context, p ProviderDefinition, logger *slog.Logger) error {
	switch p.Provider {
	case "anthropic":
		client, err := provider.NewAnthropicClient(logger)
		if err != nil {
			return err
		}
		return client.Validate(ctx)

	case "moonshot":
		client, err := provider.NewOpenAICompatClient(provider.ProviderMoonshot, logger)
		if err != nil {
			return err
		}
		return client.Validate(ctx)

	case "deepseek":
		client, err := provider.NewOpenAICompatClient(provider.ProviderDeepSeek, logger)
		if err != nil {
			return err
		}
		return client.Validate(ctx)

	default:
		return fmt.Errorf("unsupported api provider: %s", p.Provider)
	}
}

func (c *ProvidersConfig) validateACPProvider(ctx context.Context, p ProviderDefinition, logger *slog.Logger) error {
	cfg := opencode.Config{
		ConfigPath: c.OpenCodeConfigPath,
	}

	client, err := opencode.NewACPClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("create acp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if err := client.Validate(ctx); err != nil {
		return fmt.Errorf("validate acp: %w", err)
	}

	return nil
}

func (c *ProvidersConfig) validateCLIProvider(ctx context.Context, p ProviderDefinition, logger *slog.Logger) error {
	client := opencode.NewCLIClient(c.OpenCodeConfigPath)

	if err := client.Validate(ctx); err != nil {
		return fmt.Errorf("validate cli: %w", err)
	}

	return nil
}
