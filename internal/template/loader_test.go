package template

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTemplates(t *testing.T) {
	t.Run("valid directory with templates", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		template1Dir := filepath.Join(tmpDir, "template1")
		require.NoError(t, os.MkdirAll(template1Dir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(template1Dir, "config.yaml"), []byte("name: template1\ncategory: test"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(template1Dir, "template.md"), []byte("# Template 1"), 0600))

		template2Dir := filepath.Join(tmpDir, "template2")
		require.NoError(t, os.MkdirAll(template2Dir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(template2Dir, "config.yaml"), []byte("name: template2\ncategory: test"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(template2Dir, "template.md"), []byte("# Template 2"), 0600))

		templates, err := LoadTemplates(context.Background(), tmpDir)
		require.NoError(t, err)
		assert.Len(t, templates, 2)
	})

	t.Run("missing directory returns empty list", func(t *testing.T) {
		t.Parallel()

		templates, err := LoadTemplates(context.Background(), "/nonexistent/path")
		require.NoError(t, err)
		assert.Empty(t, templates)
	})

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		templates, err := LoadTemplates(context.Background(), tmpDir)
		require.NoError(t, err)
		assert.Empty(t, templates)
	})

	t.Run("directory with incomplete templates", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		validDir := filepath.Join(tmpDir, "valid")
		require.NoError(t, os.MkdirAll(validDir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(validDir, "config.yaml"), []byte("name: valid\ncategory: test"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(validDir, "template.md"), []byte("# Valid"), 0600))

		missingConfigDir := filepath.Join(tmpDir, "missing-config")
		require.NoError(t, os.MkdirAll(missingConfigDir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(missingConfigDir, "template.md"), []byte("# Missing Config"), 0600))

		missingTemplateDir := filepath.Join(tmpDir, "missing-template")
		require.NoError(t, os.MkdirAll(missingTemplateDir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(missingTemplateDir, "config.yaml"), []byte("name: missing"), 0600))

		templates, err := LoadTemplates(context.Background(), tmpDir)
		require.NoError(t, err)
		assert.Len(t, templates, 1)
		assert.Equal(t, "valid", templates[0].Name)
	})

	t.Run("path is not a directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "file.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("content"), 0600))

		_, err := LoadTemplates(context.Background(), filePath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})

	t.Run("non-existent directory returns empty list", func(t *testing.T) {
		t.Parallel()

		templates, err := LoadTemplates(context.Background(), "/nonexistent/path")
		require.NoError(t, err)
		assert.Empty(t, templates)
	})
}

func TestLoadTemplatesFromMultipleSources(t *testing.T) {
	t.Run("built-in and custom templates", func(t *testing.T) {
		t.Parallel()

		builtInDir := t.TempDir()
		customDir := t.TempDir()

		tpl1Dir := filepath.Join(builtInDir, "builtin-1")
		require.NoError(t, os.MkdirAll(tpl1Dir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(tpl1Dir, "config.yaml"), []byte("name: builtin-1\ncategory: test\ndescription: Built-in template 1"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(tpl1Dir, "template.md"), []byte("# Built-in 1"), 0600))

		tpl2Dir := filepath.Join(builtInDir, "builtin-2")
		require.NoError(t, os.MkdirAll(tpl2Dir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(tpl2Dir, "config.yaml"), []byte("name: builtin-2\ncategory: test\ndescription: Built-in template 2"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(tpl2Dir, "template.md"), []byte("# Built-in 2"), 0600))

		tpl3Dir := filepath.Join(customDir, "custom-1")
		require.NoError(t, os.MkdirAll(tpl3Dir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(tpl3Dir, "config.yaml"), []byte("name: custom-1\ncategory: test\ndescription: Custom template 1"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(tpl3Dir, "template.md"), []byte("# Custom 1"), 0600))

		builtInTemplates, err := LoadTemplates(context.Background(), builtInDir)
		require.NoError(t, err)
		assert.Len(t, builtInTemplates, 2)

		customTemplates, err := LoadTemplates(context.Background(), customDir)
		require.NoError(t, err)
		assert.Len(t, customTemplates, 1)

		assert.Equal(t, "builtin-1", builtInTemplates[0].Name)
		assert.Equal(t, "custom-1", customTemplates[0].Name)
	})

	t.Run("only custom templates", func(t *testing.T) {
		t.Parallel()

		builtInDir := t.TempDir()
		customDir := t.TempDir()

		tplDir := filepath.Join(customDir, "custom-only")
		require.NoError(t, os.MkdirAll(tplDir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(tplDir, "config.yaml"), []byte("name: custom-only\ncategory: test\ndescription: Custom only"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(tplDir, "template.md"), []byte("# Custom"), 0600))

		builtInTemplates, err := LoadTemplates(context.Background(), builtInDir)
		require.NoError(t, err)
		assert.Empty(t, builtInTemplates)

		customTemplates, err := LoadTemplates(context.Background(), customDir)
		require.NoError(t, err)
		assert.Len(t, customTemplates, 1)
	})

	t.Run("empty custom directory", func(t *testing.T) {
		t.Parallel()

		builtInDir := t.TempDir()
		customDir := t.TempDir()

		tplDir := filepath.Join(builtInDir, "builtin")
		require.NoError(t, os.MkdirAll(tplDir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(tplDir, "config.yaml"), []byte("name: builtin\ncategory: test\ndescription: Built-in"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(tplDir, "template.md"), []byte("# Built-in"), 0600))

		builtInTemplates, err := LoadTemplates(context.Background(), builtInDir)
		require.NoError(t, err)
		assert.Len(t, builtInTemplates, 1)

		customTemplates, err := LoadTemplates(context.Background(), customDir)
		require.NoError(t, err)
		assert.Empty(t, customTemplates)
	})
}

func TestLoadTemplate(t *testing.T) {
	t.Run("valid template", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		templateDir := filepath.Join(tmpDir, "test-template")
		require.NoError(t, os.MkdirAll(templateDir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(templateDir, "config.yaml"), []byte("name: test-template\ncategory: test"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(templateDir, "template.md"), []byte("# Test Template"), 0600))

		template, err := LoadTemplate(context.Background(), tmpDir, "test-template")
		require.NoError(t, err)
		assert.Equal(t, "test-template", template.Name)
		assert.Equal(t, templateDir, template.Path)
	})

	t.Run("missing template", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		_, err := LoadTemplate(context.Background(), tmpDir, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template not found")
	})

	t.Run("template missing config.yaml", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		templateDir := filepath.Join(tmpDir, "missing-config")
		require.NoError(t, os.MkdirAll(templateDir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(templateDir, "template.md"), []byte("# Missing Config"), 0600))

		_, err := LoadTemplate(context.Background(), tmpDir, "missing-config")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required file: config.yaml")
	})

	t.Run("template missing template.md", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		templateDir := filepath.Join(tmpDir, "missing-template")
		require.NoError(t, os.MkdirAll(templateDir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(templateDir, "config.yaml"), []byte("name: missing\ncategory: test"), 0600))

		_, err := LoadTemplate(context.Background(), tmpDir, "missing-template")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required file: template.md")
	})

	t.Run("template path is not a directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "not-a-directory")
		require.NoError(t, os.WriteFile(filePath, []byte("content"), 0600))

		_, err := LoadTemplate(context.Background(), tmpDir, "not-a-directory")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})

	t.Run("template missing both required files", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		templateDir := filepath.Join(tmpDir, "empty-template")
		require.NoError(t, os.MkdirAll(templateDir, 0750))

		_, err := LoadTemplate(context.Background(), tmpDir, "empty-template")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required file: config.yaml")
	})
}
