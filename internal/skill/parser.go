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
	ExplicitTriggers []Trigger
	FilePath         string
	Version          string
	Author           string
	Tags             []string
	AllowedTools     []string
	StructureVersion string
	DependsOn        []string
	DelegatesTo      map[string]string
	QualityScore     *QualityScore
}

// Trigger represents an explicit trigger for skill activation.
type Trigger struct {
	Patterns     []string `yaml:"patterns,omitempty"`
	Keywords     []string `yaml:"keywords,omitempty"`
	FilePatterns []string `yaml:"file_patterns,omitempty"`
	Weight       float64  `yaml:"weight,omitempty"`
}

// skillMetaV2 represents v2 frontmatter structure for unmarshaling.
type skillMetaV2 struct {
	Name         string            `yaml:"name"`
	Description  string            `yaml:"description"`
	Version      string            `yaml:"version"`
	Author       string            `yaml:"author"`
	Tags         []string          `yaml:"tags"`
	AllowedTools []string          `yaml:"allowedTools"`
	Triggers     []Trigger         `yaml:"triggers"`
	DependsOn    []string          `yaml:"depends_on"`
	DelegatesTo  map[string]string `yaml:"delegates_to"`
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

	for i := range meta.Triggers {
		if meta.Triggers[i].Weight == 0 {
			meta.Triggers[i].Weight = 0.7
		}
		if meta.Triggers[i].Weight < 0.0 || meta.Triggers[i].Weight > 1.0 {
			return nil, fmt.Errorf("trigger weight must be between 0.0 and 1.0, got %f", meta.Triggers[i].Weight)
		}
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

		var explicitTriggers []Trigger
		var triggers []string

		if len(v2Meta.Triggers) > 0 {
			// Use explicit triggers from frontmatter
			explicitTriggers = v2Meta.Triggers
			triggers = p.triggersToStrings(explicitTriggers)
		} else {
			// Fallback to description-based extraction with weight 0.5
			descriptionTriggers := p.extractTriggers(v2Meta.Description)
			triggers = descriptionTriggers
			explicitTriggers = p.stringsToTriggers(descriptionTriggers, 0.5)
		}

		result = &SkillMeta{
			Name:             v2Meta.Name,
			Description:      v2Meta.Description,
			Version:          v2Meta.Version,
			Author:           v2Meta.Author,
			Tags:             v2Meta.Tags,
			AllowedTools:     v2Meta.AllowedTools,
			Triggers:         triggers,
			ExplicitTriggers: explicitTriggers,
			FilePath:         path,
			StructureVersion: "v2",
			DependsOn:        v2Meta.DependsOn,
			DelegatesTo:      v2Meta.DelegatesTo,
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
			DependsOn:        nil,
			DelegatesTo:      nil,
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

// triggersToStrings converts explicit triggers to string format for backward compatibility.
func (p *Parser) triggersToStrings(explicit []Trigger) []string {
	result := make([]string, 0, len(explicit)*3)

	for _, t := range explicit {
		for _, pat := range t.Patterns {
			if pat != "" {
				result = append(result, strings.ToLower(pat))
			}
		}
		for _, kw := range t.Keywords {
			if kw != "" {
				result = append(result, strings.ToLower(kw))
			}
		}
		for _, fp := range t.FilePatterns {
			if fp != "" {
				result = append(result, strings.ToLower(fp))
			}
		}
	}

	return result
}

// stringToTrigger converts a description-based trigger to an explicit Trigger with fallback weight.
func (p *Parser) stringToTrigger(keyword string, weight float64) Trigger {
	return Trigger{
		Keywords: []string{keyword},
		Weight:   weight,
	}
}

// stringsToTriggers converts description-based triggers to explicit triggers with fallback weight.
func (p *Parser) stringsToTriggers(strings []string, weight float64) []Trigger {
	triggers := make([]Trigger, 0, len(strings))
	for _, s := range strings {
		triggers = append(triggers, p.stringToTrigger(s, weight))
	}
	return triggers
}
