package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/victorzhuk/go-ent/internal/domain"
)

var patternCache = make(map[string]*regexp.Regexp)
var cacheMutex sync.RWMutex

// MatchContext provides additional context for skill matching.
type MatchContext struct {
	Query        string   // The search query
	FileTypes    []string // File extensions (e.g., ".go", ".md")
	TaskType     string   // Task type (e.g., "implement", "review", "debug")
	ActiveSkills []string // Currently loaded skill names
}

// matchTrigger checks if a single explicit trigger matches the query and context.
func matchTrigger(trigger Trigger, query string, ctx *MatchContext) []MatchReason {
	var reasons []MatchReason
	queryLower := strings.ToLower(query)

	for _, pat := range trigger.Patterns {
		if matchesPattern(query, pat) {
			reasons = append(reasons, MatchReason{
				Type:   "pattern",
				Value:  pat,
				Weight: trigger.Weight,
			})
		}
	}

	for _, kw := range trigger.Keywords {
		if matchesKeyword(queryLower, strings.ToLower(kw)) {
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
				if matchFilePattern(fp, fileType) {
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
// Used for skills without explicit triggers for backward compatibility.
func matchDescription(skill *SkillMeta, query string) []MatchReason {
	var reasons []MatchReason
	queryLower := strings.ToLower(query)

	// Extract keywords from "Auto-activates for:" section
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

		if matchesKeyword(queryLower, kw) {
			reasons = append(reasons, MatchReason{
				Type:   "description_keyword",
				Value:  kw,
				Weight: weight,
			})
		}
	}

	return reasons
}

// scoreSkill calculates match score for a single skill based on query and context.
func scoreSkill(skill *SkillMeta, query string, ctx *MatchContext) MatchResult {
	result := MatchResult{
		Skill:     skill,
		Score:     0,
		MatchedBy: []MatchReason{},
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
			reasons := matchTrigger(trigger, query, ctx)
			for _, reason := range reasons {
				result.Score += reason.Weight
				result.MatchedBy = append(result.MatchedBy, reason)
			}
		}
	} else {
		reasons := matchDescription(skill, query)
		for _, reason := range reasons {
			result.Score += reason.Weight
			result.MatchedBy = append(result.MatchedBy, reason)
		}
	}

	return result
}

// matchesPattern checks if query matches regex pattern using a package-level cache.
// Patterns are compiled once and cached for reuse across multiple queries.
// Thread-safe: uses sync.RWMutex for concurrent read access and exclusive write access.
// Cache persists for the lifetime of the package process.
func matchesPattern(query, pattern string) bool {
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
func matchesKeyword(queryLower, keyword string) bool {
	return strings.Contains(queryLower, keyword)
}

// matchFilePattern checks if a file pattern matches a file type.
func matchFilePattern(pattern, fileType string) bool {
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

// Registry manages skill metadata and matching.
type Registry struct {
	skills        []SkillMeta
	runtimeSkills map[string]domain.Skill
	parser        *Parser
	validator     *Validator
	scorer        *QualityScorer
}

// NewRegistry creates a new skill registry.
func NewRegistry() *Registry {
	return &Registry{
		skills:        make([]SkillMeta, 0),
		runtimeSkills: make(map[string]domain.Skill),
		parser:        NewParser(),
		validator:     NewValidator(),
		scorer:        NewQualityScorer(),
	}
}

// MatchResult represents a skill match with its confidence score and reasons.
type MatchResult struct {
	Skill     *SkillMeta    // The matched skill
	Score     float64       // 0.0-1.0 confidence score
	MatchedBy []MatchReason // List of what triggered the match
}

// MatchReason explains why a skill was matched.
type MatchReason struct {
	Type   string  // "keyword", "pattern", "file_type"
	Value  string  // The specific value that matched
	Weight float64 // The trigger weight
}

// Register adds a runtime skill to the registry.
func (r *Registry) Register(skill domain.Skill) error {
	if skill == nil {
		return fmt.Errorf("skill cannot be nil")
	}

	name := skill.Name()
	if name == "" {
		return fmt.Errorf("skill name cannot be empty")
	}

	if _, exists := r.runtimeSkills[name]; exists {
		return fmt.Errorf("skill %s already registered", name)
	}

	r.runtimeSkills[name] = skill
	return nil
}

// Unregister removes a runtime skill from the registry.
func (r *Registry) Unregister(name string) error {
	if _, exists := r.runtimeSkills[name]; !exists {
		return fmt.Errorf("skill %s not found", name)
	}

	delete(r.runtimeSkills, name)
	return nil
}

// GetSkill retrieves a runtime skill by name.
func (r *Registry) GetSkill(name string) (domain.Skill, error) {
	skill, exists := r.runtimeSkills[name]
	if !exists {
		return nil, fmt.Errorf("skill %s not found", name)
	}
	return skill, nil
}

// Load scans a directory for SKILL.md files and loads their metadata.
func (r *Registry) Load(skillsPath string) error {
	return filepath.Walk(skillsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || info.Name() != "SKILL.md" {
			return nil
		}

		meta, err := r.parser.ParseSkillFile(path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		content, err := os.ReadFile(path) // #nosec G304 -- controlled skill file path
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		meta.QualityScore = r.scorer.Score(meta, string(content))
		r.skills = append(r.skills, *meta)
		return nil
	})
}

// RegisterSkill loads a skill from a given path and registers it.
func (r *Registry) RegisterSkill(name, path string) error {
	meta, err := r.parser.ParseSkillFile(path)
	if err != nil {
		return fmt.Errorf("parse skill file: %w", err)
	}

	if meta.Name != name {
		return fmt.Errorf("skill name mismatch: expected %s, got %s", name, meta.Name)
	}

	for _, s := range r.skills {
		if s.Name == name {
			return fmt.Errorf("skill %s already registered", name)
		}
	}

	r.skills = append(r.skills, *meta)
	return nil
}

// UnregisterSkill removes a skill from the metadata list.
func (r *Registry) UnregisterSkill(name string) error {
	for i, s := range r.skills {
		if s.Name == name {
			r.skills = append(r.skills[:i], r.skills[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("skill %s not found", name)
}

// MatchForContext returns skill names that match the given context.
func (r *Registry) MatchForContext(ctx domain.SkillContext) []string {
	var matched []string

	// Check runtime skills first
	for name, skill := range r.runtimeSkills {
		if skill.CanHandle(ctx) {
			matched = append(matched, name)
		}
	}

	// Then check metadata skills
	terms := r.buildSearchTerms(ctx)
	for _, skill := range r.skills {
		if r.matchesContext(skill, terms) {
			matched = append(matched, skill.Name)
		}
	}

	return matched
}

// FindMatchingSkills returns skills with match scores and reasons based on the given query and context.
func (r *Registry) FindMatchingSkills(query string, context ...*MatchContext) []MatchResult {
	if len(context) == 0 || context[0] == nil {
		return r.matchByQuery(query)
	}

	ctx := context[0]
	var results []MatchResult

	for i := range r.skills {
		result := scoreSkill(&r.skills[i], query, ctx)
		if result.Score > 0 {
			boost := r.applyContextBoosts(r.skills[i].Name, ctx)
			result.Score += boost
			results = append(results, result)
		}
	}

	for name, skill := range r.runtimeSkills {
		if strings.Contains(strings.ToLower(skill.Name()), strings.ToLower(query)) {
			results = append(results, MatchResult{
				Skill:     nil,
				Score:     1.0,
				MatchedBy: []MatchReason{{Type: "runtime", Value: name, Weight: 1.0}},
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// matchByQuery performs query-based skill matching (backward compatible).
func (r *Registry) matchByQuery(query string) []MatchResult {
	var results []MatchResult
	queryLower := strings.ToLower(query)

	for _, skill := range r.skills {
		if strings.Contains(strings.ToLower(skill.Name), queryLower) {
			results = append(results, MatchResult{
				Skill:     &skill,
				Score:     0.5,
				MatchedBy: []MatchReason{{Type: "name", Value: skill.Name, Weight: 0.5}},
			})
		}
	}

	for name, skill := range r.runtimeSkills {
		if strings.Contains(strings.ToLower(skill.Name()), queryLower) {
			results = append(results, MatchResult{
				Skill:     nil,
				Score:     0.5,
				MatchedBy: []MatchReason{{Type: "runtime", Value: name, Weight: 0.5}},
			})
		}
	}

	return results
}

// scoreSkills calculates relevance scores for all skills based on query and context.
func (r *Registry) scoreSkills(query string, ctx *MatchContext) map[string]float64 {
	scores := make(map[string]float64)
	queryLower := strings.ToLower(query)

	// Score runtime skills
	for name, skill := range r.runtimeSkills {
		if strings.Contains(strings.ToLower(skill.Name()), queryLower) {
			scores[name] = 1.0
		}
	}

	// Score metadata skills
	for _, skill := range r.skills {
		if strings.Contains(strings.ToLower(skill.Name), queryLower) {
			scores[skill.Name] = 1.0
		}
	}

	// Apply context boosts to scored skills
	for name := range scores {
		scores[name] += r.applyContextBoosts(name, ctx)
	}

	return scores
}

// applyContextBoosts calculates total boost for a skill based on context.
func (r *Registry) applyContextBoosts(skillName string, ctx *MatchContext) float64 {
	var boost float64

	boost += r.fileTypeBoost(skillName, ctx)
	boost += r.taskTypeBoost(skillName, ctx)
	boost += r.affinityBoost(skillName, ctx)

	return boost
}

// fileTypeBoost adds +0.2 if skill has file_pattern triggers matching context FileTypes.
func (r *Registry) fileTypeBoost(skillName string, ctx *MatchContext) float64 {
	if len(ctx.FileTypes) == 0 {
		return 0
	}

	skill, err := r.Get(skillName)
	if err != nil {
		return 0
	}

	for _, trigger := range skill.ExplicitTriggers {
		if len(trigger.FilePatterns) == 0 {
			continue
		}

		for _, fp := range trigger.FilePatterns {
			for _, fileType := range ctx.FileTypes {
				if r.matchesFilePattern(fp, fileType) {
					return 0.2
				}
			}
		}
	}

	return 0
}

// matchesFilePattern checks if a file pattern matches a file type.
func (r *Registry) matchesFilePattern(pattern, fileType string) bool {
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

// taskTypeBoost adds +0.15 if skill triggers match task type from query or context.
func (r *Registry) taskTypeBoost(skillName string, ctx *MatchContext) float64 {
	taskType := ctx.TaskType
	if taskType == "" {
		taskType = r.extractTaskType(ctx.Query)
	}

	if taskType == "" {
		return 0
	}

	skill, err := r.Get(skillName)
	if err != nil {
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

// extractTaskType extracts task type from query keywords.
func (r *Registry) extractTaskType(query string) string {
	queryLower := strings.ToLower(query)

	keywords := []string{"implement", "review", "debug", "test", "refactor"}
	for _, kw := range keywords {
		if strings.Contains(queryLower, kw) {
			return kw
		}
	}

	return ""
}

// affinityBoost adds +0.1 if skill is already active (avoid context switching).
func (r *Registry) affinityBoost(skillName string, ctx *MatchContext) float64 {
	for _, activeSkill := range ctx.ActiveSkills {
		if skillName == activeSkill {
			return 0.1
		}
	}
	return 0
}

// Get retrieves a skill by name.
func (r *Registry) Get(name string) (*SkillMeta, error) {
	for _, skill := range r.skills {
		if skill.Name == name {
			return &skill, nil
		}
	}
	return nil, fmt.Errorf("skill not found: %s", name)
}

// All returns all loaded skills.
func (r *Registry) All() []SkillMeta {
	return r.skills
}

// buildSearchTerms extracts searchable terms from SkillContext.
func (r *Registry) buildSearchTerms(ctx domain.SkillContext) []string {
	var terms []string

	// Add action as term
	if ctx.Action != "" {
		terms = append(terms, strings.ToLower(string(ctx.Action)))
	}

	// Add phase as term
	if ctx.Phase != "" {
		terms = append(terms, strings.ToLower(string(ctx.Phase)))
	}

	// Add agent role as term
	if ctx.Agent != "" {
		terms = append(terms, strings.ToLower(string(ctx.Agent)))
	}

	// Extract terms from metadata
	if ctx.Metadata != nil {
		for key, val := range ctx.Metadata {
			keyLower := strings.ToLower(key)
			terms = append(terms, keyLower)

			// Handle string values
			if strVal, ok := val.(string); ok {
				terms = append(terms, strings.ToLower(strVal))
			}
		}
	}

	return terms
}

// matchesContext checks if a skill's triggers match any context terms.
func (r *Registry) matchesContext(skill SkillMeta, terms []string) bool {
	if len(skill.Triggers) == 0 {
		return false
	}

	for _, trigger := range skill.Triggers {
		for _, term := range terms {
			// Exact match
			if trigger == term {
				return true
			}

			// Partial match (term contains trigger)
			if strings.Contains(term, trigger) {
				return true
			}

			// Partial match (trigger contains term)
			if strings.Contains(trigger, term) {
				return true
			}
		}
	}

	return false
}

// ValidateSkill validates a single skill by name.
func (r *Registry) ValidateSkill(name string) (*ValidationResult, error) {
	meta, err := r.Get(name)
	if err != nil {
		return nil, fmt.Errorf("get skill metadata: %w", err)
	}

	content, err := os.ReadFile(meta.FilePath) // #nosec G304 -- controlled skill file path
	if err != nil {
		return nil, fmt.Errorf("read skill file: %w", err)
	}

	result := r.validator.ValidateWithContext(meta, string(content), r)
	return result, nil
}

// ValidateAll validates all loaded skills and returns aggregate result.
func (r *Registry) ValidateAll() (*ValidationResult, error) {
	if len(r.skills) == 0 {
		return &ValidationResult{
			Valid:  true,
			Issues: []ValidationIssue{},
			Score:  nil,
		}, nil
	}

	var allIssues []ValidationIssue
	totalScore := 0.0

	for _, skill := range r.skills {
		result, err := r.ValidateSkill(skill.Name)
		if err != nil {
			return nil, fmt.Errorf("validate skill %s: %w", skill.Name, err)
		}

		allIssues = append(allIssues, result.Issues...)
		if result.Score != nil {
			totalScore += result.Score.Total
		}
	}

	avgScore := totalScore / float64(len(r.skills))
	valid := true
	for _, issue := range allIssues {
		if issue.Severity == SeverityError {
			valid = false
			break
		}
	}

	return &ValidationResult{
		Valid:  valid,
		Issues: allIssues,
		Score:  &QualityScore{Total: avgScore},
	}, nil
}

// GetQualityReport returns a map of skill names to quality scores.
func (r *Registry) GetQualityReport() map[string]float64 {
	report := make(map[string]float64, len(r.skills))
	for _, skill := range r.skills {
		if skill.QualityScore != nil {
			report[skill.Name] = skill.QualityScore.Total
		}
	}
	return report
}
