package generation

import (
	"fmt"
)

// archetypeScore represents an archetype with its match score.
type archetypeScore struct {
	Name  string
	Score float64
}

// SelectArchetype selects the best archetype based on spec analysis.
// If an explicit archetype is specified in config, it takes precedence.
func SelectArchetype(analysis *SpecAnalysis, cfg *GenerationConfig, explicitArchetype string) (string, float64, error) {
	// Explicit archetype takes precedence
	if explicitArchetype != "" {
		_, err := GetArchetype(explicitArchetype, cfg)
		if err != nil {
			return "", 0, fmt.Errorf("explicit archetype not found: %w", err)
		}
		return explicitArchetype, 1.0, nil
	}

	// Score each archetype based on pattern matches
	scores := scoreArchetypes(analysis)
	if len(scores) == 0 {
		// No patterns matched, use default
		defaultArch := "standard"
		if cfg != nil && cfg.Defaults.Archetype != "" {
			defaultArch = cfg.Defaults.Archetype
		}
		return defaultArch, 0.0, nil
	}

	// Return highest scoring archetype
	best := scores[0]
	return best.Name, best.Score, nil
}

// scoreArchetypes scores all archetypes based on pattern matches.
func scoreArchetypes(analysis *SpecAnalysis) []archetypeScore {
	scores := make(map[string]float64)

	for _, pattern := range analysis.Patterns {
		archetypes := getArchetypesForPattern(pattern.Pattern)
		for _, arch := range archetypes {
			scores[arch] += pattern.Score
		}
	}

	// Convert to sorted slice
	var result []archetypeScore
	for name, score := range scores {
		result = append(result, archetypeScore{Name: name, Score: score})
	}

	// Sort by score descending
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Score > result[i].Score {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	// Normalize scores to 0-1 range
	if len(result) > 0 && result[0].Score > 0 {
		maxScore := result[0].Score
		for i := range result {
			result[i].Score = result[i].Score / maxScore
		}
	}

	return result
}

// getArchetypesForPattern maps patterns to recommended archetypes.
func getArchetypesForPattern(pattern string) []string {
	mapping := map[string][]string{
		"crud":  {"standard", "api"},
		"api":   {"standard", "api"},
		"async": {"worker"},
		"auth":  {"standard", "api"},
		"grpc":  {"grpc"},
		"mcp":   {"mcp"},
	}

	if archetypes, ok := mapping[pattern]; ok {
		return archetypes
	}
	return []string{"standard"}
}

// EnrichAnalysisWithArchetype adds archetype and confidence to the analysis.
func EnrichAnalysisWithArchetype(analysis *SpecAnalysis, cfg *GenerationConfig, explicitArchetype string) error {
	archetype, confidence, err := SelectArchetype(analysis, cfg, explicitArchetype)
	if err != nil {
		return err
	}

	analysis.Archetype = archetype
	analysis.Confidence = confidence
	return nil
}
