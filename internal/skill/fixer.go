package skill

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// FixResult describes what was fixed in a skill file.
type FixResult struct {
	File    string
	Fixed   bool
	Changes []FixChange
}

// FixChange describes a single fix applied to a file.
type FixChange struct {
	Rule    string
	Message string
}

// Fixer applies auto-fixes to skill files.
type Fixer struct{}

// NewFixer creates a new fixer instance.
func NewFixer() *Fixer {
	return &Fixer{}
}

// FixFrontmatter normalizes YAML frontmatter formatting.
func (f *Fixer) FixFrontmatter(content string) (string, []FixChange, error) {
	var changes []FixChange

	lines := strings.Split(content, "\n")
	if len(lines) < 3 || lines[0] != "---" {
		return content, changes, nil
	}

	fmEnd := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			fmEnd = i
			break
		}
	}

	if fmEnd == -1 {
		return content, changes, nil
	}

	frontmatterLines := lines[1:fmEnd]
	originalFrontmatter := strings.Join(frontmatterLines, "\n")
	body := strings.Join(lines[fmEnd+1:], "\n")

	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(originalFrontmatter), &data); err != nil {
		return content, changes, fmt.Errorf("parse yaml: %w", err)
	}

	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(data); err != nil {
		_ = encoder.Close()
		return content, changes, fmt.Errorf("encode yaml: %w", err)
	}
	_ = encoder.Close()

	newFrontmatter := strings.TrimSpace(strings.TrimSuffix(buf.String(), "\n"))
	normalizedContent := fmt.Sprintf("---\n%s\n---\n%s", newFrontmatter, body)

	if strings.TrimSpace(originalFrontmatter) == newFrontmatter {
		return content, changes, nil
	}

	changes = append(changes, FixChange{
		Rule:    "frontmatter-format",
		Message: "normalized YAML frontmatter (sorted keys, formatted indentation)",
	})

	return normalizedContent, changes, nil
}

// FixFile fixes frontmatter and XML sections in a skill file.
func (f *Fixer) FixFile(filePath string, content string) (*FixResult, error) {
	var allChanges []FixChange

	fmNormalized, fmChanges, err := f.FixFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("fix frontmatter: %w", err)
	}

	xmlNormalized, xmlChanges, err := f.FixXMLSections(fmNormalized)
	if err != nil {
		return nil, fmt.Errorf("fix xml sections: %w", err)
	}

	normalized := xmlNormalized
	allChanges = append(allChanges, fmChanges...)
	allChanges = append(allChanges, xmlChanges...)

	if len(allChanges) == 0 && normalized == content {
		return &FixResult{
			File:    filePath,
			Fixed:   false,
			Changes: []FixChange{},
		}, nil
	}

	if err := os.WriteFile(filePath, []byte(normalized), 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	return &FixResult{
		File:    filePath,
		Fixed:   true,
		Changes: allChanges,
	}, nil
}

// HasFixableFrontmatter checks if content has fixable frontmatter issues.
func (f *Fixer) HasFixableFrontmatter(content string) bool {
	normalized, _, err := f.FixFrontmatter(content)
	if err != nil {
		return false
	}
	return normalized != content
}

// HasFixableXML checks if content has fixable XML section issues.
func (f *Fixer) HasFixableXML(content string) bool {
	normalized, _, err := f.FixXMLSections(content)
	if err != nil {
		return false
	}
	return normalized != content
}

// FixXMLSections normalizes XML-like section formatting in skill content.
func (f *Fixer) FixXMLSections(content string) (string, []FixChange, error) {
	var changes []FixChange

	tags := []string{"triggers", "role", "instructions", "constraints", "edge_cases", "examples", "output_format"}
	normalized := content

	for _, tag := range tags {
		openTag := "<" + tag + ">"
		closeTag := "</" + tag + ">"

		if !strings.Contains(normalized, openTag) {
			continue
		}

		if !strings.Contains(normalized, closeTag) {
			continue
		}

		normalized = f.normalizeTagSection(normalized, tag)
		changes = append(changes, FixChange{
			Rule:    "xml-section-format",
			Message: fmt.Sprintf("normalized <%s> section formatting", tag),
		})
	}

	if len(changes) == 0 {
		return content, changes, nil
	}

	return normalized, changes, nil
}

