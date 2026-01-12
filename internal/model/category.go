package model

type Category string

const (
	Fast  Category = "fast"
	Main  Category = "main"
	Heavy Category = "heavy"
)

// Runtime identifies tool runtimes
type Runtime string

const (
	RuntimeClaude   Runtime = "claude"
	RuntimeOpenCode Runtime = "opencode"
)

func ValidCategories() []Category {
	return []Category{Fast, Main, Heavy}
}

func IsValid(c string) bool {
	switch Category(c) {
	case Fast, Main, Heavy:
		return true
	}
	return false
}

// LegacyToCategory maps old model names to categories for backward compatibility
func LegacyToCategory(model string) Category {
	switch model {
	case "haiku", "glm-4-flash":
		return Fast
	case "sonnet":
		return Main
	case "opus", "kimi-k2":
		return Heavy
	default:
		if IsValid(model) {
			return Category(model)
		}
		return Main
	}
}
