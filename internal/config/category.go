package config

type ModelCategory string

const (
	ModelFast  ModelCategory = "fast"
	ModelMain  ModelCategory = "main"
	ModelHeavy ModelCategory = "heavy"
)

// ModelRuntime identifies tool runtimes
type ModelRuntime string

const (
	ModelRuntimeClaude   ModelRuntime = "claude"
	ModelRuntimeOpenCode ModelRuntime = "opencode"
)

func ValidModelCategories() []ModelCategory {
	return []ModelCategory{ModelFast, ModelMain, ModelHeavy}
}

func IsValidModelCategory(c string) bool {
	switch ModelCategory(c) {
	case ModelFast, ModelMain, ModelHeavy:
		return true
	}
	return false
}

// LegacyToModelCategory maps old model names to categories for backward compatibility
func LegacyToModelCategory(model string) ModelCategory {
	switch model {
	case "haiku", "glm-4-flash":
		return ModelFast
	case "sonnet":
		return ModelMain
	case "opus", "kimi-k2":
		return ModelHeavy
	default:
		if IsValidModelCategory(model) {
			return ModelCategory(model)
		}
		return ModelMain
	}
}
