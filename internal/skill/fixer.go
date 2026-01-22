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

// fixTagTypos fixes common singular→plural tag typos.
func (f *Fixer) fixTagTypos(content string) (string, []FixChange) {
	var changes []FixChange
	normalized := content

	typos := map[string]string{
		"<instruction>":  "<instructions>",
		"</instruction>": "</instructions>",
		"<example>":      "<examples>",
		"</example>":     "</examples>",
		"<constraint>":   "<constraints>",
		"</constraint>":  "</constraints>",
		"<edge_case>":    "<edge_cases>",
		"</edge_case>":   "</edge_cases>",
	}

	for typo, correct := range typos {
		if strings.Contains(normalized, typo) {
			normalized = strings.ReplaceAll(normalized, typo, correct)
			changes = append(changes, FixChange{
				Rule:    "tag-typo",
				Message: fmt.Sprintf("fixed tag typo: %s → %s", typo, correct),
			})
		}
	}

	return normalized, changes
}

// FixXMLSections normalizes XML-like section formatting in skill content.
func (f *Fixer) FixXMLSections(content string) (string, []FixChange, error) {
	var changes []FixChange

	tags := []string{"triggers", "role", "instructions", "constraints", "edge_cases", "examples", "output_format"}
	normalized := content

	normalized, typoChanges := f.fixTagTypos(normalized)
	changes = append(changes, typoChanges...)

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

// FixCommonIssues combines all auto-fix operations with trigger suggestions.
func (f *Fixer) FixCommonIssues(content, filePath string) (string, []FixChange, error) {
	var allChanges []FixChange

	fixed, valChanges, err := f.FixValidationIssues(content)
	if err != nil {
		return content, allChanges, fmt.Errorf("fix validation issues: %w", err)
	}
	allChanges = append(allChanges, valChanges...)

	fixed, triggerChanges := f.suggestTriggers(fixed, filePath)
	allChanges = append(allChanges, triggerChanges...)

	return fixed, allChanges, nil
}

// suggestTriggers adds trigger pattern suggestions based on file metadata.
func (f *Fixer) suggestTriggers(content, filePath string) (string, []FixChange) {
	var changes []FixChange

	lines := strings.Split(content, "\n")
	fmStart := -1
	fmEnd := -1
	for i := 0; i < len(lines); i++ {
		if lines[i] == "---" {
			if fmStart == -1 {
				fmStart = i
			} else if fmEnd == -1 {
				fmEnd = i
				break
			}
		}
	}

	if fmStart == -1 || fmEnd == -1 {
		return content, changes
	}

	if strings.Contains(content, "triggers:") {
		return content, changes
	}

	frontmatterLines := lines[fmStart+1 : fmEnd]
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(strings.Join(frontmatterLines, "\n")), &data); err != nil {
		return content, changes
	}

	skillName := ""
	if name, ok := data["name"].(string); ok {
		skillName = name
	}

	suggestedTriggers := f.buildTriggerSuggestions(skillName, filePath)

	if len(suggestedTriggers) == 0 {
		return content, changes
	}

	data["triggers"] = suggestedTriggers

	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(data); err != nil {
		_ = encoder.Close()
		return content, changes
	}
	_ = encoder.Close()

	newFrontmatter := strings.TrimSpace(strings.TrimSuffix(buf.String(), "\n"))
	var body string
	if fmEnd+1 < len(lines) {
		body = strings.Join(lines[fmEnd+1:], "\n")
	}
	newContent := fmt.Sprintf("---\n%s\n---\n%s", newFrontmatter, body)

	changes = append(changes, FixChange{
		Rule:    "trigger-suggestions",
		Message: fmt.Sprintf("added %d trigger suggestion(s) based on skill name and location", len(suggestedTriggers)),
	})

	changes = append(changes, f.encodeTriggerChanges(suggestedTriggers)...)

	return newContent, changes
}

// buildTriggerSuggestions generates trigger patterns based on skill metadata.
func (f *Fixer) buildTriggerSuggestions(skillName, filePath string) []map[string]interface{} {
	var triggers []map[string]interface{}

	if skillName == "" {
		return triggers
	}

	keywordPatterns := f.extractKeywords(skillName)

	for _, kw := range keywordPatterns {
		trigger := map[string]interface{}{
			"patterns": []string{fmt.Sprintf("%s.*", kw)},
			"weight":   0.7,
		}
		triggers = append(triggers, trigger)
	}

	if len(triggers) == 0 {
		triggers = append(triggers, map[string]interface{}{
			"patterns": []string{fmt.Sprintf("%s.*", skillName)},
			"weight":   0.5,
		})
	}

	return triggers
}

