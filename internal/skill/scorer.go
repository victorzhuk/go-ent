package skill

import "strings"

type QualityScorer struct{}

func NewQualityScorer() *QualityScorer {
	return &QualityScorer{}
}

func (s *QualityScorer) Score(meta *SkillMeta, content string) float64 {
	frontmatterScore := s.scoreFrontmatter(meta)
	structureScore := s.scoreStructure(content)
	contentScore := s.scoreContent(content)
	triggerScore := s.scoreTriggers(meta)

	return frontmatterScore + structureScore + contentScore + triggerScore
}

func (s *QualityScorer) scoreFrontmatter(meta *SkillMeta) float64 {
	score := 0.0

	if meta.Name != "" {
		score += 5.0
	}
	if meta.Description != "" {
		score += 5.0
	}
	if meta.Version != "" {
		score += 5.0
	}
	if len(meta.Tags) > 0 {
		score += 5.0
	}

	return score
}

func (s *QualityScorer) scoreStructure(content string) float64 {
	score := 0.0

	requiredSections := []string{"<role>", "<instructions>", "<examples>"}
	for _, section := range requiredSections {
		if strings.Contains(content, section) {
			score += 10.0
		}
	}

	return score
}

func (s *QualityScorer) scoreContent(content string) float64 {
	score := 0.0

	exampleCount := strings.Count(content, "<example>")
	if exampleCount >= 2 {
		score += 15.0
	} else if exampleCount == 1 {
		score += 10.0
	}

	if strings.Contains(content, "<edge_cases>") {
		score += 15.0
	}

	return score
}

func (s *QualityScorer) scoreTriggers(meta *SkillMeta) float64 {
	if len(meta.Triggers) == 0 {
		return 0.0
	}

	if len(meta.Triggers) >= 3 {
		return 20.0
	}

	return float64(len(meta.Triggers)) * 6.67
}
