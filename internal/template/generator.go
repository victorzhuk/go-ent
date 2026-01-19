package template

import (
	"fmt"
	"regexp"
)

var placeholderRegex = regexp.MustCompile(`\$\{(\w+)\}`)

var defaultPlaceholders = map[string]string{
	"SKILL_NAME":        "my-skill",
	"SKILL_DESCRIPTION": "",
	"AUTHOR":            "",
	"CATEGORY":          "general",
	"VERSION":           "1.0.0",
	"TAGS":              "",
}

func ReplacePlaceholders(template string, data map[string]string) (string, error) {
	if data == nil {
		return "", fmt.Errorf("data cannot be nil")
	}

	result := placeholderRegex.ReplaceAllStringFunc(template, func(match string) string {
		key := match[2 : len(match)-1]
		if value, ok := data[key]; ok {
			return value
		}
		if defaultValue, ok := defaultPlaceholders[key]; ok {
			return defaultValue
		}
		return match
	})

	return result, nil
}
