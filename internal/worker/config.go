package worker

import (
	"fmt"
	"time"

	"github.com/victorzhuk/go-ent/internal/config"
)

type ProviderDefinition config.ProviderDefinition

type Config struct {
	Providers          map[string]ProviderDefinition
	Defaults           Defaults
	Health             *HealthConfig
	OpenCodeConfigPath string
	CostTracking       *config.CostTrackingConfig
}

type Defaults config.Defaults

type HealthConfig struct {
	CheckInterval time.Duration
	WorkerTimeout time.Duration
	MaxRetries    int
	RetryDelay    time.Duration
}

func DefaultHealthConfig() *HealthConfig {
	return &HealthConfig{
		CheckInterval: 30 * time.Second,
		WorkerTimeout: 5 * time.Minute,
		MaxRetries:    3,
		RetryDelay:    5 * time.Second,
	}
}

func Load(projectRoot string) (*Config, error) {
	providersCfg, err := config.LoadProviders(projectRoot)
	if err != nil {
		return nil, err
	}

	return FromProvidersConfig(providersCfg), nil
}

func LoadFromFile(path string) (*Config, error) {
	providersCfg, err := config.LoadProvidersFromFile(path)
	if err != nil {
		return nil, err
	}

	return FromProvidersConfig(providersCfg), nil
}

func DefaultConfig() *Config {
	return &Config{
		Providers:          make(map[string]ProviderDefinition),
		Defaults:           Defaults{},
		Health:             DefaultHealthConfig(),
		OpenCodeConfigPath: config.DefaultOpenCodeConfigPath(),
		CostTracking:       config.DefaultCostTrackingConfig(),
	}
}

func FromProvidersConfig(pc *config.ProvidersConfig) *Config {
	cfg := &Config{
		Providers:          make(map[string]ProviderDefinition, len(pc.Providers)),
		Defaults:           Defaults(pc.Defaults),
		Health:             DefaultHealthConfig(),
		OpenCodeConfigPath: pc.OpenCodeConfigPath,
		CostTracking:       pc.CostTracking,
	}

	for name, def := range pc.Providers {
		cfg.Providers[name] = ProviderDefinition(def)
	}

	if pc.Health != nil {
		cfg.Health.CheckInterval = time.Duration(pc.Health.CheckInterval) * time.Second
		cfg.Health.WorkerTimeout = time.Duration(pc.Health.WorkerTimeout) * time.Second
		cfg.Health.MaxRetries = pc.Health.MaxRetries
		cfg.Health.RetryDelay = time.Duration(pc.Health.RetryDelay) * time.Second
	}

	return cfg
}

func (c *Config) GetProvider(name string) (ProviderDefinition, bool) {
	provider, exists := c.Providers[name]
	return provider, exists
}

func (c *Config) ListProviders() []string {
	names := make([]string, 0, len(c.Providers))
	for name := range c.Providers {
		names = append(names, name)
	}
	return names
}

func (c *Config) GetDefaultProvider(taskType string) (string, bool) {
	pc := config.ProvidersConfig{
		Defaults: config.Defaults(c.Defaults),
	}
	return pc.GetDefaultProvider(taskType)
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

func (c *Config) Validate() error {
	pc := &config.ProvidersConfig{
		Providers:    make(map[string]config.ProviderDefinition, len(c.Providers)),
		Defaults:     config.Defaults(c.Defaults),
		CostTracking: c.CostTracking,
	}

	for name, def := range c.Providers {
		pc.Providers[name] = config.ProviderDefinition(def)
	}

	return pc.Validate()
}
