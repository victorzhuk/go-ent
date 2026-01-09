package skill

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillMeta represents parsed skill metadata from SKILL.md files.
type SkillMeta struct {
	Name        string
	Description string
	Triggers    []string
	FilePath    string
}

// Parser handles parsing of SKILL.md files.
type Parser struct{}

// NewParser creates a new skill parser.
func NewParser() *Parser {
	return &Parser{}
}

// ParseSkillFile parses a SKILL.md file and extracts metadata.
func (p *Parser) ParseSkillFile(path string) (*SkillMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	frontmatter, err := p.extractFrontmatter(f)
	if err != nil {
		return nil, fmt.Errorf("extract frontmatter: %w", err)
	}

	var meta struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}

	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if meta.Name == "" {
		return nil, fmt.Errorf("missing name in frontmatter")
	}

	triggers := p.extractTriggers(meta.Description)

	return &SkillMeta{
		Name:        meta.Name,
		Description: meta.Description,
		Triggers:    triggers,
		FilePath:    path,
	}, nil
}

// extractFrontmatter extracts YAML frontmatter between --- delimiters.
func (p *Parser) extractFrontmatter(f *os.File) (string, error) {
	scanner := bufio.NewScanner(f)
	var lines []string
	inFrontmatter := false
	foundStart := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "---" {
			if !foundStart {
				foundStart = true
				inFrontmatter = true
				continue
			}
			// End of frontmatter
			break
		}

		if inFrontmatter {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan: %w", err)
	}

	if !foundStart {
		return "", fmt.Errorf("no frontmatter found")
	}

	return strings.Join(lines, "\n"), nil
}

// extractTriggers extracts keywords from "Auto-activates for:" in description.
func (p *Parser) extractTriggers(description string) []string {
	const prefix = "Auto-activates for:"
	idx := strings.Index(description, prefix)
	if idx == -1 {
		return nil
	}

	// Extract text after "Auto-activates for:"
	rest := description[idx+len(prefix):]

	// Find the end (period or end of string)
	endIdx := strings.Index(rest, ".")
	if endIdx == -1 {
		endIdx = len(rest)
	}
	triggerText := rest[:endIdx]

	// Split by commas
	parts := strings.Split(triggerText, ",")
	triggers := make([]string, 0, len(parts))
	for _, part := range parts {
		trigger := strings.TrimSpace(part)
		if trigger != "" {
			triggers = append(triggers, strings.ToLower(trigger))
		}
	}

	return triggers
}
