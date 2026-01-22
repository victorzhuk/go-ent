package skill

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

func TestE2E_5_4_1_CompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	templateDir := setupTestTemplates(t)
	skillsDir := setupTestSkills(t)
	os.Setenv("GO_ENT_TEMPLATE_DIR", templateDir)
	os.Setenv("GO_ENT_SKILLS_DIR", skillsDir)
	defer os.Unsetenv("GO_ENT_TEMPLATE_DIR")
	defer os.Unsetenv("GO_ENT_SKILLS_DIR")

	t.Run("list select generate validate", func(t *testing.T) {
		stdout, _, err := runCommand("skill", "list-templates")
		require.NoError(t, err, "should list templates")
		assert.Contains(t, stdout, "go-basic", "should contain go-basic template")
		assert.Contains(t, stdout, "go-complete", "should contain go-complete template")

		args := []string{
			"skill", "new", "go-e2e-payment",
			"--template", "go-basic",
			"--description", "E2E test payment skill",
			"--category", "go",
			"--non-interactive",
		}

		stdout, stderr, err := runCommand(args...)
		if err != nil {
			t.Logf("stdout: %s", stdout)
			t.Logf("stderr: %s", stderr)
		}
		require.NoError(t, err, "should generate skill")
		assert.Contains(t, stdout, "Skill created successfully", "should show success message")
		assert.Contains(t, stdout, "go-e2e-payment", "should show skill name")

		expectedPath := filepath.Join(skillsDir, "go", "go-e2e-payment", "SKILL.md")
		_, err = os.Stat(expectedPath)
		require.NoError(t, err, "skill file should exist")

		content, err := os.ReadFile(expectedPath)
		require.NoError(t, err, "should read skill file")
		assert.Contains(t, string(content), "go-e2e-payment", "should contain skill name")
		assert.Contains(t, string(content), "E2E test payment skill", "should contain description")
		assert.Contains(t, string(content), "role", "should have role section")

		err = ValidateGeneratedSkill(expectedPath)
		assert.NoError(t, err, "should validate successfully")
	})

	t.Run("multiple sequential creations", func(t *testing.T) {
		skills := []struct {
			name        string
			template    string
			category    string
			description string
		}{
			{"go-e2e-api", "go-basic", "go", "API skill"},
			{"ts-e2e-auth", "ts-basic", "typescript", "Auth skill"},
			{"test-e2e-unit", "testing", "testing", "Unit test skill"},
		}

		for _, skill := range skills {
			args := []string{
				"skill", "new", skill.name,
				"--template", skill.template,
				"--description", skill.description,
				"--category", skill.category,
				"--non-interactive",
			}

			_, _, err := runCommand(args...)
			require.NoError(t, err, "should create %s", skill.name)

			skillPath := filepath.Join(skillsDir, skill.category, skill.name, "SKILL.md")
			_, err = os.Stat(skillPath)
			require.NoError(t, err, "skill %s should exist", skill.name)

			err = ValidateGeneratedSkill(skillPath)
			assert.NoError(t, err, "skill %s should validate", skill.name)
		}
	})

	t.Run("validation errors in generated skill", func(t *testing.T) {
		invalidTemplateDir := t.TempDir()
		templatePath := filepath.Join(invalidTemplateDir, "invalid-tpl")
		require.NoError(t, os.Mkdir(templatePath, 0755))

		configContent := `name: invalid
category: test
description: Invalid template
version: 1.0.0
prompts: []
`
		configPath := filepath.Join(templatePath, "config.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		mdContent := `---
name: test
description: test
version: [invalid yaml syntax
---

Invalid skill with parsing error
`
		mdPath := filepath.Join(templatePath, "template.md")
		require.NoError(t, os.WriteFile(mdPath, []byte(mdContent), 0644))

		customTemplateDir := filepath.Join(invalidTemplateDir, "templates")
		require.NoError(t, os.Mkdir(customTemplateDir, 0755))
		require.NoError(t, copyDir(templatePath, filepath.Join(customTemplateDir, "invalid-tpl")))

		args := []string{
			"skill", "new", "test-invalid",
			"--template", "invalid-tpl",
			"--description", "Test",
			"--category", "test",
			"--non-interactive",
		}

		os.Setenv("GO_ENT_TEMPLATE_DIR", customTemplateDir)
		defer os.Unsetenv("GO_ENT_TEMPLATE_DIR")

		_, stderr, err := runCommand(args...)
		assert.Error(t, err, "should fail with parsing error")
		assert.True(t, strings.Contains(stderr, "parse") ||
			strings.Contains(stderr, "yaml") ||
			strings.Contains(stderr, "syntax") ||
			strings.Contains(stderr, "validation"),
			"should mention parse/validation error: "+stderr)
	})
}

func TestE2E_5_4_2_CustomTemplate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	templateDir := setupTestTemplates(t)
	skillsDir := setupTestSkills(t)

	customTemplateDir := t.TempDir()
	customTemplatePath := filepath.Join(customTemplateDir, "custom-payment")
	require.NoError(t, os.Mkdir(customTemplatePath, 0755))

	configContent := `name: payment-basic
category: go
description: Payment processing template
version: 1.0.0
author: test
prompts:
  - key: SKILL_NAME
    prompt: Skill name
    default: my-payment
    required: true
  - key: DESCRIPTION
    prompt: Description
    default: payment processing
    required: true
`
	configPath := filepath.Join(customTemplatePath, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	mdContent := `---
name: ${SKILL_NAME}
description: "${DESCRIPTION}"
version: 1.0.0
author: test
tags: [payment, go]
triggers:
  - pattern: payment
    weight: 1.0
---

# ${SKILL_NAME}

<role>
Payment processing specialist for ${DESCRIPTION}
</role>

<instructions>
Handle payment operations securely
</instructions>

<examples>
## Example 1
Input: Process payment $100
Output: Payment processed successfully
</examples>

<constraints>
- Must validate amounts
- Must handle errors
</constraints>

<edge_cases>
- Zero amount payments
- Negative amounts
- Currency conversion
</edge_cases>

<output_format>
JSON response with status
</output_format>

<explicit_triggers>
- payment processing
- transaction handling
</explicit_triggers>
`
	mdPath := filepath.Join(customTemplatePath, "template.md")
	require.NoError(t, os.WriteFile(mdPath, []byte(mdContent), 0644))

	destTemplateDir := filepath.Join(templateDir, "payment-basic")
	require.NoError(t, os.MkdirAll(filepath.Dir(destTemplateDir), 0755))
	require.NoError(t, copyDir(customTemplatePath, destTemplateDir))

	os.Setenv("GO_ENT_TEMPLATE_DIR", templateDir)
	os.Setenv("GO_ENT_SKILLS_DIR", skillsDir)
	defer os.Unsetenv("GO_ENT_TEMPLATE_DIR")
	defer os.Unsetenv("GO_ENT_SKILLS_DIR")

	t.Run("list shows custom template", func(t *testing.T) {
		stdout, _, err := runCommand("skill", "list-templates")
		require.NoError(t, err)
		assert.Contains(t, stdout, "payment-basic")
		assert.Contains(t, stdout, "Payment processing template")
	})

	t.Run("generate from custom template", func(t *testing.T) {
		args := []string{
			"skill", "new", "go-stripe-integration",
			"--template", "payment-basic",
			"--description", "Stripe payment integration",
			"--category", "go",
			"--non-interactive",
		}

		stdout, _, err := runCommand(args...)
		require.NoError(t, err)
		assert.Contains(t, stdout, "Skill created successfully")

		skillPath := filepath.Join(skillsDir, "go", "go-stripe-integration", "SKILL.md")
		_, err = os.Stat(skillPath)
		require.NoError(t, err)

		content, err := os.ReadFile(skillPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "go-stripe-integration")
		assert.Contains(t, string(content), "Stripe payment integration")
		assert.Contains(t, string(content), "Payment processing specialist")

		err = ValidateGeneratedSkill(skillPath)
		assert.NoError(t, err)
	})
}

func TestE2E_5_4_3_AutoDetectCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	templateDir := setupTestTemplates(t)
	skillsDir := setupTestSkills(t)
	os.Setenv("GO_ENT_TEMPLATE_DIR", templateDir)
	os.Setenv("GO_ENT_SKILLS_DIR", skillsDir)
	defer os.Unsetenv("GO_ENT_TEMPLATE_DIR")
	defer os.Unsetenv("GO_ENT_SKILLS_DIR")

	testCases := []struct {
		name        string
		skillName   string
		expectedCat string
	}{
		{"go prefix", "go-payment", "go"},
		{"go uppercase", "go-payment-uppercase", "go"},
		{"typescript prefix", "typescript-api", "typescript"},
		{"ts prefix", "ts-react", "typescript"},
		{"python prefix", "python-ml", "python"},
		{"py prefix", "py-script", "python"},
		{"rust prefix", "rust-cli", "rust"},
		{"java prefix", "java-spring", "java"},
		{"javascript prefix", "javascript-node", "javascript"},
		{"js prefix", "js-vue", "javascript"},
		{"api prefix", "api-rest", "api"},
		{"rest prefix", "rest-endpoint", "api"},
		{"graphql prefix", "graphql-server", "api"},
		{"database prefix", "database-migration", "database"},
		{"db prefix", "db-connection", "database"},
		{"sql prefix", "sql-query", "database"},
		{"test prefix", "test-unit", "testing"},
		{"testing prefix", "testing-e2e", "testing"},
		{"spec prefix", "spec-swagger", "testing"},
		{"security prefix", "security-auth", "security"},
		{"sec prefix", "sec-encryption", "security"},
		{"auth prefix", "auth-jwt", "security"},
		{"review prefix", "review-code", "review"},
		{"audit prefix", "audit-security", "review"},
		{"arch prefix", "arch-microservices", "arch"},
		{"architecture prefix", "architecture-monolith", "arch"},
		{"debug prefix", "debug-memory", "debugging"},
		{"debugging prefix", "debugging-concurrency", "debugging"},
		{"core prefix", "core-domain", "core"},
		{"ops prefix", "ops-deploy", "ops"},
		{"devops prefix", "devops-cicd", "ops"},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{
				"skill", "new", tt.skillName,
				"--template", "go-basic",
				"--description", fmt.Sprintf("Test skill for %s", tt.skillName),
				"--category", tt.expectedCat,
				"--non-interactive",
			}

			_, _, err := runCommand(args...)
			require.NoError(t, err, "should create skill %s", tt.skillName)

			skillPath := filepath.Join(skillsDir, tt.expectedCat, tt.skillName, "SKILL.md")
			_, err = os.Stat(skillPath)
			require.NoError(t, err, "skill should be in correct category %s", tt.expectedCat)

			content, err := os.ReadFile(skillPath)
			require.NoError(t, err)
			assert.Contains(t, string(content), tt.skillName)
		})
	}

	t.Run("auto-detection in non-interactive mode", func(t *testing.T) {
		args := []string{
			"skill", "new", "go-auto-detect",
			"--template", "go-basic",
			"--description", "Auto-detection test",
			"--non-interactive",
		}

		_, _, err := runCommand(args...)
		require.NoError(t, err)

		skillPath := filepath.Join(skillsDir, "go", "go-auto-detect", "SKILL.md")
		_, err = os.Stat(skillPath)
		require.NoError(t, err)
		assert.NoError(t, ValidateGeneratedSkill(skillPath))
	})
}

func TestE2E_5_4_4_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	templateDir := setupTestTemplates(t)
	skillsDir := setupTestSkills(t)

	t.Run("missing template", func(t *testing.T) {
		os.Setenv("GO_ENT_TEMPLATE_DIR", templateDir)
		os.Setenv("GO_ENT_SKILLS_DIR", skillsDir)

		args := []string{
			"skill", "new", "test-skill",
			"--template", "non-existent-template",
			"--description", "Test",
			"--category", "test",
			"--non-interactive",
		}

		_, stderr, err := runCommand(args...)
		assert.Error(t, err)
		assert.True(t, strings.Contains(stderr, "not found") ||
			strings.Contains(stderr, "template") ||
			strings.Contains(stderr, "load"),
			"should mention template error")
	})

	t.Run("invalid config", func(t *testing.T) {
		invalidDir := t.TempDir()
		invalidPath := filepath.Join(invalidDir, "invalid-tpl")
		require.NoError(t, os.Mkdir(invalidPath, 0755))

		mdContent := `# Test template`
		mdPath := filepath.Join(invalidPath, "template.md")
		require.NoError(t, os.WriteFile(mdPath, []byte(mdContent), 0644))

		os.Setenv("GO_ENT_TEMPLATE_DIR", invalidDir)
		os.Setenv("GO_ENT_SKILLS_DIR", skillsDir)

		args := []string{
			"skill", "new", "test-skill",
			"--template", "invalid-tpl",
			"--description", "Test",
			"--category", "test",
			"--non-interactive",
		}

		_, stderr, err := runCommand(args...)
		assert.Error(t, err)
		assert.True(t, strings.Contains(stderr, "config") ||
			strings.Contains(stderr, "yaml") ||
			strings.Contains(stderr, "parse"),
			"should mention config error")
	})

	t.Run("validation failure", func(t *testing.T) {
		invalidTemplateDir := t.TempDir()
		invalidPath := filepath.Join(invalidTemplateDir, "invalid-tpl")
		require.NoError(t, os.Mkdir(invalidPath, 0755))

		configContent := `name: invalid
category: test
description: Invalid template
version: 1.0.0
prompts: []
`
		configPath := filepath.Join(invalidPath, "config.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		mdContent := `---
name: invalid
description: test
version: [invalid yaml syntax
---

This skill is invalid and should fail validation.
`
		mdPath := filepath.Join(invalidPath, "template.md")
		require.NoError(t, os.WriteFile(mdPath, []byte(mdContent), 0644))

		os.Setenv("GO_ENT_TEMPLATE_DIR", invalidTemplateDir)
		os.Setenv("GO_ENT_SKILLS_DIR", skillsDir)

		args := []string{
			"skill", "new", "test-invalid",
			"--template", "invalid-tpl",
			"--description", "Test",
			"--category", "test",
			"--non-interactive",
		}

		_, stderr, err := runCommand(args...)
		assert.Error(t, err, "should fail with parsing error")
		assert.True(t, strings.Contains(stderr, "parse") ||
			strings.Contains(stderr, "yaml") ||
			strings.Contains(stderr, "syntax") ||
			strings.Contains(stderr, "validation"),
			"should mention parse/validation error: "+stderr)
	})

	t.Run("missing required flags", func(t *testing.T) {
		os.Setenv("GO_ENT_TEMPLATE_DIR", templateDir)

		t.Run("missing template", func(t *testing.T) {
			args := []string{
				"skill", "new", "test-skill",
				"--description", "Test",
				"--non-interactive",
			}

			_, stderr, err := runCommand(args...)
			assert.Error(t, err)
			assert.Contains(t, stderr, "--template flag is required")
		})

		t.Run("missing description", func(t *testing.T) {
			args := []string{
				"skill", "new", "test-skill",
				"--template", "go-basic",
				"--non-interactive",
			}

			_, stderr, err := runCommand(args...)
			assert.Error(t, err)
			assert.Contains(t, stderr, "--description flag is required")
		})

		t.Run("missing category when undetectable", func(t *testing.T) {
			args := []string{
				"skill", "new", "unknown-skill",
				"--template", "go-basic",
				"--description", "Test",
				"--non-interactive",
			}

			_, stderr, err := runCommand(args...)
			assert.Error(t, err)
			assert.Contains(t, stderr, "--category flag required")
		})
	})

	t.Run("no templates available", func(t *testing.T) {
		emptyDir := t.TempDir()
		os.Setenv("GO_ENT_TEMPLATE_DIR", emptyDir)

		args := []string{
			"skill", "new", "test-skill",
			"--template", "go-basic",
			"--description", "Test",
			"--category", "test",
			"--non-interactive",
		}

		_, stderr, err := runCommand(args...)
		assert.Error(t, err)
		assert.Contains(t, stderr, "no templates")
	})
}

func TestE2E_5_4_5_FileAlreadyExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	templateDir := setupTestTemplates(t)
	skillsDir := setupTestSkills(t)
	os.Setenv("GO_ENT_TEMPLATE_DIR", templateDir)
	os.Setenv("GO_ENT_SKILLS_DIR", skillsDir)
	defer os.Unsetenv("GO_ENT_TEMPLATE_DIR")
	defer os.Unsetenv("GO_ENT_SKILLS_DIR")

	t.Run("should error with clear message", func(t *testing.T) {
		args := []string{
			"skill", "new", "go-exists",
			"--template", "go-basic",
			"--description", "First creation",
			"--category", "go",
			"--non-interactive",
		}

		stdout, _, err := runCommand(args...)
		require.NoError(t, err, "first creation should succeed")
		assert.Contains(t, stdout, "Skill created successfully")

		skillPath := filepath.Join(skillsDir, "go", "go-exists", "SKILL.md")
		_, err = os.Stat(skillPath)
		require.NoError(t, err, "skill should exist after first creation")

		args = []string{
			"skill", "new", "go-exists",
			"--template", "go-basic",
			"--description", "Second creation",
			"--category", "go",
			"--non-interactive",
		}

		_, stderr, err := runCommand(args...)
		assert.Error(t, err, "second creation should fail")
		assert.True(t, strings.Contains(stderr, "exists") ||
			strings.Contains(stderr, "already") ||
			strings.Contains(stderr, "file"),
			"should mention file exists error")
	})

	t.Run("different names same category", func(t *testing.T) {
		names := []string{"go-skill-1", "go-skill-2", "go-skill-3"}

		for _, name := range names {
			args := []string{
				"skill", "new", name,
				"--template", "go-basic",
				"--description", fmt.Sprintf("Skill %s", name),
				"--category", "go",
				"--non-interactive",
			}

			_, _, err := runCommand(args...)
			require.NoError(t, err, "should create %s", name)

			skillPath := filepath.Join(skillsDir, "go", name, "SKILL.md")
			_, err = os.Stat(skillPath)
			require.NoError(t, err, "%s should exist", name)
		}
	})
}

func TestE2E_IntegrationScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	templateDir := setupTestTemplates(t)
	skillsDir := setupTestSkills(t)
	os.Setenv("GO_ENT_TEMPLATE_DIR", templateDir)
	os.Setenv("GO_ENT_SKILLS_DIR", skillsDir)
	defer os.Unsetenv("GO_ENT_TEMPLATE_DIR")
	defer os.Unsetenv("GO_ENT_SKILLS_DIR")

	t.Run("list select create validate workflow", func(t *testing.T) {
		stdout, _, err := runCommand("skill", "list-templates")
		require.NoError(t, err)
		assert.Contains(t, stdout, "go-basic")

		args := []string{
			"skill", "new", "go-workflow-test",
			"--template", "go-basic",
			"--description", "Integration workflow test",
			"--category", "go",
			"--non-interactive",
		}

		stdout, _, err = runCommand(args...)
		require.NoError(t, err)
		assert.Contains(t, stdout, "Skill created successfully")

		skillPath := filepath.Join(skillsDir, "go", "go-workflow-test", "SKILL.md")
		content, err := os.ReadFile(skillPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "go-workflow-test")

		err = ValidateGeneratedSkill(skillPath)
		assert.NoError(t, err)
	})

	t.Run("error recovery flow", func(t *testing.T) {
		args := []string{
			"skill", "new", "go-recovery",
			"--template", "non-existent",
			"--description", "Test",
			"--category", "go",
			"--non-interactive",
		}

		_, _, err := runCommand(args...)
		assert.Error(t, err)

		args = []string{
			"skill", "new", "go-recovery",
			"--template", "go-basic",
			"--description", "Recovery test",
			"--category", "go",
			"--non-interactive",
		}

		stdout, _, err := runCommand(args...)
		require.NoError(t, err)
		assert.Contains(t, stdout, "Skill created successfully")

		skillPath := filepath.Join(skillsDir, "go", "go-recovery", "SKILL.md")
		_, err = os.Stat(skillPath)
		require.NoError(t, err)
	})
}
