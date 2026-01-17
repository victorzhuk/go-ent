package config

import "github.com/victorzhuk/go-ent/internal/domain"

// DefaultConfig returns the default configuration when no config file exists.
// Provides sensible defaults for all configuration sections.
//
// Default Values:
//
// Version:
//   - version: "1.0" (current config format version)
//
// Agents:
//   - default: senior (balanced capability/cost ratio)
//   - architect: opus model with go-arch, go-api skills (system design)
//   - senior: sonnet model with go-code, go-db, go-test skills (implementation)
//   - developer: sonnet model with go-code, go-test skills (focused coding)
//   - delegation: auto=false (explicit agent selection required)
//   - model_tier: exploration=haiku, complexity=sonnet, critical=opus (task-based model selection)
//
// Runtime:
//   - preferred: claude-code (best integration with Claude ecosystem)
//   - fallback: [cli] (standalone execution when claude-code unavailable)
//
// Budget:
//   - daily: $10.00 USD (reasonable daily spend limit)
//   - monthly: $200.00 USD (typical project budget)
//   - per_task: $1.00 USD (prevents runaway costs on single tasks)
//   - tracking: true (enables cost monitoring)
//
// Models:
//   - opus: claude-opus-4-5-20251101 (highest capability, for architecture)
//   - sonnet: claude-sonnet-4-5-20251101 (balanced, for implementation)
//   - haiku: claude-haiku-3-5-20241022 (fastest, for simple tasks)
//
// Background:
//   - max_concurrent: 5 (concurrent background tasks limit)
//   - default_role: developer (standard implementation role)
//   - default_model: haiku (fast model for async tasks)
//   - timeout: 300s (5 minute execution limit per task)
//   - resource_limits:
//   - max_memory_mb: 512 (memory limit per agent)
//   - max_goroutines: 100 (goroutine limit per agent)
//   - max_cpu_percent: 80 (CPU limit per agent)
//
// Skills:
//   - Enabled: [go-code, go-arch, go-api, go-db, go-test] (core Go enterprise skills)
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
			ModelTier: ModelTierConfig{
				Exploration: "haiku",
				Complexity:  "sonnet",
				Critical:    "opus",
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
		Background: BackgroundConfig{
			MaxConcurrent: 5,
			DefaultRole:   "developer",
			DefaultModel:  "haiku",
			Timeout:       300,
			ResourceLimits: ResourceLimits{
				MaxMemoryMB:   512,
				MaxGoroutines: 100,
				MaxCPUPercent: 80,
			},
		},
	}
}
