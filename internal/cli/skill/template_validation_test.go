package skill

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skillpkg "github.com/victorzhuk/go-ent/internal/skill"
	"github.com/victorzhuk/go-ent/internal/template"
)

func TestGenerateAndValidateAllTemplates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	templateDir := filepath.Join("..", "..", "..", "plugins", "go-ent", "templates", "skills")

	templates, err := template.LoadTemplates(ctx, templateDir)
	require.NoError(t, err, "load templates")
	require.Len(t, templates, 11, "expect 11 built-in templates")

	templateNames := make([]string, len(templates))
	for i, tpl := range templates {
		templateNames[i] = tpl.Name
	}

	for _, templateName := range templateNames {
		t.Run(templateName, func(t *testing.T) {
			t.Parallel()

			outputPath := filepath.Join(os.TempDir(), "go-ent-test-skills", templateName, "test-"+templateName, "SKILL.md")

			cfg := &WizardConfig{
				Name:         "test-" + templateName,
				TemplateName: templateName,
				Description:  "Test skill generated from " + templateName + " template",
				Category:     templateName,
				OutputPath:   outputPath,
			}

			os.RemoveAll(filepath.Dir(outputPath))

			_, err = GenerateSkill(ctx, templateDir, cfg.TemplateName, cfg)
			require.NoError(t, err, "generate skill from template")
			defer os.RemoveAll(filepath.Dir(outputPath))

			assert.FileExists(t, outputPath, "skill file should exist")

			content, err := os.ReadFile(outputPath)
			require.NoError(t, err, "read skill file")

			parser := skillpkg.NewParser()
			meta, err := parser.ParseSkillFile(outputPath)
			require.NoError(t, err, "parse skill file")

			scorer := skillpkg.NewQualityScorer()
			qualityScore := scorer.Score(meta, string(content))

			validator := skillpkg.NewValidator()
			result := validator.Validate(meta, string(content))

			assert.True(t, result.Valid, "skill should pass validation, errors: %d", result.ErrorCount())
			assert.GreaterOrEqual(t, qualityScore.Total, 90.0, "quality score should be >= 90, got %.2f", qualityScore.Total)
		})
	}
}

func TestPlaceholderReplacement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		template    string
		data        map[string]string
		want        string
		wantErr     bool
		description string
	}{
		{
			name:        "basic replacement",
			template:    "Hello ${NAME}",
			data:        map[string]string{"NAME": "World"},
			want:        "Hello World",
			wantErr:     false,
			description: "5.3.4 - Test basic placeholder replacement",
		},
		{
			name:        "multiple placeholders",
			template:    "${GREETING} ${NAME} ${TIME}",
			data:        map[string]string{"GREETING": "Hello", "NAME": "World", "TIME": "today"},
			want:        "Hello World today",
			wantErr:     false,
			description: "5.3.4 - Test multiple placeholders",
		},
		{
			name:        "empty value",
			template:    "Value: ${EMPTY}",
			data:        map[string]string{"EMPTY": ""},
			want:        "Value: ",
			wantErr:     false,
			description: "5.3.4 - Test empty placeholder value",
		},
		{
			name:        "special characters",
			template:    "Path: ${PATH}",
			data:        map[string]string{"PATH": "/usr/local/bin"},
			want:        "Path: /usr/local/bin",
			wantErr:     false,
			description: "5.3.4 - Test special characters in placeholder",
		},
		{
			name:        "multiline value",
			template:    "Code:\n${CODE}",
			data:        map[string]string{"CODE": "func main() {\n\tprintln(\"hello\")\n}"},
			want:        "Code:\nfunc main() {\n\tprintln(\"hello\")\n}",
			wantErr:     false,
			description: "5.3.4 - Test multiline placeholder value",
		},
		{
			name:        "nested braces",
			template:    "Value: ${BRACES}",
			data:        map[string]string{"BRACES": "foo{bar}"},
			want:        "Value: foo{bar}",
			wantErr:     false,
			description: "5.3.4 - Test nested braces in value",
		},
		{
			name:        "dollar sign in value",
			template:    "Price: ${PRICE}",
			data:        map[string]string{"PRICE": "$100"},
			want:        "Price: $100",
			wantErr:     false,
			description: "5.3.4 - Test dollar sign in value",
		},
		{
			name:        "repeated placeholder",
			template:    "${NAME} says ${NAME}",
			data:        map[string]string{"NAME": "Hello"},
			want:        "Hello says Hello",
			wantErr:     false,
			description: "5.3.4 - Test repeated placeholder",
		},
		{
			name:        "default placeholder",
			template:    "Name: ${SKILL_NAME}",
			data:        map[string]string{},
			want:        "Name: my-skill",
			wantErr:     false,
			description: "5.3.4 - Test default placeholder value",
		},
		{
			name:        "nil data",
			template:    "Hello ${NAME}",
			data:        nil,
			want:        "",
			wantErr:     true,
			description: "5.3.5 - Test with nil data (should error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := template.ReplacePlaceholders(tt.template, tt.data)
			if tt.wantErr {
				assert.Error(t, err, tt.description)
			} else {
				require.NoError(t, err, tt.description)
				assert.Equal(t, tt.want, got, tt.description)
			}
		})
	}
}

