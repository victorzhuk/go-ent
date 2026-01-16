package generation

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// SpecAnalysis contains the result of analyzing a spec file.
type SpecAnalysis struct {
	Patterns   []PatternMatch `json:"patterns"`
	Components []Component    `json:"components"`
	Archetype  string         `json:"recommended_archetype"`
	Confidence float64        `json:"confidence"`
}

// PatternMatch represents a detected pattern in the spec.
type PatternMatch struct {
	Pattern  string   `json:"pattern"`
	Evidence []string `json:"evidence"`
	Score    float64  `json:"score"`
}

// Component represents an identified component from the spec.
type Component struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Templates   []string `json:"recommended_templates"`
}

// patternRule defines a pattern detection rule.
type patternRule struct {
	Name     string
	Keywords []string
	Weight   float64
}

var patterns = []patternRule{
	{
		Name:     "crud",
		Keywords: []string{"create", "read", "update", "delete", "retrieve", "list", "get by id"},
		Weight:   1.0,
	},
	{
		Name:     "api",
		Keywords: []string{"endpoint", "request", "response", "http", "rest", "api"},
		Weight:   0.9,
	},
	{
		Name:     "async",
		Keywords: []string{"queue", "worker", "background", "async", "message", "event"},
		Weight:   0.8,
	},
	{
		Name:     "auth",
		Keywords: []string{"authenticate", "authorize", "permission", "role", "token", "session"},
		Weight:   0.7,
	},
	{
		Name:     "grpc",
		Keywords: []string{"grpc", "rpc", "protobuf", "proto", "service method"},
		Weight:   0.9,
	},
	{
		Name:     "mcp",
		Keywords: []string{"mcp", "tool", "resource", "prompt", "model context protocol"},
		Weight:   1.0,
	},
}

// AnalyzeSpec analyzes a spec file to identify patterns and components.
func AnalyzeSpec(specPath string) (*SpecAnalysis, error) {
	content, err := os.ReadFile(specPath) // #nosec G304 -- controlled file path
	if err != nil {
		return nil, fmt.Errorf("read spec: %w", err)
	}

	text := strings.ToLower(string(content))
	analysis := &SpecAnalysis{
		Patterns: []PatternMatch{},
	}

	// Detect patterns
	for _, rule := range patterns {
		matches := detectPattern(text, rule)
		if matches.Score > 0 {
			analysis.Patterns = append(analysis.Patterns, matches)
		}
	}

	// Extract components from requirements
	components := extractComponents(string(content))
	analysis.Components = components

	return analysis, nil
}

// detectPattern checks if a pattern is present in the text.
func detectPattern(text string, rule patternRule) PatternMatch {
	var evidence []string
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		for _, keyword := range rule.Keywords {
			if strings.Contains(line, keyword) {
				evidence = append(evidence, strings.TrimSpace(line))
				break
			}
		}
	}

	score := 0.0
	if len(evidence) > 0 {
		// Score based on evidence count and weight
		score = float64(len(evidence)) * rule.Weight / 10.0
		if score > 1.0 {
			score = 1.0
		}
	}

	return PatternMatch{
		Pattern:  rule.Name,
		Evidence: evidence,
		Score:    score,
	}
}

// extractComponents extracts component definitions from requirements.
func extractComponents(content string) []Component {
	var components []Component

	// Match requirement patterns like "The system SHALL..."
	reqPattern := regexp.MustCompile(`(?i)the system shall\s+(.+?)(?:\.|$)`)
	matches := reqPattern.FindAllStringSubmatch(content, -1)

	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		desc := strings.TrimSpace(match[1])
		if seen[desc] || len(desc) < 10 {
			continue
		}
		seen[desc] = true

		// Infer component type from description
		compType := inferComponentType(desc)
		if compType == "" {
			continue
		}

		name := generateComponentName(desc)
		components = append(components, Component{
			Name:        name,
			Type:        compType,
			Description: desc,
			Templates:   getTemplatesForType(compType),
		})

		// Limit to prevent overwhelming output
		if len(components) >= 10 {
			break
		}
	}

	return components
}

// inferComponentType infers the component type from description.
func inferComponentType(desc string) string {
	lower := strings.ToLower(desc)

	if strings.Contains(lower, "endpoint") || strings.Contains(lower, "api") {
		return "handler"
	}
	if strings.Contains(lower, "repository") || strings.Contains(lower, "database") || strings.Contains(lower, "store") {
		return "repository"
	}
	if strings.Contains(lower, "service") || strings.Contains(lower, "use case") || strings.Contains(lower, "business logic") {
		return "usecase"
	}
	if strings.Contains(lower, "worker") || strings.Contains(lower, "background") {
		return "worker"
	}

	return "usecase" // default
}

// generateComponentName creates a component name from description.
func generateComponentName(desc string) string {
	// Extract first few words and convert to snake_case
	words := strings.Fields(desc)
	if len(words) > 3 {
		words = words[:3]
	}

	name := strings.Join(words, "_")
	name = strings.ToLower(name)
	name = regexp.MustCompile(`[^a-z0-9_]+`).ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")

	return name
}

// getTemplatesForType returns recommended templates for a component type.
func getTemplatesForType(compType string) []string {
	switch compType {
	case "handler":
		return []string{"handler", "dto"}
	case "repository":
		return []string{"repository", "models", "mappers"}
	case "usecase":
		return []string{"usecase", "dto"}
	case "worker":
		return []string{"worker", "handler"}
	default:
		return []string{}
	}
}
