package spec

import (
	"fmt"
	"regexp"
	"strings"
)

// DeltaSpec represents parsed delta operations from a change.
type DeltaSpec struct {
	Added    []Requirement
	Modified []Requirement
	Removed  []RemovedRequirement
	Renamed  []RenamedRequirement
}

// Requirement represents a parsed requirement with its content.
type Requirement struct {
	Name    string
	Content string
	Line    int
	EndLine int
}

// RemovedRequirement represents a requirement to be removed.
type RemovedRequirement struct {
	Name      string
	Reason    string
	Migration string
}

// RenamedRequirement represents a requirement rename.
type RenamedRequirement struct {
	FromName string
	ToName   string
}

// ParseDeltaSpec parses delta operations from spec content.
func ParseDeltaSpec(content string) (*DeltaSpec, error) {
	delta := &DeltaSpec{}
	lines := strings.Split(content, "\n")

	var currentSection string
	var currentReq *Requirement
	var contentLines []string

	sectionRe := regexp.MustCompile(`(?i)^##\s+(ADDED|MODIFIED|REMOVED|RENAMED)\s+Requirements?`)
	requirementRe := regexp.MustCompile(`^###\s+Requirement:\s+(.+)$`)
	renamedFromRe := regexp.MustCompile(`(?i)^-\s+FROM:\s+\x60?###\s+Requirement:\s+(.+?)\x60?$`)
	renamedToRe := regexp.MustCompile(`(?i)^-\s+TO:\s+\x60?###\s+Requirement:\s+(.+?)\x60?$`)

	saveCurrentReq := func() {
		if currentReq != nil && currentSection != "" {
			currentReq.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))
			switch strings.ToUpper(currentSection) {
			case "ADDED":
				delta.Added = append(delta.Added, *currentReq)
			case "MODIFIED":
				delta.Modified = append(delta.Modified, *currentReq)
			}
		}
		currentReq = nil
		contentLines = nil
	}

	var renameFrom, renameTo string

	for i, line := range lines {
		// Check for section headers
		if matches := sectionRe.FindStringSubmatch(line); matches != nil {
			saveCurrentReq()
			currentSection = strings.ToUpper(matches[1])
			continue
		}

		// Handle RENAMED section specially
		if strings.ToUpper(currentSection) == "RENAMED" {
			if matches := renamedFromRe.FindStringSubmatch(strings.TrimSpace(line)); matches != nil {
				renameFrom = strings.TrimSpace(matches[1])
			}
			if matches := renamedToRe.FindStringSubmatch(strings.TrimSpace(line)); matches != nil {
				renameTo = strings.TrimSpace(matches[1])
			}
			if renameFrom != "" && renameTo != "" {
				delta.Renamed = append(delta.Renamed, RenamedRequirement{
					FromName: renameFrom,
					ToName:   renameTo,
				})
				renameFrom = ""
				renameTo = ""
			}
			continue
		}

		// Check for requirement headers
		if matches := requirementRe.FindStringSubmatch(strings.TrimSpace(line)); matches != nil {
			saveCurrentReq()
			currentReq = &Requirement{
				Name: strings.TrimSpace(matches[1]),
				Line: i + 1,
			}
			contentLines = []string{line}
			continue
		}

		// Handle REMOVED section
		if strings.ToUpper(currentSection) == "REMOVED" && currentReq != nil {
			if strings.HasPrefix(strings.TrimSpace(line), "**Reason**:") {
				reason := strings.TrimPrefix(strings.TrimSpace(line), "**Reason**:")
				delta.Removed = append(delta.Removed, RemovedRequirement{
					Name:   currentReq.Name,
					Reason: strings.TrimSpace(reason),
				})
				currentReq = nil
			}
			continue
		}

		// Accumulate content for current requirement
		if currentReq != nil {
			contentLines = append(contentLines, line)
		}
	}

	saveCurrentReq()

	return delta, nil
}

// MergeDeltas merges delta operations into base spec content.
func MergeDeltas(baseSpec string, delta *DeltaSpec) (string, error) {
	result := baseSpec

	// Handle RENAMED first (before MODIFIED which might reference new names)
	for _, renamed := range delta.Renamed {
		result = renameRequirement(result, renamed.FromName, renamed.ToName)
	}

	// Handle REMOVED
	for _, removed := range delta.Removed {
		result = removeRequirement(result, removed.Name, removed.Reason)
	}

	// Handle MODIFIED (replace entire requirement)
	for _, modified := range delta.Modified {
		result = replaceRequirement(result, modified.Name, modified.Content)
	}

	// Handle ADDED (append to end)
	for _, added := range delta.Added {
		result = appendRequirement(result, added.Content)
	}

	return result, nil
}

// renameRequirement renames a requirement in the spec.
func renameRequirement(content, fromName, toName string) string {
	re := regexp.MustCompile(`(###\s+Requirement:\s+)` + regexp.QuoteMeta(fromName) + `(\s*\n)`)
	return re.ReplaceAllString(content, "${1}"+toName+"${2}")
}

// removeRequirement removes a requirement and adds a removal comment.
func removeRequirement(content, name, reason string) string {
	lines := strings.Split(content, "\n")
	var result []string

	reqRe := regexp.MustCompile(`^###\s+Requirement:\s+` + regexp.QuoteMeta(name) + `\s*$`)
	nextReqRe := regexp.MustCompile(`^###\s+Requirement:\s+`)

	inTargetReq := false
	for _, line := range lines {
		if reqRe.MatchString(strings.TrimSpace(line)) {
			inTargetReq = true
			continue
		}
		if inTargetReq && nextReqRe.MatchString(strings.TrimSpace(line)) {
			inTargetReq = false
		}
		if !inTargetReq {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// replaceRequirement replaces an existing requirement with new content.
func replaceRequirement(content, name, newContent string) string {
	lines := strings.Split(content, "\n")
	var result []string

	reqRe := regexp.MustCompile(`^###\s+Requirement:\s+` + regexp.QuoteMeta(name) + `\s*$`)
	nextReqRe := regexp.MustCompile(`^###\s+Requirement:\s+`)
	nextSectionRe := regexp.MustCompile(`^##\s+`)

	inTargetReq := false
	replaced := false

	for _, line := range lines {
		if reqRe.MatchString(strings.TrimSpace(line)) {
			inTargetReq = true
			result = append(result, strings.Split(newContent, "\n")...)
			replaced = true
			continue
		}
		if inTargetReq {
			if nextReqRe.MatchString(strings.TrimSpace(line)) || nextSectionRe.MatchString(strings.TrimSpace(line)) {
				inTargetReq = false
				result = append(result, line)
			}
			// Skip lines while in target requirement
			continue
		}
		result = append(result, line)
	}

	if !replaced {
		return fmt.Sprintf("%s\n\n%s", content, newContent)
	}

	return strings.Join(result, "\n")
}

// appendRequirement adds a new requirement to the end of the spec.
func appendRequirement(content, reqContent string) string {
	content = strings.TrimRight(content, "\n\r\t ")
	return fmt.Sprintf("%s\n\n%s\n", content, reqContent)
}
