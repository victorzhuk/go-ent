package skill

import (
	"strings"
)

func countTokens(text string) int {
	if text == "" {
		return 0
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return 0
	}

	tokenCount := int(float64(len(words)) * 1.3)
	return tokenCount
}
