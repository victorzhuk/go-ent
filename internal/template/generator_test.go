package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplacePlaceholders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		data     map[string]string
		want     string
		wantErr  bool
	}{
		{
			name:     "single placeholder",
			template: "Hello ${NAME}",
			data:     map[string]string{"NAME": "World"},
			want:     "Hello World",
			wantErr:  false,
		},
		{
			name:     "multiple placeholders",
			template: "${GREETING} ${NAME}, your age is ${AGE}",
			data:     map[string]string{"GREETING": "Hello", "NAME": "Alice", "AGE": "30"},
			want:     "Hello Alice, your age is 30",
			wantErr:  false,
		},
		{
			name:     "repeated placeholder",
			template: "${NAME} says hello to ${NAME}",
			data:     map[string]string{"NAME": "Bob"},
			want:     "Bob says hello to Bob",
			wantErr:  false,
		},
		{
			name:     "no placeholders",
			template: "Just plain text",
			data:     map[string]string{"KEY": "value"},
			want:     "Just plain text",
			wantErr:  false,
		},
		{
			name:     "empty template",
			template: "",
			data:     map[string]string{"KEY": "value"},
			want:     "",
			wantErr:  false,
		},
		{
			name:     "empty data map",
			template: "${KEY}",
			data:     map[string]string{},
			want:     "${KEY}",
			wantErr:  false,
		},
		{
			name:     "nil data map",
			template: "${KEY}",
			data:     nil,
			want:     "",
			wantErr:  true,
		},
		{
			name:     "missing placeholder in data",
			template: "Hello ${MISSING}",
			data:     map[string]string{"NAME": "World"},
			want:     "Hello ${MISSING}",
			wantErr:  false,
		},
		{
			name:     "case sensitive replacement",
			template: "${Name} vs ${NAME}",
			data:     map[string]string{"NAME": "Alice"},
			want:     "${Name} vs Alice",
			wantErr:  false,
		},
		{
			name:     "placeholder with underscores",
			template: "${SKILL_NAME} is a ${SKILL_CATEGORY}",
			data:     map[string]string{"SKILL_NAME": "my-skill", "SKILL_CATEGORY": "go"},
			want:     "my-skill is a go",
			wantErr:  false,
		},
		{
			name:     "placeholder with numbers",
			template: "${VAR1} and ${VAR2}",
			data:     map[string]string{"VAR1": "first", "VAR2": "second"},
			want:     "first and second",
			wantErr:  false,
		},
		{
			name:     "placeholder at start",
			template: "${KEY} at start",
			data:     map[string]string{"KEY": "Value"},
			want:     "Value at start",
			wantErr:  false,
		},
		{
			name:     "placeholder at end",
			template: "at end ${KEY}",
			data:     map[string]string{"KEY": "Value"},
			want:     "at end Value",
			wantErr:  false,
		},
		{
			name:     "only placeholder",
			template: "${KEY}",
			data:     map[string]string{"KEY": "Value"},
			want:     "Value",
			wantErr:  false,
		},
		{
			name:     "multiple newlines",
			template: "Line1\n${KEY}\nLine2\n${KEY}",
			data:     map[string]string{"KEY": "Value"},
			want:     "Line1\nValue\nLine2\nValue",
			wantErr:  false,
		},
		{
			name:     "special characters in replacement",
			template: "Replace with ${KEY}",
			data:     map[string]string{"KEY": "special: @#$%^&*()"},
			want:     "Replace with special: @#$%^&*()",
			wantErr:  false,
		},
		{
			name:     "escaped dollar sign",
			template: "Cost: $100",
			data:     map[string]string{"KEY": "value"},
			want:     "Cost: $100",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ReplacePlaceholders(tt.template, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReplacePlaceholders_Missing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		data     map[string]string
		want     string
		wantErr  bool
	}{
		{
			name:     "multiple missing placeholders",
			template: "${A} ${B} ${C}",
			data:     map[string]string{},
			want:     "${A} ${B} ${C}",
			wantErr:  false,
		},
		{
			name:     "mixed present and missing placeholders",
			template: "${PRESENT} ${MISSING} ${PRESENT2}",
			data:     map[string]string{"PRESENT": "val1", "PRESENT2": "val3"},
			want:     "val1 ${MISSING} val3",
			wantErr:  false,
		},
		{
			name:     "consecutive missing placeholders",
			template: "${A}${B}${C}",
			data:     map[string]string{},
			want:     "${A}${B}${C}",
			wantErr:  false,
		},
		{
			name:     "same missing placeholder multiple times",
			template: "${MISSING} text ${MISSING}",
			data:     map[string]string{},
			want:     "${MISSING} text ${MISSING}",
			wantErr:  false,
		},
		{
			name:     "missing placeholder at boundaries",
			template: "${START} middle ${END}",
			data:     map[string]string{"MIDDLE": "center"},
			want:     "${START} middle ${END}",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ReplacePlaceholders(tt.template, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReplacePlaceholders_Defaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		data     map[string]string
		want     string
		wantErr  bool
	}{
		{
			name:     "SKILL_NAME defaults to my-skill",
			template: "${SKILL_NAME}",
			data:     map[string]string{},
			want:     "my-skill",
			wantErr:  false,
		},
		{
			name:     "CATEGORY defaults to general",
			template: "${CATEGORY}",
			data:     map[string]string{},
			want:     "general",
			wantErr:  false,
		},
		{
			name:     "VERSION defaults to 1.0.0",
			template: "${VERSION}",
			data:     map[string]string{},
			want:     "1.0.0",
			wantErr:  false,
		},
		{
			name:     "AUTHOR defaults to empty",
			template: "${AUTHOR}",
			data:     map[string]string{},
			want:     "",
			wantErr:  false,
		},
		{
			name:     "SKILL_DESCRIPTION defaults to empty",
			template: "${SKILL_DESCRIPTION}",
			data:     map[string]string{},
			want:     "",
			wantErr:  false,
		},
		{
			name:     "TAGS defaults to empty",
			template: "${TAGS}",
			data:     map[string]string{},
			want:     "",
			wantErr:  false,
		},
		{
			name:     "user provided SKILL_NAME overrides default",
			template: "${SKILL_NAME}",
			data:     map[string]string{"SKILL_NAME": "custom-skill"},
			want:     "custom-skill",
			wantErr:  false,
		},
		{
			name:     "user provided CATEGORY overrides default",
			template: "${CATEGORY}",
			data:     map[string]string{"CATEGORY": "go"},
			want:     "go",
			wantErr:  false,
		},
		{
			name:     "user provided VERSION overrides default",
			template: "${VERSION}",
			data:     map[string]string{"VERSION": "2.0.0"},
			want:     "2.0.0",
			wantErr:  false,
		},
		{
			name:     "mix of defaults and user provided values",
			template: "${SKILL_NAME} v${VERSION} in ${CATEGORY}",
			data:     map[string]string{"SKILL_NAME": "my-custom-skill"},
			want:     "my-custom-skill v1.0.0 in general",
			wantErr:  false,
		},
		{
			name:     "multiple defaults used",
			template: "Name: ${SKILL_NAME}, Category: ${CATEGORY}, Version: ${VERSION}",
			data:     map[string]string{},
			want:     "Name: my-skill, Category: general, Version: 1.0.0",
			wantErr:  false,
		},
		{
			name:     "repeated default placeholder",
			template: "${SKILL_NAME} and ${SKILL_NAME}",
			data:     map[string]string{},
			want:     "my-skill and my-skill",
			wantErr:  false,
		},
		{
			name:     "unknown placeholder still kept as-is",
			template: "${SKILL_NAME} ${UNKNOWN}",
			data:     map[string]string{},
			want:     "my-skill ${UNKNOWN}",
			wantErr:  false,
		},
		{
			name:     "empty AUTHOR default in template",
			template: "Author: ${AUTHOR}",
			data:     map[string]string{},
			want:     "Author: ",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ReplacePlaceholders(tt.template, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