// normalizeTagSection normalizes a specific XML section.
func (f *Fixer) normalizeTagSection(content, tagName string) string {
	openTag := "<" + tagName + ">"
	closeTag := "</" + tagName + ">"

	openIdx := strings.Index(content, openTag)
	if openIdx == -1 {
		return content
	}

	closeIdx := strings.Index(content, closeTag)
	if closeIdx == -1 {
		return content
	}

	before := content[:openIdx+len(openTag)]
	after := content[closeIdx:]

	sectionContent := content[openIdx+len(openTag) : closeIdx]
	normalizedSection := f.normalizeSectionContent(sectionContent)

	return before + "\n" + normalizedSection + "\n" + after
}

// normalizeSectionContent normalizes content within XML tags.
func (f *Fixer) normalizeSectionContent(content string) string {
	lines := strings.Split(content, "\n")
	var normalized []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		normalized = append(normalized, trimmed)
	}

	if len(normalized) == 0 {
		return ""
	}

	return strings.Join(normalized, "\n")
}

// FixValidationIssues auto-fixes common validation issues.
func (f *Fixer) FixValidationIssues(content string) (string, []FixChange, error) {
	var changes []FixChange
	var err error

	fixed, fmChanges, err := f.fixFrontmatterFields(content)
	if err != nil {
		return content, changes, fmt.Errorf("fix frontmatter fields: %w", err)
	}
	changes = append(changes, fmChanges...)
	content = fixed

	fixed, xmlChanges, err := f.fixMissingClosingTags(content)
	if err != nil {
		return content, changes, fmt.Errorf("fix missing closing tags: %w", err)
	}
	changes = append(changes, xmlChanges...)
	content = fixed

	fixed, listChanges, err := f.fixConstraintListFormat(content)
	if err != nil {
		return content, changes, fmt.Errorf("fix constraint list format: %w", err)
	}
	changes = append(changes, listChanges...)

	return fixed, changes, nil
}

// fixFrontmatterFields adds missing required fields and normalizes invalid values.
func (f *Fixer) fixFrontmatterFields(content string) (string, []FixChange, error) {
	var changes []FixChange

	lines := strings.Split(content, "\n")
	if len(lines) < 3 || lines[0] != "---" {
		return content, changes, nil
	}

	fmEnd := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			fmEnd = i
			break
		}
	}

	if fmEnd == -1 {
		return content, changes, nil
	}

	frontmatterLines := lines[1:fmEnd]
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(strings.Join(frontmatterLines, "\n")), &data); err != nil {
		return content, changes, err
	}

	var modified bool

	if _, ok := data["name"]; !ok {
		data["name"] = "unnamed-skill"
		modified = true
		changes = append(changes, FixChange{
			Rule:    "frontmatter-name",
			Message: "added missing 'name' field with default value",
		})
	}

	if name, ok := data["name"].(string); ok {
		normalized := normalizeSkillName(name)
		if normalized != name {
			data["name"] = normalized
			modified = true
			changes = append(changes, FixChange{
				Rule:    "name-format",
				Message: fmt.Sprintf("normalized name: '%s' → '%s'", name, normalized),
			})
		}
	}

	if _, ok := data["description"]; !ok {
		data["description"] = "Auto-generated description"
		modified = true
		changes = append(changes, FixChange{
			Rule:    "frontmatter-description",
			Message: "added missing 'description' field with default value",
		})
	}

	if version, ok := data["version"]; ok && version != "" {
		var versionStr string
		switch v := version.(type) {
		case string:
			versionStr = v
		case int:
			versionStr = fmt.Sprintf("%d", v)
		case int64:
			versionStr = fmt.Sprintf("%d", v)
		case float64:
			if v == float64(int(v)) {
				versionStr = fmt.Sprintf("%.0f", v)
			} else {
				versionStr = fmt.Sprintf("%g", v)
			}
		}

		if versionStr != "" {
			normalized := normalizeVersion(versionStr)
			if normalized != versionStr {
				data["version"] = normalized
				modified = true
				changes = append(changes, FixChange{
					Rule:    "version-format",
					Message: fmt.Sprintf("normalized version: '%s' → '%s'", versionStr, normalized),
				})
			}
		}
	}

	if !modified {
		return content, changes, nil
	}

	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(data); err != nil {
		_ = encoder.Close()
		return content, changes, err
	}
	_ = encoder.Close()

	newFrontmatter := strings.TrimSpace(strings.TrimSuffix(buf.String(), "\n"))
	body := strings.Join(lines[fmEnd+1:], "\n")
	return fmt.Sprintf("---\n%s\n---\n%s", newFrontmatter, body), changes, nil
}

