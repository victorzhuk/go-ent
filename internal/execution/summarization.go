package execution

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	summarizationConfigDir  = ".go-ent"
	summarizationConfigFile = "summarization.yaml"
)

// LoadSummarizationThreshold loads summarization thresholds from .go-ent/summarization.yaml.
// Returns default thresholds if file doesn't exist.
func LoadSummarizationThreshold(projectPath string) (SummarizationThreshold, error) {
	configPath := filepath.Join(projectPath, summarizationConfigDir, summarizationConfigFile)

	data, err := os.ReadFile(configPath) //nolint:gosec
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultSummarizationThreshold(), nil
		}
		return SummarizationThreshold{}, fmt.Errorf("read summarization config: %w", err)
	}

	var config struct {
		FileCount     int    `yaml:"file_count"`
		ContextLength int    `yaml:"context_length"`
		TokenCount    int    `yaml:"token_count"`
		Model         string `yaml:"model"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return SummarizationThreshold{}, fmt.Errorf("parse summarization config: %w", err)
	}

	// Validate and apply defaults
	threshold := DefaultSummarizationThreshold()
	if config.FileCount > 0 {
		threshold.FileCount = config.FileCount
	}
	if config.ContextLength > 0 {
		threshold.ContextLength = config.ContextLength
	}
	if config.TokenCount > 0 {
		threshold.TokenCount = config.TokenCount
	}

	return threshold, nil
}

// SaveSummarizationThreshold saves summarization thresholds to .go-ent/summarization.yaml.
func SaveSummarizationThreshold(projectPath string, threshold SummarizationThreshold, model string) error {
	configDir := filepath.Join(projectPath, summarizationConfigDir)
	configPath := filepath.Join(configDir, summarizationConfigFile)

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	config := struct {
		FileCount     int    `yaml:"file_count"`
		ContextLength int    `yaml:"context_length"`
		TokenCount    int    `yaml:"token_count"`
		Model         string `yaml:"model"`
	}{
		FileCount:     threshold.FileCount,
		ContextLength: threshold.ContextLength,
		TokenCount:    threshold.TokenCount,
		Model:         model,
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal summarization config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("write summarization config: %w", err)
	}

	return nil
}
