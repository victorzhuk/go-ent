package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/victorzhuk/go-ent/internal/domain"
)

func TestModelTierConfigValidate(t *testing.T) {
	t.Run("validates all fields present", func(t *testing.T) {
		t.Parallel()

		cfg := ModelTierConfig{
			Exploration: "haiku",
			Complexity:  "sonnet",
			Critical:    "opus",
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("skips validation when all fields empty", func(t *testing.T) {
		t.Parallel()

		cfg := ModelTierConfig{}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("returns error when exploration empty", func(t *testing.T) {
		t.Parallel()

		cfg := ModelTierConfig{
			Exploration: "",
			Complexity:  "sonnet",
			Critical:    "opus",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidAgentConfig)
	})

	t.Run("returns error when complexity empty", func(t *testing.T) {
		t.Parallel()

		cfg := ModelTierConfig{
			Exploration: "haiku",
			Complexity:  "",
			Critical:    "opus",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidAgentConfig)
	})

	t.Run("returns error when critical empty", func(t *testing.T) {
		t.Parallel()

		cfg := ModelTierConfig{
			Exploration: "haiku",
			Complexity:  "sonnet",
			Critical:    "",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidAgentConfig)
	})

	t.Run("returns error when multiple fields empty", func(t *testing.T) {
		t.Parallel()

		cfg := ModelTierConfig{
			Exploration: "",
			Complexity:  "",
			Critical:    "opus",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidAgentConfig)
	})
}

func TestAgentsConfigValidateWithModelTier(t *testing.T) {
	t.Run("validates valid model tier config", func(t *testing.T) {
		t.Parallel()

		cfg := AgentsConfig{
			Default: domain.AgentRoleSenior,
			Roles: map[string]AgentRoleConfig{
				string(domain.AgentRoleSenior): {
					Model: "sonnet",
				},
			},
			ModelTier: ModelTierConfig{
				Exploration: "haiku",
				Complexity:  "sonnet",
				Critical:    "opus",
			},
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("validates with empty model tier config", func(t *testing.T) {
		t.Parallel()

		cfg := AgentsConfig{
			Default: domain.AgentRoleSenior,
			Roles: map[string]AgentRoleConfig{
				string(domain.AgentRoleSenior): {
					Model: "sonnet",
				},
			},
			ModelTier: ModelTierConfig{},
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("returns error on invalid model tier config", func(t *testing.T) {
		t.Parallel()

		cfg := AgentsConfig{
			Default: domain.AgentRoleSenior,
			Roles: map[string]AgentRoleConfig{
				string(domain.AgentRoleSenior): {
					Model: "sonnet",
				},
			},
			ModelTier: ModelTierConfig{
				Exploration: "haiku",
				Complexity:  "",
				Critical:    "opus",
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidAgentConfig)
	})
}