// fixMissingClosingTags adds missing closing tags for XML sections.
func (f *Fixer) fixMissingClosingTags(content string) (string, []FixChange, error) {
	var changes []FixChange
	normalized := content

	tags := []string{"role", "instructions", "examples", "constraints", "edge_cases", "output_format"}

	for _, tag := range tags {
		openTag := "<" + tag + ">"
		closeTag := "</" + tag + ">"

		if !strings.Contains(normalized, openTag) {
			continue
		}

		if strings.Contains(normalized, closeTag) {
			continue
		}

		openIdx := strings.LastIndex(normalized, openTag)
		if openIdx == -1 {
			continue
		}

		insertPoint := openIdx + len(openTag)
		normalized = normalized[:insertPoint] + "\n" + closeTag + normalized[insertPoint:]

		changes = append(changes, FixChange{
			Rule:    "xml-closing-tag",
			Message: fmt.Sprintf("added missing closing tag: %s", closeTag),
		})
	}

	return normalized, changes, nil
}

// fixConstraintListFormat converts non-list constraint items to list format.
func (f *Fixer) fixConstraintListFormat(content string) (string, []FixChange, error) {
	var changes []FixChange

	if !strings.Contains(content, "<constraints>") {
		return content, changes, nil
	}

	if !strings.Contains(content, "</constraints>") {
		return content, changes, nil
	}

	openTag := "<constraints>"
	closeTag := "</constraints>"
	openIdx := strings.Index(content, openTag)
	closeIdx := strings.Index(content, closeTag)

	if openIdx == -1 || closeIdx == -1 {
		return content, changes, nil
	}

	before := content[:openIdx]
	section := content[openIdx+len(openTag) : closeIdx]
	after := content[closeIdx:]

	lines := strings.Split(section, "\n")
	var normalized []string
	hasChanges := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "- ") {
			normalized = append(normalized, trimmed)
			continue
		}

		normalized = append(normalized, "- "+trimmed)
		hasChanges = true
	}

	if !hasChanges {
		return content, changes, nil
	}

	changes = append(changes, FixChange{
		Rule:    "constraint-list-format",
		Message: fmt.Sprintf("converted %d constraint(s) to list format", len(normalized)),
	})

	result := before + openTag + "\n" + strings.Join(normalized, "\n") + "\n" + closeTag + after
	return result, changes, nil
}

// normalizeSkillName normalizes a skill name to lowercase with hyphens.
func normalizeSkillName(name string) string {
	normalized := strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9-]`)
	normalized = re.ReplaceAllString(normalized, "-")
	re = regexp.MustCompile(`-+`)
	normalized = re.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	if normalized == "" {
		normalized = "unnamed-skill"
	}
	return normalized
}

// normalizeVersion normalizes a version string to semver format.
func normalizeVersion(version string) string {
	re := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)
	if re.MatchString(version) {
		return version
	}

	re = regexp.MustCompile(`^v?(\d+)\.(\d+)$`)
	if m := re.FindStringSubmatch(version); len(m) == 3 {
		return fmt.Sprintf("%s.%s.0", m[1], m[2])
	}

	re = regexp.MustCompile(`^v?(\d+)$`)
	if m := re.FindStringSubmatch(version); len(m) == 2 {
		return fmt.Sprintf("%s.0.0", m[1])
	}

	return "1.0.0"
}

// HasFixableValidationIssues checks if content has fixable validation issues.
func (f *Fixer) HasFixableValidationIssues(content string) bool {
	fixed, _, err := f.FixValidationIssues(content)
	if err != nil {
		return false
	}
	return fixed != content
}

// FixValidationFile fixes validation issues in a skill file.
func (f *Fixer) FixValidationFile(filePath string, content string) (*FixResult, error) {
	fixed, changes, err := f.FixValidationIssues(content)
	if err != nil {
		return nil, err
	}

	if len(changes) == 0 && fixed == content {
		return &FixResult{
			File:    filePath,
			Fixed:   false,
			Changes: []FixChange{},
		}, nil
	}

	if err := os.WriteFile(filePath, []byte(fixed), 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	return &FixResult{
		File:    filePath,
		Fixed:   true,
		Changes: changes,
	}, nil
}
