package model

func DefaultConfig() *Config {
	return &Config{
		Version: "1.0",
		Runtimes: map[string]Mapping{
			string(RuntimeClaude): {
				Fast:  "claude-haiku-4-5-20250429",
				Main:  "claude-sonnet-4-5-20250929",
				Heavy: "claude-opus-4-5-20250514",
			},
			string(RuntimeOpenCode): {
				Fast:  "zai-coding-plan/glm-4.7",
				Main:  "zai-coding-plan/glm-4.7",
				Heavy: "kimi-for-coding/kimi-k2-thinking",
			},
		},
		Aliases: map[string]string{
			"haiku":       string(Fast),
			"sonnet":      string(Main),
			"opus":        string(Heavy),
			"glm-4-flash": string(Fast),
			"kimi-k2":     string(Heavy),
		},
	}
}
