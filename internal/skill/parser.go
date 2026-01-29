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
	LoadLevel        LoadLevel
	Core             *CoreContent
	Full             *FullContent
}

// CoreContent holds the Level 2 content for a skill.
type CoreContent struct {
	Role         string
	Instructions string
	Constraints  string
	Examples     string
}

// FullContent holds Level 3 content for a skill (complete file body).
type FullContent struct {
	Body       string
	References []string
	Scripts    []string
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

// v3Triggers represents v3 trigger structure in frontmatter.
type v3Triggers struct {
	Keywords    []string `yaml:"keywords,omitempty"`
	FilePattern string   `yaml:"file_pattern,omitempty"`
	Weight      float64  `yaml:"weight,omitempty"`
}

// LoadLevel represents how much of a skill's content has been loaded.
type LoadLevel int

const (
	// LoadMetadata loads only frontmatter + triggers (~100 tokens)
	LoadMetadata LoadLevel = iota

	// LoadCore loads metadata + role + instructions + constraints + examples (<5k tokens)
	LoadCore

	// LoadExtended loads everything including references/, scripts/, detailed docs
	LoadExtended
)

// String returns the string representation of LoadLevel.
func (l LoadLevel) String() string {
	switch l {
	case LoadMetadata:
		return "metadata"
	case LoadCore:
		return "core"
	case LoadExtended:
		return "extended"
	default:
		return "unknown"
	}
}

// Parser handles parsing of SKILL.md files.
type Parser struct{}

// NewParser creates a new skill parser.
func NewParser() *Parser {
	return &Parser{}
}

// detectVersion checks skill format version based on content markers.
// v3: Markdown sections (## Role, ## Instructions) + triggers in frontmatter
// v2: XML tags (<role>, <instructions>) + description-based triggers
// v1: Basic frontmatter only
func (p *Parser) detectVersion(content, frontmatter string) string {
	// Check for XML tags (v2)
	if strings.Contains(content, "<role>") || strings.Contains(content, "<instructions>") {
		return "v2"
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

// ParseSkillFile parses a SKILL.md file and extracts metadata (Level 1).
// Level 1 includes: frontmatter (name, description, triggers, etc.)
// For full content loading, use UpgradeToLevel.
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
			// Convert v3 triggers to internal format
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
			// Fallback to description-based extraction
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
			LoadLevel:        LoadMetadata,
		}
	} else if version == "v2" {
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
			LoadLevel:        LoadMetadata,
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
			LoadLevel:        LoadMetadata,
		}
	}

	return result, nil
}

// UpgradeToLevel upgrades a skill's content to the specified load level.
// If the skill is already at or above the target level, returns nil.
func (p *Parser) UpgradeToLevel(meta *SkillMeta, targetLevel LoadLevel) error {
	if meta.LoadLevel >= targetLevel {
		return nil
	}

	switch targetLevel {
	case LoadCore:
		content, err := os.ReadFile(meta.FilePath) // #nosec G304 -- controlled config/template file path
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}

		contentStr := string(content)
		var core *CoreContent

		// v3 format uses Markdown sections
		if meta.StructureVersion == "v3" {
			core = &CoreContent{
				Role:         p.extractMarkdownSection(contentStr, "Role"),
				Instructions: p.extractMarkdownSection(contentStr, "Instructions"),
				Constraints:  p.extractMarkdownSection(contentStr, "Constraints"),
				Examples:     p.extractMarkdownSection(contentStr, "Examples"),
			}
		} else {
			// v2 and v1 use XML tags
			core = &CoreContent{
				Role:         p.extractXMLTag(contentStr, "role"),
				Instructions: p.extractXMLTag(contentStr, "instructions"),
				Constraints:  p.extractXMLTag(contentStr, "constraints"),
				Examples:     p.extractXMLTag(contentStr, "examples"),
			}
		}

		meta.Core = core
		meta.LoadLevel = LoadCore
		return nil
	case LoadExtended:
		if meta.Core == nil {
			if err := p.UpgradeToLevel(meta, LoadCore); err != nil {
				return err
			}
		}

		content, err := os.ReadFile(meta.FilePath) // #nosec G304 -- controlled config/template file path
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}

		full := &FullContent{
			Body: string(content),
		}
		meta.Full = full
		meta.LoadLevel = LoadExtended
		return nil
	default:
		return nil
	}
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
// Returns content from the heading until the next heading of equal or higher level.
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
