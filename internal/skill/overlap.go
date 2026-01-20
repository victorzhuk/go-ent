package skill

import (
	"strings"
)

const (
	overlapThreshold  = 0.7
	triggerWeight     = 0.7
	descriptionWeight = 0.3
)

func calculateOverlap(skill1, skill2 *SkillMeta) float64 {
	triggerOverlap := calculateTriggerOverlap(skill1, skill2)
	textSimilarity := calculateTextSimilarity(skill1.Description, skill2.Description)

	return triggerWeight*triggerOverlap + descriptionWeight*textSimilarity
}

func calculateTriggerOverlap(skill1, skill2 *SkillMeta) float64 {
	triggers1 := skill1.Triggers
	triggers2 := skill2.Triggers

	if len(triggers1) == 0 && len(triggers2) == 0 {
		return 0
	}

	set1 := make(map[string]struct{})
	for _, t := range triggers1 {
		lower := strings.ToLower(t)
		if lower != "" {
			set1[lower] = struct{}{}
		}
	}

	set2 := make(map[string]struct{})
	for _, t := range triggers2 {
		lower := strings.ToLower(t)
		if lower != "" {
			set2[lower] = struct{}{}
		}
	}

	if len(set1) == 0 && len(set2) == 0 {
		return 0
	}

	unionSize := len(set1) + len(set2)

	intersection := 0
	for trigger := range set1 {
		if _, exists := set2[trigger]; exists {
			intersection++
		}
	}

	if unionSize == 0 {
		return 0
	}

	return float64(2*intersection) / float64(unionSize)
}

func calculateTextSimilarity(text1, text2 string) float64 {
	words1 := tokenize(text1)
	words2 := tokenize(text2)

	if len(words1) == 0 && len(words2) == 0 {
		return 0
	}

	set1 := make(map[string]struct{})
	for _, w := range words1 {
		if w != "" {
			set1[w] = struct{}{}
		}
	}

	set2 := make(map[string]struct{})
	for _, w := range words2 {
		if w != "" {
			set2[w] = struct{}{}
		}
	}

	if len(set1) == 0 && len(set2) == 0 {
		return 0
	}

	unionSize := len(set1) + len(set2)

	intersection := 0
	for word := range set1 {
		if _, exists := set2[word]; exists {
			intersection++
		}
	}

	if unionSize == 0 {
		return 0
	}

	return float64(2*intersection) / float64(unionSize)
}

func tokenize(text string) []string {
	lower := strings.ToLower(text)
	words := strings.Fields(lower)

	uniqueWords := make([]string, 0, len(words))
	seen := make(map[string]struct{})

	for _, w := range words {
		w = strings.Trim(w, ".,!?;:\"'()[]{}")
		if w != "" && w != "-" && w != "—" {
			if _, exists := seen[w]; !exists {
				seen[w] = struct{}{}
				uniqueWords = append(uniqueWords, w)
			}
		}
	}

	return uniqueWords
}
