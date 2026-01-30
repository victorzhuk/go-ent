package skill

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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

// skillMetaV3 represents v3 frontmatter structure for unmarshaling.
type skillMetaV3 struct {
	Name         string            `yaml:"name"`
	Description  string            `yaml:"description"`
	Version      string            `yaml:"version"`
	Author       string            `yaml:"author"`
	Tags         []string          `yaml:"tags"`
	AllowedTools []string          `yaml:"allowed-tools"`
	Triggers     *v3Triggers       `yaml:"triggers"`
	DependsOn    []string          `yaml:"depends_on"`
	DelegatesTo  map[string]string `yaml:"delegates_to"`
}

// skillMetaV4 represents v4 frontmatter structure for unmarshaling.
// v4 uses minimal frontmatter with flat trigger array.
type skillMetaV4 struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Triggers    []string `yaml:"triggers"`
}

// v3Triggers represents v3 trigger structure in frontmatter.
type v3Triggers struct {
	Keywords    []string `yaml:"keywords,omitempty"`
	FilePattern string   `yaml:"file_pattern,omitempty"`
	Weight      float64  `yaml:"weight,omitempty"`
}

// Parser handles parsing of SKILL.md files.
type Parser struct{}

// NewParser creates a new skill parser.
func NewParser() *Parser {
	return &Parser{}
}

// detectVersion checks skill format version based on content markers.
// v4: Flat triggers array + Markdown sections (## Role, ## Instructions, ## Examples)
// v3: Object triggers + Markdown sections
// v2: XML tags (<role>, <instructions>) + description-based triggers
// v1: Basic frontmatter only
func (p *Parser) detectVersion(content, frontmatter string) string {
	// Check for XML tags (v2)
	if strings.Contains(content, "<role>") || strings.Contains(content, "<instructions>") {
		return "v2"
	}

	// Check for v4 markers: flat triggers array + required Markdown sections
	hasFlatTriggers := strings.Contains(frontmatter, "triggers:") &&
		!strings.Contains(frontmatter, "keywords:") &&
		!strings.Contains(frontmatter, "file_pattern:")
	hasV4Sections := strings.Contains(content, "## Role") &&
		strings.Contains(content, "## Instructions") &&
		strings.Contains(content, "## Examples")

	if hasFlatTriggers && hasV4Sections {
		return "v4"
	}

	// Check for v3 markers: triggers in frontmatter AND Markdown sections
	hasFrontmatterTriggers := strings.Contains(frontmatter, "triggers:")
	hasMarkdownSections := strings.Contains(content, "## Role") ||
		strings.Contains(content, "## Instructions")

	if hasFrontmatterTriggers && hasMarkdownSections {
		return "v3"
	}

	// Default to v1 for simple frontmatter-only skills
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

// parseFrontmatterV3 parses v3 frontmatter using yaml.Unmarshal.
func (p *Parser) parseFrontmatterV3(frontmatter string) (*skillMetaV3, error) {
	var meta skillMetaV3
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if meta.Name == "" {
		return nil, fmt.Errorf("missing name in frontmatter")
	}

	if meta.Triggers != nil {
		if meta.Triggers.Weight == 0 {
			meta.Triggers.Weight = 0.7
		}
		if meta.Triggers.Weight < 0.0 || meta.Triggers.Weight > 1.0 {
			return nil, fmt.Errorf("trigger weight must be between 0.0 and 1.0, got %f", meta.Triggers.Weight)
		}
	}

	return &meta, nil
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

	var result *SkillMeta

	if version == "v3" {
		v3Meta, err := p.parseFrontmatterV3(frontmatter)
		if err != nil {
			return nil, fmt.Errorf("parse v3: %w", err)
		}

		var explicitTriggers []Trigger
		var triggers []string

		if v3Meta.Triggers != nil {
			trigger := Trigger{
				Keywords:     v3Meta.Triggers.Keywords,
				FilePatterns: []string{},
				Weight:       v3Meta.Triggers.Weight,
			}
			if v3Meta.Triggers.FilePattern != "" {
				trigger.FilePatterns = append(trigger.FilePatterns, v3Meta.Triggers.FilePattern)
			}
			if trigger.Weight == 0 {
				trigger.Weight = 0.7
			}
			explicitTriggers = []Trigger{trigger}
			triggers = p.triggersToStrings(explicitTriggers)
		} else {
			descriptionTriggers := p.extractTriggers(v3Meta.Description)
			triggers = descriptionTriggers
			explicitTriggers = p.stringsToTriggers(descriptionTriggers, 0.5)
		}

		result = &SkillMeta{
			Name:             v3Meta.Name,
			Description:      v3Meta.Description,
			Version:          v3Meta.Version,
			Author:           v3Meta.Author,
			Tags:             v3Meta.Tags,
			AllowedTools:     v3Meta.AllowedTools,
			Triggers:         triggers,
			ExplicitTriggers: explicitTriggers,
			FilePath:         path,
			StructureVersion: "v3",
			DependsOn:        v3Meta.DependsOn,
			DelegatesTo:      v3Meta.DelegatesTo,
		}
	} else if version == "v2" {
		v2Meta, err := p.parseFrontmatterV2(frontmatter)
		if err != nil {
			return nil, fmt.Errorf("parse v2: %w", err)
		}

		var explicitTriggers []Trigger
		var triggers []string

		if len(v2Meta.Triggers) > 0 {
			explicitTriggers = v2Meta.Triggers
			triggers = p.triggersToStrings(explicitTriggers)
		} else {
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
	} else if version == "v4" {
		v4Meta, err := p.parseFrontmatterV4(frontmatter)
		if err != nil {
			return nil, fmt.Errorf("parse v4: %w", err)
		}

		contentStr := string(content)

		result = &SkillMeta{
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

// extractXMLTag extracts content between opening and closing XML tags.
func (p *Parser) extractXMLTag(content, tagName string) string {
	openTag := "<" + tagName + ">"
	closeTag := "</" + tagName + ">"

	openIdx := strings.Index(content, openTag)
	if openIdx == -1 {
		return ""
	}

	closeIdx := strings.Index(content, closeTag)
	if closeIdx == -1 {
		return ""
	}

	return strings.TrimSpace(content[openIdx+len(openTag) : closeIdx])
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

// loadReferences checks if references/ directory exists and returns list of valid reference files.
// Validates structure: max depth 1, no frontmatter in .md files.
func (p *Parser) loadReferences(skillDir string) ([]string, error) {
	refsDir := filepath.Join(skillDir, "references")
	if _, err := os.Stat(refsDir); os.IsNotExist(err) {
		return nil, nil
	}

	var refs []string

	err := filepath.Walk(refsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == refsDir {
			return nil
		}

		relPath, err := filepath.Rel(refsDir, path)
		if err != nil {
			return err
		}

		depth := strings.Count(relPath, string(filepath.Separator))

		if info.IsDir() {
			if depth > 1 {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) == ".md" {
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read reference %s: %w", path, err)
			}

			if strings.Contains(string(content), "---") {
				return fmt.Errorf("reference file has frontmatter: %s", path)
			}

			refs = append(refs, relPath)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk references: %w", err)
	}

	return refs, nil
}
