package skill

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/victorzhuk/go-ent/internal/skill/domain"
)

const defaultTriggerWeight = 0.7

type skillMetaV4 struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Triggers    []string `yaml:"triggers"`
}

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

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

func (p *Parser) detectCategory(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")

	skillsIdx := strings.LastIndex(path, "/skills/")
	var afterSkills string
	if skillsIdx != -1 {
		afterSkills = path[skillsIdx+len("/skills/"):]
	} else if strings.HasPrefix(path, "skills/") {
		afterSkills = path[len("skills/"):]
	} else {
		return ""
	}

	parts := strings.Split(afterSkills, "/")
	if len(parts) >= 1 && parts[0] != "" {
		return parts[0]
	}

	return ""
}

func (p *Parser) ParseSkillFile(path string) (*domain.Info, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer func() { _ = f.Close() }()

	frontmatter, err := p.extractFrontmatter(f)
	if err != nil {
		return nil, fmt.Errorf("extract frontmatter: %w", err)
	}

	content, err := os.ReadFile(path)
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

	result, err := domain.NewInfo(v4Meta.Name, v4Meta.Description, "")
	if err != nil {
		return nil, err
	}

	result.Triggers = v4Meta.Triggers
	result.ExplicitTriggers = p.stringsToTriggers(v4Meta.Triggers, defaultTriggerWeight)
	result.FilePath = path
	result.Category = p.detectCategory(path)
	result.StructureVersion = "v4"
	result.Role = p.extractMarkdownSection(contentStr, "Role")
	result.Instructions = p.extractMarkdownSection(contentStr, "Instructions")
	result.Examples = p.extractMarkdownSection(contentStr, "Examples")
	result.References = p.extractReferencesSection(contentStr)

	return result, nil
}

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

func (p *Parser) extractMarkdownSection(content, sectionName string) string {
	heading := "## " + sectionName
	startIdx := strings.Index(content, heading)
	if startIdx == -1 {
		return ""
	}

	contentStart := startIdx + len(heading)
	newlineIdx := strings.Index(content[contentStart:], "\n")
	if newlineIdx == -1 {
		return ""
	}
	contentStart += newlineIdx + 1

	remainingContent := content[contentStart:]
	nextHeadingIdx := strings.Index(remainingContent, "\n## ")

	var sectionContent string
	if nextHeadingIdx == -1 {
		sectionContent = remainingContent
	} else {
		sectionContent = remainingContent[:nextHeadingIdx]
	}

	return strings.TrimSpace(sectionContent)
}

func (p *Parser) extractReferencesSection(content string) []string {
	heading := "## References"
	startIdx := strings.Index(content, heading)
	if startIdx == -1 {
		return nil
	}

	contentStart := startIdx + len(heading)
	newlineIdx := strings.Index(content[contentStart:], "\n")
	if newlineIdx == -1 {
		return nil
	}
	contentStart += newlineIdx + 1

	remainingContent := content[contentStart:]
	nextHeadingIdx := strings.Index(remainingContent, "\n## ")

	var sectionContent string
	if nextHeadingIdx == -1 {
		sectionContent = remainingContent
	} else {
		sectionContent = remainingContent[:nextHeadingIdx]
	}

	var refs []string
	lines := strings.Split(sectionContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "- [") {
			continue
		}

		openParen := strings.LastIndex(line, "(")
		closeParen := strings.LastIndex(line, ")")
		if openParen != -1 && closeParen != -1 && openParen < closeParen {
			path := line[openParen+1 : closeParen]
			refs = append(refs, path)
		}
	}

	return refs
}

func (p *Parser) stringsToTriggers(strs []string, weight float64) []domain.Trigger {
	triggers := make([]domain.Trigger, 0, len(strs))
	for _, s := range strs {
		triggers = append(triggers, domain.Trigger{
			Keywords: []string{s},
			Weight:   weight,
		})
	}
	return triggers
}
