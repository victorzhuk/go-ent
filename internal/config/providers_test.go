package config

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadProviders(t *testing.T) {
	t.Run("file not found returns default config", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := LoadProviders(tmpDir)

		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.NotNil(t, cfg.Providers)
		assert.Empty(t, cfg.Providers)
		assert.NotNil(t, cfg.Health)
		assert.Equal(t, 30, cfg.Health.CheckInterval)
		assert.Equal(t, 300, cfg.Health.WorkerTimeout)
	})

	t.Run("valid config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".goent")
		require.NoError(t, os.Mkdir(configDir, 0755))

		configContent := `
providers:
  glm:
    method: acp
    provider: moonshot
    model: glm-4
    best_for: ["bulk", "implementation"]
    cost:
      per_1k_tokens: 0.01
      per_hour: 5.00
      per_day: 100.00
      per_month: 1000.00
    context_limit: 8192

  kimi:
    method: acp
    provider: moonshot
    model: kimi-k2
    best_for: ["large-context", "file-analysis"]
    cost:
      per_1k_tokens: 0.02
      per_hour: 6.00
      per_day: 120.00
      per_month: 1200.00
    context_limit: 128000

  haiku:
    method: api
    provider: anthropic
    model: claude-3-haiku
    best_for: ["simple", "quick"]
    cost:
      per_1k_tokens: 0.25
      per_hour: 10.00
      per_day: 200.00
      per_month: 2000.00

defaults:
  implementation: glm
  large_context: kimi
  simple_tasks: haiku

health:
  check_interval: 60
  worker_timeout: 600
  max_retries: 5
  retry_delay: 10
`
		configPath := filepath.Join(configDir, "providers.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		cfg, err := LoadProviders(tmpDir)

		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Len(t, cfg.Providers, 3)

		glm, exists := cfg.Providers["glm"]
		assert.True(t, exists)
		assert.Equal(t, MethodACP, glm.Method)
		assert.Equal(t, "moonshot", glm.Provider)
		assert.Equal(t, "glm-4", glm.Model)
		assert.Equal(t, []string{"bulk", "implementation"}, glm.BestFor)
		assert.NotNil(t, glm.Cost)
		assert.Equal(t, 0.01, glm.Cost.Per1kTokens)
		assert.Equal(t, 5.00, glm.Cost.PerHour)
		assert.Equal(t, 100.00, glm.Cost.PerDay)
		assert.Equal(t, 1000.00, glm.Cost.PerMonth)
		assert.Equal(t, 8192, glm.ContextLimit)

		kimi, exists := cfg.Providers["kimi"]
		assert.True(t, exists)
		assert.Equal(t, MethodACP, kimi.Method)
		assert.Equal(t, "moonshot", kimi.Provider)
		assert.Equal(t, "kimi-k2", kimi.Model)
		assert.Equal(t, []string{"large-context", "file-analysis"}, kimi.BestFor)
		assert.NotNil(t, kimi.Cost)
		assert.Equal(t, 0.02, kimi.Cost.Per1kTokens)
		assert.Equal(t, 6.00, kimi.Cost.PerHour)
		assert.Equal(t, 120.00, kimi.Cost.PerDay)
		assert.Equal(t, 1200.00, kimi.Cost.PerMonth)
		assert.Equal(t, 128000, kimi.ContextLimit)

		assert.Equal(t, "glm", cfg.Defaults.Implementation)
		assert.Equal(t, "kimi", cfg.Defaults.LargeContext)
		assert.Equal(t, "haiku", cfg.Defaults.SimpleTasks)

		assert.Equal(t, 60, cfg.Health.CheckInterval)
		assert.Equal(t, 600, cfg.Health.WorkerTimeout)
		assert.Equal(t, 5, cfg.Health.MaxRetries)
		assert.Equal(t, 10, cfg.Health.RetryDelay)
	})

	t.Run("invalid provider method", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".goent")
		require.NoError(t, os.Mkdir(configDir, 0755))

		configContent := `
providers:
  test:
    method: invalid
    provider: anthropic
    model: claude-3-haiku
`
		configPath := filepath.Join(configDir, "providers.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		cfg, err := LoadProviders(tmpDir)

		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "invalid method")
	})

	t.Run("missing required fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".goent")
		require.NoError(t, os.Mkdir(configDir, 0755))

		configContent := `
providers:
  test:
    method: acp
`
		configPath := filepath.Join(configDir, "providers.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		cfg, err := LoadProviders(tmpDir)

		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("default provider mapping", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".goent")
		require.NoError(t, os.Mkdir(configDir, 0755))

		configContent := `
providers:
  glm:
    method: acp
    provider: moonshot
    model: glm-4

  haiku:
    method: api
    provider: anthropic
    model: claude-3-haiku

defaults:
  implementation: glm
  simple_tasks: haiku
`
		configPath := filepath.Join(configDir, "providers.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		cfg, err := LoadProviders(tmpDir)

		require.NoError(t, err)

		provider, exists := cfg.GetDefaultProvider("implementation")
		assert.True(t, exists)
		assert.Equal(t, "glm", provider)

		provider, exists = cfg.GetDefaultProvider("simple_tasks")
		assert.True(t, exists)
		assert.Equal(t, "haiku", provider)

		provider, exists = cfg.GetDefaultProvider("unknown")
		assert.False(t, exists)
		assert.Equal(t, "", provider)
	})

	t.Run("validate invalid default", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".goent")
		require.NoError(t, os.Mkdir(configDir, 0755))

		configContent := `
providers:
  glm:
    method: acp
    provider: moonshot
    model: glm-4

defaults:
  implementation: unknown_provider
`
		configPath := filepath.Join(configDir, "providers.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		_, err := LoadProviders(tmpDir)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "references unknown provider")
	})

	t.Run("opencode config path provided", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".goent")
		require.NoError(t, os.Mkdir(configDir, 0755))

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		opencodeDir := filepath.Join(homeDir, ".config", "opencode")
		require.NoError(t, os.MkdirAll(opencodeDir, 0755))
		configFile := filepath.Join(opencodeDir, "opencode.json")
		require.NoError(t, os.WriteFile(configFile, []byte("{}"), 0644))

		configContent := `
providers:
  glm:
    method: acp
    provider: moonshot
    model: glm-4

opencode_config_path: ~/.config/opencode/opencode.json
`
		configPath := filepath.Join(configDir, "providers.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		cfg, err := LoadProviders(tmpDir)

		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "~/.config/opencode/opencode.json", cfg.OpenCodeConfigPath)
	})

	t.Run("opencode config path not provided uses default", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".goent")
		require.NoError(t, os.Mkdir(configDir, 0755))

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		opencodeDir := filepath.Join(homeDir, ".config", "opencode")
		require.NoError(t, os.MkdirAll(opencodeDir, 0755))
		configFile := filepath.Join(opencodeDir, "opencode.json")
		require.NoError(t, os.WriteFile(configFile, []byte("{}"), 0644))

		configContent := `
providers:
  glm:
    method: acp
    provider: moonshot
    model: glm-4
`
		configPath := filepath.Join(configDir, "providers.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		cfg, err := LoadProviders(tmpDir)

		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, defaultOpenCodeConfigPath, cfg.OpenCodeConfigPath)
	})

	t.Run("opencode config path validation - file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".goent")
		require.NoError(t, os.Mkdir(configDir, 0755))

		configContent := `
providers:
  glm:
    method: acp
    provider: moonshot
    model: glm-4

opencode_config_path: /nonexistent/path/opencode.json
`
		configPath := filepath.Join(configDir, "providers.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		_, err := LoadProviders(tmpDir)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config file does not exist")
	})
}

