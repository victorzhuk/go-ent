package matcher

import (
	"regexp"
	"sort"
	"strings"
	"sync"

	skdomain "github.com/victorzhuk/go-ent/internal/skill/domain"
)

var (
	patternCache = make(map[string]*regexp.Regexp)
	cacheMutex   sync.RWMutex
)

// MatchContext provides additional context for skill matching.
type MatchContext struct {
	Query        string   // The search query
	FileTypes    []string // File extensions (e.g., ".go", ".md")
	TaskType     string   // Task type (e.g., "implement", "review", "debug")
	ActiveSkills []string // Currently loaded skill names
}

// MatchReason explains why a skill was matched.
type MatchReason struct {
	Type   string  // "keyword", "pattern", "file_type"
	Value  string  // The specific value that matched
	Weight float64 // The trigger weight
}

// DelegationHint provides a hint for delegating work to another skill.
type DelegationHint struct {
	ToSkill string // Name of skill to delegate to
	Reason  string // Why delegation should happen
}

// MatchResult represents a skill match with its confidence score and reasons.
type MatchResult struct {
	Skill       *skdomain.Info   // The matched skill
	Score       float64          // 0.0-1.0 confidence score
	MatchedBy   []MatchReason    // List of what triggered the match
	Delegations []DelegationHint // Hints for skill delegation
}

// QueryMatcher finds skills by text query.
type QueryMatcher interface {
	MatchByQuery(query string) ([]MatchResult, error)
	MatchByContext(query string, ctx *MatchContext) ([]MatchResult, error)
	MatchWithScoring(query string, ctx *MatchContext) ([]MatchResult, error)
}

// CapabilityMatcher finds skills by capability requirements.
type CapabilityMatcher interface {
	MatchByCapability(cType skdomain.CapabilityType) ([]*skdomain.Info, error)
	MatchMultipleCapabilities(types []skdomain.CapabilityType) ([]*skdomain.Info, error)
	MatchToTaskRequirements(requirements map[skdomain.CapabilityType]float64) ([]*skdomain.Info, error)
}

// Matcher combines query and capability matching.
type Matcher interface {
	QueryMatcher
	CapabilityMatcher
}

// skillMatcher implements Matcher.
type skillMatcher struct {
	repo MatcherRepository
}

// MatcherRepository defines the repository interface needed by the matcher.
type MatcherRepository interface {
	ListAll() ([]*skdomain.Info, error)
}

// NewMatcher creates a new Matcher.
func NewMatcher(repo MatcherRepository) Matcher {
	return &skillMatcher{repo: repo}
}

// MatchByQuery finds skills that match a simple query string.
func (m *skillMatcher) MatchByQuery(query string) ([]MatchResult, error) {
	skills, err := m.repo.ListAll()
	if err != nil {
		return nil, err
	}

	var results []MatchResult
	queryLower := strings.ToLower(query)

	for _, skill := range skills {
		if strings.Contains(strings.ToLower(skill.Name), queryLower) {
			results = append(results, MatchResult{
				Skill:       skill,
				Score:       0.5,
				MatchedBy:   []MatchReason{{Type: "name", Value: skill.Name, Weight: 0.5}},
				Delegations: m.extractDelegations(skill),
			})
		}
	}

	return results, nil
}

// MatchByContext finds skills that match a query with additional context.
func (m *skillMatcher) MatchByContext(query string, ctx *MatchContext) ([]MatchResult, error) {
	if ctx == nil {
		return m.MatchByQuery(query)
	}

	return m.MatchWithScoring(query, ctx)
}

