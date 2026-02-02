package cli

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/victorzhuk/go-ent/pkg"
)

// loadSectionTemplate loads a section template from pkg/agents/templates/sections/
func loadSectionTemplate(name string) (*template.Template, error) {
	path := fmt.Sprintf("agents/templates/sections/%s.md.tmpl", name)

	data, err := pkg.FS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read section template %s: %w", name, err)
	}

	tpl, err := template.New(name).Funcs(templateHelpers()).Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse section template %s: %w", name, err)
	}

	return tpl, nil
}

// renderSectionTemplate renders a section template with the given data
func renderSectionTemplate(tpl *template.Template, data *AgentTemplateData) (string, error) {
	var buf bytes.Buffer

	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute section template: %w", err)
	}

	return buf.String(), nil
}

// loadAllSectionTemplates loads all section templates
func loadAllSectionTemplates() (map[string]*template.Template, error) {
	sections := make(map[string]*template.Template)

	sectionNames := []string{"_tooling", "_workflow", "_principles", "_handoff"}

	for _, name := range sectionNames {
		tpl, err := loadSectionTemplate(name)
		if err != nil {
			// Section templates are optional, so just skip if not found
			continue
		}
		sections[name] = tpl
	}

	return sections, nil
}

// assembleAgentFromSections assembles an agent from section templates
// This function will be used in Phase 3 when we migrate agents to use templates
func assembleAgentFromSections(data *AgentTemplateData, sections map[string]*template.Template) (string, error) {
	var content bytes.Buffer

	// Start with agent-specific content
	content.WriteString(data.AgentContent)
	content.WriteString("\n\n")

	// Render sections in order
	sectionOrder := []string{"_tooling", "_workflow", "_principles", "_handoff"}

	for _, name := range sectionOrder {
		tpl, ok := sections[name]
		if !ok {
			continue
		}

		rendered, err := renderSectionTemplate(tpl, data)
		if err != nil {
			return "", fmt.Errorf("render section %s: %w", name, err)
		}

		if rendered != "" {
			content.WriteString(rendered)
			content.WriteString("\n\n")
		}
	}

	return content.String(), nil
}