func TestMissingPlaceholders(t *testing.T) {
	t.Parallel()

	t.Run("extra placeholders kept as-is", func(t *testing.T) {
		t.Parallel()

		templateContent := "Hello ${NAME}, welcome to ${APP}"
		data := map[string]string{"NAME": "World"}

		result, err := template.ReplacePlaceholders(templateContent, data)
		require.NoError(t, err, "5.3.6 - extra placeholders should not error")

		assert.Contains(t, result, "Hello World", "known placeholder should be replaced")
		assert.Contains(t, result, "${APP}", "unknown placeholder should be kept as-is")
	})

	t.Run("no placeholders in data - kept as-is", func(t *testing.T) {
		t.Parallel()

		templateContent := "Hello ${NAME}"
		data := map[string]string{}

		result, err := template.ReplacePlaceholders(templateContent, data)
		require.NoError(t, err, "5.3.6 - missing placeholders should not error")

		assert.Contains(t, result, "${NAME}", "missing placeholder should be kept as-is")
	})
}

func TestQualityScoreFromGeneratedSkills(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	templateDir := filepath.Join("..", "..", "..", "plugins", "go-ent", "templates", "skills")

	templates := []string{
		"go-basic",
		"go-complete",
		"typescript-basic",
		"database",
		"testing",
		"api-design",
		"core-basic",
		"debugging-basic",
		"security",
		"review",
		"arch",
	}

	for _, templateName := range templates {
		t.Run(templateName+" quality score", func(t *testing.T) {
			t.Parallel()

			outputPath := filepath.Join(os.TempDir(), "go-ent-quality-test", templateName, "SKILL.md")

			cfg := &WizardConfig{
				Name:         "test-quality-" + templateName,
				TemplateName: templateName,
				Description:  "Quality score test for " + templateName + " template",
				Category:     templateName,
				OutputPath:   outputPath,
			}

			os.RemoveAll(filepath.Dir(outputPath))

			_, err := GenerateSkill(ctx, templateDir, cfg.TemplateName, cfg)
			require.NoError(t, err, "generate skill")
			defer os.RemoveAll(filepath.Dir(outputPath))

			content, err := os.ReadFile(outputPath)
			require.NoError(t, err, "read skill file")

			parser := skillpkg.NewParser()
			meta, err := parser.ParseSkillFile(outputPath)
			require.NoError(t, err, "parse skill file")

			scorer := skillpkg.NewQualityScorer()
			qualityScore := scorer.Score(meta, string(content))

			assert.GreaterOrEqual(t, qualityScore.Total, 90.0,
				"5.3.3 - %s template quality score should be >= 90, got %.2f",
				templateName, qualityScore.Total)
		})
	}
}

