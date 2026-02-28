// Package config provides runtime configuration management for go-ent tools.
// It handles loading model configurations for different AI platforms (Claude, OpenCode)
// and provides model alias resolution.
//
// Configuration is loaded from tool-specific YAML files and supports:
// - Model aliases (fast, main, heavy) mapping to platform-specific models
// - Tool presets for agent configuration
// - Runtime detection from environment
package config
