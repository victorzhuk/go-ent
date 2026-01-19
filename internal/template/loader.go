package template

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Template represents a skill template with metadata.
type Template struct {
	Name        string
	Path        string
	Description string
	Category    string
	Version     string
	Author      string
}

// LoadTemplates scans the specified directory and loads all valid skill templates.
// A valid template directory must contain both config.yaml and template.md files.
// Returns a slice of Template structs or an error if the directory cannot be accessed.
func LoadTemplates(ctx context.Context, dir string) ([]*Template, error) {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("access template directory %s: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("template path is not a directory: %s", dir)
	}

	var templates []*Template
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		if path == dir {
			return nil
		}

		configPath := filepath.Join(path, "config.yaml")
		templatePath := filepath.Join(path, "template.md")

		if _, err := os.Stat(configPath); err != nil {
			return nil
		}
		if _, err := os.Stat(templatePath); err != nil {
			return nil
		}

		name := filepath.Base(path)

		cfg, err := ParseConfig(configPath)
		if err != nil {
			return fmt.Errorf("parse config for %s: %w", name, err)
		}

		tpl := &Template{
			Name:        name,
			Path:        path,
			Description: cfg.Description,
			Category:    cfg.Category,
			Version:     cfg.Version,
			Author:      cfg.Author,
		}
		templates = append(templates, tpl)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("scan templates directory: %w", err)
	}

	return templates, nil
}

// LoadTemplate loads a single template by name from the specified directory.
// Returns the Template struct or an error if the template is not found or invalid.
func LoadTemplate(ctx context.Context, dir, name string) (*Template, error) {
	templatePath := filepath.Join(dir, name)

	info, err := os.Stat(templatePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template not found: %s", name)
		}
		return nil, fmt.Errorf("access template directory %s: %w", templatePath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("template path is not a directory: %s", templatePath)
	}

	configPath := filepath.Join(templatePath, "config.yaml")
	mdPath := filepath.Join(templatePath, "template.md")

	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template %s missing required file: config.yaml", name)
		}
		return nil, fmt.Errorf("access config.yaml in template %s: %w", name, err)
	}
	if _, err := os.Stat(mdPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template %s missing required file: template.md", name)
		}
		return nil, fmt.Errorf("access template.md in template %s: %w", name, err)
	}

	cfg, err := ParseConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &Template{
		Name:        name,
		Path:        templatePath,
		Description: cfg.Description,
		Category:    cfg.Category,
		Version:     cfg.Version,
		Author:      cfg.Author,
	}, nil
}
