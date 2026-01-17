package config

import (
	"testing"
)

func TestMetricsConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := MetricsConfig{
			Enabled: true,
		}
		err := cfg.Validate()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("disabled config", func(t *testing.T) {
		cfg := MetricsConfig{
			Enabled: false,
		}
		err := cfg.Validate()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("default config", func(t *testing.T) {
		cfg := MetricsConfig{}
		err := cfg.Validate()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestDefaultConfig_Metrics(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Metrics.Enabled != true {
		t.Errorf("expected metrics enabled to be true by default, got %v", cfg.Metrics.Enabled)
	}
}

func TestLoadWithEnv_Metrics(t *testing.T) {
	t.Run("overrides metrics enabled from env", func(t *testing.T) {
		getenv := func(key string) string {
			if key == "GOENT_METRICS_ENABLED" {
				return "false"
			}
			return ""
		}

		cfg, err := LoadWithEnv(".", getenv)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Metrics.Enabled {
			t.Errorf("expected metrics to be disabled from env, got %v", cfg.Metrics.Enabled)
		}
	})

	t.Run("overrides metrics disabled from env", func(t *testing.T) {
		getenv := func(key string) string {
			if key == "GOENT_METRICS_ENABLED" {
				return "true"
			}
			return ""
		}

		cfg, err := LoadWithEnv(".", getenv)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !cfg.Metrics.Enabled {
			t.Errorf("expected metrics to be enabled from env, got %v", cfg.Metrics.Enabled)
		}
	})

	t.Run("invalid metrics enabled value defaults to false", func(t *testing.T) {
		getenv := func(key string) string {
			if key == "GOENT_METRICS_ENABLED" {
				return "invalid"
			}
			return ""
		}

		cfg, err := LoadWithEnv(".", getenv)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Metrics.Enabled {
			t.Errorf("expected invalid boolean value to default to false, got %v", cfg.Metrics.Enabled)
		}
	})
}
