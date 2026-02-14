package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GenerateContextPrompt(ws *Workspace) (string, error) {
	if ws == nil {
		return "", nil
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf("## Workspace Context: %s\n\n", ws.Name))
	b.WriteString(fmt.Sprintf("This project is part of the **%s** workspace", ws.Name))

	if len(ws.Projects) > 0 {
		b.WriteString(" with these projects:\n")
		for _, p := range ws.Projects {
			desc := p.Description
			if desc == "" {
				desc = p.Name
			}
			b.WriteString(fmt.Sprintf("- %s (%s)\n", p.Name, desc))
		}
	} else {
		b.WriteString(".\n")
	}

	specs, err := listSpecSummaries(ws)
	if err == nil && len(specs) > 0 {
		b.WriteString("\n### Shared Architecture Specs\n")
		for _, s := range specs {
			b.WriteString(fmt.Sprintf("- **%s**: %s\n", s.id, s.title))
		}
	}

	b.WriteString(fmt.Sprintf("\nWorkspace specs are at: %s\n", filepath.Join(ws.Path, "openspec", "specs")))

	return b.String(), nil
}

type specSummary struct {
	id    string
	title string
}

func listSpecSummaries(ws *Workspace) ([]specSummary, error) {
	specsDir := filepath.Join(ws.Path, "openspec", "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read specs dir: %w", err)
	}

	var result []specSummary
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		title := ExtractTitle(filepath.Join(specsDir, e.Name(), "spec.md"))
		if title == "" {
			title = e.Name()
		}

		result = append(result, specSummary{
			id:    e.Name(),
			title: title,
		})
	}

	return result, nil
}

func WriteContextPrompt(outputDir string, ws *Workspace) error {
	content, err := GenerateContextPrompt(ws)
	if err != nil {
		return fmt.Errorf("generate context: %w", err)
	}

	if content == "" {
		return nil
	}

	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	path := filepath.Join(outputDir, "_workspace.md")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write _workspace.md: %w", err)
	}

	return nil
}