func TestProviderDefinition_Validate(t *testing.T) {
	t.Run("valid provider", func(t *testing.T) {
		provider := ProviderDefinition{
			Method:   MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
			BestFor:  []string{"bulk", "implementation"},
			Cost: &CostConfig{
				Per1kTokens: 0.01,
				PerHour:     5.00,
				PerDay:      100.00,
				PerMonth:    1000.00,
				Multipliers: map[string]float64{
					"acp": 1.5,
					"cli": 0.5,
					"api": 1.0,
				},
			},
			ContextLimit: 8192,
		}

		err := provider.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing method", func(t *testing.T) {
		provider := ProviderDefinition{
			Provider: "moonshot",
			Model:    "glm-4",
		}

		err := provider.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method is required")
	})

	t.Run("invalid method", func(t *testing.T) {
		provider := ProviderDefinition{
			Method:   CommunicationMethod("invalid"),
			Provider: "moonshot",
			Model:    "glm-4",
		}

		err := provider.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid method")
	})

	t.Run("unsupported provider", func(t *testing.T) {
		provider := ProviderDefinition{
			Method:   MethodAPI,
			Provider: "unknown",
			Model:    "test-model",
		}

		err := provider.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported provider")
	})

	t.Run("negative context limit", func(t *testing.T) {
		provider := ProviderDefinition{
			Method:       MethodACP,
			Provider:     "moonshot",
			Model:        "glm-4",
			ContextLimit: -1,
		}

		err := provider.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context_limit cannot be negative")
	})
}