// MatchWithScoring finds skills with detailed scoring and reasons.
func (m *skillMatcher) MatchWithScoring(query string, ctx *MatchContext) ([]MatchResult, error) {
	skills, err := m.repo.ListAll()
	if err != nil {
		return nil, err
	}

	var results []MatchResult

	for _, skill := range skills {
		result := m.scoreSkill(skill, query, ctx)
		if result.Score > 0 {
			boost := m.applyContextBoosts(skill, ctx)
			result.Score += boost
			results = append(results, result)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// MatchByCapability finds skills that match specific capabilities.
func (m *skillMatcher) MatchByCapability(cType skdomain.CapabilityType) ([]*skdomain.Info, error) {
	skills, err := m.repo.ListAll()
	if err != nil {
		return nil, err
	}

	var matched []*skdomain.Info
	for _, skill := range skills {
		for _, trigger := range skill.Triggers {
			if strings.Contains(strings.ToLower(trigger), strings.ToLower(string(cType))) {
				matched = append(matched, skill)
				break
			}
		}
	}

	return matched, nil
}

// MatchMultipleCapabilities finds skills that match multiple capabilities.
func (m *skillMatcher) MatchMultipleCapabilities(types []skdomain.CapabilityType) ([]*skdomain.Info, error) {
	skills, err := m.repo.ListAll()
	if err != nil {
		return nil, err
	}

	var matched []*skdomain.Info
	for _, skill := range skills {
		matchCount := 0
		for _, cType := range types {
			for _, trigger := range skill.Triggers {
				if strings.Contains(strings.ToLower(trigger), strings.ToLower(string(cType))) {
					matchCount++
					break
				}
			}
		}
		if matchCount > 0 {
			matched = append(matched, skill)
		}
	}

	return matched, nil
}

// MatchToTaskRequirements matches skills to task requirements.
func (m *skillMatcher) MatchToTaskRequirements(requirements map[skdomain.CapabilityType]float64) ([]*skdomain.Info, error) {
	skills, err := m.repo.ListAll()
	if err != nil {
		return nil, err
	}

	var matched []*skdomain.Info
	for _, skill := range skills {
		totalScore := 0.0
		for cType, minScore := range requirements {
			for _, trigger := range skill.Triggers {
				if strings.Contains(strings.ToLower(trigger), strings.ToLower(string(cType))) {
					totalScore += minScore
					break
				}
			}
		}
		if totalScore > 0 {
			matched = append(matched, skill)
		}
	}

	return matched, nil
}

// scoreSkill calculates match score for a single skill based on query and context.
func (m *skillMatcher) scoreSkill(skill *skdomain.Info, query string, ctx *MatchContext) MatchResult {
	result := MatchResult{
		Skill:       skill,
		Score:       0,
		MatchedBy:   []MatchReason{},
		Delegations: m.extractDelegations(skill),
	}

	queryLower := strings.ToLower(query)

	if strings.Contains(strings.ToLower(skill.Name), queryLower) {
		result.Score += 0.5
		result.MatchedBy = append(result.MatchedBy, MatchReason{
			Type:   "name",
			Value:  skill.Name,
			Weight: 0.5,
		})
	}

	if len(skill.ExplicitTriggers) > 0 {
		for _, trigger := range skill.ExplicitTriggers {
			reasons := m.matchTrigger(trigger, query, ctx)
			for _, reason := range reasons {
				result.Score += reason.Weight
				result.MatchedBy = append(result.MatchedBy, reason)
			}
		}
	} else {
		reasons := m.matchDescription(skill, query)
		for _, reason := range reasons {
			result.Score += reason.Weight
			result.MatchedBy = append(result.MatchedBy, reason)
		}
	}

	return result
}

// matchTrigger checks if a single explicit trigger matches the query and context.
func (m *skillMatcher) matchTrigger(trigger skdomain.Trigger, query string, ctx *MatchContext) []MatchReason {
	var reasons []MatchReason
	queryLower := strings.ToLower(query)

	for _, pat := range trigger.Patterns {
		if m.matchesPattern(query, pat) {
			reasons = append(reasons, MatchReason{
				Type:   "pattern",
				Value:  pat,
				Weight: trigger.Weight,
			})
		}
	}

	for _, kw := range trigger.Keywords {
		if m.matchesKeyword(queryLower, strings.ToLower(kw)) {
			reasons = append(reasons, MatchReason{
				Type:   "keyword",
				Value:  kw,
				Weight: trigger.Weight,
			})
		}
	}

	if ctx != nil {
		for _, fp := range trigger.FilePatterns {
			for _, fileType := range ctx.FileTypes {
				if m.matchFilePattern(fp, fileType) {
					reasons = append(reasons, MatchReason{
						Type:   "file_type",
						Value:  fp,
						Weight: trigger.Weight,
					})
					break
				}
			}
		}
	}

	return reasons
}

// matchDescription extracts keywords from skill description for fallback matching.
func (m *skillMatcher) matchDescription(skill *skdomain.Info, query string) []MatchReason {
	var reasons []MatchReason
	queryLower := strings.ToLower(query)

	const prefix = "Auto-activates for:"
	idx := strings.Index(skill.Description, prefix)
	if idx == -1 {
		return reasons
	}

	rest := skill.Description[idx+len(prefix):]
	endIdx := strings.Index(rest, ".")
	if endIdx == -1 {
		endIdx = len(rest)
	}
	triggerText := rest[:endIdx]

	parts := strings.Split(triggerText, ",")
	weight := 0.6

	for _, part := range parts {
		kw := strings.ToLower(strings.TrimSpace(part))
		if kw == "" {
			continue
		}

		if m.matchesKeyword(queryLower, kw) {
			reasons = append(reasons, MatchReason{
				Type:   "description_keyword",
				Value:  kw,
				Weight: weight,
			})
		}
	}

	return reasons
}

// applyContextBoosts calculates total boost for a skill based on context.
func (m *skillMatcher) applyContextBoosts(skill *skdomain.Info, ctx *MatchContext) float64 {
	if ctx == nil {
		return 0
	}

	var boost float64

	boost += m.fileTypeBoost(skill, ctx)
	boost += m.taskTypeBoost(skill, ctx)
	boost += m.affinityBoost(skill, ctx)

	return boost
}

// fileTypeBoost adds +0.2 if skill has file_pattern triggers matching context FileTypes.
func (m *skillMatcher) fileTypeBoost(skill *skdomain.Info, ctx *MatchContext) float64 {
	if len(ctx.FileTypes) == 0 {
		return 0
	}

	for _, trigger := range skill.ExplicitTriggers {
		if len(trigger.FilePatterns) == 0 {
			continue
		}

		for _, fp := range trigger.FilePatterns {
			for _, fileType := range ctx.FileTypes {
				if m.matchFilePattern(fp, fileType) {
					return 0.2
				}
			}
		}
	}

	return 0
}

// taskTypeBoost adds +0.15 if skill triggers match task type from query or context.
func (m *skillMatcher) taskTypeBoost(skill *skdomain.Info, ctx *MatchContext) float64 {
	taskType := ctx.TaskType
	if taskType == "" {
		taskType = m.extractTaskType(ctx.Query)
	}

	if taskType == "" {
		return 0
	}

	taskTypeLower := strings.ToLower(taskType)

	for _, trigger := range skill.Triggers {
		if strings.Contains(trigger, taskTypeLower) {
			return 0.15
		}
	}

	if strings.Contains(strings.ToLower(skill.Description), taskTypeLower) {
		return 0.15
	}

	for _, trigger := range skill.ExplicitTriggers {
		for _, kw := range trigger.Keywords {
			if strings.Contains(strings.ToLower(kw), taskTypeLower) {
				return 0.15
			}
		}
	}

	return 0
}

// affinityBoost adds +0.1 if skill is already active (avoid context switching).
func (m *skillMatcher) affinityBoost(skill *skdomain.Info, ctx *MatchContext) float64 {
	for _, activeSkill := range ctx.ActiveSkills {
		if skill.Name == activeSkill {
			return 0.1
		}
	}
	return 0
}

// extractTaskType extracts task type from query keywords.
func (m *skillMatcher) extractTaskType(query string) string {
	queryLower := strings.ToLower(query)

	keywords := []string{"implement", "review", "debug", "test", "refactor"}
	for _, kw := range keywords {
		if strings.Contains(queryLower, kw) {
			return kw
		}
	}

	return ""
}

// matchesPattern checks if query matches regex pattern using a package-level cache.
func (m *skillMatcher) matchesPattern(query, pattern string) bool {
	patternLower := strings.ToLower(pattern)

	cacheMutex.RLock()
	re, cached := patternCache[patternLower]
	cacheMutex.RUnlock()

	if cached {
		return re.MatchString(strings.ToLower(query))
	}

	re, err := regexp.Compile(patternLower)
	if err != nil {
		return false
	}

	cacheMutex.Lock()
	patternCache[patternLower] = re
	cacheMutex.Unlock()

	return re.MatchString(strings.ToLower(query))
}

// matchesKeyword checks if query contains keyword (exact or substring).
func (m *skillMatcher) matchesKeyword(queryLower, keyword string) bool {
	return strings.Contains(queryLower, keyword)
}

// matchFilePattern checks if a file pattern matches a file type.
func (m *skillMatcher) matchFilePattern(pattern, fileType string) bool {
	pattern = strings.ToLower(pattern)
	fileType = strings.ToLower(fileType)

	if pattern == fileType {
		return true
	}

	if strings.HasPrefix(pattern, "*") {
		ext := strings.TrimPrefix(pattern, "*")
		return fileType == ext || strings.HasSuffix(fileType, ext)
	}

	return false
}

// extractDelegations extracts delegation hints from a skill.
func (m *skillMatcher) extractDelegations(skill *skdomain.Info) []DelegationHint {
	if len(skill.DelegatesTo) == 0 {
		return nil
	}

	var hints []DelegationHint
	for toSkill, reason := range skill.DelegatesTo {
		hints = append(hints, DelegationHint{
			ToSkill: toSkill,
			Reason:  reason,
		})
	}

	return hints
}
