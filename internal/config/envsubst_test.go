package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplySubstitution(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		setupEnv      map[string]string
		expected      string
		expectedVars  []string
		expectedError error
	}{
		{
			name:     "no substitution needed",
			input:    "just a plain string",
			expected: "just a plain string",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:         "simple $VAR substitution",
			input:        "prefix $TEST_VAR suffix",
			setupEnv:     map[string]string{"TEST_VAR": "replaced"},
			expected:     "prefix replaced suffix",
			expectedVars: []string{"TEST_VAR"},
		},
		{
			name:         "braced ${VAR} substitution",
			input:        "prefix ${TEST_VAR} suffix",
			setupEnv:     map[string]string{"TEST_VAR": "replaced"},
			expected:     "prefix replaced suffix",
			expectedVars: []string{"TEST_VAR"},
		},
		{
			name:     "braced ${VAR:-default} with default",
			input:    "${TEST_VAR:-default_value}",
			expected: "default_value",
		},
		{
			name:         "braced ${VAR:-default} with env var set",
			input:        "${TEST_VAR:-default_value}",
			setupEnv:     map[string]string{"TEST_VAR": "actual_value"},
			expected:     "actual_value",
			expectedVars: []string{"TEST_VAR"},
		},
		{
			name:          "missing required ${VAR}",
			input:         "${MISSING_VAR}",
			expectedError: ErrMissingEnvVar,
		},
		{
			name:          "missing required $VAR",
			input:         "$MISSING_VAR",
			expectedError: ErrMissingEnvVar,
		},
		{
			name:         "multiple substitutions",
			input:        "${VAR1} and $VAR2 and ${VAR3:-default}",
			setupEnv:     map[string]string{"VAR1": "first", "VAR2": "second"},
			expected:     "first and second and default",
			expectedVars: []string{"VAR1", "VAR2"},
		},
		{
			name:         "multiple occurrences of same variable",
			input:        "$VAR1 and ${VAR1}",
			setupEnv:     map[string]string{"VAR1": "value"},
			expected:     "value and value",
			expectedVars: []string{"VAR1", "VAR1"},
		},
		{
			name:     "default with special characters",
			input:    "${VAR:-default with spaces}",
			expected: "default with spaces",
		},
		{
			name:         "empty default value",
			input:        "${VAR:-}",
			setupEnv:     map[string]string{"VAR": "value"},
			expected:     "value",
			expectedVars: []string{"VAR"},
		},
		{
			name:         "env var with spaces",
			input:        "${TEST_VAR}",
			setupEnv:     map[string]string{"TEST_VAR": "value with spaces"},
			expected:     "value with spaces",
			expectedVars: []string{"TEST_VAR"},
		},
		{
			name:         "env var with special chars",
			input:        "${TEST_VAR}",
			setupEnv:     map[string]string{"TEST_VAR": "value@with!special#chars"},
			expected:     "value@with!special#chars",
			expectedVars: []string{"TEST_VAR"},
		},
		{
			name:         "only substitution",
			input:        "${VAR}",
			setupEnv:     map[string]string{"VAR": "value"},
			expected:     "value",
			expectedVars: []string{"VAR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variables
			for k, v := range tt.setupEnv {
				t.Setenv(k, v)
			}

			result, usedVars, err := ApplySubstitution(tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
				assert.Equal(t, tt.expectedVars, usedVars)
			}
		})
	}
}

func TestSubstituteBraced(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		setupEnv      map[string]string
		expected      string
		expectedVars  []string
		expectedError bool
	}{
		{
			name:         "simple braced substitution",
			input:        "${VAR}",
			setupEnv:     map[string]string{"VAR": "value"},
			expected:     "value",
			expectedVars: []string{"VAR"},
		},
		{
			name:     "braced with default",
			input:    "${VAR:-default}",
			expected: "default",
		},
		{
			name:         "braced with default and env set",
			input:        "${VAR:-default}",
			setupEnv:     map[string]string{"VAR": "actual"},
			expected:     "actual",
			expectedVars: []string{"VAR"},
		},
		{
			name:          "missing required variable",
			input:         "${MISSING}",
			expectedError: true,
		},
		{
			name:         "multiple braced substitutions",
			input:        "${VAR1}${VAR2}",
			setupEnv:     map[string]string{"VAR1": "a", "VAR2": "b"},
			expected:     "ab",
			expectedVars: []string{"VAR1", "VAR2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.setupEnv {
				t.Setenv(k, v)
			}

			result, usedVars, err := substituteBraced(tt.input)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
				assert.Equal(t, tt.expectedVars, usedVars)
			}
		})
	}
}