func TestCommunicationMethod(t *testing.T) {
	t.Run("valid methods", func(t *testing.T) {
		assert.True(t, MethodACP.Valid())
		assert.True(t, MethodCLI.Valid())
		assert.True(t, MethodAPI.Valid())
	})

	t.Run("invalid method", func(t *testing.T) {
		invalidMethod := CommunicationMethod("invalid")
		assert.False(t, invalidMethod.Valid())
	})

	t.Run("string representation", func(t *testing.T) {
		assert.Equal(t, "acp", MethodACP.String())
		assert.Equal(t, "cli", MethodCLI.String())
		assert.Equal(t, "api", MethodAPI.String())
		assert.Equal(t, "unknown", CommunicationMethod("").String())
	})
}

func TestHealthConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		cfg := DefaultHealthConfig()

		assert.Equal(t, 30, cfg.CheckInterval)
		assert.Equal(t, 300, cfg.WorkerTimeout)
		assert.Equal(t, 3, cfg.MaxRetries)
		assert.Equal(t, 5, cfg.RetryDelay)
	})

	t.Run("valid config", func(t *testing.T) {
		cfg := &HealthConfig{
			CheckInterval: 60,
			WorkerTimeout: 600,
			MaxRetries:    5,
			RetryDelay:    10,
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("negative values", func(t *testing.T) {
		cfg := &HealthConfig{
			CheckInterval: -1,
			WorkerTimeout: 0,
			MaxRetries:    -5,
			RetryDelay:    -10,
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "check_interval must be positive")
	})
}

func TestProvidersConfig_ValidateProviders(t *testing.T) {
	t.Run("validates api provider", func(t *testing.T) {
		if os.Getenv("ANTHROPIC_API_KEY") == "" {
			t.Skip("ANTHROPIC_API_KEY not set - skipping integration test")
		}

		cfg := &ProvidersConfig{
			Providers: map[string]ProviderDefinition{
				"haiku": {
					Method:   MethodAPI,
					Provider: "anthropic",
					Model:    "claude-3-haiku-20240307",
				},
			},
		}

		ctx := context.Background()
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		cfg.ValidateProviders(ctx, logger)

		t.Log("Validation completed without panic")
	})

	t.Run("validates acp provider", func(t *testing.T) {
		cfg := &ProvidersConfig{
			Providers: map[string]ProviderDefinition{
				"glm": {
					Method:   MethodACP,
					Provider: "moonshot",
					Model:    "glm-4",
				},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		cfg.ValidateProviders(ctx, logger)

		t.Log("Validation completed without panic")
	})

	t.Run("validates cli provider", func(t *testing.T) {
		cfg := &ProvidersConfig{
			Providers: map[string]ProviderDefinition{
				"glm": {
					Method:   MethodCLI,
					Provider: "moonshot",
					Model:    "glm-4",
				},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		cfg.ValidateProviders(ctx, logger)

		t.Log("Validation completed without panic")
	})

	t.Run("skips invalid provider", func(t *testing.T) {
		cfg := &ProvidersConfig{
			Providers: map[string]ProviderDefinition{
				"invalid": {
					Method:   MethodAPI,
					Provider: "anthropic",
					Model:    "claude-3-haiku-20240307",
				},
			},
		}

		ctx := context.Background()
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		cfg.ValidateProviders(ctx, logger)

		t.Log("Validation completed without panic for invalid provider")
	})
}

func TestLoadProvidersWithEnvSubstitution(t *testing.T) {
	t.Run("braced variable substitution", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".goent")
		require.NoError(t, os.Mkdir(configDir, 0755))

		t.Setenv("TEST_PROVIDER", "anthropic")
		t.Setenv("TEST_MODEL", "claude-3-haiku")

		configContent := `
providers:
  test:
    method: api
    provider: ${TEST_PROVIDER}
    model: ${TEST_MODEL}
`
		configPath := filepath.Join(configDir, "providers.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		cfg, err := LoadProviders(tmpDir)

		require.NoError(t, err)
		assert.NotNil(t, cfg)
		testProvider, exists := cfg.Providers["test"]
		assert.True(t, exists)
		assert.Equal(t, "anthropic", testProvider.Provider)
		assert.Equal(t, "claude-3-haiku", testProvider.Model)
	})

	t.Run("simple variable substitution", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".goent")
		require.NoError(t, os.Mkdir(configDir, 0755))

		t.Setenv("PROVIDER_NAME", "moonshot")
		t.Setenv("MODEL_NAME", "glm-4")

		configContent := `
providers:
  test:
    method: acp
    provider: $PROVIDER_NAME
    model: $MODEL_NAME
`
		configPath := filepath.Join(configDir, "providers.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		cfg, err := LoadProviders(tmpDir)

		require.NoError(t, err)
		assert.NotNil(t, cfg)
		testProvider, exists := cfg.Providers["test"]
		assert.True(t, exists)
		assert.Equal(t, "moonshot", testProvider.Provider)
		assert.Equal(t, "glm-4", testProvider.Model)
	})

	t.Run("default value with variable missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".goent")
		require.NoError(t, os.Mkdir(configDir, 0755))

		configContent := `
providers:
  test:
    method: api
    provider: anthropic
    model: ${MISSING_MODEL:-claude-3-haiku}
`
		configPath := filepath.Join(configDir, "providers.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		cfg, err := LoadProviders(tmpDir)

		require.NoError(t, err)
		assert.NotNil(t, cfg)
		testProvider, exists := cfg.Providers["test"]
		assert.True(t, exists)
		assert.Equal(t, "claude-3-haiku", testProvider.Model)
	})

	t.Run("default value with variable present", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".goent")
		require.NoError(t, os.Mkdir(configDir, 0755))

		t.Setenv("CUSTOM_MODEL", "claude-3-sonnet")

		configContent := `
providers:
  test:
    method: api
    provider: anthropic
    model: ${CUSTOM_MODEL:-default-model}
`
		configPath := filepath.Join(configDir, "providers.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		cfg, err := LoadProviders(tmpDir)

		require.NoError(t, err)
		assert.NotNil(t, cfg)
		testProvider, exists := cfg.Providers["test"]
		assert.True(t, exists)
		assert.Equal(t, "claude-3-sonnet", testProvider.Model)
	})

	t.Run("missing required variable causes error", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".goent")
		require.NoError(t, os.Mkdir(configDir, 0755))

		configContent := `
providers:
  test:
    method: api
    provider: ${MISSING_PROVIDER}
    model: claude-3-haiku
`
		configPath := filepath.Join(configDir, "providers.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		_, err := LoadProviders(tmpDir)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrMissingEnvVar)
	})

	t.Run("opencode_config_path with substitution", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".goent")
		require.NoError(t, os.Mkdir(configDir, 0755))

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		opencodeDir := filepath.Join(homeDir, ".config", "opencode")
		require.NoError(t, os.MkdirAll(opencodeDir, 0755))
		configFile := filepath.Join(opencodeDir, "opencode.json")
		require.NoError(t, os.WriteFile(configFile, []byte("{}"), 0644))

		configContent := `
providers:
  test:
    method: api
    provider: anthropic
    model: claude-3-haiku

opencode_config_path: ${HOME:-/tmp}/.config/opencode/opencode.json
`
		configPath := filepath.Join(configDir, "providers.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		cfg, err := LoadProviders(tmpDir)

		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Contains(t, cfg.OpenCodeConfigPath, "/.config/opencode/opencode.json")
	})

}