// extractKeywords extracts trigger keywords from skill name.
func (f *Fixer) extractKeywords(name string) []string {
	keywords := []string{}
	parts := strings.Split(name, "-")

	for _, part := range parts {
		if (len(part) >= 2 && !f.isCommonWord(part)) || part == "go" {
			keywords = append(keywords, part)
		}
	}

	return keywords
}

// isCommonWord checks if a word is too generic for trigger.
func (f *Fixer) isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the":  true,
		"and":  true,
		"for":  true,
		"are":  true,
		"but":  true,
		"not":  true,
		"you":  true,
		"all":  true,
		"can":  true,
		"has":  true,
		"had":  true,
		"was":  true,
		"were": true,
		"been": true,
		"be":   true,
		"this": true,
		"that": true,
		"have": true,
		"from": true,
		"with": true,
		"use":  true,
		"make": true,
		"get":  true,
		"set":  true,
		"add":  true,
		"code": true,
	}

	return commonWords[strings.ToLower(word)]
}

// encodeTriggerChanges creates individual change entries for triggers.
func (f *Fixer) encodeTriggerChanges(triggers []map[string]interface{}) []FixChange {
	var changes []FixChange

	for _, trigger := range triggers {
		patterns := trigger["patterns"].([]string)
		if len(patterns) > 0 {
			changes = append(changes, FixChange{
				Rule:    "trigger-pattern",
				Message: fmt.Sprintf("suggested trigger pattern: %s", patterns[0]),
			})
		}
	}

	return changes
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

// DryRunFile analyzes what would be fixed without modifying the file.
func (f *Fixer) DryRunFile(content string) (*FixResult, []string, error) {
	var allChanges []FixChange
	var diffs []string

	fmNormalized, fmChanges, err := f.FixFrontmatter(content)
	if err != nil {
		return nil, nil, fmt.Errorf("fix frontmatter: %w", err)
	}
	allChanges = append(allChanges, fmChanges...)

	if fmNormalized != content {
		diff := generateDiff(content, fmNormalized)
		diffs = append(diffs, diff)
	}

	xmlNormalized, xmlChanges, err := f.FixXMLSections(fmNormalized)
	if err != nil {
		return nil, nil, fmt.Errorf("fix xml sections: %w", err)
	}
	allChanges = append(allChanges, xmlChanges...)

	if xmlNormalized != fmNormalized {
		diff := generateDiff(fmNormalized, xmlNormalized)
		diffs = append(diffs, diff)
	}

	valNormalized, valChanges, err := f.FixValidationIssues(xmlNormalized)
	if err != nil {
		return nil, nil, fmt.Errorf("fix validation issues: %w", err)
	}
	allChanges = append(allChanges, valChanges...)

	if valNormalized != xmlNormalized {
		diff := generateDiff(xmlNormalized, valNormalized)
		diffs = append(diffs, diff)
	}

	if len(allChanges) == 0 {
		return &FixResult{
			Fixed:   false,
			Changes: []FixChange{},
		}, nil, nil
	}

	return &FixResult{
		Fixed:   true,
		Changes: allChanges,
	}, diffs, nil
}

// generateDiff creates a simple line-by-line diff.
func generateDiff(original, modified string) string {
	origLines := strings.Split(original, "\n")
	modLines := strings.Split(modified, "\n")

	var diff strings.Builder

	maxLines := len(origLines)
	if len(modLines) > maxLines {
		maxLines = len(modLines)
	}

	for i := 0; i < maxLines; i++ {
		origLine := ""
		modLine := ""

		if i < len(origLines) {
			origLine = origLines[i]
		}
		if i < len(modLines) {
			modLine = modLines[i]
		}

		if origLine == modLine {
			diff.WriteString("  " + origLine + "\n")
		} else {
			if origLine != "" {
				diff.WriteString("- " + origLine + "\n")
			}
			if modLine != "" {
				diff.WriteString("+ " + modLine + "\n")
			}
		}
	}

	return diff.String()
}
