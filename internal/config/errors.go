package config

import "errors"

var (
	// ErrInvalidConfig indicates the configuration is invalid.
	ErrInvalidConfig = errors.New("invalid config")

	// ErrInvalidAgentConfig indicates the agent configuration is invalid.
	ErrInvalidAgentConfig = errors.New("invalid agent config")

	// ErrInvalidRuntimeConfig indicates the runtime configuration is invalid.
	ErrInvalidRuntimeConfig = errors.New("invalid runtime config")

	// ErrInvalidBudgetConfig indicates the budget configuration is invalid.
	ErrInvalidBudgetConfig = errors.New("invalid budget config")

	// ErrInvalidModelConfig indicates the model configuration is invalid.
	ErrInvalidModelConfig = errors.New("invalid model config")

	// ErrConfigNotFound indicates the configuration file was not found.
	ErrConfigNotFound = errors.New("config file not found")

	// ErrInvalidYAML indicates the configuration file contains invalid YAML.
	ErrInvalidYAML = errors.New("invalid yaml")
)
