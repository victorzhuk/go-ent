// Package config provides hierarchical configuration management for go-ent.
//
// # Overview
//
// The config package implements a YAML-based configuration system with
// environment variable overrides. Configuration is loaded from
// .go-ent/config.yaml in the project root, with sensible defaults when
// no config file exists.
//
// # Configuration Sections
//
// The configuration is organized into five main sections:
//
//   - Agents: Configure agent roles, models, and skills
//   - Runtime: Execution environment preferences (claude-code, opencode, cli)
//   - Budget: Spending limits and cost tracking
//   - Models: Friendly name mappings to model IDs
//   - Skills: Enabled skills and custom skill directories
//
// # Loading Configuration
//
// Basic usage:
//
//	cfg, err := config.Load("/path/to/project")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// With environment variable overrides:
//
//	cfg, err := config.LoadWithEnv("/path/to/project", os.Getenv)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Environment Variables
//
// The following environment variables override config file values:
//
//   - GOENT_BUDGET_DAILY: Override daily spending limit (float, USD)
//   - GOENT_BUDGET_MONTHLY: Override monthly spending limit (float, USD)
//   - GOENT_BUDGET_PER_TASK: Override per-task spending limit (float, USD)
//   - GOENT_RUNTIME_PREFERRED: Override preferred runtime (claude-code|opencode|cli)
//   - GOENT_AGENTS_DEFAULT: Override default agent role (architect|senior|developer|...)
//
// # Example Configuration
//
// A typical .go-ent/config.yaml file:
//
//	version: "1.0"
//
//	agents:
//	  default: senior
//	  roles:
//	    architect:
//	      model: opus
//	      skills: [go-arch, go-api]
//	    senior:
//	      model: sonnet
//	      skills: [go-code, go-db, go-test]
//
//	runtime:
//	  preferred: claude-code
//	  fallback: [cli]
//
//	budget:
//	  daily: 10.0
//	  monthly: 200.0
//	  per_task: 1.0
//	  tracking: true
//
//	models:
//	  opus: claude-opus-4-5-20251101
//	  sonnet: claude-sonnet-4-5-20251101
//	  haiku: claude-haiku-3-5-20241022
//
//	skills:
//	  enabled: [go-code, go-arch, go-api, go-db, go-test]
//
// # Default Configuration
//
// When no config file exists, DefaultConfig() provides sensible defaults:
//
//   - Default agent role: senior (balanced capability/cost)
//   - Preferred runtime: claude-code
//   - Budget: $10 daily, $200 monthly, $1 per-task
//   - Models: opus, sonnet, and haiku with latest stable versions
//   - Skills: Core Go enterprise skills enabled
//
// See DefaultConfig() documentation for complete default values.
//
// # Validation
//
// All configuration is validated on load:
//
//   - Agent role names must be valid domain.AgentRole values
//   - Runtime names must be valid domain.Runtime values
//   - Budget values must be non-negative
//   - Model mappings must be non-empty
//   - Skills must reference valid skill IDs (when skill registry is available)
//
// Validation errors are returned with descriptive messages indicating
// which field failed validation and why.
package config
