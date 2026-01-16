package agent

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parser handles parsing of agent markdown files.
type Parser struct{}

// NewParser creates a new agent parser.
func NewParser() *Parser {
	return &Parser{}
}

// ParseAgentFile parses an agent markdown file and extracts metadata.
func (p *Parser) ParseAgentFile(path string) (*AgentMeta, error) {
	f, err := os.Open(path) // #nosec G304 -- controlled config/template file path
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer func() { _ = f.Close() }()

	frontmatter, content, err := p.extractFrontmatterAndContent(f)
	if err != nil {
		return nil, fmt.Errorf("extract frontmatter: %w", err)
	}

	var meta struct {
		Name        string          `yaml:"name"`
		Description string          `yaml:"description"`
		Model       string          `yaml:"model"`
		Color       string          `yaml:"color"`
		Skills      []string        `yaml:"skills"`
		Tools       map[string]bool `yaml:"tools"`
	}

	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if meta.Name == "" {
		// Use filename as name if not in frontmatter
		meta.Name = strings.TrimSuffix(filepath.Base(path), ".md")
	}

	return &AgentMeta{
		Name:        meta.Name,
		Description: meta.Description,
		Model:       meta.Model,
		Color:       meta.Color,
		Skills:      meta.Skills,
		Tools:       meta.Tools,
		Content:     strings.TrimSpace(content),
		FilePath:    path,
	}, nil
}

// extractFrontmatterAndContent extracts YAML frontmatter and markdown content.
func (p *Parser) extractFrontmatterAndContent(f *os.File) (string, string, error) {
	scanner := bufio.NewScanner(f)
	var frontmatterLines []string
	var contentLines []string
	inFrontmatter := false
	foundStart := false
	frontmatterEnded := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "---" {
			if !foundStart {
				foundStart = true
				inFrontmatter = true
				continue
			}
			if inFrontmatter {
				// End of frontmatter
				inFrontmatter = false
				frontmatterEnded = true
				continue
			}
		}

		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
		} else if frontmatterEnded {
			contentLines = append(contentLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("scan: %w", err)
	}

	if !foundStart {
		return "", "", fmt.Errorf("no frontmatter found")
	}

	frontmatter := strings.Join(frontmatterLines, "\n")
	content := strings.Join(contentLines, "\n")

	return frontmatter, content, nil
}
