// Package template provides the template processing engine for project scaffolding.
package template

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// TemplateVars contains the variables available for template substitution.
type TemplateVars struct {
	ModulePath  string // Go module path (e.g., "github.com/user/project")
	ProjectName string // Project name (e.g., "my-project")
	GoVersion   string // Go version (e.g., "1.24")
}

// Engine processes template files with variable substitution.
type Engine struct {
	fs embed.FS
}

// NewEngine creates a new template engine using the provided embedded filesystem.
func NewEngine(fs embed.FS) *Engine {
	return &Engine{fs: fs}
}

// Process processes a single template file and writes the result to outputPath.
// The templatePath is relative to the embedded filesystem root.
func (e *Engine) Process(templatePath string, vars TemplateVars, outputPath string) error {
	content, err := e.fs.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer outFile.Close()

	if err := tmpl.Execute(outFile, vars); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	return nil
}

// ProcessAll processes all templates in templateDir and writes results to outputDir.
// Template files must have .tmpl extension, which is stripped in output filenames.
// Directory structure is preserved.
func (e *Engine) ProcessAll(templateDir string, vars TemplateVars, outputDir string) error {
	return fs.WalkDir(e.fs, templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .tmpl files
		if !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		// Calculate relative path from templateDir
		relPath, err := filepath.Rel(templateDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Strip .tmpl extension for output
		outputRelPath := strings.TrimSuffix(relPath, ".tmpl")
		outputPath := filepath.Join(outputDir, outputRelPath)

		if err := e.Process(path, vars, outputPath); err != nil {
			return err
		}

		return nil
	})
}

// ListTemplates returns all template files in the given directory.
func (e *Engine) ListTemplates(templateDir string) ([]string, error) {
	var templates []string

	err := fs.WalkDir(e.fs, templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".tmpl") {
			templates = append(templates, path)
		}

		return nil
	})

	return templates, err
}
