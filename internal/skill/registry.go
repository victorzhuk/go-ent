package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/victorzhuk/go-ent/internal/domain"
)

// MatchContext provides additional context for skill matching.
type MatchContext struct {
	Query        string   // The search query
	FileTypes    []string // File extensions (e.g., ".go", ".md")
	TaskType     string   // Task type (e.g., "implement", "review", "debug")
	ActiveSkills []string // Currently loaded skill names
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

// FindMatchingSkills returns skill names that match the given query, optionally with context.
// When context is provided, it applies context boosting to rank skills by relevance.
// When context is empty, falls back to query-only matching (backward compatible).
func (r *Registry) FindMatchingSkills(query string, context ...*MatchContext) []string {
	if len(context) == 0 || context[0] == nil {
		return r.matchByQuery(query)
	}

	ctx := context[0]
	scores := r.scoreSkills(query, ctx)

	// Sort by score descending
	type namedScore struct {
		name  string
		score float64
	}
	var ranked []namedScore
	for name, score := range scores {
		if score > 0 {
			ranked = append(ranked, namedScore{name, score})
		}
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})

	result := make([]string, len(ranked))
	for i, r := range ranked {
		result[i] = r.name
	}

	return result
}

// matchByQuery performs query-based skill matching.
func (r *Registry) matchByQuery(query string) []string {
	var matched []string
	queryLower := strings.ToLower(query)

	// Check runtime skills
	for name, skill := range r.runtimeSkills {
		if strings.Contains(strings.ToLower(skill.Name()), queryLower) {
			matched = append(matched, name)
		}
	}

	// Check metadata skills
	for _, skill := range r.skills {
		if strings.Contains(strings.ToLower(skill.Name), queryLower) {
			matched = append(matched, skill.Name)
		}
	}

	return matched
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
		if trigger.FilePattern == "" {
			continue
		}

		for _, fileType := range ctx.FileTypes {
			if r.matchesFilePattern(trigger.FilePattern, fileType) {
				return 0.2
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

	result := r.validator.Validate(meta, string(content))
	return result, nil
}

// ValidateAll validates all loaded skills and returns aggregate result.
func (r *Registry) ValidateAll() (*ValidationResult, error) {
	if len(r.skills) == 0 {
		return &ValidationResult{
			Valid:  true,
			Issues: []ValidationIssue{},
			Score:  0,
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
		totalScore += result.Score
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
		Score:  avgScore,
	}, nil
}

// GetQualityReport returns a map of skill names to quality scores.
func (r *Registry) GetQualityReport() map[string]float64 {
	report := make(map[string]float64, len(r.skills))
	for _, skill := range r.skills {
		report[skill.Name] = skill.QualityScore
	}
	return report
}
