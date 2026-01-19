package skill

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectCategory(t *testing.T) {
	tests := []struct {
		name      string
		skillName string
		want      string
	}{
		{
			name:      "Go prefix",
			skillName: "go-payment",
			want:      "go",
		},
		{
			name:      "TypeScript prefix",
			skillName: "typescript-api",
			want:      "typescript",
		},
		{
			name:      "TS prefix",
			skillName: "ts-react",
			want:      "typescript",
		},
		{
			name:      "Python prefix",
			skillName: "python-data",
			want:      "python",
		},
		{
			name:      "PY prefix",
			skillName: "py-ml",
			want:      "python",
		},
		{
			name:      "Rust prefix",
			skillName: "rust-async",
			want:      "rust",
		},
		{
			name:      "Java prefix",
			skillName: "java-spring",
			want:      "java",
		},
		{
			name:      "JS prefix",
			skillName: "js-node",
			want:      "javascript",
		},
		{
			name:      "JavaScript prefix",
			skillName: "javascript-vue",
			want:      "javascript",
		},
		{
			name:      "API prefix",
			skillName: "api-design",
			want:      "api",
		},
		{
			name:      "REST prefix",
			skillName: "rest-api",
			want:      "api",
		},
		{
			name:      "GraphQL prefix",
			skillName: "graphql-api",
			want:      "api",
		},
		{
			name:      "DB prefix",
			skillName: "db-migration",
			want:      "database",
		},
		{
			name:      "SQL prefix",
			skillName: "sql-queries",
			want:      "database",
		},
		{
			name:      "Database prefix",
			skillName: "database-pg",
			want:      "database",
		},
		{
			name:      "Test prefix",
			skillName: "test-integration",
			want:      "testing",
		},
		{
			name:      "Testing prefix",
			skillName: "testing-tdd",
			want:      "testing",
		},
		{
			name:      "Spec prefix",
			skillName: "spec-api",
			want:      "testing",
		},
		{
			name:      "Sec prefix",
			skillName: "sec-auth",
			want:      "security",
		},
		{
			name:      "Security prefix",
			skillName: "security-encryption",
			want:      "security",
		},
		{
			name:      "Auth prefix",
			skillName: "auth-oauth",
			want:      "security",
		},
		{
			name:      "Review prefix",
			skillName: "review-code",
			want:      "review",
		},
		{
			name:      "Audit prefix",
			skillName: "audit-security",
			want:      "review",
		},
		{
			name:      "Arch prefix",
			skillName: "arch-microservices",
			want:      "arch",
		},
		{
			name:      "Architecture prefix",
			skillName: "architecture-ddd",
			want:      "arch",
		},
		{
			name:      "Debug prefix",
			skillName: "debug-memory",
			want:      "debugging",
		},
		{
			name:      "Debugging prefix",
			skillName: "debugging-performance",
			want:      "debugging",
		},
		{
			name:      "Core prefix",
			skillName: "core-patterns",
			want:      "core",
		},
		{
			name:      "Ops prefix",
			skillName: "ops-k8s",
			want:      "ops",
		},
		{
			name:      "DevOps prefix",
			skillName: "devops-ci",
			want:      "ops",
		},
		{
			name:      "uppercase prefix",
			skillName: "GO-PAYMENT",
			want:      "go",
		},
		{
			name:      "mixed case prefix",
			skillName: "TypeScript-API",
			want:      "typescript",
		},
		{
			name:      "whitespace trimmed",
			skillName: "  go-payment  ",
			want:      "go",
		},
		{
			name:      "no recognized prefix",
			skillName: "payment",
			want:      "",
		},
		{
			name:      "no recognized prefix with dash",
			skillName: "my-payment",
			want:      "",
		},
		{
			name:      "empty string",
			skillName: "",
			want:      "",
		},
		{
			name:      "only whitespace",
			skillName: "  ",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectCategory(tt.skillName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetermineOutputPath(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		skillName string
		want      string
	}{
		{
			name:      "Go skill",
			category:  "go",
			skillName: "go-payment",
			want:      "plugins/go-ent/skills/go/go-payment/SKILL.md",
		},
		{
			name:      "TypeScript skill",
			category:  "typescript",
			skillName: "typescript-api",
			want:      "plugins/go-ent/skills/typescript/typescript-api/SKILL.md",
		},
		{
			name:      "API skill",
			category:  "api",
			skillName: "api-design",
			want:      "plugins/go-ent/skills/api/api-design/SKILL.md",
		},
		{
			name:      "empty category defaults to core",
			category:  "",
			skillName: "general",
			want:      "plugins/go-ent/skills/core/general/SKILL.md",
		},
		{
			name:      "custom skill name",
			category:  "go",
			skillName: "my-custom-skill",
			want:      "plugins/go-ent/skills/go/my-custom-skill/SKILL.md",
		},
		{
			name:      "special characters in name",
			category:  "go",
			skillName: "go-http-client",
			want:      "plugins/go-ent/skills/go/go-http-client/SKILL.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineOutputPath(tt.category, tt.skillName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCategoryPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		category string
	}{
		{"go prefix", "go-", "go"},
		{"typescript prefix", "typescript-", "typescript"},
		{"ts prefix", "ts-", "typescript"},
		{"python prefix", "python-", "python"},
		{"py prefix", "py-", "python"},
		{"rust prefix", "rust-", "rust"},
		{"java prefix", "java-", "java"},
		{"js prefix", "js-", "javascript"},
		{"javascript prefix", "javascript-", "javascript"},
		{"api prefix", "api-", "api"},
		{"rest prefix", "rest-", "api"},
		{"graphql prefix", "graphql-", "api"},
		{"db prefix", "db-", "database"},
		{"sql prefix", "sql-", "database"},
		{"database prefix", "database-", "database"},
		{"test prefix", "test-", "testing"},
		{"testing prefix", "testing-", "testing"},
		{"spec prefix", "spec-", "testing"},
		{"sec prefix", "sec-", "security"},
		{"security prefix", "security-", "security"},
		{"auth prefix", "auth-", "security"},
		{"review prefix", "review-", "review"},
		{"audit prefix", "audit-", "review"},
		{"arch prefix", "arch-", "arch"},
		{"architecture prefix", "architecture-", "arch"},
		{"debug prefix", "debug-", "debugging"},
		{"debugging prefix", "debugging-", "debugging"},
		{"core prefix", "core-", "core"},
		{"ops prefix", "ops-", "ops"},
		{"devops prefix", "devops-", "ops"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectCategory(tt.prefix + "skill")
			assert.Equal(t, tt.category, got)
		})
	}
}

func TestGenerateSkill(t *testing.T) {
	t.Run("generates skill from template", func(t *testing.T) {
		t.Parallel()

		templateDir := t.TempDir()
		templateName := "test-template"
		templatePath := templateDir + "/" + templateName
		templateFile := templatePath + "/template.md"
		configFile := templatePath + "/config.yaml"

		if err := os.MkdirAll(templatePath, 0755); err != nil {
			t.Fatalf("create template dir: %v", err)
		}

		templateContent := `---
name: ${SKILL_NAME}
description: "${SKILL_DESCRIPTION}"
version: "1.0.0"
author: "Test Author"
tags: ["test", "template"]
triggers:
  - pattern: "test"
    weight: 0.9
---

# ${SKILL_NAME}

<role>
You are an expert test skill with extensive experience in testing patterns.
You follow best practices and write clean, maintainable code.
</role>

<instructions>
This is a test skill for ${SKILL_NAME}.
${SKILL_DESCRIPTION}

## Testing Patterns

Follow standard testing practices.
</instructions>

<constraints>
- Write clean test code
- Follow testing best practices
</constraints>

<examples>
<example>
<input>Test input</input>
<output>Test output</output>
</example>
</examples>

<output_format>
Provide clean, well-structured test code with examples.
</output_format>
`

		if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
			t.Fatalf("write template file: %v", err)
		}

		configContent := `name: Test Template
category: test
description: A test template
author: Test Author
version: 1.0.0
prompts:
  - key: test
    prompt: Test prompt
    required: true
`

		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("write config file: %v", err)
		}

		outputDir := t.TempDir()
		cfg := &WizardConfig{
			Name:         "test-skill",
			Description:  "A test skill description",
			OutputPath:   outputDir + "/SKILL.md",
			Category:     "test",
			TemplateName: templateName,
		}

		gotPath, err := GenerateSkill(context.Background(), templateDir, templateName, cfg)
		assert.NoError(t, err)
		assert.Equal(t, cfg.OutputPath, gotPath)

		gotContent, err := os.ReadFile(gotPath)
		assert.NoError(t, err)
		assert.Contains(t, string(gotContent), cfg.Name)
		assert.Contains(t, string(gotContent), cfg.Description)
		assert.Contains(t, string(gotContent), cfg.Category)
		assert.NotContains(t, string(gotContent), "${SKILL_NAME}")
		assert.NotContains(t, string(gotContent), "${SKILL_DESCRIPTION}")
	})

	t.Run("handles missing template", func(t *testing.T) {
		t.Parallel()

		templateDir := t.TempDir()
		outputDir := t.TempDir()
		cfg := &WizardConfig{
			Name:        "test-skill",
			Description: "A test skill description",
			OutputPath:  outputDir + "/SKILL.md",
			Category:    "test",
		}

		_, err := GenerateSkill(context.Background(), templateDir, "nonexistent", cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template not found")
	})

	t.Run("creates output directory structure", func(t *testing.T) {
		t.Parallel()

		templateDir := t.TempDir()
		templateName := "test-template"
		templatePath := templateDir + "/" + templateName

		if err := os.MkdirAll(templatePath, 0755); err != nil {
			t.Fatalf("create template dir: %v", err)
		}

		templateContent := `---
name: ${SKILL_NAME}
description: "${SKILL_DESCRIPTION}"
version: "1.0.0"
author: "Test Author"
tags: ["test"]
---

# ${SKILL_NAME}

<role>
You are an expert test skill with extensive experience.
You provide clear guidance and follow best practices.
</role>

<instructions>
Test instructions for ${SKILL_NAME}.
</instructions>

<constraints>
- Follow best practices
- Write clean code
</constraints>

<examples>
<example>
<input>Test input</input>
<output>Test output</output>
</example>
</examples>

<output_format>
Provide clear, well-documented code examples.
</output_format>
`

		if err := os.WriteFile(templatePath+"/template.md", []byte(templateContent), 0644); err != nil {
			t.Fatalf("write template file: %v", err)
		}

		configContent := `name: Test
category: test
description: Test
author: Test
version: 1.0.0
prompts: []
`

		if err := os.WriteFile(templatePath+"/config.yaml", []byte(configContent), 0644); err != nil {
			t.Fatalf("write config file: %v", err)
		}

		outputDir := t.TempDir()
		nestedPath := outputDir + "/nested/dir/skill/SKILL.md"

		cfg := &WizardConfig{
			Name:        "test-skill",
			Description: "Test",
			OutputPath:  nestedPath,
			Category:    "test",
		}

		_, err := GenerateSkill(context.Background(), templateDir, templateName, cfg)
		assert.NoError(t, err)

		info, err := os.Stat(nestedPath)
		assert.NoError(t, err)
		assert.True(t, info.Mode().IsRegular())
	})
}