func TestStrictValidationForAllTemplates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	templateDir := filepath.Join("..", "..", "..", "plugins", "go-ent", "templates", "skills")

	templates, err := template.LoadTemplates(ctx, templateDir)
	require.NoError(t, err, "load templates")

	for _, tpl := range templates {
		t.Run(tpl.Name+" strict validation", func(t *testing.T) {
			t.Parallel()

			outputPath := filepath.Join(os.TempDir(), "go-ent-strict-test", tpl.Name, "SKILL.md")

			cfg := &WizardConfig{
				Name:         "test-strict-" + tpl.Name,
				TemplateName: tpl.Name,
				Description:  "Strict validation test for " + tpl.Name,
				Category:     tpl.Name,
				OutputPath:   outputPath,
			}

			os.RemoveAll(filepath.Dir(outputPath))

			_, err = GenerateSkill(ctx, templateDir, cfg.TemplateName, cfg)
			require.NoError(t, err, "generate skill")
			defer os.RemoveAll(filepath.Dir(outputPath))

			content, err := os.ReadFile(outputPath)
			require.NoError(t, err, "read skill file")

			parser := skillpkg.NewParser()
			meta, err := parser.ParseSkillFile(outputPath)
			require.NoError(t, err, "parse skill file")

			validator := skillpkg.NewValidator()
			result := validator.Validate(meta, string(content))

			assert.True(t, result.Valid,
				"5.3.2 - %s template should pass validation, errors: %d",
				tpl.Name, result.ErrorCount())
			assert.Equal(t, 0, result.ErrorCount(),
				"5.3.2 - %s template should have no validation errors",
				tpl.Name)
		})
	}
}

func TestEdgeCasePlaceholders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		template    string
		data        map[string]string
		want        string
		description string
	}{
		{
			name:        "unicode characters",
			template:    "Message: ${MSG}",
			data:        map[string]string{"MSG": "Hello 世界 🌍"},
			want:        "Message: Hello 世界 🌍",
			description: "5.3.4 - Test unicode characters",
		},
		{
			name:        "very long value",
			template:    "Content: ${CONTENT}",
			data:        map[string]string{"CONTENT": string(make([]byte, 10000))},
			want:        "Content: " + string(make([]byte, 10000)),
			description: "5.3.4 - Test very long value",
		},
		{
			name:        "quoted string in value",
			template:    "Text: ${TEXT}",
			data:        map[string]string{"TEXT": `He said "hello"`},
			want:        `Text: He said "hello"`,
			description: "5.3.4 - Test quoted string in value",
		},
		{
			name:        "backslashes in value",
			template:    "Path: ${PATH}",
			data:        map[string]string{"PATH": `C:\Users\Name`},
			want:        "Path: C:\\Users\\Name",
			description: "5.3.4 - Test backslashes in value",
		},
		{
			name:        "placeholder at start",
			template:    "${PREFIX} suffix",
			data:        map[string]string{"PREFIX": "prefix"},
			want:        "prefix suffix",
			description: "5.3.4 - Test placeholder at start of string",
		},
		{
			name:        "placeholder at end",
			template:    "prefix ${SUFFIX}",
			data:        map[string]string{"SUFFIX": "suffix"},
			want:        "prefix suffix",
			description: "5.3.4 - Test placeholder at end of string",
		},
		{
			name:        "only placeholder",
			template:    "${ONLY}",
			data:        map[string]string{"ONLY": "value"},
			want:        "value",
			description: "5.3.4 - Test string with only placeholder",
		},
		{
			name:        "consecutive placeholders",
			template:    "${A}${B}${C}",
			data:        map[string]string{"A": "1", "B": "2", "C": "3"},
			want:        "123",
			description: "5.3.4 - Test consecutive placeholders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := template.ReplacePlaceholders(tt.template, tt.data)
			require.NoError(t, err, tt.description)
			assert.Equal(t, tt.want, got, tt.description)
		})
	}
}
