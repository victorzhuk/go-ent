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
	Category         string
	Version          string
	Author           string
	Tags             []string
	AllowedTools     []string
	StructureVersion string
	DependsOn        []string
	DelegatesTo      map[string]string
	Role             string
	Instructions     string
	Examples         string
	References       []string
}

// Trigger represents an explicit trigger for skill activation.
type Trigger struct {
	Patterns     []string `yaml:"patterns,omitempty"`
	Keywords     []string `yaml:"keywords,omitempty"`
	FilePatterns []string `yaml:"file_patterns,omitempty"`
	Weight       float64  `yaml:"weight,omitempty"`
}

// skillMetaV4 represents v4 frontmatter structure for unmarshaling.
// v4 uses minimal frontmatter with flat trigger array.
type skillMetaV4 struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Triggers    []string `yaml:"triggers"`
}

// Parser handles parsing of SKILL.md files.
type Parser struct{}

// NewParser creates a new skill parser.
func NewParser() *Parser {
	return &Parser{}
}

// detectVersion checks if skill is v4 format.
// v4: Flat triggers array + Markdown sections (## Role, ## Instructions, ## Examples)
func (p *Parser) detectVersion(content, frontmatter string) string {
	hasFlatTriggers := strings.Contains(frontmatter, "triggers:")
	hasV4Sections := strings.Contains(content, "## Role") &&
		strings.Contains(content, "## Instructions") &&
		strings.Contains(content, "## Examples")

	if hasFlatTriggers && hasV4Sections {
		return "v4"
	}

	return "unknown"
}

// parseFrontmatterV4 parses v4 frontmatter using yaml.Unmarshal.
// v4 uses minimal frontmatter with only name, description, and flat triggers array.
func (p *Parser) parseFrontmatterV4(frontmatter string) (*skillMetaV4, error) {
	var meta skillMetaV4
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if meta.Name == "" {
		return nil, fmt.Errorf("missing name in frontmatter")
	}

	if meta.Description == "" {
		return nil, fmt.Errorf("missing description in frontmatter")
	}

	return &meta, nil
}

// detectCategory extracts category from skill file path.
// Expected path format: .../skills/{category}/{name}/SKILL.md
func (p *Parser) detectCategory(path string) string {
	// Normalize path separators
	path = strings.ReplaceAll(path, "\\", "/")

	// Look for "skills/" in path
	skillsIdx := strings.LastIndex(path, "/skills/")
	if skillsIdx == -1 {
		return ""
	}

	// Extract path after "skills/"
	afterSkills := path[skillsIdx+len("/skills/"):]

	// Split by "/" and get first component (category)
	parts := strings.Split(afterSkills, "/")
	if len(parts) >= 1 && parts[0] != "" {
		return parts[0]
	}

	return ""
}

// ParseSkillFile parses a SKILL.md file and extracts metadata.
// Only v4 format is supported.
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

	version := p.detectVersion(string(content), frontmatter)
	if version != "v4" {
		return nil, fmt.Errorf("unsupported skill format: %s (only v4 is supported)", version)
	}

	v4Meta, err := p.parseFrontmatterV4(frontmatter)
	if err != nil {
		return nil, fmt.Errorf("parse v4: %w", err)
	}

	contentStr := string(content)

	result := &SkillMeta{
		Name:             v4Meta.Name,
		Description:      v4Meta.Description,
		Triggers:         v4Meta.Triggers,
		ExplicitTriggers: p.stringsToTriggers(v4Meta.Triggers, 0.7),
		FilePath:         path,
		Category:         p.detectCategory(path),
		StructureVersion: "v4",
		Role:             p.extractMarkdownSection(contentStr, "Role"),
		Instructions:     p.extractMarkdownSection(contentStr, "Instructions"),
		Examples:         p.extractMarkdownSection(contentStr, "Examples"),
		References:       p.extractReferencesSection(contentStr),
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

// extractMarkdownSection extracts content under a Markdown heading (e.g., "## Role").
// Returns content from heading until next heading of equal or higher level.
func (p *Parser) extractMarkdownSection(content, sectionName string) string {
	// Look for ## SectionName
	heading := "## " + sectionName
	startIdx := strings.Index(content, heading)
	if startIdx == -1 {
		return ""
	}

	// Find start of content (after heading line)
	contentStart := startIdx + len(heading)
	newlineIdx := strings.Index(content[contentStart:], "\n")
	if newlineIdx == -1 {
		// Heading is last line
		return ""
	}
	contentStart += newlineIdx + 1

	// Find end (next ## heading or end of content)
	remainingContent := content[contentStart:]
	nextHeadingIdx := strings.Index(remainingContent, "\n## ")

	var sectionContent string
	if nextHeadingIdx == -1 {
		// No next heading, take rest of content
		sectionContent = remainingContent
	} else {
		// Take until next heading
		sectionContent = remainingContent[:nextHeadingIdx]
	}

	return strings.TrimSpace(sectionContent)
}

// extractReferencesSection extracts reference file paths from ## References section.
// Returns list of paths like ["references/constraints.md", "references/edge-cases.md"].
func (p *Parser) extractReferencesSection(content string) []string {
	heading := "## References"
	startIdx := strings.Index(content, heading)
	if startIdx == -1 {
		return nil
	}

	// Find start of content (after heading line)
	contentStart := startIdx + len(heading)
	newlineIdx := strings.Index(content[contentStart:], "\n")
	if newlineIdx == -1 {
		return nil
	}
	contentStart += newlineIdx + 1

	// Find end (next ## heading or end of content)
	remainingContent := content[contentStart:]
	nextHeadingIdx := strings.Index(remainingContent, "\n## ")

	var sectionContent string
	if nextHeadingIdx == -1 {
		sectionContent = remainingContent
	} else {
		sectionContent = remainingContent[:nextHeadingIdx]
	}

	// Extract markdown link references: [text](path)
	var refs []string
	lines := strings.Split(sectionContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "- [") {
			continue
		}

		// Extract path from [text](path)
		openParen := strings.LastIndex(line, "(")
		closeParen := strings.LastIndex(line, ")")
		if openParen != -1 && closeParen != -1 && openParen < closeParen {
			path := line[openParen+1 : closeParen]
			refs = append(refs, path)
		}
	}

	return refs
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
