package config

// DefaultModelConfig returns the default model configuration when no models.yaml file exists.
// Provides sensible defaults for model mappings across different runtimes.
func DefaultModelConfig() *ModelConfig {
	return &ModelConfig{
		Version: "1.0",
		Runtimes: map[string]ModelMapping{
			string(ModelRuntimeClaude): {
				Fast:  "claude-haiku-4-5-20251001",
				Main:  "claude-sonnet-4-5-20250929",
				Heavy: "claude-opus-4-5-20251101",
			},
			string(ModelRuntimeOpenCode): {
				Fast:  "zai-coding-plan/glm-4.5-air",
				Main:  "zai-coding-plan/glm-4.7",
				Heavy: "kimi-for-coding/k2p5",
			},
		},
		Aliases: map[string]string{
			"haiku":       string(ModelFast),
			"sonnet":      string(ModelMain),
			"opus":        string(ModelHeavy),
			"glm-4-flash": string(ModelFast),
			"kimi-k2":     string(ModelHeavy),
		},
	}
}
