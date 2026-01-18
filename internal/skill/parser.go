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
	Name             string
	Description      string
	Triggers         []string
	FilePath         string
	Version          string
	Author           string
	Tags             []string
	AllowedTools     []string
	StructureVersion string
	QualityScore     float64
}

// skillMetaV2 represents v2 frontmatter structure for unmarshaling.
type skillMetaV2 struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Version      string   `yaml:"version"`
	Author       string   `yaml:"author"`
	Tags         []string `yaml:"tags"`
	AllowedTools []string `yaml:"allowedTools"`
}

// Parser handles parsing of SKILL.md files.
type Parser struct{}

// NewParser creates a new skill parser.
func NewParser() *Parser {
	return &Parser{}
}

// detectVersion checks if content contains v2 XML tags.
func (p *Parser) detectVersion(content string) string {
	if strings.Contains(content, "<role>") || strings.Contains(content, "<instructions>") {
		return "v2"
	}
	return "v1"
}

// parseFrontmatterV2 parses v2 frontmatter using yaml.Unmarshal.
func (p *Parser) parseFrontmatterV2(frontmatter string) (*skillMetaV2, error) {
	var meta skillMetaV2
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if meta.Name == "" {
		return nil, fmt.Errorf("missing name in frontmatter")
	}

	return &meta, nil
}

// ParseSkillFile parses a SKILL.md file and extracts metadata.
func (p *Parser) ParseSkillFile(path string) (*SkillMeta, error) {
	f, err := os.Open(path) // #nosec G304 -- controlled config/template file path
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer func() { _ = f.Close() }()

	frontmatter, err := p.extractFrontmatter(f)
	if err != nil {
		return nil, fmt.Errorf("extract frontmatter: %w", err)
	}

	content, err := os.ReadFile(path) // #nosec G304 -- controlled config/template file path
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	version := p.detectVersion(string(content))

	var result *SkillMeta

	if version == "v2" {
		v2Meta, err := p.parseFrontmatterV2(frontmatter)
		if err != nil {
			return nil, fmt.Errorf("parse v2: %w", err)
		}

		triggers := p.extractTriggers(v2Meta.Description)

		result = &SkillMeta{
			Name:             v2Meta.Name,
			Description:      v2Meta.Description,
			Version:          v2Meta.Version,
			Author:           v2Meta.Author,
			Tags:             v2Meta.Tags,
			AllowedTools:     v2Meta.AllowedTools,
			Triggers:         triggers,
			FilePath:         path,
			StructureVersion: "v2",
		}
	} else {
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

		result = &SkillMeta{
			Name:             meta.Name,
			Description:      meta.Description,
			Version:          "",
			Author:           "",
			Tags:             nil,
			AllowedTools:     nil,
			Triggers:         triggers,
			FilePath:         path,
			StructureVersion: "v1",
		}
	}

	return result, nil
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