func TestSubstituteSimple(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		setupEnv      map[string]string
		expected      string
		expectedVars  []string
		expectedError bool
	}{
		{
			name:         "simple $VAR substitution",
			input:        "$VAR",
			setupEnv:     map[string]string{"VAR": "value"},
			expected:     "value",
			expectedVars: []string{"VAR"},
		},
		{
			name:         "simple with prefix",
			input:        "prefix $VAR",
			setupEnv:     map[string]string{"VAR": "value"},
			expected:     "prefix value",
			expectedVars: []string{"VAR"},
		},
		{
			name:         "simple with suffix",
			input:        "$VAR suffix",
			setupEnv:     map[string]string{"VAR": "value"},
			expected:     "value suffix",
			expectedVars: []string{"VAR"},
		},
		{
			name:         "multiple simple substitutions",
			input:        "$VAR1 and $VAR2",
			setupEnv:     map[string]string{"VAR1": "a", "VAR2": "b"},
			expected:     "a and b",
			expectedVars: []string{"VAR1", "VAR2"},
		},
		{
			name:          "missing variable",
			input:         "$MISSING",
			expectedError: true,
		},
		{
			name:         "non-word characters before $",
			input:        "path/$VAR",
			setupEnv:     map[string]string{"VAR": "value"},
			expected:     "path/value",
			expectedVars: []string{"VAR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.setupEnv {
				t.Setenv(k, v)
			}

			result, usedVars, err := substituteSimple(tt.input)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
				assert.Equal(t, tt.expectedVars, usedVars)
			}
		})
	}
}

func TestRedactSecret(t *testing.T) {
	t.Run("no vars to redact", func(t *testing.T) {
		input := "just plain text"
		result := RedactSecret(input, []string{})
		assert.Equal(t, input, result)
	})

	t.Run("redacts secret value", func(t *testing.T) {
		t.Setenv("SECRET_KEY", "my_secret_value_123")
		input := "api_key: my_secret_value_123"
		result := RedactSecret(input, []string{"SECRET_KEY"})
		assert.Equal(t, "api_key: [REDACTED]", result)
	})

	t.Run("redacts multiple occurrences", func(t *testing.T) {
		t.Setenv("SECRET_KEY", "my_secret")
		input := "key1: my_secret, key2: my_secret"
		result := RedactSecret(input, []string{"SECRET_KEY"})
		assert.Equal(t, "key1: [REDACTED], key2: [REDACTED]", result)
	})

	t.Run("redacts multiple secrets", func(t *testing.T) {
		t.Setenv("KEY1", "secret1")
		t.Setenv("KEY2", "secret2")
		input := "key1: secret1, key2: secret2"
		result := RedactSecret(input, []string{"KEY1", "KEY2"})
		assert.Equal(t, "key1: [REDACTED], key2: [REDACTED]", result)
	})

	t.Run("empty input", func(t *testing.T) {
		result := RedactSecret("", []string{"VAR"})
		assert.Equal(t, "", result)
	})

	t.Run("var not set", func(t *testing.T) {
		input := "api_key: some_value"
		result := RedactSecret(input, []string{"UNSET_VAR"})
		assert.Equal(t, input, result)
	})
}

func TestIsSecret(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		{"api_key", "api_key", true},
		{"apiKey", "apiKey", true},
		{"API_KEY", "API_KEY", true},
		{"secret", "secret", true},
		{"password", "password", true},
		{"token", "token", true},
		{"auth", "auth", true},
		{"provider", "provider", false},
		{"model", "model", false},
		{"cost", "cost", false},
		{"endpoint", "endpoint", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSecret(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}
