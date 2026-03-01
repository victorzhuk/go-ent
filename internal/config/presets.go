package config

// ModelPreset is a known model option for a given tier.
type ModelPreset struct {
	ID          string
	Description string
}

// RuntimePresets holds known model presets per tier for a runtime.
type RuntimePresets struct {
	Fast  []ModelPreset
	Main  []ModelPreset
	Heavy []ModelPreset
}

var runtimePresets = map[string]RuntimePresets{
	"claude": {
		Fast:  []ModelPreset{{ID: "haiku", Description: "Fast, cost-efficient"}},
		Main:  []ModelPreset{{ID: "sonnet", Description: "Balanced performance"}},
		Heavy: []ModelPreset{{ID: "opus", Description: "Most capable"}},
	},
	"opencode": {
		Fast: []ModelPreset{
			{ID: "anthropic/claude-haiku-4-5", Description: "Anthropic Haiku — fast"},
			{ID: "openai/gpt-4o-mini", Description: "OpenAI GPT-4o mini"},
		},
		Main: []ModelPreset{
			{ID: "anthropic/claude-sonnet-4-6", Description: "Anthropic Sonnet — balanced"},
			{ID: "openai/gpt-4o", Description: "OpenAI GPT-4o"},
		},
		Heavy: []ModelPreset{
			{ID: "anthropic/claude-opus-4-6", Description: "Anthropic Opus — most capable"},
			{ID: "openai/o1", Description: "OpenAI o1"},
		},
	},
}

// DefaultModelsForRuntime returns the default model tiers for a given runtime.
func DefaultModelsForRuntime(runtime string) ModelTiers {
	switch runtime {
	case "claude":
		return ModelTiers{
			Fast:  "haiku",
			Main:  "sonnet",
			Heavy: "opus",
		}
	case "opencode":
		return ModelTiers{
			Fast:  "anthropic/claude-haiku-4-5",
			Main:  "anthropic/claude-sonnet-4-6",
			Heavy: "anthropic/claude-opus-4-6",
		}
	default:
		return ModelTiers{}
	}
}

// PresetsForRuntime returns the known model presets for a given runtime.
func PresetsForRuntime(runtime string) (RuntimePresets, bool) {
	p, ok := runtimePresets[runtime]
	return p, ok
}
