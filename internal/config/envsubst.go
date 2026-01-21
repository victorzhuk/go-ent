package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	// ErrMissingEnvVar indicates a required environment variable is not set.
	ErrMissingEnvVar = errors.New("missing environment variable")
)

var (
	// Pattern for ${VAR_NAME}
	patternBraced = regexp.MustCompile(`\$\{([^}:]+)(?::-([^}]*))?\}`)
	// Pattern for $VAR_NAME (only at start or after non-word chars)
	patternSimple = regexp.MustCompile(`(^|[^$\w])\$([A-Za-z_][A-Za-z0-9_]*)`)
)

// ApplySubstitution replaces environment variable references in the input string.
// Supports:
// - ${VAR_NAME} - required variable
// - ${VAR:-default} - variable with default value
// - $VAR_NAME - required variable (simple syntax)
func ApplySubstitution(input string) (string, []string, error) {
	if input == "" {
		return input, nil, nil
	}

	var usedVars []string
	var err error

	// First handle ${VAR:-default} and ${VAR} patterns
	result, bracedVars, bracedErr := substituteBraced(input)
	usedVars = append(usedVars, bracedVars...)
	if bracedErr != nil {
		err = bracedErr
	}

	// Then handle simple $VAR_NAME patterns
	result, simpleVars, simpleErr := substituteSimple(result)
	usedVars = append(usedVars, simpleVars...)
	if simpleErr != nil {
		err = simpleErr
	}

	return result, usedVars, err
}

// substituteBraced handles ${VAR:-default} and ${VAR} patterns
func substituteBraced(input string) (string, []string, error) {
	var usedVars []string
	var err error

	result := patternBraced.ReplaceAllStringFunc(input, func(match string) string {
		submatches := patternBraced.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		varName := submatches[1]
		defaultValue := ""
		if len(submatches) > 2 {
			defaultValue = submatches[2]
		}

		value, exists := os.LookupEnv(varName)
		if !exists {
			if defaultValue != "" {
				return defaultValue
			}
			err = fmt.Errorf("%w: %s", ErrMissingEnvVar, varName)
			return match
		}

		usedVars = append(usedVars, varName)
		return value
	})

	return result, usedVars, err
}

// substituteSimple handles $VAR_NAME patterns
func substituteSimple(input string) (string, []string, error) {
	var usedVars []string
	var err error

	result := patternSimple.ReplaceAllStringFunc(input, func(match string) string {
		submatches := patternSimple.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		prefix := submatches[1]
		varName := submatches[2]

		value, exists := os.LookupEnv(varName)
		if !exists {
			err = fmt.Errorf("%w: %s", ErrMissingEnvVar, varName)
			return match
		}

		usedVars = append(usedVars, varName)
		return prefix + value
	})

	return result, usedVars, err
}

// RedactSecret replaces secret values in the input string with [REDACTED].
// A secret is identified by being referenced as an environment variable.
func RedactSecret(input string, usedVars []string) string {
	if input == "" || len(usedVars) == 0 {
		return input
	}

	result := input
	for _, varName := range usedVars {
		if value, exists := os.LookupEnv(varName); exists && value != "" {
			result = strings.ReplaceAll(result, value, "[REDACTED]")
		}
	}
	return result
}

// IsSecret determines if a field name likely contains a secret value.
func IsSecret(fieldName string) bool {
	lower := strings.ToLower(fieldName)
	return strings.Contains(lower, "key") ||
		strings.Contains(lower, "secret") ||
		strings.Contains(lower, "password") ||
		strings.Contains(lower, "token") ||
		strings.Contains(lower, "auth")
}
